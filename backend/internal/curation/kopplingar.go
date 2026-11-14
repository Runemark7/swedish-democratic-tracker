package curation

import (
	"context"
	"encoding/json"
	"fmt"

	"riksdagskollen/internal/curation/domain"
	"riksdagskollen/internal/curation/ports"
)

// Candidates is the bot's read: voteringar new since the last closed
// Veckobatch, with the approved Löften they might test. limit 1..200.
//
// With a mandatePeriod (e.g. "2022-2026") the read is instead every decided
// votering of that period, whenever ingested, paired with the Löften of that
// period that may take a koppling: approved ones, and archived ones with a
// quote and a source URL.
func (s *Service) Candidates(ctx context.Context, limit int, afterRef int64, mandatePeriod string) (*domain.Candidates, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	return s.repo.Candidates(ctx, limit, afterRef, mandatePeriod)
}

// SubmitSuggestions stores every suggestion: at or above the threshold as
// föreslagen, below it as bortsorterad. Both stay out of every public
// response; bortsorterade ones are kept only so they are not suggested again.
// A koppling outside the Löfte's mandate period rejects the whole batch.
func (s *Service) SubmitSuggestions(ctx context.Context, key string, in domain.SuggestionBatch) (*domain.SuggestionResponse, error) {
	if err := validKey(key); err != nil {
		return nil, err
	}
	if err := in.Normalize(); err != nil {
		return nil, err
	}
	hash := hashOf("suggest-kopplingar", in)
	out := &domain.SuggestionResponse{}
	err := s.repo.InTx(ctx, func(st ports.Store) error {
		*out = domain.SuggestionResponse{}
		// One audit entry per koppling, keyed key#i. Entry key#0 marks the batch as handled.
		first, err := lockAndFind(ctx, st, key+"#0")
		if err != nil {
			return err
		}
		if first != nil {
			if first.RequestHash != hash {
				return domain.ErrKeyReuse
			}
			out.Replayed = true
			for i := 0; ; i++ {
				e, err := st.EntryByKey(ctx, fmt.Sprintf("%s#%d", key, i))
				if err != nil {
					return err
				}
				if e == nil {
					return nil
				}
				var k domain.Koppling
				if err := json.Unmarshal(e.NewValue, &k); err != nil {
					return err
				}
				out.Results = append(out.Results, domain.SuggestionResult{KopplingID: k.ID, LoefteID: k.LoefteID,
					VoteringID: k.VoteringID, State: k.State, Status: "created"})
			}
		}
		n := 0
		for i, sg := range in.Suggestions {
			at := fmt.Sprintf("suggestion %d", i+1)
			l, err := st.LoefteByID(ctx, sg.LoefteID)
			if err != nil {
				return err
			}
			if l == nil || !l.TakesKopplingar() {
				// Same answer for unknown and non-public Löften: do not reveal proposed ones.
				return &domain.ValidationError{Msg: fmt.Sprintf(
					"%s: Löfte %d is not an approved Löfte (or an archived one with a quote and a source URL)", at, sg.LoefteID)}
			}
			v, err := st.VoteringForLoefte(ctx, sg.VoteringID, l)
			if err != nil {
				return err
			}
			if v == nil {
				return &domain.ValidationError{Msg: fmt.Sprintf("%s: unknown votering %q", at, sg.VoteringID)}
			}
			if !v.InPeriod {
				return &domain.ValidationError{Msg: fmt.Sprintf(
					"%s: votering %s is outside the mandate period %s of Löfte %d", at, sg.VoteringID, l.MandatePeriod, l.ID)}
			}
			if ex, err := st.KopplingByPair(ctx, l.ID, v.Ref); err != nil {
				return err
			} else if ex != nil {
				out.Results = append(out.Results, domain.SuggestionResult{KopplingID: ex.ID, LoefteID: l.ID,
					VoteringID: sg.VoteringID, State: ex.State, Status: "existing"})
				continue
			}
			state := domain.KopplingForeslagen
			if sg.Confidence < *in.Threshold {
				state = domain.KopplingBortsorterad
			}
			k, err := st.InsertKoppling(ctx, l.ID, v.Ref, domain.Koppling{
				LoefteID: l.ID, VoteringID: sg.VoteringID, Direction: sg.Direction, State: state,
				Confidence: sg.Confidence, Threshold: *in.Threshold, Reasoning: sg.Reasoning, Origin: Actor})
			if err != nil {
				return err
			}
			b, _ := json.Marshal(k)
			if _, err := st.AppendEntry(ctx, &domain.Entry{
				Actor: Actor, OnBehalfOf: in.OnBehalfOf, ChatRef: in.ChatRef,
				Action: domain.ActionUpsert, Entity: domain.EntityKoppling, EntityID: k.ID,
				NewValue: b, RequestKey: fmt.Sprintf("%s#%d", key, n), RequestHash: hash}); err != nil {
				return err
			}
			n++
			out.Results = append(out.Results, domain.SuggestionResult{KopplingID: k.ID, LoefteID: l.ID,
				VoteringID: sg.VoteringID, State: state, Status: "created"})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

func lockAndFind(ctx context.Context, st ports.Store, key string) (*domain.Entry, error) {
	if err := st.LockRequest(ctx, key); err != nil {
		return nil, err
	}
	return st.EntryByKey(ctx, key)
}

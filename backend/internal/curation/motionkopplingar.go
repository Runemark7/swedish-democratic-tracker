package curation

import (
	"context"
	"encoding/json"
	"fmt"

	"riksdagskollen/internal/curation/domain"
	"riksdagskollen/internal/curation/ports"
)

// SubmitMotionSuggestions stores the bot's suggested motion links: at or above
// the threshold as foreslagen, below it as bortsorterad. Neither is public;
// only a link the Redaktör confirms in a Veckobatch is. A pair suggested
// before, in any state, is left as it is.
func (s *Service) SubmitMotionSuggestions(ctx context.Context, key string, in domain.MotionSuggestionBatch) (*domain.MotionSuggestionResponse, error) {
	if err := validKey(key); err != nil {
		return nil, err
	}
	if err := in.Normalize(); err != nil {
		return nil, err
	}
	hash := hashOf("suggest-motion-kopplingar", in)
	out := &domain.MotionSuggestionResponse{}
	err := s.repo.InTx(ctx, func(st ports.Store) error {
		*out = domain.MotionSuggestionResponse{}
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
				var k domain.MotionKoppling
				if err := json.Unmarshal(e.NewValue, &k); err != nil {
					return err
				}
				out.Results = append(out.Results, domain.MotionSuggestionResult{MotionKopplingID: k.ID, Kind: k.Kind, State: k.State, Status: "created"})
			}
		}
		n := 0
		for i, sg := range in.Suggestions {
			k, err := st.ResolveMotionSuggestion(ctx, sg)
			if err != nil {
				if ve, ok := err.(*domain.ValidationError); ok {
					return &domain.ValidationError{Msg: fmt.Sprintf("suggestion %d: %s", i+1, ve.Msg)}
				}
				return err
			}
			if ex, err := st.MotionKopplingByPair(ctx, *k); err != nil {
				return err
			} else if ex != nil {
				out.Results = append(out.Results, domain.MotionSuggestionResult{MotionKopplingID: ex.ID, Kind: ex.Kind, State: ex.State, Status: "existing"})
				continue
			}
			k.State = domain.KopplingForeslagen
			if sg.Confidence < *in.Threshold {
				k.State = domain.KopplingBortsorterad
			}
			k.Confidence, k.Threshold, k.Reasoning, k.Origin = sg.Confidence, *in.Threshold, sg.Reasoning, Actor
			k, err = st.InsertMotionKoppling(ctx, *k)
			if err != nil {
				return err
			}
			if _, err := st.AppendEntry(ctx, &domain.Entry{
				Actor: Actor, OnBehalfOf: in.OnBehalfOf, ChatRef: in.ChatRef,
				Action: domain.ActionUpsert, Entity: domain.EntityMotionKoppling, EntityID: k.ID,
				NewValue: snapMK(k), RequestKey: fmt.Sprintf("%s#%d", key, n), RequestHash: hash}); err != nil {
				return err
			}
			n++
			out.Results = append(out.Results, domain.MotionSuggestionResult{MotionKopplingID: k.ID, Kind: k.Kind, State: k.State, Status: "created"})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

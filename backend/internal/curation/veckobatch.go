package curation

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"riksdagskollen/internal/curation/domain"
	"riksdagskollen/internal/curation/ports"
)

// LatestBatch and Batch are the bot's reads for the chat post.
func (s *Service) LatestBatch(ctx context.Context) (*domain.Batch, error) {
	return s.repo.LatestBatch(ctx)
}

func (s *Service) Batch(ctx context.Context, id int64) (*domain.Batch, error) {
	return s.repo.BatchByID(ctx, id)
}

// CreateBatch builds the week's Veckobatch from the strongest open
// suggestions (at most 20). Items the Redaktör did not answer in earlier
// batches are still open suggestions and come first, renumbered in this batch.
// Creating a batch closes the previous one. Nothing is decided here.
func (s *Service) CreateBatch(ctx context.Context, key string, in domain.BatchRequest) (*domain.Batch, error) {
	if err := validKey(key); err != nil {
		return nil, err
	}
	max, err := in.Normalize()
	if err != nil {
		return nil, err
	}
	hash := hashOf("create-veckobatch", struct {
		Max int
		P   domain.Provenance
	}{max, in.Provenance})
	var out *domain.Batch
	err = s.repo.InTx(ctx, func(st ports.Store) error {
		if err := st.LockBatches(ctx); err != nil {
			return err
		}
		if e, err := lockAndFind(ctx, st, key); err != nil {
			return err
		} else if e != nil {
			if e.RequestHash != hash {
				return domain.ErrKeyReuse
			}
			b, err := st.BatchByID(ctx, e.EntityID)
			if err != nil || b == nil {
				return err
			}
			b.Replayed = true
			out = b
			return nil
		}
		open, err := st.OpenSuggestions(ctx, max)
		if err != nil {
			return err
		}
		if len(open) == 0 {
			return &domain.ValidationError{Msg: "no open suggestions to put in a Veckobatch"}
		}
		if err := st.CloseOpenBatches(ctx); err != nil {
			return err
		}
		id, err := st.InsertBatch(ctx)
		if err != nil {
			return err
		}
		ids, mids := []int64{}, []int64{}
		for i, o := range open {
			if o.MotionKopplingID != 0 {
				mids = append(mids, o.MotionKopplingID)
			} else {
				ids = append(ids, o.KopplingID)
			}
			if err := st.InsertBatchItem(ctx, id, i+1, o); err != nil {
				return err
			}
		}
		nv, _ := json.Marshal(map[string]any{"id": id, "kopplingIds": ids, "motionKopplingIds": mids})
		if _, err := st.AppendEntry(ctx, &domain.Entry{
			Actor: Actor, OnBehalfOf: in.OnBehalfOf, ChatRef: in.ChatRef,
			Action: domain.ActionUpsert, Entity: domain.EntityVeckobatch, EntityID: id,
			NewValue: nv, RequestKey: key, RequestHash: hash}); err != nil {
			return err
		}
		out, err = st.BatchByID(ctx, id)
		return err
	})
	return out, err
}

// AnswerBatch applies the Redaktör's short-form answer to batch id in one
// transaction: every named item is decided or none is. ja confirms the
// koppling with the suggested direction, ja-omvaent confirms it reversed, nej
// rejects it. Unnamed items without a "rest" answer stay open and roll over.
// Each decision is its own Ändringslogg entry (key#0, key#1, …) and can be undone.
func (s *Service) AnswerBatch(ctx context.Context, key string, id int64, in domain.AnswerRequest) (*domain.AnswerResponse, error) {
	if err := validKey(key); err != nil {
		return nil, err
	}
	if err := in.Normalize(); err != nil {
		return nil, err
	}
	hash := hashOf("answer-veckobatch", struct {
		ID int64
		R  domain.AnswerRequest
	}{id, in})
	out := &domain.AnswerResponse{}
	err := s.repo.InTx(ctx, func(st ports.Store) error {
		*out = domain.AnswerResponse{Decided: []domain.ItemOutcome{}, Unanswered: []int{}}
		first, err := lockAndFind(ctx, st, key+"#0")
		if err != nil {
			return err
		}
		if first != nil {
			if first.RequestHash != hash {
				return domain.ErrKeyReuse
			}
			return replayAnswers(ctx, st, key, id, out)
		}
		b, err := st.BatchByID(ctx, id)
		if err != nil {
			return err
		}
		if b == nil {
			return domain.ErrNotFound
		}
		if b.ClosedAt != nil {
			return fmt.Errorf("%w: Veckobatch %d is closed; its open items were rolled into a later batch", domain.ErrConflict, id)
		}
		// Resolve the plan first so a bad number rejects the whole request.
		plan := map[int]string{}
		for _, a := range in.Answers {
			it := b.Item(a.Nr)
			if it == nil {
				return &domain.ValidationError{Msg: fmt.Sprintf("Veckobatch %d has no item %d", id, a.Nr)}
			}
			plan[a.Nr] = a.Answer
		}
		for _, it := range b.Items {
			if _, named := plan[it.Nr]; !named && in.Rest != "" && it.Answer == nil {
				plan[it.Nr] = in.Rest
			}
		}
		nrs := make([]int, 0, len(plan))
		for nr := range plan {
			nrs = append(nrs, nr)
		}
		sort.Ints(nrs)
		if len(nrs) == 0 {
			return &domain.ValidationError{Msg: "nothing to decide: every item is already answered"}
		}
		for _, nr := range nrs {
			it := b.Item(nr)
			if it.Answer != nil || it.State != domain.KopplingForeslagen {
				return fmt.Errorf("%w: item %d is already answered", domain.ErrConflict, nr)
			}
		}
		// Resolve every decision before writing so a refused answer ("ja men
		// omvänt" on a motion link) rejects the whole request.
		type decision struct{ state, dir string }
		dec := map[int]decision{}
		for _, nr := range nrs {
			it := b.Item(nr)
			if it.Kind == domain.ItemLoefte {
				state, dir := domain.Decision(plan[nr], it.Direction)
				dec[nr] = decision{state, dir}
				continue
			}
			state, err := domain.MotionDecision(plan[nr])
			if err != nil {
				return fmt.Errorf("item %d: %w", nr, err)
			}
			dec[nr] = decision{state, ""}
		}
		for i, nr := range nrs {
			it := b.Item(nr)
			state, dir := dec[nr].state, dec[nr].dir
			var ob, nb json.RawMessage
			entity, entityID := domain.EntityKoppling, it.KopplingID
			if it.Kind == domain.ItemLoefte {
				base := domain.Koppling{ID: it.KopplingID, LoefteID: it.LoefteID, VoteringID: it.VoteringID,
					Confidence: it.Confidence, Reasoning: it.Reasoning, VeckobatchID: &id, Nr: &nr}
				before, after := base, base
				before.State, before.Direction = it.State, it.Direction
				after.State, after.Direction, after.Answer = state, dir, plan[nr]
				if err := st.DecideItem(ctx, id, nr, it.KopplingID, state, dir, plan[nr]); err != nil {
					return err
				}
				ob, _ = json.Marshal(before)
				nb, _ = json.Marshal(after)
			} else {
				entity, entityID = domain.EntityMotionKoppling, it.MotionKopplingID
				before, err := st.MotionKopplingByID(ctx, it.MotionKopplingID)
				if err != nil {
					return err
				}
				if before == nil {
					return domain.ErrNotFound
				}
				after := *before
				after.State, after.Answer = state, plan[nr]
				if err := st.DecideMotionItem(ctx, id, nr, it.MotionKopplingID, state, plan[nr]); err != nil {
					return err
				}
				ob, nb = snapMK(before), snapMK(&after)
			}
			e, err := st.AppendEntry(ctx, &domain.Entry{
				Actor: Actor, OnBehalfOf: in.OnBehalfOf, ChatRef: in.ChatRef,
				Action: domain.ActionUpsert, Entity: entity, EntityID: entityID,
				OldValue: ob, NewValue: nb, RequestKey: fmt.Sprintf("%s#%d", key, i), RequestHash: hash})
			if err != nil {
				return err
			}
			out.Decided = append(out.Decided, domain.ItemOutcome{Nr: nr, Kind: it.Kind, KopplingID: it.KopplingID,
				MotionKopplingID: it.MotionKopplingID, Answer: plan[nr], State: state, Direction: dir, EntryID: e.ID})
		}
		for _, it := range b.Items {
			if _, done := plan[it.Nr]; !done && it.Answer == nil {
				out.Unanswered = append(out.Unanswered, it.Nr)
			}
		}
		if len(out.Unanswered) == 0 {
			out.Closed = true
			return st.CloseBatch(ctx, id)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

func replayAnswers(ctx context.Context, st ports.Store, key string, id int64, out *domain.AnswerResponse) error {
	out.Replayed = true
	for i := 0; ; i++ {
		e, err := st.EntryByKey(ctx, fmt.Sprintf("%s#%d", key, i))
		if err != nil {
			return err
		}
		if e == nil {
			break
		}
		if e.Entity == domain.EntityMotionKoppling {
			var k domain.MotionKoppling
			if err := json.Unmarshal(e.NewValue, &k); err != nil {
				return err
			}
			nr := 0
			if k.Nr != nil {
				nr = *k.Nr
			}
			out.Decided = append(out.Decided, domain.ItemOutcome{Nr: nr, Kind: k.Kind, MotionKopplingID: k.ID,
				Answer: k.Answer, State: k.State, EntryID: e.ID})
			continue
		}
		var k domain.Koppling
		if err := json.Unmarshal(e.NewValue, &k); err != nil {
			return err
		}
		nr := 0
		if k.Nr != nil {
			nr = *k.Nr
		}
		out.Decided = append(out.Decided, domain.ItemOutcome{Nr: nr, Kind: domain.ItemLoefte, KopplingID: k.ID, Answer: k.Answer,
			State: k.State, Direction: k.Direction, EntryID: e.ID})
	}
	b, err := st.BatchByID(ctx, id)
	if err != nil || b == nil {
		return err
	}
	out.Closed = b.ClosedAt != nil
	return nil
}

func snapK(k *domain.Koppling) json.RawMessage {
	b, _ := json.Marshal(k)
	return b
}

// undoKoppling reverses a Veckobatch decision: the koppling goes back to
// foreslagen with its suggested direction and the item to unanswered, so the
// suggestion is open for a later batch again. Only decisions can be undone
// (entries with an earlier value); a suggestion entry cannot.
func undoKoppling(ctx context.Context, st ports.Store, target *domain.Entry) (cur, restored json.RawMessage, err error) {
	if len(target.OldValue) == 0 || string(target.OldValue) == "null" {
		return nil, nil, &domain.ValidationError{Msg: "a suggestion entry cannot be undone; only Veckobatch decisions can"}
	}
	var want, restore domain.Koppling
	if err := json.Unmarshal(target.NewValue, &want); err != nil {
		return nil, nil, err
	}
	if err := json.Unmarshal(target.OldValue, &restore); err != nil {
		return nil, nil, err
	}
	c, err := st.KopplingByID(ctx, target.EntityID)
	if err != nil {
		return nil, nil, err
	}
	if c == nil {
		return nil, nil, domain.ErrNotFound
	}
	if c.State != want.State || c.Direction != want.Direction || c.Answer != want.Answer {
		return nil, nil, fmt.Errorf("%w: the value changed after entry %d", domain.ErrConflict, target.ID)
	}
	if want.VeckobatchID == nil || want.Nr == nil {
		return nil, nil, &domain.ValidationError{Msg: "entry has no Veckobatch item to restore"}
	}
	if err := st.DecideItem(ctx, *want.VeckobatchID, *want.Nr, c.ID, restore.State, restore.Direction, restore.Answer); err != nil {
		return nil, nil, err
	}
	return snapK(c), snapK(&restore), nil
}

func snapMK(k *domain.MotionKoppling) json.RawMessage {
	b, _ := json.Marshal(k)
	return b
}

// undoMotionKoppling reverses a Veckobatch decision on a motion link: it goes
// back to foreslagen (so it stops being public at once) and the item to
// unanswered. A suggestion entry cannot be undone.
func undoMotionKoppling(ctx context.Context, st ports.Store, target *domain.Entry) (cur, restored json.RawMessage, err error) {
	if len(target.OldValue) == 0 || string(target.OldValue) == "null" {
		return nil, nil, &domain.ValidationError{Msg: "a suggestion entry cannot be undone; only Veckobatch decisions can"}
	}
	var want, restore domain.MotionKoppling
	if err := json.Unmarshal(target.NewValue, &want); err != nil {
		return nil, nil, err
	}
	if err := json.Unmarshal(target.OldValue, &restore); err != nil {
		return nil, nil, err
	}
	c, err := st.MotionKopplingByID(ctx, target.EntityID)
	if err != nil {
		return nil, nil, err
	}
	if c == nil {
		return nil, nil, domain.ErrNotFound
	}
	if c.State != want.State || c.Answer != want.Answer {
		return nil, nil, fmt.Errorf("%w: the value changed after entry %d", domain.ErrConflict, target.ID)
	}
	if want.VeckobatchID == nil || want.Nr == nil {
		return nil, nil, &domain.ValidationError{Msg: "entry has no Veckobatch item to restore"}
	}
	if err := st.DecideMotionItem(ctx, *want.VeckobatchID, *want.Nr, c.ID, restore.State, restore.Answer); err != nil {
		return nil, nil, err
	}
	return snapMK(c), snapMK(&restore), nil
}

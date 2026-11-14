package curation

import (
	"context"
	"encoding/json"
	"fmt"

	"riksdagskollen/internal/curation/domain"
	"riksdagskollen/internal/curation/ports"
)

func snapL(l *domain.Loefte) json.RawMessage {
	if l == nil {
		return nil
	}
	b, _ := json.Marshal(l)
	return b
}

func unsnapL(raw json.RawMessage) (*domain.Loefte, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}
	var v domain.Loefte
	if err := json.Unmarshal(raw, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

func sameL(a, b *domain.Loefte) bool {
	if a == nil || b == nil {
		return a == b
	}
	return a.Equal(*b)
}

// ProposeLoefte stores a new Löfte as "proposed": invisible to every public
// endpoint until ApproveLoefte. Quote and sourceUrl are mandatory.
func (s *Service) ProposeLoefte(ctx context.Context, key string, in domain.LoefteProposal) (*domain.Result, error) {
	if err := validKey(key); err != nil {
		return nil, err
	}
	if err := in.Normalize(); err != nil {
		return nil, err
	}
	hash := hashOf("propose-loefte", in)
	var out *domain.Result
	err := s.repo.InTx(ctx, func(st ports.Store) error {
		if r, err := replay(ctx, st, key, hash); err != nil || r != nil {
			out = r
			return err
		}
		if dup, err := st.LoefteByQuote(ctx, in.Party, in.ValmanifestID, in.Quote); err != nil {
			return err
		} else if dup != nil {
			return fmt.Errorf("%w: this quote is already Löfte %d (%s)", domain.ErrConflict, dup.ID, dup.Visibility)
		}
		ins, err := st.InsertLoefte(ctx, in.Snapshot())
		if err != nil {
			return err
		}
		e, err := st.AppendEntry(ctx, &domain.Entry{
			Actor: Actor, OnBehalfOf: in.OnBehalfOf, ChatRef: in.ChatRef,
			Action: domain.ActionUpsert, Entity: domain.EntityLoefte, EntityID: ins.ID,
			NewValue: snapL(ins), RequestKey: key, RequestHash: hash,
		})
		out = &domain.Result{Changed: true, Entry: e}
		return err
	})
	return out, err
}

// change applies mutate to Löfte id inside one logged transaction. mutate
// returns the next snapshot, or the same one for a no-op.
func (s *Service) changeLoefte(ctx context.Context, key, kind string, id int64, req any, p domain.Provenance,
	mutate func(cur domain.Loefte) (domain.Loefte, error)) (*domain.Result, error) {
	if err := validKey(key); err != nil {
		return nil, err
	}
	hash := hashOf(kind, struct {
		ID  int64
		Req any
	}{id, req})
	var out *domain.Result
	err := s.repo.InTx(ctx, func(st ports.Store) error {
		if r, err := replay(ctx, st, key, hash); err != nil || r != nil {
			out = r
			return err
		}
		cur, err := st.LoefteByID(ctx, id)
		if err != nil {
			return err
		}
		if cur == nil {
			return domain.ErrNotFound
		}
		next, err := mutate(*cur)
		if err != nil {
			return err
		}
		if cur.Equal(next) {
			out = &domain.Result{Changed: false}
			return nil
		}
		if err := st.UpdateLoefte(ctx, next); err != nil {
			return err
		}
		e, err := st.AppendEntry(ctx, &domain.Entry{
			Actor: Actor, OnBehalfOf: p.OnBehalfOf, ChatRef: p.ChatRef,
			Action: domain.ActionUpsert, Entity: domain.EntityLoefte, EntityID: id,
			OldValue: snapL(cur), NewValue: snapL(&next), RequestKey: key, RequestHash: hash,
		})
		out = &domain.Result{Changed: true, Entry: e}
		return err
	})
	return out, err
}

func notEditable(l domain.Loefte) error {
	return fmt.Errorf("%w: Löfte %d is %s and cannot be changed", domain.ErrConflict, l.ID, l.Visibility)
}

func sameStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// EditLoefte replaces the editable fields of a proposed or approved Löfte;
// on an archived one only quote, locator and sourceUrl may differ.
// The party, Valmanifest and visibility stay as they are.
func (s *Service) EditLoefte(ctx context.Context, key string, id int64, in domain.LoefteEdit) (*domain.Result, error) {
	if err := in.Normalize(); err != nil {
		return nil, err
	}
	return s.changeLoefte(ctx, key, "edit-loefte", id, in, in.Provenance, func(cur domain.Loefte) (domain.Loefte, error) {
		if cur.Visibility == domain.VisArchived {
			// Archived Löften: only quote, locator (källdokument) and sourceUrl may change.
			if in.Text != cur.Text || in.Topic != cur.Topic || in.Specificity != cur.Specificity ||
				!sameStrings(in.Utskott, cur.Utskott) {
				return cur, fmt.Errorf("%w: Löfte %d is archived; only quote, locator and sourceUrl can be changed", domain.ErrConflict, cur.ID)
			}
			cur.Quote, cur.Locator, cur.SourceURL = in.Quote, in.Locator, in.SourceURL
			return cur, nil
		}
		if cur.Visibility != domain.VisProposed && cur.Visibility != domain.VisApproved {
			return cur, notEditable(cur)
		}
		cur.Text, cur.Quote, cur.Locator, cur.SourceURL = in.Text, in.Quote, in.Locator, in.SourceURL
		cur.Topic, cur.Specificity, cur.Utskott = in.Topic, in.Specificity, in.Utskott
		return cur, nil
	})
}

// ApproveLoefte makes a proposed Löfte public. Approving an approved one is a no-op.
func (s *Service) ApproveLoefte(ctx context.Context, key string, id int64, p domain.Provenance) (*domain.Result, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}
	return s.changeLoefte(ctx, key, "approve-loefte", id, p, p, func(cur domain.Loefte) (domain.Loefte, error) {
		switch cur.Visibility {
		case domain.VisApproved:
			return cur, nil
		case domain.VisProposed:
			if cur.Quote == "" || cur.SourceURL == "" {
				return cur, &domain.ValidationError{Msg: "a Löfte needs a quote and a source URL to be approved"}
			}
			cur.Visibility = domain.VisApproved
			return cur, nil
		}
		return cur, notEditable(cur)
	})
}

// StrikeLoefte withdraws a proposed or approved Löfte from the public view.
// The row stays (struck) so the decision is traceable. Striking a struck one is a no-op.
func (s *Service) StrikeLoefte(ctx context.Context, key string, id int64, p domain.Provenance) (*domain.Result, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}
	return s.changeLoefte(ctx, key, "strike-loefte", id, p, p, func(cur domain.Loefte) (domain.Loefte, error) {
		switch cur.Visibility {
		case domain.VisStruck:
			return cur, nil
		case domain.VisProposed, domain.VisApproved:
			cur.Visibility = domain.VisStruck
			return cur, nil
		}
		return cur, notEditable(cur)
	})
}

func undoLoefte(ctx context.Context, st ports.Store, target *domain.Entry) (cur, restored json.RawMessage, err error) {
	want, err := unsnapL(target.NewValue)
	if err != nil {
		return nil, nil, err
	}
	restore, err := unsnapL(target.OldValue)
	if err != nil {
		return nil, nil, err
	}
	c, err := st.LoefteByID(ctx, target.EntityID)
	if err != nil {
		return nil, nil, err
	}
	if !sameL(c, want) {
		return nil, nil, fmt.Errorf("%w: the value changed after entry %d", domain.ErrConflict, target.ID)
	}
	switch {
	case restore == nil:
		err = st.DeleteLoefte(ctx, target.EntityID)
	case c == nil:
		_, err = st.InsertLoefte(ctx, *restore)
	default:
		err = st.UpdateLoefte(ctx, *restore)
	}
	return snapL(c), snapL(restore), err
}

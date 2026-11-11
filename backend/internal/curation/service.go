// Package curation is the write side of ADR-003: token-protected, idempotent
// changes that each leave an append-only Ändringslogg entry in the same
// transaction, and can be undone by a further logged entry.
package curation

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"riksdagskollen/internal/curation/domain"
	"riksdagskollen/internal/curation/ports"
)

// Actor is recorded for every write: the holder of the write token.
const Actor = "bot"

type Service struct{ repo ports.Repository }

func NewService(repo ports.Repository) *Service { return &Service{repo: repo} }

func hashOf(kind string, v any) string {
	b, _ := json.Marshal(struct {
		K string
		V any
	}{kind, v})
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// replay returns the stored entry for key when this request was already
// handled, or an error when the key belongs to a different request.
func replay(ctx context.Context, st ports.Store, key, hash string) (*domain.Result, error) {
	if err := st.LockRequest(ctx, key); err != nil {
		return nil, err
	}
	e, err := st.EntryByKey(ctx, key)
	if err != nil || e == nil {
		return nil, err
	}
	if e.RequestHash != hash {
		return nil, domain.ErrKeyReuse
	}
	return &domain.Result{Changed: true, Replayed: true, Entry: e}, nil
}

func snap(v *domain.Valmanifest) json.RawMessage {
	if v == nil {
		return nil
	}
	b, _ := json.Marshal(v)
	return b
}

func unsnap(raw json.RawMessage) (*domain.Valmanifest, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}
	var v domain.Valmanifest
	if err := json.Unmarshal(raw, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

func same(a, b *domain.Valmanifest) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

// UpsertValmanifest creates or updates the Valmanifest keyed by
// (party, electionYear, title). key is the request's Idempotency-Key.
func (s *Service) UpsertValmanifest(ctx context.Context, key string, in domain.ValmanifestInput) (*domain.Result, error) {
	if err := validKey(key); err != nil {
		return nil, err
	}
	if err := in.Normalize(); err != nil {
		return nil, err
	}
	hash := hashOf("upsert-valmanifest", in)
	var out *domain.Result
	err := s.repo.InTx(ctx, func(st ports.Store) error {
		if r, err := replay(ctx, st, key, hash); err != nil || r != nil {
			out = r
			return err
		}
		cur, err := st.ValmanifestByKey(ctx, in.Party, in.ElectionYear, in.Title)
		if err != nil {
			return err
		}
		next := in.Snapshot()
		var old *domain.Valmanifest
		if cur != nil {
			next.ID = cur.ID
			if *cur == next {
				out = &domain.Result{Changed: false}
				return nil
			}
			old = cur
			if err := st.UpdateValmanifest(ctx, next); err != nil {
				return err
			}
		} else {
			ins, err := st.InsertValmanifest(ctx, next)
			if err != nil {
				return err
			}
			next = *ins
		}
		e, err := st.AppendEntry(ctx, &domain.Entry{
			Actor: Actor, OnBehalfOf: in.OnBehalfOf, ChatRef: in.ChatRef,
			Action: domain.ActionUpsert, Entity: domain.EntityValmanifest, EntityID: next.ID,
			OldValue: snap(old), NewValue: snap(&next), RequestKey: key, RequestHash: hash,
		})
		out = &domain.Result{Changed: true, Entry: e}
		return err
	})
	return out, err
}

// Undo reverses entry id, provided the entity still holds the value that
// entry wrote. The undo is itself logged. A since-changed value is ErrConflict.
func (s *Service) Undo(ctx context.Context, key string, id int64, p domain.Provenance) (*domain.Result, error) {
	if err := validKey(key); err != nil {
		return nil, err
	}
	if err := p.Validate(); err != nil {
		return nil, err
	}
	hash := hashOf("undo", struct {
		ID int64
		P  domain.Provenance
	}{id, p})
	var out *domain.Result
	err := s.repo.InTx(ctx, func(st ports.Store) error {
		if r, err := replay(ctx, st, key, hash); err != nil || r != nil {
			out = r
			return err
		}
		target, err := st.EntryByID(ctx, id)
		if err != nil {
			return err
		}
		if target == nil {
			return domain.ErrNotFound
		}
		var cur, restore json.RawMessage
		switch target.Entity {
		case domain.EntityValmanifest:
			cur, restore, err = undoValmanifest(ctx, st, target)
		case domain.EntityLoefte:
			cur, restore, err = undoLoefte(ctx, st, target)
		case domain.EntityKoppling:
			cur, restore, err = undoKoppling(ctx, st, target)
		default:
			return &domain.ValidationError{Msg: fmt.Sprintf("undo of %q entries is not supported", target.Entity)}
		}
		if err != nil {
			return err
		}
		e, err := st.AppendEntry(ctx, &domain.Entry{
			Actor: Actor, OnBehalfOf: p.OnBehalfOf, ChatRef: p.ChatRef,
			Action: domain.ActionUndo, Entity: target.Entity, EntityID: target.EntityID,
			OldValue: cur, NewValue: restore, Undoes: &target.ID,
			RequestKey: key, RequestHash: hash,
		})
		out = &domain.Result{Changed: true, Entry: e}
		return err
	})
	return out, err
}

// Entries lists the Ändringslogg newest first. beforeID > 0 pages backwards.
func (s *Service) Entries(ctx context.Context, limit int, beforeID int64) ([]*domain.Entry, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	entries, err := s.repo.ListEntries(ctx, limit, beforeID)
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		redactPublic(e)
	}
	return entries, nil
}

// redactPublic withholds the side of a Löfte change that was not public.
// The repository already drops entries where neither side was public, so a
// proposed Löfte (text, quote, source) never leaves through the log.
func redactPublic(e *domain.Entry) {
	if e.Entity != domain.EntityLoefte {
		return
	}
	hide := func(raw json.RawMessage) json.RawMessage {
		if len(raw) == 0 || string(raw) == "null" {
			return raw
		}
		var v struct {
			Visibility string `json:"visibility"`
		}
		if json.Unmarshal(raw, &v) == nil && domain.PublicVisibility(v.Visibility) {
			return raw
		}
		e.Redacted = true
		return nil
	}
	e.OldValue, e.NewValue = hide(e.OldValue), hide(e.NewValue)
}

func validKey(key string) error {
	if key == "" || len(key) > 200 {
		return &domain.ValidationError{Msg: "Idempotency-Key header is required (max 200 characters)"}
	}
	return nil
}

// undoValmanifest reverses a Valmanifest entry and returns the snapshots
// before and after the reversal.
func undoValmanifest(ctx context.Context, st ports.Store, target *domain.Entry) (cur, restored json.RawMessage, err error) {
	want, err := unsnap(target.NewValue)
	if err != nil {
		return nil, nil, err
	}
	restore, err := unsnap(target.OldValue)
	if err != nil {
		return nil, nil, err
	}
	c, err := st.ValmanifestByID(ctx, target.EntityID)
	if err != nil {
		return nil, nil, err
	}
	if !same(c, want) {
		return nil, nil, fmt.Errorf("%w: the value changed after entry %d", domain.ErrConflict, target.ID)
	}
	switch {
	case restore == nil:
		err = st.DeleteValmanifest(ctx, target.EntityID)
	case c == nil:
		_, err = st.InsertValmanifest(ctx, *restore)
	default:
		err = st.UpdateValmanifest(ctx, *restore)
	}
	return snap(c), snap(restore), err
}

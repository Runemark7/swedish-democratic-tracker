// Package loften serves Valmanifest and Löfte as public, read-only data.
package loften

import (
	"context"
	"errors"
	"slices"
	"sort"
	"time"

	"riksdagskollen/internal/loften/domain"
	"riksdagskollen/internal/loften/ports"
)

type Service struct{ repo ports.Repository }

func NewService(repo ports.Repository) *Service { return &Service{repo: repo} }

// List returns public Löften, optionally for one party ("" = all).
func (s *Service) List(ctx context.Context, party string) ([]*domain.Loefte, error) {
	return s.repo.ListPublic(ctx, party)
}

// Get returns a public Löfte, or nil when it does not exist or is not public.
func (s *Service) Get(ctx context.Context, id int) (*domain.Loefte, error) {
	return s.repo.GetPublic(ctx, id)
}

// Valmanifest lists Valmanifest, newest election first, optionally for one party.
func (s *Service) Valmanifest(ctx context.Context, party string) ([]*domain.Valmanifest, error) {
	return s.repo.ListValmanifest(ctx, party)
}

// StatusFilter selects public Löften by Löftesstatus slug: "for", "emot",
// "blandat", "inte-provat" or "arkiv" (archived Löften have no status).
var statusSlugs = map[string]domain.Status{
	"for": domain.StatusFor, "emot": domain.StatusEmot,
	"blandat": domain.StatusBlandat, "inte-provat": domain.StatusInteProvat,
}

// ErrBadStatus is returned for an unknown status filter.
var ErrBadStatus = errors.New("unknown status filter")

// Filter narrows a list of Löften. Zero fields mean "any".
type Filter struct {
	Party, Utskott, Status string
	// Betankande is "riksmöte:beteckning": only Löften with a Bekräftad
	// koppling to that betänkande.
	Betankande string
	// Mandatperiod is a mandate period code such as "2026-2030". It is
	// independent of Status: any period combines with any Löftesstatus.
	Mandatperiod string
}

// inPeriod keeps the Löften of one mandate period ("" = every period).
func inPeriod(ls []*domain.Loefte, period string) []*domain.Loefte {
	if period == "" {
		return ls
	}
	kept := ls[:0]
	for _, l := range ls {
		if l.MandatePeriod == period {
			kept = append(kept, l)
		}
	}
	return kept
}

func (s *Service) views(ctx context.Context, ls []*domain.Loefte, withKopplingar bool) ([]*domain.LoefteView, error) {
	ids := make([]int, len(ls))
	for i, l := range ls {
		ids[i] = l.ID
	}
	ks, err := s.repo.ConfirmedKopplingar(ctx, ids)
	if err != nil {
		return nil, err
	}
	wins, err := s.repo.Windows(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*domain.LoefteView, 0, len(ls))
	for _, l := range ls {
		confirmed := ks[l.ID]
		in := make([]domain.VoteInput, len(confirmed))
		byID := map[int]domain.ConfirmedKoppling{}
		for i, c := range confirmed {
			in[i] = c.VoteInput
			byID[c.KopplingID] = c
		}
		res := domain.StatusOf(l.Archived, wins[l.MandatePeriod], in)
		v := &domain.LoefteView{Loefte: *l}
		if res.Status != domain.StatusNone {
			st := res.Status
			v.Status = &st
			if !res.ChangedAt.IsZero() {
				at := res.ChangedAt.UTC().Format(time.RFC3339)
				v.StatusChangedAt = &at
				v.StatusReason = res.ChangeReason
			}
		}
		if withKopplingar {
			v.Kopplingar = []domain.KopplingView{}
			if l.Archived {
				// No status for archived Löften: their votes (if any) are facts, listed unevaluated.
				for _, c := range confirmed {
					ev := domain.EvaluatedVote{VoteInput: c.VoteInput, Effect: domain.EffectNone,
						Explanation: "Arkiverat Löfte: vi visar omröstningen men ingen Löftesstatus."}
					v.Kopplingar = append(v.Kopplingar, domain.ToView(ev, c))
				}
			}
			for _, ev := range res.Votes {
				v.Kopplingar = append(v.Kopplingar, domain.ToView(ev, byID[ev.KopplingID]))
			}
		}
		out = append(out, v)
	}
	return out, nil
}

// ListWithStatus lists public Löften with Löftesstatus, filtered by party,
// utskott, mandatperiod and status.
func (s *Service) ListWithStatus(ctx context.Context, f Filter) ([]*domain.LoefteView, error) {
	want, ok := statusSlugs[f.Status]
	if f.Status != "" && f.Status != "arkiv" && !ok {
		return nil, ErrBadStatus
	}
	ls, err := s.repo.ListPublic(ctx, f.Party)
	if err != nil {
		return nil, err
	}
	ls = inPeriod(ls, f.Mandatperiod)
	if f.Utskott != "" {
		kept := ls[:0]
		for _, l := range ls {
			if slices.Contains(l.Utskott, f.Utskott) {
				kept = append(kept, l)
			}
		}
		ls = kept
	}
	if f.Betankande != "" {
		refs, err := s.repo.LoefteForBetankanden(ctx, []string{f.Betankande})
		if err != nil {
			return nil, err
		}
		in := map[int]bool{}
		for _, r := range refs[f.Betankande] {
			in[r.ID] = true
		}
		kept := ls[:0]
		for _, l := range ls {
			if in[l.ID] {
				kept = append(kept, l)
			}
		}
		ls = kept
	}
	vs, err := s.views(ctx, ls, false)
	if err != nil {
		return nil, err
	}
	if f.Status == "" {
		return vs, nil
	}
	out := vs[:0]
	for _, v := range vs {
		switch {
		case f.Status == "arkiv" && v.Archived:
			out = append(out, v)
		case f.Status != "arkiv" && v.Status != nil && *v.Status == want:
			out = append(out, v)
		}
	}
	return out, nil
}

// GetWithStatus is one public Löfte with status and every Bekräftad koppling.
func (s *Service) GetWithStatus(ctx context.Context, id int) (*domain.LoefteView, error) {
	l, err := s.repo.GetPublic(ctx, id)
	if err != nil || l == nil {
		return nil, err
	}
	vs, err := s.views(ctx, []*domain.Loefte{l}, true)
	if err != nil {
		return nil, err
	}
	return vs[0], nil
}

// Koppling is one Bekräftad koppling by its stable id, nil if it is not
// Bekräftad or its Löfte is not public.
func (s *Service) Koppling(ctx context.Context, id int) (*domain.KopplingView, error) {
	c, err := s.repo.ConfirmedKopplingByID(ctx, id)
	if err != nil || c == nil {
		return nil, err
	}
	l, err := s.GetWithStatus(ctx, c.LoefteID)
	if err != nil || l == nil {
		return nil, err
	}
	for i := range l.Kopplingar {
		if l.Kopplingar[i].ID == id {
			return &l.Kopplingar[i], nil
		}
	}
	return nil, nil
}

// Summary counts public Löften per party and Löftesstatus, in party order,
// for one mandate period ("" = every period). A party without Löften in the
// period has no row.
func (s *Service) Summary(ctx context.Context, period string) ([]domain.PartySummary, error) {
	ls, err := s.repo.ListPublic(ctx, "")
	if err != nil {
		return nil, err
	}
	ls = inPeriod(ls, period)
	vs, err := s.views(ctx, ls, false)
	if err != nil {
		return nil, err
	}
	idx := map[string]*domain.PartySummary{}
	var order []string
	for _, v := range vs {
		p := idx[v.Party]
		if p == nil {
			p = &domain.PartySummary{Party: v.Party}
			idx[v.Party] = p
			order = append(order, v.Party)
		}
		p.Total++
		switch {
		case v.Status == nil:
			p.Arkiv++
		case *v.Status == domain.StatusFor:
			p.For++
		case *v.Status == domain.StatusEmot:
			p.Emot++
		case *v.Status == domain.StatusBlandat:
			p.Blandat++
		default:
			p.InteProvat++
		}
	}
	sort.Strings(order)
	out := make([]domain.PartySummary, 0, len(order))
	for _, p := range order {
		out = append(out, *idx[p])
	}
	return out, nil
}

// ForBetankanden returns, per "riksmöte:beteckning", the Löften a decision
// touches. Every key is present, with an empty list when none.
func (s *Service) ForBetankanden(ctx context.Context, keys []string) (map[string][]domain.LoefteRef, error) {
	got, err := s.repo.LoefteForBetankanden(ctx, keys)
	if err != nil {
		return nil, err
	}
	for _, k := range keys {
		if got[k] == nil {
			got[k] = []domain.LoefteRef{}
		}
	}
	return got, nil
}

// PartyTiles summarises Löftesstatus for all eight parties, always in
// alphabetical order of party code and with zeros for a party without Löften.
func (s *Service) PartyTiles(ctx context.Context) ([]domain.PartySummary, error) {
	sum, err := s.Summary(ctx, "")
	if err != nil {
		return nil, err
	}
	by := map[string]domain.PartySummary{}
	for _, p := range sum {
		by[p.Party] = p
	}
	out := make([]domain.PartySummary, 0, len(domain.Parties))
	for _, code := range domain.Parties {
		p := by[code]
		p.Party = code
		out = append(out, p)
	}
	return out, nil
}

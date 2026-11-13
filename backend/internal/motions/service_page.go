package motions

import (
	"context"
	"errors"
	"log/slog"
	"sort"

	"riksdagskollen/internal/motions/domain"
	"riksdagskollen/internal/motions/ports"
	votes "riksdagskollen/internal/votes/domain"
	votesports "riksdagskollen/internal/votes/ports"
)

// AvvikareView is a member who voted against their party's Partilinje.
type AvvikareView struct {
	PoliticianID, Name, Party, Result string
	Partilinje                        votes.Partilinje
}

// PartyLine is a party's tally and Partilinje in one votering.
type PartyLine struct {
	Party                        string
	Ja, Nej, Avstar, Franvarande int
	Partilinje                   votes.Partilinje
}

// PunktView is one förslagspunkt on the betänkande page.
type PunktView struct {
	domain.Forslagspunkt
	VoteringStored          bool
	Parties                 []PartyLine
	Avvikare                []AvvikareView
	Reservations, Yttranden []domain.Statement
	Motions                 []domain.PunktMotion
}

// BetankandePage is the canonical page of one riksmöte + beteckning.
type BetankandePage struct {
	Riksmote, Beteckning, Title, Organ, DokID string
	// Held is true when the betänkande's förslagspunkter are stored; false
	// means the punkter come from the voteringar alone.
	Held                                                  bool
	HTMLRead, HasReasoningSection, HasReservationsSection bool
	// Status, DecisionDate and Decided come from Riksdagen's listing read
	// live. Decided is nil when that read failed or found nothing.
	Status, DecisionDate string
	Decided              *bool
	Punkter              []PunktView
}

// ErrBetankandeNotFound: neither stored, voted nor listed by Riksdagen.
var ErrBetankandeNotFound = errors.New("betänkande not found")

// Betankande builds the page. The decision state is Riksdagen's: its listing
// row (beslutsdag, beslutad, status) through RiksdagDocument.Decided, never a
// timestamp of ours.
func (s *Service) Betankande(ctx context.Context, rm, bet string) (*BetankandePage, error) {
	// Every row is keyed on the beteckning as Riksdagen spells it (SfU18,
	// FöU2); a URL typed sfu18 or FÖU2 means the same betänkande.
	bet, err := s.repo.CanonicalBeteckning(ctx, rm, bet)
	if err != nil {
		return nil, err
	}
	page := &BetankandePage{Riksmote: rm, Beteckning: bet, Punkter: []PunktView{}}

	ref, listed, lerr := s.rd.FindBetankande(ctx, rm, bet)
	if lerr != nil {
		slog.Warn("betankande page: listing not read", "rm", rm, "bet", bet, "error", lerr)
	}
	if listed {
		doc := votesports.RiksdagDocument{DecisionDate: ref.DecisionDate, Status: ref.Status, ChamberDecided: ref.ChamberDecided}
		d := doc.Decided(s.clock())
		page.Decided, page.Status, page.DecisionDate = &d, ref.Status, ref.DecisionDate
		page.Title, page.DokID = ref.Title, ref.DokID
		if ref.Beteckning != "" {
			page.Beteckning = ref.Beteckning
		}
	}

	stored, err := s.repo.GetBetankande(ctx, rm, bet)
	switch {
	case err == nil:
		page.Held = true
		page.Title, page.Organ, page.DokID = stored.Title, stored.Organ, stored.DokID
		page.HTMLRead, page.HasReasoningSection, page.HasReservationsSection = stored.HTMLRead, stored.HasReasoningSection, stored.HasReservationsSection
	case !errors.Is(err, ports.ErrNotFound):
		return nil, err
	}

	var pts []domain.Forslagspunkt
	if page.Held {
		if pts, err = s.repo.Punkter(ctx, rm, bet); err != nil {
			return nil, err
		}
	} else {
		vps, err := s.repo.VoteringPunkter(ctx, rm, bet)
		if err != nil {
			return nil, err
		}
		for _, v := range vps {
			if page.Title == "" {
				page.Title = v.Title
			}
			pts = append(pts, domain.Forslagspunkt{Number: v.Punkt, Beslutstyp: "röstning", VoteringID: v.VoteringID})
		}
		if len(pts) == 0 && !listed {
			if lerr != nil {
				return nil, lerr
			}
			return nil, ErrBetankandeNotFound
		}
	}

	stmts, err := s.repo.Statements(ctx, rm, bet)
	if err != nil {
		return nil, err
	}
	motions, err := s.repo.PunktMotions(ctx, rm, bet)
	if err != nil {
		return nil, err
	}
	for _, p := range pts {
		v := PunktView{Forslagspunkt: p, Motions: motions[p.Number]}
		for _, st := range stmts {
			if st.Punkt != p.Number {
				continue
			}
			if st.Kind == domain.KindReservation {
				v.Reservations = append(v.Reservations, st)
			} else {
				v.Yttranden = append(v.Yttranden, st)
			}
		}
		if p.VoteringID != "" {
			if err := s.fillVotes(ctx, &v); err != nil {
				return nil, err
			}
		}
		page.Punkter = append(page.Punkter, v)
	}
	return page, nil
}

func (s *Service) fillVotes(ctx context.Context, v *PunktView) error {
	bs, err := s.repo.Ballots(ctx, v.VoteringID)
	if err != nil {
		return err
	}
	if len(bs) == 0 {
		return nil
	}
	v.VoteringStored = true
	names := map[string]string{}
	vv := make([]*votes.Vote, 0, len(bs))
	for _, b := range bs {
		names[b.PoliticianID] = b.Name
		vv = append(vv, &votes.Vote{VoteringID: v.VoteringID, PoliticianID: b.PoliticianID, Party: b.Party, VoteResult: votes.VoteResult(b.Result)})
	}
	for party, t := range votes.TallyByParty(vv) {
		v.Parties = append(v.Parties, PartyLine{Party: party, Ja: t.Ja, Nej: t.Nej, Avstar: t.Avstar, Franvarande: t.Franvarande, Partilinje: votes.PartilinjeFor(party, t)})
	}
	sort.Slice(v.Parties, func(i, j int) bool { return v.Parties[i].Party < v.Parties[j].Party })
	for _, a := range votes.Avvikare(vv) {
		v.Avvikare = append(v.Avvikare, AvvikareView{PoliticianID: a.PoliticianID, Name: names[a.PoliticianID], Party: a.Party, Result: string(a.VoteResult), Partilinje: a.Partilinje})
	}
	return nil
}

// Resolve backs the redirects: which riksmöte a beteckning (and punkt) lives in.
func (s *Service) Resolve(ctx context.Context, bet, punkt, rm string) (string, bool, error) {
	return s.repo.ResolveBetankande(ctx, bet, punkt, rm)
}

// CanonicalBeteckning returns the beteckning as it is stored for the riksmöte,
// matched without regard to case; unchanged when nothing is stored.
func (s *Service) CanonicalBeteckning(ctx context.Context, rm, bet string) (string, error) {
	return s.repo.CanonicalBeteckning(ctx, rm, bet)
}

// DebateLimit caps the speeches returned for one debate.
const DebateLimit = 300

// Debate reads the speeches linked to a betänkande, by its dok_id.
func (s *Service) Debate(ctx context.Context, betDokID string) (*domain.Debate, error) {
	return s.repo.Debate(ctx, betDokID, DebateLimit)
}

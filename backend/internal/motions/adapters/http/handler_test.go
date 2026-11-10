package http

import (
	"testing"

	"riksdagskollen/internal/motions/domain"
)

func TestToResponseAggregatesYrkandenIntoMotionOutcome(t *testing.T) {
	m := &domain.Motion{DokID: "X", Subtyp: "Kommittémotion", Yrkanden: []domain.Yrkande{
		{Number: 1, ChamberDecision: "Bifall"},
		{Number: 2, ChamberDecision: "Avslag"},
		{Number: 3, ChamberDecision: "Avslag"},
		{Number: 4},
	}}
	link := func(n int) domain.YrkandeDecision {
		return domain.YrkandeDecision{YrkandeNumber: n, Riksmote: "2021/22", Betankande: "SkU16",
			Punkt: domain.Forslagspunkt{Number: "1", Beslutstyp: "acklamation"}}
	}
	r := toResponse(m, []domain.YrkandeDecision{link(1), link(2), link(3)})
	if r.Yrkanden[3].Outcome != domain.OutcomeEjAvgjord {
		t.Errorf("unlinked yrkande = %q", r.Yrkanden[3].Outcome)
	}
	if r.Outcome.Word != domain.OutcomeEjAvgjord || r.Outcome.Total != 4 {
		t.Errorf("one undecided yrkande: %+v", r.Outcome)
	}
	r = toResponse(&domain.Motion{DokID: "X", Yrkanden: m.Yrkanden[:3]}, []domain.YrkandeDecision{link(1), link(2), link(3)})
	if r.Outcome.Word != domain.OutcomeDelvis || r.Outcome.Counts[domain.OutcomeBifallen] != 1 || r.Outcome.Counts[domain.OutcomeAvslagen] != 2 {
		t.Errorf("mixed: %+v", r.Outcome)
	}
}

func TestToDecisionShowsRiksdagensOwnWordForUnknownBeslutstyp(t *testing.T) {
	d := toDecision(domain.YrkandeDecision{Punkt: domain.Forslagspunkt{Beslutstyp: "Bordläggning"}}, "Återremitterad")
	if d.ModeLabel != "Bordläggning" || d.Outcome != "Återremitterad" {
		t.Errorf("%+v", d)
	}
}

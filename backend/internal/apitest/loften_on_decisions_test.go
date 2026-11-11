package apitest

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Seam tests for Löften on decisions, motions and the front page (issue #162).
// Counts come from Bekräftade kopplingar on public Löften only.

func addRawKoppling(t *testing.T, s *kSeed, loefte, votering int64, state string) {
	t.Helper()
	if _, err := s.h.DB.Exec(context.Background(),
		`INSERT INTO kopplingar (loefte_id, votering_ref, direction, state, confidence, threshold, reasoning, origin)
		 VALUES ($1,$2,'Ja',$3,0.9,0.6,'test','test')`, loefte, votering, state); err != nil {
		t.Fatal(err)
	}
}

type decisionsSeed struct {
	*kSeed
	a, b int64
}

// SkU2 (votering K-NEW) is a decided betänkande in the listing. Two public
// Löften have a Bekräftad koppling to it; a föreslagen koppling, a proposed
// Löfte and an avvisad koppling must not count.
func startDecisions(t *testing.T) *decisionsSeed {
	t.Helper()
	s := startK(t)
	ctx := context.Background()
	if _, err := s.h.DB.Exec(ctx,
		`INSERT INTO beslut_listing (riksmote, beteckning, dok_id, title, organ, status, beslutsdag, beslutad, sort_date)
		 VALUES ('2026/27','SkU2','D2','Skatt på kapital','SkU','Avslutat','2026-10-20',true,'2026-10-20')
		 ON CONFLICT DO NOTHING`); err != nil {
		t.Fatal(err)
	}
	d := &decisionsSeed{kSeed: s}
	d.a = s.loefteSkU
	d.b = seedLoefte(t, s.h, "M", "Lägre kapitalskatt", "approved", "SkU")
	addRawKoppling(t, s, d.a, s.vNew, "bekraftad")
	addRawKoppling(t, s, d.b, s.vNew, "bekraftad")
	addRawKoppling(t, s, s.loefteAU, s.vNew, "foreslagen")
	addRawKoppling(t, s, s.loefteProposed, s.vNew, "bekraftad") // Löfte not public
	c := seedLoefte(t, s.h, "V", "Avvisad", "approved", "SkU")
	addRawKoppling(t, s, c, s.vNew, "avvisad")
	return d
}

func TestBetankandeAndFrontPageCountOnlyBekraftadeKopplingar(t *testing.T) {
	d := startDecisions(t)

	fp := decode[struct {
		Decided []struct {
			Beteckning  string `json:"beteckning"`
			LoefteCount int    `json:"loefteCount"`
		} `json:"decided"`
		PartyTiles []struct {
			Party      string `json:"party"`
			For        int    `json:"for"`
			InteProvat int    `json:"inteProvat"`
			Total      int    `json:"total"`
		} `json:"partyTiles"`
	}](t, publicGet(t, d.h, "/api/frontpage", 200))
	found := false
	for _, r := range fp.Decided {
		if r.Beteckning == "SkU2" {
			found = true
			if r.LoefteCount != 2 {
				t.Errorf("front page SkU2 loefteCount = %d, want 2", r.LoefteCount)
			}
		}
	}
	if !found {
		t.Fatalf("SkU2 not in front page feed: %+v", fp.Decided)
	}

}

// Recorded chain (fixtures/decision-chain): motion 2021/22:3638 y9 was decided
// in SkU16 punkt 6, by a votering we hold. Two public Löften have a Bekräftad
// koppling to that votering, a third only a föreslagen one.
func TestBetankandeMotionAndLinkedListShowTheSameLoefter(t *testing.T) {
	h := startChain(t, chainDir(t, "decision-chain"), "{2017/18,2018/19,2021/22}")
	runChain(h)
	ctx := context.Background()
	var vref int64
	if err := h.DB.QueryRow(ctx, `SELECT v.id FROM voteringar v JOIN forslagspunkter f ON upper(f.votering_id)=upper(v.votering_id)
		WHERE f.riksmote='2021/22' AND f.beteckning='SkU16' AND f.punkt='6'`).Scan(&vref); err != nil {
		t.Fatal(err)
	}
	var a, b, c int64
	for i, p := range []*int64{&a, &b, &c} {
		party := []string{"M", "S", "V"}[i]
		if err := h.DB.QueryRow(ctx,
			`INSERT INTO loften (party, text, quote, source_url, utskott, mandate_period, visibility)
			 VALUES ($1,'Test '||$1,'citat','https://example.org','{SkU}','2026-2030','approved') RETURNING id`, party).Scan(p); err != nil {
			t.Fatal(err)
		}
	}
	for _, k := range []struct {
		id    int64
		state string
	}{{a, "bekraftad"}, {b, "bekraftad"}, {c, "foreslagen"}} {
		if _, err := h.DB.Exec(ctx, `INSERT INTO kopplingar (loefte_id, votering_ref, direction, state, confidence, threshold, reasoning, origin)
			VALUES ($1,$2,'Ja',$3,0.9,0.6,'t','t')`, k.id, vref, k.state); err != nil {
			t.Fatal(err)
		}
	}

	bet := decode[struct {
		LoefteCount int `json:"loefteCount"`
		Loften      []struct {
			ID    int    `json:"id"`
			Party string `json:"party"`
		} `json:"loften"`
	}](t, publicGet(t, h, "/api/betankanden/2021%2F22/SkU16", 200))
	if bet.LoefteCount != 2 || len(bet.Loften) != 2 {
		t.Fatalf("betänkande Löften = %+v", bet)
	}
	// "Berör N löften" links to exactly those Löften.
	list := decode[[]struct{ ID int }](t, publicGet(t, h, "/api/loften?betankande=2021%2F22:SkU16", 200))
	if len(list) != 2 {
		t.Fatalf("filtered list = %+v", list)
	}
	ids := map[int]bool{list[0].ID: true, list[1].ID: true}
	for _, l := range bet.Loften {
		if !ids[l.ID] {
			t.Errorf("Löfte %d in the betänkande but not in the linked list", l.ID)
		}
	}
	publicGet(t, h, "/api/loften?betankande=nonsense", 400)
	if other := decode[[]struct{ ID int }](t, publicGet(t, h, "/api/loften?betankande=2017%2F18:SkU11", 200)); len(other) != 0 {
		t.Errorf("another betänkande lists %+v", other)
	}

	mot := decode[struct {
		Yrkanden []struct {
			Decisions []struct {
				LoefteCount int `json:"loefteCount"`
			} `json:"decisions"`
		} `json:"yrkanden"`
		Loften []struct{ ID int } `json:"loften"`
	}](t, publicGet(t, h, "/api/motions/2021%2F22/3638", 200))
	if len(mot.Loften) != 2 {
		t.Errorf("motion Löften = %+v", mot.Loften)
	}
	n := 0
	for _, y := range mot.Yrkanden {
		for _, d := range y.Decisions {
			if d.LoefteCount == 2 {
				n++
			}
		}
	}
	if n == 0 {
		t.Errorf("no decision of the motion carries the count: %+v", mot.Yrkanden)
	}
}

func TestPartyTilesAreAlphabeticalAndCountBekraftadeOnly(t *testing.T) {
	d := startDecisions(t)
	var codes []string
	tiles := decode[struct {
		PartyTiles []struct {
			Party string `json:"party"`
			Total int    `json:"total"`
			For   int    `json:"for"`
			Inte  int    `json:"inteProvat"`
		} `json:"partyTiles"`
	}](t, publicGet(t, d.h, "/api/frontpage", 200)).PartyTiles
	for _, p := range tiles {
		codes = append(codes, p.Party)
	}
	if got := strings.Join(codes, ","); got != "C,KD,L,M,MP,S,SD,V" {
		t.Fatalf("tiles = %s", got)
	}
	// Counts shift but the order does not, whatever the counts are.
	ballotsFor(t, d.h, d.vNew, "M", 9, 1, 0)
	tiles2 := decode[struct {
		PartyTiles []struct {
			Party string `json:"party"`
			For   int    `json:"for"`
			Inte  int    `json:"inteProvat"`
		} `json:"partyTiles"`
	}](t, publicGet(t, d.h, "/api/frontpage", 200)).PartyTiles
	for i, p := range tiles2 {
		if p.Party != codes[i] {
			t.Fatalf("order changed with counts: %v", tiles2)
		}
		if p.Party == "M" && p.For != 1 {
			t.Errorf("M för = %d, want 1", p.For)
		}
		// S has two public approved Löften; the föreslagen koppling gives no status.
		if p.Party == "S" && p.Inte != 2 {
			t.Errorf("S inteProvat = %d, want 2", p.Inte)
		}
	}
	// A request cannot reorder them.
	rec := d.h.Do(httptest.NewRequest(http.MethodGet, "/api/frontpage?party=V&sort=count", nil))
	if rec.Code != 200 {
		t.Fatal(rec.Code)
	}
}

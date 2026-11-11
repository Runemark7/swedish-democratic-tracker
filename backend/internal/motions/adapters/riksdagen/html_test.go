package riksdagen

import (
	"riksdagskollen/internal/motions/domain"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The HTML files are the recorded data.riksdagen.se/dokument/<dok_id>.html
// responses kept with the seam fixtures.
func recordedHTML(t *testing.T, name string) *domain.BetankandeHTML {
	t.Helper()
	f, err := os.Open(filepath.Join("..", "..", "..", "apitest", "fixtures", "html", name))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	h, err := ParseBetankandeHTML(f)
	if err != nil {
		t.Fatal(err)
	}
	return h
}

func findStatement(h *domain.BetankandeHTML, kind string, n int) *domain.Statement {
	for i := range h.Statements {
		if h.Statements[i].Kind == kind && h.Statements[i].Number == n {
			return &h.Statements[i]
		}
	}
	return nil
}

func TestParseSkU11_2017_18_ReservationsAndReasoning(t *testing.T) {
	h := recordedHTML(t, "H501SkU11.html")
	nres := 0
	for _, s := range h.Statements {
		if s.Kind == domain.KindReservation {
			nres++
		}
	}
	if nres != 9 {
		t.Fatalf("reservations = %d, want 9 (the document says \"nio reservationer\")", nres)
	}
	r8 := findStatement(h, domain.KindReservation, 8)
	if r8 == nil || r8.Punkt != "6" || r8.Rubrik != "Övriga frågor" || len(r8.Parties) != 1 || r8.Parties[0] != "M" {
		t.Fatalf("reservation 8 = %+v", r8)
	}
	if len(r8.Signers) != 4 || r8.Signers[0] != "Cecilia Widegren (M)" || r8.Signers[3] != "Lotta Finstorp (M)" {
		t.Errorf("signers = %q", r8.Signers)
	}
	if !strings.Contains(r8.Proposal, "2017/18:3293 av Maria Malmer Stenergard m.fl. (M)") {
		t.Errorf("proposal = %q", r8.Proposal)
	}
	if !strings.Contains(r8.Text, "Vi föreslår en utredning") || strings.Contains(r8.Text, "Larry Söder") {
		t.Errorf("text = %q", r8.Text)
	}
	r9 := findStatement(h, domain.KindReservation, 9)
	if r9 == nil || r9.Punkt != "6" || r9.Parties[0] != "KD" || len(r9.Signers) != 1 || r9.Signers[0] != "Larry Söder (KD)" {
		t.Fatalf("reservation 9 = %+v", r9)
	}
	if strings.Contains(r9.Text, "Förteckning över behandlade förslag") || strings.Contains(r9.Text, "2017/18:57 av") {
		t.Errorf("the last reservation ran into the bilaga: %q", r9.Text)
	}
	r5 := findStatement(h, domain.KindReservation, 5)
	if r5 == nil || strings.Join(r5.Parties, ",") != "M,C,L,KD" || r5.Punkt != "2" {
		t.Errorf("reservation 5 = %+v", r5)
	}
	got := h.ReasoningFor("Övriga frågor")
	if !strings.Contains(got, "Utskottet gör ingen annan bedömning om virtuella varor än Skatteverket.") {
		t.Errorf("reasoning for Övriga frågor = %q", got)
	}
	if strings.Contains(got, "Motionerna") || strings.Contains(got, "kommittémotion 2017/18:3293") {
		t.Errorf("reasoning includes the motion summaries that come before Utskottets ställningstagande")
	}
	if h.ReasoningFor("Rubrik som inte finns") != "" {
		t.Error("an unknown rubrik must have no reasoning")
	}
}

func TestParseSkU16_2021_22_Reservation12IsM(t *testing.T) {
	h := recordedHTML(t, "H901SkU16.html")
	r := findStatement(h, domain.KindReservation, 12)
	if r == nil || r.Punkt != "6" || len(r.Parties) != 1 || r.Parties[0] != "M" {
		t.Fatalf("reservation 12 = %+v", r)
	}
	if findStatement(h, domain.KindReservation, 14) == nil || findStatement(h, domain.KindReservation, 15) != nil {
		t.Error("expected exactly reservations 1-14")
	}
	if h.ReasoningFor("Övriga mervärdesskattefrågor") == "" {
		t.Error("no reasoning for Övriga mervärdesskattefrågor")
	}
}

func TestParseSkU15_2022_23(t *testing.T) {
	h := recordedHTML(t, "HA01SkU15.html")
	n := 0
	for _, s := range h.Statements {
		if s.Kind == domain.KindReservation {
			n++
			if s.Punkt == "" || len(s.Parties) == 0 || s.Text == "" || len(s.Signers) == 0 {
				t.Errorf("incomplete reservation %+v", s)
			}
		}
	}
	if n != 7 {
		t.Errorf("reservations = %d, want 7", n)
	}
	for _, rub := range []string{"Skattesatser m.m.", "Undantag från skattskyldighet", "Mervärdesskattepliktig verksamhet vid fastighetsuthyrning", "Övriga mervärdesskattefrågor"} {
		if h.ReasoningFor(rub) == "" {
			t.Errorf("no reasoning for %q", rub)
		}
	}
}

func TestParseSkU11_2018_19_ReservationOfTheVotedAlternative(t *testing.T) {
	h := recordedHTML(t, "H601SkU11.html")
	r := findStatement(h, domain.KindReservation, 2)
	if r == nil || r.Punkt != "1" || r.Parties[0] != "SD" {
		t.Fatalf("reservation 2 = %+v", r)
	}
}

func TestParseSarskiltYttrande(t *testing.T) {
	f, err := os.Open(filepath.Join("testdata", "H901SoU1-sarskilda-yttranden-excerpt.html"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	h, err := ParseBetankandeHTML(f)
	if err != nil {
		t.Fatal(err)
	}
	y := findStatement(h, domain.KindSarskiltYttrande, 1)
	if y == nil || y.Punkt != "1" || y.Rubrik != "Statens budget inom utgiftsområde 9" || len(y.Parties) != 1 || y.Parties[0] != "S" {
		t.Fatalf("särskilt yttrande 1 = %+v", y)
	}
	if len(y.Signers) != 5 || y.Signers[0] != "Ann-Christin Ahlberg (S)" || y.Text == "" {
		t.Errorf("signers/text = %q / %q", y.Signers, y.Text)
	}
	if findStatement(h, domain.KindSarskiltYttrande, 2) == nil {
		t.Error("expected a second särskilt yttrande")
	}
}

func TestParseUnrecognisedDocumentYieldsNothing(t *testing.T) {
	h, err := ParseBetankandeHTML(strings.NewReader("<p>Ett dokument utan avsnitt</p>"))
	if err != nil {
		t.Fatal(err)
	}
	if len(h.Statements) != 0 || h.ReasoningFor("Övriga frågor") != "" {
		t.Errorf("invented content: %+v", h)
	}
}

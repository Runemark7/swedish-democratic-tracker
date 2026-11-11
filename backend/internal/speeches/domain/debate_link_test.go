package domain_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"riksdagskollen/internal/speeches/domain"
)

// The speech lists are Riksdagen's /anforandelista responses recorded in
// apitest/fixtures/speech-debate (reduced to one debate each plus an
// unrelated row).
type recordedSpeech struct {
	DokDatum       string `json:"dok_datum"`
	Avsnittsrubrik string `json:"avsnittsrubrik"`
	RelDokID       string `json:"rel_dok_id"`
	Talare         string `json:"talare"`
}

func recorded(t *testing.T, file string) []recordedSpeech {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "..", "apitest", "fixtures", "speech-debate", file))
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		Body struct {
			Anforandelista struct {
				Anforande []recordedSpeech `json:"anforande"`
			} `json:"anforandelista"`
		} `json:"body"`
	}
	if err := json.Unmarshal(raw, &f); err != nil {
		t.Fatal(err)
	}
	return f.Body.Anforandelista.Anforande
}

func forLinking(t *testing.T, r recordedSpeech) domain.SpeechForLinking {
	t.Helper()
	d, err := time.Parse("2006-01-02", r.DokDatum)
	if err != nil {
		t.Fatal(err)
	}
	return domain.SpeechForLinking{RelatedDokID: r.RelDokID, TopicHeading: r.Avsnittsrubrik, Date: d}
}

// UBU15 2002/03 "Fristående skolor": title and debate date as listed by
// Riksdagen (apitest/fixtures and motions/adapters/riksdagen/testdata).
var ubu15 = domain.DebateCandidate{
	DokID: "GQ01UBU15", Title: "Fristående skolor", DebateDate: time.Date(2003, 5, 28, 0, 0, 0, 0, time.UTC),
}

func TestRelDokIDNamingABetankandeLinksWithoutAnyCandidate(t *testing.T) {
	n := 0
	for _, r := range recorded(t, "speeches-2021-22-MP.json") {
		if r.RelDokID != "H901MJU33" {
			continue
		}
		n++
		l, ok := domain.LinkSpeech(forLinking(t, r), nil)
		if !ok || l.BetankandeDokID != "H901MJU33" || l.Rule != domain.LinkRuleRelDokID {
			t.Fatalf("%+v → %+v ok=%v", r, l, ok)
		}
	}
	if n == 0 {
		t.Fatal("recording has no H901MJU33 speech")
	}
}

func TestSpeechWithoutRelDokIDLinksByTitleAndDebateDate2002(t *testing.T) {
	n := 0
	for _, file := range []string{"speeches-2002-03-S.json", "speeches-2002-03-V.json", "speeches-2002-03-M.json"} {
		for _, r := range recorded(t, file) {
			if r.Avsnittsrubrik != "Fristående skolor" && r.Avsnittsrubrik != "Fristående skolor " {
				continue
			}
			n++
			if r.RelDokID != "" {
				t.Fatalf("recording should lack rel_dok_id: %+v", r)
			}
			l, ok := domain.LinkSpeech(forLinking(t, r), []domain.DebateCandidate{ubu15})
			if !ok || l.BetankandeDokID != "GQ01UBU15" || l.Rule != domain.LinkRuleTitleDate {
				t.Fatalf("%+v → %+v ok=%v", r, l, ok)
			}
		}
	}
	if n == 0 {
		t.Fatal("recording has no 2002/03 debate rows")
	}
}

func TestSpeechMatchingNothingStaysUnlinked(t *testing.T) {
	rows := recorded(t, "speeches-2021-22-MP.json")
	var unrelated *recordedSpeech
	for i := range rows {
		if rows[i].RelDokID == "" {
			unrelated = &rows[i]
		}
	}
	if unrelated == nil {
		t.Fatal("recording has no unrelated row")
	}
	if l, ok := domain.LinkSpeech(forLinking(t, *unrelated), []domain.DebateCandidate{ubu15}); ok {
		t.Fatalf("linked an unrelated speech: %+v", l)
	}

	// Same heading on another day, and the right day under another heading.
	s := domain.SpeechForLinking{TopicHeading: "Fristående skolor", Date: time.Date(2003, 5, 29, 0, 0, 0, 0, time.UTC)}
	if _, ok := domain.LinkSpeech(s, []domain.DebateCandidate{ubu15}); ok {
		t.Error("linked on title alone")
	}
	s = domain.SpeechForLinking{TopicHeading: "Rättsväsendet", Date: ubu15.DebateDate}
	if _, ok := domain.LinkSpeech(s, []domain.DebateCandidate{ubu15}); ok {
		t.Error("linked on date alone")
	}
}

func TestRelDokIDOfAnInterpellationFallsThroughAndAmbiguityIsNotGuessed(t *testing.T) {
	s := domain.SpeechForLinking{RelatedDokID: "H910398", TopicHeading: "Fristående skolor", Date: ubu15.DebateDate}
	l, ok := domain.LinkSpeech(s, []domain.DebateCandidate{ubu15})
	if !ok || l.Rule != domain.LinkRuleTitleDate {
		t.Fatalf("interpellation id must not count as a betänkande: %+v ok=%v", l, ok)
	}
	twin := ubu15
	twin.DokID = "GQ01UBU99"
	if l, ok := domain.LinkSpeech(s, []domain.DebateCandidate{ubu15, twin}); ok {
		t.Fatalf("two betänkanden share title and day, yet linked %+v", l)
	}
}

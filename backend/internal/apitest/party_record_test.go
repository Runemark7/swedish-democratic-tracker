package apitest

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

type partyRecordBody struct {
	Party     string `json:"party"`
	Decisions []struct {
		Riksmote   string `json:"riksmote"`
		Beteckning string `json:"beteckning"`
		Decided    bool   `json:"decided"`
		Partilinje string `json:"partilinje"`
		SourceURL  string `json:"sourceUrl"`
	} `json:"decisions"`
	Motions []listRow `json:"motions"`
}

func getPartyRecord(t *testing.T, h *Harness, party string) (partyRecordBody, int) {
	t.Helper()
	rec := h.Do(httptest.NewRequest(http.MethodGet, "/api/parties/"+party+"/record", nil))
	var b partyRecordBody
	_ = json.Unmarshal(rec.Body.Bytes(), &b)
	return b, rec.Code
}

// M's Kommittémotioner on the party page include 2017/18:3293, Avslagen.
func TestPartyRecordListsMsKommittemotionerWithOutcomes(t *testing.T) {
	h := listHarness(t)
	b, code := getPartyRecord(t, h, "M")
	if code != 200 {
		t.Fatalf("status %d", code)
	}
	found := false
	for _, m := range b.Motions {
		if m.Kind != "Kommittémotion" {
			t.Errorf("%s:%s is a %s", m.Riksmote, m.Beteckning, m.Kind)
		}
		if m.Riksmote == "2017/18" && m.Beteckning == "3293" {
			found = true
			if m.Outcome.Word != "Avslagen" {
				t.Errorf("3293 outcome = %q", m.Outcome.Word)
			}
		}
	}
	if !found {
		t.Fatalf("2017/18:3293 missing from %d motions", len(b.Motions))
	}
	// 4202 is a Flerpartimotion, not a Kommittémotion.
	for _, m := range b.Motions {
		if m.Beteckning == "4202" {
			t.Error("4202 listed")
		}
	}
}

// Recent decisions carry the party's own Partilinje and link to the source.
func TestPartyRecordDecisionsCarryThePartysPartilinje(t *testing.T) {
	h := startBeslut(t)
	want := map[string]string{"S": "Nej", "M": "Ja", "V": "Nej"}
	for party, pl := range want {
		b, code := getPartyRecord(t, h, party)
		if code != 200 {
			t.Fatalf("status %d", code)
		}
		var got string
		for _, d := range b.Decisions {
			if !d.Decided {
				t.Errorf("%s: undecided row %s", party, d.Beteckning)
			}
			if d.Beteckning == "UbU30" {
				got = d.Partilinje
				if d.SourceURL == "" {
					t.Error("no sourceUrl")
				}
			}
		}
		if got != pl {
			t.Errorf("%s UbU30 partilinje = %q, want %q", party, got, pl)
		}
	}
}

func TestPartyRecordRejectsABadPartyCode(t *testing.T) {
	h := listHarness(t)
	if _, code := getPartyRecord(t, h, "M;DROP"); code != 400 {
		t.Errorf("status %d", code)
	}
}

// An unknown party code answers as before (200, nothing listed) and, since
// ballots have no party index, must not probe every ballot to find that out.
func TestPartyRecordForAnUnknownPartyIsEmpty(t *testing.T) {
	h := listHarness(t)
	b, code := getPartyRecord(t, h, "ZZ")
	if code != 200 || b.Party != "ZZ" || len(b.Decisions) != 0 || len(b.Motions) != 0 {
		t.Fatalf("status %d, body %+v", code, b)
	}
}

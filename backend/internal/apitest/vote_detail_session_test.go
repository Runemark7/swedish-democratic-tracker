package apitest

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

// TestVoteDetailNeverMixesRiksmoten covers gap share-card-wrong-session. The
// share card (/votes/:beteckning/:punkt/dela) renders GET /api/votes/{bet}/{punkt}
// as-is: its "Riksdagsomröstning · <riksmöte>" line is the response's session
// and its party grid is the response's partyBreakdown. Beteckning numbering
// restarts every riksmöte, so the same beteckning + punkt exists in several.
//
// The recorded votering is 2025/26 UbU30 punkt 2. A second votering with the
// same beteckning and punkt is seeded in 2005/06, with the party codes of that
// time. Each riksmöte must come back alone when asked for, and an address
// without a riksmöte must answer with one riksmöte only (the latest), never a
// blend of both.
func TestVoteDetailNeverMixesRiksmoten(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	dir := filepath.Join(filepath.Dir(file), "fixtures", "one-votering")
	h := Start(t, dir, WithNow(time.Date(2026, 8, 13, 12, 0, 0, 0, time.UTC)))
	ctx := context.Background()
	h.RunWorker(ctx, "fetch-politicians")
	h.RunWorker(ctx, "fetch-votes")

	const oldSession, newSession = "2005/06", "2025/26"
	const oldTitle = "Äldre betänkande med samma beteckning"
	var ref int
	if err := h.DB.QueryRow(ctx,
		`INSERT INTO voteringar (votering_id, beteckning, forslagspunkt, session, document_title, system_datum)
		 VALUES ('SEED-OLD-UBU30-2', $1, $2, $3, $4, '2006-03-01 12:00:00+00') RETURNING id`,
		recordedBeteckning, recordedPunkt, oldSession, oldTitle).Scan(&ref); err != nil {
		t.Fatalf("seed old votering: %v", err)
	}
	// Three members of the recording, re-used as stand-ins under the party
	// codes the 2005/06 record carries.
	if _, err := h.DB.Exec(ctx,
		`INSERT INTO ballots (votering_ref, politician_id, party, vote_result)
		 SELECT $1, politician_id, 'fp', 'Nej' FROM ballots
		 WHERE votering_ref = (SELECT id FROM voteringar WHERE votering_id = $2)
		 ORDER BY politician_id LIMIT 3`, ref, recordedVoteringID); err != nil {
		t.Fatalf("seed old ballots: %v", err)
	}

	type detail struct {
		Session        string `json:"session"`
		DocumentTitle  string `json:"documentTitle"`
		Status         string `json:"status"`
		PartyBreakdown []struct {
			Party string `json:"party"`
			Ja    int    `json:"jaCount"`
			Nej   int    `json:"nejCount"`
		} `json:"partyBreakdown"`
	}
	get := func(session string) detail {
		t.Helper()
		target := "/api/votes/" + recordedBeteckning + "/" + recordedPunkt
		if session != "" {
			target += "?session=" + url.QueryEscape(session)
		}
		rec := h.Do(httptest.NewRequest(http.MethodGet, target, nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("GET %s status = %d, body = %s", target, rec.Code, rec.Body.String())
		}
		var d detail
		if err := json.Unmarshal(rec.Body.Bytes(), &d); err != nil {
			t.Fatalf("decode: %v body=%s", err, rec.Body.String())
		}
		return d
	}
	parties := func(d detail) map[string]bool {
		m := map[string]bool{}
		for _, p := range d.PartyBreakdown {
			m[p.Party] = true
		}
		return m
	}

	t.Run("2005/06 asked for by riksmöte", func(t *testing.T) {
		d := get(oldSession)
		if d.Session != oldSession {
			t.Errorf("session = %q, want %q", d.Session, oldSession)
		}
		if d.DocumentTitle != oldTitle {
			t.Errorf("documentTitle = %q, want %q", d.DocumentTitle, oldTitle)
		}
		if p := parties(d); len(p) != 1 || !p["fp"] {
			t.Errorf("parties = %v, want only fp", p)
		}
		if d.Status != "Avslag" {
			t.Errorf("status = %q, want Avslag", d.Status)
		}
	})

	t.Run("2025/26 asked for by riksmöte", func(t *testing.T) {
		d := get(newSession)
		if d.Session != newSession {
			t.Errorf("session = %q, want %q", d.Session, newSession)
		}
		if d.DocumentTitle == oldTitle {
			t.Errorf("documentTitle is the 2005/06 title")
		}
		p := parties(d)
		if p["fp"] {
			t.Errorf("2025/26 card carries the 2005/06 party fp: %v", p)
		}
		if !p["S"] || !p["M"] {
			t.Errorf("2025/26 parties missing: %v", p)
		}
	})

	t.Run("address without riksmöte answers with one riksmöte only", func(t *testing.T) {
		d, scoped := get(""), get(newSession)
		if d.Session != newSession {
			t.Errorf("session = %q, want the latest, %q", d.Session, newSession)
		}
		if d.DocumentTitle == oldTitle {
			t.Errorf("documentTitle is the 2005/06 title")
		}
		if p := parties(d); p["fp"] {
			t.Errorf("unscoped card blends in the 2005/06 party fp: %v", p)
		}
		if len(d.PartyBreakdown) != len(scoped.PartyBreakdown) {
			t.Fatalf("unscoped has %d parties, 2025/26 has %d", len(d.PartyBreakdown), len(scoped.PartyBreakdown))
		}
		for i, p := range d.PartyBreakdown {
			if p != scoped.PartyBreakdown[i] {
				t.Errorf("party %s tally %+v differs from 2025/26's %+v", p.Party, p, scoped.PartyBreakdown[i])
			}
		}
	})
}

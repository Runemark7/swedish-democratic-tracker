package apitest

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// Gap raw-html-entities-in-proposal-text. Riksdagen serves a förslagspunkt's
// text with HTML entities and <br /> ("Riksdagen<br />a) godk&auml;nner
// &auml;ndrade ramar f&ouml;r utgiftsomr&aring;den ... 19&#160;571&#160;000&#160;000
// kronor"). fixtures/forslagstext-entities is 2025/26:FiU47 recorded on
// 2026-11-13: its listing row, its dokumentstatus (reduced, see the note) and
// its HTML. The page API must give plain text with real å, ä and ö.
//
// The daily worker reads a betänkande once a stored motion points at it, so
// the test stores the key of motion 2025/26:4199 (dok_id HD024199), which
// Riksdagen lists as behandlas_i 2025/26:FiU47, without its text.
func TestForslagstextIsServedAsPlainTextWithSwedishLetters(t *testing.T) {
	h := Start(t, chainDir(t, "forslagstext-entities"))
	ctx := context.Background()
	if _, err := h.DB.Exec(ctx, `
		INSERT INTO motions (dok_id, riksmote, beteckning, systemdatum, behandlas_i)
		VALUES ('HD024199', '2025/26', '4199', '2026-06-17 14:16:28', '2025/26:FiU47')`); err != nil {
		t.Fatal(err)
	}
	if _, err := h.DB.Exec(ctx, `
		INSERT INTO motion_yrkanden (dok_id, nummer, behandlas_i) VALUES ('HD024199', 2, '2025/26:FiU47')`); err != nil {
		t.Fatal(err)
	}
	h.App.SetNow(func() time.Time { return time.Date(2026, 6, 18, 12, 0, 0, 0, time.UTC) })
	h.RunWorker(ctx, "fetch-betankanden")
	rec := h.Do(httptest.NewRequest(http.MethodGet, "/api/betankanden/2025%2F26/FiU47", nil))
	if rec.Code != 200 {
		t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
	}
	var b struct {
		Punkter []struct {
			Number string `json:"number"`
			Text   string `json:"text"`
		} `json:"punkter"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &b); err != nil {
		t.Fatal(err)
	}
	if len(b.Punkter) == 0 {
		t.Fatal("no förslagspunkter")
	}
	var all string
	for _, p := range b.Punkter {
		all += p.Text + "\n"
		for _, raw := range []string{"&ouml;", "&auml;", "&aring;", "&Aring;", "&#", "&amp;", "<br", "<BR", "\u00ad"} {
			if strings.Contains(p.Text, raw) {
				t.Errorf("punkt %s text holds raw %q: %.200q", p.Number, raw, p.Text)
			}
		}
	}
	for _, want := range []string{
		"Riksdagen\na) godkänner ändrade ramar för utgiftsområden och anvisar ändrade och nya anslag enligt utskottets förslag i bilaga 4,\nb) bemyndigar",
		"högst 19\u00a0571\u00a0000\u00a0000 kronor",
		"Transportstyrelsen",
		"1:17 Åtgärder för landsbygdens miljö",
	} {
		if !strings.Contains(all, want) {
			t.Errorf("text lacks %q", want)
		}
	}
}

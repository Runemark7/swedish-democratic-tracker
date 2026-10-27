package workers_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	ingPorts "riksdagskollen/internal/ingestion/ports"
	"riksdagskollen/internal/ingestion/workers"
	"riksdagskollen/internal/speeches"
	speechdomain "riksdagskollen/internal/speeches/domain"
	speechports "riksdagskollen/internal/speeches/ports"
)

type fakeSpeechClient struct {
	mu       sync.Mutex
	sessions map[string]bool
	rows     []*speechdomain.Speech
	err      error
}

func (f *fakeSpeechClient) FetchSpeeches(_ context.Context, flt speechports.FetchSpeechesFilter) ([]*speechdomain.Speech, error) {
	f.mu.Lock()
	if f.sessions == nil {
		f.sessions = map[string]bool{}
	}
	f.sessions[flt.Session] = true
	f.mu.Unlock()
	if f.err != nil {
		return nil, f.err
	}
	return f.rows, nil
}

func (f *fakeSpeechClient) FetchSpeechText(context.Context, string, string) (string, error) {
	return "", nil
}

// fakeSpeechRepo satisfies ports.SpeechRepository. Only UpsertMany is exercised
// by the worker; the rest are read paths it never touches.
type fakeSpeechRepo struct{ upserted int }

func (r *fakeSpeechRepo) UpsertMany(_ context.Context, ss []*speechdomain.Speech) error {
	r.upserted += len(ss)
	return nil
}
func (r *fakeSpeechRepo) GetByID(context.Context, int) (*speechdomain.Speech, error) { return nil, nil }
func (r *fakeSpeechRepo) ListByPolitician(context.Context, string) ([]*speechdomain.Speech, error) {
	return nil, nil
}
func (r *fakeSpeechRepo) ListRecent(context.Context, int) ([]*speechdomain.Speech, error) {
	return nil, nil
}
func (r *fakeSpeechRepo) ListUnprocessed(context.Context, int) ([]*speechdomain.Speech, error) {
	return nil, nil
}
func (r *fakeSpeechRepo) MarkProcessed(context.Context, int) error { return nil }
func (r *fakeSpeechRepo) ListMissingText(context.Context, int) ([]*speechdomain.Speech, error) {
	return nil, nil
}
func (r *fakeSpeechRepo) UpdateText(context.Context, int, string) error { return nil }
func (r *fakeSpeechRepo) ListByDocument(context.Context, string) ([]*speechdomain.Speech, error) {
	return nil, nil
}
func (r *fakeSpeechRepo) ListByParty(context.Context, string, int) ([]*speechdomain.Speech, error) {
	return nil, nil
}

type fakeSpeechCursors struct{ cur *ingPorts.Cursor }

func (c *fakeSpeechCursors) Get(context.Context, string) (*ingPorts.Cursor, error) { return c.cur, nil }
func (c *fakeSpeechCursors) Upsert(_ context.Context, x ingPorts.Cursor) error {
	c.cur = &x
	return nil
}

type fakeMandates struct {
	periods []ingPorts.MandatePeriod
	err     error
}

func (f fakeMandates) ListMandatePeriods(context.Context) ([]ingPorts.MandatePeriod, error) {
	return f.periods, f.err
}

// The worker asks for the sessions derived from the period table and the
// cursor, not a list compiled into the binary.
func TestSpeechesWorker_UsesDerivedSessions(t *testing.T) {
	now := time.Now().UTC()
	periods := []ingPorts.MandatePeriod{{
		Code:      "2026-2030",
		Start:     time.Date(2026, 9, 13, 0, 0, 0, 0, time.UTC),
		End:       time.Date(2030, 9, 8, 0, 0, 0, 0, time.UTC),
		Riksmoten: []string{"2026/27", "2027/28", "2028/29", "2029/30"},
	}}
	since := time.Date(2026, 8, 13, 0, 0, 0, 0, time.UTC)
	client := &fakeSpeechClient{}
	w := workers.NewSpeechesWorker(
		speeches.NewService(&fakeSpeechRepo{}, client),
		&fakeSpeechCursors{cur: &ingPorts.Cursor{DataType: "speeches", LastDate: &since}},
	).WithMandates(fakeMandates{periods: periods})
	_ = w.Run(context.Background())

	want := workers.SpeechSessions(periods, now, since)
	for _, rm := range want {
		if !client.sessions[rm] {
			t.Errorf("riksmöte %s was never requested (wanted %v)", rm, want)
		}
	}
	if client.sessions["2027/28"] || client.sessions["2024/25"] {
		t.Errorf("requested %v, want only %v", client.sessions, want)
	}
}

func TestSpeechesWorker_MandateReadErrorStillAsksForTheCurrentSession(t *testing.T) {
	client := &fakeSpeechClient{}
	w := workers.NewSpeechesWorker(
		speeches.NewService(&fakeSpeechRepo{}, client), &fakeSpeechCursors{},
	).WithMandates(fakeMandates{err: errors.New("db down")})
	_ = w.Run(context.Background())

	current := workers.CurrentRiksmote(time.Now().UTC())
	if !client.sessions[current] {
		t.Errorf("sessions requested = %v, want the current riksmöte %s", client.sessions, current)
	}
}

// A failed fetch must not move the cursor: that is how speeches fell eleven
// months behind while the worker reported success every day.
func TestSpeechesWorker_CursorUnchangedOnFetchError(t *testing.T) {
	before := time.Date(2025, 9, 4, 0, 0, 0, 0, time.UTC)
	cur := &fakeSpeechCursors{cur: &ingPorts.Cursor{DataType: "speeches", LastDate: &before}}
	w := workers.NewSpeechesWorker(
		speeches.NewService(&fakeSpeechRepo{}, &fakeSpeechClient{err: errors.New("down")}), cur)
	_ = w.Run(context.Background())

	if !cur.cur.LastDate.Equal(before) {
		t.Errorf("cursor moved to %v on error, want unchanged %v", cur.cur.LastDate, before)
	}
}

// An empty result is not evidence of freshness.
func TestSpeechesWorker_CursorUnchangedWhenNothingFetched(t *testing.T) {
	before := time.Date(2025, 9, 4, 0, 0, 0, 0, time.UTC)
	cur := &fakeSpeechCursors{cur: &ingPorts.Cursor{DataType: "speeches", LastDate: &before}}
	w := workers.NewSpeechesWorker(
		speeches.NewService(&fakeSpeechRepo{}, &fakeSpeechClient{}), cur)
	_ = w.Run(context.Background())

	if !cur.cur.LastDate.Equal(before) {
		t.Errorf("cursor moved to %v with no rows, want unchanged %v", cur.cur.LastDate, before)
	}
}

// The cursor advances only as far as data actually retrieved.
func TestSpeechesWorker_CursorAdvancesToNewestFetched(t *testing.T) {
	newest := time.Date(2026, 6, 17, 0, 0, 0, 0, time.UTC)
	older := time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC)
	client := &fakeSpeechClient{rows: []*speechdomain.Speech{{Date: older}, {Date: newest}}}
	cur := &fakeSpeechCursors{}
	w := workers.NewSpeechesWorker(speeches.NewService(&fakeSpeechRepo{}, client), cur)

	if err := w.Run(context.Background()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if cur.cur == nil || cur.cur.LastDate == nil || !cur.cur.LastDate.Equal(newest) {
		t.Errorf("cursor = %v, want %v", cur.cur, newest)
	}
}

// /anforandelista cannot paginate, so a response at the fetch limit means the
// party-riksmöte was truncated with no way to retrieve the remainder. Storing
// it would understate the record with no detectable gap, so fail loudly.
func TestSpeechesWorker_ErrorsOnPossibleTruncation(t *testing.T) {
	rows := make([]*speechdomain.Speech, workers.SpeechFetchLimit)
	for i := range rows {
		rows[i] = &speechdomain.Speech{Date: time.Date(2026, 6, 17, 0, 0, 0, 0, time.UTC)}
	}
	w := workers.NewSpeechesWorker(
		speeches.NewService(&fakeSpeechRepo{}, &fakeSpeechClient{rows: rows}),
		&fakeSpeechCursors{})

	if err := w.Run(context.Background()); err == nil {
		t.Error("expected an error when a party-riksmöte fills the fetch limit")
	}
}

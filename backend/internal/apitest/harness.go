// Package apitest builds the real HTTP API against a migrated Postgres and a
// replay of recorded Riksdagen responses.
//
// See README.md in this directory for how to record a fixture and write a test.
package apitest

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"riksdagskollen/internal/api"
)

// Harness is one test's API: its own database, a replay server, and the
// production router pointed at that server.
type Harness struct {
	t      *testing.T
	DB     *pgxpool.Pool
	App    *api.Router
	replay *Replay
}

// Option configures a harness.
type Option func(*Harness)

// WithNow pins the clock the votes ingestion worker uses to choose a riksmöte.
// Record the fixture for that riksmöte. Production leaves the clock unset.
func WithNow(now time.Time) Option {
	return func(h *Harness) {
		h.App.SetNow(func() time.Time { return now })
	}
}

// Start migrates a fresh database, serves fixtureDir, and builds the router.
func Start(t *testing.T, fixtureDir string, opts ...Option) *Harness {
	t.Helper()
	db := OpenDB(t)
	replay := NewReplay(t, fixtureDir)
	app, err := api.NewRouter(db, "", replay.URL)
	if err != nil {
		t.Fatalf("router: %v", err)
	}
	h := &Harness{t: t, DB: db, App: app, replay: replay}
	for _, opt := range opts {
		opt(h)
	}
	return h
}

// RunWorker runs one ingestion worker by its registered name.
func (h *Harness) RunWorker(ctx context.Context, name string) {
	h.t.Helper()
	if err := h.App.RunWorker(ctx, name); err != nil {
		h.t.Fatalf("worker %s: %v", name, err)
	}
}

// Do sends req to the router without opening a port.
func (h *Harness) Do(req *http.Request) *httptest.ResponseRecorder {
	h.t.Helper()
	rec := httptest.NewRecorder()
	h.App.ServeHTTP(rec, req)
	return rec
}

// Served reports how many recorded responses were returned for path.
// Zero means the client did not use the replay server.
func (h *Harness) Served(path string) int {
	n := 0
	needle := " " + path + "?"
	for key, hits := range h.replay.Hits() {
		if strings.Contains(key, needle) {
			n += hits
		}
	}
	return n
}

// ReplayURL is the base URL of the replay server, for tests that build their
// own Riksdagen client (for example to run a service outside the scheduler).
func (h *Harness) ReplayURL() string { return h.replay.URL }

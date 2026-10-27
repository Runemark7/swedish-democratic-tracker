package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"riksdagskollen/internal/ingestion"
	ingestionhttp "riksdagskollen/internal/ingestion/adapters/http"
)

// A DATE cast to timestamptz is midnight in the session TimeZone. Under
// Europe/Stockholm on 2026-10-21 that is 2026-10-20 22:00 UTC, which the
// report then printed as 2026-10-20. The stored day has to survive that zone.
func TestSnapshot_SpeechDateIsTheStoredCalendarDay(t *testing.T) {
	dsn := startPostgres(t)
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)

	if _, err := pool.Exec(ctx, `
		CREATE TABLE speeches (date DATE NOT NULL);
		CREATE TABLE voteringar (system_datum TIMESTAMPTZ);
		CREATE TABLE ingestion_runs (
			id            BIGSERIAL PRIMARY KEY,
			worker_name   TEXT NOT NULL,
			started_at    TIMESTAMPTZ NOT NULL,
			finished_at   TIMESTAMPTZ,
			status        TEXT NOT NULL,
			rows_affected INT,
			error_msg     TEXT
		);
		INSERT INTO speeches (date) VALUES ('2026-10-21'), ('2026-08-13');
		INSERT INTO voteringar (system_datum) VALUES ('2026-10-15 12:52:10+00');
	`); err != nil {
		t.Fatal(err)
	}

	var zone string
	if err := pool.QueryRow(ctx, "SHOW timezone").Scan(&zone); err != nil {
		t.Fatal(err)
	}
	if zone != "Europe/Stockholm" {
		t.Fatalf("timezone = %q, want Europe/Stockholm", zone)
	}

	snap, err := NewFreshnessRepository(pool).Snapshot(ctx)
	if err != nil {
		t.Fatal(err)
	}
	wantSpeech := time.Date(2026, 10, 21, 0, 0, 0, 0, time.UTC)
	if snap.SpeechesLatest == nil || !snap.SpeechesLatest.Equal(wantSpeech) {
		got := "<nil>"
		if snap.SpeechesLatest != nil {
			got = snap.SpeechesLatest.UTC().Format(time.RFC3339)
		}
		t.Fatalf("speech = %s, want 2026-10-21T00:00:00Z", got)
	}
	wantVote := time.Date(2026, 10, 15, 12, 52, 10, 0, time.UTC)
	if snap.VotesLatest == nil || !snap.VotesLatest.Equal(wantVote) {
		got := "<nil>"
		if snap.VotesLatest != nil {
			got = snap.VotesLatest.UTC().Format(time.RFC3339)
		}
		t.Fatalf("vote = %s, want 2026-10-15T12:52:10Z", got)
	}

	now := time.Date(2026, 12, 27, 6, 0, 0, 0, time.UTC)
	alerts := ingestion.Alerts(snap, now)
	joined := strings.Join(alerts, "\n")
	for _, want := range []string{"2026-10-21", "2026-10-15"} {
		if !strings.Contains(joined, want) {
			t.Errorf("alerts = %v, want a line naming %s", alerts, want)
		}
	}

	h := ingestionhttp.NewHandler(NewFreshnessRepository(pool), nil)
	rec := httptest.NewRecorder()
	h.ServeFreshness(rec, httptest.NewRequest(http.MethodGet, "/api/ingestion/freshness", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body %s", rec.Code, rec.Body.String())
	}
	var body struct {
		Speeches struct {
			LatestDate *string `json:"latestDate"`
		} `json:"speeches"`
		Votes struct {
			LatestDate *string `json:"latestDate"`
			LatestAt   *string `json:"latestAt"`
		} `json:"votes"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Speeches.LatestDate == nil || *body.Speeches.LatestDate != "2026-10-21" {
		t.Fatalf("reported speech date = %v, want 2026-10-21", body.Speeches.LatestDate)
	}
	if body.Votes.LatestDate == nil || *body.Votes.LatestDate != "2026-10-15" {
		t.Fatalf("reported vote date = %v, want 2026-10-15", body.Votes.LatestDate)
	}
	if body.Votes.LatestAt == nil || *body.Votes.LatestAt != "2026-10-15T12:52:10Z" {
		t.Fatalf("reported vote time = %v, want 2026-10-15T12:52:10Z", body.Votes.LatestAt)
	}
}

func startPostgres(t *testing.T) string {
	t.Helper()
	initdb := postgresBin("initdb")
	pgctl := postgresBin("pg_ctl")
	if initdb == "" || pgctl == "" {
		t.Skip("postgres binaries not installed")
	}

	dir := t.TempDir()
	data := filepath.Join(dir, "data")
	if out, err := exec.Command(initdb, "-D", data, "--username=postgres", "--auth=trust", "--no-sync", "--locale=C", "--encoding=UTF8").CombinedOutput(); err != nil {
		t.Fatalf("initdb: %v\n%s", err, out)
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close()

	logPath := filepath.Join(dir, "postgres.log")
	start := exec.Command(pgctl, "-D", data, "-l", logPath, "-w", "start",
		"-o", fmt.Sprintf("-p %d -k %s -c listen_addresses=127.0.0.1", port, dir))
	if out, err := start.CombinedOutput(); err != nil {
		log, _ := os.ReadFile(logPath)
		t.Fatalf("pg_ctl start: %v\n%s\n%s", err, out, log)
	}
	t.Cleanup(func() {
		_ = exec.Command(pgctl, "-D", data, "-w", "-m", "immediate", "stop").Run()
	})

	dsn := fmt.Sprintf("postgres://postgres@127.0.0.1:%d/postgres?sslmode=disable", port)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	conn, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	if _, err := conn.Exec(ctx, "ALTER DATABASE postgres SET timezone TO 'Europe/Stockholm'"); err != nil {
		t.Fatal(err)
	}
	return dsn
}

func postgresBin(name string) string {
	if p, err := exec.LookPath(name); err == nil {
		return p
	}
	for _, version := range []string{"17", "16", "15"} {
		p := "/usr/lib/postgresql/" + version + "/bin/" + name
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

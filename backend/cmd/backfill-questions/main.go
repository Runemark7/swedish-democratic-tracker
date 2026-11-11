// Command backfill-questions ingests skriftliga frågor, svar and
// interpellationer for the given riksmöten. It is opt-in: nothing in the API
// process or the scheduler runs it, and it must not be run on deploy. Every
// write is an upsert keyed on dok_id and systemdatum, so it can be stopped and
// started again.
package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	ingestionPG "riksdagskollen/internal/ingestion/adapters/postgres"
	"riksdagskollen/internal/ingestion/workers"
	"riksdagskollen/internal/questions"
	questionspg "riksdagskollen/internal/questions/adapters/postgres"
	questionsrd "riksdagskollen/internal/questions/adapters/riksdagen"
)

func main() {
	rm := flag.String("rm", "", "comma-separated riksmöten, e.g. 2021/22 (required)")
	delay := flag.Duration("delay", 200*time.Millisecond, "pause between requests to Riksdagen")
	flag.Parse()
	if *rm == "" {
		slog.Error("-rm is required")
		os.Exit(2)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		slog.Error("DATABASE_URL is required")
		os.Exit(1)
	}
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		slog.Error("connect", "error", err)
		os.Exit(1)
	}
	defer pool.Close()
	workers.FetchDelay = *delay
	svc := questions.NewService(questionspg.NewRepository(pool), questionsrd.NewClientWithBaseURL(""))
	w := workers.NewQuestionsWorker(svc, ingestionPG.NewMandateRepository(pool))
	for _, r := range strings.Split(*rm, ",") {
		if err := w.SyncRiksmote(ctx, r, 1<<30); err != nil {
			slog.Error("riksmöte failed", "rm", r, "error", err)
			os.Exit(1)
		}
	}
}

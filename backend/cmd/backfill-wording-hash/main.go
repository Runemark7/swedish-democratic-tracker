// Command backfill-wording-hash fills the wording hash of yrkanden stored
// before Återkommande krav existed (issue #145).
//
// New and changed motions get their hash on ingest, so this is only needed for
// rows already in the database. It is deliberately opt-in: nothing in the API
// process or the scheduler runs it. It only touches rows whose hash is NULL,
// walks them in (dok_id, nummer) order in small batches, and makes no external
// requests, so it is idempotent and safe to interrupt and run again.
package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"

	"github.com/jackc/pgx/v5/pgxpool"

	"riksdagskollen/internal/motions"
	motionpg "riksdagskollen/internal/motions/adapters/postgres"
)

func main() {
	batch := flag.Int("batch", 500, "yrkanden read and written per database round trip")
	flag.Parse()

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

	svc := motions.NewService(motionpg.NewRepository(pool), nil)
	n, err := svc.BackfillWordingHashes(ctx, *batch)
	if err != nil {
		slog.Error("backfill", "written", n, "error", err)
		os.Exit(1)
	}
	slog.Info("wording hashes written", "count", n)
}

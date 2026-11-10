// Command recompute-origins re-resolves every votering's proposal origin per
// riksmöte (issue #133).
//
// Until #133 the origin step wrote one resolved origin to every votering that
// shared a beteckning and förslagspunkt, whatever its riksmöte, so e.g. every
// SkU16 punkt 6 carried the 2021/22 motion. The code is fixed; this command
// repairs the rows already written. It is deliberately opt-in: nothing in the
// API process or the scheduler runs it.
//
// Two steps, each its own flag so the destructive-looking one is explicit:
//
//	recompute-origins -mark        flag every votering for re-enrichment
//	recompute-origins              resolve all flagged voteringar (resumable)
//	recompute-origins -mark -run   both, in order
//
// The mark step only sets origin_enriched = false; existing origins stay
// visible until each votering is re-resolved. The resolve step walks in
// (riksmöte, beteckning, förslagspunkt) order, unflags each votering as it is
// written, and pauses -delay between requests. If it is interrupted, start it
// again without -mark and it carries on where it stopped. Running it with
// nothing flagged does nothing.
package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"riksdagskollen/internal/votes"
	votepg "riksdagskollen/internal/votes/adapters/postgres"
	voterd "riksdagskollen/internal/votes/adapters/riksdagen"
)

func main() {
	mark := flag.Bool("mark", false, "flag every votering for re-enrichment (run once, before resolving)")
	run := flag.Bool("run", false, "resolve flagged voteringar (implied when -mark is not given)")
	delay := flag.Duration("delay", 200*time.Millisecond, "pause between Riksdagen requests")
	batch := flag.Int("batch", 100, "voteringar read per database round trip")
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

	svc := votes.NewService(votepg.NewRepository(pool), voterd.NewClient())
	svc.SetOriginDelay(*delay)

	if *mark {
		n, err := svc.MarkOriginsForRecompute(ctx)
		if err != nil {
			slog.Error("mark", "error", err)
			os.Exit(1)
		}
		slog.Info("flagged voteringar for re-enrichment", "count", n)
	}
	if *run || !*mark {
		started := time.Now()
		n, err := svc.RecomputeOrigins(ctx, *batch)
		if err != nil {
			slog.Error("recompute stopped; re-run without -mark to resume", "resolved", n, "error", err)
			os.Exit(1)
		}
		slog.Info("recompute complete", "resolved", n, "elapsed", time.Since(started))
	}
}

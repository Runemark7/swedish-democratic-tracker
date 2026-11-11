// Command backfill-motions ingests motions and betänkanden for every riksmöte
// with voteringar, newest first back to 2002/03 (issue #144).
//
// The daily chain only takes the current riksmöte. This is the operator's
// command for the rest. It is deliberately opt-in: nothing in the API process
// or the scheduler runs it, and it must not be run on deploy.
//
// It works one riksmöte at a time, motions first and then the betänkanden the
// riksmöte's motions point at, and records coverage per riksmöte and source in
// motion_backfill. Start it again after an interruption (Ctrl-C, a reset
// connection, a failed riksmöte) and it skips what is complete and carries on;
// every write is an upsert, so repeating work changes nothing. It pauses
// -delay between requests to data.riksdagen.se.
//
// After the walk it recomputes the yrkande links of every betänkande held from
// the stored förslagspunkter (no requests); a votering's proposed-by is derived
// from those links when it is read. -skip-relink leaves that out.
//
// It ends by printing the report: coverage, every decision value seen with
// counts, the values Outcome does not map (shown verbatim on the site), and
// how many voteringar have an origin. -report prints it without fetching.
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
	"riksdagskollen/internal/motions"
	motionpg "riksdagskollen/internal/motions/adapters/postgres"
	motionrd "riksdagskollen/internal/motions/adapters/riksdagen"
)

func main() {
	rm := flag.String("rm", "", "comma-separated riksmöten to take, e.g. 2021/22 (default: every riksmöte with voteringar)")
	from := flag.String("from", workers.OldestBackfillRiksmote, "oldest riksmöte to take")
	sources := flag.String("sources", "", "motioner, betankanden or both (default both)")
	delay := flag.Duration("delay", 200*time.Millisecond, "pause between requests to Riksdagen")
	force := flag.Bool("force", false, "redo riksmöten recorded as complete")
	skipRelink := flag.Bool("skip-relink", false, "do not recompute the yrkande links (and so every votering's origin) after the walk")
	reportOnly := flag.Bool("report", false, "print the report from what is stored; fetch nothing")
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

	workers.FetchDelay = *delay
	svc := motions.NewService(motionpg.NewRepository(pool), motionrd.NewClientWithBaseURL(""))
	b := workers.NewMotionsBackfill(svc, ingestionPG.NewMandateRepository(pool))

	var rep *workers.Report
	if *reportOnly {
		rep, err = b.Report(ctx)
	} else {
		opts := workers.BackfillOptions{From: *from, Force: *force, SkipRelink: *skipRelink}
		if *rm != "" {
			opts.Riksmoten = strings.Split(*rm, ",")
		}
		switch *sources {
		case "":
		case "motioner", "betankanden":
			opts.Sources = []string{*sources}
		default:
			slog.Error("-sources must be motioner or betankanden")
			os.Exit(2)
		}
		rep, err = b.Run(ctx, opts)
	}
	if rep != nil {
		rep.Write(os.Stdout)
	}
	if err != nil {
		slog.Error("backfill", "error", err)
		os.Exit(1)
	}
	if rep != nil && (rep.Failed > 0 || rep.Interrupted) {
		// A partial run must not be mistaken for a complete one.
		os.Exit(1)
	}
}

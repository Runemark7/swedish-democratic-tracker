// Command backfill-speeches ingests anföranden for the riksmöten before the
// ones the daily worker takes (issue #147), newest first.
//
// It is deliberately opt-in: nothing in the API process or the scheduler runs
// it, and it must not be run on deploy. It works one riksmöte at a time: the
// speech lists (one request per party), a politicians row for every speaker
// Riksdagen's person data knows, the speeches, the text of each (one request
// per speech), and the link to the betänkande each debates (the rules of
// #146). Coverage and storage growth are recorded per riksmöte in
// speech_backfill. Start it again after an interruption and it skips the
// riksmöten that are complete; every write is an upsert, so repeating work
// changes nothing. It pauses -delay between requests to data.riksdagen.se.
//
// The default range is 2014/15 up to the current riksmöte. The report ends
// with the storage each riksmöte added, to size the older periods by (#148).
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"riksdagskollen/internal/ingestion/workers"
	"riksdagskollen/internal/politicians"
	polpg "riksdagskollen/internal/politicians/adapters/postgres"
	polrd "riksdagskollen/internal/politicians/adapters/riksdagen"
	"riksdagskollen/internal/speeches"
	spgg "riksdagskollen/internal/speeches/adapters/postgres"
	sprd "riksdagskollen/internal/speeches/adapters/riksdagen"
)

func main() {
	rm := flag.String("rm", "", "comma-separated riksmöten to take, e.g. 2017/18 (default: -from to -to)")
	from := flag.String("from", workers.OldestSpeechBackfillRiksmote, "oldest riksmöte to take")
	to := flag.String("to", "", "newest riksmöte to take (default: the current one)")
	delay := flag.Duration("delay", 200*time.Millisecond, "pause between requests to Riksdagen")
	force := flag.Bool("force", false, "redo riksmöten recorded as complete")
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
	sp := speeches.NewService(spgg.NewRepository(pool), sprd.NewClientWithBaseURL(""))
	pol := politicians.NewService(polpg.NewRepository(pool), polrd.NewClientWithBaseURL(""))
	b := workers.NewSpeechesBackfill(sp, pol)

	opts := workers.SpeechBackfillOptions{From: *from, To: *to, Force: *force}
	if *rm != "" {
		opts.Riksmoten = strings.Split(*rm, ",")
	}
	rep, err := b.Run(ctx, opts)
	if rep != nil {
		rep.Write(os.Stdout)
	}
	if err != nil {
		slog.Error("backfill", "error", err)
		os.Exit(1)
	}
	if len(rep.Rows) == 0 {
		fmt.Println("inga riksmöten i intervallet")
	}
	if rep.Failed > 0 || rep.Interrupted {
		// A partial run must not be mistaken for a complete one.
		os.Exit(1)
	}
}

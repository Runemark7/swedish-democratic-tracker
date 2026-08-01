// Command backfill retrieves the complete voting record for a mandate period.
//
// The daily worker only ever fetches forward from its cursor, so it cannot
// recover a gap. This command exists to fill one: as of 2026-08-01 production
// held 32 of the 2 562 voteringar in the 2022-2026 mandate — roughly 1.2% of
// the record the site exists to present.
//
// Strategy (see docs/superpowers/plans/2026-08-01-pre-election-record.md):
// enumerate voteringar per riksmöte via /dokumentlista, which paginates and
// reports a total, then fetch ballots one betänkande at a time. /voteringlista
// ignores `p` and caps `sz` at 10 000, so it cannot be paged directly.
//
// Safe to re-run: every write is an upsert, so an interrupted run resumes by
// simply starting again.
package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"riksdagskollen/internal/ingestion/workers"
	"riksdagskollen/internal/votes"
	votepg "riksdagskollen/internal/votes/adapters/postgres"
	voterd "riksdagskollen/internal/votes/adapters/riksdagen"
	voteports "riksdagskollen/internal/votes/ports"
)

const enumPageSize = 200

func main() {
	only := flag.String("rm", "", "backfill a single riksmöte, e.g. 2025/26")
	delay := flag.Duration("delay", 200*time.Millisecond, "pause between API requests")
	flag.Parse()

	ctx := context.Background()

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

	riksmoten := workers.MandateRiksmoten
	if *only != "" {
		riksmoten = []string{*only}
	}

	started := time.Now()
	for _, rm := range riksmoten {
		if err := backfillRiksmote(ctx, pool, svc, rm, *delay); err != nil {
			slog.Error("riksmöte failed", "rm", rm, "error", err)
		}
	}
	slog.Info("backfill complete", "riksmoten", len(riksmoten), "elapsed", time.Since(started))
}

func backfillRiksmote(ctx context.Context, pool *pgxpool.Pool, svc *votes.Service, rm string, delay time.Duration) error {
	// 1. Enumerate every votering in the riksmöte. @traffar is the coverage
	//    denominator: Riksdagen's count, not ours.
	seen := map[string]bool{}
	var order []string
	expected := 0
	unreachable := 0
	// Latest decision date, taken from the enumeration's `datum`. Ballot rows
	// carry only `systemdatum`, which is when Riksdagen last *touched* the
	// record -- 2022/23 ballots show 2026 timestamps -- so it must never be
	// presented as a decision date.
	var lastDecision time.Time

	for page := 1; ; page++ {
		refs, total, err := svc.ListVoteringar(ctx, rm, page, enumPageSize)
		if err != nil {
			return err
		}
		if total > 0 {
			expected = total
		}
		for _, r := range refs {
			if r.Beteckning == "" {
				// A votering on a motion rather than a betänkande. There is no
				// beteckning to partition by, so this strategy cannot fetch it.
				// Counted, not silently dropped.
				unreachable++
				continue
			}
			if d, err := time.Parse("2006-01-02", r.Date); err == nil && d.After(lastDecision) {
				lastDecision = d
			}
			if !seen[r.Beteckning] {
				seen[r.Beteckning] = true
				order = append(order, r.Beteckning)
			}
		}
		if len(refs) < enumPageSize {
			break
		}
		time.Sleep(delay)
	}
	slog.Info("enumerated", "rm", rm, "expected", expected,
		"betankanden", len(order), "unreachable", unreachable)

	// 2. One request per betänkande returns the whole chamber for it.
	//    Since is deliberately left zero: a backfill must not apply the
	//    incremental cutoff, or it re-skips the very gap it exists to fill.
	stored := 0
	var empty []string
	for i, bet := range order {
		vv, err := svc.FetchAndStore(ctx, voteports.FetchVotesFilter{
			Session: rm, Beteckning: bet,
		})
		if err != nil {
			slog.Warn("fetch failed", "rm", rm, "bet", bet, "error", err)
			continue
		}
		if len(vv) == 0 {
			// Riksdagen lists a votering document for this betänkande but
			// publishes no ballots for it. Named rather than absorbed into an
			// unexplained shortfall, so the gap is attributable.
			empty = append(empty, bet)
		}
		stored += len(vv)
		if (i+1)%25 == 0 || i+1 == len(order) {
			slog.Info("progress", "rm", rm, "betankanden", i+1, "of", len(order), "rows", stored)
		}
		time.Sleep(delay)
	}

	// 3. Record what we actually hold, so the site can state its own coverage.
	var ingested int
	if err := pool.QueryRow(ctx, `
		SELECT count(DISTINCT beteckning || ':' || forslagspunkt)
		FROM votes WHERE session = $1`, rm).Scan(&ingested); err != nil {
		return err
	}
	var lastDate *time.Time
	if !lastDecision.IsZero() {
		lastDate = &lastDecision
	}

	if _, err := pool.Exec(ctx, `
		INSERT INTO ingestion_coverage
			(riksmote, expected_voteringar, ingested_voteringar, unreachable_voteringar, last_vote_date, checked_at)
		VALUES ($1, $2, $3, $4, $5, now())
		ON CONFLICT (riksmote) DO UPDATE SET
			expected_voteringar    = EXCLUDED.expected_voteringar,
			ingested_voteringar    = EXCLUDED.ingested_voteringar,
			unreachable_voteringar = EXCLUDED.unreachable_voteringar,
			last_vote_date         = EXCLUDED.last_vote_date,
			checked_at             = now()`,
		rm, expected, ingested, unreachable, lastDate); err != nil {
		return err
	}

	// Unexplained shortfall: anything missing beyond the voteringar we know we
	// cannot reach. This is the number worth investigating.
	if len(empty) > 0 {
		slog.Warn("betänkanden enumerated with no published ballots",
			"rm", rm, "count", len(empty), "beteckningar", empty)
	}

	// Unexplained shortfall: missing beyond the voteringar we know we cannot
	// reach. Anything left here is worth investigating.
	shortfall := expected - ingested - unreachable
	slog.Info("riksmöte done", "rm", rm,
		"ingested", ingested, "expected", expected,
		"unreachable", unreachable, "emptyBetankanden", len(empty),
		"unexplained", shortfall, "rows", stored)
	return nil
}

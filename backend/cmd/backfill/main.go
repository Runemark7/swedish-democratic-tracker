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
// Scope comes from the mandate_periods table rather than a list compiled here,
// so this command and the period-scoped queries the site serves can never
// disagree about which riksmöten a period contains. With no flags it walks
// every period the table holds, oldest first — all six are meant to be held
// complete, so that is the honest default rather than a subset we picked. Use
// -period to work one at a time.
//
// Safe to re-run: every write is an upsert, so an interrupted run resumes by
// simply starting again.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"riksdagskollen/internal/votes"
	votepg "riksdagskollen/internal/votes/adapters/postgres"
	voterd "riksdagskollen/internal/votes/adapters/riksdagen"
	votedomain "riksdagskollen/internal/votes/domain"
	voteports "riksdagskollen/internal/votes/ports"
)

const enumPageSize = 200

// upstreamBackoffs is the wait before each retry of a reset or timeout.
// Run 3 died at the last sessions when data.riksdagen.se started resetting
// connections; the Job then retried the whole walk and failed immediately.
var upstreamBackoffs = []time.Duration{2 * time.Second, 8 * time.Second, 20 * time.Second}

func main() {
	only := flag.String("rm", "", "backfill a single riksmöte, e.g. 2025/26")
	period := flag.String("period", "", "backfill one mandate period, e.g. 2018-2022 (default: every period in mandate_periods)")
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

	var riksmoten []string
	switch {
	case *only != "":
		riksmoten = []string{*only}
	default:
		riksmoten, err = loadRiksmoten(ctx, pool, *period)
		if err != nil {
			slog.Error("resolve scope", "period", *period, "error", err)
			os.Exit(1)
		}
	}
	slog.Info("backfill scope", "period", *period, "riksmoten", riksmoten)

	started := time.Now()
	failed := 0
	for _, rm := range riksmoten {
		if err := backfillRiksmote(ctx, pool, svc, rm, *delay); err != nil {
			// Keep going: one riksmöte failing is not a reason to abandon the
			// rest, and every write is an upsert so a later re-run repairs it.
			slog.Error("riksmöte failed", "rm", rm, "error", err)
			failed++
		}
	}
	slog.Info("backfill complete", "riksmoten", len(riksmoten), "failed", failed,
		"elapsed", time.Since(started))
	if failed > 0 {
		// Exit non-zero so a partial run cannot be mistaken for a complete one
		// by whatever invoked it.
		os.Exit(1)
	}
}

// loadRiksmoten resolves the scope from mandate_periods. The table is the
// single source for which riksmöten a period contains -- the same rows the
// period-scoped queries in committees and votes join against -- so a backfill
// cannot fetch a window the site would then not count.
//
// An empty period means every period the table holds, oldest first. Ordering by
// start_date is not cosmetic: it puts the earliest, least-certain riksmöten
// first, where a scope or coverage problem shows up while the run is still
// short enough to abandon cheaply.
func loadRiksmoten(ctx context.Context, pool *pgxpool.Pool, period string) ([]string, error) {
	if period != "" {
		var rms []string
		err := pool.QueryRow(ctx,
			`SELECT riksmoten FROM mandate_periods WHERE code = $1`, period).Scan(&rms)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("no mandate period %q in mandate_periods", period)
		}
		if err != nil {
			return nil, err
		}
		return rms, nil
	}

	rows, err := pool.Query(ctx,
		`SELECT unnest(riksmoten) FROM mandate_periods ORDER BY start_date`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var all []string
	for rows.Next() {
		var rm string
		if err := rows.Scan(&rm); err != nil {
			return nil, err
		}
		all = append(all, rm)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(all) == 0 {
		// Not an empty result to shrug at: with no periods declared the command
		// would report "backfill complete" having fetched nothing.
		return nil, errors.New("mandate_periods is empty; nothing to backfill")
	}
	return all, nil
}

// transientUpstream reports errors worth retrying. A reset or a timeout from
// data.riksdagen.se is not a statement that the session is missing.
func transientUpstream(err error) bool {
	if err == nil {
		return false
	}
	var nerr net.Error
	if errors.As(err, &nerr) && nerr.Timeout() {
		return true
	}
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "connection reset") ||
		strings.Contains(s, "connection refused") ||
		strings.Contains(s, "broken pipe") ||
		strings.Contains(s, "timeout") ||
		strings.Contains(s, "deadline exceeded") ||
		strings.Contains(s, "eof") ||
		strings.Contains(s, "returned 429") ||
		strings.Contains(s, "returned 500") ||
		strings.Contains(s, "returned 502") ||
		strings.Contains(s, "returned 503") ||
		strings.Contains(s, "returned 504")
}

// retryUpstream calls fn, and on a transient error waits and tries again.
// A non-transient error is returned immediately. The last error is returned
// when the attempts run out.
func retryUpstream(ctx context.Context, what string, fn func() error) error {
	var err error
	for attempt := 0; ; attempt++ {
		err = fn()
		if err == nil || !transientUpstream(err) || attempt >= len(upstreamBackoffs) {
			return err
		}
		wait := upstreamBackoffs[attempt]
		slog.Warn("upstream retry", "what", what, "attempt", attempt+1, "error", err, "sleep", wait)
		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
}

// heldComplete is the same comparison the end-of-session log uses for an
// unexplained shortfall. Unreachable voteringar are not a gap. A surplus
// over Riksdagen's listing is still complete.
func heldComplete(expected, unreachable, ingested int) bool {
	return expected > 0 && ingested+unreachable >= expected
}

// sessionAlreadyComplete reports whether this riksmöte is already stored
// against the denominator recorded for it. Enumeration can fail after the
// data is in; that must not fail the Job, or a retry pod walks every session
// and exits non-zero while the record is whole.
func sessionAlreadyComplete(ctx context.Context, pool *pgxpool.Pool, rm string) (bool, error) {
	var expected *int
	var unreachable int
	err := pool.QueryRow(ctx, `
		SELECT expected_voteringar, unreachable_voteringar
		FROM ingestion_coverage WHERE riksmote = $1`, rm).Scan(&expected, &unreachable)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if expected == nil || *expected <= 0 {
		return false, nil
	}
	var ingested int
	if err := pool.QueryRow(ctx, `
		SELECT count(DISTINCT beteckning || ':' || forslagspunkt)
		FROM voteringar WHERE session = $1`, rm).Scan(&ingested); err != nil {
		return false, err
	}
	return heldComplete(*expected, unreachable, ingested), nil
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
		var refs []voteports.VoteringRef
		var total int
		err := retryUpstream(ctx, rm+" page "+fmt.Sprint(page), func() error {
			var lerr error
			refs, total, lerr = svc.ListVoteringar(ctx, rm, page, enumPageSize)
			return lerr
		})
		if err != nil {
			complete, cerr := sessionAlreadyComplete(ctx, pool, rm)
			if cerr != nil {
				return fmt.Errorf("enumerate %s: %w (coverage check: %v)", rm, err, cerr)
			}
			if complete {
				slog.Warn("enumeration failed; session already complete", "rm", rm, "page", page, "error", err)
				return nil
			}
			return fmt.Errorf("enumerate %s: %w", rm, err)
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
		var vv []*votedomain.Vote
		err := retryUpstream(ctx, rm+"/"+bet, func() error {
			var ferr error
			vv, ferr = svc.FetchAndStore(ctx, voteports.FetchVotesFilter{
				Session: rm, Beteckning: bet,
			})
			return ferr
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
	// Counted from the two tables directly: the voteringar we hold for the
	// riksmöte, and the ballots cast in them. DISTINCT stays on the vote point
	// because two voteringar can share a beteckning and förslagspunkt.
	var ingested, ballotsHeld int
	if err := pool.QueryRow(ctx, `
		SELECT
			(SELECT count(DISTINCT beteckning || ':' || forslagspunkt)
			   FROM voteringar WHERE session = $1),
			(SELECT count(*)
			   FROM ballots b JOIN voteringar vg ON vg.id = b.votering_ref
			  WHERE vg.session = $1)`, rm).Scan(&ingested, &ballotsHeld); err != nil {
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

	// Ballots Riksdagen published for this riksmöte that we do not hold.
	//
	// UpsertMany drops any ballot whose member is absent from politicians, so
	// "rows fetched" is not "rows stored" -- 2002/03 fetched 226 152 and stored
	// 224 058, a 0.9% gap of substitute members we have no profile for.
	// Reporting only the fetched figure would state a completeness we do not
	// have. Recomputed from the fetch each run rather than from an insert
	// delta, so it stays true on a re-run, where every write is an update.
	//
	// This does not move the votering counts: the missing ballots belong to
	// voteringar other members still cover, so coverage stays complete at the
	// votering level while being short at the ballot level. Both are stated.
	droppedBallots := stored - ballotsHeld
	if droppedBallots < 0 {
		// The riksmöte holds more than this run fetched -- an earlier run
		// reached ballots this one did not. Not a shortfall to report.
		droppedBallots = 0
	}

	// Voteringar missing beyond the ones we know we cannot reach, and the ones
	// we hold beyond what Riksdagen listed. Both directions occur, so they are
	// reported as two named numbers rather than one signed one.
	//
	// The surplus is not a counting error. /dokumentlista?doktyp=votering does
	// not list every votering: fetching a whole betänkande from /voteringlista
	// returns förslagspunkter the listing omits -- 2005/06 yielded five (JuSoU1
	// punkt 10/19/40/46 and UbU14 punkt 104), each a full 349-ballot chamber,
	// and 2007/08 ten more. So `expected` is Riksdagen's published count, not a
	// ceiling on what exists, and coverage stated against it can exceed 100%.
	// Measured 2026-08-22; every riksmöte from 2014/15 on came out exact, so
	// this is confined to the older record.
	shortfall := expected - ingested - unreachable
	surplus := 0
	if shortfall < 0 {
		surplus, shortfall = -shortfall, 0
	}
	slog.Info("riksmöte done", "rm", rm,
		"ingested", ingested, "expected", expected,
		"unreachable", unreachable, "emptyBetankanden", len(empty),
		"unexplained", shortfall, "beyondListing", surplus,
		"ballotsFetched", stored, "ballotsHeld", ballotsHeld,
		"ballotsDropped", droppedBallots)
	if surplus > 0 {
		slog.Warn("voteringar held beyond Riksdagen's own listing; coverage denominator is not a ceiling",
			"rm", rm, "count", surplus)
	}
	if droppedBallots > 0 {
		slog.Warn("ballots dropped: no politician profile for the member",
			"rm", rm, "count", droppedBallots)
	}
	return nil
}

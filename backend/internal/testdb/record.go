// Package testdb prepares the database the existing Postgres integration tests
// were written against.
//
// Those tests pin counts and votering ids from the ingested record (116 AU
// voteringar, the NU paging pair, two tied party lines). Migrations create the
// schema and the seeded goals, not that record. CI applies the migrations to
// an empty Postgres and then runs the tests, so an empty voteringar table is
// filled with the smallest set of rows that makes those pins true.
//
// A database that already holds voteringar — the local dev record — is left
// alone. The pins there come from ingestion, and adding these rows would
// change the counts.
package testdb

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// lockKey serialises the empty-record seed across packages. go test runs
// packages in parallel, and both the votes and committees suites call
// EnsureVoteRecord.
const lockKey int64 = 131131

type voteRow struct {
	id         string
	beteckning string
	punkt      string
	session    string
	ballot     bool
}

// EnsureVoteRecord inserts the pinned vote record when voteringar is empty.
func EnsureVoteRecord(ctx context.Context, pool *pgxpool.Pool) error {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("acquire: %w", err)
	}
	defer conn.Release()

	if _, err := conn.Exec(ctx, `SELECT pg_advisory_lock($1)`, lockKey); err != nil {
		return fmt.Errorf("lock: %w", err)
	}
	defer conn.Exec(context.Background(), `SELECT pg_advisory_unlock($1)`, lockKey)

	var n int
	if err := conn.QueryRow(ctx, `SELECT count(*) FROM voteringar`).Scan(&n); err != nil {
		return fmt.Errorf("count voteringar: %w", err)
	}
	if n > 0 {
		return nil
	}

	tx, err := conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := insertVoteRecord(ctx, tx); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}

func insertVoteRecord(ctx context.Context, tx pgx.Tx) error {
	if _, err := tx.Exec(ctx, `
		INSERT INTO politicians (intressent_id, first_name, last_name, party) VALUES
			('ci-s', 'Ci', 'Ess', 'S'),
			('ci-dash', 'Ci', 'Dash', '-')
	`); err != nil {
		return fmt.Errorf("politicians: %w", err)
	}
	for _, party := range []string{"MP", "L"} {
		n := 14
		if party == "L" {
			n = 16
		}
		for i := 0; i < n; i++ {
			if _, err := tx.Exec(ctx, `
				INSERT INTO politicians (intressent_id, first_name, last_name, party)
				VALUES ($1, 'Ci', $2, $2)`,
				fmt.Sprintf("ci-%s-%d", party, i), party); err != nil {
				return fmt.Errorf("politician %s: %w", party, err)
			}
		}
	}

	rows := votePlan()
	const batch = 400
	for start := 0; start < len(rows); start += batch {
		end := start + batch
		if end > len(rows) {
			end = len(rows)
		}
		if err := insertVoteringBatch(ctx, tx, rows[start:end]); err != nil {
			return err
		}
	}

	var ballotIDs []string
	for _, row := range rows {
		if row.ballot {
			ballotIDs = append(ballotIDs, row.id)
		}
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO ballots (votering_ref, politician_id, party, vote_result)
		SELECT id, 'ci-s', 'S', 'Ja'
		FROM voteringar
		WHERE votering_id = ANY($1)`, ballotIDs); err != nil {
		return fmt.Errorf("ballots: %w", err)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO ballots (votering_ref, politician_id, party, vote_result)
		SELECT id, 'ci-dash', '-', 'Ja'
		FROM voteringar
		WHERE votering_id = 'ci-au-1'`); err != nil {
		return fmt.Errorf("dash ballot: %w", err)
	}

	// KrU7 punkt 5, 2022/23: MP split 7 Avstår / 7 Ja.
	if err := insertTie(ctx, tx, "922CFC33-E4D2-4F3C-A819-E2927BE94B51", "MP",
		repeat("Avstår", 7, "Ja", 7)); err != nil {
		return err
	}
	// MJU19 punkt 4, 2025/26: L split 8 Frånvarande / 8 Ja.
	if err := insertTie(ctx, tx, "9D3AD261-F412-4DCB-80E8-6D36D3465864", "L",
		repeat("Frånvarande", 8, "Ja", 8)); err != nil {
		return err
	}
	return nil
}

func insertVoteringBatch(ctx context.Context, tx pgx.Tx, rows []voteRow) error {
	var b strings.Builder
	args := make([]any, 0, len(rows)*4)
	b.WriteString(`INSERT INTO voteringar (votering_id, beteckning, forslagspunkt, session) VALUES `)
	for i, row := range rows {
		if i > 0 {
			b.WriteByte(',')
		}
		n := i * 4
		fmt.Fprintf(&b, "($%d,$%d,$%d,$%d)", n+1, n+2, n+3, n+4)
		args = append(args, row.id, row.beteckning, row.punkt, row.session)
	}
	if _, err := tx.Exec(ctx, b.String(), args...); err != nil {
		return fmt.Errorf("voteringar: %w", err)
	}
	return nil
}

func insertTie(ctx context.Context, tx pgx.Tx, voteringID, party string, results []string) error {
	var ref int
	if err := tx.QueryRow(ctx, `SELECT id FROM voteringar WHERE votering_id = $1`, voteringID).Scan(&ref); err != nil {
		return fmt.Errorf("tie votering %s: %w", voteringID, err)
	}
	for i, result := range results {
		if _, err := tx.Exec(ctx, `
			INSERT INTO ballots (votering_ref, politician_id, party, vote_result)
			VALUES ($1, $2, $3, $4)`,
			ref, fmt.Sprintf("ci-%s-%d", party, i), party, result); err != nil {
			return fmt.Errorf("tie ballot %s: %w", voteringID, err)
		}
	}
	return nil
}

func repeat(a string, na int, b string, nb int) []string {
	out := make([]string, 0, na+nb)
	for i := 0; i < na; i++ {
		out = append(out, a)
	}
	for i := 0; i < nb; i++ {
		out = append(out, b)
	}
	return out
}

// votePlan is the row set the committee and vote integration tests assert.
// NU rows are appended in the order ListByCommitteeWithPositions returns
// them (session, numeric beteckning, numeric förslagspunkt, votering id),
// so the paging test's indices 149 and 150 land on the NU7 pair.
func votePlan() []voteRow {
	var rows []voteRow
	add := func(id, bet, punkt, session string, ballot bool) {
		rows = append(rows, voteRow{id, bet, punkt, session, ballot})
	}
	addPair := func(session, bet, punkt, a, b string) {
		if b < a {
			a, b = b, a
		}
		add(a, bet, punkt, session, true)
		add(b, bet, punkt, session, true)
	}

	for i := 1; i <= 116; i++ {
		add(fmt.Sprintf("ci-au-%d", i), fmt.Sprintf("AU%d", i), "1", "2022/23", true)
	}
	// 139 NU voteringar sort before every later session. The ten voteringar
	// on earlier duplicated points also sort before 2025/26 NU7 punkt 2,
	// which puts that pair at indices 149 and 150.
	for i := 1; i <= 139; i++ {
		add(fmt.Sprintf("ci-nu-202223-%d", i), fmt.Sprintf("NU%d", i), "1", "2022/23", true)
	}
	addPair("2023/24", "NU1", "2",
		"35B5E914-F3FD-45C4-AE5F-A9539273E361",
		"6B741FC0-B732-4CBA-90FF-A16AD5AA1044")
	addPair("2023/24", "NU5", "2", "ci-nu-202324-nu5-a", "ci-nu-202324-nu5-b")
	addPair("2023/24", "NU14", "1", "ci-nu-202324-nu14-a", "ci-nu-202324-nu14-b")
	addPair("2024/25", "NU1", "2", "ci-nu-202425-nu1-a", "ci-nu-202425-nu1-b")
	addPair("2025/26", "NU1", "1", "ci-nu-202526-nu1-a", "ci-nu-202526-nu1-b")
	addPair("2025/26", "NU7", "2",
		"A971F760-DB68-4099-AB44-5E947B2A326D",
		"BD92BDCE-1D0F-4748-A5DF-B17F18A66DE7")
	addPair("2025/26", "NU16", "8", "ci-nu-202526-nu16-a", "ci-nu-202526-nu16-b")
	addPair("2025/26", "NU22", "1", "ci-nu-202526-nu22-a", "ci-nu-202526-nu22-b")
	addPair("2025/26", "NU27", "5", "ci-nu-202526-nu27-a", "ci-nu-202526-nu27-b")
	// 43 fillers after the pair, so offset 150 still returns a full page of 50.
	for i := 100; i <= 142; i++ {
		add(fmt.Sprintf("ci-nu-202526-%d", i), fmt.Sprintf("NU%d", i), "1", "2025/26", true)
	}

	add("922CFC33-E4D2-4F3C-A819-E2927BE94B51", "KrU7", "5", "2022/23", true)
	add("9D3AD261-F412-4DCB-80E8-6D36D3465864", "MJU19", "4", "2025/26", true)

	// One row each so the derived committee list has the standing committees
	// plus UFöU. UU, UbU and UFöU carry a ballot: the wildcard test only
	// notices a missing CommitteeCode filter when those rows come back.
	for _, code := range []string{"CU", "FöU", "JuU", "KU", "SfU", "SkU", "SoU", "TU", "UbU", "UU", "UFöU"} {
		ballot := code == "UbU" || code == "UU" || code == "UFöU"
		add("ci-"+strings.ToLower(code), code+"1", "1", "2022/23", ballot)
	}
	// The rest of the 2 576 floor measured on 2026-08-10. No ballots: the
	// committee list counts voteringar, and these must not show up as NU or AU.
	for i := 1; i <= 2247; i++ {
		add(fmt.Sprintf("ci-fiu-%d", i), fmt.Sprintf("FiU%d", i), "1", "2022/23", false)
	}
	return rows
}

package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"riksdagskollen/internal/motions/domain"
	"riksdagskollen/internal/motions/ports"
)

func (r *Repository) UpsertBeslutListing(ctx context.Context, rows []ports.BetankandeRef) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	for _, b := range rows {
		var day *time.Time
		if t, err := time.Parse("2006-01-02", b.DecisionDate); err == nil {
			day = &t
		}
		sortDate := b.Published
		if day != nil {
			sortDate = *day
		}
		if sortDate.IsZero() {
			sortDate = b.SystemDatum
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO beslut_listing (riksmote, beteckning, dok_id, title, organ, status, beslutsdag, beslutad, sort_date, synced_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9::date, now())
			ON CONFLICT (riksmote, beteckning) DO UPDATE SET
				dok_id = EXCLUDED.dok_id, title = EXCLUDED.title, organ = EXCLUDED.organ, status = EXCLUDED.status,
				beslutsdag = EXCLUDED.beslutsdag, beslutad = EXCLUDED.beslutad, sort_date = EXCLUDED.sort_date, synced_at = now()`,
			b.Riksmote, b.Beteckning, b.DokID, b.Title, b.Organ, b.Status, day, b.ChamberDecided, sortDate); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *Repository) BeslutListed(ctx context.Context, rm string) (int, error) {
	var n int
	err := r.db.QueryRow(ctx, `SELECT count(*) FROM beslut_listing WHERE riksmote = $1`, rm).Scan(&n)
	return n, err
}

// decidedSQL is the decided rule of votes/ports.RiksdagDocument.Decided,
// over the stored columns; $1 is today. A test holds the two together.
const decidedSQL = `(CASE
	WHEN l.beslutad IS NOT NULL THEN l.beslutad
	WHEN lower(btrim(l.status)) = 'planerat' THEN false
	WHEN l.beslutsdag IS NULL THEN false
	ELSE l.beslutsdag <= $1::date END)`

func (r *Repository) ListBeslut(ctx context.Context, f domain.BeslutFilter, today time.Time, after *domain.BeslutCursor, limit int) ([]domain.BeslutRow, error) {
	var out []domain.BeslutRow
	if f.State != domain.BeslutDecided && (after == nil || !after.Decided) {
		rows, err := r.beslutQuery(ctx, f, today, false, after, limit)
		if err != nil {
			return nil, err
		}
		out = rows
		after = nil
	}
	if f.State != domain.BeslutUpcoming && len(out) < limit {
		rows, err := r.beslutQuery(ctx, f, today, true, after, limit-len(out))
		if err != nil {
			return nil, err
		}
		out = append(out, rows...)
	}
	return out, nil
}

func (r *Repository) beslutQuery(ctx context.Context, f domain.BeslutFilter, today time.Time, decided bool, after *domain.BeslutCursor, limit int) ([]domain.BeslutRow, error) {
	args := []any{today.UTC().Format("2006-01-02")}
	arg := func(v any) string { args = append(args, v); return fmt.Sprintf("$%d", len(args)) }
	where := []string{decidedSQL + " = " + arg(decided)}
	if len(f.Utskott) > 0 {
		where = append(where, `upper(l.organ) = ANY(`+arg(upperAll(f.Utskott))+`)`)
	}
	// Each test is a scalar subquery with LIMIT 1 (not EXISTS): EXISTS under OR
	// is planned as a hashed subplan, and the ballots one then reads all 5.5M
	// ballots. A scalar subquery stays a correlated index probe per listing
	// row, and the LATERAL ... LIMIT 1 below keeps the ballots probe on
	// ballots_pkey (votering_ref, politician_id).
	motionSide := func(p string) string {
		return `(SELECT true FROM yrkande_forslagspunkt y JOIN motion_authors a ON a.dok_id = y.motion_dok_id
				WHERE y.riksmote = l.riksmote AND y.beteckning = l.beteckning AND upper(a.party) = ANY(` + p + `) LIMIT 1) IS TRUE`
	}
	reservationSide := func(p string) string {
		return `(SELECT true FROM betankande_statements s, unnest(s.parties) sp
				WHERE s.riksmote = l.riksmote AND s.beteckning = l.beteckning AND s.kind = 'reservation' AND upper(sp) = ANY(` + p + `) LIMIT 1) IS TRUE`
	}
	if len(f.Parties) > 0 {
		p := arg(upperAll(f.Parties))
		where = append(where, `(`+motionSide(p)+` OR `+reservationSide(p)+`)`)
	}
	if len(f.Involved) > 0 {
		p := arg(upperAll(f.Involved))
		where = append(where, `(`+motionSide(p)+` OR `+reservationSide(p)+` OR
			(SELECT true FROM voteringar v
			 CROSS JOIN LATERAL (SELECT 1 FROM ballots b WHERE b.votering_ref = v.id AND upper(b.party) = ANY(`+p+`) LIMIT 1) vb
			 WHERE v.session = l.riksmote AND v.beteckning = l.beteckning LIMIT 1) IS TRUE)`)
	}
	// Decided rows run newest first by their decision date (else publication
	// date); upcoming rows run soonest planned day first, undated ones last.
	eff := "l.sort_date"
	order := "eff DESC, l.riksmote DESC, l.beteckning DESC"
	cmp := "<"
	if !decided {
		eff = "coalesce(l.beslutsdag, DATE '9999-12-31')"
		order, cmp = "eff ASC, l.riksmote ASC, l.beteckning ASC", ">"
	}
	if after != nil {
		where = append(where, fmt.Sprintf(`(%s, l.riksmote, l.beteckning) %s (%s::date, %s::text, %s::text)`,
			eff, cmp, arg(after.Date.UTC().Format("2006-01-02")), arg(after.Riksmote), arg(after.Beteckning)))
	}
	rows, err := r.db.Query(ctx, `
		SELECT l.riksmote, l.beteckning, l.dok_id, l.title, l.organ, l.status,
		       coalesce(to_char(l.beslutsdag, 'YYYY-MM-DD'), ''), l.beslutad, `+eff+` AS eff, `+decidedSQL+`
		FROM beslut_listing l
		WHERE `+strings.Join(where, " AND ")+`
		ORDER BY `+order+`
		LIMIT `+arg(limit), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.BeslutRow
	for rows.Next() {
		var b domain.BeslutRow
		if err := rows.Scan(&b.Riksmote, &b.Beteckning, &b.DokID, &b.Title, &b.Organ, &b.Status, &b.DecisionDate, &b.ChamberDecided, &b.SortDate, &b.SQLDecided); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

func (r *Repository) BeslutDetails(ctx context.Context, keys []string) (map[string]domain.BeslutDetail, error) {
	out := map[string]domain.BeslutDetail{}
	if len(keys) == 0 {
		return out, nil
	}
	// The keys are "riksmote:beteckning"; a riksmöte has no colon. Joining
	// against the unnest'ed pairs lets forslagspunkter_pkey and
	// voteringar(session, beteckning) be used (a concatenated key cannot).
	rms, bets := make([]string, len(keys)), make([]string, len(keys))
	for i, k := range keys {
		rms[i], bets[i], _ = strings.Cut(k, ":")
	}
	pr, err := r.db.Query(ctx, `
		SELECT f.riksmote || ':' || f.beteckning, f.beslutstyp
		FROM unnest($1::text[], $2::text[]) AS k(riksmote, beteckning)
		JOIN forslagspunkter f ON f.riksmote = k.riksmote AND f.beteckning = k.beteckning
		ORDER BY 1, f.punkt`, rms, bets)
	if err != nil {
		return nil, err
	}
	for pr.Next() {
		var k, t string
		if err := pr.Scan(&k, &t); err != nil {
			pr.Close()
			return nil, err
		}
		d := out[k]
		d.Beslutstyper = append(d.Beslutstyper, t)
		out[k] = d
	}
	pr.Close()
	if err := pr.Err(); err != nil {
		return nil, err
	}
	br, err := r.db.Query(ctx, `
		SELECT v.session || ':' || v.beteckning, v.votering_id, b.party, b.vote_result, count(*)
		FROM unnest($1::text[], $2::text[]) AS k(riksmote, beteckning)
		JOIN voteringar v ON v.session = k.riksmote AND v.beteckning = k.beteckning
		JOIN ballots b ON b.votering_ref = v.id
		GROUP BY 1, 2, 3, 4`, rms, bets)
	if err != nil {
		return nil, err
	}
	defer br.Close()
	for br.Next() {
		var k string
		var c domain.BallotCount
		if err := br.Scan(&k, &c.VoteringID, &c.Party, &c.Result, &c.N); err != nil {
			return nil, err
		}
		d := out[k]
		d.Ballots = append(d.Ballots, c)
		out[k] = d
	}
	return out, br.Err()
}

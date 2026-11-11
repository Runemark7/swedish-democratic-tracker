package postgres

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"riksdagskollen/internal/motions/domain"
)

// kindSQL derives the kind label from Riksdagen's subtyp and the signers, the
// same rule as domain.Motion.Kind. An unknown subtyp is returned as given.
const kindSQL = `CASE lower(m.subtyp)
	WHEN 'kommittémotion' THEN 'Kommittémotion'
	WHEN 'partimotion' THEN 'Partimotion'
	WHEN 'enskild motion' THEN CASE WHEN (
		SELECT count(DISTINCT upper(a.party)) FROM motion_authors a
		WHERE a.dok_id = m.dok_id AND (a.role = '' OR lower(a.role) = 'undertecknare') AND a.party NOT IN ('', '-')
	) > 1 THEN 'Flerpartimotion' ELSE 'Enskild motion' END
	ELSE m.subtyp END`

func likeEscape(s string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
}

func upperAll(in []string) []string {
	out := make([]string, len(in))
	for i, s := range in {
		out[i] = strings.ToUpper(s)
	}
	return out
}

// ListPage implements ports.Repository. The order is (decided, key, dok_id)
// descending, where decided means a linked förslagspunkt has a decision and
// key is then the newest linked betänkande's date, else the filing date.
func (r *Repository) ListPage(ctx context.Context, f domain.ListFilter, after *domain.ListCursor, limit int) ([]domain.ListRow, error) {
	filters := func(arg func(any) string) []string {
		var where []string

		if len(f.Parties) > 0 {
			where = append(where, `EXISTS (SELECT 1 FROM motion_authors a WHERE a.dok_id = m.dok_id AND upper(a.party) = ANY(`+arg(upperAll(f.Parties))+`))`)
		}
		if len(f.Authors) > 0 {
			where = append(where, `EXISTS (SELECT 1 FROM motion_authors a WHERE a.dok_id = m.dok_id AND a.intressent_id = ANY(`+arg(f.Authors)+`))`)
		}
		if len(f.Utskott) > 0 {
			// The utskott a motion is filed to, or the one whose betänkande decided
			// one of its yrkanden (2018/19:2900 is filed to FiU and decided in SkU11).
			p := arg(upperAll(f.Utskott))
			where = append(where, `(upper(m.organ) = ANY(`+p+`) OR EXISTS (
			SELECT 1 FROM yrkande_forslagspunkt l
			WHERE l.motion_dok_id = m.dok_id AND upper(substring(l.beteckning from '^[^0-9]+')) = ANY(`+p+`)))`)
		}
		if len(f.Riksmoten) > 0 {
			where = append(where, `m.riksmote = ANY(`+arg(f.Riksmoten)+`)`)
		}
		if len(f.Kinds) > 0 {
			where = append(where, `(`+kindSQL+`) = ANY(`+arg(f.Kinds)+`)`)
		}
		for _, term := range strings.Fields(f.Query) {
			p := arg("%" + likeEscape(term) + "%")
			where = append(where, `(m.title ILIKE `+p+` OR EXISTS (SELECT 1 FROM motion_yrkanden y WHERE y.dok_id = m.dok_id AND y.wording ILIKE `+p+`))`)
		}
		return where
	}
	// Decided motions are the few that hold a decided förslagspunkt; they are
	// read first, newest decision first, then the undecided by filing date.
	// Each branch has its own ORDER BY ... LIMIT so neither has to rank every
	// motion (the old single query computed the decision for all of them).
	var out []domain.ListRow
	if after == nil || after.Decided {
		rows, err := r.listBranch(ctx, true, filters, after, limit)
		if err != nil {
			return nil, err
		}
		out = rows
	}
	if len(out) < limit {
		rows, err := r.listBranch(ctx, false, filters, after, limit-len(out))
		if err != nil {
			return nil, err
		}
		out = append(out, rows...)
	}
	byDok := map[string]int{}
	var dokIDs []string
	for i := range out {
		byDok[out[i].Motion.DokID] = i
		dokIDs = append(dokIDs, out[i].Motion.DokID)
	}
	if len(out) == 0 {
		return out, nil
	}

	ar, err := r.db.Query(ctx, `SELECT dok_id, ordning, intressent_id, name, party, role FROM motion_authors WHERE dok_id = ANY($1) ORDER BY dok_id, ordning`, dokIDs)
	if err != nil {
		return nil, err
	}
	for ar.Next() {
		var id string
		var a domain.Author
		if err := ar.Scan(&id, &a.Order, &a.PoliticianID, &a.Name, &a.Party, &a.Role); err != nil {
			ar.Close()
			return nil, err
		}
		m := &out[byDok[id]].Motion
		m.Authors = append(m.Authors, a)
	}
	ar.Close()
	if err := ar.Err(); err != nil {
		return nil, err
	}

	yr, err := r.db.Query(ctx, `SELECT dok_id, nummer, kammaren FROM motion_yrkanden WHERE dok_id = ANY($1) ORDER BY dok_id, nummer`, dokIDs)
	if err != nil {
		return nil, err
	}
	for yr.Next() {
		var id string
		var y domain.Yrkande
		if err := yr.Scan(&id, &y.Number, &y.ChamberDecision); err != nil {
			yr.Close()
			return nil, err
		}
		m := &out[byDok[id]].Motion
		m.Yrkanden = append(m.Yrkanden, y)
	}
	yr.Close()
	if err := yr.Err(); err != nil {
		return nil, err
	}

	dr, err := r.db.Query(ctx, `
		SELECT motion_dok_id, yrkande_nummer, decision, riksmote || ':' || beteckning FROM yrkande_forslagspunkt
		WHERE motion_dok_id = ANY($1) ORDER BY motion_dok_id, yrkande_nummer, riksmote, beteckning, punkt`, dokIDs)
	if err != nil {
		return nil, err
	}
	defer dr.Close()
	for dr.Next() {
		var id, dec, bet string
		var n int
		if err := dr.Scan(&id, &n, &dec, &bet); err != nil {
			return nil, err
		}
		row := &out[byDok[id]]
		row.Decision[n] = append(row.Decision[n], dec)
		if !slices.Contains(row.HandledIn, bet) {
			row.HandledIn = append(row.HandledIn, bet)
		}
	}
	return out, dr.Err()
}

// undecidedKeySQL is the sort key of a motion with no decision: its filing date.
const undecidedKeySQL = `COALESCE(m.motion_date::timestamp, 'epoch'::timestamp)`

// decidedSubquery lists the motions with a decided förslagspunkt and the date
// of the newest betänkande among those. A motion whose betänkanden carry no
// date is not decided here, as before.
const decidedSubquery = `(
	SELECT l.motion_dok_id AS dok_id, max(b.systemdatum) AS at
	FROM yrkande_forslagspunkt l
	JOIN forslagspunkter p ON p.riksmote = l.riksmote AND p.beteckning = l.beteckning AND p.punkt = l.punkt
	JOIN betankanden b ON b.riksmote = l.riksmote AND b.beteckning = l.beteckning
	WHERE p.beslutstyp <> ''
	GROUP BY l.motion_dok_id
	HAVING max(b.systemdatum) IS NOT NULL
) dec`

// listBranch reads the decided or the undecided half of the list: filters
// plus, when the cursor lies in this half, the cursor condition.
func (r *Repository) listBranch(ctx context.Context, decided bool, filters func(func(any) string) []string, after *domain.ListCursor, limit int) ([]domain.ListRow, error) {
	const cols = `m.dok_id, m.riksmote, m.beteckning, m.title, m.subtyp, m.motion_date, m.organ, m.tilldelat, m.status`
	var args []any
	arg := func(v any) string { args = append(args, v); return fmt.Sprintf("$%d", len(args)) }
	cond := filters(arg)
	var q string
	if decided {
		if after != nil {
			cond = append(cond, fmt.Sprintf(`(dec.at, m.dok_id) < (%s::timestamp, %s::text)`, arg(after.Key.UTC()), arg(after.DokID)))
		}
		q = `SELECT ` + cols + `, true, dec.at FROM ` + decidedSubquery + ` JOIN motions m ON m.dok_id = dec.dok_id`
		if len(cond) > 0 {
			q += ` WHERE ` + strings.Join(cond, " AND ")
		}
		q += ` ORDER BY dec.at DESC, m.dok_id DESC LIMIT ` + arg(limit)
	} else {
		if after != nil && !after.Decided {
			cond = append(cond, fmt.Sprintf(`(%s, m.dok_id) < (%s::timestamp, %s::text)`, undecidedKeySQL, arg(after.Key.UTC()), arg(after.DokID)))
		}
		cond = append([]string{`NOT EXISTS (
			SELECT 1 FROM yrkande_forslagspunkt l
			JOIN forslagspunkter p ON p.riksmote = l.riksmote AND p.beteckning = l.beteckning AND p.punkt = l.punkt
			JOIN betankanden b ON b.riksmote = l.riksmote AND b.beteckning = l.beteckning
			WHERE l.motion_dok_id = m.dok_id AND p.beslutstyp <> '' AND b.systemdatum IS NOT NULL)`}, cond...)
		q = `SELECT ` + cols + `, false, ` + undecidedKeySQL + ` FROM motions m WHERE ` + strings.Join(cond, " AND ") +
			` ORDER BY ` + undecidedKeySQL + ` DESC, m.dok_id DESC LIMIT ` + arg(limit)
	}
	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.ListRow
	for rows.Next() {
		var row domain.ListRow
		var key time.Time
		m := &row.Motion
		if err := rows.Scan(&m.DokID, &m.Riksmote, &m.Beteckning, &m.Title, &m.Subtyp, &m.Date, &m.Organ, &m.Tilldelat, &m.Status, &row.Decided, &key); err != nil {
			return nil, err
		}
		row.Key = key
		row.Cursor = domain.ListCursor{Decided: row.Decided, Key: key, DokID: m.DokID}
		row.Decision = map[int][]string{}
		out = append(out, row)
	}
	return out, rows.Err()
}

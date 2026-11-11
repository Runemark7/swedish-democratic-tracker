package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"riksdagskollen/internal/loften/domain"
	votes "riksdagskollen/internal/votes/domain"
)

type Repository struct{ db *pgxpool.Pool }

func NewRepository(db *pgxpool.Pool) *Repository { return &Repository{db: db} }

// publicFilter is the single place that decides what is public.
const publicFilter = `visibility IN ('approved', 'archived')`

const loefteCols = `id, party, valmanifest_id, text, coalesce(quote,''), coalesce(locator,''),
	coalesce(source_url,''), coalesce(source_document,''), coalesce(topic,''), utskott,
	mandate_period, visibility, created_at`

func scanLoefte(row pgx.Row) (*domain.Loefte, error) {
	var l domain.Loefte
	var vis string
	if err := row.Scan(&l.ID, &l.Party, &l.ValmanifestID, &l.Text, &l.Quote, &l.Locator,
		&l.SourceURL, &l.SourceDoc, &l.Topic, &l.Utskott, &l.MandatePeriod, &vis, &l.CreatedAt); err != nil {
		return nil, err
	}
	l.Visibility = domain.Visibility(vis)
	l.Archived = l.Visibility == domain.Archived
	if l.Utskott == nil {
		l.Utskott = []string{}
	}
	return &l, nil
}

func (r *Repository) ListPublic(ctx context.Context, party string) ([]*domain.Loefte, error) {
	rows, err := r.db.Query(ctx,
		`SELECT `+loefteCols+` FROM loften WHERE `+publicFilter+` AND ($1 = '' OR party = $1)
		 ORDER BY party, topic, id`, party)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []*domain.Loefte{}
	for rows.Next() {
		l, err := scanLoefte(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

func (r *Repository) GetPublic(ctx context.Context, id int) (*domain.Loefte, error) {
	l, err := scanLoefte(r.db.QueryRow(ctx,
		`SELECT `+loefteCols+` FROM loften WHERE id = $1 AND `+publicFilter, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return l, err
}

func (r *Repository) ListValmanifest(ctx context.Context, party string) ([]*domain.Valmanifest, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, party, election_year, title, published_on, url, coalesce(pdf_url,''), state
		 FROM valmanifest WHERE ($1 = '' OR party = $1)
		 ORDER BY election_year DESC, party, id`, party)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []*domain.Valmanifest{}
	for rows.Next() {
		var v domain.Valmanifest
		if err := rows.Scan(&v.ID, &v.Party, &v.ElectionYear, &v.Title, &v.PublishedOn, &v.URL, &v.PDFURL, &v.State); err != nil {
			return nil, err
		}
		out = append(out, &v)
	}
	return out, rows.Err()
}

// confirmedSQL: the single place a koppling becomes public. Only state
// 'bekraftad' on a public Löfte is read, and the tally is the Löfte's own party.
const confirmedSQL = `
SELECT k.id, k.loefte_id, v.votering_id, v.beteckning, v.forslagspunkt, v.session, v.system_datum,
       k.direction, coalesce(k.decided_at, k.created_at),
       coalesce((SELECT bt.title FROM forslagspunkter f JOIN betankanden bt USING (riksmote, beteckning)
                  WHERE upper(f.votering_id) = upper(v.votering_id) ORDER BY f.riksmote DESC LIMIT 1), ''),
       count(*) FILTER (WHERE b.vote_result = 'Ja'), count(*) FILTER (WHERE b.vote_result = 'Nej'),
       count(*) FILTER (WHERE b.vote_result = 'Avstår'), count(*) FILTER (WHERE b.vote_result = 'Frånvarande')
  FROM kopplingar k
  JOIN loften l ON l.id = k.loefte_id AND l.visibility IN ('approved', 'archived')
  JOIN voteringar v ON v.id = k.votering_ref
  LEFT JOIN ballots b ON b.votering_ref = v.id AND b.party = l.party
 WHERE k.state = 'bekraftad' AND `

const confirmedGroup = ` GROUP BY k.id, v.id, l.id ORDER BY coalesce(k.decided_at, k.created_at), k.id`

func scanConfirmed(rows pgx.Rows) (*domain.ConfirmedKoppling, error) {
	var c domain.ConfirmedKoppling
	var ja, nej, avst, fr int
	if err := rows.Scan(&c.KopplingID, &c.LoefteID, &c.VoteringID, &c.Beteckning, &c.Forslagspunkt, &c.Session,
		&c.VoteDate, &c.Direction, &c.ConfirmedAt, &c.Betankande, &ja, &nej, &avst, &fr); err != nil {
		return nil, err
	}
	c.Tally = votes.Tally{Ja: ja, Nej: nej, Avstar: avst, Franvarande: fr}
	return &c, nil
}

func (r *Repository) ConfirmedKopplingar(ctx context.Context, ids []int) (map[int][]domain.ConfirmedKoppling, error) {
	out := map[int][]domain.ConfirmedKoppling{}
	if len(ids) == 0 {
		return out, nil
	}
	rows, err := r.db.Query(ctx, confirmedSQL+`k.loefte_id = ANY($1)`+confirmedGroup, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		c, err := scanConfirmed(rows)
		if err != nil {
			return nil, err
		}
		out[c.LoefteID] = append(out[c.LoefteID], *c)
	}
	return out, rows.Err()
}

func (r *Repository) ConfirmedKopplingByID(ctx context.Context, id int) (*domain.ConfirmedKoppling, error) {
	rows, err := r.db.Query(ctx, confirmedSQL+`k.id = $1`+confirmedGroup, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, rows.Err()
	}
	return scanConfirmed(rows)
}

func (r *Repository) Windows(ctx context.Context) (map[string][]string, error) {
	rows, err := r.db.Query(ctx, `SELECT code, riksmoten FROM mandate_periods`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string][]string{}
	for rows.Next() {
		var code string
		var rm []string
		if err := rows.Scan(&code, &rm); err != nil {
			return nil, err
		}
		out[code] = rm
	}
	return out, rows.Err()
}

// LoefteForBetankanden: kopplingar are the small side, so start there and
// reach the betänkande through the förslagspunkt that names the votering.
func (r *Repository) LoefteForBetankanden(ctx context.Context, keys []string) (map[string][]domain.LoefteRef, error) {
	out := map[string][]domain.LoefteRef{}
	if len(keys) == 0 {
		return out, nil
	}
	rows, err := r.db.Query(ctx, `
SELECT DISTINCT f.riksmote || ':' || f.beteckning AS key, l.party, l.id, l.text
  FROM kopplingar k
  JOIN loften l ON l.id = k.loefte_id AND l.`+publicFilter+`
  JOIN voteringar v ON v.id = k.votering_ref
  JOIN forslagspunkter f ON f.votering_id <> '' AND upper(f.votering_id) = upper(v.votering_id)
 WHERE k.state = 'bekraftad' AND f.riksmote || ':' || f.beteckning = ANY($1)
 ORDER BY key, l.party, l.id`, keys)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var key string
		var ref domain.LoefteRef
		if err := rows.Scan(&key, &ref.Party, &ref.ID, &ref.Text); err != nil {
			return nil, err
		}
		out[key] = append(out[key], ref)
	}
	return out, rows.Err()
}

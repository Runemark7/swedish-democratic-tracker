package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"riksdagskollen/internal/loften/domain"
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

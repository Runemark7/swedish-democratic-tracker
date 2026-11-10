// Package postgres is the SQL implementation of the motions repository.
package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"riksdagskollen/internal/motions/domain"
	"riksdagskollen/internal/motions/ports"
)

type Repository struct{ db *pgxpool.Pool }

func NewRepository(db *pgxpool.Pool) *Repository { return &Repository{db: db} }

func (r *Repository) Upsert(ctx context.Context, m *domain.Motion) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `
		INSERT INTO motions (dok_id, riksmote, beteckning, title, subtyp, motion_date,
		                     organ, tilldelat, status, systemdatum, behandlas_i, fetched_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11, now())
		ON CONFLICT (dok_id) DO UPDATE SET
			riksmote = EXCLUDED.riksmote, beteckning = EXCLUDED.beteckning,
			title = EXCLUDED.title, subtyp = EXCLUDED.subtyp,
			motion_date = EXCLUDED.motion_date, organ = EXCLUDED.organ,
			tilldelat = EXCLUDED.tilldelat, status = EXCLUDED.status,
			systemdatum = EXCLUDED.systemdatum, behandlas_i = EXCLUDED.behandlas_i,
			fetched_at = now()`,
		m.DokID, m.Riksmote, m.Beteckning, m.Title, m.Subtyp, m.Date,
		m.Organ, m.Tilldelat, m.Status, m.SystemDatum, m.BehandlasI); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, `DELETE FROM motion_authors WHERE dok_id = $1`, m.DokID); err != nil {
		return err
	}
	for _, a := range m.Authors {
		if _, err := tx.Exec(ctx, `
			INSERT INTO motion_authors (dok_id, ordning, intressent_id, name, party, role)
			VALUES ($1,$2,$3,$4,$5,$6)`,
			m.DokID, a.Order, a.PoliticianID, a.Name, a.Party, a.Role); err != nil {
			return err
		}
	}

	if _, err := tx.Exec(ctx, `DELETE FROM motion_yrkanden WHERE dok_id = $1`, m.DokID); err != nil {
		return err
	}
	for _, y := range m.Yrkanden {
		if _, err := tx.Exec(ctx, `
			INSERT INTO motion_yrkanden (dok_id, nummer, wording, utskottet, kammaren, behandlas_i, behandlas_i_punkt)
			VALUES ($1,$2,$3,$4,$5,$6,$7)`,
			m.DokID, y.Number, y.Wording, y.UtskottProposal, y.ChamberDecision, y.BehandlasI, y.BehandlasIPunkt); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *Repository) Get(ctx context.Context, riksmote, beteckning string) (*domain.Motion, error) {
	var m domain.Motion
	err := r.db.QueryRow(ctx, `
		SELECT dok_id, riksmote, beteckning, title, subtyp, motion_date, organ, tilldelat,
		       status, systemdatum, behandlas_i
		FROM motions WHERE riksmote = $1 AND beteckning = $2`, riksmote, beteckning).
		Scan(&m.DokID, &m.Riksmote, &m.Beteckning, &m.Title, &m.Subtyp, &m.Date, &m.Organ,
			&m.Tilldelat, &m.Status, &m.SystemDatum, &m.BehandlasI)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ports.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	rows, err := r.db.Query(ctx, `
		SELECT ordning, intressent_id, name, party, role
		FROM motion_authors WHERE dok_id = $1 ORDER BY ordning`, m.DokID)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var a domain.Author
		if err := rows.Scan(&a.Order, &a.PoliticianID, &a.Name, &a.Party, &a.Role); err != nil {
			rows.Close()
			return nil, err
		}
		m.Authors = append(m.Authors, a)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}

	yrows, err := r.db.Query(ctx, `
		SELECT nummer, wording, utskottet, kammaren, behandlas_i, behandlas_i_punkt
		FROM motion_yrkanden WHERE dok_id = $1 ORDER BY nummer`, m.DokID)
	if err != nil {
		return nil, err
	}
	defer yrows.Close()
	for yrows.Next() {
		var y domain.Yrkande
		if err := yrows.Scan(&y.Number, &y.Wording, &y.UtskottProposal, &y.ChamberDecision,
			&y.BehandlasI, &y.BehandlasIPunkt); err != nil {
			return nil, err
		}
		m.Yrkanden = append(m.Yrkanden, y)
	}
	return &m, yrows.Err()
}

func (r *Repository) StoredVersions(ctx context.Context, riksmote string) (map[string]time.Time, error) {
	rows, err := r.db.Query(ctx, `SELECT dok_id, systemdatum FROM motions WHERE riksmote = $1`, riksmote)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]time.Time{}
	for rows.Next() {
		var id string
		var t time.Time
		if err := rows.Scan(&id, &t); err != nil {
			return nil, err
		}
		out[id] = t
	}
	return out, rows.Err()
}

func (r *Repository) UpsertCoverage(ctx context.Context, riksmote, source string, expected int) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO motion_coverage (riksmote, source, expected, last_enumerated_at)
		VALUES ($1,$2,$3, now())
		ON CONFLICT (riksmote, source) DO UPDATE
		SET expected = EXCLUDED.expected, last_enumerated_at = now()`, riksmote, source, expected)
	return err
}

func (r *Repository) Coverage(ctx context.Context) ([]domain.Coverage, error) {
	rows, err := r.db.Query(ctx, `
		SELECT c.riksmote, c.source, c.expected,
		       (SELECT count(*) FROM motions m WHERE m.riksmote = c.riksmote),
		       c.last_enumerated_at
		FROM motion_coverage c
		ORDER BY c.riksmote DESC, c.source`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Coverage
	for rows.Next() {
		var c domain.Coverage
		if err := rows.Scan(&c.Riksmote, &c.Source, &c.Expected, &c.Held, &c.LastEnumeratedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

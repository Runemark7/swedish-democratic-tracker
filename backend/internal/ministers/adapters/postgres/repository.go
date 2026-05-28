package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"riksdagskollen/internal/ministers/domain"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

const ministerCols = `id::text, name, title, department, department_code, party, politician_id, photo_url, bio, active, created_at`

func (r *Repository) ListActive(ctx context.Context) ([]*domain.Minister, error) {
	rows, err := r.db.Query(ctx,
		`SELECT `+ministerCols+` FROM ministers WHERE active = true ORDER BY department, name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMinisters(rows)
}

func (r *Repository) GetByID(ctx context.Context, id string) (*domain.Minister, error) {
	m := &domain.Minister{}
	err := r.db.QueryRow(ctx,
		`SELECT `+ministerCols+` FROM ministers WHERE id::text = $1`, id).
		Scan(&m.ID, &m.Name, &m.Title, &m.Department, &m.DepartmentCode,
			&m.Party, &m.PoliticianID, &m.PhotoURL, &m.Bio, &m.Active, &m.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return m, err
}

func (r *Repository) ListProposalsByDept(ctx context.Context, deptCode string, limit int) ([]*domain.Proposal, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, title, department_code, riksdag_year, published_at, url
		 FROM proposals WHERE department_code = $1
		 ORDER BY published_at DESC NULLS LAST LIMIT $2`,
		deptCode, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*domain.Proposal
	for rows.Next() {
		p := &domain.Proposal{}
		if err := rows.Scan(&p.ID, &p.Title, &p.DepartmentCode, &p.RiksdagYear, &p.PublishedAt, &p.URL); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *Repository) UpsertProposal(ctx context.Context, p *domain.Proposal) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO proposals (id, title, department_code, riksdag_year, published_at, url)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 ON CONFLICT (id) DO UPDATE SET title = EXCLUDED.title, published_at = EXCLUDED.published_at`,
		p.ID, p.Title, p.DepartmentCode, p.RiksdagYear, p.PublishedAt, p.URL)
	return err
}

func scanMinisters(rows pgx.Rows) ([]*domain.Minister, error) {
	var out []*domain.Minister
	for rows.Next() {
		m := &domain.Minister{}
		if err := rows.Scan(&m.ID, &m.Name, &m.Title, &m.Department, &m.DepartmentCode,
			&m.Party, &m.PoliticianID, &m.PhotoURL, &m.Bio, &m.Active, &m.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

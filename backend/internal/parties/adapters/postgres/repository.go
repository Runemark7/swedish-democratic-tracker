package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"riksdagskollen/internal/parties/domain"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

const selectCols = `code, name, founded_year, ideology, color_hex, text_hex, active, website_url, created_at`

func (r *Repository) ListAll(ctx context.Context) ([]*domain.Party, error) {
	rows, err := r.db.Query(ctx, `SELECT `+selectCols+` FROM parties ORDER BY code`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*domain.Party
	for rows.Next() {
		p := &domain.Party{}
		if err := rows.Scan(&p.Code, &p.Name, &p.FoundedYear, &p.Ideology,
			&p.ColorHex, &p.TextHex, &p.Active, &p.WebsiteURL, &p.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *Repository) GetByCode(ctx context.Context, code string) (*domain.Party, error) {
	p := &domain.Party{}
	err := r.db.QueryRow(ctx, `SELECT `+selectCols+` FROM parties WHERE code = $1`, code).
		Scan(&p.Code, &p.Name, &p.FoundedYear, &p.Ideology,
			&p.ColorHex, &p.TextHex, &p.Active, &p.WebsiteURL, &p.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return p, err
}

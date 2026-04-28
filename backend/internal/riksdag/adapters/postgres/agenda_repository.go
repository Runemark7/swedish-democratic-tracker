package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"riksdagskollen/internal/riksdag/domain"
)

type AgendaRepository struct {
	db *pgxpool.Pool
}

func NewAgendaRepository(db *pgxpool.Pool) *AgendaRepository {
	return &AgendaRepository{db: db}
}

func (r *AgendaRepository) ListAgenda(ctx context.Context) ([]domain.AgendaItem, error) {
	const q = `
		SELECT id, title, description, source, status
		FROM riksdag_agenda
		ORDER BY sort_order
	`
	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.AgendaItem
	for rows.Next() {
		var a domain.AgendaItem
		if err := rows.Scan(&a.ID, &a.Title, &a.Description, &a.Source, &a.Status); err != nil {
			return nil, err
		}
		items = append(items, a)
	}
	return items, rows.Err()
}

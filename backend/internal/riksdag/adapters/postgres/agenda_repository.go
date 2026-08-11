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

// ListAgenda returns the government documents we have registered.
//
// There is no single-item read: an item is a title and a link to a primary
// source, so a detail page would show nothing the list does not already carry.
// The endpoint and its page were retired with the paraphrase they existed for.
func (r *AgendaRepository) ListAgenda(ctx context.Context) ([]domain.AgendaItem, error) {
	const q = `
		SELECT id, title, source, issuer, url, to_char(published, 'YYYY-MM-DD')
		FROM riksdag_agenda
		ORDER BY sort_order
	`
	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []domain.AgendaItem{}
	for rows.Next() {
		var a domain.AgendaItem
		if err := rows.Scan(&a.ID, &a.Title, &a.Source, &a.Issuer, &a.URL, &a.Published); err != nil {
			return nil, err
		}
		items = append(items, a)
	}
	return items, rows.Err()
}

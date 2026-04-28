package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"riksdagskollen/internal/riksdag/domain"
)

type GovRepository struct {
	db *pgxpool.Pool
}

func NewGovRepository(db *pgxpool.Pool) *GovRepository {
	return &GovRepository{db: db}
}

func (r *GovRepository) GetCurrentGovernment(ctx context.Context) (*domain.Government, error) {
	const govQ = `
		SELECT type_label, valid_from
		FROM riksdag_government
		WHERE valid_to IS NULL
		ORDER BY valid_from DESC
		LIMIT 1
	`
	var gov domain.Government
	if err := r.db.QueryRow(ctx, govQ).Scan(&gov.TypeLabel, &gov.ValidFrom); err != nil {
		return nil, err
	}

	const partiesQ = `
		SELECT name, short, seats, color, role, sort_order
		FROM riksdag_government_parties gp
		JOIN riksdag_government g ON g.id = gp.government_id
		WHERE g.valid_to IS NULL
		ORDER BY sort_order
	`
	rows, err := r.db.Query(ctx, partiesQ)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var p domain.GovernmentParty
		if err := rows.Scan(&p.Name, &p.Short, &p.Seats, &p.Color, &p.Role, &p.SortOrder); err != nil {
			return nil, err
		}
		switch p.Role {
		case "governing":
			gov.Parties = append(gov.Parties, p)
		case "support":
			gov.Support = append(gov.Support, p)
		case "opposition":
			gov.Opposition = append(gov.Opposition, p)
		}
	}
	return &gov, rows.Err()
}

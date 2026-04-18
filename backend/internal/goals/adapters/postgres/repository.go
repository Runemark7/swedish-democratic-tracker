package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"riksdagskollen/internal/goals/domain"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

const selectCols = `id, party, goal_text, topic, specificity, source_document, keywords, relevant_committees, created_at`

func (r *Repository) Create(ctx context.Context, g *domain.Goal) error {
	const q = `INSERT INTO party_goals (party, goal_text, topic, specificity, source_document, keywords, relevant_committees)
		VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id, created_at`
	return r.db.QueryRow(ctx, q,
		g.Party, g.GoalText, g.Topic, string(g.Specificity), g.SourceDocument,
		g.Keywords, g.RelevantCommittees,
	).Scan(&g.ID, &g.CreatedAt)
}

func (r *Repository) GetByID(ctx context.Context, id int) (*domain.Goal, error) {
	q := "SELECT " + selectCols + " FROM party_goals WHERE id = $1"
	g, err := scanGoal(r.db.QueryRow(ctx, q, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return g, err
}

func (r *Repository) ListAll(ctx context.Context) ([]*domain.Goal, error) {
	q := "SELECT " + selectCols + " FROM party_goals ORDER BY party, topic, id"
	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanGoals(rows)
}

func (r *Repository) ListByParty(ctx context.Context, party string) ([]*domain.Goal, error) {
	q := "SELECT " + selectCols + " FROM party_goals WHERE party = $1 ORDER BY topic, id"
	rows, err := r.db.Query(ctx, q, party)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanGoals(rows)
}

func (r *Repository) ListByTopic(ctx context.Context, party, topic string) ([]*domain.Goal, error) {
	q := "SELECT " + selectCols + " FROM party_goals WHERE party = $1 AND topic = $2 ORDER BY id"
	rows, err := r.db.Query(ctx, q, party, topic)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanGoals(rows)
}

type scanner interface{ Scan(dest ...any) error }

func scanGoal(s scanner) (*domain.Goal, error) {
	var g domain.Goal
	var specificity string
	err := s.Scan(&g.ID, &g.Party, &g.GoalText, &g.Topic, &specificity,
		&g.SourceDocument, &g.Keywords, &g.RelevantCommittees, &g.CreatedAt)
	g.Specificity = domain.Specificity(specificity)
	return &g, err
}

func scanGoals(rows pgx.Rows) ([]*domain.Goal, error) {
	var gg []*domain.Goal
	for rows.Next() {
		g, err := scanGoal(rows)
		if err != nil {
			return nil, err
		}
		gg = append(gg, g)
	}
	return gg, rows.Err()
}

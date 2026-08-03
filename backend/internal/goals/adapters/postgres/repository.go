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

const selectCols = `id, party, goal_text, topic, specificity, source_document, source_url, source_quote, keywords, relevant_committees, created_at`

// recordViewFilter excludes 2026 campaign material from the record view. The
// 2022-2026 votes cannot speak to a promise made in 2026, and showing the two
// side by side reads as "promised this, didn't do it" about a brand-new promise.
//
// It is also what keeps party comparison honest: only one party currently has
// 2026 material on the site, so including it would compare one party's current
// platform against everyone else's four-year-old manifesto.
//
// A 2026 platform view ships only when all eight parties are represented
// (wayfinder #83) and does not exist yet. Goals with a NULL cycle are standing
// party programmes and stay in the record view.
const recordViewFilter = `(election_cycle IS NULL OR election_cycle <> '2026')`

func (r *Repository) Create(ctx context.Context, g *domain.Goal) error {
	const q = `INSERT INTO party_goals (party, goal_text, topic, specificity, source_document, source_url, source_quote, keywords, relevant_committees)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id, created_at`
	return r.db.QueryRow(ctx, q,
		g.Party, g.GoalText, g.Topic, string(g.Specificity), g.SourceDocument,
		nullIfEmpty(g.SourceURL), nullIfEmpty(g.SourceQuote),
		g.Keywords, g.RelevantCommittees,
	).Scan(&g.ID, &g.CreatedAt)
}

func (r *Repository) GetByID(ctx context.Context, id int) (*domain.Goal, error) {
	// Filtered too: a goal excluded from the record view must not be
	// addressable within it, or a direct link would render a 2022-2026 vote
	// analysis against a promise made in 2026.
	q := "SELECT " + selectCols + " FROM party_goals WHERE id = $1 AND " + recordViewFilter
	g, err := scanGoal(r.db.QueryRow(ctx, q, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return g, err
}

func (r *Repository) ListAll(ctx context.Context) ([]*domain.Goal, error) {
	q := "SELECT " + selectCols + " FROM party_goals WHERE " + recordViewFilter + " ORDER BY party, topic, id"
	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanGoals(rows)
}

func (r *Repository) ListByParty(ctx context.Context, party string) ([]*domain.Goal, error) {
	q := "SELECT " + selectCols + " FROM party_goals WHERE party = $1 AND " + recordViewFilter + " ORDER BY topic, id"
	rows, err := r.db.Query(ctx, q, party)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanGoals(rows)
}

func (r *Repository) ListByTopic(ctx context.Context, party, topic string) ([]*domain.Goal, error) {
	q := "SELECT " + selectCols + " FROM party_goals WHERE party = $1 AND topic = $2 AND " + recordViewFilter + " ORDER BY id"
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
	var srcURL, srcQuote *string
	err := s.Scan(&g.ID, &g.Party, &g.GoalText, &g.Topic, &specificity,
		&g.SourceDocument, &srcURL, &srcQuote, &g.Keywords, &g.RelevantCommittees, &g.CreatedAt)
	g.Specificity = domain.Specificity(specificity)
	if srcURL != nil {
		g.SourceURL = *srcURL
	}
	if srcQuote != nil {
		g.SourceQuote = *srcQuote
	}
	return &g, err
}

func nullIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
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

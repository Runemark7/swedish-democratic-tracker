package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"riksdagskollen/internal/promises/domain"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

const selectCols = `id, politician_id, speech_id, promise_text, topic, specificity, keywords, extracted_at`

func (r *Repository) Create(ctx context.Context, p *domain.Promise) error {
	const q = `INSERT INTO promises (politician_id, speech_id, promise_text, topic, specificity, keywords)
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, extracted_at`
	return r.db.QueryRow(ctx, q,
		p.PoliticianID, p.SpeechID, p.PromiseText, p.Topic, string(p.Specificity), p.Keywords,
	).Scan(&p.ID, &p.ExtractedAt)
}

func (r *Repository) GetByID(ctx context.Context, id int) (*domain.Promise, error) {
	q := "SELECT " + selectCols + " FROM promises WHERE id = $1"
	p, err := scanPromise(r.db.QueryRow(ctx, q, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return p, err
}

func (r *Repository) ListByPolitician(ctx context.Context, politicianID, topic string) ([]*domain.Promise, error) {
	args := []any{politicianID}
	where := "WHERE politician_id = $1"
	if topic != "" {
		where += " AND topic = $2"
		args = append(args, topic)
	}
	q := "SELECT " + selectCols + " FROM promises " + where + " ORDER BY extracted_at DESC"
	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pp []*domain.Promise
	for rows.Next() {
		p, err := scanPromise(rows)
		if err != nil {
			return nil, err
		}
		pp = append(pp, p)
	}
	return pp, rows.Err()
}

type scanner interface{ Scan(dest ...any) error }

func scanPromise(s scanner) (*domain.Promise, error) {
	var p domain.Promise
	var spec string
	err := s.Scan(&p.ID, &p.PoliticianID, &p.SpeechID, &p.PromiseText,
		&p.Topic, &spec, &p.Keywords, &p.ExtractedAt)
	p.Specificity = domain.Specificity(spec)
	return &p, err
}

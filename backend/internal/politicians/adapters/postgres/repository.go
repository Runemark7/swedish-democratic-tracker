package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"riksdagskollen/internal/politicians/domain"
	"riksdagskollen/internal/politicians/ports"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetByID(ctx context.Context, id string) (*domain.Politician, error) {
	const q = `
		SELECT intressent_id, first_name, last_name, party, constituency, image_url, is_active, updated_at,
		       email, member_since, committees
		FROM politicians
		WHERE intressent_id = $1`

	var p domain.Politician
	var constituency, imageURL, email *string
	var since *int
	var committees []byte
	err := r.db.QueryRow(ctx, q, id).Scan(
		&p.IntressentID, &p.FirstName, &p.LastName, &p.Party,
		&constituency, &imageURL, &p.IsActive, &p.UpdatedAt,
		&email, &since, &committees)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	if constituency != nil {
		p.Constituency = *constituency
	}
	if imageURL != nil {
		p.ImageURL = *imageURL
	}
	if email != nil {
		p.Email = *email
	}
	if since != nil {
		p.MemberSince = *since
	}
	if len(committees) > 0 {
		if err := json.Unmarshal(committees, &p.Committees); err != nil {
			return nil, fmt.Errorf("decode committees for %s: %w", id, err)
		}
	}
	return &p, nil
}

func (r *Repository) GetMany(ctx context.Context, ids []string) (map[string]*domain.Politician, error) {
	out := map[string]*domain.Politician{}
	if len(ids) == 0 {
		return out, nil
	}
	rows, err := r.db.Query(ctx, `
		SELECT intressent_id, first_name, last_name, party, constituency, image_url, is_active, updated_at
		FROM politicians
		WHERE intressent_id = ANY($1)`, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		p, err := scanPolitician(rows)
		if err != nil {
			return nil, err
		}
		out[p.IntressentID] = p
	}
	return out, rows.Err()
}

func (r *Repository) List(ctx context.Context, f ports.ListFilter) (ports.ListResult, error) {
	args := []any{}
	where := "WHERE 1=1"
	n := 1

	if f.Party != "" {
		where += " AND party = " + ph(n)
		args = append(args, f.Party)
		n++
	}
	if f.ActiveOnly {
		where += " AND is_active = true"
	}

	// Count
	var total int
	countQ := "SELECT COUNT(*) FROM politicians " + where
	if err := r.db.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return ports.ListResult{}, err
	}

	// Page
	offset := (f.Page - 1) * f.PageSize
	args = append(args, f.PageSize, offset)
	listQ := `SELECT intressent_id, first_name, last_name, party, constituency, image_url, is_active, updated_at
		FROM politicians ` + where + fmt.Sprintf(` ORDER BY last_name, first_name LIMIT $%d OFFSET $%d`, n, n+1)

	rows, err := r.db.Query(ctx, listQ, args...)
	if err != nil {
		return ports.ListResult{}, err
	}
	defer rows.Close()

	var pp []*domain.Politician
	for rows.Next() {
		p, err := scanPolitician(rows)
		if err != nil {
			return ports.ListResult{}, err
		}
		pp = append(pp, p)
	}
	return ports.ListResult{Politicians: pp, Total: total}, rows.Err()
}

func (r *Repository) UpsertMany(ctx context.Context, pp []*domain.Politician) error {
	const q = `
		INSERT INTO politicians (intressent_id, first_name, last_name, party, constituency, image_url, is_active, updated_at,
			email, member_since, committees)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (intressent_id) DO UPDATE SET
			first_name    = EXCLUDED.first_name,
			last_name     = EXCLUDED.last_name,
			party         = EXCLUDED.party,
			constituency  = EXCLUDED.constituency,
			image_url     = EXCLUDED.image_url,
			is_active     = EXCLUDED.is_active,
			updated_at    = EXCLUDED.updated_at,
			email         = EXCLUDED.email,
			member_since  = EXCLUDED.member_since,
			committees    = EXCLUDED.committees`

	batch := &pgx.Batch{}
	for _, p := range pp {
		var email *string
		if p.Email != "" {
			email = &p.Email
		}
		var since *int
		if p.MemberSince > 0 {
			since = &p.MemberSince
		}
		var committees []byte
		if len(p.Committees) > 0 {
			var err error
			if committees, err = json.Marshal(p.Committees); err != nil {
				return err
			}
		}
		batch.Queue(q, p.IntressentID, p.FirstName, p.LastName, p.Party,
			p.Constituency, p.ImageURL, p.IsActive, time.Now(), email, since, committees)
	}
	br := r.db.SendBatch(ctx, batch)
	defer br.Close()

	for range pp {
		if _, err := br.Exec(); err != nil {
			return err
		}
	}
	return br.Close()
}

type scanner interface {
	Scan(dest ...any) error
}

func scanPolitician(s scanner) (*domain.Politician, error) {
	var p domain.Politician
	var constituency, imageURL *string
	err := s.Scan(
		&p.IntressentID, &p.FirstName, &p.LastName, &p.Party,
		&constituency, &imageURL, &p.IsActive, &p.UpdatedAt,
	)
	if constituency != nil {
		p.Constituency = *constituency
	}
	if imageURL != nil {
		p.ImageURL = *imageURL
	}
	return &p, err
}

func ph(n int) string { return fmt.Sprintf("$%d", n) }

// Missing returns the ids that have no politicians row.
func (r *Repository) Missing(ctx context.Context, ids []string) ([]string, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	rows, err := r.db.Query(ctx, `
		SELECT id FROM unnest($1::text[]) AS id
		WHERE NOT EXISTS (SELECT 1 FROM politicians WHERE intressent_id = id)
		ORDER BY id`, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

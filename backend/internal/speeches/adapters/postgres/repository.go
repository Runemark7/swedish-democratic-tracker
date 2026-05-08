package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"riksdagskollen/internal/speeches/domain"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

const selectCols = `id, dok_id, anforande_nummer, politician_id, party, date, topic_heading, speech_text, related_dok_id, ai_processed, created_at`

func (r *Repository) GetByID(ctx context.Context, id int) (*domain.Speech, error) {
	q := "SELECT " + selectCols + " FROM speeches WHERE id = $1"
	row := r.db.QueryRow(ctx, q, id)
	s, err := scanSpeech(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return s, err
}

func (r *Repository) ListByPolitician(ctx context.Context, politicianID string) ([]*domain.Speech, error) {
	q := "SELECT " + selectCols + " FROM speeches WHERE politician_id = $1 ORDER BY date DESC"
	rows, err := r.db.Query(ctx, q, politicianID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSpeeches(rows)
}

func (r *Repository) UpsertMany(ctx context.Context, ss []*domain.Speech) error {
	const q = `
		INSERT INTO speeches (dok_id, anforande_nummer, politician_id, party, date, topic_heading, speech_text, related_dok_id)
		SELECT $1, $2, $3, $4, $5, $6, $7, $8
		WHERE EXISTS (SELECT 1 FROM politicians WHERE intressent_id = $3)
		ON CONFLICT (dok_id, anforande_nummer) DO UPDATE SET
			topic_heading = EXCLUDED.topic_heading,
			speech_text   = EXCLUDED.speech_text`

	batch := &pgx.Batch{}
	for _, s := range ss {
		batch.Queue(q, s.DokID, s.AnforandeNummer, s.PoliticianID, s.Party,
			s.Date, s.TopicHeading, s.SpeechText, s.RelatedDokID)
	}
	br := r.db.SendBatch(ctx, batch)
	defer br.Close()
	for range ss {
		if _, err := br.Exec(); err != nil {
			return err
		}
	}
	return br.Close()
}

func (r *Repository) ListUnprocessed(ctx context.Context, limit int) ([]*domain.Speech, error) {
	q := "SELECT " + selectCols + " FROM speeches WHERE ai_processed = false ORDER BY date DESC LIMIT $1"
	rows, err := r.db.Query(ctx, q, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSpeeches(rows)
}

func (r *Repository) ListRecent(ctx context.Context, limit int) ([]*domain.Speech, error) {
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	q := "SELECT " + selectCols + " FROM speeches ORDER BY date DESC, id DESC LIMIT $1"
	rows, err := r.db.Query(ctx, q, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSpeeches(rows)
}

func (r *Repository) ListMissingText(ctx context.Context, limit int) ([]*domain.Speech, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	q := "SELECT " + selectCols + " FROM speeches " +
		"WHERE speech_text IS NULL OR speech_text = '' " +
		"ORDER BY date DESC, id DESC LIMIT $1"
	rows, err := r.db.Query(ctx, q, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSpeeches(rows)
}

func (r *Repository) UpdateText(ctx context.Context, id int, text string) error {
	_, err := r.db.Exec(ctx, "UPDATE speeches SET speech_text = $1 WHERE id = $2", text, id)
	return err
}

func (r *Repository) MarkProcessed(ctx context.Context, id int) error {
	_, err := r.db.Exec(ctx, "UPDATE speeches SET ai_processed = true WHERE id = $1", id)
	return err
}

type scanner interface{ Scan(dest ...any) error }

func scanSpeech(s scanner) (*domain.Speech, error) {
	var sp domain.Speech
	var anforandeNummer, topicHeading, speechText, relatedDokID *string
	err := s.Scan(&sp.ID, &sp.DokID, &anforandeNummer, &sp.PoliticianID,
		&sp.Party, &sp.Date, &topicHeading, &speechText, &relatedDokID,
		&sp.AIProcessed, &sp.CreatedAt)
	if anforandeNummer != nil {
		sp.AnforandeNummer = *anforandeNummer
	}
	if topicHeading != nil {
		sp.TopicHeading = *topicHeading
	}
	if speechText != nil {
		sp.SpeechText = *speechText
	}
	if relatedDokID != nil {
		sp.RelatedDokID = *relatedDokID
	}
	return &sp, err
}

func scanSpeeches(rows pgx.Rows) ([]*domain.Speech, error) {
	var ss []*domain.Speech
	for rows.Next() {
		s, err := scanSpeech(rows)
		if err != nil {
			return nil, err
		}
		ss = append(ss, s)
	}
	return ss, rows.Err()
}

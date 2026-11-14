package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"riksdagskollen/internal/speeches/domain"
	"riksdagskollen/internal/speeches/ports"
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
		INSERT INTO speeches (dok_id, anforande_nummer, politician_id, party, date, topic_heading, speech_text, related_dok_id, riksmote)
		SELECT $1, $2, $3, $4, $5, $6, $7, $8, NULLIF($9, '')
		WHERE EXISTS (SELECT 1 FROM politicians WHERE intressent_id = $3)
		ON CONFLICT (dok_id, anforande_nummer) DO UPDATE SET
			topic_heading  = EXCLUDED.topic_heading,
			-- The list carries no text; a re-listed speech keeps the text it was given.
			speech_text    = COALESCE(NULLIF(EXCLUDED.speech_text, ''), speeches.speech_text),
			related_dok_id = COALESCE(NULLIF(EXCLUDED.related_dok_id, ''), speeches.related_dok_id),
			riksmote       = COALESCE(EXCLUDED.riksmote, speeches.riksmote)`

	batch := &pgx.Batch{}
	for _, s := range ss {
		batch.Queue(q, s.DokID, s.AnforandeNummer, s.PoliticianID, s.Party,
			s.Date, s.TopicHeading, s.SpeechText, s.RelatedDokID, s.Riksmote)
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

func (r *Repository) ListByDocument(ctx context.Context, dokID string) ([]*domain.Speech, error) {
	if dokID == "" {
		return []*domain.Speech{}, nil
	}
	q := "SELECT " + selectCols + " FROM speeches " +
		"WHERE related_dok_id = $1 " +
		"ORDER BY anforande_nummer::int ASC, id ASC"
	rows, err := r.db.Query(ctx, q, dokID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSpeeches(rows)
}

func (r *Repository) ListByParty(ctx context.Context, party string, limit int) ([]*domain.Speech, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if party == "" {
		return []*domain.Speech{}, nil
	}
	q := "SELECT " + selectCols + " FROM speeches " +
		"WHERE party = $1 " +
		"ORDER BY date DESC, id DESC LIMIT $2"
	rows, err := r.db.Query(ctx, q, party, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSpeeches(rows)
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

// DebateCandidates lists the stored betänkanden that have a debate date.
func (r *Repository) DebateCandidates(ctx context.Context) ([]domain.DebateCandidate, error) {
	rows, err := r.db.Query(ctx, `
		SELECT dok_id, title, debate_date
		FROM betankanden
		WHERE found AND dok_id <> '' AND debate_date IS NOT NULL`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.DebateCandidate
	for rows.Next() {
		var c domain.DebateCandidate
		var d time.Time
		if err := rows.Scan(&c.DokID, &c.Title, &d); err != nil {
			return nil, err
		}
		c.DebateDate = d
		out = append(out, c)
	}
	return out, rows.Err()
}

// ListForLinking reads at most limit speeches with id > afterID, in id order.
// A riksmote restricts it to that riksmöte's speeches.
func (r *Repository) ListForLinking(ctx context.Context, riksmote string, afterID, limit int) ([]ports.LinkRow, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, coalesce(related_dok_id, ''), coalesce(topic_heading, ''), date,
		       coalesce(betankande_dok_id, ''), coalesce(betankande_link_rule, '')
		FROM speeches WHERE id > $1 AND ($3 = '' OR riksmote = $3) ORDER BY id LIMIT $2`, afterID, limit, riksmote)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ports.LinkRow
	for rows.Next() {
		var l ports.LinkRow
		if err := rows.Scan(&l.ID, &l.RelatedDokID, &l.TopicHeading, &l.Date, &l.CurrentDokID, &l.CurrentRule); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

// SetLinks writes the links in one batch. An empty DokID clears the link.
func (r *Repository) SetLinks(ctx context.Context, links []ports.LinkUpdate) error {
	batch := &pgx.Batch{}
	for _, l := range links {
		batch.Queue(`
			UPDATE speeches
			SET betankande_dok_id = NULLIF($2, ''), betankande_link_rule = NULLIF($3, ''), betankande_linked_at = now()
			WHERE id = $1`, l.ID, l.DokID, l.Rule)
	}
	br := r.db.SendBatch(ctx, batch)
	defer br.Close()
	for range links {
		if _, err := br.Exec(); err != nil {
			return err
		}
	}
	return br.Close()
}

// MissingTextIn lists up to limit speeches of the riksmöte with no text, oldest id first.
func (r *Repository) MissingTextIn(ctx context.Context, riksmote string, limit int) ([]*domain.Speech, error) {
	q := "SELECT " + selectCols + " FROM speeches " +
		"WHERE riksmote = $1 AND (speech_text IS NULL OR speech_text = '') ORDER BY id LIMIT $2"
	rows, err := r.db.Query(ctx, q, riksmote, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSpeeches(rows)
}

func (r *Repository) RiksmoteCounts(ctx context.Context, riksmote string) (domain.RiksmoteCounts, error) {
	var c domain.RiksmoteCounts
	err := r.db.QueryRow(ctx, `
		SELECT count(*),
		       count(betankande_dok_id),
		       count(*) FILTER (WHERE speech_text IS NULL OR speech_text = ''),
		       coalesce(sum(octet_length(speech_text)), 0)
		FROM speeches WHERE riksmote = $1`, riksmote).
		Scan(&c.Stored, &c.Linked, &c.MissingText, &c.TextBytes)
	return c, err
}

func (r *Repository) StorageSizes(ctx context.Context) (table, index int64, err error) {
	err = r.db.QueryRow(ctx, `SELECT pg_table_size('speeches'), pg_indexes_size('speeches')`).Scan(&table, &index)
	return table, index, err
}

func (r *Repository) SaveBackfill(ctx context.Context, run domain.BackfillRun) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO speech_backfill
			(riksmote, expected, text_bytes, rows_added, table_bytes_added, index_bytes_added,
			 politicians_created, unresolved_speakers, speakerless_speeches, completed_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $10, CASE WHEN $9 THEN now() END, now())
		ON CONFLICT (riksmote) DO UPDATE SET
			expected            = EXCLUDED.expected,
			text_bytes          = EXCLUDED.text_bytes,
			rows_added          = speech_backfill.rows_added + EXCLUDED.rows_added,
			table_bytes_added   = speech_backfill.table_bytes_added + EXCLUDED.table_bytes_added,
			index_bytes_added   = speech_backfill.index_bytes_added + EXCLUDED.index_bytes_added,
			politicians_created = speech_backfill.politicians_created + EXCLUDED.politicians_created,
			unresolved_speakers = EXCLUDED.unresolved_speakers,
			speakerless_speeches = EXCLUDED.speakerless_speeches,
			completed_at        = CASE WHEN $9 THEN coalesce(speech_backfill.completed_at, now()) END,
			updated_at          = now()`,
		run.Riksmote, run.Expected, run.TextBytes, run.RowsAdded, run.TableBytesAdded, run.IndexBytesAdded,
		run.PoliticiansCreated, run.Unresolved, run.Complete, run.Speakerless)
	return err
}

func (r *Repository) ListBackfill(ctx context.Context) ([]domain.Backfill, error) {
	rows, err := r.db.Query(ctx, `
		SELECT b.riksmote, b.expected, coalesce(c.stored, 0), coalesce(c.linked, 0),
		       b.text_bytes, b.rows_added, b.table_bytes_added, b.index_bytes_added,
		       b.politicians_created, b.unresolved_speakers, b.speakerless_speeches, b.completed_at
		FROM speech_backfill b
		LEFT JOIN (
			SELECT riksmote, count(*) AS stored, count(betankande_dok_id) AS linked
			FROM speeches WHERE riksmote IS NOT NULL GROUP BY riksmote
		) c ON c.riksmote = b.riksmote
		ORDER BY b.riksmote DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Backfill
	for rows.Next() {
		var b domain.Backfill
		if err := rows.Scan(&b.Riksmote, &b.Expected, &b.Stored, &b.Linked, &b.TextBytes, &b.RowsAdded,
			&b.TableBytesAdded, &b.IndexBytesAdded, &b.PoliticiansCreated, &b.Unresolved, &b.Speakerless, &b.CompletedAt); err != nil {
			return nil, err
		}
		b.Unlinked = b.Stored - b.Linked
		out = append(out, b)
	}
	return out, rows.Err()
}

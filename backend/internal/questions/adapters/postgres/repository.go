// Package postgres stores questions.
package postgres

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"riksdagskollen/internal/questions/domain"
	"riksdagskollen/internal/questions/ports"
)

type Repository struct{ db *pgxpool.Pool }

func NewRepository(db *pgxpool.Pool) *Repository { return &Repository{db: db} }

func (r *Repository) Upsert(ctx context.Context, q *domain.Question) error {
	refs := q.MotionRefs
	if refs == nil {
		refs = []string{}
	}
	_, err := r.db.Exec(ctx, `
		INSERT INTO questions (dok_id, kind, riksmote, beteckning, title, asked_date, status, systemdatum,
			asker_id, asker_name, asker_party, addressed_id, addressed_name, addressed_party,
			answerer_id, answerer_name, answerer_party, answer_dok_id, answer_date, body_text, motion_refs, fetched_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21, now())
		ON CONFLICT (dok_id) DO UPDATE SET
			title = EXCLUDED.title, asked_date = EXCLUDED.asked_date, status = EXCLUDED.status,
			systemdatum = EXCLUDED.systemdatum,
			asker_id = EXCLUDED.asker_id, asker_name = EXCLUDED.asker_name, asker_party = EXCLUDED.asker_party,
			addressed_id = EXCLUDED.addressed_id, addressed_name = EXCLUDED.addressed_name, addressed_party = EXCLUDED.addressed_party,
			answerer_id = EXCLUDED.answerer_id, answerer_name = EXCLUDED.answerer_name, answerer_party = EXCLUDED.answerer_party,
			answer_dok_id = EXCLUDED.answer_dok_id, answer_date = EXCLUDED.answer_date,
			body_text = EXCLUDED.body_text, motion_refs = EXCLUDED.motion_refs, fetched_at = now()`,
		q.DokID, q.Kind, q.Riksmote, q.Beteckning, q.Title, q.AskedDate, q.Status, q.SystemDatum,
		q.Asker.ID, q.Asker.Name, q.Asker.Party, q.Addressed.ID, q.Addressed.Name, q.Addressed.Party,
		q.Answerer.ID, q.Answerer.Name, q.Answerer.Party, q.AnswerDokID, q.AnswerDate, q.BodyText, refs)
	return err
}

func (r *Repository) StoredVersions(ctx context.Context, kind, rm string) (map[string]time.Time, error) {
	rows, err := r.db.Query(ctx, `SELECT dok_id, systemdatum FROM questions WHERE kind = $1 AND riksmote = $2`, kind, rm)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]time.Time{}
	for rows.Next() {
		var id string
		var t time.Time
		if err := rows.Scan(&id, &t); err != nil {
			return nil, err
		}
		out[id] = t
	}
	return out, rows.Err()
}

func (r *Repository) UpsertCoverage(ctx context.Context, rm, kind string, expected int) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO question_coverage (riksmote, kind, expected, last_enumerated_at) VALUES ($1,$2,$3, now())
		ON CONFLICT (riksmote, kind) DO UPDATE SET expected = EXCLUDED.expected, last_enumerated_at = now()`,
		rm, kind, expected)
	return err
}

func (r *Repository) Coverage(ctx context.Context) ([]domain.Coverage, error) {
	rows, err := r.db.Query(ctx, `
		SELECT c.riksmote, c.kind, c.expected,
		       (SELECT count(*) FROM questions q WHERE q.riksmote = c.riksmote AND q.kind = c.kind),
		       c.last_enumerated_at
		FROM question_coverage c ORDER BY c.riksmote DESC, c.kind`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Coverage{}
	for rows.Next() {
		var c domain.Coverage
		if err := rows.Scan(&c.Riksmote, &c.Kind, &c.Expected, &c.Held, &c.LastEnumeratedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *Repository) ForLinking(ctx context.Context, rm string) ([]domain.Question, error) {
	rows, err := r.db.Query(ctx, `
		SELECT dok_id, riksmote, beteckning, body_text, motion_refs FROM questions WHERE riksmote = $1 ORDER BY dok_id`, rm)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Question
	for rows.Next() {
		var q domain.Question
		if err := rows.Scan(&q.DokID, &q.Riksmote, &q.Beteckning, &q.BodyText, &q.MotionRefs); err != nil {
			return nil, err
		}
		out = append(out, q)
	}
	return out, rows.Err()
}

func (r *Repository) MotionDokID(ctx context.Context, rm, bet string) (string, bool, error) {
	var id string
	err := r.db.QueryRow(ctx, `SELECT dok_id FROM motions WHERE riksmote = $1 AND beteckning = $2`, rm, bet).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", false, nil
	}
	return id, err == nil, err
}

func (r *Repository) ReplaceLinks(ctx context.Context, qid string, links []domain.Link) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	keep := make([]string, 0, len(links))
	for _, l := range links {
		keep = append(keep, l.MotionDokID)
		if _, err := tx.Exec(ctx, `
			INSERT INTO question_motion_links (question_dok_id, motion_dok_id, rule) VALUES ($1,$2,$3)
			ON CONFLICT (question_dok_id, motion_dok_id) DO UPDATE SET rule = EXCLUDED.rule`,
			qid, l.MotionDokID, l.Rule); err != nil {
			return err
		}
	}
	if _, err := tx.Exec(ctx, `DELETE FROM question_motion_links WHERE question_dok_id = $1 AND NOT (motion_dok_id = ANY($2))`, qid, keep); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

const itemCols = `q.dok_id, q.kind, q.riksmote, q.beteckning, q.title, q.asked_date, q.status,
	q.asker_id, q.asker_name, q.asker_party, q.addressed_id, q.addressed_name, q.addressed_party,
	q.answerer_id, q.answerer_name, q.answerer_party, q.answer_date, q.answer_dok_id`

func scanItem(row pgx.Row, extra ...any) (domain.Item, error) {
	var it domain.Item
	dest := []any{&it.DokID, &it.Kind, &it.Riksmote, &it.Beteckning, &it.Title, &it.AskedDate, &it.Status,
		&it.Asker.ID, &it.Asker.Name, &it.Asker.Party, &it.Addressed.ID, &it.Addressed.Name, &it.Addressed.Party,
		&it.Answerer.ID, &it.Answerer.Name, &it.Answerer.Party, &it.AnswerDate, &it.AnswerDokID}
	err := row.Scan(append(dest, extra...)...)
	it.SourceURL = domain.SourceURL(it.DokID)
	return it, err
}

func (r *Repository) ByMember(ctx context.Context, id string, limit int) ([]domain.Item, error) {
	rows, err := r.db.Query(ctx, `
		SELECT `+itemCols+`, CASE WHEN q.asker_id = $1 THEN 'asker' ELSE 'answerer' END
		FROM questions q
		WHERE q.asker_id = $1 OR q.answerer_id = $1
		ORDER BY coalesce(q.answer_date, q.asked_date) DESC, q.riksmote DESC, q.kind, q.dok_id DESC
		LIMIT $2`, id, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Item{}
	for rows.Next() {
		var role string
		it, err := scanItem(rows, &role)
		if err != nil {
			return nil, err
		}
		it.Role = role
		out = append(out, it)
	}
	return out, rows.Err()
}

// publicLinks is every link that may be shown: the rule-made ones, and the
// prose citations the Redaktör confirmed in a Veckobatch (issue #163). A
// suggested, rejected or undone one is not in it.
const publicLinks = `(
	SELECT question_dok_id, motion_dok_id, rule FROM question_motion_links
	UNION ALL
	SELECT question_dok_id, motion_dok_id, 'redaktor-confirmed' FROM motion_kopplingar
	 WHERE kind = 'fraga' AND state = 'bekraftad'
)`

func (r *Repository) ByMotion(ctx context.Context, rm, bet string, limit int) ([]domain.Item, error) {
	rows, err := r.db.Query(ctx, `
		SELECT `+itemCols+`, l.rule
		FROM `+publicLinks+` l
		JOIN questions q ON q.dok_id = l.question_dok_id
		JOIN motions m ON m.dok_id = l.motion_dok_id
		WHERE m.riksmote = $1 AND m.beteckning = $2
		ORDER BY coalesce(q.answer_date, q.asked_date) DESC, q.dok_id DESC
		LIMIT $3`, rm, bet, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Item{}
	for rows.Next() {
		var rule string
		it, err := scanItem(rows, &rule)
		if err != nil {
			return nil, err
		}
		it.Rule = rule
		out = append(out, it)
	}
	return out, rows.Err()
}

func (r *Repository) Get(ctx context.Context, kind, rm, bet string) (*domain.Detail, error) {
	var text string
	it, err := scanItem(r.db.QueryRow(ctx, `SELECT `+itemCols+`, q.body_text FROM questions q
		WHERE q.kind = $1 AND q.riksmote = $2 AND q.beteckning = $3`, kind, rm, bet), &text)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ports.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	d := &domain.Detail{Item: it, Text: text, Debate: []domain.DebateSpeech{}, Motions: []domain.LinkedMotion{}}
	rows, err := r.db.Query(ctx, `
		SELECT m.dok_id, m.riksmote, m.beteckning, m.title, l.rule
		FROM `+publicLinks+` l JOIN motions m ON m.dok_id = l.motion_dok_id
		WHERE l.question_dok_id = $1 ORDER BY m.riksmote, m.beteckning`, it.DokID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var m domain.LinkedMotion
		if err := rows.Scan(&m.DokID, &m.Riksmote, &m.Beteckning, &m.Title, &m.Rule); err != nil {
			return nil, err
		}
		d.Motions = append(d.Motions, m)
	}
	return d, rows.Err()
}

// Debate finds an interpellation's anföranden among the stored speeches by
// the avsnittsrubrik ("Svar på interpellation 2022/23:397 om ...") and the day
// the interpellation was answered in the chamber.
func (r *Repository) Debate(ctx context.Context, d *domain.Detail) ([]domain.DebateSpeech, error) {
	out := []domain.DebateSpeech{}
	if d.AnswerDate == nil {
		return out, nil
	}
	prefix := "svar på interpellation " + strings.ToLower(d.Riksmote+":"+d.Beteckning)
	rows, err := r.db.Query(ctx, `
		SELECT s.id, s.dok_id, s.politician_id, p.first_name || ' ' || p.last_name, s.party, s.date,
		       left(coalesce(s.speech_text, ''), 1500)
		FROM speeches s JOIN politicians p ON p.intressent_id = s.politician_id
		WHERE s.date = $1
		  AND (lower(s.topic_heading) = $2 OR lower(s.topic_heading) LIKE $2 || ' %')
		ORDER BY s.dok_id, nullif(regexp_replace(s.anforande_nummer, '\D', '', 'g'), '')::int, s.id
		LIMIT 200`, d.AnswerDate, prefix)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var sp domain.DebateSpeech
		if err := rows.Scan(&sp.SpeechID, &sp.DokID, &sp.PoliticianID, &sp.Name, &sp.Party, &sp.Date, &sp.Excerpt); err != nil {
			return nil, err
		}
		out = append(out, sp)
	}
	return out, rows.Err()
}

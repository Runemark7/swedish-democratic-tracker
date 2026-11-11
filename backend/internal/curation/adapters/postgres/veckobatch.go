package postgres

import (
	"context"
	"errors"
	"sort"
	"time"

	"github.com/jackc/pgx/v5"

	"riksdagskollen/internal/curation/domain"
)

type querier interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func (s *store) LockBatches(ctx context.Context) error {
	_, err := s.tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended('veckobatch:create', 0))`)
	return err
}

// Rolled-over kopplingar (already in an earlier batch, never answered) come
// first so the weekly cap cannot starve them; then strongest first.
func (s *store) OpenSuggestions(ctx context.Context, max int) ([]domain.OpenSuggestion, error) {
	rows, err := s.tx.Query(ctx, `
		SELECT kid, mkid, rolled FROM (
		  SELECT k.id AS kid, 0 AS mkid, k.veckobatch_id IS NOT NULL AS rolled, k.veckobatch_id AS vb, k.confidence AS conf, k.id AS ord, 0 AS src
		    FROM kopplingar k WHERE k.state = 'foreslagen'
		  UNION ALL
		  SELECT 0, k.id, k.veckobatch_id IS NOT NULL, k.veckobatch_id, k.confidence, k.id, 1
		    FROM motion_kopplingar k WHERE k.state = 'foreslagen'
		) u
		ORDER BY rolled DESC, vb ASC NULLS LAST, conf DESC, src ASC, ord ASC
		LIMIT $1`, max)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.OpenSuggestion
	for rows.Next() {
		var o domain.OpenSuggestion
		if err := rows.Scan(&o.KopplingID, &o.MotionKopplingID, &o.RolledOver); err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	return out, rows.Err()
}

func (s *store) CloseOpenBatches(ctx context.Context) error {
	_, err := s.tx.Exec(ctx, `UPDATE veckobatch SET closed_at = now() WHERE closed_at IS NULL`)
	return err
}

func (s *store) InsertBatch(ctx context.Context) (int64, error) {
	var id int64
	err := s.tx.QueryRow(ctx, `INSERT INTO veckobatch DEFAULT VALUES RETURNING id`).Scan(&id)
	return id, err
}

func (s *store) InsertBatchItem(ctx context.Context, batchID int64, nr int, o domain.OpenSuggestion) error {
	if o.MotionKopplingID != 0 {
		if _, err := s.tx.Exec(ctx, `INSERT INTO veckobatch_items (veckobatch_id, nr, motion_koppling_id) VALUES ($1,$2,$3)`,
			batchID, nr, o.MotionKopplingID); err != nil {
			return err
		}
		_, err := s.tx.Exec(ctx, `UPDATE motion_kopplingar SET veckobatch_id = $1 WHERE id = $2`, batchID, o.MotionKopplingID)
		return err
	}
	if _, err := s.tx.Exec(ctx, `INSERT INTO veckobatch_items (veckobatch_id, nr, koppling_id) VALUES ($1,$2,$3)`,
		batchID, nr, o.KopplingID); err != nil {
		return err
	}
	_, err := s.tx.Exec(ctx, `UPDATE kopplingar SET veckobatch_id = $1 WHERE id = $2`, batchID, o.KopplingID)
	return err
}

func (s *store) BatchByID(ctx context.Context, id int64) (*domain.Batch, error) {
	return loadBatch(ctx, s.tx, id, true)
}

func (r *Repository) BatchByID(ctx context.Context, id int64) (*domain.Batch, error) {
	return loadBatch(ctx, r.db, id, false)
}

func (r *Repository) LatestBatch(ctx context.Context) (*domain.Batch, error) {
	var id int64
	err := r.db.QueryRow(ctx, `SELECT id FROM veckobatch ORDER BY id DESC LIMIT 1`).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return loadBatch(ctx, r.db, id, false)
}

func loadBatch(ctx context.Context, q querier, id int64, lock bool) (*domain.Batch, error) {
	lockB, lockK := "", ""
	if lock {
		lockB, lockK = " FOR UPDATE", " FOR UPDATE OF k"
	}
	b := domain.Batch{Items: []domain.BatchItem{}}
	err := q.QueryRow(ctx, `SELECT id, created_at, closed_at FROM veckobatch WHERE id=$1`+lockB, id).
		Scan(&b.ID, &b.CreatedAt, &b.ClosedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	rows, err := q.Query(ctx, `
		SELECT i.nr, k.id, k.state, EXISTS (SELECT 1 FROM veckobatch_items o WHERE o.koppling_id = k.id AND o.veckobatch_id < i.veckobatch_id),
		       l.id, l.party, l.text, coalesce(l.quote,''), coalesce(l.locator,''), coalesce(l.source_url,''),
		       v.votering_id, v.beteckning, v.forslagspunkt,
		       coalesce((SELECT bt.title FROM forslagspunkter f JOIN betankanden bt USING (riksmote, beteckning)
		                  WHERE upper(f.votering_id) = upper(v.votering_id) ORDER BY f.riksmote DESC LIMIT 1), ''),
		       v.system_datum, k.direction, k.confidence::float8, k.reasoning, i.answer
		  FROM veckobatch_items i
		  JOIN kopplingar k ON k.id = i.koppling_id
		  JOIN loften l ON l.id = k.loefte_id
		  JOIN voteringar v ON v.id = k.votering_ref
		 WHERE i.veckobatch_id = $1
		 ORDER BY i.nr`+lockK, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var it domain.BatchItem
		var d *time.Time
		if err := rows.Scan(&it.Nr, &it.KopplingID, &it.State, &it.RolledOver, &it.LoefteID, &it.Party, &it.LoefteText,
			&it.Quote, &it.Locator, &it.SourceURL, &it.VoteringID, &it.Beteckning, &it.Forslagspunkt, &it.Betankande,
			&d, &it.Direction, &it.Confidence, &it.Reasoning, &it.Answer); err != nil {
			return nil, err
		}
		if d != nil {
			ds := d.UTC().Format(time.RFC3339)
			it.Datum = &ds
		}
		it.Kind = domain.ItemLoefte
		b.Items = append(b.Items, it)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	rows.Close()
	mitems, err := loadMotionItems(ctx, q, id, lockK)
	if err != nil {
		return nil, err
	}
	b.Items = append(b.Items, mitems...)
	sort.Slice(b.Items, func(i, j int) bool { return b.Items[i].Nr < b.Items[j].Nr })
	return &b, nil
}

func (s *store) DecideItem(ctx context.Context, batchID int64, nr int, kopplingID int64, state, direction, answer string) error {
	if _, err := s.tx.Exec(ctx, `UPDATE kopplingar SET state=$2, direction=$3,
		decided_at = CASE WHEN $2 IN ('bekraftad','avvisad') THEN now() ELSE NULL END WHERE id=$1`,
		kopplingID, state, direction); err != nil {
		return err
	}
	_, err := s.tx.Exec(ctx, `UPDATE veckobatch_items SET answer = NULLIF($4,''),
		answered_at = CASE WHEN $4 = '' THEN NULL ELSE now() END
		WHERE veckobatch_id=$1 AND nr=$2 AND koppling_id=$3`, batchID, nr, kopplingID, answer)
	return err
}

func (s *store) CloseBatch(ctx context.Context, id int64) error {
	_, err := s.tx.Exec(ctx, `UPDATE veckobatch SET closed_at = now() WHERE id=$1 AND closed_at IS NULL`, id)
	return err
}

func (s *store) KopplingByID(ctx context.Context, id int64) (*domain.Koppling, error) {
	var k domain.Koppling
	err := s.tx.QueryRow(ctx, `SELECT k.id, k.loefte_id, v.votering_id, k.direction, k.state,
		k.confidence::float8, k.threshold::float8, k.reasoning, k.origin
		FROM kopplingar k JOIN voteringar v ON v.id = k.votering_ref
		WHERE k.id=$1 FOR UPDATE OF k`, id).
		Scan(&k.ID, &k.LoefteID, &k.VoteringID, &k.Direction, &k.State, &k.Confidence, &k.Threshold, &k.Reasoning, &k.Origin)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	// The item in the koppling's latest batch carries nr and answer.
	var bid int64
	var nr int
	var ans *string
	err = s.tx.QueryRow(ctx, `SELECT i.veckobatch_id, i.nr, i.answer FROM veckobatch_items i
		WHERE i.koppling_id=$1 ORDER BY i.veckobatch_id DESC LIMIT 1`, id).Scan(&bid, &nr, &ans)
	if err == nil {
		k.VeckobatchID, k.Nr = &bid, &nr
		if ans != nil {
			k.Answer = *ans
		}
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}
	return &k, nil
}

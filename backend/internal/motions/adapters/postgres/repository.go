// Package postgres is the SQL implementation of the motions repository.
package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"riksdagskollen/internal/motions/domain"
	"riksdagskollen/internal/motions/ports"
)

type Repository struct{ db *pgxpool.Pool }

func NewRepository(db *pgxpool.Pool) *Repository { return &Repository{db: db} }

func (r *Repository) Upsert(ctx context.Context, m *domain.Motion) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `
		INSERT INTO motions (dok_id, riksmote, beteckning, title, subtyp, motion_date,
		                     organ, tilldelat, status, systemdatum, behandlas_i, fetched_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11, now())
		ON CONFLICT (dok_id) DO UPDATE SET
			riksmote = EXCLUDED.riksmote, beteckning = EXCLUDED.beteckning,
			title = EXCLUDED.title, subtyp = EXCLUDED.subtyp,
			motion_date = EXCLUDED.motion_date, organ = EXCLUDED.organ,
			tilldelat = EXCLUDED.tilldelat, status = EXCLUDED.status,
			systemdatum = EXCLUDED.systemdatum, behandlas_i = EXCLUDED.behandlas_i,
			fetched_at = now()`,
		m.DokID, m.Riksmote, m.Beteckning, m.Title, m.Subtyp, m.Date,
		m.Organ, m.Tilldelat, m.Status, m.SystemDatum, m.BehandlasI); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, `DELETE FROM motion_authors WHERE dok_id = $1`, m.DokID); err != nil {
		return err
	}
	for _, a := range m.Authors {
		if _, err := tx.Exec(ctx, `
			INSERT INTO motion_authors (dok_id, ordning, intressent_id, name, party, role)
			VALUES ($1,$2,$3,$4,$5,$6)`,
			m.DokID, a.Order, a.PoliticianID, a.Name, a.Party, a.Role); err != nil {
			return err
		}
	}

	if _, err := tx.Exec(ctx, `DELETE FROM motion_yrkanden WHERE dok_id = $1`, m.DokID); err != nil {
		return err
	}
	for _, y := range m.Yrkanden {
		if _, err := tx.Exec(ctx, `
			INSERT INTO motion_yrkanden (dok_id, nummer, wording, utskottet, kammaren, behandlas_i, behandlas_i_punkt)
			VALUES ($1,$2,$3,$4,$5,$6,$7)`,
			m.DokID, y.Number, y.Wording, y.UtskottProposal, y.ChamberDecision, y.BehandlasI, y.BehandlasIPunkt); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *Repository) Get(ctx context.Context, riksmote, beteckning string) (*domain.Motion, error) {
	var m domain.Motion
	err := r.db.QueryRow(ctx, `
		SELECT dok_id, riksmote, beteckning, title, subtyp, motion_date, organ, tilldelat,
		       status, systemdatum, behandlas_i
		FROM motions WHERE riksmote = $1 AND beteckning = $2`, riksmote, beteckning).
		Scan(&m.DokID, &m.Riksmote, &m.Beteckning, &m.Title, &m.Subtyp, &m.Date, &m.Organ,
			&m.Tilldelat, &m.Status, &m.SystemDatum, &m.BehandlasI)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ports.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	rows, err := r.db.Query(ctx, `
		SELECT ordning, intressent_id, name, party, role
		FROM motion_authors WHERE dok_id = $1 ORDER BY ordning`, m.DokID)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var a domain.Author
		if err := rows.Scan(&a.Order, &a.PoliticianID, &a.Name, &a.Party, &a.Role); err != nil {
			rows.Close()
			return nil, err
		}
		m.Authors = append(m.Authors, a)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}

	yrows, err := r.db.Query(ctx, `
		SELECT nummer, wording, utskottet, kammaren, behandlas_i, behandlas_i_punkt
		FROM motion_yrkanden WHERE dok_id = $1 ORDER BY nummer`, m.DokID)
	if err != nil {
		return nil, err
	}
	defer yrows.Close()
	for yrows.Next() {
		var y domain.Yrkande
		if err := yrows.Scan(&y.Number, &y.Wording, &y.UtskottProposal, &y.ChamberDecision,
			&y.BehandlasI, &y.BehandlasIPunkt); err != nil {
			return nil, err
		}
		m.Yrkanden = append(m.Yrkanden, y)
	}
	return &m, yrows.Err()
}

func (r *Repository) StoredVersions(ctx context.Context, riksmote string) (map[string]time.Time, error) {
	rows, err := r.db.Query(ctx, `SELECT dok_id, systemdatum FROM motions WHERE riksmote = $1`, riksmote)
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

func (r *Repository) UpsertCoverage(ctx context.Context, riksmote, source string, expected int) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO motion_coverage (riksmote, source, expected, last_enumerated_at)
		VALUES ($1,$2,$3, now())
		ON CONFLICT (riksmote, source) DO UPDATE
		SET expected = EXCLUDED.expected, last_enumerated_at = now()`, riksmote, source, expected)
	return err
}

func (r *Repository) Coverage(ctx context.Context) ([]domain.Coverage, error) {
	rows, err := r.db.Query(ctx, `
		SELECT c.riksmote, c.source, c.expected,
		       (SELECT count(*) FROM motions m WHERE m.riksmote = c.riksmote),
		       c.last_enumerated_at
		FROM motion_coverage c
		ORDER BY c.riksmote DESC, c.source`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Coverage
	for rows.Next() {
		var c domain.Coverage
		if err := rows.Scan(&c.Riksmote, &c.Source, &c.Expected, &c.Held, &c.LastEnumeratedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *Repository) PendingBetankanden(ctx context.Context, limit int) ([]ports.BetRef, error) {
	rows, err := r.db.Query(ctx, `
		WITH refs AS (
			SELECT split_part(h, ':', 1) AS rm, split_part(h, ':', 2) AS bet, max(fetched_at) AS motion_fetched
			FROM (
				SELECT coalesce(nullif(y.behandlas_i, ''), m.behandlas_i) AS h, m.fetched_at
				FROM motion_yrkanden y JOIN motions m ON m.dok_id = y.dok_id
			) x
			WHERE h ~ '^[0-9]{4}/[0-9]{2}:.+'
			GROUP BY 1, 2
		)
		SELECT r.rm, r.bet
		FROM refs r
		LEFT JOIN betankanden b ON b.riksmote = r.rm AND b.beteckning = r.bet
		WHERE b.riksmote IS NULL
		   OR (b.found AND b.fetched_at < r.motion_fetched)
		   OR (NOT b.found AND b.fetched_at < now() - interval '1 day')
		ORDER BY r.rm DESC, r.bet
		LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ports.BetRef
	for rows.Next() {
		var b ports.BetRef
		if err := rows.Scan(&b.Riksmote, &b.Beteckning); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

func (r *Repository) SaveBetankande(ctx context.Context, b *domain.Betankande) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, `
		INSERT INTO betankanden (riksmote, beteckning, dok_id, found, title, organ, systemdatum, fetched_at)
		VALUES ($1,$2,$3,true,$4,$5,$6, now())
		ON CONFLICT (riksmote, beteckning) DO UPDATE SET
			dok_id = EXCLUDED.dok_id, found = true, title = EXCLUDED.title,
			organ = EXCLUDED.organ, systemdatum = EXCLUDED.systemdatum, fetched_at = now()`,
		b.Riksmote, b.Beteckning, b.DokID, b.Title, b.Organ, nullTime(b.SystemDatum)); err != nil {
		return err
	}
	nums := make([]string, 0, len(b.Punkter))
	for _, p := range b.Punkter {
		nums = append(nums, p.Number)
		if _, err := tx.Exec(ctx, `
			INSERT INTO forslagspunkter (riksmote, beteckning, punkt, rubrik, forslag_text, beslutstyp, votering_id, alternativ, vinnare)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
			ON CONFLICT (riksmote, beteckning, punkt) DO UPDATE SET
				rubrik = EXCLUDED.rubrik, forslag_text = EXCLUDED.forslag_text,
				beslutstyp = EXCLUDED.beslutstyp, votering_id = EXCLUDED.votering_id,
				alternativ = EXCLUDED.alternativ, vinnare = EXCLUDED.vinnare`,
			b.Riksmote, b.Beteckning, p.Number, p.Rubrik, p.Text, p.Beslutstyp, p.VoteringID, p.Alternativ, p.Vinnare); err != nil {
			return err
		}
	}
	if _, err := tx.Exec(ctx, `
		DELETE FROM forslagspunkter WHERE riksmote = $1 AND beteckning = $2 AND punkt <> ALL($3)`,
		b.Riksmote, b.Beteckning, nums); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func nullTime(t time.Time) any {
	if t.IsZero() {
		return nil
	}
	return t
}

func (r *Repository) MarkBetankandeMissing(ctx context.Context, rm, bet string) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO betankanden (riksmote, beteckning, found, fetched_at) VALUES ($1,$2,false, now())
		ON CONFLICT (riksmote, beteckning) DO UPDATE SET found = false, fetched_at = now()`, rm, bet)
	return err
}

func (r *Repository) YrkandenHandledIn(ctx context.Context, rm, bet string) ([]ports.YrkandeKey, error) {
	rows, err := r.db.Query(ctx, `
		SELECT y.dok_id, y.nummer, m.riksmote, m.beteckning, y.behandlas_i_punkt
		FROM motion_yrkanden y JOIN motions m ON m.dok_id = y.dok_id
		WHERE coalesce(nullif(y.behandlas_i, ''), m.behandlas_i) = $1
		ORDER BY y.dok_id, y.nummer`, rm+":"+bet)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ports.YrkandeKey
	for rows.Next() {
		var k ports.YrkandeKey
		if err := rows.Scan(&k.MotionDokID, &k.Number, &k.MotionRiksmote, &k.MotionBeteckn, &k.BehandlasIPunkt); err != nil {
			return nil, err
		}
		out = append(out, k)
	}
	return out, rows.Err()
}

func (r *Repository) Punkter(ctx context.Context, rm, bet string) ([]domain.Forslagspunkt, error) {
	rows, err := r.db.Query(ctx, `
		SELECT punkt, rubrik, forslag_text, beslutstyp, votering_id, alternativ, vinnare
		FROM forslagspunkter WHERE riksmote = $1 AND beteckning = $2
		ORDER BY nullif(regexp_replace(punkt, '\D', '', 'g'), '')::int NULLS LAST, punkt`, rm, bet)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Forslagspunkt
	for rows.Next() {
		var p domain.Forslagspunkt
		if err := rows.Scan(&p.Number, &p.Rubrik, &p.Text, &p.Beslutstyp, &p.VoteringID, &p.Alternativ, &p.Vinnare); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *Repository) ReplaceLinks(ctx context.Context, key ports.YrkandeKey, rm, bet string, links []domain.PunktLink) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, `
		DELETE FROM yrkande_forslagspunkt
		WHERE motion_dok_id = $1 AND yrkande_nummer = $2 AND riksmote = $3 AND beteckning = $4`,
		key.MotionDokID, key.Number, rm, bet); err != nil {
		return err
	}
	for _, l := range links {
		if _, err := tx.Exec(ctx, `
			INSERT INTO yrkande_forslagspunkt (motion_dok_id, yrkande_nummer, riksmote, beteckning, punkt, linked_by, decision)
			VALUES ($1,$2,$3,$4,$5,$6,$7)
			ON CONFLICT DO NOTHING`,
			key.MotionDokID, key.Number, rm, bet, l.Punkt, l.How, l.Decision); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *Repository) Decisions(ctx context.Context, motionDokID string) ([]domain.YrkandeDecision, error) {
	rows, err := r.db.Query(ctx, `
		SELECT l.yrkande_nummer, l.linked_by, l.decision, l.riksmote, l.beteckning,
		       b.title, b.dok_id,
		       p.punkt, p.rubrik, p.forslag_text, p.beslutstyp, p.votering_id, p.alternativ, p.vinnare,
		       EXISTS (SELECT 1 FROM voteringar v WHERE p.votering_id <> '' AND upper(v.votering_id) = upper(p.votering_id))
		FROM yrkande_forslagspunkt l
		JOIN forslagspunkter p ON p.riksmote = l.riksmote AND p.beteckning = l.beteckning AND p.punkt = l.punkt
		JOIN betankanden b ON b.riksmote = l.riksmote AND b.beteckning = l.beteckning
		WHERE l.motion_dok_id = $1
		ORDER BY l.yrkande_nummer, l.riksmote, l.beteckning,
		         nullif(regexp_replace(l.punkt, '\D', '', 'g'), '')::int NULLS LAST, l.punkt`, motionDokID)
	if err != nil {
		return nil, err
	}
	var out []domain.YrkandeDecision
	for rows.Next() {
		var d domain.YrkandeDecision
		if err := rows.Scan(&d.YrkandeNumber, &d.LinkedBy, &d.Decision, &d.Riksmote, &d.Betankande,
			&d.BetTitle, &d.BetDokID,
			&d.Punkt.Number, &d.Punkt.Rubrik, &d.Punkt.Text, &d.Punkt.Beslutstyp, &d.Punkt.VoteringID,
			&d.Punkt.Alternativ, &d.Punkt.Vinnare, &d.VoteringStored); err != nil {
			rows.Close()
			return nil, err
		}
		out = append(out, d)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for i := range out {
		if !out[i].VoteringStored {
			continue
		}
		trows, err := r.db.Query(ctx, `
			SELECT b.party,
			       count(*) FILTER (WHERE b.vote_result = 'Ja'),
			       count(*) FILTER (WHERE b.vote_result = 'Nej'),
			       count(*) FILTER (WHERE b.vote_result = 'Avstår'),
			       count(*) FILTER (WHERE b.vote_result = 'Frånvarande')
			FROM voteringar v JOIN ballots b ON b.votering_ref = v.id
			WHERE upper(v.votering_id) = upper($1)
			GROUP BY b.party ORDER BY b.party`, out[i].Punkt.VoteringID)
		if err != nil {
			return nil, err
		}
		for trows.Next() {
			var t domain.PartyTally
			if err := trows.Scan(&t.Party, &t.Ja, &t.Nej, &t.Avstar, &t.Franvarande); err != nil {
				trows.Close()
				return nil, err
			}
			out[i].Parties = append(out[i].Parties, t)
		}
		trows.Close()
		if err := trows.Err(); err != nil {
			return nil, err
		}
	}
	return out, nil
}

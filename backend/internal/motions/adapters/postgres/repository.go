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
			INSERT INTO motion_yrkanden (dok_id, nummer, wording, utskottet, kammaren, behandlas_i, behandlas_i_punkt, wording_hash)
			VALUES ($1,$2,$3,$4,$5,$6,$7, NULLIF($8,''))`,
			m.DokID, y.Number, y.Wording, y.UtskottProposal, y.ChamberDecision, y.BehandlasI, y.BehandlasIPunkt,
			domain.WordingHash(y.Wording)); err != nil {
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
		SELECT nummer, wording, utskottet, kammaren, behandlas_i, behandlas_i_punkt, COALESCE(wording_hash, '')
		FROM motion_yrkanden WHERE dok_id = $1 ORDER BY nummer`, m.DokID)
	if err != nil {
		return nil, err
	}
	defer yrows.Close()
	for yrows.Next() {
		var y domain.Yrkande
		if err := yrows.Scan(&y.Number, &y.Wording, &y.UtskottProposal, &y.ChamberDecision,
			&y.BehandlasI, &y.BehandlasIPunkt, &y.WordingHash); err != nil {
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

func (r *Repository) PendingBetankanden(ctx context.Context, rm string, limit int) ([]ports.BetRef, error) {
	rows, err := r.db.Query(ctx, `
		WITH refs AS (
			SELECT split_part(h, ':', 1) AS rm, split_part(h, ':', 2) AS bet, max(fetched_at) AS motion_fetched
			FROM (
				SELECT coalesce(nullif(y.behandlas_i, ''), m.behandlas_i) AS h, m.fetched_at, m.riksmote AS mrm
				FROM motion_yrkanden y JOIN motions m ON m.dok_id = y.dok_id
			) x
			WHERE h ~ '^[0-9]{4}/[0-9]{2}:.+' AND ($2 = '' OR mrm = $2)
			GROUP BY 1, 2
		)
		SELECT r.rm, r.bet
		FROM refs r
		LEFT JOIN betankanden b ON b.riksmote = r.rm AND b.beteckning = r.bet
		WHERE b.riksmote IS NULL
		   OR (b.found AND b.fetched_at < r.motion_fetched)
		   OR (b.found AND b.html_fetched_at IS NULL
		       AND (b.html_attempted_at IS NULL OR b.html_attempted_at < now() - interval '1 day'))
		   OR (b.found AND b.debate_checked_at IS NULL)
		   OR (NOT b.found AND b.fetched_at < now() - interval '1 day')
		ORDER BY r.rm DESC, r.bet
		LIMIT $1`, limit, rm)
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
		INSERT INTO betankanden (riksmote, beteckning, dok_id, found, title, organ, systemdatum, fetched_at, debate_date, debate_checked_at)
		VALUES ($1,$2,$3,true,$4,$5,$6, now(), $7, now())
		ON CONFLICT (riksmote, beteckning) DO UPDATE SET
			dok_id = EXCLUDED.dok_id, found = true, title = EXCLUDED.title,
			organ = EXCLUDED.organ, systemdatum = EXCLUDED.systemdatum, fetched_at = now(),
			debate_date = EXCLUDED.debate_date, debate_checked_at = now()`,
		b.Riksmote, b.Beteckning, b.DokID, b.Title, b.Organ, nullTime(b.SystemDatum), nullTime(b.DebateDate)); err != nil {
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

func (r *Repository) MarkHTMLAttempt(ctx context.Context, rm, bet string) error {
	_, err := r.db.Exec(ctx, `UPDATE betankanden SET html_attempted_at = now() WHERE riksmote = $1 AND beteckning = $2`, rm, bet)
	return err
}

func (r *Repository) SaveBetankandeHTML(ctx context.Context, rm, bet string, h *domain.BetankandeHTML) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, `DELETE FROM betankande_statements WHERE riksmote = $1 AND beteckning = $2`, rm, bet); err != nil {
		return err
	}
	for _, s := range h.Statements {
		parties, signers := s.Parties, s.Signers
		if parties == nil {
			parties = []string{}
		}
		if signers == nil {
			signers = []string{}
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO betankande_statements (riksmote, beteckning, kind, nummer, rubrik, punkt, parties, signers, proposal, body)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
			ON CONFLICT DO NOTHING`,
			rm, bet, s.Kind, s.Number, s.Rubrik, s.Punkt, parties, signers, s.Proposal, s.Text); err != nil {
			return err
		}
	}
	rows, err := tx.Query(ctx, `SELECT punkt, rubrik FROM forslagspunkter WHERE riksmote = $1 AND beteckning = $2`, rm, bet)
	if err != nil {
		return err
	}
	reasoning := map[string]string{}
	for rows.Next() {
		var punkt, rubrik string
		if err := rows.Scan(&punkt, &rubrik); err != nil {
			rows.Close()
			return err
		}
		reasoning[punkt] = h.ReasoningFor(rubrik)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}
	for punkt, text := range reasoning {
		if _, err := tx.Exec(ctx, `UPDATE forslagspunkter SET reasoning = $4 WHERE riksmote = $1 AND beteckning = $2 AND punkt = $3`, rm, bet, punkt, text); err != nil {
			return err
		}
	}
	if _, err := tx.Exec(ctx, `
		UPDATE betankanden SET html_fetched_at = now(), html_attempted_at = now(),
		       html_has_reasoning = $3, html_has_reservations = $4
		WHERE riksmote = $1 AND beteckning = $2`, rm, bet, h.HasReasoning, h.HasReservations); err != nil {
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
		SELECT punkt, rubrik, forslag_text, beslutstyp, votering_id, alternativ, vinnare, reasoning
		FROM forslagspunkter WHERE riksmote = $1 AND beteckning = $2
		ORDER BY nullif(regexp_replace(punkt, '\D', '', 'g'), '')::int NULLS LAST, punkt`, rm, bet)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Forslagspunkt
	for rows.Next() {
		var p domain.Forslagspunkt
		if err := rows.Scan(&p.Number, &p.Rubrik, &p.Text, &p.Beslutstyp, &p.VoteringID, &p.Alternativ, &p.Vinnare, &p.Reasoning); err != nil {
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
		       p.reasoning, b.html_fetched_at IS NOT NULL, b.html_has_reasoning, b.html_has_reservations,
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
			&d.Punkt.Alternativ, &d.Punkt.Vinnare, &d.Punkt.Reasoning,
			&d.HTMLRead, &d.HasReasoningSection, &d.HasReservationsSection, &d.VoteringStored); err != nil {
			rows.Close()
			return nil, err
		}
		out = append(out, d)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}
	stmts := map[[2]string][]domain.Statement{}
	for i := range out {
		key := [2]string{out[i].Riksmote, out[i].Betankande}
		if _, ok := stmts[key]; !ok {
			list, err := r.statements(ctx, key[0], key[1])
			if err != nil {
				return nil, err
			}
			stmts[key] = list
		}
		for _, s := range stmts[key] {
			if s.Punkt != out[i].Punkt.Number {
				continue
			}
			if s.Kind == domain.KindReservation {
				out[i].Reservations = append(out[i].Reservations, s)
			} else {
				out[i].Yttranden = append(out[i].Yttranden, s)
			}
		}
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

func (r *Repository) RecurringMembers(ctx context.Context, hash string) ([]domain.RecurringMember, error) {
	rows, err := r.db.Query(ctx, `
		SELECT m.dok_id, m.riksmote, m.beteckning, y.nummer, m.motion_date, y.kammaren,
		       COALESCE((SELECT array_agg(s.party ORDER BY s.first)
		                 FROM (SELECT a.party, min(a.ordning) AS first
		                       FROM motion_authors a
		                       WHERE a.dok_id = m.dok_id AND a.party NOT IN ('', '-')
		                         AND lower(a.role) IN ('', 'undertecknare')
		                       GROUP BY a.party) s), '{}')
		FROM motion_yrkanden y JOIN motions m ON m.dok_id = y.dok_id
		WHERE y.wording_hash = $1
		ORDER BY m.riksmote, nullif(regexp_replace(m.beteckning, '\D', '', 'g'), '')::int NULLS LAST,
		         m.beteckning, y.nummer`, hash)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.RecurringMember
	for rows.Next() {
		var mem domain.RecurringMember
		if err := rows.Scan(&mem.DokID, &mem.Riksmote, &mem.Beteckning, &mem.Number, &mem.Date,
			&mem.ChamberDecision, &mem.Parties); err != nil {
			return nil, err
		}
		out = append(out, mem)
	}
	return out, rows.Err()
}

func (r *Repository) SetWordingHash(ctx context.Context, dokID string, number int, hash string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE motion_yrkanden SET wording_hash = NULLIF($3, '')
		WHERE dok_id = $1 AND nummer = $2`, dokID, number, hash)
	return err
}

func (r *Repository) BackfillWordingHashes(ctx context.Context, after ports.YrkandeCursor, limit int) (int, int, ports.YrkandeCursor, error) {
	last := after
	rows, err := r.db.Query(ctx, `
		SELECT dok_id, nummer, wording FROM motion_yrkanden
		WHERE wording_hash IS NULL AND (dok_id, nummer) > ($1, $2)
		ORDER BY dok_id, nummer LIMIT $3`, after.DokID, after.Number, limit)
	if err != nil {
		return 0, 0, last, err
	}
	type row struct {
		dokID   string
		number  int
		wording string
	}
	var todo []row
	for rows.Next() {
		var x row
		if err := rows.Scan(&x.dokID, &x.number, &x.wording); err != nil {
			rows.Close()
			return 0, 0, last, err
		}
		todo = append(todo, x)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return 0, 0, last, err
	}
	written := 0
	for _, x := range todo {
		last = ports.YrkandeCursor{DokID: x.dokID, Number: x.number}
		h := domain.WordingHash(x.wording)
		if h == "" {
			continue // nothing to group on; the cursor moves past it
		}
		if err := r.SetWordingHash(ctx, x.dokID, x.number, h); err != nil {
			return written, len(todo), last, err
		}
		written++
	}
	return written, len(todo), last, nil
}

func (r *Repository) Statements(ctx context.Context, rm, bet string) ([]domain.Statement, error) {
	return r.statements(ctx, rm, bet)
}

func (r *Repository) statements(ctx context.Context, rm, bet string) ([]domain.Statement, error) {
	rows, err := r.db.Query(ctx, `
		SELECT kind, nummer, rubrik, punkt, parties, signers, proposal, body
		FROM betankande_statements WHERE riksmote = $1 AND beteckning = $2
		ORDER BY kind, nummer`, rm, bet)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Statement
	for rows.Next() {
		var s domain.Statement
		if err := rows.Scan(&s.Kind, &s.Number, &s.Rubrik, &s.Punkt, &s.Parties, &s.Signers, &s.Proposal, &s.Text); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *Repository) GetBetankande(ctx context.Context, rm, bet string) (*domain.BetankandeStored, error) {
	var b domain.BetankandeStored
	var sd *time.Time
	err := r.db.QueryRow(ctx, `
		SELECT riksmote, beteckning, dok_id, title, organ, systemdatum,
		       html_fetched_at IS NOT NULL, html_has_reasoning, html_has_reservations
		FROM betankanden WHERE riksmote = $1 AND beteckning = $2 AND found`, rm, bet).
		Scan(&b.Riksmote, &b.Beteckning, &b.DokID, &b.Title, &b.Organ, &sd,
			&b.HTMLRead, &b.HasReasoningSection, &b.HasReservationsSection)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ports.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if sd != nil {
		b.SystemDatum = *sd
	}
	return &b, nil
}

func (r *Repository) PunktMotions(ctx context.Context, rm, bet string) (map[string][]domain.PunktMotion, error) {
	rows, err := r.db.Query(ctx, `
		SELECT l.punkt, m.dok_id, m.riksmote, m.beteckning, m.title, m.subtyp, l.yrkande_nummer,
		       COALESCE((SELECT array_agg(s.party ORDER BY s.first)
		                 FROM (SELECT a.party, min(a.ordning) AS first FROM motion_authors a
		                       WHERE a.dok_id = m.dok_id AND a.party NOT IN ('', '-') GROUP BY a.party) s), '{}')
		FROM yrkande_forslagspunkt l JOIN motions m ON m.dok_id = l.motion_dok_id
		WHERE l.riksmote = $1 AND l.beteckning = $2
		ORDER BY l.punkt, m.riksmote, nullif(regexp_replace(m.beteckning, '\D', '', 'g'), '')::int NULLS LAST, m.dok_id, l.yrkande_nummer`, rm, bet)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string][]domain.PunktMotion{}
	for rows.Next() {
		var punkt string
		var m domain.PunktMotion
		var nr int
		if err := rows.Scan(&punkt, &m.DokID, &m.Riksmote, &m.Beteckning, &m.Title, &m.Subtyp, &nr, &m.Parties); err != nil {
			return nil, err
		}
		list := out[punkt]
		if n := len(list); n > 0 && list[n-1].DokID == m.DokID {
			list[n-1].Yrkanden = append(list[n-1].Yrkanden, nr)
		} else {
			m.Yrkanden = []int{nr}
			list = append(list, m)
		}
		out[punkt] = list
	}
	return out, rows.Err()
}

func (r *Repository) Ballots(ctx context.Context, voteringID string) ([]domain.Ballot, error) {
	rows, err := r.db.Query(ctx, `
		SELECT b.politician_id, trim(p.first_name || ' ' || p.last_name), b.party, b.vote_result
		FROM voteringar v JOIN ballots b ON b.votering_ref = v.id
		JOIN politicians p ON p.intressent_id = b.politician_id
		WHERE upper(v.votering_id) = upper($1)
		ORDER BY b.party, b.politician_id`, voteringID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Ballot
	for rows.Next() {
		var b domain.Ballot
		if err := rows.Scan(&b.PoliticianID, &b.Name, &b.Party, &b.Result); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

func (r *Repository) VoteringPunkter(ctx context.Context, rm, bet string) ([]domain.VoteringPunkt, error) {
	rows, err := r.db.Query(ctx, `
		SELECT votering_id, forslagspunkt, coalesce(document_title, '')
		FROM voteringar WHERE session = $1 AND beteckning = $2
		ORDER BY nullif(regexp_replace(forslagspunkt, '\D', '', 'g'), '')::int NULLS LAST, forslagspunkt, votering_id`, rm, bet)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.VoteringPunkt
	for rows.Next() {
		var v domain.VoteringPunkt
		if err := rows.Scan(&v.VoteringID, &v.Punkt, &v.Title); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func (r *Repository) ResolveBetankande(ctx context.Context, bet, punkt, rm string) (string, bool, error) {
	// Riksmöten written YYYY/YY sort correctly as text.
	var found *string
	err := r.db.QueryRow(ctx, `
		SELECT max(rm) FROM (
			SELECT riksmote AS rm FROM betankanden WHERE found AND upper(beteckning) = upper($1)
			  AND ($2 = '' OR EXISTS (SELECT 1 FROM forslagspunkter f WHERE f.riksmote = betankanden.riksmote
			       AND f.beteckning = betankanden.beteckning AND f.punkt = $2))
			UNION
			SELECT session FROM voteringar WHERE upper(beteckning) = upper($1) AND ($2 = '' OR forslagspunkt = $2)
		) x WHERE ($3 = '' OR rm = $3)`, bet, punkt, rm).Scan(&found)
	if err != nil {
		return "", false, err
	}
	if found == nil {
		return "", false, nil
	}
	return *found, true, nil
}

func (r *Repository) BackfillProgress(ctx context.Context) ([]domain.BackfillProgress, error) {
	rows, err := r.db.Query(ctx, `
		SELECT riksmote, source, expected, held, completed_at, updated_at
		FROM motion_backfill ORDER BY riksmote DESC, source`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.BackfillProgress
	for rows.Next() {
		var p domain.BackfillProgress
		if err := rows.Scan(&p.Riksmote, &p.Source, &p.Expected, &p.Held, &p.CompletedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *Repository) SaveBackfillProgress(ctx context.Context, p domain.BackfillProgress) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO motion_backfill (riksmote, source, expected, held, completed_at, updated_at)
		VALUES ($1,$2,$3,$4,$5, now())
		ON CONFLICT (riksmote, source) DO UPDATE SET
			expected = EXCLUDED.expected, held = EXCLUDED.held,
			completed_at = COALESCE(motion_backfill.completed_at, EXCLUDED.completed_at),
			updated_at = now()`,
		p.Riksmote, p.Source, p.Expected, p.Held, p.CompletedAt)
	return err
}

func (r *Repository) BetankandeCounts(ctx context.Context, rm string) (int, int, error) {
	var expected, held int
	err := r.db.QueryRow(ctx, `
		WITH refs AS (
			SELECT DISTINCT split_part(h, ':', 1) AS rm, split_part(h, ':', 2) AS bet
			FROM (
				SELECT coalesce(nullif(y.behandlas_i, ''), m.behandlas_i) AS h
				FROM motion_yrkanden y JOIN motions m ON m.dok_id = y.dok_id
				WHERE m.riksmote = $1
			) x
			WHERE h ~ '^[0-9]{4}/[0-9]{2}:.+'
		)
		SELECT count(*), count(b.riksmote) FILTER (WHERE b.found)
		FROM refs r LEFT JOIN betankanden b ON b.riksmote = r.rm AND b.beteckning = r.bet`, rm).Scan(&expected, &held)
	return expected, held, err
}

func (r *Repository) MotionCount(ctx context.Context, rm string) (int, error) {
	var n int
	err := r.db.QueryRow(ctx, `SELECT count(*) FROM motions WHERE riksmote = $1`, rm).Scan(&n)
	return n, err
}

func (r *Repository) BackfillRiksmoten(ctx context.Context) ([]string, error) {
	rows, err := r.db.Query(ctx, `
		SELECT rm FROM (
			SELECT DISTINCT session AS rm FROM voteringar
			UNION
			SELECT unnest(riksmoten) FROM mandate_periods
		) x WHERE rm ~ '^[0-9]{4}/[0-9]{2}$'
		ORDER BY rm DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var rm string
		if err := rows.Scan(&rm); err != nil {
			return nil, err
		}
		out = append(out, rm)
	}
	return out, rows.Err()
}

func (r *Repository) DecisionVocabulary(ctx context.Context) ([]domain.DecisionValue, error) {
	rows, err := r.db.Query(ctx, `
		SELECT field, value, count(*) FROM (
			SELECT 'kammaren' AS field, kammaren AS value FROM motion_yrkanden
			UNION ALL SELECT 'beslutstyp', beslutstyp FROM forslagspunkter
			UNION ALL SELECT 'vinnare', vinnare FROM forslagspunkter
			UNION ALL SELECT 'punktbeslut', decision FROM yrkande_forslagspunkt
		) v
		WHERE btrim(value) <> ''
		GROUP BY field, value
		ORDER BY field, count(*) DESC, value`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.DecisionValue
	for rows.Next() {
		var v domain.DecisionValue
		if err := rows.Scan(&v.Field, &v.Value, &v.Count); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func (r *Repository) HeldBetankanden(ctx context.Context) ([]ports.BetRef, error) {
	rows, err := r.db.Query(ctx, `SELECT riksmote, beteckning FROM betankanden WHERE found ORDER BY riksmote DESC, beteckning`)
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

func (r *Repository) OriginCoverage(ctx context.Context) (int, int, error) {
	var total, with int
	err := r.db.QueryRow(ctx, `
		SELECT count(*),
		       count(*) FILTER (WHERE EXISTS (
		           SELECT 1 FROM yrkande_forslagspunkt l
		           WHERE l.riksmote = v.session AND upper(l.beteckning) = upper(v.beteckning)
		             AND l.punkt = v.forslagspunkt))
		FROM voteringar v`).Scan(&total, &with)
	return total, with, err
}

// Debate reads the speeches linked to a betänkande (speeches.betankande_dok_id).
func (r *Repository) Debate(ctx context.Context, betDokID string, limit int) (*domain.Debate, error) {
	d := &domain.Debate{Speeches: []domain.DebateSpeech{}}
	if betDokID == "" {
		return d, nil
	}
	if err := r.db.QueryRow(ctx, `SELECT count(*) FROM speeches WHERE upper(betankande_dok_id) = upper($1)`, betDokID).Scan(&d.Total); err != nil {
		return nil, err
	}
	if d.Total == 0 {
		return d, nil
	}
	rows, err := r.db.Query(ctx, `
		SELECT s.id, s.dok_id, coalesce(s.anforande_nummer, ''), s.politician_id,
		       p.first_name || ' ' || p.last_name, s.party, s.date,
		       left(coalesce(s.speech_text, ''), 1500), coalesce(s.betankande_link_rule, '')
		FROM speeches s JOIN politicians p ON p.intressent_id = s.politician_id
		WHERE upper(s.betankande_dok_id) = upper($1)
		ORDER BY s.date, s.dok_id, nullif(regexp_replace(s.anforande_nummer, '\D', '', 'g'), '')::int, s.id
		LIMIT $2`, betDokID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var sp domain.DebateSpeech
		if err := rows.Scan(&sp.SpeechID, &sp.DokID, &sp.AnforandeNummer, &sp.PoliticianID, &sp.Name, &sp.Party, &sp.Date, &sp.Excerpt, &sp.LinkRule); err != nil {
			return nil, err
		}
		d.Speeches = append(d.Speeches, sp)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(d.Speeches) > 0 {
		d.Date = d.Speeches[0].Date
	}
	return d, nil
}

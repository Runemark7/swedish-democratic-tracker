package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"riksdagskollen/internal/curation/domain"
)

// A votering lies in a Löfte's mandate period when its riksmöte belongs to the
// period and, if the date is known, the date is between the election days that
// open and close it. A period that does not exist contains nothing.
func (s *store) VoteringForLoefte(ctx context.Context, voteringID string, l *domain.Loefte) (*domain.Votering, error) {
	var v domain.Votering
	err := s.tx.QueryRow(ctx, `
		SELECT v.id, v.votering_id,
		       COALESCE((SELECT v.session = ANY(mp.riksmoten)
		                        AND (v.system_datum IS NULL
		                             OR (v.system_datum::date BETWEEN mp.start_date AND mp.end_date))
		                   FROM mandate_periods mp WHERE mp.code = $2), false)
		  FROM voteringar v WHERE v.votering_id = $1`, voteringID, l.MandatePeriod).
		Scan(&v.Ref, &v.VoteringID, &v.InPeriod)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &v, err
}

func (s *store) KopplingByPair(ctx context.Context, loefteID, ref int64) (*domain.Koppling, error) {
	var k domain.Koppling
	err := s.tx.QueryRow(ctx, `SELECT k.id, k.loefte_id, v.votering_id, k.direction, k.state,
		k.confidence::float8, k.threshold::float8, k.reasoning, k.origin
		FROM kopplingar k JOIN voteringar v ON v.id = k.votering_ref
		WHERE k.loefte_id=$1 AND k.votering_ref=$2 FOR UPDATE OF k`, loefteID, ref).
		Scan(&k.ID, &k.LoefteID, &k.VoteringID, &k.Direction, &k.State, &k.Confidence, &k.Threshold, &k.Reasoning, &k.Origin)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &k, err
}

func (s *store) InsertKoppling(ctx context.Context, loefteID, ref int64, k domain.Koppling) (*domain.Koppling, error) {
	k.LoefteID = loefteID
	err := s.tx.QueryRow(ctx, `INSERT INTO kopplingar
		(loefte_id, votering_ref, direction, state, confidence, threshold, reasoning, origin)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id`,
		loefteID, ref, k.Direction, k.State, k.Confidence, k.Threshold, k.Reasoning, k.Origin).Scan(&k.ID)
	if pgCode(err) == "23505" {
		return nil, fmt.Errorf("%w: this Löfte and votering already have a koppling", domain.ErrConflict)
	}
	return &k, err
}

// Candidates: decided voteringar ingested after the last closed Veckobatch,
// inside the mandate period of an approved Löfte whose utskott matches the
// votering's (the letters of its beteckning, SkU16 → SkU), paired with that
// Löfte. Pairs already suggested, in any state, are left out so the bot never
// re-judges them. A Löfte with no utskott is never a candidate: utskott is
// what narrows the pairs.
func (r *Repository) Candidates(ctx context.Context, limit int, afterRef int64) (*domain.Candidates, error) {
	out := &domain.Candidates{Voteringar: []domain.Candidate{}}
	var since *time.Time
	if err := r.db.QueryRow(ctx, `SELECT max(created_at) FROM veckobatch WHERE closed_at IS NOT NULL`).Scan(&since); err != nil {
		return nil, err
	}
	if since != nil {
		s := since.UTC().Format(time.RFC3339)
		out.Since = &s
	}
	rows, err := r.db.Query(ctx, `
		WITH pairs AS (
		  SELECT v.id AS vref, l.id AS lid
		    FROM voteringar v
		    JOIN loften l ON l.visibility = 'approved'
		    JOIN mandate_periods mp ON mp.code = l.mandate_period
		     AND v.session = ANY(mp.riksmoten)
		     AND (v.system_datum IS NULL OR v.system_datum::date BETWEEN mp.start_date AND mp.end_date)
		   WHERE v.id > $3
		     AND ($1::timestamptz IS NULL OR v.created_at > $1)
		     AND upper(substring(v.beteckning from '^[^0-9]+')) IN (SELECT upper(u) FROM unnest(l.utskott) u)
		     AND NOT EXISTS (SELECT 1 FROM kopplingar k WHERE k.loefte_id = l.id AND k.votering_ref = v.id)
		), picked AS (
		  SELECT DISTINCT vref FROM pairs ORDER BY vref LIMIT $2
		)
		SELECT v.id, v.votering_id, v.beteckning, v.forslagspunkt, v.session, v.system_datum,
		       upper(substring(v.beteckning from '^[^0-9]+')),
		       l.id, l.party, l.text, coalesce(l.quote,''), coalesce(l.locator,''), coalesce(l.source_url,''),
		       coalesce(l.topic,''), l.utskott
		  FROM picked p
		  JOIN pairs ON pairs.vref = p.vref
		  JOIN voteringar v ON v.id = p.vref
		  JOIN loften l ON l.id = pairs.lid
		 ORDER BY v.id, l.id`, since, limit, afterRef)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var last int64
	for rows.Next() {
		var ref int64
		var c domain.Candidate
		var d *time.Time
		var l domain.CandidateLoefte
		if err := rows.Scan(&ref, &c.VoteringID, &c.Beteckning, &c.Forslagspkt, &c.Session, &d, &c.Utskott,
			&l.ID, &l.Party, &l.Text, &l.Quote, &l.Locator, &l.SourceURL, &l.Topic, &l.Utskott); err != nil {
			return nil, err
		}
		if ref != last {
			if d != nil {
				ds := d.UTC().Format(time.RFC3339)
				c.Datum = &ds
			}
			c.Motions, c.Loften = []domain.CandidateMotion{}, []domain.CandidateLoefte{}
			out.Voteringar = append(out.Voteringar, c)
			last = ref
		}
		cur := &out.Voteringar[len(out.Voteringar)-1]
		cur.Loften = append(cur.Loften, l)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	rows.Close()
	if len(out.Voteringar) == limit {
		out.Next = &last
	}
	for i := range out.Voteringar {
		if err := r.fillContext(ctx, &out.Voteringar[i]); err != nil {
			return nil, err
		}
	}
	return out, nil
}

// fillContext adds the betänkande, förslagspunkt and the motions and yrkanden
// decided there, so the bot can judge without further calls.
func (r *Repository) fillContext(ctx context.Context, c *domain.Candidate) error {
	err := r.db.QueryRow(ctx, `
		SELECT f.riksmote, f.beteckning, b.title, f.punkt, f.rubrik, f.forslag_text, f.alternativ
		  FROM forslagspunkter f JOIN betankanden b USING (riksmote, beteckning)
		 WHERE upper(f.votering_id) = upper($1) ORDER BY f.riksmote DESC, f.beteckning, f.punkt LIMIT 1`, c.VoteringID).
		Scan(&c.Betankande.Riksmote, &c.Betankande.Beteckning, &c.Betankande.Title, &c.Betankande.Punkt,
			&c.Betankande.Rubrik, &c.Betankande.Forslag, &c.Betankande.Alternativ)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil // not every votering has its betänkande ingested yet
	}
	if err != nil {
		return err
	}
	rows, err := r.db.Query(ctx, `
		SELECT m.riksmote, m.beteckning, m.title, coalesce((SELECT a.party FROM motion_authors a WHERE a.dok_id = m.dok_id ORDER BY a.ordning LIMIT 1),''),
		       y.nummer, y.wording
		  FROM yrkande_forslagspunkt yf
		  JOIN motions m ON m.dok_id = yf.motion_dok_id
		  JOIN motion_yrkanden y ON y.dok_id = yf.motion_dok_id AND y.nummer = yf.yrkande_nummer
		 WHERE yf.riksmote = $1 AND yf.beteckning = $2 AND yf.punkt = $3
		 ORDER BY m.beteckning, y.nummer LIMIT 100`, c.Betankande.Riksmote, c.Betankande.Beteckning, c.Betankande.Punkt)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var m domain.CandidateMotion
		if err := rows.Scan(&m.Riksmote, &m.Beteckning, &m.Title, &m.Party, &m.Yrkande, &m.Wording); err != nil {
			return err
		}
		c.Motions = append(c.Motions, m)
	}
	return rows.Err()
}

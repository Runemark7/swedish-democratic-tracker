package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"riksdagskollen/internal/votes/domain"
	"riksdagskollen/internal/votes/ports"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) ListByPolitician(ctx context.Context, f ports.ListVotesFilter) (ports.ListVotesResult, error) {
	// Columns are qualified because the ballot and the votering now come from
	// different tables: politician_id is the member's, everything else describes
	// the question they voted on.
	args := []any{f.PoliticianID}
	where := "WHERE b.politician_id = $1"
	n := 2

	if f.Session != "" {
		where += fmt.Sprintf(" AND vg.session = $%d", n)
		args = append(args, f.Session)
		n++
	}

	const from = `FROM ballots b JOIN voteringar vg ON vg.id = b.votering_ref `

	var total int
	if err := r.db.QueryRow(ctx, "SELECT COUNT(*) "+from+where, args...).Scan(&total); err != nil {
		return ports.ListVotesResult{}, err
	}

	offset := (f.Page - 1) * f.PageSize
	args = append(args, f.PageSize, offset)
	q := fmt.Sprintf(`SELECT vg.votering_id, b.politician_id, b.party, b.vote_result, vg.beteckning, vg.forslagspunkt,
		vg.session, vg.dok_id, vg.document_title, vg.created_at
		%s%s ORDER BY vg.created_at DESC LIMIT $%d OFFSET $%d`, from, where, n, n+1)

	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return ports.ListVotesResult{}, err
	}
	defer rows.Close()

	var vv []*domain.Vote
	for rows.Next() {
		v, err := scanVote(rows)
		if err != nil {
			return ports.ListVotesResult{}, err
		}
		vv = append(vv, v)
	}
	return ports.ListVotesResult{Votes: vv, Total: total}, rows.Err()
}

func (r *Repository) ListByBeteckning(ctx context.Context, beteckning, punkt, session string) ([]*domain.Vote, error) {
	const q = `SELECT vg.votering_id, b.politician_id, b.party, b.vote_result, vg.beteckning, vg.forslagspunkt,
		vg.session, vg.dok_id, vg.document_title, vg.created_at
		FROM ballots b
		JOIN voteringar vg ON vg.id = b.votering_ref
		WHERE vg.beteckning = $1 AND vg.forslagspunkt = $2
		  AND ($3 = '' OR vg.session = $3)
		ORDER BY b.party, b.politician_id`

	rows, err := r.db.Query(ctx, q, beteckning, punkt, session)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var vv []*domain.Vote
	for rows.Next() {
		v, err := scanVote(rows)
		if err != nil {
			return nil, err
		}
		vv = append(vv, v)
	}
	return vv, rows.Err()
}

func (r *Repository) UpsertMany(ctx context.Context, vv []*domain.Vote) error {
	// Two tables, one statement. The votering is upserted first and returns its
	// id, which the ballot then references -- so a ballot can never be written
	// against a votering that does not exist.
	//
	// DO UPDATE rather than DO NOTHING on the votering: it must return a row in
	// every case, including a re-run where the votering is already present, or
	// the ballot insert would select from an empty CTE and silently drop.
	//
	// The deprecated proposed_by_party / proposal_type / proposal_dok_id /
	// origin_enriched columns are deliberately absent: nothing writes them any
	// more (proposed-by is derived from the Motioner link at read time).
	const q = `
		WITH vg AS (
			INSERT INTO voteringar (votering_id, beteckning, forslagspunkt, session, dok_id, system_datum)
			VALUES ($1, $5, $6, $7, $8, $9)
			ON CONFLICT (votering_id) DO UPDATE SET
				beteckning    = EXCLUDED.beteckning,
				forslagspunkt = EXCLUDED.forslagspunkt,
				session       = EXCLUDED.session,
				dok_id        = EXCLUDED.dok_id,
				system_datum  = EXCLUDED.system_datum
			RETURNING id
		)
		INSERT INTO ballots (votering_ref, politician_id, party, vote_result)
		SELECT vg.id, $2, $3, $4 FROM vg
		WHERE EXISTS (SELECT 1 FROM politicians WHERE intressent_id = $2)
		ON CONFLICT (votering_ref, politician_id) DO UPDATE SET
			party       = EXCLUDED.party,
			vote_result = EXCLUDED.vote_result`

	batch := &pgx.Batch{}
	for _, v := range vv {
		// system_datum may be zero for rows that predate the column; store NULL
		// rather than year 1, so "no date" is distinguishable from a real one.
		var sd any
		if !v.SystemDatum.IsZero() {
			sd = v.SystemDatum
		}
		batch.Queue(q, v.VoteringID, v.PoliticianID, v.Party, string(v.VoteResult),
			v.Beteckning, v.Forslagspunkt, v.Session, v.DokID, sd)
	}
	br := r.db.SendBatch(ctx, batch)
	defer br.Close()
	for range vv {
		if _, err := br.Exec(); err != nil {
			return err
		}
	}
	return br.Close()
}

func (r *Repository) ListDistinctVotes(ctx context.Context, f ports.ListDistinctVotesFilter) (ports.ListDistinctVotesResult, error) {
	var total int
	// One row per riksmöte + beteckning + förslagspunkt. DISTINCT survives the
	// move to voteringar: it is not collapsing duplicated ballots but two
	// genuinely separate voteringar that share all three — 18 of them in
	// 2022-2026. Not gated on the deprecated origin_enriched flag.
	if err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM (SELECT DISTINCT session, beteckning, forslagspunkt FROM voteringar) sub`,
	).Scan(&total); err != nil {
		return ports.ListDistinctVotesResult{}, err
	}

	offset := (f.Page - 1) * f.PageSize
	// The title comes from the betänkande the Motioner link holds when there is
	// one, else the title stored on the votering by the retired origin step.
	const q = `SELECT DISTINCT ON (vg.session, vg.beteckning, vg.forslagspunkt)
		vg.session, vg.beteckning, vg.forslagspunkt, COALESCE(NULLIF(bt.title, ''), vg.document_title, '')
		FROM voteringar vg
		LEFT JOIN betankanden bt ON bt.riksmote = vg.session AND bt.beteckning = vg.beteckning AND bt.found
		ORDER BY vg.session DESC, vg.beteckning, vg.forslagspunkt
		LIMIT $1 OFFSET $2`

	rows, err := r.db.Query(ctx, q, f.PageSize, offset)
	if err != nil {
		return ports.ListDistinctVotesResult{}, err
	}
	defer rows.Close()

	var ss []ports.VoteSummary
	for rows.Next() {
		var s ports.VoteSummary
		if err := rows.Scan(&s.Riksmote, &s.Beteckning, &s.Forslagspunkt, &s.DocumentTitle); err != nil {
			return ports.ListDistinctVotesResult{}, err
		}
		ss = append(ss, s)
	}
	return ports.ListDistinctVotesResult{Votes: ss, Total: total}, rows.Err()
}

func (r *Repository) ListDistinctByCommitteePrefix(ctx context.Context, prefix string) ([]ports.VoteSummary, error) {
	const q = `SELECT DISTINCT vg.session, vg.beteckning, vg.forslagspunkt, COALESCE(NULLIF(bt.title, ''), vg.document_title)
		FROM voteringar vg
		LEFT JOIN betankanden bt ON bt.riksmote = vg.session AND bt.beteckning = vg.beteckning AND bt.found
		WHERE vg.beteckning LIKE $1 || '%'
		  AND COALESCE(NULLIF(bt.title, ''), vg.document_title) IS NOT NULL
		ORDER BY vg.session, vg.beteckning, vg.forslagspunkt`

	rows, err := r.db.Query(ctx, q, prefix)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ss []ports.VoteSummary
	for rows.Next() {
		var s ports.VoteSummary
		if err := rows.Scan(&s.Riksmote, &s.Beteckning, &s.Forslagspunkt, &s.DocumentTitle); err != nil {
			return nil, err
		}
		ss = append(ss, s)
	}
	return ss, rows.Err()
}

// ProposedByForPoints derives proposed-by from the stored Motioner link, keyed
// on riksmöte + beteckning + förslagspunkt.
func (r *Repository) ProposedByForPoints(ctx context.Context, keys []ports.PointKey) (map[ports.PointKey]domain.ProposedBy, error) {
	out := make(map[ports.PointKey]domain.ProposedBy, len(keys))
	if len(keys) == 0 {
		return out, nil
	}
	rms, bets, punkts := make([]string, len(keys)), make([]string, len(keys)), make([]string, len(keys))
	for i, k := range keys {
		rms[i], bets[i], punkts[i] = k.Riksmote, k.Beteckning, k.Punkt
		out[k] = domain.ProposedBy{Motions: []domain.ProposingMotion{}, Propositions: []domain.PropositionRef{}}
	}
	const keysCTE = `WITH k AS (SELECT * FROM unnest($1::text[], $2::text[], $3::text[]) AS t(rm, bet, p))`

	// Punkt text (propositions) and the alternative put to the vote.
	prows, err := r.db.Query(ctx, keysCTE+`
		SELECT f.riksmote, f.beteckning, f.punkt, f.forslag_text, f.alternativ
		FROM forslagspunkter f JOIN k ON k.rm = f.riksmote AND k.bet = f.beteckning AND k.p = f.punkt`,
		rms, bets, punkts)
	if err != nil {
		return nil, err
	}
	for prows.Next() {
		var key ports.PointKey
		var text, alt string
		if err := prows.Scan(&key.Riksmote, &key.Beteckning, &key.Punkt, &text, &alt); err != nil {
			prows.Close()
			return nil, err
		}
		pb := out[key]
		pb.Propositions = domain.PropositionsIn(text)
		if pb.Propositions == nil {
			pb.Propositions = []domain.PropositionRef{}
		}
		pb.Alternativ = alt
		out[key] = pb
	}
	prows.Close()
	if err := prows.Err(); err != nil {
		return nil, err
	}

	// Motions and yrkanden the punkt decided, with the motion's authors' parties.
	lrows, err := r.db.Query(ctx, keysCTE+`
		SELECT l.riksmote, l.beteckning, l.punkt, m.dok_id, m.riksmote, m.beteckning, m.title, m.subtyp,
		       l.yrkande_nummer,
		       COALESCE((SELECT array_agg(a.party ORDER BY a.ordning) FROM motion_authors a
		                 WHERE a.dok_id = m.dok_id AND a.party <> ''), '{}')
		FROM yrkande_forslagspunkt l
		JOIN k ON k.rm = l.riksmote AND k.bet = l.beteckning AND k.p = l.punkt
		JOIN motions m ON m.dok_id = l.motion_dok_id
		ORDER BY l.riksmote, l.beteckning, l.punkt, m.dok_id, l.yrkande_nummer`,
		rms, bets, punkts)
	if err != nil {
		return nil, err
	}
	defer lrows.Close()
	for lrows.Next() {
		var key ports.PointKey
		var m domain.ProposingMotion
		var nr int
		var authorParties []string
		if err := lrows.Scan(&key.Riksmote, &key.Beteckning, &key.Punkt, &m.DokID, &m.Riksmote, &m.Beteckning,
			&m.Title, &m.Subtyp, &nr, &authorParties); err != nil {
			return nil, err
		}
		pb := out[key]
		if n := len(pb.Motions); n > 0 && pb.Motions[n-1].DokID == m.DokID {
			pb.Motions[n-1].Yrkanden = append(pb.Motions[n-1].Yrkanden, nr)
		} else {
			seen := map[string]bool{}
			m.Parties = []string{}
			for _, p := range authorParties {
				if !seen[p] {
					seen[p] = true
					m.Parties = append(m.Parties, p)
				}
			}
			m.Yrkanden = []int{nr}
			pb.Motions = append(pb.Motions, m)
		}
		out[key] = pb
	}
	return out, lrows.Err()
}

type scanner interface {
	Scan(dest ...any) error
}

func scanVote(s scanner) (*domain.Vote, error) {
	var v domain.Vote
	var dokID, documentTitle *string
	err := s.Scan(
		&v.VoteringID, &v.PoliticianID, &v.Party, &v.VoteResult,
		&v.Beteckning, &v.Forslagspunkt, &v.Session, &dokID,
		&documentTitle, &v.CreatedAt,
	)
	if dokID != nil {
		v.DokID = *dokID
	}
	if documentTitle != nil {
		v.ProposalOrigin.DocumentTitle = *documentTitle
	}
	return &v, err
}

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
		vg.session, vg.dok_id, vg.proposed_by_party, vg.proposal_type, vg.proposal_dok_id, vg.document_title,
		vg.origin_enriched, vg.created_at
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
		vg.session, vg.dok_id, vg.proposed_by_party, vg.proposal_type, vg.proposal_dok_id, vg.document_title,
		vg.origin_enriched, vg.created_at
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
	// The proposal_* columns are deliberately absent. They belong to the
	// votering and are written by enrich-vote-origins; listing them here with
	// their zero values would erase enrichment on every re-ingest.
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

// ListVotePointsWithoutOrigin returns vote points still needing origin
// resolution, in (session, beteckning, forslagspunkt) order after `after`.
func (r *Repository) ListVotePointsWithoutOrigin(ctx context.Context, after ports.VotePoint, limit int) ([]ports.VotePoint, error) {
	// Reads voteringar directly: one row per votering, so no DISTINCT is needed.
	// Rows are keyed on session as well, because the same beteckning and
	// förslagspunkt recur every riksmöte.
	const q = `SELECT beteckning, forslagspunkt, session, COALESCE(dok_id, '')
		FROM voteringar
		WHERE origin_enriched = false
		  AND (session, beteckning, forslagspunkt) > ($2, $3, $4)
		ORDER BY session, beteckning, forslagspunkt
		LIMIT $1`

	rows, err := r.db.Query(ctx, q, limit, after.Session, after.Beteckning, after.Forslagspunkt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pp []ports.VotePoint
	for rows.Next() {
		var v ports.VotePoint
		if err := rows.Scan(&v.Beteckning, &v.Forslagspunkt, &v.Session, &v.DokID); err != nil {
			return nil, err
		}
		pp = append(pp, v)
	}
	return pp, rows.Err()
}

// MarkAllOriginsForRecompute flags every votering for re-enrichment. It leaves
// the stored origin untouched; the recompute overwrites it per riksmöte.
func (r *Repository) MarkAllOriginsForRecompute(ctx context.Context) (int64, error) {
	tag, err := r.db.Exec(ctx, `UPDATE voteringar SET origin_enriched = false WHERE origin_enriched = true`)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

// UpdateProposalOriginForPoint applies one origin to the vote point and reports
// how many rows it touched.
//
// The count now means voteringar rather than ballots: origin is a property of
// the votering, so one point that previously updated ~350 rows updates one or
// two. Callers use it only to detect "nothing matched", which still holds.
//
// Keyed on session + beteckning + förslagspunkt. Without the session, SkU16
// punkt 6 of 2021/22 overwrote the same punkt of every other riksmöte.
func (r *Repository) UpdateProposalOriginForPoint(ctx context.Context, session, beteckning, forslagspunkt string, o domain.ProposalOrigin) (int64, error) {
	const q = `
		UPDATE voteringar SET
			proposed_by_party = $4,
			proposal_type     = $5,
			proposal_dok_id   = $6,
			document_title    = $7,
			origin_enriched   = true
		WHERE session = $1 AND beteckning = $2 AND forslagspunkt = $3`

	var proposalType *string
	if o.ProposalType != "" {
		s := string(o.ProposalType)
		proposalType = &s
	}
	var proposedBy, dokID, title *string
	if o.ProposedByParty != "" {
		proposedBy = &o.ProposedByParty
	}
	if o.ProposalDokID != "" {
		dokID = &o.ProposalDokID
	}
	if o.DocumentTitle != "" {
		title = &o.DocumentTitle
	}

	tag, err := r.db.Exec(ctx, q, session, beteckning, forslagspunkt, proposedBy, proposalType, dokID, title)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

func (r *Repository) ListDistinctVotes(ctx context.Context, f ports.ListDistinctVotesFilter) (ports.ListDistinctVotesResult, error) {
	var total int
	// DISTINCT survives the move to voteringar, unlike elsewhere: it is not
	// collapsing duplicated ballots but two genuinely separate voteringar that
	// share a beteckning and förslagspunkt — 18 of them in 2022-2026. This
	// listing is by vote point, so those still count once.
	if err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM (SELECT DISTINCT beteckning, forslagspunkt FROM voteringar WHERE origin_enriched = true) sub`,
	).Scan(&total); err != nil {
		return ports.ListDistinctVotesResult{}, err
	}

	offset := (f.Page - 1) * f.PageSize
	const q = `SELECT DISTINCT ON (beteckning, forslagspunkt)
		beteckning, forslagspunkt, COALESCE(document_title, ''), proposed_by_party, proposal_type
		FROM voteringar
		WHERE origin_enriched = true
		ORDER BY beteckning, forslagspunkt
		LIMIT $1 OFFSET $2`

	rows, err := r.db.Query(ctx, q, f.PageSize, offset)
	if err != nil {
		return ports.ListDistinctVotesResult{}, err
	}
	defer rows.Close()

	var ss []ports.VoteSummary
	for rows.Next() {
		var s ports.VoteSummary
		var proposedBy, propType *string
		if err := rows.Scan(&s.Beteckning, &s.Forslagspunkt, &s.DocumentTitle, &proposedBy, &propType); err != nil {
			return ports.ListDistinctVotesResult{}, err
		}
		if proposedBy != nil {
			s.ProposedByParty = *proposedBy
		}
		if propType != nil {
			s.ProposalType = *propType
		}
		ss = append(ss, s)
	}
	return ports.ListDistinctVotesResult{Votes: ss, Total: total}, rows.Err()
}

func (r *Repository) ListDistinctByCommitteePrefix(ctx context.Context, prefix string) ([]ports.VoteSummary, error) {
	const q = `SELECT DISTINCT beteckning, forslagspunkt, document_title, proposed_by_party, proposal_type
		FROM voteringar
		WHERE beteckning LIKE $1 || '%'
		  AND origin_enriched = true
		  AND document_title IS NOT NULL
		ORDER BY beteckning, forslagspunkt`

	rows, err := r.db.Query(ctx, q, prefix)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ss []ports.VoteSummary
	for rows.Next() {
		var s ports.VoteSummary
		var proposedBy, propType *string
		if err := rows.Scan(&s.Beteckning, &s.Forslagspunkt, &s.DocumentTitle, &proposedBy, &propType); err != nil {
			return nil, err
		}
		if proposedBy != nil {
			s.ProposedByParty = *proposedBy
		}
		if propType != nil {
			s.ProposalType = *propType
		}
		ss = append(ss, s)
	}
	return ss, rows.Err()
}

type scanner interface {
	Scan(dest ...any) error
}

func scanVote(s scanner) (*domain.Vote, error) {
	var v domain.Vote
	var dokID, proposedByParty, proposalType, proposalDokID, documentTitle *string
	err := s.Scan(
		&v.VoteringID, &v.PoliticianID, &v.Party, &v.VoteResult,
		&v.Beteckning, &v.Forslagspunkt, &v.Session, &dokID,
		&proposedByParty, &proposalType, &proposalDokID, &documentTitle,
		&v.OriginEnriched, &v.CreatedAt,
	)
	if dokID != nil {
		v.DokID = *dokID
	}
	if proposedByParty != nil {
		v.ProposalOrigin.ProposedByParty = *proposedByParty
	}
	if proposalType != nil {
		v.ProposalOrigin.ProposalType = domain.ProposalType(*proposalType)
	}
	if proposalDokID != nil {
		v.ProposalOrigin.ProposalDokID = *proposalDokID
	}
	if documentTitle != nil {
		v.ProposalOrigin.DocumentTitle = *documentTitle
	}
	return &v, err
}

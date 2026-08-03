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
	args := []any{f.PoliticianID}
	where := "WHERE politician_id = $1"
	n := 2

	if f.Session != "" {
		where += fmt.Sprintf(" AND session = $%d", n)
		args = append(args, f.Session)
		n++
	}

	var total int
	if err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM votes "+where, args...).Scan(&total); err != nil {
		return ports.ListVotesResult{}, err
	}

	offset := (f.Page - 1) * f.PageSize
	args = append(args, f.PageSize, offset)
	q := fmt.Sprintf(`SELECT id, votering_id, politician_id, party, vote_result, beteckning, forslagspunkt,
		session, dok_id, proposed_by_party, proposal_type, proposal_dok_id, document_title, origin_enriched, created_at
		FROM votes %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, where, n, n+1)

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

func (r *Repository) ListByBeteckning(ctx context.Context, beteckning, punkt string) ([]*domain.Vote, error) {
	const q = `SELECT id, votering_id, politician_id, party, vote_result, beteckning, forslagspunkt,
		session, dok_id, proposed_by_party, proposal_type, proposal_dok_id, document_title, origin_enriched, created_at
		FROM votes WHERE beteckning = $1 AND forslagspunkt = $2 ORDER BY party, politician_id`

	rows, err := r.db.Query(ctx, q, beteckning, punkt)
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
	const q = `
		INSERT INTO votes (votering_id, politician_id, party, vote_result, beteckning, forslagspunkt, session, dok_id, system_datum)
		SELECT $1, $2, $3, $4, $5, $6, $7, $8, $9
		WHERE EXISTS (SELECT 1 FROM politicians WHERE intressent_id = $2)
		ON CONFLICT (votering_id, politician_id) DO UPDATE SET
			vote_result   = EXCLUDED.vote_result,
			beteckning    = EXCLUDED.beteckning,
			forslagspunkt = EXCLUDED.forslagspunkt,
			session       = EXCLUDED.session,
			dok_id        = EXCLUDED.dok_id,
			system_datum  = EXCLUDED.system_datum`

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

func (r *Repository) UpdateProposalOrigin(ctx context.Context, voteringID, politicianID string, o domain.ProposalOrigin) error {
	const q = `
		UPDATE votes SET
			proposed_by_party = $3,
			proposal_type     = $4,
			proposal_dok_id   = $5,
			document_title    = $6,
			origin_enriched   = true
		WHERE votering_id = $1 AND politician_id = $2`

	var proposalType *string
	if o.ProposalType != "" {
		s := string(o.ProposalType)
		proposalType = &s
	}
	var proposedBy *string
	if o.ProposedByParty != "" {
		proposedBy = &o.ProposedByParty
	}
	var proposalDokID *string
	if o.ProposalDokID != "" {
		proposalDokID = &o.ProposalDokID
	}
	var docTitle *string
	if o.DocumentTitle != "" {
		docTitle = &o.DocumentTitle
	}

	_, err := r.db.Exec(ctx, q, voteringID, politicianID,
		proposedBy, proposalType, proposalDokID, docTitle)
	return err
}

func (r *Repository) ListWithoutOrigin(ctx context.Context, limit int) ([]*domain.Vote, error) {
	const q = `SELECT id, votering_id, politician_id, party, vote_result, beteckning, forslagspunkt,
		session, dok_id, proposed_by_party, proposal_type, proposal_dok_id, document_title, origin_enriched, created_at
		FROM votes WHERE origin_enriched = false LIMIT $1`

	rows, err := r.db.Query(ctx, q, limit)
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

func (r *Repository) ListDistinctVotes(ctx context.Context, f ports.ListDistinctVotesFilter) (ports.ListDistinctVotesResult, error) {
	var total int
	if err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM (SELECT DISTINCT beteckning, forslagspunkt FROM votes WHERE origin_enriched = true) sub`,
	).Scan(&total); err != nil {
		return ports.ListDistinctVotesResult{}, err
	}

	offset := (f.Page - 1) * f.PageSize
	const q = `SELECT DISTINCT ON (beteckning, forslagspunkt)
		beteckning, forslagspunkt, COALESCE(document_title, ''), proposed_by_party, proposal_type
		FROM votes
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
		FROM votes
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
		&v.ID, &v.VoteringID, &v.PoliticianID, &v.Party, &v.VoteResult,
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

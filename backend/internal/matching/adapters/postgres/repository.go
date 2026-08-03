package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"riksdagskollen/internal/matching/domain"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) UpsertGoalVoteMatch(ctx context.Context, m *domain.GoalVoteMatch) error {
	const q = `
		INSERT INTO goal_vote_matches
			(goal_id, beteckning, forslagspunkt, relevance_score, aligned_direction,
			 explanation, proposed_by_party, proposal_type, context_note, document_title)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		ON CONFLICT (goal_id, beteckning, forslagspunkt) DO UPDATE SET
			relevance_score   = EXCLUDED.relevance_score,
			aligned_direction = EXCLUDED.aligned_direction,
			explanation       = EXCLUDED.explanation,
			context_note      = EXCLUDED.context_note,
			document_title    = EXCLUDED.document_title`
	_, err := r.db.Exec(ctx, q,
		m.GoalID, m.Beteckning, m.Forslagspunkt, m.RelevanceScore,
		string(m.AlignedDirection), m.Explanation, m.ProposedByParty, m.ProposalType,
		m.ContextNote, m.DocumentTitle)
	return err
}

func (r *Repository) ListMatchesByGoal(ctx context.Context, goalID int) ([]*domain.GoalVoteMatch, error) {
	const q = `SELECT id, goal_id, beteckning, forslagspunkt, relevance_score, aligned_direction,
		explanation, proposed_by_party, proposal_type, context_note, document_title, verified, created_at
		FROM goal_vote_matches WHERE goal_id = $1 ORDER BY relevance_score DESC`

	rows, err := r.db.Query(ctx, q, goalID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mm []*domain.GoalVoteMatch
	for rows.Next() {
		var m domain.GoalVoteMatch
		var dir string
		var explanation, proposedByParty, proposalType, contextNote, documentTitle *string
		err := rows.Scan(&m.ID, &m.GoalID, &m.Beteckning, &m.Forslagspunkt, &m.RelevanceScore,
			&dir, &explanation, &proposedByParty, &proposalType,
			&contextNote, &documentTitle, &m.Verified, &m.CreatedAt)
		if err != nil {
			return nil, err
		}
		m.AlignedDirection = domain.AlignmentDirection(dir)
		if explanation != nil {
			m.Explanation = *explanation
		}
		if proposedByParty != nil {
			m.ProposedByParty = *proposedByParty
		}
		if proposalType != nil {
			m.ProposalType = *proposalType
		}
		if contextNote != nil {
			m.ContextNote = *contextNote
		}
		if documentTitle != nil {
			m.DocumentTitle = *documentTitle
		}
		mm = append(mm, &m)
	}
	return mm, rows.Err()
}

func (r *Repository) UpsertPromiseVoteMatch(ctx context.Context, m *domain.PromiseVoteMatch) error {
	const q = `
		INSERT INTO promise_vote_matches
			(promise_id, vote_id, relevance_score, alignment, explanation, proposed_by_party, proposal_type)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		ON CONFLICT (promise_id, vote_id) DO UPDATE SET
			relevance_score  = EXCLUDED.relevance_score,
			alignment        = EXCLUDED.alignment,
			explanation      = EXCLUDED.explanation`
	_, err := r.db.Exec(ctx, q,
		m.PromiseID, m.VoteID, m.RelevanceScore, string(m.Alignment),
		m.Explanation, m.ProposedByParty, m.ProposalType)
	return err
}

func (r *Repository) ListMatchesByPromise(ctx context.Context, promiseID int) ([]*domain.PromiseVoteMatch, error) {
	const q = `SELECT id, promise_id, vote_id, relevance_score, alignment, explanation,
		proposed_by_party, proposal_type, verified, created_at
		FROM promise_vote_matches WHERE promise_id = $1 ORDER BY relevance_score DESC`

	rows, err := r.db.Query(ctx, q, promiseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mm []*domain.PromiseVoteMatch
	for rows.Next() {
		var m domain.PromiseVoteMatch
		var alignment string
		var explanation, proposedByParty, proposalType *string
		err := rows.Scan(&m.ID, &m.PromiseID, &m.VoteID, &m.RelevanceScore,
			&alignment, &explanation, &proposedByParty, &proposalType,
			&m.Verified, &m.CreatedAt)
		if err != nil {
			return nil, err
		}
		m.Alignment = domain.PromiseAlignment(alignment)
		if explanation != nil {
			m.Explanation = *explanation
		}
		if proposedByParty != nil {
			m.ProposedByParty = *proposedByParty
		}
		if proposalType != nil {
			m.ProposalType = *proposalType
		}
		mm = append(mm, &m)
	}
	return mm, rows.Err()
}

const scorecardColumns = `party, goal_id, goal_text, topic, relevant_votes`

func (r *Repository) GetPartyScorecard(ctx context.Context, party string) ([]*domain.ScorecardRow, error) {
	const q = `SELECT ` + scorecardColumns + `
		FROM party_scorecards WHERE party = $1 ORDER BY topic, goal_id`
	return r.queryScorecards(ctx, q, party)
}

func (r *Repository) GetAllPartyScorecards(ctx context.Context) ([]*domain.ScorecardRow, error) {
	const q = `SELECT ` + scorecardColumns + `
		FROM party_scorecards ORDER BY party, topic, goal_id`
	return r.queryScorecards(ctx, q)
}

func (r *Repository) RefreshScorecards(ctx context.Context) error {
	_, err := r.db.Exec(ctx, "REFRESH MATERIALIZED VIEW CONCURRENTLY party_scorecards")
	return err
}

func (r *Repository) queryScorecards(ctx context.Context, q string, args ...any) ([]*domain.ScorecardRow, error) {
	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rr []*domain.ScorecardRow
	for rows.Next() {
		var row domain.ScorecardRow
		if err := rows.Scan(&row.Party, &row.GoalID, &row.GoalText, &row.Topic,
			&row.RelevantVotes); err != nil {
			return nil, err
		}
		rr = append(rr, &row)
	}
	return rr, rows.Err()
}

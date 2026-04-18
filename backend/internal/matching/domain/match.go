package domain

import "time"

type AlignmentDirection string

const (
	AlignmentJa      AlignmentDirection = "Ja"
	AlignmentNej     AlignmentDirection = "Nej"
	AlignmentAvstar  AlignmentDirection = "Avstår"
	AlignmentUnclear AlignmentDirection = "unclear"
)

type PromiseAlignment string

const (
	AlignmentSupports    PromiseAlignment = "supports"
	AlignmentContradicts PromiseAlignment = "contradicts"
	AlignmentUnknown     PromiseAlignment = "unclear"
)

// GoalVoteMatch links a party goal to a relevant parliamentary vote.
type GoalVoteMatch struct {
	ID               int                `json:"id"`
	GoalID           int                `json:"goalId"`
	Beteckning       string             `json:"beteckning"`
	Forslagspunkt    string             `json:"forslagspunkt"`
	RelevanceScore   float64            `json:"relevanceScore"`
	AlignedDirection AlignmentDirection `json:"alignedDirection"`
	Explanation      string             `json:"explanation,omitempty"`
	ProposedByParty  string             `json:"proposedByParty,omitempty"`
	ProposalType     string             `json:"proposalType,omitempty"`
	ContextNote      string             `json:"contextNote,omitempty"`
	DocumentTitle    string             `json:"documentTitle,omitempty"`
	Verified         bool               `json:"verified"`
	CreatedAt        time.Time          `json:"createdAt"`
}

// PromiseVoteMatch links a politician's promise to a specific vote.
type PromiseVoteMatch struct {
	ID              int              `json:"id"`
	PromiseID       int              `json:"promiseId"`
	VoteID          int              `json:"voteId"`
	RelevanceScore  float64          `json:"relevanceScore"`
	Alignment       PromiseAlignment `json:"alignment"`
	Explanation     string           `json:"explanation,omitempty"`
	ProposedByParty string           `json:"proposedByParty,omitempty"`
	ProposalType    string           `json:"proposalType,omitempty"`
	Verified        bool             `json:"verified"`
	CreatedAt       time.Time        `json:"createdAt"`
}

// ScorecardRow is a single row from the party_scorecards materialized view.
type ScorecardRow struct {
	Party         string  `json:"party"`
	GoalID        int     `json:"goalId"`
	GoalText      string  `json:"goalText"`
	Topic         string  `json:"topic"`
	RelevantVotes int     `json:"relevantVotes"`
	AlignmentPct  float64 `json:"alignmentPct"`
}

// RelevanceScore is returned by the AI service for a goal-vote pair.
type RelevanceScore struct {
	Relevant         bool               `json:"relevant"`
	AlignedDirection AlignmentDirection `json:"alignedDirection"`
	Confidence       float64            `json:"confidence"`
	ContextNote      string             `json:"contextNote,omitempty"`
}

// ExtractedPromise is a promise extracted from a speech by the AI service.
type ExtractedPromise struct {
	PromiseText string   `json:"promiseText"`
	Topic       string   `json:"topic"`
	Specificity string   `json:"specificity"`
	Keywords    []string `json:"keywords,omitempty"`
}

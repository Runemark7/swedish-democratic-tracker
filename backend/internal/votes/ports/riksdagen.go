package ports

import (
	"context"
	"time"

	"riksdagskollen/internal/votes/domain"
)

type FetchVotesFilter struct {
	Session      string
	Party        string
	PoliticianID string
	Beteckning   string
	Size         int
	Since        time.Time // client-side cutoff: skip votes on or before this date
	// NOTE: /voteringlista does NOT paginate. Verified 2026-08-01: the `p`
	// parameter is silently ignored (p=1, p=2 and p=3 return byte-identical
	// rows) and `sz` is capped at 10 000 regardless of the value requested.
	// A whole party-riksmöte (~81 000 ballots) therefore cannot be retrieved
	// by this endpoint alone.
	//
	// Callers must instead partition by Beteckning: one request per betänkande
	// returns every party's ballots for it and stays well under the cap
	// (AU9 = 1 047 rows). Enumerate betänkanden via /dokumentlista, which does
	// paginate correctly and reports @traffar.
}

// VoteringRef identifies one votering — a single förslagspunkt decided by a
// vote. Enumerated from /dokumentlista, which (unlike /voteringlista) paginates
// correctly and reports a total.
type VoteringRef struct {
	Beteckning  string // "UbU31"
	Organ       string // "UbU"
	DokID       string
	Date        string // "2026-06-17"
	SystemDatum time.Time
}

// RiksdagDocument is a recent betänkande (committee report) from the Riksdagen Open Data API.
type RiksdagDocument struct {
	Title      string
	Organ      string // e.g. "SoU", "TU", "UbU", "CU"
	Date       string // "2025-04-15"
	Beteckning string // e.g. "SoU12"
}

// BetankandeInfo is the metadata for a single betänkande from the dokumentlista API.
type BetankandeInfo struct {
	DokID   string
	Title   string
	Date    string
	Status  string
	Session string
	Organ   string
}

type RiksdagenVoteClient interface {
	FetchVotes(ctx context.Context, f FetchVotesFilter) ([]*domain.Vote, error)
	// ListVoteringar enumerates voteringar for a riksmöte, newest first.
	// Returns one page plus the total the API reports (@traffar), which is also
	// the coverage denominator. page is 1-based.
	ListVoteringar(ctx context.Context, rm string, page, size int) ([]VoteringRef, int, error)
	FetchDocumentStatus(ctx context.Context, dokID string) (*domain.DocumentStatus, error)
	// FetchDocuments returns the most recent betänkanden from the given committee organs.
	FetchDocuments(ctx context.Context, organs []string, count int) ([]RiksdagDocument, error)
	FetchBetankandeByBeteckning(ctx context.Context, beteckning string) (*BetankandeInfo, error)
}

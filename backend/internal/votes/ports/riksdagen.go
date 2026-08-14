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

// RiksdagDocument is a betänkande (committee report) from the Riksdagen Open
// Data API. It is not necessarily a decision: the list carries betänkanden that
// have only been planned, and they sort newest-first alongside decided ones.
type RiksdagDocument struct {
	Title      string
	Organ      string // e.g. "SoU", "TU", "UbU", "CU"
	Beteckning string // e.g. "SoU12"

	// Date is `datum`, which tracks when the document was published or last
	// touched — never when the chamber decided. On every decided betänkande
	// checked on 2026-08-14, `beslutsdag` fell one to six days after it. Do not
	// render this as a decision date.
	Date string // "2025-04-15"

	// DecisionDate is `beslutsdag`, empty when the chamber has not decided.
	// Its emptiness is the authoritative "not yet decided" signal — `beslutad`
	// is not decoded, because it arrives with an inconsistent JSON type.
	DecisionDate string

	// Status is Riksdagen's own lifecycle word, e.g. "Webbpublicering" for a
	// published decision or "planerat" for a betänkande scheduled but not yet
	// written. Shown verbatim; never translated into a claim of our own.
	Status string
}

// Decided reports whether the chamber has decided this betänkande.
//
// A feed sorted by `datum` puts planned betänkanden for a coming riksmöte at the
// top — 15 of 40 on 2026-08-14, five of them in the first eight rows. Anything
// presenting these as decisions publishes a future event as a past one.
func (d RiksdagDocument) Decided() bool { return d.DecisionDate != "" }

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
	// FetchRecentBetankanden returns recent betänkanden across all committees.
	FetchRecentBetankanden(ctx context.Context, count int) ([]RiksdagDocument, error)
	FetchBetankandeByBeteckning(ctx context.Context, beteckning string) (*BetankandeInfo, error)
}

package ports

import (
	"context"
	"strings"
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

	// DecisionDate is `beslutsdag`. Riksdagen fills it for planned items too:
	// on 2026-10-27, betänkanden with status "planerat" carried beslutsdag
	// values in December 2026, and published-but-not-yet-voted reports
	// (status "Webbpublicering", beslutad 0) carried the day the chamber is
	// scheduled to vote. Empty still means no date. It is not, by itself,
	// evidence that the chamber has decided.
	DecisionDate string

	// Status is Riksdagen's own lifecycle word, e.g. "Webbpublicering" for a
	// published report or "planerat" for a betänkande scheduled but not yet
	// written. Shown verbatim; never translated into a claim of our own.
	Status string

	// ChamberDecided is Riksdagen's `beslutad` flag when the payload carried
	// one. 0 means the chamber has not decided, including when DecisionDate
	// is a future scheduled day. Nil means the field was absent or unreadable
	// and Decided falls back to status and the date.
	ChamberDecided *bool

	// Riksmote is Riksdagen's `rm`, e.g. "2026/27". Beteckning numbering
	// restarts each riksmöte, so a link built from the beteckning alone can
	// open another year's betänkande.
	Riksmote string
}

// Decided reports whether the chamber has decided this betänkande as of now.
//
// A feed sorted by `datum` puts planned betänkanden at the top. Treating a
// non-empty beslutsdag as a decision published those as things that had
// already happened: on 2026-10-27 the newest rows were status "planerat"
// with beslutsdag in December and beslutad 0.
func (d RiksdagDocument) Decided(now time.Time) bool {
	if d.ChamberDecided != nil {
		return *d.ChamberDecided
	}
	if strings.EqualFold(strings.TrimSpace(d.Status), "planerat") {
		return false
	}
	if d.DecisionDate == "" {
		return false
	}
	t, err := time.Parse("2006-01-02", d.DecisionDate)
	if err != nil {
		return false
	}
	utc := now.UTC()
	today := time.Date(utc.Year(), utc.Month(), utc.Day(), 0, 0, 0, 0, time.UTC)
	day := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
	return !day.After(today)
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
	// FetchRecentBetankanden returns recent betänkanden across all committees.
	FetchRecentBetankanden(ctx context.Context, count int) ([]RiksdagDocument, error)
	// FetchBetankandeByBeteckning looks up one betänkande. session, when set,
	// is the riksmöte ("2026/27"): without it the same beteckning matches
	// every year and the first hit is whichever year the API returns.
	FetchBetankandeByBeteckning(ctx context.Context, beteckning, session string) (*BetankandeInfo, error)
}

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
}

// RiksdagDocument is a recent betänkande (committee report) from the Riksdagen Open Data API.
type RiksdagDocument struct {
	Title      string
	Organ      string // e.g. "SoU", "TU", "UbU", "CU"
	Date       string // "2025-04-15"
	Beteckning string // e.g. "SoU12"
}

type RiksdagenVoteClient interface {
	FetchVotes(ctx context.Context, f FetchVotesFilter) ([]*domain.Vote, error)
	FetchDocumentStatus(ctx context.Context, dokID string) (*domain.DocumentStatus, error)
	// FetchDocuments returns the most recent betänkanden from the given committee organs.
	FetchDocuments(ctx context.Context, organs []string, count int) ([]RiksdagDocument, error)
}

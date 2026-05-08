package ports

import (
	"context"
	"time"

	"riksdagskollen/internal/speeches/domain"
)

type FetchSpeechesFilter struct {
	Session      string
	Party        string
	PoliticianID string
	Since        time.Time
	Size         int
}

type RiksdagenSpeechClient interface {
	FetchSpeeches(ctx context.Context, f FetchSpeechesFilter) ([]*domain.Speech, error)
	// FetchSpeechText fetches the full prose text for one anförande.
	// The /anforandelista endpoint only returns metadata; the body lives at
	// /anforande/{dokID}-{anforandeNummer}.json.
	FetchSpeechText(ctx context.Context, dokID, anforandeNummer string) (string, error)
}

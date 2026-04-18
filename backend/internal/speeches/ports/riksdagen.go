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
}

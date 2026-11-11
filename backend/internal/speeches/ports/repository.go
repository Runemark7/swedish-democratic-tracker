package ports

import (
	"context"
	"time"

	"riksdagskollen/internal/speeches/domain"
)

type SpeechRepository interface {
	GetByID(ctx context.Context, id int) (*domain.Speech, error)
	ListByPolitician(ctx context.Context, politicianID string) ([]*domain.Speech, error)
	ListRecent(ctx context.Context, limit int) ([]*domain.Speech, error)
	UpsertMany(ctx context.Context, ss []*domain.Speech) error
	ListUnprocessed(ctx context.Context, limit int) ([]*domain.Speech, error)
	MarkProcessed(ctx context.Context, id int) error
	// ListMissingText returns speeches with empty speech_text, ordered
	// newest first so user-facing pages get backfilled first.
	ListMissingText(ctx context.Context, limit int) ([]*domain.Speech, error)
	// UpdateText overwrites speech_text for one row.
	UpdateText(ctx context.Context, id int, text string) error
	// ListByDocument returns speeches whose rel_dok_id matches the
	// given betänkande document id, ordered by anforande_nummer ASC
	// (chronological within the debate).
	ListByDocument(ctx context.Context, dokID string) ([]*domain.Speech, error)
	// ListByParty returns speeches by all members of one party,
	// newest first, capped at limit (default 50, max 200).
	ListByParty(ctx context.Context, party string, limit int) ([]*domain.Speech, error)

	// DebateCandidates lists the stored betänkanden with a known debate date.
	DebateCandidates(ctx context.Context) ([]domain.DebateCandidate, error)
	// ListForLinking reads at most limit speeches with id > afterID, in id
	// order, with the link they currently carry.
	ListForLinking(ctx context.Context, afterID, limit int) ([]LinkRow, error)
	// SetLinks writes links; an empty DokID clears a link.
	SetLinks(ctx context.Context, links []LinkUpdate) error
}

// LinkRow is a speech as the linking reads it.
type LinkRow struct {
	ID                         int
	RelatedDokID, TopicHeading string
	Date                       time.Time
	CurrentDokID, CurrentRule  string
}

// LinkUpdate sets one speech's link; an empty DokID clears it.
type LinkUpdate struct {
	ID          int
	DokID, Rule string
}

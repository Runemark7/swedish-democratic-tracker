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
	// order, with the link they currently carry. A riksmote limits it to that
	// riksmöte; empty reads them all.
	ListForLinking(ctx context.Context, riksmote string, afterID, limit int) ([]LinkRow, error)

	// MissingTextIn lists up to limit speeches of the riksmöte with no text.
	MissingTextIn(ctx context.Context, riksmote string, limit int) ([]*domain.Speech, error)
	// RiksmoteCounts counts what is stored for one riksmöte.
	RiksmoteCounts(ctx context.Context, riksmote string) (domain.RiksmoteCounts, error)
	// StorageSizes is the speeches table's size (heap and TOAST) and its indexes'.
	StorageSizes(ctx context.Context) (table, index int64, err error)
	// SaveBackfill records one run's work on a riksmöte (see domain.BackfillRun).
	SaveBackfill(ctx context.Context, run domain.BackfillRun) error
	// ListBackfill returns the riksmöten the backfill has worked on, newest
	// first, with the counts read live from the speeches.
	ListBackfill(ctx context.Context) ([]domain.Backfill, error)
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

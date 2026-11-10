// Package ports declares the interfaces the motions service depends on.
package ports

import (
	"context"
	"errors"
	"time"

	"riksdagskollen/internal/motions/domain"
)

// ErrNotFound is returned when no motion has the given riksmöte and beteckning.
var ErrNotFound = errors.New("motion not found")

// MotionRef is one row of Riksdagen's motion listing.
type MotionRef struct {
	DokID       string
	SystemDatum time.Time
}

// RiksdagenClient reads motions from data.riksdagen.se.
type RiksdagenClient interface {
	// ListMotions returns one page (1-based) of the riksmöte's motions plus
	// the total Riksdagen reports for it (@traffar).
	ListMotions(ctx context.Context, rm string, page, size int) ([]MotionRef, int, error)
	// FetchMotion reads one motion with authors and yrkanden.
	FetchMotion(ctx context.Context, dokID string) (*domain.Motion, error)
	// FindBetankande looks a betänkande up by riksmöte and beteckning
	// ("2017/18", "SkU11"). found is false when Riksdagen lists none.
	FindBetankande(ctx context.Context, rm, bet string) (ref BetankandeRef, found bool, err error)
	// FetchBetankande reads a betänkande with its förslagspunkter.
	FetchBetankande(ctx context.Context, dokID string) (*domain.Betankande, error)
}

// BetankandeRef is one row of Riksdagen's betänkande listing.
type BetankandeRef struct {
	DokID       string
	SystemDatum time.Time
}

// BetRef names a betänkande some yrkande points at.
type BetRef struct{ Riksmote, Beteckning string }

// YrkandeKey identifies a stored yrkande to link.
type YrkandeKey struct {
	MotionDokID     string
	Number          int
	MotionRiksmote  string
	MotionBeteckn   string
	BehandlasIPunkt string
}

// Repository stores motions.
type Repository interface {
	// Upsert writes the motion and replaces its authors and yrkanden in one
	// transaction, keyed on dok_id.
	Upsert(ctx context.Context, m *domain.Motion) error
	Get(ctx context.Context, riksmote, beteckning string) (*domain.Motion, error)
	// StoredVersions maps dok_id to the systemdatum held, for one riksmöte.
	StoredVersions(ctx context.Context, riksmote string) (map[string]time.Time, error)
	UpsertCoverage(ctx context.Context, riksmote, source string, expected int) error
	// Coverage lists every recorded riksmöte × source with the held count.
	Coverage(ctx context.Context) ([]domain.Coverage, error)

	// PendingBetankanden lists, at most limit, betänkanden that yrkanden point
	// at and that are not held, were not found a day ago, or are older than the
	// newest motion pointing at them.
	PendingBetankanden(ctx context.Context, limit int) ([]BetRef, error)
	// SaveBetankande upserts the betänkande and its punkter.
	SaveBetankande(ctx context.Context, b *domain.Betankande) error
	// MarkBetankandeMissing records that Riksdagen lists no such betänkande.
	MarkBetankandeMissing(ctx context.Context, rm, bet string) error
	// YrkandenHandledIn lists the yrkanden whose betänkande is rm:bet.
	YrkandenHandledIn(ctx context.Context, rm, bet string) ([]YrkandeKey, error)
	// Punkter reads the stored förslagspunkter of a betänkande.
	Punkter(ctx context.Context, rm, bet string) ([]domain.Forslagspunkt, error)
	// ReplaceLinks sets the links of one yrkande into one betänkande,
	// idempotently.
	ReplaceLinks(ctx context.Context, key YrkandeKey, rm, bet string, links []domain.PunktLink) error
	// Decisions reads the stored links of a motion with their punkt, betänkande
	// and, for a votering we hold, the party tallies.
	Decisions(ctx context.Context, motionDokID string) ([]domain.YrkandeDecision, error)
}

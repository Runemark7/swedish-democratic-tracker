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
}

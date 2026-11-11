// Package ports declares the interfaces the questions service depends on.
package ports

import (
	"context"
	"errors"
	"time"

	"riksdagskollen/internal/questions/domain"
)

// ErrNotFound is returned when no question has the given identity.
var ErrNotFound = errors.New("question not found")

// Ref is one row of Riksdagen's listing.
type Ref struct {
	DokID       string
	SystemDatum time.Time
}

// RiksdagenClient reads frågor, svar and interpellationer.
type RiksdagenClient interface {
	// List returns one page (1-based) of the riksmöte's documents of a kind,
	// plus the total Riksdagen reports (@traffar).
	List(ctx context.Context, kind, rm string, page, size int) ([]Ref, int, error)
	// Fetch reads one fråga or interpellation with its people, text and
	// references. For a fråga it also reads the svar document it points at.
	Fetch(ctx context.Context, kind, dokID string) (*domain.Question, error)
}

// Repository stores questions.
type Repository interface {
	// Upsert writes the question, keyed on dok_id. Links are untouched.
	Upsert(ctx context.Context, q *domain.Question) error
	// StoredVersions maps dok_id to the systemdatum held, for one riksmöte
	// and kind.
	StoredVersions(ctx context.Context, kind, rm string) (map[string]time.Time, error)
	UpsertCoverage(ctx context.Context, rm, kind string, expected int) error
	Coverage(ctx context.Context) ([]domain.Coverage, error)

	// ForLinking lists the questions of a riksmöte with the text and
	// references the rules read.
	ForLinking(ctx context.Context, rm string) ([]domain.Question, error)
	// MotionDokID finds a held motion by "rm:beteckning".
	MotionDokID(ctx context.Context, rm, bet string) (string, bool, error)
	// ReplaceLinks sets the links of one question, idempotently.
	ReplaceLinks(ctx context.Context, questionDokID string, links []domain.Link) error

	// ByMember lists the questions a member asked or answered, newest first.
	ByMember(ctx context.Context, politicianID string, limit int) ([]domain.Item, error)
	// ByMotion lists the questions linked to a motion, newest first.
	ByMotion(ctx context.Context, riksmote, beteckning string, limit int) ([]domain.Item, error)
	// Get reads one question; the text and linked motions are included.
	Get(ctx context.Context, kind, riksmote, beteckning string) (*domain.Detail, error)
	// Debate lists the anföranden of an interpellation's debate: speeches whose
	// avsnittsrubrik names the interpellation, given on its debate date.
	Debate(ctx context.Context, d *domain.Detail) ([]domain.DebateSpeech, error)
}

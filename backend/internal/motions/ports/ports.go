// Package ports declares the interfaces the motions service depends on.
package ports

import (
	"context"
	"errors"
	"time"

	"riksdagskollen/internal/motions/domain"
)

// ErrHTMLUnavailable is returned when Riksdagen serves no HTML for a betänkande.
var ErrHTMLUnavailable = errors.New("betänkande html not available")

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
	// ListBetankanden returns one page (1-based) of the riksmöte's betänkanden
	// as Riksdagen lists them, with the total Riksdagen reports.
	ListBetankanden(ctx context.Context, rm string, page, size int) ([]BetankandeRef, int, error)
	// FetchBetankande reads a betänkande with its förslagspunkter.
	FetchBetankande(ctx context.Context, dokID string) (*domain.Betankande, error)
	// FetchBetankandeHTML reads the betänkande document and parses its
	// reservations, särskilda yttranden and the utskott's reasoning. It
	// returns ErrHTMLUnavailable when Riksdagen has no HTML for it.
	FetchBetankandeHTML(ctx context.Context, dokID string) (*domain.BetankandeHTML, error)
}

// BetankandeRef is one row of Riksdagen's betänkande listing.
type BetankandeRef struct {
	DokID       string
	Riksmote    string
	Beteckning  string
	Organ       string // "SkU"
	Published   time.Time
	SystemDatum time.Time
	Title       string
	// Status, DecisionDate (`beslutsdag`) and ChamberDecided (`beslutad`) are
	// Riksdagen's own, read live: a stored copy would go stale the day the
	// chamber decides. ChamberDecided is nil when the field was absent.
	Status         string
	DecisionDate   string
	ChamberDecided *bool
	// DebateDate is the "Debatt" date of the listing's statusrad; zero when the
	// betänkande was not debated (or has no date yet).
	DebateDate time.Time
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

// YrkandeCursor is a position in the (dok_id, nummer) order. The zero value
// is the start.
type YrkandeCursor struct {
	DokID  string
	Number int
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
	// newest motion pointing at them. A non-empty rm keeps only betänkanden
	// that motions of that riksmöte point at.
	PendingBetankanden(ctx context.Context, rm string, limit int) ([]BetRef, error)
	// SaveBetankande upserts the betänkande and its punkter.
	SaveBetankande(ctx context.Context, b *domain.Betankande) error
	// SaveBetankandeHTML replaces the stored statements of a betänkande and
	// sets the reasoning of each punkt whose rubrik has exactly one section.
	SaveBetankandeHTML(ctx context.Context, rm, bet string, h *domain.BetankandeHTML) error
	// MarkHTMLAttempt records a failed HTML read so it is retried tomorrow,
	// not on every run.
	MarkHTMLAttempt(ctx context.Context, rm, bet string) error
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
	// ListPage reads at most limit motions matching the SQL-expressible part
	// of the filter (everything but Outcomes), most recently decided first,
	// strictly after the cursor. Rows carry what the outcome is computed from.
	ListPage(ctx context.Context, f domain.ListFilter, after *domain.ListCursor, limit int) ([]domain.ListRow, error)

	// RecurringMembers lists every stored yrkande with the wording hash,
	// earliest riksmöte first. The members carry no Outcome yet.
	RecurringMembers(ctx context.Context, hash string) ([]domain.RecurringMember, error)
	// ConfirmedKravHashes lists the wording hashes the Redaktör confirmed to be
	// the same Återkommande krav as hash (transitively), hash itself excluded.
	ConfirmedKravHashes(ctx context.Context, hash string) ([]string, error)
	// SetWordingHash stores the hash of one yrkande.
	SetWordingHash(ctx context.Context, dokID string, number int, hash string) error
	// BackfillWordingHashes walks yrkanden without a hash in (dok_id, nummer)
	// order after the cursor, at most limit of them, writes the hash of each
	// and returns how many it wrote, how many it looked at, and the new
	// cursor. Looking at none means the walk is done.
	// GetBetankande reads a stored, found betänkande without its punkter.
	GetBetankande(ctx context.Context, rm, bet string) (*domain.BetankandeStored, error)
	// CanonicalBeteckning returns the stored spelling of a beteckning in a
	// riksmöte, matched case-insensitively across the betänkanden, the
	// listing and the voteringar; bet unchanged when none matches.
	CanonicalBeteckning(ctx context.Context, rm, bet string) (string, error)
	// Statements reads the reservations and särskilda yttranden of a betänkande.
	Statements(ctx context.Context, rm, bet string) ([]domain.Statement, error)
	// PunktMotions lists, per punkt number, the motions and yrkanden decided there.
	PunktMotions(ctx context.Context, rm, bet string) (map[string][]domain.PunktMotion, error)
	// Ballots reads the ballots of one votering (case-insensitive id).
	Ballots(ctx context.Context, voteringID string) ([]domain.Ballot, error)
	// VoteringPunkter lists the voteringar we hold for a riksmöte + beteckning.
	VoteringPunkter(ctx context.Context, rm, bet string) ([]domain.VoteringPunkt, error)
	// Debate reads at most limit speeches linked to the betänkande with this
	// dok_id, in speaking order, and how many are linked in all.
	Debate(ctx context.Context, betDokID string, limit int) (*domain.Debate, error)
	// ResolveBetankande finds the riksmöte for a beteckning (and, with punkt,
	// a votering's riksmöte). rm, when given, is taken as is if it exists;
	// otherwise the most recent riksmöte holding it wins.
	ResolveBetankande(ctx context.Context, bet, punkt, rm string) (string, bool, error)

	// UpsertBeslutListing stores Riksdagen's listing rows, replacing the
	// status, beslutsdag and beslutad of rows already held.
	UpsertBeslutListing(ctx context.Context, rows []BetankandeRef) error
	// BeslutListed counts the stored listing rows of a riksmöte.
	BeslutListed(ctx context.Context, rm string) (int, error)
	// ListBeslut reads at most limit listing rows matching f, upcoming first
	// (soonest first) and then decided (newest first), strictly after the
	// cursor. today is the date the decided rule compares beslutsdag with.
	ListBeslut(ctx context.Context, f domain.BeslutFilter, today time.Time, after *domain.BeslutCursor, limit int) ([]domain.BeslutRow, error)
	// BeslutDetails reads, for the given "riksmöte:beteckning" keys, the
	// beslutstyp of each stored förslagspunkt and each party's ballots per
	// votering.
	BeslutDetails(ctx context.Context, keys []string) (map[string]domain.BeslutDetail, error)

	BackfillWordingHashes(ctx context.Context, after YrkandeCursor, limit int) (written, scanned int, last YrkandeCursor, err error)

	// BackfillProgress lists the recorded progress rows, newest riksmöte first.
	BackfillProgress(ctx context.Context) ([]domain.BackfillProgress, error)
	// SaveBackfillProgress upserts one row. A zero CompletedAt keeps the row
	// open; it never clears a completion already recorded.
	SaveBackfillProgress(ctx context.Context, p domain.BackfillProgress) error
	// BetankandeCounts counts the distinct betänkanden the motions of rm point
	// at and how many of them are held.
	BetankandeCounts(ctx context.Context, rm string) (expected, held int, err error)
	// MotionCount counts the motions held for rm.
	MotionCount(ctx context.Context, rm string) (int, error)
	// BackfillRiksmoten lists, newest first, the riksmöten the voteringar and
	// the declared mandate periods cover.
	BackfillRiksmoten(ctx context.Context) ([]string, error)
	// DecisionVocabulary counts every distinct decision value stored, per field.
	DecisionVocabulary(ctx context.Context) ([]domain.DecisionValue, error)
	// HeldBetankanden lists every betänkande held, for relinking.
	HeldBetankanden(ctx context.Context) ([]BetRef, error)
	// OriginCoverage counts voteringar and how many of them have at least one
	// decided yrkande linked to their förslagspunkt.
	OriginCoverage(ctx context.Context) (voteringar, withOrigin int, err error)
}

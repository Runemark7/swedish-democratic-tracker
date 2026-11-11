package ports

import (
	"context"

	"riksdagskollen/internal/loften/domain"
)

// Repository is read-only. Implementations must return only public Löften
// (approved or archived); a proposed Löfte is unreachable through this port.
type Repository interface {
	ListPublic(ctx context.Context, party string) ([]*domain.Loefte, error)
	GetPublic(ctx context.Context, id int) (*domain.Loefte, error)
	// ConfirmedKopplingar returns only Bekräftade kopplingar (never föreslagen,
	// avvisad or bortsorterad) of the given Löften, with the Löfte's own
	// party's ballots in each votering.
	ConfirmedKopplingar(ctx context.Context, loefteIDs []int) (map[int][]domain.ConfirmedKoppling, error)
	// ConfirmedKopplingByID is nil when the koppling is not Bekräftad or its Löfte is not public.
	ConfirmedKopplingByID(ctx context.Context, id int) (*domain.ConfirmedKoppling, error)
	// Windows maps mandate period code to its riksmöten.
	Windows(ctx context.Context) (map[string][]string, error)
	// LoefteForBetankanden maps "riksmöte:beteckning" to the public Löften with
	// a Bekräftad koppling to a votering of that betänkande. Other kopplingar
	// and non-public Löften never appear.
	LoefteForBetankanden(ctx context.Context, keys []string) (map[string][]domain.LoefteRef, error)
	ListValmanifest(ctx context.Context, party string) ([]*domain.Valmanifest, error)
}

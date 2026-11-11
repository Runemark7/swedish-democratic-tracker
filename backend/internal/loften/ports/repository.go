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
	ListValmanifest(ctx context.Context, party string) ([]*domain.Valmanifest, error)
}

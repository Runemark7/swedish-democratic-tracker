package ports

import (
	"context"

	"riksdagskollen/internal/goals/domain"
)

type GoalRepository interface {
	Create(ctx context.Context, g *domain.Goal) error
	GetByID(ctx context.Context, id int) (*domain.Goal, error)
	ListAll(ctx context.Context) ([]*domain.Goal, error)
	ListByParty(ctx context.Context, party string) ([]*domain.Goal, error)
	ListByTopic(ctx context.Context, party, topic string) ([]*domain.Goal, error)
}

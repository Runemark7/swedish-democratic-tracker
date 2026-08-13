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
	// ListByCommittee returns the goals whose relevant_committees contains the
	// code, alphabetically by party. An empty result is a normal answer: one
	// committee has no seeded goals, and the caller states that as a gap in our
	// data rather than as the parties having nothing to say.
	ListByCommittee(ctx context.Context, code string) ([]*domain.Goal, error)
}

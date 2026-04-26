package riksdag

import (
	"context"
	"log/slog"
	"sort"

	"riksdagskollen/internal/riksdag/domain"
	"riksdagskollen/internal/riksdag/ports"
)

type Service struct {
	primary  ports.AuthorityClient
	fallback ports.AuthorityClient
}

func NewService(primary, fallback ports.AuthorityClient) *Service {
	return &Service{primary: primary, fallback: fallback}
}

func (s *Service) GetAuthorities(ctx context.Context) ([]domain.Authority, error) {
	data, err := s.primary.FetchAuthorities(ctx)
	if err != nil {
		slog.Warn("authority primary fetch failed, using fallback", "error", err)
		data, err = s.fallback.FetchAuthorities(ctx)
		if err != nil {
			return nil, err
		}
	}

	result := make([]domain.Authority, len(data))
	for i, d := range data {
		history := make([]domain.YearlyExpenditure, len(d.History))
		for j, h := range d.History {
			history[j] = domain.YearlyExpenditure{
				Year:            h.Year,
				ExpenditureMdkr: h.ExpenditureMdkr,
			}
		}
		result[i] = domain.Authority{
			Name:            d.Name,
			Role:            d.Role,
			Headcount:       d.Headcount,
			ExpenditureMdkr: d.ExpenditureMdkr,
			Year:            d.Year,
			History:         history,
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].ExpenditureMdkr > result[j].ExpenditureMdkr
	})
	return result, nil
}

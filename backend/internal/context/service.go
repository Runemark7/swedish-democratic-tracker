package context

import (
	"context"
	"fmt"

	"riksdagskollen/internal/context/domain"
	"riksdagskollen/internal/context/ports"
)

// topicMeta defines the static mapping from topic slug to budget areas,
// committees, and human-readable description.
type topicMetadata struct {
	UOCodes     []string
	Committees  []string
	Description string
}

var topicMeta = map[string]topicMetadata{
	"sjukvard": {
		UOCodes:     []string{"UO9", "UO10"},
		Committees:  []string{"SoU", "SfU"},
		Description: "Hälsovård, sjukvård och social omsorg (UO9) samt ekonomisk trygghet vid sjukdom och rehabilitering (UO10).",
	},
	"forsvar": {
		UOCodes:     []string{"UO6"},
		Committees:  []string{"FöU"},
		Description: "Försvar och samhällets krisberedskap (UO6) — inkluderar Försvarsmakten, Myndigheten för samhällsskydd och beredskap.",
	},
	"invandring": {
		UOCodes:     []string{"UO8", "UO13"},
		Committees:  []string{"SfU", "AU"},
		Description: "Migration och asylmottagning (UO8) samt integration och jämställdhet (UO13).",
	},
	"skatt": {
		UOCodes:     []string{"UO3", "UO27"},
		Committees:  []string{"SkU", "FiU"},
		Description: "Skatt, tull och exekution (UO3) samt avgifter till EU (UO27).",
	},
	"skola": {
		UOCodes:     []string{"UO15", "UO16"},
		Committees:  []string{"UbU"},
		Description: "Studiestöd (UO15) samt utbildning och universitetsforskning (UO16).",
	},
	"klimat": {
		UOCodes:     []string{"UO20"},
		Committees:  []string{"MJU"},
		Description: "Allmän miljö- och naturvård (UO20) — inkluderar klimatåtgärder, biologisk mångfald och hav.",
	},
	"brott": {
		UOCodes:     []string{"UO4"},
		Committees:  []string{"JuU"},
		Description: "Rättsväsendet (UO4) — inkluderar polis, åklagare, domstolar och kriminalvård.",
	},
	"bostad": {
		UOCodes:     []string{"UO18"},
		Committees:  []string{"CU"},
		Description: "Samhällsplanering och bostadsförsörjning (UO18) — inkluderar bostadsbidrag och planberedskap.",
	},
	"arbete": {
		UOCodes:     []string{"UO13", "UO14"},
		Committees:  []string{"AU"},
		Description: "Arbetsmarknad och arbetsliv (UO14) samt integration och jämställdhet (UO13).",
	},
	"energi": {
		UOCodes:     []string{"UO21"},
		Committees:  []string{"NU"},
		Description: "Energi (UO21) — inkluderar energiproduktion, elnät och energieffektivisering.",
	},
	"other": {
		UOCodes:     []string{},
		Committees:  []string{},
		Description: "Övriga politikområden utan direkt budgetkoppling.",
	},
}

// Service resolves topic context by combining the static topic map
// with live budget data from the repository.
type Service struct {
	repo ports.ContextRepository
}

func NewService(repo ports.ContextRepository) *Service {
	return &Service{repo: repo}
}

// GetTopicContext returns budget context for a given topic slug.
// Returns (nil, nil) for an unknown topic (caller should 404).
func (s *Service) GetTopicContext(ctx context.Context, topic string) (*domain.TopicContext, error) {
	meta, ok := topicMeta[topic]
	if !ok {
		return nil, nil
	}

	result := &domain.TopicContext{
		Topic:       topic,
		Description: meta.Description,
		Committees:  meta.Committees,
	}

	// Fetch budget trend for related UO codes (skip if no UO codes defined)
	if len(meta.UOCodes) > 0 {
		areas, err := s.repo.GetAreaTrend(ctx, meta.UOCodes)
		if err != nil {
			return nil, fmt.Errorf("get area trend: %w", err)
		}
		result.RelatedAreas = areas
	}

	// Fetch funding context for the latest budget year
	year, err := s.repo.GetLatestBudgetYear(ctx)
	if err != nil {
		return nil, fmt.Errorf("get latest budget year: %w", err)
	}
	if year > 0 {
		funding, err := s.repo.GetFundingContext(ctx, year)
		if err != nil {
			return nil, fmt.Errorf("get funding context: %w", err)
		}
		result.FundingContext = funding
	}

	return result, nil
}

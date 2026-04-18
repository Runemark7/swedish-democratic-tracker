package votes

import (
	"context"
	"fmt"
	"strings"
	"time"

	"riksdagskollen/internal/votes/domain"
	"riksdagskollen/internal/votes/ports"
)

type Service struct {
	repo      ports.VoteRepository
	riksdagen ports.RiksdagenVoteClient
}

func NewService(repo ports.VoteRepository, riksdagen ports.RiksdagenVoteClient) *Service {
	return &Service{repo: repo, riksdagen: riksdagen}
}

func (s *Service) SyncVotes(ctx context.Context, f ports.FetchVotesFilter) error {
	vv, err := s.riksdagen.FetchVotes(ctx, f)
	if err != nil {
		return err
	}
	return s.repo.UpsertMany(ctx, vv)
}

// EnrichOrigins fetches proposal-origin data for votes that haven't been enriched yet.
func (s *Service) EnrichOrigins(ctx context.Context, batchSize int) (int, error) {
	vv, err := s.repo.ListWithoutOrigin(ctx, batchSize)
	if err != nil {
		return 0, err
	}

	enriched := 0
	for _, v := range vv {
		origin, err := s.resolveOrigin(ctx, v)
		if err != nil {
			// Log and continue — don't abort the whole batch
			continue
		}
		if err := s.repo.UpdateProposalOrigin(ctx, v.VoteringID, v.PoliticianID, origin); err != nil {
			return enriched, err
		}
		enriched++
		// Rate limit: be gentle on the Riksdagen API
		time.Sleep(100 * time.Millisecond)
	}
	return enriched, nil
}

// resolveOrigin traces beteckning → dokumentstatus → prop/mot → proposing party.
func (s *Service) resolveOrigin(ctx context.Context, v *domain.Vote) (domain.ProposalOrigin, error) {
	// Use the vote's dok_id directly (e.g. "HC01AU10") — this is the correct format
	// for the dokumentstatus API. Falls back to constructing from session+beteckning.
	dokID := v.DokID
	if dokID == "" {
		session := strings.ReplaceAll(v.Session, "/", "")
		dokID = fmt.Sprintf("%s:%s", session, v.Beteckning)
	}

	ds, err := s.riksdagen.FetchDocumentStatus(ctx, dokID)
	if err != nil || ds == nil {
		return domain.ProposalOrigin{}, err
	}

	origin := domain.ProposalOrigin{
		DocumentTitle: ds.Title,
		ProposalDokID: dokID,
	}

	for _, ref := range ds.References {
		switch ref.RefDokTyp {
		case "prop":
			origin.ProposedByParty = "Regeringen"
			origin.ProposalType = domain.ProposalProp
			origin.ProposalDokID = ref.RefDokID
			return origin, nil
		case "mot":
			// Fetch the motion's document to get the proposing party
			motDS, err := s.riksdagen.FetchDocumentStatus(ctx, ref.RefDokID)
			if err == nil && motDS != nil {
				for _, mr := range motDS.References {
					if mr.PartyBet != "" {
						origin.ProposedByParty = mr.PartyBet
						break
					}
				}
			}
			origin.ProposalType = domain.ProposalMot
			origin.ProposalDokID = ref.RefDokID
			return origin, nil
		}
	}
	// Fallback: the betänkande itself
	origin.ProposalType = domain.ProposalBet
	return origin, nil
}

func (s *Service) ListByPolitician(ctx context.Context, f ports.ListVotesFilter) (ports.ListVotesResult, error) {
	if f.Page == 0 {
		f.Page = 1
	}
	if f.PageSize == 0 {
		f.PageSize = 50
	}
	return s.repo.ListByPolitician(ctx, f)
}

func (s *Service) ListDistinctByCommitteePrefix(ctx context.Context, prefix string) ([]ports.VoteSummary, error) {
	return s.repo.ListDistinctByCommitteePrefix(ctx, prefix)
}

func (s *Service) ListDistinctVotes(ctx context.Context, f ports.ListDistinctVotesFilter) (ports.ListDistinctVotesResult, error) {
	if f.Page == 0 {
		f.Page = 1
	}
	if f.PageSize == 0 {
		f.PageSize = 50
	}
	return s.repo.ListDistinctVotes(ctx, f)
}

func (s *Service) ListByBeteckning(ctx context.Context, beteckning, punkt string) ([]*domain.Vote, error) {
	return s.repo.ListByBeteckning(ctx, beteckning, punkt)
}

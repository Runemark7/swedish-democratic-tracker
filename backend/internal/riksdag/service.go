package riksdag

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/jackc/pgx/v5"

	"riksdagskollen/internal/riksdag/domain"
	"riksdagskollen/internal/riksdag/ports"
)

var ErrNotFound = errors.New("authority not found")

type Service struct {
	primary         ports.AuthorityClient
	fallback        ports.AuthorityClient
	headcount       ports.HeadcountClient       // optional; nil means use static data
	kpiRepo         ports.KpiRepository         // optional
	govRepo         ports.GovRepository         // optional
	agendaRepo      ports.AgendaRepository      // optional
	liveVotesRepo   ports.LiveVotesRepository   // optional
	agencyIntelRepo ports.AgencyIntelRepository // optional
	authorityRepo   ports.AuthorityRepository   // optional; backs ListMyndigheter
}

func (s *Service) SetKpiRepo(r ports.KpiRepository)                 { s.kpiRepo = r }
func (s *Service) SetGovRepo(r ports.GovRepository)                 { s.govRepo = r }
func (s *Service) SetAgendaRepo(r ports.AgendaRepository)           { s.agendaRepo = r }
func (s *Service) SetLiveVotesRepo(r ports.LiveVotesRepository)     { s.liveVotesRepo = r }
func (s *Service) SetAgencyIntelRepo(r ports.AgencyIntelRepository) { s.agencyIntelRepo = r }
func (s *Service) SetAuthorityRepo(r ports.AuthorityRepository)     { s.authorityRepo = r }

// MyndigheterList is the paginated response for ListMyndigheter.
type MyndigheterList struct {
	Items    []domain.RegisteredAuthority `json:"items"`
	Total    int                          `json:"total"`
	Page     int                          `json:"page"`
	PageSize int                          `json:"pageSize"`
}

// ListMyndigheter returns the searchable/filterable agency list backed by the
// authorities table (Phase 1 register + Phase 2 enrichment). Pagination is
// 1-indexed; page <= 0 is treated as 1, pageSize is clamped to [1, 200].
func (s *Service) ListMyndigheter(ctx context.Context, query string, underGovernment *bool, page, pageSize int) (*MyndigheterList, error) {
	if s.authorityRepo == nil {
		return nil, errors.New("authority repository not configured")
	}
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 50
	}
	if pageSize > 200 {
		pageSize = 200
	}

	filter := ports.AuthorityFilter{
		Query:           query,
		UnderGovernment: underGovernment,
		Limit:           pageSize,
		Offset:          (page - 1) * pageSize,
	}
	items, err := s.authorityRepo.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("list authorities: %w", err)
	}
	total, err := s.authorityRepo.Count(ctx, ports.AuthorityFilter{Query: query, UnderGovernment: underGovernment})
	if err != nil {
		return nil, fmt.Errorf("count authorities: %w", err)
	}
	return &MyndigheterList{Items: items, Total: total, Page: page, PageSize: pageSize}, nil
}

func (s *Service) ListKpis(ctx context.Context) ([]domain.Kpi, error) {
	if s.kpiRepo == nil {
		return nil, errors.New("kpi repository not configured")
	}
	return s.kpiRepo.ListKpis(ctx)
}

func (s *Service) GetGovernment(ctx context.Context) (*domain.Government, error) {
	if s.govRepo == nil {
		return nil, errors.New("government repository not configured")
	}
	return s.govRepo.GetCurrentGovernment(ctx)
}

func (s *Service) ListAgenda(ctx context.Context) ([]domain.AgendaItem, error) {
	if s.agendaRepo == nil {
		return nil, errors.New("agenda repository not configured")
	}
	return s.agendaRepo.ListAgenda(ctx)
}

func (s *Service) GetAgendaItem(ctx context.Context, id int) (*domain.AgendaItem, error) {
	if s.agendaRepo == nil {
		return nil, errors.New("agenda repository not configured")
	}
	item, err := s.agendaRepo.GetAgendaItem(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return item, nil
}

func (s *Service) ListLiveVotes(ctx context.Context, limit int) ([]domain.LiveVote, error) {
	if s.liveVotesRepo == nil {
		return nil, errors.New("live votes repository not configured")
	}
	return s.liveVotesRepo.ListLiveVotes(ctx, limit)
}

func NewService(primary, fallback ports.AuthorityClient) *Service {
	return &Service{primary: primary, fallback: fallback}
}

func NewServiceWithSCB(primary, fallback ports.AuthorityClient, headcount ports.HeadcountClient) *Service {
	return &Service{primary: primary, fallback: fallback, headcount: headcount}
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

	headcounts := s.fetchHeadcounts()

	result := make([]domain.Authority, len(data))
	for i, d := range data {
		history := make([]domain.YearlyExpenditure, len(d.History))
		for j, h := range d.History {
			history[j] = domain.YearlyExpenditure{
				Year:            h.Year,
				ExpenditureMdkr: h.ExpenditureMdkr,
				BudgetMdkr:      h.BudgetMdkr,
			}
		}

		hc := d.HeadcountInt
		var hcHistory []domain.YearlyHeadcount
		if scbData, ok := headcounts[d.Name]; ok {
			if scbData.HeadcountInt > 0 {
				hc = scbData.HeadcountInt
			}
			hcHistory = make([]domain.YearlyHeadcount, len(scbData.HeadcountHistory))
			for k, h := range scbData.HeadcountHistory {
				hcHistory[k] = domain.YearlyHeadcount{Year: h.Year, HeadcountInt: h.HeadcountInt}
			}
		}

		result[i] = domain.Authority{
			Slug:             toSlug(d.Name),
			Name:             d.Name,
			Role:             d.Role,
			Ministry:         d.Ministry,
			Headcount:        d.Headcount,
			HeadcountInt:     hc,
			Description:      d.Description,
			WebsiteURL:       d.WebsiteURL,
			AnnualReportURL:  d.AnnualReportURL,
			ExpenditureMdkr:  d.ExpenditureMdkr,
			BudgetMdkr:       d.BudgetMdkr,
			Year:             d.Year,
			History:          history,
			HeadcountHistory: hcHistory,
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].ExpenditureMdkr > result[j].ExpenditureMdkr
	})
	return result, nil
}

func (s *Service) GetAuthority(ctx context.Context, slug string) (*domain.AuthorityDetail, error) {
	// Build mandate map from static data on each call (cheap — 11 entries).
	mandateBySlug := s.buildMandateMap(ctx)

	all, err := s.GetAuthorities(ctx)
	if err != nil {
		return nil, err
	}
	for _, a := range all {
		if a.Slug != slug {
			continue
		}

		info := mandateBySlug[slug]
		detail := &domain.AuthorityDetail{
			Authority:  a,
			Mandate:    info.text,
			MandateURL: info.url,
		}

		// Regleringsbrev are generated from the Statsliggaren ID (no DB needed).
		if info.statsliggarenID > 0 && info.activeSince > 0 {
			detail.Regleringsbrev = generateRegleringsbrev(info.statsliggarenID, info.activeSince)
		}

		if s.agencyIntelRepo != nil {
			if dec, err := s.agencyIntelRepo.GetDecisions(ctx, slug); err == nil {
				detail.RecentDecisions = dec
			} else {
				slog.Warn("failed to get decisions", "slug", slug, "error", err)
			}
		}

		return detail, nil
	}
	return nil, ErrNotFound
}

type mandateInfo struct {
	text, url       string
	statsliggarenID int
	activeSince     int
}

// buildMandateMap fetches static authority data and builds slug→mandate lookup.
func (s *Service) buildMandateMap(ctx context.Context) map[string]mandateInfo {
	data, err := s.fallback.FetchAuthorities(ctx)
	if err != nil {
		return nil
	}
	m := make(map[string]mandateInfo, len(data))
	for _, d := range data {
		m[toSlug(d.Name)] = mandateInfo{
			text:            d.Mandate,
			url:             d.MandateURL,
			statsliggarenID: d.StatsliggarenID,
			activeSince:     d.ActiveSince,
		}
	}
	return m
}

const statsliggarenBase = "https://www.statskontoret.se/statsliggaren/regleringsbrev"

// generateRegleringsbrev produces one entry per year from activeSince to the current year,
// each linking to the Statskontoret Statsliggaren for that agency.
func generateRegleringsbrev(id, activeSince int) []domain.Regleringsbrev {
	currentYear := time.Now().Year()
	result := make([]domain.Regleringsbrev, 0, currentYear-activeSince+1)
	for y := currentYear; y >= activeSince; y-- {
		result = append(result, domain.Regleringsbrev{
			Year:  y,
			Title: fmt.Sprintf("Regleringsbrev för budgetåret %d", y),
			URL:   fmt.Sprintf("%s/%d/%d/senaste", statsliggarenBase, id, y),
		})
	}
	return result
}

// fetchHeadcounts fetches SCB headcount data using its own background context so it is
// never cancelled by the HTTP request context (which may already be partially consumed
// by the Statskontoret ZIP download that runs first).
func (s *Service) fetchHeadcounts() map[string]ports.HeadcountData {
	if s.headcount == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	data, err := s.headcount.FetchHeadcounts(ctx)
	if err != nil {
		slog.Warn("scb headcount fetch failed", "error", err)
		return nil
	}
	m := make(map[string]ports.HeadcountData, len(data))
	for _, d := range data {
		m[d.Name] = d
	}
	return m
}

// toSlug produces a stable URL slug from an agency name.
// Rules: lowercase, Swedish chars mapped (å/ä→a, ö→o), spaces→hyphens, non-ASCII stripped.
func toSlug(name string) string {
	replacer := strings.NewReplacer(
		"å", "a", "Å", "a",
		"ä", "a", "Ä", "a",
		"ö", "o", "Ö", "o",
		" ", "-",
	)
	s := strings.ToLower(replacer.Replace(name))
	var b strings.Builder
	for _, r := range s {
		if r <= unicode.MaxASCII && (r == '-' || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

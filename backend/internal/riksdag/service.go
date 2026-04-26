package riksdag

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"sort"
	"strings"
	"unicode"

	"riksdagskollen/internal/riksdag/domain"
	"riksdagskollen/internal/riksdag/ports"
)

var ErrNotFound = errors.New("authority not found")

type Service struct {
	primary   ports.AuthorityClient
	fallback  ports.AuthorityClient
	headcount ports.HeadcountClient // optional; nil means use static data
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

	headcounts := s.fetchHeadcounts(ctx)

	result := make([]domain.Authority, len(data))
	for i, d := range data {
		history := make([]domain.YearlyExpenditure, len(d.History))
		for j, h := range d.History {
			history[j] = domain.YearlyExpenditure{
				Year:            h.Year,
				ExpenditureMdkr: h.ExpenditureMdkr,
			}
		}
		hc := d.HeadcountInt
		if override, ok := headcounts[d.Name]; ok {
			hc = override
		}
		result[i] = domain.Authority{
			Slug:            toSlug(d.Name),
			Name:            d.Name,
			Role:            d.Role,
			Ministry:        d.Ministry,
			Headcount:       d.Headcount,
			HeadcountInt:    hc,
			Description:     d.Description,
			WebsiteURL:      d.WebsiteURL,
			AnnualReportURL: d.AnnualReportURL,
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

func (s *Service) GetAuthority(ctx context.Context, slug string) (*domain.AuthorityDetail, error) {
	all, err := s.GetAuthorities(ctx)
	if err != nil {
		return nil, err
	}
	for _, a := range all {
		if a.Slug == slug {
			return &domain.AuthorityDetail{
				Authority:         a,
				RegleringsbrevURL: regleringsbrevURL(a.Name),
			}, nil
		}
	}
	return nil, ErrNotFound
}

// fetchHeadcounts returns name→headcountInt from the SCB client, or empty map on failure/nil.
func (s *Service) fetchHeadcounts(ctx context.Context) map[string]int {
	if s.headcount == nil {
		return nil
	}
	data, err := s.headcount.FetchHeadcounts(ctx)
	if err != nil {
		slog.Warn("scb headcount fetch failed", "error", err)
		return nil
	}
	m := make(map[string]int, len(data))
	for _, d := range data {
		m[d.Name] = d.HeadcountInt
	}
	return m
}

func regleringsbrevURL(name string) string {
	return fmt.Sprintf(
		"https://data.riksdagen.se/dokumentlista/?doktyp=Rb&titel=%s&utformat=json&sz=200",
		url.QueryEscape(name),
	)
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

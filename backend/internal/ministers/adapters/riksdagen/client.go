package riksdagen

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"riksdagskollen/internal/ministers/domain"
)

const baseURL = "https://data.riksdagen.se"

type Client struct {
	http *http.Client
}

func NewClient() *Client {
	return &Client{http: &http.Client{Timeout: 30 * time.Second}}
}

type rdDoc struct {
	ID          string `json:"id"`
	Titel       string `json:"titel"`
	Organ       string `json:"organ"`
	Rm          string `json:"rm"`
	Datum       string `json:"datum"`
	Dokumenturl string `json:"dokumenturl"`
}

type rdResponse struct {
	Dokumentlista struct {
		Dokument []rdDoc `json:"dokument"`
	} `json:"dokumentlista"`
}

func (c *Client) FetchProposals(ctx context.Context, deptCode string, year string) ([]*domain.Proposal, error) {
	url := fmt.Sprintf("%s/dokumentlista/?doktyp=Prop&organ=%s&rm=%s&sz=50&utformat=json",
		baseURL, deptCode, year)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var rd rdResponse
	if err := json.NewDecoder(resp.Body).Decode(&rd); err != nil {
		return nil, err
	}

	out := make([]*domain.Proposal, 0, len(rd.Dokumentlista.Dokument))
	for _, d := range rd.Dokumentlista.Dokument {
		p := &domain.Proposal{
			ID:             d.ID,
			Title:          strings.TrimSpace(d.Titel),
			DepartmentCode: d.Organ,
			RiksdagYear:    d.Rm,
			URL:            "https://www.riksdagen.se/sv/dokument-och-lagar/dokument/" + d.ID,
		}
		if d.Datum != "" {
			if t, err := time.Parse("2006-01-02", d.Datum); err == nil {
				p.PublishedAt = &t
			}
		}
		out = append(out, p)
	}
	return out, nil
}

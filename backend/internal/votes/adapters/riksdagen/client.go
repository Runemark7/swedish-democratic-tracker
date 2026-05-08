package riksdagen

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"riksdagskollen/internal/votes/domain"
	"riksdagskollen/internal/votes/ports"
)

const baseURL = "https://data.riksdagen.se"

type Client struct {
	http *http.Client
}

func NewClient() *Client {
	return &Client{http: &http.Client{Timeout: 30 * time.Second}}
}

func (c *Client) FetchVotes(ctx context.Context, f ports.FetchVotesFilter) ([]*domain.Vote, error) {
	size := f.Size
	if size == 0 {
		size = 500
	}
	url := fmt.Sprintf("%s/voteringlista/?rm=%s&parti=%s&iid=%s&bet=%s&sz=%d&utformat=json",
		baseURL, f.Session, f.Party, f.PoliticianID, f.Beteckning, size)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("riksdagen votes API returned %d", resp.StatusCode)
	}

	var payload struct {
		Voteringlista struct {
			Votering []struct {
				VoteringID    string `json:"votering_id"`
				IntressentID  string `json:"intressent_id"`
				Namn          string `json:"namn"`
				Parti         string `json:"parti"`
				Rost          string `json:"rost"`
				Beteckning    string `json:"beteckning"`
				Forslagspunkt string `json:"punkt"`
				Rm            string `json:"rm"`
				DokID         string `json:"dok_id"`
				Systemdatum   string `json:"systemdatum"`
			} `json:"votering"`
		} `json:"voteringlista"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode votes: %w", err)
	}

	vv := make([]*domain.Vote, 0, len(payload.Voteringlista.Votering))
	for _, v := range payload.Voteringlista.Votering {
		// Client-side date filter for incremental sync
		if !f.Since.IsZero() && v.Systemdatum != "" {
			if t, err := time.Parse("2006-01-02 15:04:05", v.Systemdatum); err == nil && !t.After(f.Since) {
				continue
			}
		}
		vv = append(vv, &domain.Vote{
			VoteringID:    v.VoteringID,
			PoliticianID:  v.IntressentID,
			Party:         v.Parti,
			VoteResult:    domain.VoteResult(v.Rost),
			Beteckning:    v.Beteckning,
			Forslagspunkt: v.Forslagspunkt,
			Session:       v.Rm,
			DokID:         v.DokID,
		})
	}
	return vv, nil
}

// FetchDocuments returns recent betänkanden from the given committee organs.
// status defaults to "Bifall" — >90% of betänkanden pass; full votering lookup is future work.
// TODO: integrate nämndärenden API (lankadedata.se) for regional/municipal council decisions when available.
func (c *Client) FetchDocuments(ctx context.Context, organs []string, count int) ([]ports.RiksdagDocument, error) {
	url := fmt.Sprintf("%s/dokumentlista/?organ=%s&typ=bet&utformat=json&sz=%d&sort=datum&sortorder=desc",
		baseURL, strings.Join(organs, ","), count)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("riksdagen dokumentlista API returned %d", resp.StatusCode)
	}

	var payload struct {
		Dokumentlista struct {
			Dokument []struct {
				Titel      string `json:"titel"`
				Organ      string `json:"organ"`
				Datum      string `json:"datum"`
				Beteckning string `json:"beteckning"`
			} `json:"dokument"`
		} `json:"dokumentlista"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode dokumentlista: %w", err)
	}

	docs := make([]ports.RiksdagDocument, 0, len(payload.Dokumentlista.Dokument))
	for _, d := range payload.Dokumentlista.Dokument {
		docs = append(docs, ports.RiksdagDocument{
			Title:      d.Titel,
			Organ:      d.Organ,
			Date:       d.Datum,
			Beteckning: d.Beteckning,
		})
	}
	return docs, nil
}

// FetchDocumentStatus fetches /dokumentstatus/{dok_id}.json and parses proposal origin.
func (c *Client) FetchDocumentStatus(ctx context.Context, dokID string) (*domain.DocumentStatus, error) {
	url := fmt.Sprintf("%s/dokumentstatus/%s.json", baseURL, dokID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("riksdagen dokumentstatus API returned %d for %s", resp.StatusCode, dokID)
	}

	var payload struct {
		Dokumentstatus struct {
			Dokument struct {
				DokID      string `json:"dok_id"`
				Titel      string `json:"titel"`
				Typ        string `json:"typ"`
				Datum      string `json:"datum"`
				Undertitel string `json:"undertitel"`
				Summary    string `json:"summary"`
				HTML       string `json:"html"`
			} `json:"dokument"`
			Dokreferens struct {
				Referens []struct {
					RefDokTyp string `json:"ref_dok_typ"`
					RefDokID  string `json:"ref_dok_id"`
				} `json:"referens"`
			} `json:"dokreferens"`
			Dokintressent struct {
				Intressent []struct {
					Partibet string `json:"partibet"`
				} `json:"intressent"`
			} `json:"dokintressent"`
		} `json:"dokumentstatus"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode dokumentstatus %s: %w", dokID, err)
	}

	ds := payload.Dokumentstatus
	status := &domain.DocumentStatus{
		DokID:    ds.Dokument.DokID,
		Title:    ds.Dokument.Titel,
		Type:     ds.Dokument.Typ,
		Date:     ds.Dokument.Datum,
		Subtitle: ds.Dokument.Undertitel,
		Summary:  stripHTML(ds.Dokument.Summary),
		BodyHTML: ds.Dokument.HTML,
	}
	for _, ref := range ds.Dokreferens.Referens {
		r := domain.DocumentReference{
			RefDokTyp: strings.ToLower(ref.RefDokTyp),
			RefDokID:  ref.RefDokID,
		}
		// For motions, fetch the party from the referencing document's intressent
		status.References = append(status.References, r)
	}
	// Primary author party (used when the doc itself is a motion)
	if len(ds.Dokintressent.Intressent) > 0 {
		status.References = append(status.References, domain.DocumentReference{
			RefDokTyp: ds.Dokument.Typ,
			RefDokID:  ds.Dokument.DokID,
			PartyBet:  ds.Dokintressent.Intressent[0].Partibet,
		})
	}
	return status, nil
}

func (c *Client) FetchBetankandeByBeteckning(ctx context.Context, beteckning string) (*ports.BetankandeInfo, error) {
	url := fmt.Sprintf("%s/dokumentlista/?bet=%s&typ=bet&utformat=json&sz=1", baseURL, beteckning)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("riksdagen dokumentlista API returned %d", resp.StatusCode)
	}

	var payload struct {
		Dokumentlista struct {
			Dokument []struct {
				DokID  string `json:"id"`
				Titel  string `json:"titel"`
				Datum  string `json:"datum"`
				Status string `json:"status"`
				Rm     string `json:"rm"`
				Organ  string `json:"organ"`
			} `json:"dokument"`
		} `json:"dokumentlista"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode dokumentlista: %w", err)
	}

	docs := payload.Dokumentlista.Dokument
	if len(docs) == 0 {
		return nil, nil
	}
	d := docs[0]
	return &ports.BetankandeInfo{
		DokID:   d.DokID,
		Title:   d.Titel,
		Date:    d.Datum,
		Status:  d.Status,
		Session: d.Rm,
		Organ:   d.Organ,
	}, nil
}

func stripHTML(s string) string {
	var b strings.Builder
	inTag := false
	for _, r := range s {
		switch {
		case r == '<':
			inTag = true
		case r == '>':
			inTag = false
		case !inTag:
			b.WriteRune(r)
		}
	}
	return strings.TrimSpace(b.String())
}

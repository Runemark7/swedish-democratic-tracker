package riksdagen

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"riksdagskollen/internal/votes/domain"
	"riksdagskollen/internal/votes/ports"
)

const defaultBaseURL = "https://data.riksdagen.se"

type Client struct {
	http    *http.Client
	baseURL string
}

func NewClient() *Client {
	return &Client{
		http:    &http.Client{Timeout: 30 * time.Second},
		baseURL: defaultBaseURL,
	}
}

func (c *Client) FetchVotes(ctx context.Context, f ports.FetchVotesFilter) ([]*domain.Vote, error) {
	size := f.Size
	if size == 0 {
		// The endpoint caps at 10 000 rows and ignores `p`, so ask for the
		// maximum and partition the work by Beteckning. See FetchVotesFilter.
		size = 10000
	}
	url := fmt.Sprintf("%s/voteringlista/?rm=%s&parti=%s&iid=%s&bet=%s&sz=%d&utformat=json",
		c.baseURL, f.Session, f.Party, f.PoliticianID, f.Beteckning, size)

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
		// Riksdagen's own timestamp. Parsed once: it drives both the
		// incremental cutoff below and the ingestion cursor high-water mark.
		var systemDatum time.Time
		if v.Systemdatum != "" {
			if t, err := time.Parse("2006-01-02 15:04:05", v.Systemdatum); err == nil {
				systemDatum = t
			}
		}
		// Client-side date filter for incremental sync
		if !f.Since.IsZero() && !systemDatum.IsZero() && !systemDatum.After(f.Since) {
			continue
		}
		vv = append(vv, &domain.Vote{
			SystemDatum:   systemDatum,
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

// forslagspunktSuffix matches the förslagspunkt Riksdagen concatenates onto the
// beteckning in /dokumentlista for riksmöten 2002/03 through 2013/14 --
// "FIU20p1" rather than the "FiU20" the later riksmöten return. Sampling every
// riksmöte from 2002/03 to 2025/26 on 2026-08-22 put the break at 2014/15, with
// no mixed riksmöte on either side.
//
// The digit before the "p" is required so this cannot bite a beteckning that
// merely ends in a letter and a number.
var forslagspunktSuffix = regexp.MustCompile(`^(.*[0-9])p[0-9]+$`)

// betankandeBeteckning returns the beteckning of the betänkande a votering
// belongs to, which is the unit /voteringlista files ballots under.
//
// Callers partition the ballot fetch by this value, so leaving the older form
// intact asks for bet=FIU20p1 -- a string that matches no ballots at all, while
// bet=FIU20 returns them. That silently emptied every fetch for the three
// mandate periods before 2014.
//
// Stripping loses nothing: each ballot row carries its own förslagspunkt, so
// the punkt is read from the record rather than recovered from this string.
func betankandeBeteckning(s string) string {
	if m := forslagspunktSuffix.FindStringSubmatch(s); m != nil {
		return m[1]
	}
	return s
}

// ListVoteringar enumerates voteringar for a riksmöte, newest first.
//
// This uses /dokumentlista rather than /voteringlista because only the former
// paginates: /voteringlista silently ignores `p` and caps `sz` at 10 000, so it
// cannot enumerate a riksmöte on its own. The @traffar total returned here is
// also the coverage denominator the site publishes.
func (c *Client) ListVoteringar(ctx context.Context, rm string, page, size int) ([]ports.VoteringRef, int, error) {
	if page == 0 {
		page = 1
	}
	if size == 0 {
		size = 200
	}
	url := fmt.Sprintf(
		"%s/dokumentlista/?doktyp=votering&rm=%s&utformat=json&sz=%d&p=%d&sort=datum&sortorder=desc",
		c.baseURL, rm, size, page)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, 0, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, 0, fmt.Errorf("riksdagen dokumentlista returned %d", resp.StatusCode)
	}

	var payload struct {
		Dokumentlista struct {
			Traffar  string `json:"@traffar"`
			Dokument []struct {
				DokID       string `json:"dok_id"`
				Beteckning  string `json:"beteckning"`
				Organ       string `json:"organ"`
				Datum       string `json:"datum"`
				Systemdatum string `json:"systemdatum"`
			} `json:"dokument"`
		} `json:"dokumentlista"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, 0, fmt.Errorf("decode dokumentlista: %w", err)
	}

	total := 0
	_, _ = fmt.Sscanf(payload.Dokumentlista.Traffar, "%d", &total)

	refs := make([]ports.VoteringRef, 0, len(payload.Dokumentlista.Dokument))
	for _, d := range payload.Dokumentlista.Dokument {
		var sd time.Time
		if t, err := time.Parse("2006-01-02 15:04:05", d.Systemdatum); err == nil {
			sd = t
		}
		refs = append(refs, ports.VoteringRef{
			Beteckning:  betankandeBeteckning(d.Beteckning),
			Organ:       d.Organ,
			DokID:       d.DokID,
			Date:        d.Datum,
			SystemDatum: sd,
		})
	}
	return refs, total, nil
}

// FetchDocuments returns recent betänkanden from the given committee organs.
// status defaults to "Bifall" — >90% of betänkanden pass; full votering lookup is future work.
// TODO: integrate nämndärenden API (lankadedata.se) for regional/municipal council decisions when available.
//
// NOTE: Riksdagen ignores `organ` on /dokumentlista. Verified 2026-08-14:
// organ=SoU returns 74 783 hits whose organs include AU, JuU and UbU, and
// omitting the parameter gives the identical result. Callers that need a
// specific committee must filter the returned documents themselves — passing
// organs here selects nothing.
func (c *Client) FetchDocuments(ctx context.Context, organs []string, count int) ([]ports.RiksdagDocument, error) {
	return c.fetchDocumentList(ctx, strings.Join(organs, ","), count)
}

// FetchRecentBetankanden returns the most recently published betänkanden across
// every committee, newest first.
//
// Sends no organ parameter. A feed filtered to committees we chose would
// publish a worldview while claiming to show what happened; the period is the
// selection, and the calendar defines the period.
//
// This does not change what Riksdagen returns — see the note on FetchDocuments.
func (c *Client) FetchRecentBetankanden(ctx context.Context, count int) ([]ports.RiksdagDocument, error) {
	return c.fetchDocumentList(ctx, "", count)
}

// fetchDocumentList fetches /dokumentlista, filtered to betänkanden and sorted
// newest first. organParam is included in the query only when non-empty.
func (c *Client) fetchDocumentList(ctx context.Context, organParam string, count int) ([]ports.RiksdagDocument, error) {
	url := fmt.Sprintf("%s/dokumentlista/?typ=bet&utformat=json&sz=%d&sort=datum&sortorder=desc",
		c.baseURL, count)
	if organParam != "" {
		url += "&organ=" + organParam
	}

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

	// beslutsdag and status are decoded because datum alone cannot tell a
	// decided betänkande from a planned one, and the list carries both. Dropping
	// them is what let planned betänkanden for a coming riksmöte render as the
	// most recent decisions. `beslutad` is deliberately not decoded: it arrives
	// with an inconsistent JSON type, and an empty beslutsdag answers the same
	// question without the guesswork.
	var payload struct {
		Dokumentlista struct {
			Dokument []struct {
				Titel      string `json:"titel"`
				Organ      string `json:"organ"`
				Datum      string `json:"datum"`
				Beteckning string `json:"beteckning"`
				Beslutsdag string `json:"beslutsdag"`
				Status     string `json:"status"`
			} `json:"dokument"`
		} `json:"dokumentlista"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode dokumentlista: %w", err)
	}

	docs := make([]ports.RiksdagDocument, 0, len(payload.Dokumentlista.Dokument))
	for _, d := range payload.Dokumentlista.Dokument {
		docs = append(docs, ports.RiksdagDocument{
			Title:        d.Titel,
			Organ:        d.Organ,
			Date:         d.Datum,
			Beteckning:   d.Beteckning,
			DecisionDate: d.Beslutsdag,
			Status:       d.Status,
		})
	}
	return docs, nil
}

// FetchDocumentStatus fetches /dokumentstatus/{dok_id}.json and parses proposal origin.
func (c *Client) FetchDocumentStatus(ctx context.Context, dokID string) (*domain.DocumentStatus, error) {
	url := fmt.Sprintf("%s/dokumentstatus/%s.json", c.baseURL, dokID)

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
				Beteckning string `json:"beteckning"`
			} `json:"dokument"`
			Dokreferens struct {
				Referens []struct {
					RefDokTyp string `json:"ref_dok_typ"`
					RefDokID  string `json:"ref_dok_id"`
				} `json:"referens"`
			} `json:"dokreferens"`
			Dokintressent struct {
				Intressent []struct {
					IntressentID string `json:"intressent_id"`
					Namn         string `json:"namn"`
					Partibet     string `json:"partibet"`
					Roll         string `json:"roll"`
				} `json:"intressent"`
			} `json:"dokintressent"`
			Dokuppgift struct {
				Uppgift []struct {
					Kod  string `json:"kod"`
					Text string `json:"text"`
				} `json:"uppgift"`
			} `json:"dokuppgift"`
		} `json:"dokumentstatus"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode dokumentstatus %s: %w", dokID, err)
	}

	ds := payload.Dokumentstatus
	status := &domain.DocumentStatus{
		DokID:      ds.Dokument.DokID,
		Title:      ds.Dokument.Titel,
		Type:       ds.Dokument.Typ,
		Date:       ds.Dokument.Datum,
		Subtitle:   ds.Dokument.Undertitel,
		Summary:    stripHTML(ds.Dokument.Summary),
		Beteckning: ds.Dokument.Beteckning,
		BodyHTML:   ds.Dokument.HTML,
	}
	for _, ref := range ds.Dokreferens.Referens {
		r := domain.DocumentReference{
			RefDokTyp: strings.ToLower(ref.RefDokTyp),
			RefDokID:  ref.RefDokID,
		}
		// For motions, fetch the party from the referencing document's intressent
		status.References = append(status.References, r)
	}
	for _, i := range ds.Dokintressent.Intressent {
		status.Intressenter = append(status.Intressenter, domain.Intressent{
			IntressentID: i.IntressentID,
			Name:         i.Namn,
			Party:        i.Partibet,
			Role:         i.Roll,
		})
	}
	for _, u := range ds.Dokuppgift.Uppgift {
		switch u.Kod {
		case "debattdatumtid":
			status.DebattDate = u.Text
		case "beslutdatumtid":
			status.BeslutDate = u.Text
		case "statustext":
			status.StatusText = u.Text
		case "notis":
			status.Notis = u.Text
		}
	}
	// Preserve existing behaviour: first intressent's party feeds proposal-origin tracing via References.
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
	url := fmt.Sprintf("%s/dokumentlista/?bet=%s&typ=bet&utformat=json&sz=1", c.baseURL, beteckning)

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

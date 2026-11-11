// Package riksdagen reads motions from data.riksdagen.se.
package riksdagen

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"riksdagskollen/internal/motions/domain"
	"riksdagskollen/internal/motions/ports"
)

const defaultBaseURL = "https://data.riksdagen.se"

type Client struct {
	http    *http.Client
	baseURL string
}

// NewClientWithBaseURL points the client at baseURL; empty means production.
func NewClientWithBaseURL(baseURL string) *Client {
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	return &Client{http: &http.Client{Timeout: 30 * time.Second}, baseURL: strings.TrimRight(baseURL, "/")}
}

func (c *Client) getJSON(ctx context.Context, u string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("riksdagen returned %d for %s", resp.StatusCode, req.URL.Path)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

// ListMotions pages /dokumentlista for doktyp=mot. The riksmöte is sent in its
// slash form; the client encodes it.
func (c *Client) ListMotions(ctx context.Context, rm string, page, size int) ([]ports.MotionRef, int, error) {
	if page == 0 {
		page = 1
	}
	if size == 0 {
		size = 200
	}
	q := url.Values{}
	q.Set("doktyp", "mot")
	q.Set("rm", rm)
	q.Set("utformat", "json")
	q.Set("sz", strconv.Itoa(size))
	q.Set("p", strconv.Itoa(page))
	q.Set("sort", "datum")
	q.Set("sortorder", "desc")

	var payload struct {
		Dokumentlista struct {
			Traffar  string `json:"@traffar"`
			Dokument flexList[struct {
				DokID       string `json:"dok_id"`
				Systemdatum string `json:"systemdatum"`
			}] `json:"dokument"`
		} `json:"dokumentlista"`
	}
	if err := c.getJSON(ctx, c.baseURL+"/dokumentlista/?"+q.Encode(), &payload); err != nil {
		return nil, 0, fmt.Errorf("list motions: %w", err)
	}
	total, _ := strconv.Atoi(payload.Dokumentlista.Traffar)
	refs := make([]ports.MotionRef, 0, len(payload.Dokumentlista.Dokument))
	for _, d := range payload.Dokumentlista.Dokument {
		refs = append(refs, ports.MotionRef{DokID: d.DokID, SystemDatum: parseTime(d.Systemdatum)})
	}
	return refs, total, nil
}

// FetchMotion reads /dokumentstatus/{dok_id}.json.
func (c *Client) FetchMotion(ctx context.Context, dokID string) (*domain.Motion, error) {
	var payload struct {
		Status struct {
			Dokument struct {
				DokID       string `json:"dok_id"`
				Rm          string `json:"rm"`
				Beteckning  string `json:"beteckning"`
				Subtyp      string `json:"subtyp"`
				Organ       string `json:"organ"`
				Datum       string `json:"datum"`
				Systemdatum string `json:"systemdatum"`
				Titel       string `json:"titel"`
				Status      string `json:"status"`
			} `json:"dokument"`
			Forslag struct {
				Forslag flexList[forslag] `json:"forslag"`
			} `json:"dokforslag"`
			Intressenter struct {
				Intressent flexList[intressent] `json:"intressent"`
			} `json:"dokintressent"`
			Uppgifter struct {
				Uppgift flexList[uppgift] `json:"uppgift"`
			} `json:"dokuppgift"`
			Referenser struct {
				Referens flexList[referens] `json:"referens"`
			} `json:"dokreferens"`
		} `json:"dokumentstatus"`
	}
	if err := c.getJSON(ctx, c.baseURL+"/dokumentstatus/"+url.PathEscape(dokID)+".json", &payload); err != nil {
		return nil, fmt.Errorf("fetch motion %s: %w", dokID, err)
	}
	d := payload.Status.Dokument
	if d.DokID == "" || d.Rm == "" || d.Beteckning == "" {
		return nil, fmt.Errorf("fetch motion %s: response has no dok_id, rm or beteckning", dokID)
	}

	m := &domain.Motion{
		DokID: d.DokID, Riksmote: d.Rm, Beteckning: d.Beteckning, Title: d.Titel,
		Subtyp: d.Subtyp, Organ: d.Organ, Status: d.Status,
		SystemDatum: parseTime(d.Systemdatum),
	}
	if t := parseTime(d.Datum); !t.IsZero() {
		day := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
		m.Date = &day
	}
	for _, u := range payload.Status.Uppgifter.Uppgift {
		if u.Kod == "tilldelat" {
			m.Tilldelat = u.Text
		}
	}
	for _, r := range payload.Status.Referenser.Referens {
		if r.Referenstyp == "behandlas_i" {
			m.BehandlasI = r.Uppgift
			break
		}
	}
	for i, a := range payload.Status.Intressenter.Intressent {
		order, err := strconv.Atoi(a.Ordning)
		if err != nil {
			order = i + 1
		}
		m.Authors = append(m.Authors, domain.Author{
			Order: order, PoliticianID: a.IntressentID, Name: a.Namn, Party: a.Partibet, Role: a.Roll,
		})
	}
	for i, f := range payload.Status.Forslag.Forslag {
		n, err := strconv.Atoi(f.Nummer)
		if err != nil {
			n = i + 1
		}
		m.Yrkanden = append(m.Yrkanden, domain.Yrkande{
			Number: n, Wording: f.Lydelse, UtskottProposal: f.Utskottet, ChamberDecision: f.Kammaren,
			BehandlasI: f.BehandlasI, BehandlasIPunkt: f.BehandlasIPunkt,
		})
	}
	return m, nil
}

type forslag struct {
	Nummer          string `json:"nummer"`
	Lydelse         string `json:"lydelse"`
	Utskottet       string `json:"utskottet"`
	Kammaren        string `json:"kammaren"`
	BehandlasI      string `json:"behandlas_i"`
	BehandlasIPunkt string `json:"behandlas_i_punkt"`
}

type intressent struct {
	IntressentID string `json:"intressent_id"`
	Namn         string `json:"namn"`
	Partibet     string `json:"partibet"`
	Ordning      string `json:"ordning"`
	Roll         string `json:"roll"`
}

type uppgift struct {
	Kod  string `json:"kod"`
	Text string `json:"text"`
}

type referens struct {
	Referenstyp string `json:"referenstyp"`
	Uppgift     string `json:"uppgift"`
}

// flexList decodes what Riksdagen serves as an object when there is one item,
// an array when there are several, and null or "" when there are none.
type flexList[T any] []T

func (l *flexList[T]) UnmarshalJSON(b []byte) error {
	s := strings.TrimSpace(string(b))
	switch {
	case s == "" || s == "null" || s == `""`:
		*l = nil
		return nil
	case s[0] == '[':
		var arr []T
		if err := json.Unmarshal(b, &arr); err != nil {
			return err
		}
		*l = arr
		return nil
	case s[0] == '{':
		var one T
		if err := json.Unmarshal(b, &one); err != nil {
			return err
		}
		*l = []T{one}
		return nil
	}
	return fmt.Errorf("unexpected JSON %.20q", s)
}

func parseTime(s string) time.Time {
	for _, layout := range []string{"2006-01-02 15:04:05", "2006-01-02"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t
		}
	}
	return time.Time{}
}

// FindBetankande lists /dokumentlista for doktyp=bet and picks the exact
// riksmöte and beteckning.
func (c *Client) FindBetankande(ctx context.Context, rm, bet string) (ports.BetankandeRef, bool, error) {
	q := url.Values{}
	q.Set("doktyp", "bet")
	q.Set("rm", rm)
	q.Set("bet", bet)
	q.Set("sz", "5")
	q.Set("utformat", "json")
	var payload struct {
		Dokumentlista struct {
			Dokument flexList[struct {
				DokID       string          `json:"dok_id"`
				Rm          string          `json:"rm"`
				Beteckning  string          `json:"beteckning"`
				Systemdatum string          `json:"systemdatum"`
				Titel       string          `json:"titel"`
				Status      string          `json:"status"`
				Beslutsdag  string          `json:"beslutsdag"`
				Beslutad    json.RawMessage `json:"beslutad"`
			}] `json:"dokument"`
		} `json:"dokumentlista"`
	}
	if err := c.getJSON(ctx, c.baseURL+"/dokumentlista/?"+q.Encode(), &payload); err != nil {
		return ports.BetankandeRef{}, false, fmt.Errorf("find betänkande %s:%s: %w", rm, bet, err)
	}
	for _, d := range payload.Dokumentlista.Dokument {
		if d.Rm == rm && strings.EqualFold(d.Beteckning, bet) && d.DokID != "" {
			return ports.BetankandeRef{
				DokID: d.DokID, SystemDatum: parseTime(d.Systemdatum), Title: d.Titel, Status: d.Status,
				DecisionDate: d.Beslutsdag, ChamberDecided: parseBeslutad(d.Beslutad),
			}, true, nil
		}
	}
	return ports.BetankandeRef{}, false, nil
}

// ListBetankanden lists /dokumentlista for doktyp=bet in one riksmöte, newest
// first, one page at a time.
func (c *Client) ListBetankanden(ctx context.Context, rm string, page, size int) ([]ports.BetankandeRef, int, error) {
	q := url.Values{}
	q.Set("doktyp", "bet")
	q.Set("rm", rm)
	q.Set("sz", strconv.Itoa(size))
	q.Set("p", strconv.Itoa(page))
	q.Set("sort", "datum")
	q.Set("sortorder", "desc")
	q.Set("utformat", "json")
	var payload struct {
		Dokumentlista struct {
			Traffar  string `json:"@traffar"`
			Dokument flexList[struct {
				DokID       string          `json:"dok_id"`
				Rm          string          `json:"rm"`
				Beteckning  string          `json:"beteckning"`
				Organ       string          `json:"organ"`
				Datum       string          `json:"datum"`
				Systemdatum string          `json:"systemdatum"`
				Titel       string          `json:"titel"`
				Status      string          `json:"status"`
				Beslutsdag  string          `json:"beslutsdag"`
				Beslutad    json.RawMessage `json:"beslutad"`
			}] `json:"dokument"`
		} `json:"dokumentlista"`
	}
	if err := c.getJSON(ctx, c.baseURL+"/dokumentlista/?"+q.Encode(), &payload); err != nil {
		return nil, 0, fmt.Errorf("list betänkanden %s page %d: %w", rm, page, err)
	}
	total, _ := strconv.Atoi(payload.Dokumentlista.Traffar)
	var out []ports.BetankandeRef
	for _, d := range payload.Dokumentlista.Dokument {
		if d.DokID == "" || d.Beteckning == "" || d.Rm != rm {
			continue
		}
		out = append(out, ports.BetankandeRef{
			DokID: d.DokID, Riksmote: d.Rm, Beteckning: d.Beteckning, Organ: d.Organ,
			Published: parseTime(d.Datum), SystemDatum: parseTime(d.Systemdatum), Title: d.Titel, Status: d.Status,
			DecisionDate: d.Beslutsdag, ChamberDecided: parseBeslutad(d.Beslutad),
		})
	}
	return out, total, nil
}

// FetchBetankande reads /dokumentstatus/{dok_id}.json and its
// dokutskottsforslag.
func (c *Client) FetchBetankande(ctx context.Context, dokID string) (*domain.Betankande, error) {
	var payload struct {
		Status struct {
			Dokument struct {
				DokID       string `json:"dok_id"`
				Rm          string `json:"rm"`
				Beteckning  string `json:"beteckning"`
				Organ       string `json:"organ"`
				Systemdatum string `json:"systemdatum"`
				Titel       string `json:"titel"`
			} `json:"dokument"`
			Forslag struct {
				Forslag flexList[struct {
					Punkt      string          `json:"punkt"`
					Rubrik     string          `json:"rubrik"`
					Forslag    string          `json:"forslag"`
					Beslutstyp string          `json:"beslutstyp"`
					VoteringID string          `json:"votering_id"`
					Summary    json.RawMessage `json:"votering_sammanfattning_html"`
					Vinnare    string          `json:"vinnare"`
				}] `json:"utskottsforslag"`
			} `json:"dokutskottsforslag"`
		} `json:"dokumentstatus"`
	}
	if err := c.getJSON(ctx, c.baseURL+"/dokumentstatus/"+url.PathEscape(dokID)+".json", &payload); err != nil {
		return nil, fmt.Errorf("fetch betänkande %s: %w", dokID, err)
	}
	d := payload.Status.Dokument
	if d.DokID == "" || d.Rm == "" || d.Beteckning == "" {
		return nil, fmt.Errorf("fetch betänkande %s: response has no dok_id, rm or beteckning", dokID)
	}
	b := &domain.Betankande{
		DokID: d.DokID, Riksmote: d.Rm, Beteckning: d.Beteckning, Title: d.Titel,
		Organ: d.Organ, SystemDatum: parseTime(d.Systemdatum),
	}
	for _, f := range payload.Status.Forslag.Forslag {
		if f.Punkt == "" {
			continue
		}
		b.Punkter = append(b.Punkter, domain.Forslagspunkt{
			Number: f.Punkt, Rubrik: f.Rubrik, Text: f.Forslag, Beslutstyp: f.Beslutstyp,
			VoteringID: f.VoteringID, Alternativ: votedAlternative(f.Summary), Vinnare: f.Vinnare,
		})
	}
	return b, nil
}

var alternativeRe = regexp.MustCompile(`(?i)\bmot\b.*(förslag|reservation)|förslag.*\bmot\b`)

// votedAlternative finds the line naming the alternative put to the vote
// ("Utskottets förslag mot reservation 12 (M)") in the votering summary,
// which Riksdagen serves as differently shaped HTML-to-JSON objects. It
// returns "" when there is none.
func votedAlternative(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return ""
	}
	var found string
	var walk func(any)
	walk = func(n any) {
		if found != "" {
			return
		}
		switch t := n.(type) {
		case string:
			s := strings.TrimSpace(t)
			if len(s) < 200 && alternativeRe.MatchString(s) {
				found = s
			}
		case []any:
			for _, e := range t {
				walk(e)
			}
		case map[string]any:
			keys := make([]string, 0, len(t))
			for k := range t {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				walk(t[k])
			}
		}
	}
	walk(v)
	return found
}

// parseBeslutad reads Riksdagen's `beslutad` flag, which arrives as the
// strings "0"/"1", a number or a bool. nil means absent or unreadable.
func parseBeslutad(raw json.RawMessage) *bool {
	t := strings.Trim(strings.TrimSpace(string(raw)), `"`)
	switch strings.ToLower(t) {
	case "1", "true":
		v := true
		return &v
	case "0", "false":
		v := false
		return &v
	}
	return nil
}

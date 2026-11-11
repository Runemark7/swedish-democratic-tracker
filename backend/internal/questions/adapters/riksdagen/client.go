// Package riksdagen reads skriftliga frågor, svar and interpellationer from
// data.riksdagen.se.
package riksdagen

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"riksdagskollen/internal/questions/domain"
	"riksdagskollen/internal/questions/ports"
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

func doktyp(kind string) (string, error) {
	switch kind {
	case domain.KindFraga:
		return "fr", nil
	case domain.KindInterpellation:
		return "ip", nil
	}
	return "", fmt.Errorf("unknown question kind %q", kind)
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

// List pages /dokumentlista for doktyp=fr or ip.
func (c *Client) List(ctx context.Context, kind, rm string, page, size int) ([]ports.Ref, int, error) {
	typ, err := doktyp(kind)
	if err != nil {
		return nil, 0, err
	}
	if page == 0 {
		page = 1
	}
	if size == 0 {
		size = 200
	}
	q := url.Values{}
	q.Set("doktyp", typ)
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
		return nil, 0, fmt.Errorf("list %s: %w", typ, err)
	}
	total, _ := strconv.Atoi(payload.Dokumentlista.Traffar)
	refs := make([]ports.Ref, 0, len(payload.Dokumentlista.Dokument))
	for _, d := range payload.Dokumentlista.Dokument {
		refs = append(refs, ports.Ref{DokID: d.DokID, SystemDatum: parseTime(d.Systemdatum)})
	}
	return refs, total, nil
}

type status struct {
	Dokument struct {
		DokID       string `json:"dok_id"`
		Rm          string `json:"rm"`
		Beteckning  string `json:"beteckning"`
		Datum       string `json:"datum"`
		Systemdatum string `json:"systemdatum"`
		Titel       string `json:"titel"`
		Status      string `json:"status"`
		HTML        string `json:"html"`
	} `json:"dokument"`
	Aktivitet struct {
		Aktivitet flexList[aktivitet] `json:"aktivitet"`
	} `json:"dokaktivitet"`
	Intressenter struct {
		Intressent flexList[intressent] `json:"intressent"`
	} `json:"dokintressent"`
	Referenser struct {
		Referens flexList[referens] `json:"referens"`
	} `json:"dokreferens"`
}

type aktivitet struct {
	Kod   string `json:"kod"`
	Namn  string `json:"namn"`
	Datum string `json:"datum"`
}

type intressent struct {
	IntressentID string `json:"intressent_id"`
	Namn         string `json:"namn"`
	Partibet     string `json:"partibet"`
	Roll         string `json:"roll"`
}

type referens struct {
	Referenstyp string `json:"referenstyp"`
	RefDokID    string `json:"ref_dok_id"`
	RefDokTyp   string `json:"ref_dok_typ"`
	RefDokRm    string `json:"ref_dok_rm"`
	RefDokBet   string `json:"ref_dok_bet"`
}

func (c *Client) status(ctx context.Context, dokID string) (*status, error) {
	var payload struct {
		Status status `json:"dokumentstatus"`
	}
	if err := c.getJSON(ctx, c.baseURL+"/dokumentstatus/"+url.PathEscape(dokID)+".json", &payload); err != nil {
		return nil, fmt.Errorf("fetch %s: %w", dokID, err)
	}
	d := payload.Status.Dokument
	if d.DokID == "" || d.Rm == "" || d.Beteckning == "" {
		return nil, fmt.Errorf("fetch %s: response has no dok_id, rm or beteckning", dokID)
	}
	return &payload.Status, nil
}

func person(s *status, roles ...string) domain.Person {
	for _, role := range roles {
		for _, p := range s.Intressenter.Intressent {
			if p.Roll == role {
				return domain.Person{ID: p.IntressentID, Name: p.Namn, Party: p.Partibet}
			}
		}
	}
	return domain.Person{}
}

func day(s string) *time.Time {
	t := parseTime(s)
	if t.IsZero() {
		return nil
	}
	d := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
	return &d
}

// Fetch reads a fråga or interpellation. For a fråga it follows the
// `frågesvar` reference to the svar document and takes the answer date and the
// answering minister from it.
func (c *Client) Fetch(ctx context.Context, kind, dokID string) (*domain.Question, error) {
	if _, err := doktyp(kind); err != nil {
		return nil, err
	}
	s, err := c.status(ctx, dokID)
	if err != nil {
		return nil, err
	}
	d := s.Dokument
	q := &domain.Question{
		DokID: d.DokID, Kind: kind, Riksmote: d.Rm, Beteckning: d.Beteckning,
		Title: d.Titel, AskedDate: day(d.Datum), Status: d.Status,
		SystemDatum: parseTime(d.Systemdatum),
		Asker:       person(s, "undertecknare", "fragestallare"),
		Addressed:   person(s, "stalldtill"),
		Answerer:    person(s, "besvaradav"),
		BodyText:    plainText(d.HTML),
	}
	for _, r := range s.Referenser.Referens {
		switch {
		case r.RefDokTyp == "mot" && r.RefDokRm != "" && r.RefDokBet != "":
			q.MotionRefs = append(q.MotionRefs, domain.Ref(r.RefDokRm, r.RefDokBet))
		case kind == domain.KindFraga && r.Referenstyp == "frågesvar" && r.RefDokTyp == "frs":
			q.AnswerDokID = r.RefDokID
		}
	}
	switch kind {
	case domain.KindFraga:
		if q.AnswerDokID != "" {
			a, err := c.status(ctx, q.AnswerDokID)
			if err != nil {
				return nil, err
			}
			q.AnswerDate = day(a.Dokument.Datum)
			if p := person(a, "besvaradav"); p.ID != "" {
				q.Answerer = p
			}
			if p := person(a, "fragestallare"); p.ID != "" && q.Asker.ID == "" {
				q.Asker = p
			}
		}
	case domain.KindInterpellation:
		// The interpellation is answered in the chamber on the day of its
		// "Besvarad" activity.
		for _, a := range s.Aktivitet.Aktivitet {
			if a.Namn == "Besvarad" {
				q.AnswerDate = day(a.Datum)
			}
		}
	}
	return q, nil
}

var (
	styleRe = regexp.MustCompile(`(?is)<style.*?</style>`)
	tagRe   = regexp.MustCompile(`(?s)<[^>]+>`)
	spaceRe = regexp.MustCompile(`[ \t\x{00a0}]+`)
	nlRe    = regexp.MustCompile(`\s*\n\s*`)
)

// plainText turns the document's HTML into text: no style block, no tags,
// entities decoded, whitespace collapsed.
func plainText(h string) string {
	h = styleRe.ReplaceAllString(h, "")
	h = strings.NewReplacer("</p>", "\n", "<br>", "\n", "<br/>", "\n", "</h2>", "\n").Replace(h)
	h = tagRe.ReplaceAllString(h, " ")
	h = html.UnescapeString(h)
	h = spaceRe.ReplaceAllString(h, " ")
	h = nlRe.ReplaceAllString(h, "\n")
	return strings.TrimSpace(h)
}

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

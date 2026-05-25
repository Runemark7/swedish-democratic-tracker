package registret

import (
	"context"
	"fmt"
	"html"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
	"unicode"

	"riksdagskollen/internal/riksdag/ports"
)

// group is a Myndighetsregistret category. The group label is sent verbatim as
// the HamtaMynd "mynd" parameter; typ/principal/underGov classify the rows.
type group struct {
	label     string
	typ       string
	principal string
	underGov  bool
}

// groups enumerates the six categories from the #MyId selector. Iterating all
// of them yields the full register (~449).
var groups = []group{
	{"Statliga förvaltningsmyndigheter", "Förvaltningsmyndighet", "Regeringen", true},
	{"Myndigheter under riksdagen", "Riksdagsmyndighet", "Riksdagen", false},
	{"Statliga affärsverk", "Affärsverk", "Regeringen", true},
	{"AP-fonder", "AP-fond", "Regeringen", true},
	{"Sveriges domstolar samt Domstolsverket", "Domstol", "Riksdagen", false},
	{"Svenska utlandsmyndigheter", "Utlandsmyndighet", "Regeringen", true},
}

var (
	rowRe   = regexp.MustCompile(`(?s)<tr[^>]*>(.*?)</tr>`)
	cellRe  = regexp.MustCompile(`(?s)<td[^>]*>(.*?)</td>`)
	tagRe   = regexp.MustCompile(`<[^>]+>`)
	orgNrRe = regexp.MustCompile(`^\d{6}-\d{4}$`)
)

func cleanCell(s string) string {
	s = tagRe.ReplaceAllString(s, "")
	return strings.TrimSpace(html.UnescapeString(s))
}

// slugify mirrors riksdag.toSlug so detail-page slugs stay consistent.
func slugify(name string) string {
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

const hamtaURL = "https://myndighetsregistret.scb.se/Myndighet/HamtaMynd"

type Client struct {
	http *http.Client
}

func NewClient() *Client {
	return &Client{http: &http.Client{Timeout: 30 * time.Second}}
}

func (c *Client) FetchRegister(ctx context.Context) ([]ports.RegisterEntry, error) {
	byOrg := make(map[string]ports.RegisterEntry)
	var firstErr error

	for i, g := range groups {
		entries, err := c.fetchGroup(ctx, g)
		if err != nil {
			// Best-effort: remember the first error, keep going so one bad group
			// does not lose the others.
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		for _, e := range entries {
			byOrg[e.OrgNumber] = e // dedupe across groups by org number
		}
		if i < len(groups)-1 {
			time.Sleep(400 * time.Millisecond)
		}
	}

	if len(byOrg) == 0 {
		if firstErr != nil {
			return nil, fmt.Errorf("register: all groups failed: %w", firstErr)
		}
		return nil, fmt.Errorf("register: no entries parsed")
	}

	out := make([]ports.RegisterEntry, 0, len(byOrg))
	for _, e := range byOrg {
		out = append(out, e)
	}
	return out, nil
}

func (c *Client) fetchGroup(ctx context.Context, g group) ([]ports.RegisterEntry, error) {
	body := fmt.Sprintf(`{"mynd":%q}`, g.label)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, hamtaURL, strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("User-Agent", "riksdagskollen/1.0")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("hamtamynd %q: %w", g.label, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("hamtamynd %q: HTTP %d", g.label, resp.StatusCode)
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return parseAgencyTable(string(raw), g), nil
}

func parseAgencyTable(fragment string, g group) []ports.RegisterEntry {
	var out []ports.RegisterEntry
	for _, rm := range rowRe.FindAllStringSubmatch(fragment, -1) {
		cells := cellRe.FindAllStringSubmatch(rm[1], -1)
		if len(cells) < 2 {
			continue
		}
		name := cleanCell(cells[0][1])
		org := cleanCell(cells[1][1])
		if name == "" || !orgNrRe.MatchString(org) {
			continue
		}
		var sfs, web string
		if len(cells) > 2 {
			sfs = cleanCell(cells[2][1])
		}
		if len(cells) > 3 {
			web = cleanCell(cells[3][1])
		}
		out = append(out, ports.RegisterEntry{
			OrgNumber:       org,
			Slug:            slugify(name),
			Name:            name,
			Type:            g.typ,
			PrincipalBody:   g.principal,
			UnderGovernment: g.underGov,
			Website:         web,
			SFS:             sfs,
		})
	}
	return out
}

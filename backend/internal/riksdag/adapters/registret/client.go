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

// schema identifies the cell layout returned by HamtaMynd for a given group.
// SCB returns three different table shapes — see group definitions below.
type schema int

const (
	schemaStandard schema = iota // Namn, Org-nr, SFS, WebbAdress
	schemaCourt                  // Namn, CfarNr (8-digit), Postadress, Postnr, Postort, WebbAdress
	schemaForeign                // Land, LopNr, Namn, Ambassadör/Generalkonsul, WebbAdress
)

// group is a Myndighetsregistret category. The label is sent verbatim as the
// HamtaMynd "mynd" parameter; typ/principal/underGov classify the rows; schema
// picks the row parser.
type group struct {
	label     string
	typ       string
	principal string
	underGov  bool
	schema    schema
}

// groups enumerates the six categories from the #MyId selector. Iterating all
// of them yields the full register (~449).
var groups = []group{
	{"Statliga förvaltningsmyndigheter", "Förvaltningsmyndighet", "Regeringen", true, schemaStandard},
	{"Myndigheter under riksdagen", "Riksdagsmyndighet", "Riksdagen", false, schemaStandard},
	{"Statliga affärsverk", "Affärsverk", "Regeringen", true, schemaStandard},
	{"AP-fonder", "AP-fond", "Regeringen", true, schemaStandard},
	{"Sveriges domstolar samt Domstolsverket", "Domstol", "Riksdagen", false, schemaCourt},
	{"Svenska utlandsmyndigheter", "Utlandsmyndighet", "Regeringen", true, schemaForeign},
}

var (
	rowRe    = regexp.MustCompile(`(?s)<tr[^>]*>(.*?)</tr>`)
	cellRe   = regexp.MustCompile(`(?s)<td[^>]*>(.*?)</td>`)
	tagRe    = regexp.MustCompile(`<[^>]+>`)
	orgNrRe  = regexp.MustCompile(`^\d{6}-\d{4}$`)
	cfarRe   = regexp.MustCompile(`^\d{6,10}$`)
	lopnrRe  = regexp.MustCompile(`^\d+$`)
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

// parseAgencyTable dispatches to the right row parser for the group's schema.
func parseAgencyTable(fragment string, g group) []ports.RegisterEntry {
	var out []ports.RegisterEntry
	for _, rm := range rowRe.FindAllStringSubmatch(fragment, -1) {
		cells := cellRe.FindAllStringSubmatch(rm[1], -1)
		if len(cells) < 2 {
			continue
		}
		var (
			e  ports.RegisterEntry
			ok bool
		)
		switch g.schema {
		case schemaStandard:
			e, ok = parseStandardRow(cells, g)
		case schemaCourt:
			e, ok = parseCourtRow(cells, g)
		case schemaForeign:
			e, ok = parseForeignRow(cells, g)
		}
		if ok {
			out = append(out, e)
		}
	}
	return out
}

// schemaStandard: cells = [Namn, Org-nr, SFS, WebbAdress].
func parseStandardRow(cells [][]string, g group) (ports.RegisterEntry, bool) {
	name := cleanCell(cells[0][1])
	org := cleanCell(cells[1][1])
	if name == "" || !orgNrRe.MatchString(org) {
		return ports.RegisterEntry{}, false
	}
	var sfs, web string
	if len(cells) > 2 {
		sfs = cleanCell(cells[2][1])
	}
	if len(cells) > 3 {
		web = cleanCell(cells[3][1])
	}
	return ports.RegisterEntry{
		OrgNumber:       org,
		Slug:            slugify(name),
		Name:            name,
		Type:            g.typ,
		PrincipalBody:   g.principal,
		UnderGovernment: g.underGov,
		Website:         web,
		SFS:             sfs,
	}, true
}

// schemaCourt: cells = [Namn, CfarNr, Postadress, Postnr, Postort, WebbAdress].
// Courts share Domstolsverket's organisationsnummer, so synthesize a stable
// per-court key from the CfarNr (the workplace identifier in SCB's register).
func parseCourtRow(cells [][]string, g group) (ports.RegisterEntry, bool) {
	if len(cells) < 2 {
		return ports.RegisterEntry{}, false
	}
	name := cleanCell(cells[0][1])
	cfar := cleanCell(cells[1][1])
	if name == "" || !cfarRe.MatchString(cfar) {
		return ports.RegisterEntry{}, false
	}
	var web string
	if len(cells) > 5 {
		web = cleanCell(cells[5][1])
	}
	return ports.RegisterEntry{
		OrgNumber:       "cfar:" + cfar,
		Slug:            slugify(name),
		Name:            name,
		Type:            g.typ,
		PrincipalBody:   g.principal,
		UnderGovernment: g.underGov,
		Website:         web,
	}, true
}

// schemaForeign: cells = [Land, LopNr, Namn, Ambassadör/Generalkonsul, WebbAdress].
// Embassies share Utrikesdepartementet's organisationsnummer, so synthesize a
// stable key from LopNr (the register's sequence number for the mission).
func parseForeignRow(cells [][]string, g group) (ports.RegisterEntry, bool) {
	if len(cells) < 3 {
		return ports.RegisterEntry{}, false
	}
	lopnr := cleanCell(cells[1][1])
	name := cleanCell(cells[2][1])
	if name == "" || !lopnrRe.MatchString(lopnr) {
		return ports.RegisterEntry{}, false
	}
	var web string
	if len(cells) > 4 {
		web = cleanCell(cells[4][1])
	}
	return ports.RegisterEntry{
		OrgNumber:       "utland:" + lopnr,
		Slug:            slugify(name),
		Name:            name,
		Type:            g.typ,
		PrincipalBody:   g.principal,
		UnderGovernment: g.underGov,
		Website:         web,
	}, true
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

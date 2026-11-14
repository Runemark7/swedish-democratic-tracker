package riksdagen

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/net/html"

	"riksdagskollen/internal/motions/domain"
	"riksdagskollen/internal/motions/ports"
)

// block is one paragraph or heading of the betänkande, in document order.
type block struct {
	tag, class, text string
}

func (b block) isSection() bool { return b.class == "Avsnittsrubrik" }

func collectBlocks(n *html.Node, out *[]block) {
	if n.Type == html.ElementNode && (n.Data == "p" || (len(n.Data) == 2 && n.Data[0] == 'h' && n.Data[1] >= '1' && n.Data[1] <= '6')) {
		var class string
		for _, a := range n.Attr {
			if a.Key == "class" {
				class = a.Val
			}
		}
		var sb strings.Builder
		var text func(*html.Node)
		text = func(c *html.Node) {
			if c.Type == html.TextNode {
				sb.WriteString(c.Data)
			}
			for k := c.FirstChild; k != nil; k = k.NextSibling {
				text(k)
			}
		}
		text(n)
		*out = append(*out, block{tag: n.Data, class: class, text: cleanText(sb.String())})
		return
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		collectBlocks(c, out)
	}
}

func cleanText(s string) string {
	s = strings.NewReplacer("­", "", " ", " ", "​", "").Replace(s)
	return strings.Join(strings.Fields(s), " ")
}

var (
	headingRe = regexp.MustCompile(`^(.*?),\s*punkt\s+(\d+[a-z]?)\s*\(([^)]*)\)\s*$`)
	// A betänkande with one förslagspunkt names no punkt: "Vilande förslag (V, MP)".
	soleHeadingRe = regexp.MustCompile(`^(.*?)\s*\(((?:\p{Lu}{1,3})(?:,\s*\p{Lu}{1,3})*)\)\s*$`)
	numberRe      = regexp.MustCompile(`^(\d+)\.$`)
	splitRe       = regexp.MustCompile(`,\s+|\s+och\s+`)
)

// ParseBetankandeHTML reads reservations, särskilda yttranden and the
// utskott's reasoning out of a betänkande's HTML by section heading:
// "Utskottets överväganden" (one h2 per rubrik, reasoning under "Utskottets
// ställningstagande"), "Reservationer" and "Särskilda yttranden". A section it
// does not find is left empty; nothing is guessed.
func ParseBetankandeHTML(r io.Reader) (*domain.BetankandeHTML, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return nil, err
	}
	var blocks []block
	collectBlocks(doc, &blocks)

	out := &domain.BetankandeHTML{}
	section := ""
	var rubrik string
	var reasoning []string
	capturing := false
	flushReasoning := func() {
		if rubrik != "" && len(reasoning) > 0 {
			out.Reasoning = append(out.Reasoning, domain.RubrikReasoning{Rubrik: rubrik, Text: strings.Join(reasoning, "\n\n")})
		}
		reasoning, capturing = nil, false
	}

	var cur *domain.Statement
	var curKind string
	var pendingNumber int
	stage := "" // "signers", "proposal", "text"
	var proposal, text []string
	flushStatement := func() {
		if cur != nil {
			cur.Proposal = strings.Join(proposal, "\n")
			cur.Text = strings.Join(text, "\n\n")
			out.Statements = append(out.Statements, *cur)
		}
		cur, proposal, text, stage = nil, nil, nil, ""
	}

	for _, b := range blocks {
		if b.isSection() {
			flushReasoning()
			flushStatement()
			rubrik = ""
			t := strings.ToLower(b.text)
			switch {
			case strings.HasPrefix(t, "utskottets överväganden"):
				section = "reasoning"
				out.HasReasoning = true
			case t == "reservationer" || t == "reservation":
				section, curKind = "statements", domain.KindReservation
				out.HasReservations = true
			case strings.HasPrefix(t, "särskilda yttranden") || strings.HasPrefix(t, "särskilt yttrande"):
				section, curKind = "statements", domain.KindSarskiltYttrande
			default:
				section = ""
			}
			continue
		}
		switch section {
		case "reasoning":
			switch {
			case b.tag == "h2":
				flushReasoning()
				rubrik = b.text
			case b.class == "R3" && strings.EqualFold(b.text, "Utskottets ställningstagande"):
				capturing = true
			case capturing && b.text != "" && b.class != "Mellanrum":
				reasoning = append(reasoning, b.text)
			}
		case "statements":
			switch {
			case b.class == "Bilaga" || b.class == "Bilagerubrik" || b.class == "Rubrik2bilaga" || b.tag == "h2" || b.tag == "h4":
				flushStatement()
				section = ""
			case b.class == "NormalNoll":
				flushStatement()
				if m := numberRe.FindStringSubmatch(b.text); m != nil {
					pendingNumber, _ = strconv.Atoi(m[1])
				} else {
					pendingNumber = 0
				}
			case b.class == "Reservationsrubrik" || b.class == "Srskiltyttranderubrik":
				flushStatement()
				cur = &domain.Statement{Kind: curKind, Number: pendingNumber}
				parties := ""
				if m := headingRe.FindStringSubmatch(b.text); m != nil {
					cur.Rubrik, cur.Punkt, parties = m[1], m[2], m[3]
				} else if m := soleHeadingRe.FindStringSubmatch(b.text); m != nil {
					cur.Rubrik, parties = m[1], m[2]
				} else {
					cur.Rubrik = b.text
				}
				for _, p := range strings.Split(parties, ",") {
					if p = strings.TrimSpace(p); p != "" {
						cur.Parties = append(cur.Parties, p)
					}
				}
				// The number belongs to this statement only.
				pendingNumber = 0
				stage = "signers"
			case cur == nil || b.text == "" || b.class == "Mellanrum":
			case b.class == "R4" && strings.EqualFold(b.text, "Förslag till riksdagsbeslut"):
				stage = "proposal"
			case b.class == "R4" && strings.EqualFold(b.text, "Ställningstagande"):
				stage = "text"
			case stage == "signers":
				if cur.Kind == domain.KindReservation {
					if strings.HasPrefix(b.text, "av ") {
						cur.Signers = splitSigners(strings.TrimPrefix(b.text, "av "))
						stage = "afterSigners"
					}
				} else {
					if s := strings.TrimSuffix(strings.TrimSuffix(b.text, ":"), " anför"); s != b.text && !strings.HasPrefix(s, "Jag") {
						cur.Signers = splitSigners(s)
					} else {
						text = append(text, b.text)
					}
					stage = "text"
				}
			case stage == "proposal":
				proposal = append(proposal, b.text)
			default:
				text = append(text, b.text)
			}
		}
	}
	flushReasoning()
	flushStatement()
	numberSole(out.Statements)
	return out, nil
}

// numberSole gives number 1 to a statement that is the only one of its kind
// and printed without a number ("Reservation", not "Reservationer"), as
// Riksdagen's own data does ("Utskottets förslag mot reservation 1").
func numberSole(ss []domain.Statement) {
	count := map[string]int{}
	for _, s := range ss {
		count[s.Kind]++
	}
	for i := range ss {
		if ss[i].Number == 0 && count[ss[i].Kind] == 1 {
			ss[i].Number = 1
		}
	}
}

func splitSigners(s string) []string {
	s = strings.TrimSuffix(strings.TrimSpace(s), ".")
	var out []string
	for _, p := range splitRe.Split(s, -1) {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// FetchBetankandeHTML reads /dokument/{dok_id}.html and parses it by section
// heading. A 404 is ErrHTMLUnavailable; any other failure is returned as is
// so the caller retries later.
func (c *Client) FetchBetankandeHTML(ctx context.Context, dokID string) (*domain.BetankandeHTML, error) {
	u := c.baseURL + "/dokument/" + url.PathEscape(dokID) + ".html"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch betänkande html %s: %w", dokID, err)
	}
	defer resp.Body.Close()
	switch {
	case resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusGone:
		return nil, ports.ErrHTMLUnavailable
	case resp.StatusCode != http.StatusOK:
		return nil, fmt.Errorf("riksdagen returned %d for %s", resp.StatusCode, req.URL.Path)
	}
	h, err := ParseBetankandeHTML(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("parse betänkande html %s: %w", dokID, err)
	}
	return h, nil
}

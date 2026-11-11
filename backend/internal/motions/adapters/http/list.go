package http

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"riksdagskollen/internal/motions"
	"riksdagskollen/internal/motions/domain"
)

const (
	defaultListLimit = 25
	maxListLimit     = 100
)

type listRowResponse struct {
	DokID       string          `json:"dokId"`
	Riksmote    string          `json:"riksmote"`
	Beteckning  string          `json:"beteckning"`
	Title       string          `json:"title"`
	Subtyp      string          `json:"subtyp"`
	Kind        string          `json:"kind"`
	Parties     []string        `json:"parties"`
	Date        *string         `json:"date"`
	Utskott     string          `json:"utskott"`
	UtskottName string          `json:"utskottName"`
	HandledIn   []string        `json:"handledIn"`
	SourceURL   string          `json:"sourceUrl"`
	Outcome     outcomeResponse `json:"outcome"`
}

type listResponse struct {
	Data       []listRowResponse `json:"data"`
	NextCursor *string           `json:"nextCursor"`
}

// csv reads a repeated or comma-separated query parameter.
func csv(vals []string) []string {
	var out []string
	for _, v := range vals {
		for _, part := range strings.Split(v, ",") {
			if p := strings.TrimSpace(part); p != "" {
				out = append(out, p)
			}
		}
	}
	return out
}

func encodeCursor(c domain.ListCursor) string {
	d := 0
	if c.Decided {
		d = 1
	}
	raw := fmt.Sprintf("%d|%s|%s", d, c.Key.UTC().Format(time.RFC3339Nano), c.DokID)
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

func decodeCursor(s string) (*domain.ListCursor, bool) {
	b, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return nil, false
	}
	parts := strings.SplitN(string(b), "|", 3)
	if len(parts) != 3 || (parts[0] != "0" && parts[0] != "1") || parts[2] == "" {
		return nil, false
	}
	t, err := time.Parse(time.RFC3339Nano, parts[1])
	if err != nil {
		return nil, false
	}
	return &domain.ListCursor{Decided: parts[0] == "1", Key: t, DokID: parts[2]}, true
}

var authorPattern = regexp.MustCompile(`^[0-9]{6,20}$`)

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	f := domain.ListFilter{
		Parties:  csv(q["party"]),
		Utskott:  csv(q["utskott"]),
		Kinds:    csv(q["kind"]),
		Outcomes: csv(q["outcome"]),
		Authors:  csv(q["author"]),
		Query:    strings.TrimSpace(q.Get("q")),
	}
	f.Riksmoten = csv(q["riksmote"])
	for _, rm := range f.Riksmoten {
		if !riksmotePattern.MatchString(rm) {
			jsonError(w, "invalid riksmote; write it 2017/18", http.StatusBadRequest)
			return
		}
	}
	for _, k := range f.Kinds {
		if !slices.Contains(domain.FilterableKinds, k) {
			jsonError(w, "invalid kind; one of "+strings.Join(domain.FilterableKinds, ", "), http.StatusBadRequest)
			return
		}
	}
	for _, o := range f.Outcomes {
		if !slices.Contains(domain.FilterableOutcomes, o) {
			jsonError(w, "invalid outcome; one of "+strings.Join(domain.FilterableOutcomes, ", "), http.StatusBadRequest)
			return
		}
	}
	for _, p := range append(slices.Clone(f.Parties), f.Utskott...) {
		if !beteckningPattern.MatchString(p) {
			jsonError(w, "invalid party or utskott", http.StatusBadRequest)
			return
		}
	}
	for _, a := range f.Authors {
		if !authorPattern.MatchString(a) {
			jsonError(w, "invalid author", http.StatusBadRequest)
			return
		}
	}
	if len(f.Query) > 200 {
		jsonError(w, "q is too long", http.StatusBadRequest)
		return
	}
	limit := defaultListLimit
	if v := q.Get("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 || n > maxListLimit {
			jsonError(w, fmt.Sprintf("limit must be 1..%d", maxListLimit), http.StatusBadRequest)
			return
		}
		limit = n
	}
	var after *domain.ListCursor
	if v := q.Get("cursor"); v != "" {
		c, ok := decodeCursor(v)
		if !ok {
			jsonError(w, "invalid cursor", http.StatusBadRequest)
			return
		}
		after = c
	}
	res, err := h.svc.List(r.Context(), f, after, limit)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, toListResponse(res))
}

func toListResponse(res *motions.ListResult) listResponse {
	out := listResponse{Data: make([]listRowResponse, 0, len(res.Items))}
	for _, it := range res.Items {
		m := it.Motion
		kind, _ := m.Kind()
		row := listRowResponse{
			DokID: m.DokID, Riksmote: m.Riksmote, Beteckning: m.Beteckning, Title: m.Title,
			Subtyp: m.Subtyp, Kind: kind, Parties: m.Parties(),
			HandledIn: append([]string{}, it.HandledIn...), Utskott: m.Organ, UtskottName: m.Tilldelat, SourceURL: sourceURLBase + m.DokID + ".html",
			Outcome: outcomeResponse{Word: it.Outcome.Word, Counts: it.Outcome.Counts, Total: it.Outcome.Total},
		}
		if m.Date != nil {
			d := m.Date.Format("2006-01-02")
			row.Date = &d
		}
		out.Data = append(out.Data, row)
	}
	if res.Next != nil {
		c := encodeCursor(*res.Next)
		out.NextCursor = &c
	}
	return out
}

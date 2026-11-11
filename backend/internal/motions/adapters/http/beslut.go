package http

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	committees "riksdagskollen/internal/committees/domain"
	"riksdagskollen/internal/motions"
	"riksdagskollen/internal/motions/domain"
)

type beslutPartyResponse struct {
	Party      string `json:"party"`
	Partilinje string `json:"partilinje"`
	Voteringar int    `json:"voteringar"`
}

type beslutRowResponse struct {
	Riksmote    string `json:"riksmote"`
	Beteckning  string `json:"beteckning"`
	Title       string `json:"title"`
	Utskott     string `json:"utskott"`
	UtskottName string `json:"utskottName"`
	// Date is the decision date for a decided row and the planned beslutsdag
	// for an upcoming one; null when Riksdagen gives none.
	Date      *string `json:"date"`
	Decided   bool    `json:"decided"`
	Status    string  `json:"status"`
	Outcome   string  `json:"outcome"`
	SourceURL string  `json:"sourceUrl"`
	// Parties is empty for an upcoming row.
	Parties []beslutPartyResponse `json:"parties"`
	// LoefteCount is "berör N löften": Löften with a Bekräftad koppling to a
	// votering of this betänkande. Zero for an upcoming row.
	LoefteCount int `json:"loefteCount"`
}

type beslutListResponse struct {
	Data       []beslutRowResponse `json:"data"`
	NextCursor *string             `json:"nextCursor"`
}

func encodeBeslutCursor(c domain.BeslutCursor) string {
	d := 0
	if c.Decided {
		d = 1
	}
	raw := fmt.Sprintf("%d|%s|%s|%s", d, c.Date.Format("2006-01-02"), c.Riksmote, c.Beteckning)
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

func decodeBeslutCursor(s string) (*domain.BeslutCursor, bool) {
	b, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return nil, false
	}
	parts := strings.Split(string(b), "|")
	if len(parts) != 4 || (parts[0] != "0" && parts[0] != "1") || !riksmotePattern.MatchString(parts[2]) || !beteckningPattern.MatchString(parts[3]) {
		return nil, false
	}
	t, err := time.Parse("2006-01-02", parts[1])
	if err != nil {
		return nil, false
	}
	return &domain.BeslutCursor{Decided: parts[0] == "1", Date: t, Riksmote: parts[2], Beteckning: parts[3]}, true
}

// listBeslut serves GET /beslut: betänkanden, upcoming first and then decided,
// filtered by state, utskott and party, paginated by cursor.
func (h *Handler) listBeslut(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	f := domain.BeslutFilter{
		State:   q.Get("state"),
		Utskott: csv(q["utskott"]),
		Parties: csv(q["party"]),
	}
	if f.State != "" && f.State != domain.BeslutDecided && f.State != domain.BeslutUpcoming {
		jsonError(w, "invalid state; one of decided, upcoming", http.StatusBadRequest)
		return
	}
	for _, p := range append(slices.Clone(f.Parties), f.Utskott...) {
		if !beteckningPattern.MatchString(p) {
			jsonError(w, "invalid party or utskott", http.StatusBadRequest)
			return
		}
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
	var after *domain.BeslutCursor
	if v := q.Get("cursor"); v != "" {
		c, ok := decodeBeslutCursor(v)
		if !ok {
			jsonError(w, "invalid cursor", http.StatusBadRequest)
			return
		}
		after = c
	}
	res, err := h.svc.ListBeslut(r.Context(), f, after, limit)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	out := toBeslutResponse(res)
	if !h.addLoefteCounts(r.Context(), w, out.Data) {
		return
	}
	jsonOK(w, out)
}

func toBeslutResponse(res *motions.BeslutResult) beslutListResponse {
	out := beslutListResponse{Data: make([]beslutRowResponse, 0, len(res.Items))}
	for _, it := range res.Items {
		row := beslutRowResponse{
			Riksmote: it.Riksmote, Beteckning: it.Beteckning, Title: it.Title,
			Utskott: it.Organ, UtskottName: committees.DisplayName(it.Organ),
			Decided: it.Decided, Status: it.Status, Outcome: it.Outcome,
			SourceURL: sourceURLBase + it.DokID + ".html",
			Parties:   make([]beslutPartyResponse, 0, len(it.Parties)),
		}
		if it.DecisionDate != "" {
			d := it.DecisionDate
			row.Date = &d
		} else if it.Decided {
			d := it.SortDate.Format("2006-01-02")
			row.Date = &d
		}
		for _, p := range it.Parties {
			row.Parties = append(row.Parties, beslutPartyResponse{Party: p.Party, Partilinje: p.Partilinje, Voteringar: p.Voteringar})
		}
		out.Data = append(out.Data, row)
	}
	if res.Next != nil {
		c := encodeBeslutCursor(*res.Next)
		out.NextCursor = &c
	}
	return out
}

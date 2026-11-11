package http

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"riksdagskollen/internal/loften"
)

var partyPattern = regexp.MustCompile(`^[A-Z]{1,4}$`)

type Handler struct{ svc *loften.Service }

func NewHandler(svc *loften.Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) Routes(r chi.Router) {
	r.Get("/loften", h.list)
	r.Get("/loften/{id}", h.get)
	r.Get("/valmanifest", h.valmanifest)
	r.Get("/parties/{party}/loften", h.list)
	r.Get("/parties/{party}/valmanifest", h.valmanifest)
}

// partyParam reads the party from the path or ?party=. ok is false (and a 400
// written) when it is present but malformed.
func partyParam(w http.ResponseWriter, r *http.Request) (string, bool) {
	p := chi.URLParam(r, "party")
	if p == "" {
		p = r.URL.Query().Get("party")
	}
	p = strings.ToUpper(p)
	if p != "" && !partyPattern.MatchString(p) {
		jsonError(w, "invalid party", http.StatusBadRequest)
		return "", false
	}
	return p, true
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	party, ok := partyParam(w, r)
	if !ok {
		return
	}
	out, err := h.svc.List(r.Context(), party)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, out)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	l, err := h.svc.Get(r.Context(), id)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if l == nil {
		jsonError(w, "not found", http.StatusNotFound)
		return
	}
	jsonOK(w, l)
}

func (h *Handler) valmanifest(w http.ResponseWriter, r *http.Request) {
	party, ok := partyParam(w, r)
	if !ok {
		return
	}
	out, err := h.svc.Valmanifest(r.Context(), party)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, out)
}

func jsonOK(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

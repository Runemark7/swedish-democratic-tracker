// Package http is the write API's inbound adapter. Routes exist only when a
// token is configured; the token is never logged or echoed.
package http

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"riksdagskollen/internal/curation"
	"riksdagskollen/internal/curation/domain"
)

type Handler struct {
	svc   *curation.Service
	token [32]byte
}

func NewHandler(svc *curation.Service, token string) *Handler {
	return &Handler{svc: svc, token: sha256.Sum256([]byte(token))}
}

// PublicRoutes is the read side: what changed and when.
func (h *Handler) PublicRoutes(r chi.Router) { r.Get("/aendringslogg", h.list) }

// WriteRoutes registers the token-protected writes under the given router.
func (h *Handler) WriteRoutes(r chi.Router) {
	r.Use(h.auth)
	r.Post("/valmanifest", h.upsertValmanifest)
	r.Post("/aendringslogg/{id}/undo", h.undo)
	r.Get("/candidates", h.candidates)
	r.Post("/kopplingar", h.submitSuggestions)
	r.Post("/loften", h.proposeLoefte)
	r.Post("/loften/{id}", h.editLoefte)
	r.Post("/loften/{id}/approve", h.approveLoefte)
	r.Post("/loften/{id}/strike", h.strikeLoefte)
}

func (h *Handler) auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
		sum := sha256.Sum256([]byte(got))
		if !ok || subtle.ConstantTimeCompare(sum[:], h.token[:]) != 1 {
			w.Header().Set("WWW-Authenticate", `Bearer realm="write"`)
			fail(w, http.StatusUnauthorized, "unauthorized", "missing or invalid bearer token")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (h *Handler) upsertValmanifest(w http.ResponseWriter, r *http.Request) {
	var in domain.ValmanifestInput
	if !readJSON(w, r, &in) {
		return
	}
	res, err := h.svc.UpsertValmanifest(r.Context(), r.Header.Get("Idempotency-Key"), in)
	h.respond(w, res, err)
}

// candidates is a read, but it exposes approved Löften next to voteringar the
// bot has not judged yet, so it sits behind the write token.
func (h *Handler) candidates(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	after, _ := strconv.ParseInt(r.URL.Query().Get("after"), 10, 64)
	c, err := h.svc.Candidates(r.Context(), limit, after)
	if err != nil {
		slog.Error("candidates", "error", err)
		fail(w, http.StatusInternalServerError, "internal", "internal error")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(c)
}

func (h *Handler) submitSuggestions(w http.ResponseWriter, r *http.Request) {
	var in domain.SuggestionBatch
	if !readJSON(w, r, &in) {
		return
	}
	res, err := h.svc.SubmitSuggestions(r.Context(), r.Header.Get("Idempotency-Key"), in)
	if err != nil {
		h.respond(w, nil, err)
		return
	}
	if res.Replayed {
		w.Header().Set("Idempotent-Replayed", "true")
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(res)
}

func (h *Handler) proposeLoefte(w http.ResponseWriter, r *http.Request) {
	var in domain.LoefteProposal
	if !readJSON(w, r, &in) {
		return
	}
	res, err := h.svc.ProposeLoefte(r.Context(), r.Header.Get("Idempotency-Key"), in)
	h.respond(w, res, err)
}

func pathID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id <= 0 {
		fail(w, http.StatusBadRequest, "bad_request", "invalid id")
		return 0, false
	}
	return id, true
}

func (h *Handler) editLoefte(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var in domain.LoefteEdit
	if !readJSON(w, r, &in) {
		return
	}
	res, err := h.svc.EditLoefte(r.Context(), r.Header.Get("Idempotency-Key"), id, in)
	h.respond(w, res, err)
}

func (h *Handler) approveLoefte(w http.ResponseWriter, r *http.Request) {
	h.decide(w, r, h.svc.ApproveLoefte)
}
func (h *Handler) strikeLoefte(w http.ResponseWriter, r *http.Request) {
	h.decide(w, r, h.svc.StrikeLoefte)
}

func (h *Handler) decide(w http.ResponseWriter, r *http.Request,
	fn func(context.Context, string, int64, domain.Provenance) (*domain.Result, error)) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var p domain.Provenance
	if !readJSON(w, r, &p) {
		return
	}
	res, err := fn(r.Context(), r.Header.Get("Idempotency-Key"), id, p)
	h.respond(w, res, err)
}

func (h *Handler) undo(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id <= 0 {
		fail(w, http.StatusBadRequest, "bad_request", "invalid entry id")
		return
	}
	var p domain.Provenance
	if !readJSON(w, r, &p) {
		return
	}
	res, err := h.svc.Undo(r.Context(), r.Header.Get("Idempotency-Key"), id, p)
	h.respond(w, res, err)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	before, _ := strconv.ParseInt(r.URL.Query().Get("before"), 10, 64)
	entries, err := h.svc.Entries(r.Context(), limit, before)
	if err != nil {
		slog.Error("aendringslogg list", "error", err)
		fail(w, http.StatusInternalServerError, "internal", "internal error")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"entries": entries})
}

func (h *Handler) respond(w http.ResponseWriter, res *domain.Result, err error) {
	var ve *domain.ValidationError
	switch {
	case err == nil:
		if res.Replayed {
			w.Header().Set("Idempotent-Replayed", "true")
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(res)
	case errors.As(err, &ve):
		fail(w, http.StatusUnprocessableEntity, "validation", ve.Msg)
	case errors.Is(err, domain.ErrNotFound):
		fail(w, http.StatusNotFound, "not_found", "not found")
	case errors.Is(err, domain.ErrKeyReuse):
		fail(w, http.StatusConflict, "idempotency_key_reused", err.Error())
	case errors.Is(err, domain.ErrConflict):
		fail(w, http.StatusConflict, "conflict", err.Error())
	default:
		slog.Error("write api", "error", err)
		fail(w, http.StatusInternalServerError, "internal", "internal error")
	}
}

func readJSON(w http.ResponseWriter, r *http.Request, v any) bool {
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<10))
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		fail(w, http.StatusBadRequest, "bad_request", "invalid JSON body: "+err.Error())
		return false
	}
	if _, err := dec.Token(); err != io.EOF {
		fail(w, http.StatusBadRequest, "bad_request", "invalid JSON body: trailing data")
		return false
	}
	return true
}

func fail(w http.ResponseWriter, status int, code, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg, "code": code})
}

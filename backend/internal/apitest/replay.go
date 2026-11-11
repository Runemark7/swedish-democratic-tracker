package apitest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"
)

// Replay serves recorded Riksdagen responses. A request that matches no
// fixture fails the test and returns an error status, so a missing recording
// cannot fall through to the live API.
type Replay struct {
	URL  string
	hits func() map[string]int
}

// NewReplay loads every *.json fixture in dir and serves them on a local
// server for the duration of the test.
func NewReplay(t *testing.T, dir string) *Replay {
	t.Helper()
	h, err := loadReplay(dir, func(format string, args ...any) {
		t.Errorf(format, args...)
	})
	if err != nil {
		t.Fatalf("replay fixtures: %v", err)
	}
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return &Replay{URL: srv.URL, hits: h.Hits}
}

// Hits returns how many times each fixture key was served. The key is
// "METHOD /path?query" with the query canonically encoded.
func (r *Replay) Hits() map[string]int {
	if r.hits == nil {
		return nil
	}
	return r.hits()
}

type fixtureFile struct {
	Method string            `json:"method"`
	Path   string            `json:"path"`
	Query  map[string]string `json:"query"`
	Status int               `json:"status"`
	Body   json.RawMessage   `json:"body"`
	// BodyFile names a recorded raw response (HTML) next to the fixture,
	// instead of an inline JSON body. ContentType is served with it.
	BodyFile    string `json:"bodyFile"`
	ContentType string `json:"contentType"`
}

type storedFixture struct {
	status      int
	body        []byte
	contentType string
}

type replayHandler struct {
	byKey  map[string]storedFixture
	onMiss func(string, ...any)
	mu     sync.Mutex
	hits   map[string]int
}

func (h *replayHandler) Hits() map[string]int {
	h.mu.Lock()
	defer h.mu.Unlock()
	out := make(map[string]int, len(h.hits))
	for k, n := range h.hits {
		out[k] = n
	}
	return out
}

func (h *replayHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	key := fixtureKey(r.Method, r.URL.Path, r.URL.Query())
	f, ok := h.byKey[key]
	if !ok {
		msg := fmt.Sprintf("unrecorded Riksdagen request: %s %s?%s", r.Method, r.URL.Path, r.URL.RawQuery)
		if h.onMiss != nil {
			h.onMiss("%s", msg)
		}
		http.Error(w, msg, http.StatusNotImplemented)
		return
	}
	h.mu.Lock()
	h.hits[key]++
	h.mu.Unlock()
	ct := f.contentType
	if ct == "" {
		ct = "application/json"
	}
	w.Header().Set("Content-Type", ct)
	w.WriteHeader(f.status)
	_, _ = w.Write(f.body)
}

func loadReplay(dir string, onMiss func(string, ...any)) (*replayHandler, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || !strings.EqualFold(filepath.Ext(e.Name()), ".json") {
			continue
		}
		names = append(names, e.Name())
	}
	sort.Strings(names)
	if len(names) == 0 {
		return nil, fmt.Errorf("no json fixtures in %s", dir)
	}

	byKey := make(map[string]storedFixture, len(names))
	for _, name := range names {
		raw, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return nil, err
		}
		var f fixtureFile
		if err := json.Unmarshal(raw, &f); err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
		if f.Method == "" || f.Path == "" {
			return nil, fmt.Errorf("%s: method and path are required", name)
		}
		body := []byte(f.Body)
		if f.BodyFile != "" {
			if len(f.Body) != 0 {
				return nil, fmt.Errorf("%s: set body or bodyFile, not both", name)
			}
			body, err = os.ReadFile(filepath.Join(dir, f.BodyFile))
			if err != nil {
				return nil, fmt.Errorf("%s: %w", name, err)
			}
		} else if len(f.Body) == 0 {
			return nil, fmt.Errorf("%s: body is required", name)
		}
		q := url.Values{}
		for k, v := range f.Query {
			q.Set(k, v)
		}
		key := fixtureKey(f.Method, f.Path, q)
		if _, exists := byKey[key]; exists {
			return nil, fmt.Errorf("%s: duplicate fixture for %s", name, key)
		}
		status := f.Status
		if status == 0 {
			status = http.StatusOK
		}
		byKey[key] = storedFixture{status: status, body: append([]byte(nil), body...), contentType: f.ContentType}
	}

	return &replayHandler{byKey: byKey, onMiss: onMiss, hits: map[string]int{}}, nil
}

func fixtureKey(method, path string, q url.Values) string {
	return strings.ToUpper(method) + " " + path + "?" + q.Encode()
}

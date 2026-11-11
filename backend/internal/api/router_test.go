package api

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestMain(m *testing.M) {
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
	os.Exit(m.Run())
}

// routesBeforeExtraction is the route list cmd/api registered inline before
// NewRouter existed. The constructor must serve this same list, with the same
// six root middlewares on every route: request id, real IP, logger, recoverer,
// the 100/minute rate limit, and CORS.
var routesBeforeExtraction = []string{
	"GET /api/betankanden/resolve",
	"GET /api/betankanden/{riksmote}/{beteckning}",
	"GET /api/beslut",
	"GET /api/budget/areas",
	"GET /api/budget/areas/{code}",
	"GET /api/budget/years",
	"GET /api/budget/years/{year}",
	"GET /api/budget/years/{year}/compare/{compareYear}",
	"GET /api/committees",
	"GET /api/committees/{code}",
	"GET /api/committees/{code}/goals",
	"GET /api/committees/{code}/votes",
	"GET /api/context/topic/{topic}",
	"GET /api/documents/{dokId}",
	"GET /api/documents/{dokId}/full",
	"GET /api/ingestion/freshness",
	"GET /api/ministers",
	"GET /api/ministers/{id}",
	"GET /api/motions",
	"GET /api/motions/backfill",
	"GET /api/motions/coverage",
	"GET /api/motions/{riksmote}/{beteckning}",
	"GET /api/municipalities",
	"GET /api/municipalities/budget/area/{areaName}",
	"GET /api/municipalities/kpi/{kpiCode}/ranking",
	"GET /api/municipalities/{code}",
	"GET /api/municipalities/{code}/budget/history",
	"GET /api/municipalities/{code}/kpi",
	"GET /api/municipalities/{code}/kpi-ranks",
	"GET /api/municipalities/{code}/population-trend",
	"GET /api/municipalities/{code}/procurement",
	"GET /api/municipalities/{code}/spending",
	"GET /api/parties",
	"GET /api/parties/{party}/goals",
	"GET /api/parties/{party}/goals/{goalId}/votes",
	"GET /api/party-meta",
	"GET /api/party-meta/{code}",
	"GET /api/politicians",
	"GET /api/politicians/{id}",
	"GET /api/politicians/{id}/promises",
	"GET /api/politicians/{id}/votes",
	"GET /api/regions",
	"GET /api/regions/budget/area/{areaName}",
	"GET /api/regions/kpi/{kpiCode}/ranking",
	"GET /api/regions/{code}",
	"GET /api/regions/{code}/budget",
	"GET /api/regions/{code}/budget/history",
	"GET /api/regions/{code}/kpi",
	"GET /api/regions/{code}/kpi-ranks",
	"GET /api/regions/{code}/plan",
	"GET /api/riksdag/agenda",
	"GET /api/riksdag/authorities",
	"GET /api/riksdag/authorities/list",
	"GET /api/riksdag/authorities/stats",
	"GET /api/riksdag/authorities/{slug}",
	"GET /api/riksdag/coverage",
	"GET /api/riksdag/government",
	"GET /api/riksdag/kpis",
	"GET /api/riksdag/live-votes",
	"GET /api/speeches/by-document/{dokId}",
	"GET /api/speeches/by-party/{party}",
	"GET /api/speeches/by-politician/{intressentId}",
	"GET /api/speeches/recent",
	"GET /api/speeches/{id}",
	"GET /api/summary",
	"GET /api/votes",
	"GET /api/votes/recent",
	"GET /api/votes/{beteckning}/{punkt}",
	"GET /healthz",
	"GET /readyz",
}

func TestRegisteredRoutesMatchThePreExtractionList(t *testing.T) {
	without := mustRouter(t, "")
	withToken := mustRouter(t, "phase-2-token")

	got := walkRoutes(t, without)
	tokenRoutes := walkRoutes(t, withToken)
	if strings.Join(got, "\n") != strings.Join(tokenRoutes, "\n") {
		t.Fatalf("a write token changed the route list:\nwithout:\n%s\nwith token:\n%s",
			strings.Join(got, "\n"), strings.Join(tokenRoutes, "\n"))
	}

	want := append([]string(nil), routesBeforeExtraction...)
	sort.Strings(want)
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("registered routes changed:\nmissing or extra lines below\n got:\n%s\nwant:\n%s",
			strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
}

func TestCORSAndRateLimitUnchanged(t *testing.T) {
	t.Run("cors", func(t *testing.T) {
		t.Setenv("CORS_ALLOWED_ORIGIN", "https://riksdagskollen.test")
		app := mustRouter(t, "")
		req := httptest.NewRequest(http.MethodOptions, "/healthz", nil)
		rec := httptest.NewRecorder()
		app.ServeHTTP(rec, req)
		if rec.Code != http.StatusNoContent {
			t.Fatalf("OPTIONS status = %d, want 204", rec.Code)
		}
		if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "https://riksdagskollen.test" {
			t.Fatalf("allow-origin = %q", got)
		}
		if got := rec.Header().Get("Access-Control-Allow-Methods"); got != "GET, OPTIONS" {
			t.Fatalf("allow-methods = %q", got)
		}
		if got := rec.Header().Get("Access-Control-Allow-Headers"); got != "Content-Type" {
			t.Fatalf("allow-headers = %q", got)
		}
	})

	t.Run("default origin", func(t *testing.T) {
		t.Setenv("CORS_ALLOWED_ORIGIN", "")
		app := mustRouter(t, "")
		req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
		rec := httptest.NewRecorder()
		app.ServeHTTP(rec, req)
		if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
			t.Fatalf("default allow-origin = %q", got)
		}
	})

	t.Run("rate limit", func(t *testing.T) {
		app := mustRouter(t, "")
		for i := 1; i <= 101; i++ {
			req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
			rec := httptest.NewRecorder()
			app.ServeHTTP(rec, req)
			if i == 1 && rec.Header().Get("X-RateLimit-Limit") != "100" {
				t.Fatalf("rate limit header = %q, want 100", rec.Header().Get("X-RateLimit-Limit"))
			}
			if i <= 100 && rec.Code != http.StatusOK {
				t.Fatalf("request %d status = %d, want 200", i, rec.Code)
			}
			if i == 101 && rec.Code != http.StatusTooManyRequests {
				t.Fatalf("request 101 status = %d, want 429", rec.Code)
			}
		}
	})
}

func mustRouter(t *testing.T, writeToken string) *Router {
	t.Helper()
	app, err := NewRouter(nil, writeToken, "")
	if err != nil {
		t.Fatal(err)
	}
	return app
}

func walkRoutes(t *testing.T, app *Router) []string {
	t.Helper()
	var got []string
	err := chi.Walk(app.mux, func(method, route string, _ http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		got = append(got, method+" "+route)
		if route == "/healthz" {
			var names []string
			for _, mw := range middlewares {
				names = append(names, runtime.FuncForPC(reflect.ValueOf(mw).Pointer()).Name())
			}
			joined := strings.Join(names, ",")
			for _, want := range []string{"RequestID", "RealIP", "Logger", "Recoverer", "httprate", "corsMiddleware"} {
				if !strings.Contains(joined, want) {
					t.Errorf("middleware stack = %s, missing %s", joined, want)
				}
			}
		}
		if len(middlewares) != 6 {
			t.Errorf("%s %s has %d middlewares, want 6", method, route, len(middlewares))
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(got)
	return got
}

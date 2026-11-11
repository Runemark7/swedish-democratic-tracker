package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// retiredRoutes answers the endpoints that belonged to the goals, keyword
// matching, scorecards and old promises features (issue #164) with 410 Gone and
// a pointer to the Löften API that replaced them.
//
// Kept for one release so a client that still calls them learns where to go
// instead of getting an opaque 404. Delete this file in the release after.
func retiredRoutes(api chi.Router) {
	for _, path := range []string{
		"/parties",
		"/parties/{party}/goals",
		"/parties/{party}/goals/{goalId}/votes",
		"/committees/{code}/goals",
		"/politicians/{id}/promises",
	} {
		api.Get(path, goneToLoften)
	}
}

func goneToLoften(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Link", `</api/loften>; rel="successor-version"`)
	w.WriteHeader(http.StatusGone)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"error":   "gone",
		"message": "This endpoint was retired. Party goals and promises are now Löften.",
		"loften":  "/api/loften",
		"summary": "/api/loften/summary",
	})
}

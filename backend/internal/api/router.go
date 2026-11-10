// Package api builds the HTTP router the process serves.
//
// NewRouter is the seam tests use: a database pool, the Phase 2 write token
// (empty until that API exists) and the Riksdagen base URL. cmd/api calls the
// same constructor. Passing a replay server URL makes every Riksdagen client
// read recorded responses instead of the live API.
package api

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	"github.com/jackc/pgx/v5/pgxpool"

	"riksdagskollen/internal/budget"
	budgetHTTP "riksdagskollen/internal/budget/adapters/http"
	budgetPG "riksdagskollen/internal/budget/adapters/postgres"
	committeesFeature "riksdagskollen/internal/committees"
	committeesHTTP "riksdagskollen/internal/committees/adapters/http"
	committeesPG "riksdagskollen/internal/committees/adapters/postgres"
	topiccontext "riksdagskollen/internal/context"
	contextHTTP "riksdagskollen/internal/context/adapters/http"
	contextPG "riksdagskollen/internal/context/adapters/postgres"
	"riksdagskollen/internal/goals"
	goalsHTTP "riksdagskollen/internal/goals/adapters/http"
	goalsPG "riksdagskollen/internal/goals/adapters/postgres"
	"riksdagskollen/internal/ingestion"
	ingestionHTTP "riksdagskollen/internal/ingestion/adapters/http"
	ingestionPG "riksdagskollen/internal/ingestion/adapters/postgres"
	"riksdagskollen/internal/ingestion/workers"
	"riksdagskollen/internal/matching"
	matchingPG "riksdagskollen/internal/matching/adapters/postgres"
	matchingStub "riksdagskollen/internal/matching/adapters/stub"
	"riksdagskollen/internal/ministers"
	ministersHTTP "riksdagskollen/internal/ministers/adapters/http"
	ministersPG "riksdagskollen/internal/ministers/adapters/postgres"
	ministersRD "riksdagskollen/internal/ministers/adapters/riksdagen"
	"riksdagskollen/internal/motions"
	motionsHTTP "riksdagskollen/internal/motions/adapters/http"
	motionsPG "riksdagskollen/internal/motions/adapters/postgres"
	motionsRD "riksdagskollen/internal/motions/adapters/riksdagen"
	"riksdagskollen/internal/parties"
	partiesHTTP "riksdagskollen/internal/parties/adapters/http"
	partiesPG "riksdagskollen/internal/parties/adapters/postgres"
	"riksdagskollen/internal/politicians"
	politiciansHTTP "riksdagskollen/internal/politicians/adapters/http"
	politiciansPG "riksdagskollen/internal/politicians/adapters/postgres"
	politiciansRD "riksdagskollen/internal/politicians/adapters/riksdagen"
	"riksdagskollen/internal/promises"
	promisesHTTP "riksdagskollen/internal/promises/adapters/http"
	promisesPG "riksdagskollen/internal/promises/adapters/postgres"
	"riksdagskollen/internal/regions"
	regionsHTTP "riksdagskollen/internal/regions/adapters/http"
	koladaAdapter "riksdagskollen/internal/regions/adapters/kolada"
	regionsPG "riksdagskollen/internal/regions/adapters/postgres"
	scbAdapter "riksdagskollen/internal/regions/adapters/scb"
	tedAdapter "riksdagskollen/internal/regions/adapters/ted"
	"riksdagskollen/internal/riksdag"
	riksdagHTTP "riksdagskollen/internal/riksdag/adapters/http"
	riksdagMF "riksdagskollen/internal/riksdag/adapters/myndighetsforteckning"
	riksdagPG "riksdagskollen/internal/riksdag/adapters/postgres"
	riksdagRegistret "riksdagskollen/internal/riksdag/adapters/registret"
	riksdagRD "riksdagskollen/internal/riksdag/adapters/riksdagen"
	riksdagSCB "riksdagskollen/internal/riksdag/adapters/scb"
	riksdagSCBMacro "riksdagskollen/internal/riksdag/adapters/scbmacro"
	riksdagStatic "riksdagskollen/internal/riksdag/adapters/static"
	riksdagSK "riksdagskollen/internal/riksdag/adapters/statskontoret"
	"riksdagskollen/internal/speeches"
	speechesHTTP "riksdagskollen/internal/speeches/adapters/http"
	speechesPG "riksdagskollen/internal/speeches/adapters/postgres"
	speechesRD "riksdagskollen/internal/speeches/adapters/riksdagen"
	"riksdagskollen/internal/votes"
	votesHTTP "riksdagskollen/internal/votes/adapters/http"
	votesPG "riksdagskollen/internal/votes/adapters/postgres"
	votesRD "riksdagskollen/internal/votes/adapters/riksdagen"
)

const productionRiksdagenBaseURL = "https://data.riksdagen.se"

// Router is the HTTP API. It implements http.Handler. Start and
// RunInitialSync arm the ingestion scheduler; tests leave them uncalled and
// use RunWorker instead, so a test does not start the cron or the boot sync.
type Router struct {
	mux        *chi.Mux
	sched      *ingestion.Scheduler
	clk        *clock
	writeToken string
}

// NewRouter builds the production router. writeToken is empty until the
// Phase 2 write API exists; a non-empty token does not register routes yet.
// riksdagenBaseURL is the origin for every data.riksdagen.se client. Empty
// means the production origin.
func NewRouter(db *pgxpool.Pool, writeToken, riksdagenBaseURL string) (*Router, error) {
	riksdagenBaseURL = normalizeBase(riksdagenBaseURL)
	allowedOrigin := envOr("CORS_ALLOWED_ORIGIN", "http://localhost:5173")
	clk := &clock{}

	polRepo := politiciansPG.NewRepository(db)
	polRD := politiciansRD.NewClientWithBaseURL(riksdagenBaseURL)
	polSvc := politicians.NewService(polRepo, polRD)
	polHandler := politiciansHTTP.NewHandler(polSvc)

	speechRepo := speechesPG.NewRepository(db)
	speechRD := speechesRD.NewClientWithBaseURL(riksdagenBaseURL)
	speechSvc := speeches.NewService(speechRepo, speechRD)
	speechHandler := speechesHTTP.NewHandler(speechSvc, polSvc)

	voteRepo := votesPG.NewRepository(db)
	voteRD := votesRD.NewClientWithBaseURL(riksdagenBaseURL)
	voteSvc := votes.NewService(voteRepo, voteRD)
	voteHandler := votesHTTP.NewHandler(voteSvc)

	motionSvc := motions.NewService(motionsPG.NewRepository(db), motionsRD.NewClientWithBaseURL(riksdagenBaseURL))
	motionHandler := motionsHTTP.NewHandler(motionSvc)

	goalRepo := goalsPG.NewRepository(db)
	goalSvc := goals.NewService(goalRepo)

	matchRepo := matchingPG.NewRepository(db)
	aiStub := matchingStub.NewAIService()
	matchSvc := matching.NewService(matchRepo, aiStub)

	goalsHandler := goalsHTTP.NewHandler(goalSvc, matchSvc)

	promiseRepo := promisesPG.NewRepository(db)
	promiseSvc := promises.NewService(promiseRepo)
	promisesHandler := promisesHTTP.NewHandler(promiseSvc, matchSvc)

	budgetRepo := budgetPG.NewRepository(db)
	budgetSvc := budget.NewService(budgetRepo)
	budgetHandler := budgetHTTP.NewHandler(budgetSvc)

	contextRepo := contextPG.NewRepository(db)
	contextSvc := topiccontext.NewService(contextRepo)
	contextHandler := contextHTTP.NewHandler(contextSvc)

	regionsRepo := regionsPG.NewRepository(db)
	koladaCl := koladaAdapter.NewClient()
	scbCl := scbAdapter.NewClient()
	tedCl := tedAdapter.NewClient()
	regionsSvc := regions.NewService(regionsRepo, koladaCl, scbCl, tedCl)
	regionsHandler := regionsHTTP.NewHandler(regionsSvc)

	riksdagSvc := riksdag.NewServiceWithSCB(riksdagSK.NewClient(), riksdagStatic.NewClient(), riksdagSCB.NewClient())
	kpiRepo := riksdagPG.NewKpiRepository(db)
	riksdagSvc.SetKpiRepo(kpiRepo)
	riksdagSvc.SetCoverageRepo(riksdagPG.NewCoverageRepository(db))
	riksdagSvc.SetGovRepo(riksdagPG.NewGovRepository(db))
	riksdagSvc.SetAgendaRepo(riksdagPG.NewAgendaRepository(db))
	riksdagSvc.SetLiveVotesRepo(riksdagPG.NewLiveVotesRepository(db))
	agencyIntelRepo := riksdagPG.NewAgencyIntelRepository(db)
	authorityRepo := riksdagPG.NewAuthorityRepository(db)
	riksdagSvc.SetAgencyIntelRepo(agencyIntelRepo)
	riksdagSvc.SetAuthorityRepo(authorityRepo)
	riksdagHandler := riksdagHTTP.NewHandler(riksdagSvc)

	partiesRepo := partiesPG.NewRepository(db)
	partiesSvc := parties.NewService(partiesRepo)
	partiesHandler := partiesHTTP.NewHandler(partiesSvc)

	ministersRepo := ministersPG.NewRepository(db)
	ministersRDClient := ministersRD.NewClientWithBaseURL(riksdagenBaseURL)
	ministersSvc := ministers.NewService(ministersRepo, ministersRDClient)
	ministersHandler := ministersHTTP.NewHandler(ministersSvc)

	committeesSvc := committeesFeature.NewService(committeesPG.NewRepository(db))
	committeesHandler := committeesHTTP.NewHandler(committeesSvc)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(httprate.LimitByIP(100, time.Minute))
	r.Use(corsMiddleware(allowedOrigin))

	// Liveness only. A database blip must not restart the pod: readiness is
	// /readyz, which pings the database and drops the pod out of service
	// until the ping succeeds again.
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"status":"ok"}`)
	})
	r.Get("/readyz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Content-Type", "application/json")
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := db.Ping(ctx); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprint(w, `{"status":"unavailable"}`)
			return
		}
		fmt.Fprint(w, `{"status":"ok"}`)
	})

	var apiRouter chi.Router
	r.Route("/api", func(api chi.Router) {
		apiRouter = api
		polHandler.Routes(api)
		voteHandler.Routes(api)
		motionHandler.Routes(api)
		speechHandler.Routes(api)
		goalsHandler.Routes(api)
		promisesHandler.Routes(api)
		budgetHandler.Routes(api)
		contextHandler.Routes(api)
		regionsHandler.Routes(api)
		riksdagHandler.Routes(api)
		partiesHandler.Routes(api)
		ministersHandler.Routes(api)
		committeesHandler.Routes(api)

		// TODO: return aggregate 24h decision counts per level for the homepage pulse strip.
		// Shape: { riksdag: number, region: number, kommun: number, total: number, buckets: number[] }
		api.Get("/summary", func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"ok":true}`)
		})
	})

	cursorRepo := ingestionPG.NewCursorRepository(db)
	ingestionRunsRepo := ingestionPG.NewIngestionRunRepository(db)
	freshnessRepo := ingestionPG.NewFreshnessRepository(db)
	pollWorker := workers.NewPoliticiansWorker(polSvc)
	speechWorker := workers.NewSpeechesWorker(speechSvc, cursorRepo).WithMandates(ingestionPG.NewMandateRepository(db))
	enrichSpeechesWorker := workers.NewEnrichSpeechesWorker(speechSvc, 200)
	voteWorker := workers.NewVotesWorker(voteSvc, cursorRepo)
	voteWorker.SetNow(clk.Now)
	keywordWorker := workers.NewKeywordMatcherWorker(goalSvc, voteSvc, matchSvc)
	refreshWorker := workers.NewRefreshScorecardsWorker(matchSvc)
	coverageWorker := workers.NewCoverageWorker(voteSvc, ingestionPG.NewCoverageRepo(db))

	agencyIntelWorker := workers.NewAgencyIntelWorker(riksdagRD.NewAgencyClientWithBaseURL(riksdagenBaseURL), agencyIntelRepo, agencyInfoList())
	authoritiesWorker := workers.NewAuthoritiesWorker(riksdagRegistret.NewClient(), authorityRepo)
	headcountWorker := workers.NewAuthorityHeadcountWorker(riksdagMF.NewClient(), authorityRepo)
	expenditureWorker := workers.NewAuthorityExpenditureWorker(riksdagSK.NewClient(), authorityRepo)
	nationalKpisWorker := workers.NewNationalKpisWorker(riksdagSCBMacro.NewClient(), kpiRepo)

	sched := ingestion.NewScheduler()
	sched.SetRunRecorder(ingestionRunsRepo)
	sched.SetFreshness(freshnessRepo)
	if err := sched.RegisterDefaults(pollWorker, speechWorker, voteWorker, keywordWorker, refreshWorker, coverageWorker); err != nil {
		slog.Error("failed to register ingestion workers", "error", err)
		return nil, err
	}
	motionsWorker := workers.NewMotionsWorker(motionSvc, ingestionPG.NewMandateRepository(db))
	motionsWorker.SetNow(clk.Now)
	if err := sched.RegisterSync("@daily", &motionsWorker); err != nil {
		slog.Error("failed to register fetch-motions worker", "error", err)
		return nil, err
	}
	betankandenWorker := workers.NewBetankandenWorker(motionSvc)
	if err := sched.RegisterSync("@daily", &betankandenWorker); err != nil {
		slog.Error("failed to register fetch-betankanden worker", "error", err)
		return nil, err
	}
	if err := sched.RegisterSync("@daily", &enrichSpeechesWorker); err != nil {
		slog.Error("failed to register enrich-speech-text worker", "error", err)
		return nil, err
	}
	if err := sched.RegisterSync("@weekly", &agencyIntelWorker); err != nil {
		slog.Error("failed to register agency-intel worker", "error", err)
		return nil, err
	}
	if err := sched.RegisterSync("@weekly", &authoritiesWorker); err != nil {
		slog.Error("failed to register authorities worker", "error", err)
	}
	if err := sched.RegisterSync("@weekly", &headcountWorker); err != nil {
		slog.Error("failed to register authority-headcount worker", "error", err)
	}
	if err := sched.RegisterSync("@weekly", &expenditureWorker); err != nil {
		slog.Error("failed to register authority-expenditure worker", "error", err)
	}
	if err := sched.RegisterSync("@daily", &nationalKpisWorker); err != nil {
		slog.Error("failed to register national-kpis worker", "error", err)
	}
	regionBudgetWorker := workers.NewRegionBudgetWorker(regionsSvc, ingestionRunsRepo)
	if err := sched.RegisterSync("@weekly", &regionBudgetWorker); err != nil {
		slog.Error("failed to register region-budget worker", "error", err)
		return nil, err
	}
	munBudgetWorker := workers.NewMunicipalityBudgetWorker(regionsSvc, ingestionRunsRepo)
	if err := sched.RegisterSync("@weekly", &munBudgetWorker); err != nil {
		slog.Error("failed to register municipality-budget worker", "error", err)
		return nil, err
	}
	proposalsWorker := workers.NewProposalsWorker(ministersSvc)
	if err := sched.RegisterSync("@weekly", &proposalsWorker); err != nil {
		slog.Error("failed to register proposals worker", "error", err)
		return nil, err
	}
	apiRouter.Get("/ingestion/freshness", ingestionHTTP.NewHandler(freshnessRepo, sched.Catalog()).ServeFreshness)

	rt := &Router{mux: r, sched: sched, clk: clk, writeToken: writeToken}
	// Phase 2 registers the write API here when writeToken is set. Until then
	// the call adds no routes, which the route-list test locks.
	rt.registerWriteRoutes()
	return rt, nil
}

// registerWriteRoutes is the Phase 2 extension point. An empty token and a
// configured token both leave the public route list unchanged.
func (r *Router) registerWriteRoutes() {
	if r.writeToken == "" {
		return
	}
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}

// Start arms the daily ingestion cron. Tests do not call it.
func (r *Router) Start() { r.sched.Start() }

// Stop stops the ingestion cron.
func (r *Router) Stop() { r.sched.Stop() }

// RunInitialSync runs the whole ingestion chain once, including weekly workers.
func (r *Router) RunInitialSync(ctx context.Context) error {
	return r.sched.RunInitialSync(ctx)
}

// RunWorker runs one registered ingestion worker and returns its error.
func (r *Router) RunWorker(ctx context.Context, name string) error {
	return r.sched.RunNamed(ctx, name)
}

// SetNow pins the clock the votes worker uses to choose a riksmöte.
// Production leaves it unset, which means time.Now.
func (r *Router) SetNow(now func() time.Time) { r.clk.Set(now) }

type clock struct {
	mu  sync.Mutex
	now func() time.Time
}

func (c *clock) Now() time.Time {
	c.mu.Lock()
	fn := c.now
	c.mu.Unlock()
	if fn == nil {
		return time.Now()
	}
	return fn()
}

func (c *clock) Set(now func() time.Time) {
	c.mu.Lock()
	c.now = now
	c.mu.Unlock()
}

func normalizeBase(base string) string {
	base = trimSlash(base)
	if base == "" {
		return productionRiksdagenBaseURL
	}
	return base
}

func trimSlash(s string) string {
	for len(s) > 0 && s[len(s)-1] == '/' {
		s = s[:len(s)-1]
	}
	return s
}

func corsMiddleware(origin string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// agencyInfoList returns the static list of tracked agencies for the ingestion worker.
// Slugs must match the toSlug() function in the riksdag service.
func agencyInfoList() []workers.AgencyInfo {
	return []workers.AgencyInfo{
		{Name: "Polismyndigheten", Slug: "polismyndigheten"},
		{Name: "Kriminalvården", Slug: "kriminalvarden"},
		{Name: "Försäkringskassan", Slug: "forsakringskassan"},
		{Name: "Skatteverket", Slug: "skatteverket"},
		{Name: "Sveriges Domstolar", Slug: "sveriges-domstolar"},
		{Name: "Arbetsförmedlingen", Slug: "arbetsformedlingen"},
		{Name: "Migrationsverket", Slug: "migrationsverket"},
		{Name: "Tullverket", Slug: "tullverket"},
		{Name: "Åklagarmyndigheten", Slug: "aklagarmyndigheten"},
		{Name: "Säkerhetspolisen", Slug: "sakerhetspolisen"},
	}
}

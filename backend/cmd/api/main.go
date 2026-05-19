package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"

	// Feature: politicians
	politiciansHTTP "riksdagskollen/internal/politicians/adapters/http"
	politiciansPG "riksdagskollen/internal/politicians/adapters/postgres"
	politiciansRD "riksdagskollen/internal/politicians/adapters/riksdagen"
	"riksdagskollen/internal/politicians"

	// Feature: speeches
	speechesPG "riksdagskollen/internal/speeches/adapters/postgres"
	speechesRD "riksdagskollen/internal/speeches/adapters/riksdagen"
	speechesHTTP "riksdagskollen/internal/speeches/adapters/http"
	"riksdagskollen/internal/speeches"

	// Feature: votes
	votesHTTP "riksdagskollen/internal/votes/adapters/http"
	votesPG "riksdagskollen/internal/votes/adapters/postgres"
	votesRD "riksdagskollen/internal/votes/adapters/riksdagen"
	"riksdagskollen/internal/votes"

	// Feature: goals + matching
	goalsHTTP "riksdagskollen/internal/goals/adapters/http"
	goalsPG "riksdagskollen/internal/goals/adapters/postgres"
	"riksdagskollen/internal/goals"
	matchingPG "riksdagskollen/internal/matching/adapters/postgres"
	matchingStub "riksdagskollen/internal/matching/adapters/stub"
	"riksdagskollen/internal/matching"

	// Feature: promises
	promisesHTTP "riksdagskollen/internal/promises/adapters/http"
	promisesPG "riksdagskollen/internal/promises/adapters/postgres"
	"riksdagskollen/internal/promises"

	// Feature: budget
	budgetHTTP "riksdagskollen/internal/budget/adapters/http"
	budgetPG "riksdagskollen/internal/budget/adapters/postgres"
	"riksdagskollen/internal/budget"

	// Feature: context
	contextHTTP "riksdagskollen/internal/context/adapters/http"
	contextPG "riksdagskollen/internal/context/adapters/postgres"
	topiccontext "riksdagskollen/internal/context"

	// Feature: regions + municipalities
	regionsHTTP "riksdagskollen/internal/regions/adapters/http"
	koladaAdapter "riksdagskollen/internal/regions/adapters/kolada"
	regionsPG "riksdagskollen/internal/regions/adapters/postgres"
	scbAdapter "riksdagskollen/internal/regions/adapters/scb"
	tedAdapter "riksdagskollen/internal/regions/adapters/ted"
	"riksdagskollen/internal/regions"
	"riksdagskollen/internal/regions/seeder"

	// Feature: riksdag
	riksdagHTTP "riksdagskollen/internal/riksdag/adapters/http"
	riksdagPG "riksdagskollen/internal/riksdag/adapters/postgres"
	riksdagRD "riksdagskollen/internal/riksdag/adapters/riksdagen"
	riksdagSCB "riksdagskollen/internal/riksdag/adapters/scb"
	riksdagStatic "riksdagskollen/internal/riksdag/adapters/static"
	riksdagSK "riksdagskollen/internal/riksdag/adapters/statskontoret"
	"riksdagskollen/internal/riksdag"

	// Ingestion
	"riksdagskollen/internal/ingestion"
	ingestionPG "riksdagskollen/internal/ingestion/adapters/postgres"
	"riksdagskollen/internal/ingestion/workers"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	dbURL := mustEnv("DATABASE_URL")
	port := envOr("PORT", "8080")
	allowedOrigin := envOr("CORS_ALLOWED_ORIGIN", "http://localhost:5173")
	migrationsPath := envOr("MIGRATIONS_PATH", "file://migrations")

	// -- Database --
	db, err := connectDB(context.Background(), dbURL)
	if err != nil {
		slog.Error("database unavailable after retries", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// -- Migrations --
	m, err := migrate.New(migrationsPath, dbURL)
	if err != nil {
		slog.Error("failed to create migrator", "error", err)
		os.Exit(1)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		slog.Error("migration failed", "error", err)
		os.Exit(1)
	}
	slog.Info("migrations applied")

	// -- Election seeder (runs in background if SEED_ELECTIONS=true) --
	// The seeder is idempotent: per-kommun DELETE+INSERT guarded by total>0,
	// so transient SCB failures preserve existing data. Running on every
	// startup fills in any gaps left by previous partial seeds.
	if envOr("SEED_ELECTIONS", "false") == "true" {
		go func() {
			slog.Info("running election seeder (idempotent fill)")
			if err := seeder.Run(context.Background(), db); err != nil {
				slog.Error("election seeder failed", "error", err)
			}
		}()
	}

	// -- Wire features --
	polRepo := politiciansPG.NewRepository(db)
	polRD := politiciansRD.NewClient()
	polSvc := politicians.NewService(polRepo, polRD)
	polHandler := politiciansHTTP.NewHandler(polSvc)

	speechRepo := speechesPG.NewRepository(db)
	speechRD := speechesRD.NewClient()
	speechSvc := speeches.NewService(speechRepo, speechRD)
	speechHandler := speechesHTTP.NewHandler(speechSvc, polSvc)

	voteRepo := votesPG.NewRepository(db)
	voteRD := votesRD.NewClient()
	voteSvc := votes.NewService(voteRepo, voteRD)
	voteHandler := votesHTTP.NewHandler(voteSvc)

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
	riksdagSvc.SetKpiRepo(riksdagPG.NewKpiRepository(db))
	riksdagSvc.SetGovRepo(riksdagPG.NewGovRepository(db))
	riksdagSvc.SetAgendaRepo(riksdagPG.NewAgendaRepository(db))
	riksdagSvc.SetLiveVotesRepo(riksdagPG.NewLiveVotesRepository(db))
	agencyIntelRepo := riksdagPG.NewAgencyIntelRepository(db)
	riksdagSvc.SetAgencyIntelRepo(agencyIntelRepo)
	riksdagHandler := riksdagHTTP.NewHandler(riksdagSvc)

	// -- Router --
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(httprate.LimitByIP(100, time.Minute))
	r.Use(corsMiddleware(allowedOrigin))

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"status":"ok"}`)
	})

	r.Route("/api", func(r chi.Router) {
		polHandler.Routes(r)
		voteHandler.Routes(r)
		speechHandler.Routes(r)
		goalsHandler.Routes(r)
		promisesHandler.Routes(r)
		budgetHandler.Routes(r)
		contextHandler.Routes(r)
		regionsHandler.Routes(r)
		riksdagHandler.Routes(r)

		// TODO: return aggregate 24h decision counts per level for the homepage pulse strip.
		// Shape: { riksdag: number, region: number, kommun: number, total: number, buckets: number[] }
		r.Get("/summary", func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"ok":true}`)
		})
	})

	// -- Ingestion scheduler --
	cursorRepo := ingestionPG.NewCursorRepository(db)
	ingestionRunsRepo := ingestionPG.NewIngestionRunRepository(db)
	pollWorker := workers.NewPoliticiansWorker(polSvc)
	speechWorker := workers.NewSpeechesWorker(speechSvc, cursorRepo)
	enrichSpeechesWorker := workers.NewEnrichSpeechesWorker(speechSvc, 200)
	voteWorker := workers.NewVotesWorker(voteSvc, cursorRepo)
	enrichWorker := workers.NewEnrichOriginsWorker(voteSvc)
	keywordWorker := workers.NewKeywordMatcherWorker(goalSvc, voteSvc, matchSvc)
	refreshWorker := workers.NewRefreshScorecardsWorker(matchSvc)

	agencyIntelWorker := workers.NewAgencyIntelWorker(riksdagRD.NewAgencyClient(), agencyIntelRepo, agencyInfoList())

	sched := ingestion.NewScheduler()
	if err := sched.RegisterDefaults(pollWorker, speechWorker, voteWorker, enrichWorker, keywordWorker, refreshWorker); err != nil {
		slog.Error("failed to register ingestion workers", "error", err)
		os.Exit(1)
	}
	if err := sched.RegisterSync("@daily", &enrichSpeechesWorker); err != nil {
		slog.Error("failed to register enrich-speech-text worker", "error", err)
		os.Exit(1)
	}
	if err := sched.RegisterSync("@weekly", &agencyIntelWorker); err != nil {
		slog.Error("failed to register agency-intel worker", "error", err)
		os.Exit(1)
	}
	regionBudgetWorker := workers.NewRegionBudgetWorker(regionsSvc, ingestionRunsRepo)
	if err := sched.Register("@weekly", &regionBudgetWorker); err != nil {
		slog.Error("failed to register region-budget worker", "error", err)
		os.Exit(1)
	}
	sched.Start()
	defer sched.Stop()

	// -- Initial data sync (background) --
	if envOr("INITIAL_SYNC", "true") == "true" {
		go func() {
			syncCtx, syncCancel := context.WithTimeout(context.Background(), 20*time.Minute)
			defer syncCancel()
			if err := sched.RunInitialSync(syncCtx); err != nil {
				slog.Error("initial sync failed", "error", err)
			}
		}()
	}

	// -- HTTP server --
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("server listening", "port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("shutdown error", "error", err)
	}
}

// connectDB retries the pgxpool connection with exponential backoff.
// Needed because Docker's internal DNS may not resolve service names
// immediately even after the healthcheck passes.
func connectDB(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	const maxAttempts = 10
	backoff := 1 * time.Second
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		db, err := pgxpool.New(ctx, dsn)
		if err == nil {
			if pingErr := db.Ping(ctx); pingErr == nil {
				slog.Info("database connected", "attempt", attempt)
				return db, nil
			} else {
				db.Close()
				err = pingErr
			}
		}
		if attempt == maxAttempts {
			return nil, err
		}
		slog.Warn("database not ready, retrying", "attempt", attempt, "backoff", backoff, "error", err)
		time.Sleep(backoff)
		if backoff < 16*time.Second {
			backoff *= 2
		}
	}
	return nil, fmt.Errorf("database unreachable after %d attempts", maxAttempts)
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

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		slog.Error("required environment variable missing", "key", key)
		os.Exit(1)
	}
	return v
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

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

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"

	"riksdagskollen/internal/api"
	"riksdagskollen/internal/regions/seeder"
)

// riksdagenBaseURL is the production data.riksdagen.se origin. Tests build the
// same router with a replay server URL instead of this constant.
const riksdagenBaseURL = "https://data.riksdagen.se"

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	dbURL := mustEnv("DATABASE_URL")
	port := envOr("PORT", "8080")
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

	// WRITE_TOKEN is empty until the Phase 2 write API exists. NewRouter
	// accepts it so that API can be registered without a new constructor.
	app, err := api.NewRouter(db, os.Getenv("WRITE_TOKEN"), riksdagenBaseURL)
	if err != nil {
		os.Exit(1)
	}

	// -- Ingestion scheduler --
	app.Start()
	defer app.Stop()

	// -- Initial data sync (background) --
	if envOr("INITIAL_SYNC", "true") == "true" {
		go func() {
			// Sequential chain, one timeout per worker. 20 minutes for the
			// whole chain cancelled every worker after the first slow one,
			// including the speech catch-up a deploy relies on.
			syncCtx, syncCancel := context.WithTimeout(context.Background(), 3*time.Hour)
			defer syncCancel()
			if err := app.RunInitialSync(syncCtx); err != nil {
				slog.Error("initial sync failed", "error", err)
			}
		}()
	}

	// -- HTTP server --
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      app,
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

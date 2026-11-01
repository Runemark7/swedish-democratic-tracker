package apitest

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
)

// OpenDB creates a database, applies every migration with golang-migrate, and
// drops the database when the test ends. TEST_DATABASE_URL wins over
// DATABASE_URL so CI's Postgres service is used when both are set. The test
// is skipped when neither is set.
//
// The role in the URL must be allowed to create databases. Each call gets its
// own database, so tests do not see each other's rows.
func OpenDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	raw := os.Getenv("TEST_DATABASE_URL")
	if raw == "" {
		raw = os.Getenv("DATABASE_URL")
	}
	if raw == "" {
		t.Skip("TEST_DATABASE_URL or DATABASE_URL not set — skipping API seam tests")
	}

	adminDSN, err := withDatabase(raw, "postgres")
	if err != nil {
		t.Fatalf("admin database url: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	admin, err := pgxpool.New(ctx, adminDSN)
	if err != nil {
		t.Fatalf("connect admin database: %v", err)
	}

	name, err := testDatabaseName()
	if err != nil {
		admin.Close()
		t.Fatal(err)
	}
	if _, err := admin.Exec(ctx, `CREATE DATABASE "`+name+`"`); err != nil {
		admin.Close()
		t.Fatalf("create test database: %v", err)
	}

	var pool *pgxpool.Pool
	t.Cleanup(func() {
		if pool != nil {
			pool.Close()
		}
		dropCtx, dropCancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer dropCancel()
		if _, err := admin.Exec(dropCtx, `DROP DATABASE IF EXISTS "`+name+`" WITH (FORCE)`); err != nil {
			t.Errorf("drop test database %s: %v", name, err)
		}
		admin.Close()
	})

	dbURL, err := withDatabase(raw, name)
	if err != nil {
		t.Fatal(err)
	}
	m, err := migrate.New(migrationsURL(), dbURL)
	if err != nil {
		t.Fatalf("migrator: %v", err)
	}
	upErr := m.Up()
	srcErr, dbErr := m.Close()
	if upErr != nil && upErr != migrate.ErrNoChange {
		t.Fatalf("migrate: %v", upErr)
	}
	if srcErr != nil {
		t.Fatalf("close migration source: %v", srcErr)
	}
	if dbErr != nil {
		t.Fatalf("close migration database: %v", dbErr)
	}

	pool, err = pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Fatalf("connect test database: %v", err)
	}
	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("ping test database: %v", err)
	}
	return pool
}

func testDatabaseName() (string, error) {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	// Letter prefix, hex body. Safe to interpolate as an identifier.
	return "rk" + hex.EncodeToString(buf), nil
}

func withDatabase(raw, name string) (string, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", err
	}
	u.Path = "/" + name
	return u.String(), nil
}

func migrationsURL() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		panic("apitest: runtime.Caller failed")
	}
	dir := filepath.Dir(file)
	for i := 0; i < 6; i++ {
		cand := filepath.Join(dir, "migrations")
		if st, err := os.Stat(cand); err == nil && st.IsDir() {
			return "file://" + filepath.ToSlash(cand)
		}
		dir = filepath.Dir(dir)
	}
	panic("apitest: migrations directory not found")
}

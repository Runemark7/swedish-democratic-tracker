package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"riksdagskollen/internal/politicians"
	politiciansPG "riksdagskollen/internal/politicians/adapters/postgres"
	politiciansRD "riksdagskollen/internal/politicians/adapters/riksdagen"
	"riksdagskollen/internal/regions"
	koladaAdapter "riksdagskollen/internal/regions/adapters/kolada"
	regionsPG "riksdagskollen/internal/regions/adapters/postgres"
	scbAdapter "riksdagskollen/internal/regions/adapters/scb"
	tedAdapter "riksdagskollen/internal/regions/adapters/ted"
	speechesRD "riksdagskollen/internal/speeches/adapters/riksdagen"
	"riksdagskollen/internal/votes"
	votesPG "riksdagskollen/internal/votes/adapters/postgres"
	votesRD "riksdagskollen/internal/votes/adapters/riksdagen"
)

func main() {
	// Overall budget. The audit makes thousands of sequential SCB/Kolada
	// calls (the per-municipality KPI sweep dominates); whatever has not
	// completed when this elapses is recorded as a "context deadline" error
	// row rather than aborting the run. The Riksdag set runs first (below) so
	// it always finishes regardless of how long the region/kommun sweep takes.
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Minute)
	defer cancel()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		fmt.Fprintln(os.Stderr, "DATABASE_URL is required")
		os.Exit(2)
	}
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "db connect: %v\n", err)
		os.Exit(2)
	}
	defer pool.Close()

	// Construct services + upstream clients — mirrors cmd/api/main.go.
	regionsRepo := regionsPG.NewRepository(pool)
	koladaCl := koladaAdapter.NewClient()
	scbCl := scbAdapter.NewClient()
	tedCl := tedAdapter.NewClient()
	regionsSvc := regions.NewService(regionsRepo, koladaCl, scbCl, tedCl)

	polRepo := politiciansPG.NewRepository(pool)
	polCl := politiciansRD.NewClient()
	polSvc := politicians.NewService(polRepo, polCl)

	voteRepo := votesPG.NewRepository(pool)
	voteCl := votesRD.NewClient()
	voteSvc := votes.NewService(voteRepo, voteCl)

	speechCl := speechesRD.NewClient()

	// Run the small, fast Riksdag set first so it always completes. The
	// region/kommun sweep makes thousands of sequential SCB/Kolada calls and
	// can exhaust the whole budget; running it last keeps it from starving the
	// Riksdag checks. renderReport groups by tier, so order here only affects
	// which checks finish, not the report layout.
	var results []Result
	results = append(results, riksdagenChecks(ctx, polSvc, polCl, voteSvc, voteCl, speechCl)...)
	results = append(results, regionAndKommunChecks(ctx, regionsSvc, koladaCl)...)

	now := time.Now().UTC()
	doc := renderReport(results, now)
	outPath := filepath.Join("docs", "superpowers", "audits",
		fmt.Sprintf("%s-value-verification.md", now.Format("2006-01-02")))
	if err := os.WriteFile(outPath, []byte(doc), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "write report: %v\n", err)
		os.Exit(1)
	}

	var review int
	for _, r := range results {
		if r.Status == StatusReview {
			review++
		}
	}
	fmt.Printf("verifyvalues: %d checks, %d need review → %s\n", len(results), review, outPath)
}

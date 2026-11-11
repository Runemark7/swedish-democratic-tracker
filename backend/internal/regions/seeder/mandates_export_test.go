package seeder

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestFetchMandateTotals_LiveSmoke(t *testing.T) {
	if testing.Short() || os.Getenv("SCB_LIVE_SMOKE") == "" {
		t.Skip("hits live SCB (flaky since Nov 2026); set SCB_LIVE_SMOKE=1 to run")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	mun, reg, err := FetchMandateTotals(ctx, &http.Client{Timeout: 20 * time.Second})
	if err != nil {
		t.Fatalf("FetchMandateTotals: %v", err)
	}
	if len(reg) < 20 || len(reg) > 25 {
		t.Errorf("region count = %d, want ~21", len(reg))
	}
	if len(mun) < 250 {
		t.Errorf("municipality count = %d, want ~290", len(mun))
	}
	for name, total := range reg {
		if total <= 0 {
			t.Errorf("region %q total = %d", name, total)
		}
	}
}

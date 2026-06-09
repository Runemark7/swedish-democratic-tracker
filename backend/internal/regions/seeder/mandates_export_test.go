package seeder

import (
	"context"
	"net/http"
	"testing"
	"time"
)

func TestFetchMandateTotals_LiveSmoke(t *testing.T) {
	if testing.Short() {
		t.Skip("hits live SCB; run without -short")
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

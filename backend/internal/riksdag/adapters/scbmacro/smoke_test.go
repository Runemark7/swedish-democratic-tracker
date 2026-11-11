package scbmacro

import (
	"context"
	"os"
	"testing"
	"time"
)

// TestLiveSmoke calls SCB's live API. It stays opt-in so CI never depends on
// SCB being up (set SCB_LIVE_SMOKE=1). It checks the client end to end
// (v2 first, v1 fallback) and logs whether each series is served by v2 alone,
// so the v1 fallback can be dropped once v2 answers for both. PxWebApi v1 is
// switched off at the turn of 2026/2027.
func TestLiveSmoke(t *testing.T) {
	if testing.Short() || os.Getenv("SCB_LIVE_SMOKE") == "" {
		t.Skip("set SCB_LIVE_SMOKE=1 to run against the live SCB API")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	c := NewClient()
	for _, s := range []series{kpiSeries, akuSeries} {
		if pts, err := c.fetchV2(ctx, s); err != nil {
			t.Logf("%s: v2 %s failed (v1 fallback needed): %v", s.name, s.v2Table, err)
		} else {
			t.Logf("%s: v2 %s ok %+v", s.name, s.v2Table, pts)
		}
	}
	inf, err := c.FetchInflationRate(ctx)
	if err != nil {
		t.Fatalf("inflation: %v", err)
	}
	t.Logf("inflation latest=%+v", inf)
	un, err := c.FetchUnemploymentRate(ctx)
	if err != nil {
		t.Fatalf("unemployment: %v", err)
	}
	t.Logf("unemployment latest=%+v", un)
	if inf[0].Value < -5 || inf[0].Value > 20 {
		t.Errorf("inflation %v out of sane range", inf[0])
	}
	if un[0].Value < 1 || un[0].Value > 20 {
		t.Errorf("unemployment %v out of sane range", un[0])
	}
}

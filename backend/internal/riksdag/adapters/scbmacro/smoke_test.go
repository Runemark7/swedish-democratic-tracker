package scbmacro

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestLiveSmoke(t *testing.T) {
	// Opt-in: it calls SCB's live API, which CI must not depend on. Since
	// 2026-11-11 the v1 AKU table (AKURLBefK) answers 400 to our query, which
	// failed every main build; the fix of the client itself is tracked separately.
	if testing.Short() || os.Getenv("SCB_LIVE_SMOKE") == "" {
		t.Skip("set SCB_LIVE_SMOKE=1 to run against the live SCB API")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	c := NewClient()
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

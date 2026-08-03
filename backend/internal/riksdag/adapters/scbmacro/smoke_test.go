package scbmacro

import (
	"context"
	"testing"
	"time"
)

func TestLiveSmoke(t *testing.T) {
	if testing.Short() {
		t.Skip()
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

package budget

import (
	"context"
	"testing"
	"time"
)

func TestAcquireRelease(t *testing.T) {
	cfg := Config{
		Budgets: map[Connector]struct {
			Capacity int64         `yaml:"capacity"`
			Refill   int64         `yaml:"refill"`
			Period   time.Duration `yaml:"period"`
		}{
			GoogleText: {Capacity: 10, Refill: 10, Period: time.Second},
		},
		SplitRatio: 0.7,
	}
	g := NewGuard(cfg)
	g.Rebalance()

	ctx := context.Background()
	if err := g.Acquire(ctx, AcquireOpts{Connector: GoogleText, Split: Primaries, Tokens: 5, Deadline: time.Second}); err != nil {
		t.Fatalf("expected acquire ok, got %v", err)
	}
	g.Release(GoogleText, 3, Primaries)

	if err := g.Acquire(ctx, AcquireOpts{Connector: GoogleText, Split: Secondaries, Tokens: 2, Deadline: time.Second}); err != nil {
		t.Fatalf("expected acquire ok, got %v", err)
	}
}

func TestEarlyStop(t *testing.T) {
	pw := ProgressWindow{LastNNewUnique: 5, LastNCalls: 200}
	if !EarlyStop(pw, 0.05) {
		t.Fatalf("expected early stop")
	}
}

package workflow

import (
	"context"
	"time"
)

// Stats captures running counters for the current job.
type Stats struct {
	APICalls       int
	NewUniqueItems int
	TotalItemsSeen int
	// Optional: future fields (errors, retries, etc.)
}

// BudgetGuard encapsulates the stopping criteria.
type BudgetGuard struct {
	MaxAPICalls      int
	MaxWallClock     time.Duration
	MinNewUniqueRate float64 // e.g., 0.05 means 5% new items vs total seen

	StartTime time.Time
}

// ShouldStop decides whether to halt based on configured budgets.
// This is written to be deterministic and easily unit-testable.
func (b BudgetGuard) ShouldStop(ctx context.Context, s Stats) bool {
	// Guard 1: API budget
	if b.MaxAPICalls > 0 && s.APICalls >= b.MaxAPICalls {
		return true
	}

	// Guard 2: Wall clock
	if b.MaxWallClock > 0 && !b.StartTime.IsZero() && time.Since(b.StartTime) >= b.MaxWallClock {
		return true
	}

	// Guard 3: New unique rate
	if b.MinNewUniqueRate > 0 && s.TotalItemsSeen > 0 {
		rate := float64(s.NewUniqueItems) / float64(s.TotalItemsSeen)
		if rate < b.MinNewUniqueRate {
			return true
		}
	}

	return false
}

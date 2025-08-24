package main

import (
	"context"
	"fmt"
	"time"

	"github.com/Sreeram-ganesan/jaunt-data-scout/epics/orchestration-step-fns/go/internal/workflow"
)

func main() {
	ctx := context.Background()

	bg := workflow.BudgetGuard{
		MaxAPICalls:      1000,
		MaxWallClock:     30 * time.Minute,
		MinNewUniqueRate: 0.05,
		StartTime:        time.Now(),
	}
	stats := workflow.Stats{
		APICalls:       0,
		NewUniqueItems: 0,
		TotalItemsSeen: 0,
	}

	shouldStop := bg.ShouldStop(ctx, stats)
	fmt.Printf("Budget check (initial): shouldStop=%v\n", shouldStop)
}
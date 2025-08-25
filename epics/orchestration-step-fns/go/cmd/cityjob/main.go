package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	b "github.com/Sreeram-ganesan/jaunt-data-scout/epics/orchestration-step-fns/go/internal/budget"
	cfg "github.com/Sreeram-ganesan/jaunt-data-scout/epics/orchestration-step-fns/go/internal/config"
	wf "github.com/Sreeram-ganesan/jaunt-data-scout/epics/orchestration-step-fns/go/internal/workflow"
)

func main() {
	ctx := context.Background()

	// Load defaults
	defaultPath := filepath.Join("config", "defaults.yaml")
	if env := os.Getenv("CONFIG_PATH"); env != "" {
		defaultPath = env
	}
	rd, err := cfg.LoadDefaults(defaultPath)
	if err != nil {
		fmt.Printf("failed to load config: %v\n", err)
		return
	}
	cfg.ApplyEnvOverrides(&rd)
	bcfg := cfg.BuildBudgetConfig(rd)
	guard := b.NewGuard(bcfg)
	guard.Rebalance()

	// Example workflow guard wired from config
	bg := wf.BudgetGuard{
		MaxAPICalls:      rd.CityDefaults.Budgets.MaxAPICalls,
		MaxWallClock:     time.Duration(rd.CityDefaults.Budgets.MaxWallClockHours) * time.Hour,
		MinNewUniqueRate: rd.CityDefaults.EarlyStop.MinNewUniqueRate,
		StartTime:        time.Now(),
	}

	stats := wf.Stats{}
	shouldStop := bg.ShouldStop(ctx, stats)
	fmt.Printf("Loaded config: %s\n", rd.String())
	fmt.Printf("Budget check (initial): shouldStop=%v\n", shouldStop)
	_ = guard // placeholder to suppress unused; in real states Acquire() would be used per connector
}

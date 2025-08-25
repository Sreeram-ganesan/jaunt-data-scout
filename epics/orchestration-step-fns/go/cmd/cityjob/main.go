package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	b "github.com/Sreeram-ganesan/jaunt-data-scout/epics/orchestration-step-fns/go/internal/budget"
	cfg "github.com/Sreeram-ganesan/jaunt-data-scout/epics/orchestration-step-fns/go/internal/config"
	obs "github.com/Sreeram-ganesan/jaunt-data-scout/epics/orchestration-step-fns/go/internal/observability"
	wf "github.com/Sreeram-ganesan/jaunt-data-scout/epics/orchestration-step-fns/go/internal/workflow"
)

func main() {
	ctx := context.Background()

	// Ensure we have a correlation_id for this execution
	ctx = obs.EnsureCorrelationID(ctx)
	ctx = context.WithValue(ctx, "run_id", fmt.Sprintf("cityjob-run-%d", time.Now().Unix()))
	ctx = context.WithValue(ctx, "split", "primary")

	// Create logger with correlation_id
	logger := obs.LogWithCorrelationID(ctx, log.Default())

	logger.Printf("Starting cityjob execution")

	// Load defaults
	defaultPath := filepath.Join("config", "defaults.yaml")
	if env := os.Getenv("CONFIG_PATH"); env != "" {
		defaultPath = env
	}
	rd, err := cfg.LoadDefaults(defaultPath)
	if err != nil {
		logger.Printf("failed to load config: %v", err)
		return
	}
	cfg.ApplyEnvOverrides(&rd)
	bcfg := cfg.BuildBudgetConfig(rd)
	guard := b.NewGuard(bcfg)
	guard.Rebalance()

	logger.Printf("Loaded config successfully")

	// Example workflow guard wired from config
	bg := wf.BudgetGuard{
		MaxAPICalls:      rd.CityDefaults.Budgets.MaxAPICalls,
		MaxWallClock:     time.Duration(rd.CityDefaults.Budgets.MaxWallClockHours) * time.Hour,
		MinNewUniqueRate: rd.CityDefaults.EarlyStop.MinNewUniqueRate,
		StartTime:        time.Now(),
	}

	stats := wf.Stats{}
	shouldStop := bg.ShouldStop(ctx, stats)
	
	logger.Printf("Configuration loaded: %s", rd.String())
	logger.Printf("Budget check (initial): shouldStop=%v", shouldStop)

	// Demonstrate EMF metrics emission
	obs.CountCall(ctx, "cityjob", "initialize", "config", "edinburgh")
	obs.RecordDurationMS(ctx, "cityjob", "initialize", "config", "edinburgh", 50.0)
	
	logger.Printf("EMF metrics emitted for initialization")
	logger.Printf("Cityjob execution completed")
	
	_ = guard // placeholder to suppress unused; in real states Acquire() would be used per connector
}

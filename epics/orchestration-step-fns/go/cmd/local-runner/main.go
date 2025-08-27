package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	b "github.com/Sreeram-ganesan/jaunt-data-scout/epics/orchestration-step-fns/go/internal/budget"
	cfg "github.com/Sreeram-ganesan/jaunt-data-scout/epics/orchestration-step-fns/go/internal/config"
	obs "github.com/Sreeram-ganesan/jaunt-data-scout/epics/orchestration-step-fns/go/internal/observability"
	"github.com/Sreeram-ganesan/jaunt-data-scout/epics/orchestration-step-fns/go/internal/states"
	"github.com/Sreeram-ganesan/jaunt-data-scout/epics/orchestration-step-fns/go/internal/types"
	wf "github.com/Sreeram-ganesan/jaunt-data-scout/epics/orchestration-step-fns/go/internal/workflow"
)

// WorkflowInput represents the input configuration for a workflow run
type WorkflowInput struct {
	JobID     string                 `json:"job_id"`
	City      string                 `json:"city"`
	Seed      SeedConfig             `json:"seed"`
	Budgets   map[string]interface{} `json:"budgets"`
	Config    map[string]interface{} `json:"config,omitempty"`
	S3Prefix  string                 `json:"s3_prefix,omitempty"`
}

type SeedConfig struct {
	Type   string            `json:"type"`
	Center types.Coordinates `json:"center"`
	RadiusKM int             `json:"radius_km"`
}

// LocalWorkflowRunner implements the complete end-to-end workflow locally
type LocalWorkflowRunner struct {
	logger       *log.Logger
	budgetGuard  *b.Guard
	workflowGuard wf.BudgetGuard
	stateHandlers []states.StateHandler
	outputPath    string
}

// NewLocalWorkflowRunner creates a new local workflow runner
func NewLocalWorkflowRunner(configPath, outputPath string) (*LocalWorkflowRunner, error) {
	logger := log.New(os.Stdout, "[WorkflowRunner] ", log.LstdFlags)

	// Load configuration
	rd, err := cfg.LoadDefaults(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Apply environment overrides
	cfg.ApplyEnvOverrides(&rd)

	// Build budget configuration
	bcfg := cfg.BuildBudgetConfig(rd)
	guard := b.NewGuard(bcfg)
	guard.Rebalance()

	// Create workflow budget guard
	workflowGuard := wf.BudgetGuard{
		MaxAPICalls:      rd.CityDefaults.Budgets.MaxAPICalls,
		MaxWallClock:     time.Duration(rd.CityDefaults.Budgets.MaxWallClockHours) * time.Hour,
		MinNewUniqueRate: rd.CityDefaults.EarlyStop.MinNewUniqueRate,
		StartTime:        time.Now(),
	}

	// Initialize state handlers in order
	stateHandlers := []states.StateHandler{
		states.NewDiscoverWebSourcesHandler(),
		states.NewDiscoverTargetsHandler(),
		states.NewSeedPrimariesHandler(),
		states.NewExpandNeighborsHandler(),
		states.NewTileSweepHandler(),
		states.NewWebFetchHandler(),
		states.NewExtractWithLLMHandler(),
		states.NewGeocodeValidateHandler(),
		states.NewDedupeCanonicalizeHandler(),
		states.NewPersistHandler(),
		states.NewRankHandler(),
		states.NewFinalizeHandler(),
	}

	return &LocalWorkflowRunner{
		logger:        logger,
		budgetGuard:   guard,
		workflowGuard: workflowGuard,
		stateHandlers: stateHandlers,
		outputPath:    outputPath,
	}, nil
}

// RunWorkflow executes the complete workflow for a given input
func (r *LocalWorkflowRunner) RunWorkflow(ctx context.Context, input WorkflowInput) (*types.WorkflowResult, error) {
	startTime := time.Now()
	r.logger.Printf("Starting workflow for city: %s (job: %s)", input.City, input.JobID)

	// Ensure correlation ID in context
	ctx = obs.EnsureCorrelationID(ctx)
	correlationID := obs.FromContext(ctx)

	// Initialize workflow stats
	stats := wf.Stats{}
	allMetrics := make(map[string]states.StateMetrics)
	
	// Initialize state input
	stateInput := states.StateInput{
		City:          input.City,
		CorrelationID: correlationID,
		JobID:         input.JobID,
		Center:        input.Seed.Center,
		RadiusKM:      input.Seed.RadiusKM,
		Config:        input.Config,
	}

	var locations []types.Location
	var urls []states.WebSource
	sourcesUsed := make(map[string]bool)

	// Execute each state in sequence
	for i, handler := range r.stateHandlers {
		r.logger.Printf("Executing state %d/%d: %s", i+1, len(r.stateHandlers), handler.Name())

		// Check budget constraints before each state
		if r.workflowGuard.ShouldStop(ctx, stats) {
			r.logger.Printf("Budget constraints exceeded, stopping workflow")
			break
		}

		// Update state input with current data
		stateInput.Locations = locations
		stateInput.URLs = urls

		// Execute state
		output, err := handler.Execute(ctx, stateInput)
		if err != nil {
			return nil, fmt.Errorf("state %s failed: %w", handler.Name(), err)
		}

		// Update workflow state
		if len(output.Locations) > 0 {
			locations = output.Locations
		}
		if len(output.URLs) > 0 {
			urls = output.URLs
		}
		// Note: manifest is available in output.Manifest if needed

		// Update stats
		stats.APICalls += output.Metrics.APICalls
		stats.TotalItemsSeen += output.Metrics.ItemsFound
		stats.NewUniqueItems += output.Metrics.ItemsFound // Simplified for demo

		// Record metrics
		allMetrics[handler.Name()] = output.Metrics

		// Track sources used
		for _, loc := range locations {
			sourcesUsed[loc.Source] = true
		}
		for _, url := range urls {
			sourcesUsed[url.Source] = true
		}

		r.logger.Printf("State %s completed: found %d items, %d API calls, %v duration", 
			handler.Name(), output.Metrics.ItemsFound, output.Metrics.APICalls, output.Metrics.Duration)
	}

	// Calculate final summary
	var sourcesList []string
	for source := range sourcesUsed {
		sourcesList = append(sourcesList, source)
	}

	primaryCount := 0
	secondaryCount := 0
	for _, loc := range locations {
		if loc.Type == types.LocationTypePrimary {
			primaryCount++
		} else {
			secondaryCount++
		}
	}

	processingTime := time.Since(startTime)
	result := &types.WorkflowResult{
		JobID:         input.JobID,
		City:          input.City,
		CorrelationID: correlationID,
		CompletedAt:   time.Now(),
		Summary: types.Summary{
			TotalLocations:     len(locations),
			PrimaryLocations:   primaryCount,
			SecondaryLocations: secondaryCount,
			SourcesUsed:        sourcesList,
			ProcessingTimeMS:   processingTime.Milliseconds(),
			APICalls:           stats.APICalls,
		},
		Locations: locations,
	}

	r.logger.Printf("Workflow completed: %d locations (%d primary, %d secondary) in %v",
		len(locations), primaryCount, secondaryCount, processingTime)

	return result, nil
}

// WriteJSONL writes the workflow result to a JSONL file
func (r *LocalWorkflowRunner) WriteJSONL(result *types.WorkflowResult) error {
	// Ensure output directory exists
	if err := os.MkdirAll(filepath.Dir(r.outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	file, err := os.Create(r.outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)

	// Write summary line
	summaryLine := map[string]interface{}{
		"type":           "summary",
		"job_id":         result.JobID,
		"city":           result.City,
		"correlation_id": result.CorrelationID,
		"completed_at":   result.CompletedAt,
		"summary":        result.Summary,
	}
	
	if err := encoder.Encode(summaryLine); err != nil {
		return fmt.Errorf("failed to write summary: %w", err)
	}

	// Write each location as a separate line
	for _, location := range result.Locations {
		locationLine := map[string]interface{}{
			"type":     "location",
			"job_id":   result.JobID,
			"city":     result.City,
			"location": location,
		}
		
		if err := encoder.Encode(locationLine); err != nil {
			return fmt.Errorf("failed to write location: %w", err)
		}
	}

	r.logger.Printf("Wrote %d lines to %s", len(result.Locations)+1, r.outputPath)
	return nil
}

// loadEnvFile loads environment variables from a .env file
func loadEnvFile(filename string) error {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return nil // .env file is optional
	}

	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		
		// Remove quotes if present
		if (strings.HasPrefix(value, `"`) && strings.HasSuffix(value, `"`)) ||
		   (strings.HasPrefix(value, `'`) && strings.HasSuffix(value, `'`)) {
			value = value[1 : len(value)-1]
		}

		os.Setenv(key, value)
	}

	return scanner.Err()
}

func main() {
	ctx := context.Background()

	// Load .env file if it exists
	if err := loadEnvFile(".env"); err != nil {
		log.Printf("Warning: could not load .env file: %v", err)
	}

	// Get configuration from environment or use defaults
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config/defaults.yaml"
	}

	outputPath := os.Getenv("OUTPUT_PATH") 
	if outputPath == "" {
		outputPath = "output/edinburgh-results.jsonl"
	}

	inputFile := "examples/input.edinburgh.json"
	if len(os.Args) > 1 {
		inputFile = os.Args[1]
	}

	// Load input configuration
	inputData, err := os.ReadFile(inputFile)
	if err != nil {
		log.Fatalf("Failed to read input file %s: %v", inputFile, err)
	}

	var workflowInput WorkflowInput
	if err := json.Unmarshal(inputData, &workflowInput); err != nil {
		log.Fatalf("Failed to parse input file: %v", err)
	}

	// Create workflow runner
	runner, err := NewLocalWorkflowRunner(configPath, outputPath)
	if err != nil {
		log.Fatalf("Failed to create workflow runner: %v", err)
	}

	// Run the workflow
	result, err := runner.RunWorkflow(ctx, workflowInput)
	if err != nil {
		log.Fatalf("Workflow failed: %v", err)
	}

	// Write results to JSONL file
	if err := runner.WriteJSONL(result); err != nil {
		log.Fatalf("Failed to write output: %v", err)
	}

	log.Printf("Workflow completed successfully!")
	log.Printf("Results written to: %s", outputPath)
	log.Printf("Summary: %d locations (%d primary, %d secondary) from %v sources",
		result.Summary.TotalLocations,
		result.Summary.PrimaryLocations, 
		result.Summary.SecondaryLocations,
		result.Summary.SourcesUsed)
}
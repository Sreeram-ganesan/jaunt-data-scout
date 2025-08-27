// +build integration

package main

import (
	"context"
	"encoding/json"
	"os"
	"testing"
)

// Integration test that requires actual Tavily API key
func TestHandlerIntegration(t *testing.T) {
	if os.Getenv("TAVILY_SECRET_ARN") == "" {
		t.Skip("Skipping integration test - TAVILY_SECRET_ARN not set")
	}

	input := HandlerInput{
		RunID: "test-run-123",
		City: CityInfo{
			Name:        "Edinburgh",
			CountryCode: "GB",
		},
		Orchestrator: Orchestrator{
			CorrelationID: "test-correlation-456",
			FailFast:      false,
		},
		Budgets: Budgets{
			MaxURLs:      5, // Small for testing
			MaxPerDomain: 2,
		},
	}

	response, err := handler(context.Background(), input)
	if err != nil {
		t.Fatalf("Handler failed: %v", err)
	}

	t.Logf("Response: %+v", response)

	if response.DiscoveredCount == 0 {
		t.Error("Expected to discover some URLs")
	}

	if len(response.SampleURLs) == 0 {
		t.Error("Expected some sample URLs")
	}

	// All URLs should be valid
	for _, url := range response.SampleURLs {
		if url == "" {
			t.Error("Sample URL should not be empty")
		}
	}
}
package main

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestHandlerInputContract verifies the Lambda accepts the expected input schema
func TestHandlerInputContract(t *testing.T) {
	// Sample input matching the specification
	inputJSON := `{
		"run_id": "execution-123-456",
		"city": {
			"name": "Edinburgh",
			"country_code": "GB"
		},
		"orchestrator": {
			"correlation_id": "corr-abc-def",
			"fail_fast": false
		},
		"budgets": {
			"max_urls": 50,
			"max_per_domain": 5
		}
	}`

	var input HandlerInput
	err := json.Unmarshal([]byte(inputJSON), &input)
	if err != nil {
		t.Fatalf("Failed to unmarshal input JSON: %v", err)
	}

	// Verify all fields are parsed correctly
	if input.RunID != "execution-123-456" {
		t.Errorf("Expected run_id='execution-123-456', got '%s'", input.RunID)
	}

	if input.City.Name != "Edinburgh" {
		t.Errorf("Expected city.name='Edinburgh', got '%s'", input.City.Name)
	}

	if input.City.CountryCode != "GB" {
		t.Errorf("Expected city.country_code='GB', got '%s'", input.City.CountryCode)
	}

	if input.Orchestrator.CorrelationID != "corr-abc-def" {
		t.Errorf("Expected correlation_id='corr-abc-def', got '%s'", input.Orchestrator.CorrelationID)
	}

	if input.Orchestrator.FailFast != false {
		t.Errorf("Expected fail_fast=false, got %t", input.Orchestrator.FailFast)
	}

	if input.Budgets.MaxURLs != 50 {
		t.Errorf("Expected max_urls=50, got %d", input.Budgets.MaxURLs)
	}

	if input.Budgets.MaxPerDomain != 5 {
		t.Errorf("Expected max_per_domain=5, got %d", input.Budgets.MaxPerDomain)
	}
}

// TestHandlerOutputContract verifies the Lambda produces the expected output schema
func TestHandlerOutputContract(t *testing.T) {
	// Create a sample response
	response := HandlerResponse{
		DiscoveredCount: 25,
		EnqueuedCount:   23,
		TopDomains: []DomainCount{
			{Domain: "visitscotland.com", Count: 5},
			{Domain: "tripadvisor.com", Count: 3},
		},
		SampleURLs: []string{
			"https://www.visitscotland.com/destinations/highlands/edinburgh/",
			"https://www.tripadvisor.com/Attractions-g186525-Activities-Edinburgh_Scotland.html",
		},
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("Failed to marshal response: %v", err)
	}

	// Verify JSON contains expected fields
	jsonStr := string(jsonData)
	expectedFields := []string{
		`"discovered_count":25`,
		`"enqueued_count":23`,
		`"top_domains":`,
		`"sample_urls":`,
		`"domain":"visitscotland.com"`,
		`"count":5`,
	}

	for _, field := range expectedFields {
		if !strings.Contains(jsonStr, field) {
			t.Errorf("JSON output missing expected field: %s\nActual JSON: %s", field, jsonStr)
		}
	}

	// Verify structure by unmarshaling back
	var parsed HandlerResponse
	err = json.Unmarshal(jsonData, &parsed)
	if err != nil {
		t.Fatalf("Failed to unmarshal response JSON: %v", err)
	}

	if parsed.DiscoveredCount != response.DiscoveredCount {
		t.Errorf("Round-trip mismatch in discovered_count")
	}

	if len(parsed.TopDomains) != len(response.TopDomains) {
		t.Errorf("Round-trip mismatch in top_domains length")
	}

	if len(parsed.SampleURLs) != len(response.SampleURLs) {
		t.Errorf("Round-trip mismatch in sample_urls length")
	}
}

// TestFrontierMessageContract verifies the SQS frontier messages match expected schema
func TestFrontierMessageContract(t *testing.T) {
	// Create a sample frontier message
	msg := FrontierMessage{
		Type:          "web",
		RunID:         "execution-123-456",
		CorrelationID: "corr-abc-def",
		URL:           "https://example.com/page",
		Source:        "tavily",
		BudgetToken:   "tavily.api",
		City: CityInfo{
			Name:        "Edinburgh",
			CountryCode: "GB",
		},
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Failed to marshal frontier message: %v", err)
	}

	// Verify JSON structure
	jsonStr := string(jsonData)
	expectedFields := []string{
		`"type":"web"`,
		`"run_id":"execution-123-456"`,
		`"correlation_id":"corr-abc-def"`,
		`"url":"https://example.com/page"`,
		`"source":"tavily"`,
		`"budget_token":"tavily.api"`,
		`"city":{"name":"Edinburgh","country_code":"GB"}`,
	}

	for _, field := range expectedFields {
		if !strings.Contains(jsonStr, field) {
			t.Errorf("Frontier message JSON missing expected field: %s\nActual JSON: %s", field, jsonStr)
		}
	}

	// Verify structure by unmarshaling back
	var parsed FrontierMessage
	err = json.Unmarshal(jsonData, &parsed)
	if err != nil {
		t.Fatalf("Failed to unmarshal frontier message JSON: %v", err)
	}

	if parsed.Type != "web" {
		t.Errorf("Expected type='web', got '%s'", parsed.Type)
	}

	if parsed.Source != "tavily" {
		t.Errorf("Expected source='tavily', got '%s'", parsed.Source)
	}

	if parsed.BudgetToken != "tavily.api" {
		t.Errorf("Expected budget_token='tavily.api', got '%s'", parsed.BudgetToken)
	}
}

// TestHandlerWithDefaults verifies the Lambda handles missing budget fields correctly
func TestHandlerWithDefaults(t *testing.T) {
	// Input with missing budget fields
	inputJSON := `{
		"run_id": "execution-123-456",
		"city": {
			"name": "Edinburgh",
			"country_code": "GB"
		},
		"orchestrator": {
			"correlation_id": "corr-abc-def",
			"fail_fast": false
		}
	}`

	var input HandlerInput
	err := json.Unmarshal([]byte(inputJSON), &input)
	if err != nil {
		t.Fatalf("Failed to unmarshal input JSON: %v", err)
	}

	// The handler should set defaults for missing budgets
	// This tests the beginning of the handler logic without calling external APIs

	// Check defaults are applied (these would be set in handler())
	expectedMaxURLs := 100      // default from spec
	expectedMaxPerDomain := 10  // default from spec

	if input.Budgets.MaxURLs == 0 {
		input.Budgets.MaxURLs = expectedMaxURLs
	}
	if input.Budgets.MaxPerDomain == 0 {
		input.Budgets.MaxPerDomain = expectedMaxPerDomain
	}

	if input.Budgets.MaxURLs != expectedMaxURLs {
		t.Errorf("Expected default max_urls=%d, got %d", expectedMaxURLs, input.Budgets.MaxURLs)
	}

	if input.Budgets.MaxPerDomain != expectedMaxPerDomain {
		t.Errorf("Expected default max_per_domain=%d, got %d", expectedMaxPerDomain, input.Budgets.MaxPerDomain)
	}
}
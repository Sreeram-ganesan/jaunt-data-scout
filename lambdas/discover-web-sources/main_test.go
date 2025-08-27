package main

import (
	"strings"
	"testing"
)

func TestExtractDomain(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"https://example.com/path", "example.com"},
		{"http://subdomain.example.com", "subdomain.example.com"},
		{"https://example.com:8080/path", "example.com:8080"},
		{"invalid-url", ""},
	}

	for _, tt := range tests {
		result := extractDomain(tt.input)
		if result != tt.expected {
			t.Errorf("extractDomain(%s) = %s; want %s", tt.input, result, tt.expected)
		}
	}
}

func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			"https://example.com/path?utm_source=google&utm_medium=cpc&normal=keep#fragment",
			"https://example.com/path?normal=keep",
		},
		{
			"http://example.com/path?fbclid=123&gclid=456&keep=yes",
			"http://example.com/path?keep=yes",
		},
		{
			"example.com/path", 
			"https://example.com/path",
		},
	}

	for _, tt := range tests {
		result := normalizeURL(tt.input)
		if result != tt.expected {
			t.Errorf("normalizeURL(%s) = %s; want %s", tt.input, result, tt.expected)
		}
	}
}

func TestGenerateQueries(t *testing.T) {
	city := CityInfo{
		Name:        "Edinburgh",
		CountryCode: "GB",
	}

	queries := generateQueries(city)
	
	if len(queries) == 0 {
		t.Error("Expected non-empty queries")
	}

	// Check that city and country are properly substituted
	for _, query := range queries {
		if !containsWord(query, "Edinburgh") {
			t.Errorf("Query should contain city name 'Edinburgh': %s", query)
		}
		if !containsWord(query, "GB") {
			t.Errorf("Query should contain country code 'GB': %s", query)
		}
	}
}

func TestProcessResults(t *testing.T) {
	results := []DiscoveryResult{
		{URL: "https://example.com/1", Score: 0.9, Domain: "example.com"},
		{URL: "https://example.com/2", Score: 0.8, Domain: "example.com"},
		{URL: "https://example.com/3", Score: 0.7, Domain: "example.com"}, // Should be filtered by domain limit
		{URL: "https://other.com/1", Score: 0.6, Domain: "other.com"},
		{URL: "https://example.com/1", Score: 0.5, Domain: "example.com"}, // Duplicate
	}

	budgets := Budgets{
		MaxURLs:      10,
		MaxPerDomain: 2,
	}

	processed := processResults(results, budgets)

	// Check deduplication
	urlSet := make(map[string]bool)
	for _, result := range processed {
		if urlSet[result.URL] {
			t.Errorf("Duplicate URL found: %s", result.URL)
		}
		urlSet[result.URL] = true
	}

	// Check domain limit
	domainCounts := make(map[string]int)
	for _, result := range processed {
		domainCounts[result.Domain]++
		if domainCounts[result.Domain] > budgets.MaxPerDomain {
			t.Errorf("Domain %s exceeds limit of %d", result.Domain, budgets.MaxPerDomain)
		}
	}

	// Check ordering (should be by score descending)
	for i := 1; i < len(processed); i++ {
		if processed[i-1].Score < processed[i].Score {
			t.Error("Results should be ordered by score descending")
		}
	}
}

func TestBuildResponse(t *testing.T) {
	results := []DiscoveryResult{
		{URL: "https://example.com/1", Domain: "example.com"},
		{URL: "https://example.com/2", Domain: "example.com"},
		{URL: "https://other.com/1", Domain: "other.com"},
	}

	response := buildResponse(results, 3)

	if response.DiscoveredCount != 3 {
		t.Errorf("Expected discovered_count=3, got %d", response.DiscoveredCount)
	}

	if response.EnqueuedCount != 3 {
		t.Errorf("Expected enqueued_count=3, got %d", response.EnqueuedCount)
	}

	if len(response.TopDomains) == 0 {
		t.Error("Expected non-empty top_domains")
	}

	// Check that example.com has count 2
	found := false
	for _, domain := range response.TopDomains {
		if domain.Domain == "example.com" && domain.Count == 2 {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected example.com with count 2 in top_domains")
	}

	if len(response.SampleURLs) != 3 {
		t.Errorf("Expected 3 sample URLs, got %d", len(response.SampleURLs))
	}
}

// Helper function to check if a query contains a word
func containsWord(text, word string) bool {
	return strings.Contains(strings.ToLower(text), strings.ToLower(word))
}
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

// Input structures matching the contract
type HandlerInput struct {
	RunID        string      `json:"run_id"`
	City         CityInfo    `json:"city"`
	Orchestrator Orchestrator `json:"orchestrator"`
	Budgets      Budgets     `json:"budgets"`
}

type CityInfo struct {
	Name        string `json:"name"`
	CountryCode string `json:"country_code"`
}

type Orchestrator struct {
	CorrelationID string `json:"correlation_id"`
	FailFast      bool   `json:"fail_fast"`
}

type Budgets struct {
	MaxURLs      int `json:"max_urls"`
	MaxPerDomain int `json:"max_per_domain"`
}

// Output structure
type HandlerResponse struct {
	DiscoveredCount int               `json:"discovered_count"`
	EnqueuedCount   int               `json:"enqueued_count"`
	TopDomains      []DomainCount     `json:"top_domains"`
	SampleURLs      []string          `json:"sample_urls"`
}

type DomainCount struct {
	Domain string `json:"domain"`
	Count  int    `json:"count"`
}

// Tavily API structures
type TavilySearchRequest struct {
	Query            string   `json:"query"`
	MaxResults       int      `json:"max_results"`
	SearchDepth      string   `json:"search_depth"`
	IncludeAnswer    bool     `json:"include_answer"`
	IncludeImages    bool     `json:"include_images"`
	IncludeRawContent bool     `json:"include_raw_content"`
	IncludeDomains   []string `json:"include_domains,omitempty"`
	ExcludeDomains   []string `json:"exclude_domains,omitempty"`
}

type TavilySearchResponse struct {
	Query   string           `json:"query"`
	Results []TavilyResult   `json:"results"`
}

type TavilyResult struct {
	Title   string  `json:"title"`
	URL     string  `json:"url"`
	Content string  `json:"content"`
	Score   float64 `json:"score"`
}

// SQS Frontier message structure
type FrontierMessage struct {
	Type          string   `json:"type"`
	RunID         string   `json:"run_id"`
	CorrelationID string   `json:"correlation_id"`
	URL           string   `json:"url"`
	Source        string   `json:"source"`
	BudgetToken   string   `json:"budget_token"`
	City          CityInfo `json:"city"`
}

// Configuration
type Config struct {
	TavilySecretARN string
	FrontierQueueURL string
	ProjectPrefix   string
	Environment     string
}

// Discovery context for URL processing
type DiscoveryResult struct {
	URL    string
	Score  float64
	Domain string
	Title  string
}

var (
	cfg *Config
	awsCfg aws.Config
	secretsClient *secretsmanager.Client
	sqsClient *sqs.Client
	httpClient *http.Client
)

// Default query templates for city discovery
var queryTemplates = []string{
	"{city} {country} restaurants",
	"{city} {country} cafes coffee",
	"{city} {country} bars pubs nightlife",
	"{city} {country} things to do attractions",
	"{city} {country} events venues",
	"{city} {country} tourism tourist information",
	"{city} {country} official city website",
	"{city} {country} government site",
	"{city} {country} hotels accommodation",
	"{city} {country} museums galleries",
}

func init() {
	initializeLambda()
}

func initializeLambda() {
	// Initialize AWS config
	var err error
	awsCfg, err = config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("Failed to load AWS config: %v", err)
	}

	secretsClient = secretsmanager.NewFromConfig(awsCfg)
	sqsClient = sqs.NewFromConfig(awsCfg)
	
	// HTTP client with timeouts
	httpClient = &http.Client{
		Timeout: 30 * time.Second,
	}

	// Load configuration from environment variables
	cfg = &Config{
		TavilySecretARN:  getenv("TAVILY_SECRET_ARN", ""),
		FrontierQueueURL: getenv("FRONTIER_QUEUE_URL", ""),
		ProjectPrefix:    getenv("PROJECT_PREFIX", "jaunt"),
		Environment:      getenv("ENVIRONMENT", "dev"),
	}

	// Only validate required env vars if they're actually needed (not during tests)
	if os.Getenv("LAMBDA_RUNTIME_API") != "" {
		if cfg.TavilySecretARN == "" {
			log.Fatal("TAVILY_SECRET_ARN environment variable is required")
		}
		if cfg.FrontierQueueURL == "" {
			log.Fatal("FRONTIER_QUEUE_URL environment variable is required")
		}
	}
}

func main() {
	lambda.Start(handler)
}

func handler(ctx context.Context, input HandlerInput) (HandlerResponse, error) {
	// Set default budgets if not provided
	if input.Budgets.MaxURLs == 0 {
		input.Budgets.MaxURLs = 100
	}
	if input.Budgets.MaxPerDomain == 0 {
		input.Budgets.MaxPerDomain = 10
	}

	logStructured("discovery_started", map[string]interface{}{
		"run_id": input.RunID,
		"correlation_id": input.Orchestrator.CorrelationID,
		"city": input.City,
		"budgets": input.Budgets,
	})

	// Get Tavily API key
	tavilyKey, err := getTavilyAPIKey(ctx)
	if err != nil {
		emitMetric("tavily.errors", 1, input.RunID, input.Orchestrator.CorrelationID)
		return HandlerResponse{}, fmt.Errorf("failed to get Tavily API key: %w", err)
	}

	// Generate discovery queries
	queries := generateQueries(input.City)
	
	// Discover URLs using Tavily
	allResults := make([]DiscoveryResult, 0)
	for _, query := range queries {
		emitMetric("tavily.calls", 1, input.RunID, input.Orchestrator.CorrelationID)
		
		results, err := searchTavily(ctx, tavilyKey, query)
		if err != nil {
			log.Printf("Error searching Tavily for query '%s': %v", query, err)
			emitMetric("tavily.errors", 1, input.RunID, input.Orchestrator.CorrelationID)
			continue
		}
		
		for _, result := range results {
			allResults = append(allResults, DiscoveryResult{
				URL:    result.URL,
				Score:  result.Score,
				Domain: extractDomain(result.URL),
				Title:  result.Title,
			})
		}
	}

	// Deduplicate and apply budget constraints
	finalResults := processResults(allResults, input.Budgets)
	
	emitMetric("urls.discovered", float64(len(finalResults)), input.RunID, input.Orchestrator.CorrelationID)

	// Enqueue to SQS frontier
	enqueuedCount, err := enqueueFrontierMessages(ctx, finalResults, input)
	if err != nil {
		return HandlerResponse{}, fmt.Errorf("failed to enqueue frontier messages: %w", err)
	}

	emitMetric("urls.enqueued", float64(enqueuedCount), input.RunID, input.Orchestrator.CorrelationID)

	// Build response
	response := buildResponse(finalResults, enqueuedCount)
	
	logStructured("discovery_completed", map[string]interface{}{
		"run_id": input.RunID,
		"correlation_id": input.Orchestrator.CorrelationID,
		"discovered_count": response.DiscoveredCount,
		"enqueued_count": response.EnqueuedCount,
	})

	return response, nil
}

func getTavilyAPIKey(ctx context.Context) (string, error) {
	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(cfg.TavilySecretARN),
	}

	result, err := secretsClient.GetSecretValue(ctx, input)
	if err != nil {
		return "", err
	}

	// Parse the secret value (assuming it's JSON with an "api_key" field)
	var secret map[string]string
	if err := json.Unmarshal([]byte(*result.SecretString), &secret); err != nil {
		return "", fmt.Errorf("failed to parse secret: %w", err)
	}

	apiKey, ok := secret["api_key"]
	if !ok {
		return "", fmt.Errorf("api_key not found in secret")
	}

	return apiKey, nil
}

func generateQueries(city CityInfo) []string {
	queries := make([]string, 0, len(queryTemplates))
	for _, template := range queryTemplates {
		query := strings.ReplaceAll(template, "{city}", city.Name)
		query = strings.ReplaceAll(query, "{country}", city.CountryCode)
		queries = append(queries, query)
	}
	return queries
}

func searchTavily(ctx context.Context, apiKey, query string) ([]TavilyResult, error) {
	requestBody := TavilySearchRequest{
		Query:       query,
		MaxResults:  15,
		SearchDepth: "basic",
		IncludeAnswer: false,
		IncludeImages: false,
		IncludeRawContent: false,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.tavily.com/search", strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("User-Agent", cfg.ProjectPrefix+"-discover-web-sources/1.0")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Tavily API returned status %d", resp.StatusCode)
	}

	var response TavilySearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return response.Results, nil
}

func extractDomain(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	return parsed.Host
}

func normalizeURL(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	// Strip fragment
	parsed.Fragment = ""
	
	// Strip common tracking parameters
	values := parsed.Query()
	trackingParams := []string{"utm_source", "utm_medium", "utm_campaign", "utm_term", "utm_content", "fbclid", "gclid"}
	for _, param := range trackingParams {
		values.Del(param)
	}
	parsed.RawQuery = values.Encode()

	// Normalize scheme
	if parsed.Scheme == "" {
		parsed.Scheme = "https"
	}

	return parsed.String()
}

func processResults(results []DiscoveryResult, budgets Budgets) []DiscoveryResult {
	// Normalize URLs and deduplicate
	urlSeen := make(map[string]bool)
	domainCounts := make(map[string]int)
	dedupedResults := make([]DiscoveryResult, 0)

	for _, result := range results {
		normalizedURL := normalizeURL(result.URL)
		if urlSeen[normalizedURL] {
			continue
		}
		urlSeen[normalizedURL] = true

		domain := extractDomain(normalizedURL)
		if domainCounts[domain] >= budgets.MaxPerDomain {
			continue
		}

		result.URL = normalizedURL
		result.Domain = domain
		dedupedResults = append(dedupedResults, result)
		domainCounts[domain]++
	}

	// Sort by score (higher is better)
	sort.Slice(dedupedResults, func(i, j int) bool {
		return dedupedResults[i].Score > dedupedResults[j].Score
	})

	// Apply max URLs budget
	if len(dedupedResults) > budgets.MaxURLs {
		dedupedResults = dedupedResults[:budgets.MaxURLs]
	}

	return dedupedResults
}

func enqueueFrontierMessages(ctx context.Context, results []DiscoveryResult, input HandlerInput) (int, error) {
	if len(results) == 0 {
		return 0, nil
	}

	messages := make([]types.SendMessageBatchRequestEntry, 0, len(results))
	
	for i, result := range results {
		frontierMsg := FrontierMessage{
			Type:          "web",
			RunID:         input.RunID,
			CorrelationID: input.Orchestrator.CorrelationID,
			URL:           result.URL,
			Source:        "tavily",
			BudgetToken:   "tavily.api",
			City:          input.City,
		}

		msgBody, err := json.Marshal(frontierMsg)
		if err != nil {
			log.Printf("Failed to marshal frontier message: %v", err)
			continue
		}

		messages = append(messages, types.SendMessageBatchRequestEntry{
			Id:          aws.String(fmt.Sprintf("msg-%d", i)),
			MessageBody: aws.String(string(msgBody)),
		})
	}

	// Send in batches of 10 (SQS limit)
	batchSize := 10
	totalSent := 0
	
	for i := 0; i < len(messages); i += batchSize {
		end := i + batchSize
		if end > len(messages) {
			end = len(messages)
		}
		
		batch := messages[i:end]
		input := &sqs.SendMessageBatchInput{
			QueueUrl: aws.String(cfg.FrontierQueueURL),
			Entries:  batch,
		}

		output, err := sqsClient.SendMessageBatch(ctx, input)
		if err != nil {
			log.Printf("Failed to send message batch: %v", err)
			continue
		}

		totalSent += len(output.Successful)
		
		if len(output.Failed) > 0 {
			log.Printf("Failed to send %d messages in batch", len(output.Failed))
		}
	}

	return totalSent, nil
}

func buildResponse(results []DiscoveryResult, enqueuedCount int) HandlerResponse {
	// Count domains
	domainCounts := make(map[string]int)
	for _, result := range results {
		domainCounts[result.Domain]++
	}

	// Sort domains by count
	type domainCount struct {
		domain string
		count  int
	}
	var domains []domainCount
	for domain, count := range domainCounts {
		domains = append(domains, domainCount{domain: domain, count: count})
	}
	sort.Slice(domains, func(i, j int) bool {
		return domains[i].count > domains[j].count
	})

	// Top 10 domains
	topDomains := make([]DomainCount, 0, 10)
	for i, dc := range domains {
		if i >= 10 {
			break
		}
		topDomains = append(topDomains, DomainCount{
			Domain: dc.domain,
			Count:  dc.count,
		})
	}

	// Sample URLs (first 10)
	sampleURLs := make([]string, 0, 10)
	for i, result := range results {
		if i >= 10 {
			break
		}
		sampleURLs = append(sampleURLs, result.URL)
	}

	return HandlerResponse{
		DiscoveredCount: len(results),
		EnqueuedCount:   enqueuedCount,
		TopDomains:      topDomains,
		SampleURLs:      sampleURLs,
	}
}

func logStructured(event string, data map[string]interface{}) {
	logEntry := map[string]interface{}{
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"event":     event,
		"service":   "discover-web-sources",
	}
	for k, v := range data {
		logEntry[k] = v
	}

	jsonData, _ := json.Marshal(logEntry)
	fmt.Println(string(jsonData))
}

func emitMetric(metricName string, value float64, runID, correlationID string) {
	// EMF format for CloudWatch Metrics
	metric := map[string]interface{}{
		"_aws": map[string]interface{}{
			"Timestamp": time.Now().UnixMilli(),
			"CloudWatchMetrics": []map[string]interface{}{
				{
					"Namespace": fmt.Sprintf("%s/DiscoverWebSources", cfg.ProjectPrefix),
					"Dimensions": [][]string{
						{"Service", "Environment", "City"},
						{"Service", "Environment"},
						{"Service"},
					},
					"Metrics": []map[string]interface{}{
						{
							"Name": metricName,
							"Unit": "Count",
						},
					},
				},
			},
		},
		"Service":       "discover-web-sources",
		"Environment":   cfg.Environment,
		"City":          "unknown",
		"RunID":         runID,
		"CorrelationID": correlationID,
		metricName:      value,
	}

	jsonData, _ := json.Marshal(metric)
	fmt.Println(string(jsonData))
}

func getenv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
package states

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/Sreeram-ganesan/jaunt-data-scout/epics/orchestration-step-fns/go/internal/observability"
	"github.com/Sreeram-ganesan/jaunt-data-scout/epics/orchestration-step-fns/go/internal/types"
)

// StateHandler represents a single workflow state
type StateHandler interface {
	Name() string
	Execute(ctx context.Context, input StateInput) (StateOutput, error)
}

// StateInput contains the input for a state execution
type StateInput struct {
	City          string                 `json:"city"`
	CorrelationID string                 `json:"correlation_id"`
	JobID         string                 `json:"job_id"`
	Center        types.Coordinates      `json:"center"`
	RadiusKM      int                    `json:"radius_km"`
	Locations     []types.Location       `json:"locations,omitempty"`
	URLs          []WebSource            `json:"urls,omitempty"`
	Config        map[string]interface{} `json:"config,omitempty"`
}

// StateOutput contains the output of a state execution
type StateOutput struct {
	Locations []types.Location `json:"locations,omitempty"`
	URLs      []WebSource      `json:"urls,omitempty"`
	Manifest  *Manifest        `json:"manifest,omitempty"`
	Metrics   StateMetrics     `json:"metrics"`
}

type WebSource struct {
	URL        string   `json:"url"`
	Title      string   `json:"title,omitempty"`
	Domain     string   `json:"domain"`
	TrustScore *float64 `json:"trust_score,omitempty"`
	Source     string   `json:"source"`
}

type Manifest struct {
	QueryPackVersion string      `json:"query_pack_version"`
	Queries          []string    `json:"queries"`
	Candidates       []WebSource `json:"candidates"`
	Citations        []string    `json:"citations"`
}

type StateMetrics struct {
	Duration    time.Duration `json:"duration"`
	APICalls    int           `json:"api_calls"`
	ItemsFound  int           `json:"items_found"`
	TokensUsed  int64         `json:"tokens_used,omitempty"`
	BytesFetched int64        `json:"bytes_fetched,omitempty"`
}

// DiscoverWebSourcesHandler implements web source discovery using mock Tavily-like functionality
type DiscoverWebSourcesHandler struct {
	logger *log.Logger
}

func NewDiscoverWebSourcesHandler() *DiscoverWebSourcesHandler {
	return &DiscoverWebSourcesHandler{
		logger: log.New(os.Stdout, "[DiscoverWebSources] ", log.LstdFlags),
	}
}

func (h *DiscoverWebSourcesHandler) Name() string {
	return "DiscoverWebSources"
}

func (h *DiscoverWebSourcesHandler) Execute(ctx context.Context, input StateInput) (StateOutput, error) {
	start := time.Now()
	h.logger.Printf("Starting web source discovery for city: %s", input.City)

	observability.CountCall(ctx, h.Name(), "execute", "start", input.City)

	// Mock implementation - in production this would use Tavily API
	mockWebSources := h.generateMockWebSources(input.City)

	metrics := StateMetrics{
		Duration:   time.Since(start),
		APICalls:   len(mockWebSources) / 10, // Simulate API calls needed
		ItemsFound: len(mockWebSources),
	}

	observability.CountCall(ctx, h.Name(), "complete", "success", input.City)
	observability.RecordDurationMS(ctx, h.Name(), "execute", "total", input.City, float64(metrics.Duration.Milliseconds()))

	h.logger.Printf("Discovered %d web sources in %v", len(mockWebSources), metrics.Duration)

	return StateOutput{
		URLs:    mockWebSources,
		Metrics: metrics,
	}, nil
}

func (h *DiscoverWebSourcesHandler) generateMockWebSources(city string) []WebSource {
	sources := []WebSource{
		{
			URL:        fmt.Sprintf("https://visitscotland.com/%s", city),
			Title:      fmt.Sprintf("Visit %s - Official Tourism Site", city),
			Domain:     "visitscotland.com",
			TrustScore: &[]float64{0.95}[0],
			Source:     "tavily_mock",
		},
		{
			URL:        fmt.Sprintf("https://tripadvisor.com/attractions-%s", city),
			Title:      fmt.Sprintf("Top Attractions in %s", city),
			Domain:     "tripadvisor.com",
			TrustScore: &[]float64{0.85}[0],
			Source:     "tavily_mock",
		},
		{
			URL:        fmt.Sprintf("https://en.wikipedia.org/wiki/%s", city),
			Title:      fmt.Sprintf("%s - Wikipedia", city),
			Domain:     "wikipedia.org",
			TrustScore: &[]float64{0.90}[0],
			Source:     "tavily_mock",
		},
	}

	// Add some city-specific mock data sources
	if city == "Edinburgh" {
		sources = append(sources, []WebSource{
			{
				URL:        "https://data.gov.uk/dataset/edinburgh-attractions",
				Title:      "Edinburgh Attractions Open Data",
				Domain:     "data.gov.uk",
				TrustScore: &[]float64{0.88}[0],
				Source:     "tavily_mock",
			},
			{
				URL:        "https://www.edinburgh.gov.uk/business-licensing/places-entertainment",
				Title:      "Edinburgh Council - Places of Entertainment",
				Domain:     "edinburgh.gov.uk",
				TrustScore: &[]float64{0.92}[0],
				Source:     "tavily_mock",
			},
		}...)
	}

	return sources
}

// DiscoverTargetsHandler implements target discovery and manifest generation
type DiscoverTargetsHandler struct {
	logger *log.Logger
}

func NewDiscoverTargetsHandler() *DiscoverTargetsHandler {
	return &DiscoverTargetsHandler{
		logger: log.New(os.Stdout, "[DiscoverTargets] ", log.LstdFlags),
	}
}

func (h *DiscoverTargetsHandler) Name() string {
	return "DiscoverTargets"
}

func (h *DiscoverTargetsHandler) Execute(ctx context.Context, input StateInput) (StateOutput, error) {
	start := time.Now()
	h.logger.Printf("Discovering data targets for city: %s", input.City)

	observability.CountCall(ctx, h.Name(), "execute", "start", input.City)

	// Create manifest from web sources
	manifest := &Manifest{
		QueryPackVersion: "v1",
		Queries: []string{
			fmt.Sprintf("%s attractions", input.City),
			fmt.Sprintf("%s restaurants", input.City),
			fmt.Sprintf("%s hotels", input.City),
			fmt.Sprintf("%s museums", input.City),
			fmt.Sprintf("%s parks", input.City),
		},
		Candidates: input.URLs,
		Citations:  make([]string, len(input.URLs)),
	}

	for i, url := range input.URLs {
		manifest.Citations[i] = fmt.Sprintf("Retrieved from %s on %s", url.URL, time.Now().Format("2006-01-02"))
	}

	metrics := StateMetrics{
		Duration:   time.Since(start),
		APICalls:   5, // Mock API calls for queries
		ItemsFound: len(input.URLs),
	}

	observability.CountCall(ctx, h.Name(), "complete", "success", input.City)
	observability.RecordDurationMS(ctx, h.Name(), "execute", "total", input.City, float64(metrics.Duration.Milliseconds()))

	h.logger.Printf("Generated manifest with %d candidates in %v", len(manifest.Candidates), metrics.Duration)

	return StateOutput{
		Manifest: manifest,
		URLs:     input.URLs,
		Metrics:  metrics,
	}, nil
}

// SeedPrimariesHandler implements primary location seeding
type SeedPrimariesHandler struct {
	logger *log.Logger
}

func NewSeedPrimariesHandler() *SeedPrimariesHandler {
	return &SeedPrimariesHandler{
		logger: log.New(os.Stdout, "[SeedPrimaries] ", log.LstdFlags),
	}
}

func (h *SeedPrimariesHandler) Name() string {
	return "SeedPrimaries"
}

func (h *SeedPrimariesHandler) Execute(ctx context.Context, input StateInput) (StateOutput, error) {
	start := time.Now()
	h.logger.Printf("Seeding primary locations for city: %s", input.City)

	observability.CountCall(ctx, h.Name(), "execute", "start", input.City)

	// Generate mock primary locations
	primaries := h.generateMockPrimaryLocations(input.City, input.Center, input.CorrelationID)

	metrics := StateMetrics{
		Duration:   time.Since(start),
		APICalls:   len(primaries), // Simulate API calls to fetch details
		ItemsFound: len(primaries),
	}

	observability.CountCall(ctx, h.Name(), "complete", "success", input.City)
	observability.RecordDurationMS(ctx, h.Name(), "execute", "total", input.City, float64(metrics.Duration.Milliseconds()))

	h.logger.Printf("Seeded %d primary locations in %v", len(primaries), metrics.Duration)

	return StateOutput{
		Locations: primaries,
		Metrics:   metrics,
	}, nil
}

func (h *SeedPrimariesHandler) generateMockPrimaryLocations(city string, center types.Coordinates, correlationID string) []types.Location {
	locations := []types.Location{}
	
	// Create some mock primary attractions for Edinburgh
	if city == "Edinburgh" {
		mockPrimaries := []struct {
			name     string
			category string
			lat, lng float64
			rating   float64
		}{
			{"Edinburgh Castle", "historic_site", 55.9486, -3.1999, 4.5},
			{"Royal Mile", "tourist_attraction", 55.9507, -3.1844, 4.3},
			{"Palace of Holyroodhouse", "historic_site", 55.9529, -3.1722, 4.2},
			{"Arthur's Seat", "natural_feature", 55.9444, -3.1618, 4.4},
			{"National Museum of Scotland", "museum", 55.9469, -3.1901, 4.6},
			{"Camera Obscura", "tourist_attraction", 55.9487, -3.1970, 4.1},
			{"St Giles' Cathedral", "church", 55.9496, -3.1906, 4.3},
			{"Princes Street Gardens", "park", 55.9518, -3.2020, 4.2},
		}

		for i, mock := range mockPrimaries {
			locations = append(locations, types.Location{
				ID:   fmt.Sprintf("primary_%d", i+1),
				Name: mock.name,
				Type: types.LocationTypePrimary,
				Coordinates: types.Coordinates{
					Lat: mock.lat,
					Lng: mock.lng,
				},
				Address: types.Address{
					FormattedAddress: fmt.Sprintf("%s, Edinburgh, Scotland", mock.name),
					Locality:         "Edinburgh",
					Region:          "Scotland",
					Country:         "United Kingdom",
				},
				Category:              mock.category,
				Rating:                &mock.rating,
				Source:                "google_mock",
				SourceID:              fmt.Sprintf("google_place_%d", i+1),
				Confidence:            0.95,
				CreatedAt:            time.Now(),
				CorrelationID:        correlationID,
				CoordinatesConfidence: 0.98,
				AdditionalContent: map[string]interface{}{
					"google": map[string]interface{}{
						"place_id":           fmt.Sprintf("ChIJ%d", rand.Int31()),
						"types":              []string{mock.category, "establishment", "point_of_interest"},
						"business_status":    "OPERATIONAL",
						"user_ratings_total": rand.Intn(1000) + 100,
					},
					"signals": map[string]interface{}{
						"has_wikipedia":    true,
						"photo_count":      rand.Intn(50) + 10,
						"novelty":          0.8,
					},
				},
			})
		}
	}

	return locations
}
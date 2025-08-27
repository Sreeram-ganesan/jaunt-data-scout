package states

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"time"

	"github.com/Sreeram-ganesan/jaunt-data-scout/epics/orchestration-step-fns/go/internal/observability"
	"github.com/Sreeram-ganesan/jaunt-data-scout/epics/orchestration-step-fns/go/internal/types"
)

// ExpandNeighborsHandler implements neighbor expansion around primaries
type ExpandNeighborsHandler struct {
	logger *log.Logger
}

func NewExpandNeighborsHandler() *ExpandNeighborsHandler {
	return &ExpandNeighborsHandler{
		logger: log.New(os.Stdout, "[ExpandNeighbors] ", log.LstdFlags),
	}
}

func (h *ExpandNeighborsHandler) Name() string {
	return "ExpandNeighbors"
}

func (h *ExpandNeighborsHandler) Execute(ctx context.Context, input StateInput) (StateOutput, error) {
	start := time.Now()
	h.logger.Printf("Expanding neighbors around %d primary locations", len(input.Locations))

	observability.CountCall(ctx, h.Name(), "execute", "start", input.City)

	// Generate secondary locations around each primary
	var allLocations []types.Location
	allLocations = append(allLocations, input.Locations...) // Keep primaries

	secondaries := h.generateSecondaryLocations(input.Locations, input.CorrelationID)
	allLocations = append(allLocations, secondaries...)

	metrics := StateMetrics{
		Duration:   time.Since(start),
		APICalls:   len(secondaries) / 5, // Simulate batch API calls
		ItemsFound: len(secondaries),
	}

	observability.CountCall(ctx, h.Name(), "complete", "success", input.City)
	observability.RecordDurationMS(ctx, h.Name(), "execute", "total", input.City, float64(metrics.Duration.Milliseconds()))

	h.logger.Printf("Generated %d secondary locations in %v", len(secondaries), metrics.Duration)

	return StateOutput{
		Locations: allLocations,
		Metrics:   metrics,
	}, nil
}

func (h *ExpandNeighborsHandler) generateSecondaryLocations(primaries []types.Location, correlationID string) []types.Location {
	var secondaries []types.Location
	
	for i, primary := range primaries {
		// Generate 10-20 secondary locations around each primary
		numSecondaries := 10 + rand.Intn(11) // 10-20
		
		for j := 0; j < numSecondaries; j++ {
			// Generate coordinates within ~800m of the primary
			lat := primary.Coordinates.Lat + (rand.Float64()-0.5)*0.01 // ~±500m
			lng := primary.Coordinates.Lng + (rand.Float64()-0.5)*0.01
			
			secondary := types.Location{
				ID:   fmt.Sprintf("secondary_%d_%d", i+1, j+1),
				Name: h.generateSecondaryName(primary.Category, j),
				Type: types.LocationTypeSecondary,
				Coordinates: types.Coordinates{
					Lat: lat,
					Lng: lng,
				},
				Address: types.Address{
					FormattedAddress: fmt.Sprintf("Near %s, Edinburgh, Scotland", primary.Name),
					Locality:         "Edinburgh",
					Region:          "Scotland",
					Country:         "United Kingdom",
				},
				Category:              h.getSecondaryCategory(primary.Category),
				Rating:                h.generateSecondaryRating(),
				Source:                "google_nearby_mock",
				SourceID:              fmt.Sprintf("google_nearby_%d_%d", i+1, j+1),
				Confidence:            0.75 + rand.Float64()*0.20, // 0.75-0.95
				CreatedAt:            time.Now(),
				CorrelationID:        correlationID,
				CoordinatesConfidence: 0.85 + rand.Float64()*0.10, // 0.85-0.95
				AdditionalContent: map[string]interface{}{
					"google": map[string]interface{}{
						"place_id":           fmt.Sprintf("ChIJ%d", rand.Int31()),
						"types":              []string{h.getSecondaryCategory(primary.Category), "establishment"},
						"business_status":    "OPERATIONAL",
						"user_ratings_total": rand.Intn(200) + 10,
					},
					"signals": map[string]interface{}{
						"anchor_primary": primary.ID,
						"distance_m":     rand.Intn(800) + 50, // 50-850m
						"photo_count":    rand.Intn(20) + 1,
					},
				},
			}
			
			// Calculate adjacency score
			adjacencyScore := 0.6*0.8 + 0.2*rand.Float64() + 0.2*0.7 // proximity + novelty + coverage
			secondary.AdjacencyScore = &adjacencyScore
			
			secondaries = append(secondaries, secondary)
		}
	}
	
	return secondaries
}

func (h *ExpandNeighborsHandler) generateSecondaryName(primaryCategory string, index int) string {
	names := map[string][]string{
		"historic_site": {"Historic Inn", "Heritage Cafe", "Old Town Shop", "Medieval Restaurant", "Castle View Hotel"},
		"tourist_attraction": {"Tourist Info Center", "Souvenir Shop", "City Tour Office", "Local Gallery", "Visitor Cafe"},
		"natural_feature": {"Mountain View Cafe", "Hiking Gear Store", "Nature Center", "Outdoor Shop", "Trail Restaurant"},
		"museum": {"Museum Shop", "Cultural Cafe", "Art Gallery", "Learning Center", "Heritage Store"},
		"church": {"Parish Hall", "Religious Shop", "Pilgrimage Center", "Sacred Cafe", "Faith Bookstore"},
		"park": {"Park Cafe", "Garden Center", "Recreation Store", "Picnic Shop", "Nature Gift Shop"},
	}
	
	if nameList, exists := names[primaryCategory]; exists && index < len(nameList) {
		return fmt.Sprintf("%s %d", nameList[index%len(nameList)], index+1)
	}
	
	return fmt.Sprintf("Local Business %d", index+1)
}

func (h *ExpandNeighborsHandler) getSecondaryCategory(primaryCategory string) string {
	categoryMap := map[string]string{
		"historic_site":       "restaurant",
		"tourist_attraction":  "store",
		"natural_feature":     "cafe",
		"museum":             "gift_shop",
		"church":             "community_center", 
		"park":               "recreation",
	}
	
	if category, exists := categoryMap[primaryCategory]; exists {
		return category
	}
	return "establishment"
}

func (h *ExpandNeighborsHandler) generateSecondaryRating() *float64 {
	rating := 3.0 + rand.Float64()*2.0 // 3.0-5.0
	return &rating
}

// TileSweepHandler implements H3 tile sweep for coverage
type TileSweepHandler struct {
	logger *log.Logger
}

func NewTileSweepHandler() *TileSweepHandler {
	return &TileSweepHandler{
		logger: log.New(os.Stdout, "[TileSweep] ", log.LstdFlags),
	}
}

func (h *TileSweepHandler) Name() string {
	return "TileSweep"
}

func (h *TileSweepHandler) Execute(ctx context.Context, input StateInput) (StateOutput, error) {
	start := time.Now()
	h.logger.Printf("Performing H3 tile sweep for under-dense areas")

	observability.CountCall(ctx, h.Name(), "execute", "start", input.City)

	// Mock H3 tile sweep - in production this would use actual H3 library
	additionalLocations := h.performMockTileSweep(input.Center, input.RadiusKM, input.CorrelationID)

	allLocations := append(input.Locations, additionalLocations...)

	metrics := StateMetrics{
		Duration:   time.Since(start),
		APICalls:   len(additionalLocations) / 10, // Simulate batch API calls
		ItemsFound: len(additionalLocations),
	}

	observability.CountCall(ctx, h.Name(), "complete", "success", input.City)
	observability.RecordDurationMS(ctx, h.Name(), "execute", "total", input.City, float64(metrics.Duration.Milliseconds()))

	h.logger.Printf("Added %d locations from tile sweep in %v", len(additionalLocations), metrics.Duration)

	return StateOutput{
		Locations: allLocations,
		Metrics:   metrics,
	}, nil
}

func (h *TileSweepHandler) performMockTileSweep(center types.Coordinates, radiusKM int, correlationID string) []types.Location {
	var locations []types.Location
	
	// Generate some locations in under-dense areas (mock H3 cells)
	for i := 0; i < 20; i++ { // Mock 20 additional locations from tile sweep
		// Generate coordinates within the radius but in "sparse" areas
		angle := rand.Float64() * 2 * math.Pi
		distance := rand.Float64() * float64(radiusKM) * 0.009 // Rough conversion to degrees
		
		lat := center.Lat + distance*math.Cos(angle)
		lng := center.Lng + distance*math.Sin(angle)
		
		location := types.Location{
			ID:   fmt.Sprintf("tile_sweep_%d", i+1),
			Name: fmt.Sprintf("Area Location %d", i+1),
			Type: types.LocationTypeSecondary,
			Coordinates: types.Coordinates{
				Lat: lat,
				Lng: lng,
			},
			Address: types.Address{
				FormattedAddress: fmt.Sprintf("Edinburgh Area %d, Edinburgh, Scotland", i+1),
				Locality:         "Edinburgh", 
				Region:          "Scotland",
				Country:         "United Kingdom",
			},
			Category:              "local_business",
			Rating:                h.generateTileSweepRating(),
			Source:                "osm_mock",
			SourceID:              fmt.Sprintf("osm_way_%d", rand.Int31()),
			Confidence:            0.60 + rand.Float64()*0.25, // 0.60-0.85
			CreatedAt:            time.Now(),
			CorrelationID:        correlationID,
			CoordinatesConfidence: 0.80 + rand.Float64()*0.15, // 0.80-0.95
			AdditionalContent: map[string]interface{}{
				"osm": map[string]interface{}{
					"way_id":      rand.Int31(),
					"tags":        map[string]string{"amenity": "restaurant", "name": fmt.Sprintf("Area Location %d", i+1)},
					"h3_index_9":  fmt.Sprintf("h3_9_%x", rand.Int31()),
					"h3_index_12": fmt.Sprintf("h3_12_%x", rand.Int31()),
				},
				"signals": map[string]interface{}{
					"source":     "tile_sweep",
					"coverage":   true,
					"sparse_area": true,
				},
			},
		}
		
		locations = append(locations, location)
	}
	
	return locations
}

func (h *TileSweepHandler) generateTileSweepRating() *float64 {
	// Tile sweep locations might have lower ratings on average
	rating := 2.5 + rand.Float64()*2.0 // 2.5-4.5
	return &rating
}

// WebFetchHandler implements web content fetching
type WebFetchHandler struct {
	logger *log.Logger
}

func NewWebFetchHandler() *WebFetchHandler {
	return &WebFetchHandler{
		logger: log.New(os.Stdout, "[WebFetch] ", log.LstdFlags),
	}
}

func (h *WebFetchHandler) Name() string {
	return "WebFetch"
}

func (h *WebFetchHandler) Execute(ctx context.Context, input StateInput) (StateOutput, error) {
	start := time.Now()
	h.logger.Printf("Fetching web content from %d URLs", len(input.URLs))

	observability.CountCall(ctx, h.Name(), "execute", "start", input.City)

	// Mock web fetching - simulate downloading and caching content
	fetchedContent := h.mockWebFetch(input.URLs)

	metrics := StateMetrics{
		Duration:     time.Since(start),
		APICalls:     0, // No API calls, just HTTP fetches
		ItemsFound:   len(fetchedContent),
		BytesFetched: int64(len(fetchedContent) * 50000), // Simulate ~50KB per page
	}

	observability.CountCall(ctx, h.Name(), "complete", "success", input.City)
	observability.RecordDurationMS(ctx, h.Name(), "execute", "total", input.City, float64(metrics.Duration.Milliseconds()))

	h.logger.Printf("Fetched %d web pages (%d bytes) in %v", len(fetchedContent), metrics.BytesFetched, metrics.Duration)

	return StateOutput{
		URLs:     fetchedContent,
		Metrics:  metrics,
		Locations: input.Locations, // Pass through existing locations
	}, nil
}

func (h *WebFetchHandler) mockWebFetch(urls []WebSource) []WebSource {
	// Simulate successful fetches with cached content
	for i := range urls {
		// Mock: add cache metadata
		if urls[i].Source == "" {
			urls[i].Source = "web_fetch_mock"
		}
	}
	return urls
}
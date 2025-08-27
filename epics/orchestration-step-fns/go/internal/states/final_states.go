package states

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/Sreeram-ganesan/jaunt-data-scout/epics/orchestration-step-fns/go/internal/observability"
	"github.com/Sreeram-ganesan/jaunt-data-scout/epics/orchestration-step-fns/go/internal/types"
)

// ExtractWithLLMHandler implements LLM-based content extraction
type ExtractWithLLMHandler struct {
	logger *log.Logger
}

func NewExtractWithLLMHandler() *ExtractWithLLMHandler {
	return &ExtractWithLLMHandler{
		logger: log.New(os.Stdout, "[ExtractWithLLM] ", log.LstdFlags),
	}
}

func (h *ExtractWithLLMHandler) Name() string {
	return "ExtractWithLLM"
}

func (h *ExtractWithLLMHandler) Execute(ctx context.Context, input StateInput) (StateOutput, error) {
	start := time.Now()
	h.logger.Printf("Extracting structured data from %d web sources using LLM", len(input.URLs))

	observability.CountCall(ctx, h.Name(), "execute", "start", input.City)

	// Mock LLM extraction - simulate structured extraction from web content
	extractedLocations := h.mockLLMExtraction(input.URLs, input.CorrelationID)

	// Combine with existing locations
	allLocations := append(input.Locations, extractedLocations...)

	// Mock token usage
	tokensUsed := int64(len(input.URLs) * 1500) // ~1500 tokens per URL processed

	metrics := StateMetrics{
		Duration:   time.Since(start),
		APICalls:   len(input.URLs), // One LLM call per URL
		ItemsFound: len(extractedLocations),
		TokensUsed: tokensUsed,
	}

	observability.CountCall(ctx, h.Name(), "complete", "success", input.City)
	observability.RecordDurationMS(ctx, h.Name(), "execute", "total", input.City, float64(metrics.Duration.Milliseconds()))

	h.logger.Printf("Extracted %d locations using %d tokens in %v", len(extractedLocations), tokensUsed, metrics.Duration)

	return StateOutput{
		Locations: allLocations,
		Metrics:   metrics,
	}, nil
}

func (h *ExtractWithLLMHandler) mockLLMExtraction(urls []WebSource, correlationID string) []types.Location {
	var locations []types.Location

	for i, url := range urls {
		// Mock: extract 1-3 locations per URL
		numExtracted := 1 + rand.Intn(3)

		for j := 0; j < numExtracted; j++ {
			// Generate Edinburgh coordinates (mock extraction)
			lat := 55.9533 + (rand.Float64()-0.5)*0.05 // ±2.5km around center
			lng := -3.1883 + (rand.Float64()-0.5)*0.05

			location := types.Location{
				ID:   fmt.Sprintf("llm_extracted_%d_%d", i+1, j+1),
				Name: h.generateExtractedLocationName(url.Domain, j),
				Type: types.LocationTypeSecondary,
				Coordinates: types.Coordinates{
					Lat: lat,
					Lng: lng,
				},
				Address: types.Address{
					FormattedAddress: fmt.Sprintf("Extracted Location %d, Edinburgh, Scotland", j+1),
					Locality:         "Edinburgh",
					Region:          "Scotland",
					Country:         "United Kingdom",
				},
				Category:              h.inferCategoryFromURL(url.URL),
				Rating:                h.generateLLMExtractedRating(),
				Source:                "llm_extracted",
				SourceID:              fmt.Sprintf("llm_%d_%d", i+1, j+1),
				Confidence:            0.65 + rand.Float64()*0.25, // 0.65-0.90
				CreatedAt:            time.Now(),
				CorrelationID:        correlationID,
				CoordinatesConfidence: 0.60 + rand.Float64()*0.30, // 0.60-0.90 (lower for LLM)
				AdditionalContent: map[string]interface{}{
					"web": map[string]interface{}{
						"source_url":    url.URL,
						"source_domain": url.Domain,
						"trust_score":   url.TrustScore,
						"extraction_confidences": map[string]float64{
							"name":        0.85 + rand.Float64()*0.10,
							"coordinates": 0.60 + rand.Float64()*0.30,
							"category":    0.70 + rand.Float64()*0.20,
						},
					},
					"signals": map[string]interface{}{
						"llm_extracted":   true,
						"needs_geocoding": rand.Float64() < 0.3, // 30% might need geocoding
						"novelty":         0.7,
					},
				},
			}

			locations = append(locations, location)
		}
	}

	return locations
}

func (h *ExtractWithLLMHandler) generateExtractedLocationName(domain string, index int) string {
	domainNames := map[string][]string{
		"visitscotland.com":  {"Highland Restaurant", "Scottish Heritage Site", "Celtic Museum", "Royal Gardens"},
		"tripadvisor.com":    {"Top Rated Cafe", "Tourist Favorite", "Must Visit Pub", "Hidden Gem Store"},
		"wikipedia.org":      {"Historical Landmark", "Cultural Site", "Notable Building", "Memorial Garden"},
		"data.gov.uk":        {"Public Facility", "Community Center", "Government Building", "Public Park"},
		"edinburgh.gov.uk":   {"City Service", "Municipal Building", "Public Library", "Recreation Center"},
	}

	if names, exists := domainNames[domain]; exists && index < len(names) {
		return fmt.Sprintf("%s %d", names[index%len(names)], index+1)
	}

	return fmt.Sprintf("Web Extracted Location %d", index+1)
}

func (h *ExtractWithLLMHandler) inferCategoryFromURL(url string) string {
	if strings.Contains(url, "attraction") {
		return "tourist_attraction"
	} else if strings.Contains(url, "restaurant") || strings.Contains(url, "dining") {
		return "restaurant"
	} else if strings.Contains(url, "museum") {
		return "museum"
	} else if strings.Contains(url, "park") || strings.Contains(url, "garden") {
		return "park"
	} else if strings.Contains(url, "gov") {
		return "government"
	}
	return "establishment"
}

func (h *ExtractWithLLMHandler) generateLLMExtractedRating() *float64 {
	// LLM extracted ratings might be less reliable
	rating := 3.0 + rand.Float64()*1.8 // 3.0-4.8
	return &rating
}

// GeocodeValidateHandler implements geocoding validation
type GeocodeValidateHandler struct {
	logger *log.Logger
}

func NewGeocodeValidateHandler() *GeocodeValidateHandler {
	return &GeocodeValidateHandler{
		logger: log.New(os.Stdout, "[GeocodeValidate] ", log.LstdFlags),
	}
}

func (h *GeocodeValidateHandler) Name() string {
	return "GeocodeValidate"
}

func (h *GeocodeValidateHandler) Execute(ctx context.Context, input StateInput) (StateOutput, error) {
	start := time.Now()
	h.logger.Printf("Validating and geocoding %d locations", len(input.Locations))

	observability.CountCall(ctx, h.Name(), "execute", "start", input.City)

	// Process locations for geocoding validation
	validatedLocations, geocodingCalls := h.validateAndGeocodeLocations(input.Locations)

	metrics := StateMetrics{
		Duration:   time.Since(start),
		APICalls:   geocodingCalls,
		ItemsFound: len(validatedLocations),
	}

	observability.CountCall(ctx, h.Name(), "complete", "success", input.City)
	observability.RecordDurationMS(ctx, h.Name(), "execute", "total", input.City, float64(metrics.Duration.Milliseconds()))

	h.logger.Printf("Validated %d locations with %d geocoding calls in %v", len(validatedLocations), geocodingCalls, metrics.Duration)

	return StateOutput{
		Locations: validatedLocations,
		Metrics:   metrics,
	}, nil
}

func (h *GeocodeValidateHandler) validateAndGeocodeLocations(locations []types.Location) ([]types.Location, int) {
	geocodingCalls := 0

	for i := range locations {
		// Check if location needs geocoding (low coordinates confidence)
		if locations[i].CoordinatesConfidence < 0.75 {
			// Mock geocoding - improve coordinates and confidence
			locations[i].Coordinates.Lat += (rand.Float64() - 0.5) * 0.001 // Small adjustment
			locations[i].Coordinates.Lng += (rand.Float64() - 0.5) * 0.001
			locations[i].CoordinatesConfidence = 0.85 + rand.Float64()*0.10 // Improve to 0.85-0.95

			geocodingCalls++

			// Add geocoding metadata
			if locations[i].AdditionalContent == nil {
				locations[i].AdditionalContent = make(map[string]interface{})
			}
			locations[i].AdditionalContent["geocoding"] = map[string]interface{}{
				"provider":         "nominatim_mock",
				"original_confidence": locations[i].CoordinatesConfidence,
				"geocoded":           true,
				"geocoded_at":        time.Now(),
			}
		}

		// Validate address completeness
		if locations[i].Address.FormattedAddress == "" {
			locations[i].Address.FormattedAddress = fmt.Sprintf("%s, %s, %s",
				locations[i].Name, locations[i].Address.Locality, locations[i].Address.Country)
		}
	}

	return locations, geocodingCalls
}

// DedupeCanonicalizeHandler implements deduplication and canonicalization
type DedupeCanonicalizeHandler struct {
	logger *log.Logger
}

func NewDedupeCanonicalizeHandler() *DedupeCanonicalizeHandler {
	return &DedupeCanonicalizeHandler{
		logger: log.New(os.Stdout, "[DedupeCanonicalize] ", log.LstdFlags),
	}
}

func (h *DedupeCanonicalizeHandler) Name() string {
	return "DedupeCanonicalize"
}

func (h *DedupeCanonicalizeHandler) Execute(ctx context.Context, input StateInput) (StateOutput, error) {
	start := time.Now()
	h.logger.Printf("Deduplicating and canonicalizing %d locations", len(input.Locations))

	observability.CountCall(ctx, h.Name(), "execute", "start", input.City)

	// Perform deduplication
	deduped := h.deduplicateLocations(input.Locations)
	canonical := h.canonicalizeLocations(deduped)

	metrics := StateMetrics{
		Duration:   time.Since(start),
		APICalls:   0, // Local processing only
		ItemsFound: len(canonical),
	}

	observability.CountCall(ctx, h.Name(), "complete", "success", input.City)
	observability.RecordDurationMS(ctx, h.Name(), "execute", "total", input.City, float64(metrics.Duration.Milliseconds()))

	duplicatesRemoved := len(input.Locations) - len(canonical)
	h.logger.Printf("Removed %d duplicates, canonicalized %d locations in %v", duplicatesRemoved, len(canonical), metrics.Duration)

	return StateOutput{
		Locations: canonical,
		Metrics:   metrics,
	}, nil
}

func (h *DedupeCanonicalizeHandler) deduplicateLocations(locations []types.Location) []types.Location {
	var deduped []types.Location
	seen := make(map[string]bool)

	for _, loc := range locations {
		// Create a key for deduplication based on name and coordinates
		key := fmt.Sprintf("%s_%.6f_%.6f", loc.Name, loc.Coordinates.Lat, loc.Coordinates.Lng)

		// Simple proximity check - locations within ~60m are considered duplicates
		isDupe := false
		for _, existing := range deduped {
			if h.calculateDistance(loc.Coordinates, existing.Coordinates) < 60.0 && 
			   h.nameSimilarity(loc.Name, existing.Name) > 0.80 {
				isDupe = true
				break
			}
		}

		if !isDupe && !seen[key] {
			seen[key] = true
			deduped = append(deduped, loc)
		}
	}

	return deduped
}

func (h *DedupeCanonicalizeHandler) canonicalizeLocations(locations []types.Location) []types.Location {
	// Canonicalize data format and quality
	for i := range locations {
		// Standardize names
		locations[i].Name = strings.TrimSpace(locations[i].Name)

		// Ensure minimum confidence
		if locations[i].Confidence < 0.5 {
			locations[i].Confidence = 0.5
		}

		// Add canonical metadata
		if locations[i].AdditionalContent == nil {
			locations[i].AdditionalContent = make(map[string]interface{})
		}
		locations[i].AdditionalContent["canonical"] = map[string]interface{}{
			"canonicalized_at": time.Now(),
			"quality_score":    locations[i].Confidence,
			"dedupe_checked":   true,
		}
	}

	return locations
}

func (h *DedupeCanonicalizeHandler) calculateDistance(coord1, coord2 types.Coordinates) float64 {
	// Simple Euclidean distance approximation in meters
	// In production, use Haversine formula
	latDiff := coord1.Lat - coord2.Lat
	lngDiff := coord1.Lng - coord2.Lng
	return (latDiff*latDiff + lngDiff*lngDiff) * 111000 // Rough conversion to meters
}

func (h *DedupeCanonicalizeHandler) nameSimilarity(name1, name2 string) float64 {
	// Simple similarity check - in production use more sophisticated algorithms
	name1 = strings.ToLower(strings.TrimSpace(name1))
	name2 = strings.ToLower(strings.TrimSpace(name2))
	
	if name1 == name2 {
		return 1.0
	}
	
	// Check if one contains the other
	if strings.Contains(name1, name2) || strings.Contains(name2, name1) {
		return 0.85
	}
	
	// Very basic word overlap check
	words1 := strings.Fields(name1)
	words2 := strings.Fields(name2)
	
	common := 0
	for _, w1 := range words1 {
		for _, w2 := range words2 {
			if w1 == w2 {
				common++
				break
			}
		}
	}
	
	if len(words1) == 0 || len(words2) == 0 {
		return 0.0
	}
	
	return float64(common) / float64(len(words1)+len(words2)-common)
}

// PersistHandler implements data persistence (mock)
type PersistHandler struct {
	logger *log.Logger
}

func NewPersistHandler() *PersistHandler {
	return &PersistHandler{
		logger: log.New(os.Stdout, "[Persist] ", log.LstdFlags),
	}
}

func (h *PersistHandler) Name() string {
	return "Persist"
}

func (h *PersistHandler) Execute(ctx context.Context, input StateInput) (StateOutput, error) {
	start := time.Now()
	h.logger.Printf("Persisting %d locations to storage", len(input.Locations))

	observability.CountCall(ctx, h.Name(), "execute", "start", input.City)

	// Mock persistence - in production this would write to Postgres
	persistedLocations := h.mockPersistLocations(input.Locations)

	metrics := StateMetrics{
		Duration:   time.Since(start),
		APICalls:   0, // Database operations, not API calls
		ItemsFound: len(persistedLocations),
	}

	observability.CountCall(ctx, h.Name(), "complete", "success", input.City)
	observability.RecordDurationMS(ctx, h.Name(), "execute", "total", input.City, float64(metrics.Duration.Milliseconds()))

	h.logger.Printf("Persisted %d locations in %v", len(persistedLocations), metrics.Duration)

	return StateOutput{
		Locations: persistedLocations,
		Metrics:   metrics,
	}, nil
}

func (h *PersistHandler) mockPersistLocations(locations []types.Location) []types.Location {
	// Mock: add persistence metadata
	for i := range locations {
		if locations[i].AdditionalContent == nil {
			locations[i].AdditionalContent = make(map[string]interface{})
		}
		locations[i].AdditionalContent["persistence"] = map[string]interface{}{
			"persisted_at": time.Now(),
			"table":        "t_locations_prd",
			"external_id":  locations[i].SourceID,
		}
	}
	return locations
}

// RankHandler implements location ranking
type RankHandler struct {
	logger *log.Logger
}

func NewRankHandler() *RankHandler {
	return &RankHandler{
		logger: log.New(os.Stdout, "[Rank] ", log.LstdFlags),
	}
}

func (h *RankHandler) Name() string {
	return "Rank"
}

func (h *RankHandler) Execute(ctx context.Context, input StateInput) (StateOutput, error) {
	start := time.Now()
	h.logger.Printf("Ranking %d locations", len(input.Locations))

	observability.CountCall(ctx, h.Name(), "execute", "start", input.City)

	// Calculate ranks for all locations
	rankedLocations := h.calculateRanks(input.Locations)

	metrics := StateMetrics{
		Duration:   time.Since(start),
		APICalls:   0, // Local computation
		ItemsFound: len(rankedLocations),
	}

	observability.CountCall(ctx, h.Name(), "complete", "success", input.City)
	observability.RecordDurationMS(ctx, h.Name(), "execute", "total", input.City, float64(metrics.Duration.Milliseconds()))

	h.logger.Printf("Calculated ranks for %d locations in %v", len(rankedLocations), metrics.Duration)

	return StateOutput{
		Locations: rankedLocations,
		Metrics:   metrics,
	}, nil
}

func (h *RankHandler) calculateRanks(locations []types.Location) []types.Location {
	// Calculate content rank for each location
	for i := range locations {
		rank := h.calculateContentRank(locations[i])
		locations[i].ContentRank = &rank

		// Add ranking metadata
		if locations[i].AdditionalContent == nil {
			locations[i].AdditionalContent = make(map[string]interface{})
		}
		locations[i].AdditionalContent["score_breakdown"] = map[string]interface{}{
			"popularity_score": h.calculatePopularityScore(locations[i]),
			"authority_score":  h.calculateAuthorityScore(locations[i]),
			"geo_score":       h.calculateGeoScore(locations[i]),
			"novelty_score":   h.calculateNoveltyScore(locations[i]),
			"final_rank":      rank,
		}
	}

	// Sort by rank for better organization
	sort.Slice(locations, func(i, j int) bool {
		if locations[i].ContentRank == nil {
			return false
		}
		if locations[j].ContentRank == nil {
			return true
		}
		return *locations[i].ContentRank > *locations[j].ContentRank
	})

	return locations
}

func (h *RankHandler) calculateContentRank(loc types.Location) float64 {
	// Composite ranking algorithm: w_pop + w_auth + w_geo + w_novelty
	popScore := h.calculatePopularityScore(loc)
	authScore := h.calculateAuthorityScore(loc)
	geoScore := h.calculateGeoScore(loc)
	noveltyScore := h.calculateNoveltyScore(loc)

	// Weights (should be configurable)
	wPop := 0.30
	wAuth := 0.25
	wGeo := 0.20
	wNovelty := 0.25

	rank := wPop*popScore + wAuth*authScore + wGeo*geoScore + wNovelty*noveltyScore
	return rank
}

func (h *RankHandler) calculatePopularityScore(loc types.Location) float64 {
	if loc.Rating != nil {
		// Normalize rating to 0-1 scale
		return (*loc.Rating - 1.0) / 4.0 // 1-5 star rating normalized
	}
	return 0.5 // Default middle score
}

func (h *RankHandler) calculateAuthorityScore(loc types.Location) float64 {
	score := 0.5 // Base score

	// Boost for authoritative sources
	switch loc.Source {
	case "google_mock":
		score += 0.3
	case "wikipedia", "wiki":
		score += 0.25
	case "government", "gov":
		score += 0.35
	case "osm_mock":
		score += 0.15
	case "llm_extracted":
		score += 0.05
	}

	// Check for additional authority signals
	if loc.AdditionalContent != nil {
		if signals, ok := loc.AdditionalContent["signals"].(map[string]interface{}); ok {
			if hasWiki, ok := signals["has_wikipedia"].(bool); ok && hasWiki {
				score += 0.1
			}
		}
	}

	// Ensure score is between 0 and 1
	if score > 1.0 {
		score = 1.0
	}
	return score
}

func (h *RankHandler) calculateGeoScore(loc types.Location) float64 {
	// For primaries, give higher geo score
	if loc.Type == types.LocationTypePrimary {
		return 0.8
	}
	
	// Secondary locations have variable geo centrality
	return 0.3 + rand.Float64()*0.4 // 0.3-0.7
}

func (h *RankHandler) calculateNoveltyScore(loc types.Location) float64 {
	// Check for novelty indicators
	novelty := 0.5 // Base score

	if loc.AdditionalContent != nil {
		if signals, ok := loc.AdditionalContent["signals"].(map[string]interface{}); ok {
			if nov, ok := signals["novelty"].(float64); ok {
				novelty = nov
			}
		}
	}

	// LLM extracted locations are often more novel
	if loc.Source == "llm_extracted" {
		novelty += 0.2
	}

	// Tile sweep locations might be novel discoveries
	if strings.Contains(loc.ID, "tile_sweep") {
		novelty += 0.15
	}

	if novelty > 1.0 {
		novelty = 1.0
	}
	return novelty
}

// FinalizeHandler implements the final workflow step
type FinalizeHandler struct {
	logger *log.Logger
}

func NewFinalizeHandler() *FinalizeHandler {
	return &FinalizeHandler{
		logger: log.New(os.Stdout, "[Finalize] ", log.LstdFlags),
	}
}

func (h *FinalizeHandler) Name() string {
	return "Finalize"
}

func (h *FinalizeHandler) Execute(ctx context.Context, input StateInput) (StateOutput, error) {
	start := time.Now()
	h.logger.Printf("Finalizing workflow for %d locations", len(input.Locations))

	observability.CountCall(ctx, h.Name(), "execute", "start", input.City)

	// Final validation and cleanup
	finalLocations := h.finalizeLocations(input.Locations)

	metrics := StateMetrics{
		Duration:   time.Since(start),
		APICalls:   0,
		ItemsFound: len(finalLocations),
	}

	observability.CountCall(ctx, h.Name(), "complete", "success", input.City)
	observability.RecordDurationMS(ctx, h.Name(), "execute", "total", input.City, float64(metrics.Duration.Milliseconds()))

	// Calculate final statistics
	primaryCount := 0
	secondaryCount := 0
	for _, loc := range finalLocations {
		if loc.Type == types.LocationTypePrimary {
			primaryCount++
		} else {
			secondaryCount++
		}
	}

	h.logger.Printf("Finalized %d locations (%d primary, %d secondary) in %v", 
		len(finalLocations), primaryCount, secondaryCount, metrics.Duration)

	return StateOutput{
		Locations: finalLocations,
		Metrics:   metrics,
	}, nil
}

func (h *FinalizeHandler) finalizeLocations(locations []types.Location) []types.Location {
	// Final cleanup and validation
	for i := range locations {
		// Ensure all required fields are present
		if locations[i].CreatedAt.IsZero() {
			locations[i].CreatedAt = time.Now()
		}

		// Add finalization metadata
		if locations[i].AdditionalContent == nil {
			locations[i].AdditionalContent = make(map[string]interface{})
		}
		locations[i].AdditionalContent["workflow"] = map[string]interface{}{
			"finalized_at": time.Now(),
			"version":      "1.0",
			"workflow":     "jaunt-data-scout-local",
		}
	}

	return locations
}
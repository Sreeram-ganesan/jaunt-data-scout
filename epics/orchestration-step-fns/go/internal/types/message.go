package types

import "time"

// FrontierMessage is the canonical frontier message schema.
type FrontierMessage struct {
	Lat           float64 `json:"lat"`
	Lng           float64 `json:"lng"`
	RadiusMeters  int     `json:"radius"`
	Category      string  `json:"category"`
	CorrelationID string  `json:"correlation_id,omitempty"`
	City          string  `json:"city,omitempty"`
}

// Location represents a discovered place/location
type Location struct {
	ID                    string                 `json:"id"`
	Name                  string                 `json:"name"`
	Type                  LocationType           `json:"type"` // primary or secondary
	Coordinates           Coordinates            `json:"coordinates"`
	Address               Address                `json:"address,omitempty"`
	Category              string                 `json:"category,omitempty"`
	Rating                *float64               `json:"rating,omitempty"`
	Source                string                 `json:"source"`
	SourceID              string                 `json:"source_id,omitempty"`
	Confidence            float64                `json:"confidence"`
	AdditionalContent     map[string]interface{} `json:"additional_content,omitempty"`
	CreatedAt             time.Time              `json:"created_at"`
	CorrelationID         string                 `json:"correlation_id"`
	ContentRank           *float64               `json:"content_rank,omitempty"`
	AdjacencyScore        *float64               `json:"adjacency_score,omitempty"`
	TrustScore            *float64               `json:"trust_score,omitempty"`
	CoordinatesConfidence float64                `json:"coordinates_confidence"`
}

type LocationType string

const (
	LocationTypePrimary   LocationType = "primary"
	LocationTypeSecondary LocationType = "secondary"
)

type Coordinates struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type Address struct {
	FormattedAddress string `json:"formatted_address,omitempty"`
	StreetNumber     string `json:"street_number,omitempty"`
	Route            string `json:"route,omitempty"`
	Locality         string `json:"locality,omitempty"`
	Region           string `json:"region,omitempty"`
	Country          string `json:"country,omitempty"`
	PostalCode       string `json:"postal_code,omitempty"`
}

// WorkflowResult represents the final output of the workflow
type WorkflowResult struct {
	JobID         string    `json:"job_id"`
	City          string    `json:"city"`
	CorrelationID string    `json:"correlation_id"`
	CompletedAt   time.Time `json:"completed_at"`
	Summary       Summary   `json:"summary"`
	Locations     []Location `json:"locations"`
}

type Summary struct {
	TotalLocations     int `json:"total_locations"`
	PrimaryLocations   int `json:"primary_locations"`
	SecondaryLocations int `json:"secondary_locations"`
	SourcesUsed        []string `json:"sources_used"`
	ProcessingTimeMS   int64 `json:"processing_time_ms"`
	APICalls           int   `json:"api_calls"`
}

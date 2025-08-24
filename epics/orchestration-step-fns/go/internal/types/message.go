package types

// FrontierMessage is the canonical frontier message schema.
type FrontierMessage struct {
	Lat           float64 `json:"lat"`
	Lng           float64 `json:"lng"`
	RadiusMeters  int     `json:"radius"`
	Category      string  `json:"category"`
	CorrelationID string  `json:"correlation_id,omitempty"`
	City          string  `json:"city,omitempty"`
}
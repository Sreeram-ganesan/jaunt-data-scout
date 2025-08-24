package frontier

import (
    "errors"
    "time"
)

type Envelope struct {
    Type                  string   `json:"type"` // "maps" or "web"
    City                  string   `json:"city"`
    CorrelationID         string   `json:"correlation_id"`
    BudgetToken           string   `json:"budget_token,omitempty"`
    TrustScore            *float64 `json:"trust_score,omitempty"`
    CoordinatesConfidence *float64 `json:"coordinates_confidence,omitempty"`
    EnqueuedAt            int64    `json:"enqueued_at"`
}

type MapsMessage struct {
    Envelope
    Lat float64  `json:"lat"`
    Lng float64  `json:"lng"`
    Rad float64  `json:"radius"`
    Cat *string  `json:"category,omitempty"`
}

type WebMessage struct {
    Envelope
    SourceURL  string `json:"source_url"`
    SourceName string `json:"source_name"`
    SourceType string `json:"source_type"`
    CrawlDepth int    `json:"crawl_depth"`
}

func NewEnvelope(msgType, city, correlationID string) Envelope {
    return Envelope{
        Type:          msgType,
        City:          city,
        CorrelationID: correlationID,
        EnqueuedAt:    time.Now().Unix(),
    }
}

func (m MapsMessage) Validate() error {
    if m.Type != "maps" { return errors.New("type must be 'maps'") }
    if m.City == "" || m.CorrelationID == "" { return errors.New("city and correlation_id required") }
    if m.Rad <= 0 { return errors.New("radius must be > 0") }
    return nil
}

func (w WebMessage) Validate() error {
    if w.Type != "web" { return errors.New("type must be 'web'") }
    if w.City == "" || w.CorrelationID == "" { return errors.New("city and correlation_id required") }
    if w.SourceURL == "" || w.SourceType == "" { return errors.New("source_url and source_type required") }
    return nil
}
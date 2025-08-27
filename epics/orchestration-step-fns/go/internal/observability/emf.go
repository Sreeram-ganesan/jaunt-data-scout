package observability

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

const (
	// EMF Namespace for all JauntDataScout metrics
	EMFNamespace = "JauntDataScout"
)

// EMFMetric represents a single metric for AWS Embedded Metric Format
type EMFMetric struct {
	MetricName    string
	Unit          string
	Value         float64
	Service       string
	State         string
	Connector     string
	City          string
	RunID         string
	Split         string
	CorrelationID string
}

// EMFEnvelope represents the complete EMF log structure
type EMFEnvelope struct {
	AWSEMFTimestamp int64             `json:"_aws"`
	CloudWatchLogs  *EMFCloudWatch    `json:"_aws"`
	Dimensions      [][]string        `json:"_aws"`
	MetricName      string            `json:"_aws"`
	Namespace       string            `json:"_aws"`
	Metadata        map[string]interface{} `json:",inline"`
}

// EMFCloudWatch contains CloudWatch-specific EMF metadata
type EMFCloudWatch struct {
	Timestamp  int64            `json:"Timestamp"`
	CloudWatchMetrics []EMFMetricDef `json:"CloudWatchMetrics"`
}

// EMFMetricDef defines a metric within the EMF structure
type EMFMetricDef struct {
	Namespace  string     `json:"Namespace"`
	Dimensions [][]string `json:"Dimensions"`
	MetricName string     `json:"MetricName"`
	Unit       string     `json:"Unit"`
}

// emitEMF outputs a metric in AWS Embedded Metric Format to stdout
// This will be picked up by CloudWatch Logs and automatically converted to CloudWatch metrics
func emitEMF(ctx context.Context, metric *EMFMetric) {
	// Create the dimensions for the metric
	dimensions := [][]string{
		{"Service", "State", "Connector", "City"},
		{"Service", "State"},
		{"Service"},
	}

	// Add correlation_id dimension if present
	if metric.CorrelationID != "" {
		dimensions = append(dimensions, []string{"Service", "State", "Connector", "City", "CorrelationID"})
	}

	// Build the EMF envelope
	envelope := map[string]interface{}{
		"_aws": map[string]interface{}{
			"Timestamp": time.Now().UnixMilli(),
			"CloudWatchMetrics": []map[string]interface{}{
				{
					"Namespace":  EMFNamespace,
					"Dimensions": dimensions,
					"Metrics": []map[string]interface{}{
						{
							"Name": metric.MetricName,
							"Unit": metric.Unit,
						},
					},
				},
			},
		},
		// Dimension values
		"Service":   metric.Service,
		"State":     metric.State,
		"Connector": metric.Connector,
		"City":      metric.City,
		"RunID":     metric.RunID,
		"Split":     metric.Split,
		// Metric value
		metric.MetricName: metric.Value,
	}

	// Add correlation_id if present
	if metric.CorrelationID != "" {
		envelope["CorrelationID"] = metric.CorrelationID
	}

	// Marshal to JSON and output to stdout
	jsonData, err := json.Marshal(envelope)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to marshal EMF metric: %v\n", err)
		return
	}

	// Output to stdout - this will be captured by CloudWatch Logs
	fmt.Println(string(jsonData))
}

// EmitCustomEMF allows emitting a custom metric with arbitrary dimensions
func EmitCustomEMF(ctx context.Context, namespace, metricName, unit string, value float64, dimensions map[string]string) {
	// Create dimension arrays
	dimNames := make([]string, 0, len(dimensions))
	for key := range dimensions {
		dimNames = append(dimNames, key)
	}

	envelope := map[string]interface{}{
		"_aws": map[string]interface{}{
			"Timestamp": time.Now().UnixMilli(),
			"CloudWatchMetrics": []map[string]interface{}{
				{
					"Namespace":  namespace,
					"Dimensions": [][]string{dimNames},
					"Metrics": []map[string]interface{}{
						{
							"Name": metricName,
							"Unit": unit,
						},
					},
				},
			},
		},
		metricName: value,
	}

	// Add all dimensions to the envelope
	for key, value := range dimensions {
		envelope[key] = value
	}

	jsonData, err := json.Marshal(envelope)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to marshal custom EMF metric: %v\n", err)
		return
	}

	fmt.Println(string(jsonData))
}
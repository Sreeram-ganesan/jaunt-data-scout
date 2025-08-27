package observability

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFromContext_WithCorrelationID(t *testing.T) {
	ctx := context.Background()
	expectedID := "test-correlation-123"
	
	ctx = WithCorrelationID(ctx, expectedID)
	actualID := FromContext(ctx)
	
	assert.Equal(t, expectedID, actualID)
}

func TestFromContext_WithoutCorrelationID(t *testing.T) {
	ctx := context.Background()
	actualID := FromContext(ctx)
	
	assert.Empty(t, actualID)
}

func TestEnsureCorrelationID_GeneratesWhenMissing(t *testing.T) {
	ctx := context.Background()
	
	// Initially no correlation ID
	assert.Empty(t, FromContext(ctx))
	
	// EnsureCorrelationID should generate one
	ctx = EnsureCorrelationID(ctx)
	correlationID := FromContext(ctx)
	
	assert.NotEmpty(t, correlationID)
	assert.Len(t, correlationID, 36) // UUID v4 format: 8-4-4-4-12 characters + 4 hyphens
}

func TestEnsureCorrelationID_PreservesExisting(t *testing.T) {
	ctx := context.Background()
	existingID := "existing-id-123"
	
	ctx = WithCorrelationID(ctx, existingID)
	ctx = EnsureCorrelationID(ctx)
	
	assert.Equal(t, existingID, FromContext(ctx))
}

func TestLogWithCorrelationID(t *testing.T) {
	ctx := context.Background()
	correlationID := "test-123"
	ctx = WithCorrelationID(ctx, correlationID)
	
	// This test verifies the logger is created properly
	// In a real scenario, you'd capture the output to verify the formatting
	logger := LogWithCorrelationID(ctx, nil)
	assert.NotNil(t, logger)
	assert.Equal(t, correlationID, logger.correlationID)
}

func TestGenerateUUID_Format(t *testing.T) {
	uuid, err := generateUUID()
	assert.NoError(t, err)
	assert.Len(t, uuid, 36) // 8-4-4-4-12 format
	assert.Contains(t, uuid, "-")
}

func TestEMFMetricCreation(t *testing.T) {
	ctx := context.Background()
	ctx = WithCorrelationID(ctx, "test-correlation-id")
	ctx = context.WithValue(ctx, "run_id", "run-123")
	ctx = context.WithValue(ctx, "split", "primary")
	
	// Test creating an EMF metric structure
	metric := &EMFMetric{
		MetricName:    "TestMetric",
		Unit:          "Count",
		Value:         42.0,
		Service:       "test-service",
		State:         "test-state",
		Connector:     "test-connector",
		City:          "edinburgh",
		RunID:         extractRunID(ctx),
		Split:         extractSplit(ctx),
		CorrelationID: FromContext(ctx),
	}
	
	assert.Equal(t, "TestMetric", metric.MetricName)
	assert.Equal(t, "Count", metric.Unit)
	assert.Equal(t, 42.0, metric.Value)
	assert.Equal(t, "test-service", metric.Service)
	assert.Equal(t, "test-state", metric.State)
	assert.Equal(t, "test-connector", metric.Connector)
	assert.Equal(t, "edinburgh", metric.City)
	assert.Equal(t, "run-123", metric.RunID)
	assert.Equal(t, "primary", metric.Split)
	assert.Equal(t, "test-correlation-id", metric.CorrelationID)
}

func TestExtractRunID(t *testing.T) {
	tests := []struct {
		name     string
		ctx      context.Context
		expected string
	}{
		{
			name:     "with run_id in context",
			ctx:      context.WithValue(context.Background(), "run_id", "run-123"),
			expected: "run-123",
		},
		{
			name:     "without run_id in context",
			ctx:      context.Background(),
			expected: "unknown",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractRunID(tt.ctx)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractSplit(t *testing.T) {
	tests := []struct {
		name     string
		ctx      context.Context
		expected string
	}{
		{
			name:     "with split in context",
			ctx:      context.WithValue(context.Background(), "split", "secondary"),
			expected: "secondary",
		},
		{
			name:     "without split in context",
			ctx:      context.Background(),
			expected: "unknown",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractSplit(tt.ctx)
			assert.Equal(t, tt.expected, result)
		})
	}
}
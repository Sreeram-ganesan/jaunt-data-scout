package observability

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/stretchr/testify/assert"
)

func TestReadCorrelationIDFromSQS_WithAttribute(t *testing.T) {
	correlationID := "test-correlation-123"
	dataType := "String"
	message := &types.Message{
		MessageAttributes: map[string]types.MessageAttributeValue{
			CorrelationIDAttribute: {
				StringValue: &correlationID,
				DataType:    &dataType,
			},
		},
	}

	result := ReadCorrelationIDFromSQS(message)
	assert.Equal(t, correlationID, result)
}

func TestReadCorrelationIDFromSQS_WithoutAttribute(t *testing.T) {
	dataType := "String"
	message := &types.Message{
		MessageAttributes: map[string]types.MessageAttributeValue{
			"other_attribute": {
				StringValue: stringPtr("some-value"),
				DataType:    &dataType,
			},
		},
	}

	result := ReadCorrelationIDFromSQS(message)
	assert.Empty(t, result)
}

func TestReadCorrelationIDFromSQS_NilAttributes(t *testing.T) {
	message := &types.Message{
		MessageAttributes: nil,
	}

	result := ReadCorrelationIDFromSQS(message)
	assert.Empty(t, result)
}

func TestWriteCorrelationIDToSQS_NewMap(t *testing.T) {
	correlationID := "test-correlation-456"
	
	result := WriteCorrelationIDToSQS(nil, correlationID)
	
	assert.NotNil(t, result)
	assert.Contains(t, result, CorrelationIDAttribute)
	
	attr := result[CorrelationIDAttribute]
	assert.Equal(t, "String", *attr.DataType)
	assert.NotNil(t, attr.StringValue)
	assert.Equal(t, correlationID, *attr.StringValue)
}

func TestWriteCorrelationIDToSQS_ExistingMap(t *testing.T) {
	correlationID := "test-correlation-789"
	dataType := "String"
	existing := map[string]types.MessageAttributeValue{
		"existing_attr": {
			StringValue: stringPtr("existing-value"),
			DataType:    &dataType,
		},
	}
	
	result := WriteCorrelationIDToSQS(existing, correlationID)
	
	// Should preserve existing attributes
	assert.Contains(t, result, "existing_attr")
	assert.Contains(t, result, CorrelationIDAttribute)
	
	// Check correlation ID is set correctly
	attr := result[CorrelationIDAttribute]
	assert.Equal(t, "String", *attr.DataType)
	assert.NotNil(t, attr.StringValue)
	assert.Equal(t, correlationID, *attr.StringValue)
	
	// Check existing attribute is preserved
	existingAttr := result["existing_attr"]
	assert.Equal(t, "existing-value", *existingAttr.StringValue)
}

func TestContextFromSQSMessage_WithCorrelationID(t *testing.T) {
	ctx := context.Background()
	correlationID := "message-correlation-123"
	dataType := "String"
	
	message := &types.Message{
		MessageAttributes: map[string]types.MessageAttributeValue{
			CorrelationIDAttribute: {
				StringValue: &correlationID,
				DataType:    &dataType,
			},
		},
	}
	
	resultCtx := ContextFromSQSMessage(ctx, message)
	
	assert.Equal(t, correlationID, FromContext(resultCtx))
}

func TestContextFromSQSMessage_WithoutCorrelationID(t *testing.T) {
	ctx := context.Background()
	dataType := "String"
	
	message := &types.Message{
		MessageAttributes: map[string]types.MessageAttributeValue{
			"other_attr": {
				StringValue: stringPtr("some-value"),
				DataType:    &dataType,
			},
		},
	}
	
	resultCtx := ContextFromSQSMessage(ctx, message)
	
	// Should generate a new correlation ID
	correlationID := FromContext(resultCtx)
	assert.NotEmpty(t, correlationID)
	assert.Len(t, correlationID, 36) // UUID format
}

func TestBatchMessageInput(t *testing.T) {
	// Test the BatchMessageInput structure
	delaySeconds := int32(10)
	dataType := "String"
	
	batch := BatchMessageInput{
		ID:   "msg-001",
		Body: "test message body",
		MessageAttributes: map[string]types.MessageAttributeValue{
			"test_attr": {
				StringValue: stringPtr("test-value"),
				DataType:    &dataType,
			},
		},
		DelaySeconds: &delaySeconds,
	}
	
	assert.Equal(t, "msg-001", batch.ID)
	assert.Equal(t, "test message body", batch.Body)
	assert.NotNil(t, batch.MessageAttributes)
	assert.Equal(t, int32(10), *batch.DelaySeconds)
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}
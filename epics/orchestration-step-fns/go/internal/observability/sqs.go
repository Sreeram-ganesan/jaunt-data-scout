package observability

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

const (
	// SQS message attribute name for correlation_id
	CorrelationIDAttribute = "correlation_id"
)

// ReadCorrelationIDFromSQS extracts correlation_id from SQS message attributes
// Returns the correlation_id if found, empty string otherwise
func ReadCorrelationIDFromSQS(message *types.Message) string {
	if message.MessageAttributes == nil {
		return ""
	}
	
	if attr, ok := message.MessageAttributes[CorrelationIDAttribute]; ok && attr.StringValue != nil {
		return *attr.StringValue
	}
	
	return ""
}

// WriteCorrelationIDToSQS adds correlation_id to SQS message attributes
// If messageAttributes is nil, it creates a new map
func WriteCorrelationIDToSQS(messageAttributes map[string]types.MessageAttributeValue, correlationID string) map[string]types.MessageAttributeValue {
	if messageAttributes == nil {
		messageAttributes = make(map[string]types.MessageAttributeValue)
	}
	
	dataType := "String"
	messageAttributes[CorrelationIDAttribute] = types.MessageAttributeValue{
		DataType:    &dataType,
		StringValue: &correlationID,
	}
	
	return messageAttributes
}

// ContextFromSQSMessage creates a context with correlation_id extracted from SQS message
// If no correlation_id is found in the message, it generates a new one
func ContextFromSQSMessage(ctx context.Context, message *types.Message) context.Context {
	correlationID := ReadCorrelationIDFromSQS(message)
	if correlationID != "" {
		return WithCorrelationID(ctx, correlationID)
	}
	
	// Generate a new correlation_id if none found
	return EnsureCorrelationID(ctx)
}

// SQSPublishWithCorrelationID is a helper to publish to SQS with correlation_id from context
// Example usage for sending messages to SQS
func SQSPublishWithCorrelationID(ctx context.Context, sqsClient *sqs.Client, queueURL, messageBody string, additionalAttributes map[string]types.MessageAttributeValue) (*sqs.SendMessageOutput, error) {
	// Ensure we have a correlation_id
	ctx = EnsureCorrelationID(ctx)
	correlationID := FromContext(ctx)
	
	// Prepare message attributes
	messageAttributes := WriteCorrelationIDToSQS(additionalAttributes, correlationID)
	
	// Send message
	return sqsClient.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:          &queueURL,
		MessageBody:       &messageBody,
		MessageAttributes: messageAttributes,
	})
}

// SQSBatchPublishWithCorrelationID is a helper to batch publish to SQS with correlation_id
func SQSBatchPublishWithCorrelationID(ctx context.Context, sqsClient *sqs.Client, queueURL string, messages []BatchMessageInput) (*sqs.SendMessageBatchOutput, error) {
	// Ensure we have a correlation_id
	ctx = EnsureCorrelationID(ctx)
	correlationID := FromContext(ctx)
	
	// Prepare batch entries
	entries := make([]types.SendMessageBatchRequestEntry, len(messages))
	for i, msg := range messages {
		messageAttributes := WriteCorrelationIDToSQS(msg.MessageAttributes, correlationID)
		
		entries[i] = types.SendMessageBatchRequestEntry{
			Id:                &msg.ID,
			MessageBody:       &msg.Body,
			MessageAttributes: messageAttributes,
		}
		
		if msg.DelaySeconds != nil {
			entries[i].DelaySeconds = *msg.DelaySeconds
		}
	}
	
	return sqsClient.SendMessageBatch(ctx, &sqs.SendMessageBatchInput{
		QueueUrl: &queueURL,
		Entries:  entries,
	})
}

// BatchMessageInput represents a single message in a batch
type BatchMessageInput struct {
	ID                string
	Body              string
	MessageAttributes map[string]types.MessageAttributeValue
	DelaySeconds      *int32
}

// ExampleUsage demonstrates how to use the SQS correlation ID utilities
func ExampleUsage() {
	/*
	// Example 1: Processing incoming SQS messages
	func handleSQSMessage(message *types.Message) error {
		ctx := context.Background()
		
		// Extract correlation_id from message and add to context
		ctx = ContextFromSQSMessage(ctx, message)
		
		// Create logger with correlation_id
		logger := LogWithCorrelationID(ctx, log.Default())
		logger.Printf("Processing message: %s", *message.Body)
		
		// Process the message...
		processMessage(ctx, *message.Body)
		
		return nil
	}
	
	// Example 2: Sending SQS messages with correlation_id
	func sendProcessingRequest(ctx context.Context, sqsClient *sqs.Client, data string) error {
		queueURL := "https://sqs.us-east-1.amazonaws.com/123456789012/frontier-queue"
		
		// This will automatically include correlation_id from context (or generate one)
		_, err := SQSPublishWithCorrelationID(ctx, sqsClient, queueURL, data, nil)
		if err != nil {
			return fmt.Errorf("failed to send message: %w", err)
		}
		
		return nil
	}
	
	// Example 3: Batch sending with correlation_id
	func sendBatch(ctx context.Context, sqsClient *sqs.Client, messages []string) error {
		queueURL := "https://sqs.us-east-1.amazonaws.com/123456789012/frontier-queue"
		
		batchMessages := make([]BatchMessageInput, len(messages))
		for i, msg := range messages {
			batchMessages[i] = BatchMessageInput{
				ID:   fmt.Sprintf("msg-%d", i),
				Body: msg,
			}
		}
		
		_, err := SQSBatchPublishWithCorrelationID(ctx, sqsClient, queueURL, batchMessages)
		return err
	}
	*/
	fmt.Println("See function comments for SQS correlation ID usage examples")
}
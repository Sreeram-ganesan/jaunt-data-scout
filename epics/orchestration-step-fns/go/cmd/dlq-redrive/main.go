package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"

	obs "github.com/Sreeram-ganesan/jaunt-data-scout/epics/orchestration-step-fns/go/internal/observability"
	frontier "github.com/Sreeram-ganesan/jaunt-data-scout/epics/orchestration-step-fns/go/internal/frontier"
)

const (
	DefaultMaxMessages = 10
	DefaultWaitTime    = 5
)

type Config struct {
	DLQUrl       string
	FrontierUrl  string
	Region       string
	MaxMessages  int32
	DryRun       bool
	WaitTime     int32
}

type DLQMessage struct {
	ReceiptHandle string                 `json:"-"`
	MessageId     string                 `json:"message_id"`
	Body          string                 `json:"body"`
	Attributes    map[string]string      `json:"attributes"`
	CorrelationID string                 `json:"correlation_id"`
	ParsedBody    interface{}            `json:"parsed_body,omitempty"`
	Error         string                 `json:"error,omitempty"`
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "list":
		handleList()
	case "inspect":
		handleInspect()
	case "redrive":
		handleRedrive()
	case "redrive-all":
		handleRedriveAll()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("DLQ Re-drive Tool for Jaunt Data Scout")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  dlq-redrive list [--dlq-url <url>] [--max-messages <n>]")
	fmt.Println("    List messages in the DLQ")
	fmt.Println()
	fmt.Println("  dlq-redrive inspect --message-id <id> [--dlq-url <url>]")
	fmt.Println("    Inspect a specific message in detail")
	fmt.Println()
	fmt.Println("  dlq-redrive redrive --message-id <id> [--frontier-url <url>] [--dlq-url <url>] [--dry-run]")
	fmt.Println("    Re-drive a specific message to the frontier queue")
	fmt.Println()
	fmt.Println("  dlq-redrive redrive-all [--frontier-url <url>] [--dlq-url <url>] [--dry-run] [--max-messages <n>]")
	fmt.Println("    Re-drive all messages from DLQ to frontier queue")
	fmt.Println()
	fmt.Println("Environment Variables:")
	fmt.Println("  DLQ_URL      - DLQ URL (required)")
	fmt.Println("  FRONTIER_URL - Frontier queue URL (required for redrive operations)")
	fmt.Println("  AWS_REGION   - AWS region (default: us-east-1)")
	fmt.Println()
}

func handleList() {
	cfg := parseConfigForList()
	
	ctx := context.Background()
	ctx = obs.EnsureCorrelationID(ctx)
	logger := obs.LogWithCorrelationID(ctx, log.Default())
	
	awsCfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(cfg.Region))
	if err != nil {
		logger.Printf("Failed to load AWS config: %v", err)
		os.Exit(1)
	}
	
	sqsClient := sqs.NewFromConfig(awsCfg)
	
	messages, err := receiveDLQMessages(ctx, sqsClient, cfg)
	if err != nil {
		logger.Printf("Failed to receive DLQ messages: %v", err)
		os.Exit(1)
	}
	
	if len(messages) == 0 {
		fmt.Println("No messages found in DLQ")
		return
	}
	
	fmt.Printf("Found %d message(s) in DLQ:\n\n", len(messages))
	
	for i, msg := range messages {
		fmt.Printf("Message %d:\n", i+1)
		fmt.Printf("  ID: %s\n", msg.MessageId)
		fmt.Printf("  Correlation ID: %s\n", msg.CorrelationID)
		
		// Try to parse the message body
		if err := parseMessageBody(&msg); err != nil {
			fmt.Printf("  Parse Error: %s\n", err.Error())
		} else {
			if envelope, ok := msg.ParsedBody.(frontier.Envelope); ok {
				fmt.Printf("  Type: %s\n", envelope.Type)
				fmt.Printf("  City: %s\n", envelope.City)
				fmt.Printf("  Enqueued At: %s\n", time.Unix(envelope.EnqueuedAt, 0).Format(time.RFC3339))
			}
		}
		
		fmt.Printf("  Body Preview: %.100s...\n", msg.Body)
		fmt.Println()
	}
}

func handleInspect() {
	messageId := getRequiredArg("--message-id")
	cfg := parseConfigForList()
	
	ctx := context.Background()
	ctx = obs.EnsureCorrelationID(ctx)
	logger := obs.LogWithCorrelationID(ctx, log.Default())
	
	awsCfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(cfg.Region))
	if err != nil {
		logger.Printf("Failed to load AWS config: %v", err)
		os.Exit(1)
	}
	
	sqsClient := sqs.NewFromConfig(awsCfg)
	
	messages, err := receiveDLQMessages(ctx, sqsClient, cfg)
	if err != nil {
		logger.Printf("Failed to receive DLQ messages: %v", err)
		os.Exit(1)
	}
	
	var targetMessage *DLQMessage
	for _, msg := range messages {
		if msg.MessageId == messageId {
			targetMessage = &msg
			break
		}
	}
	
	if targetMessage == nil {
		fmt.Printf("Message with ID %s not found in DLQ\n", messageId)
		os.Exit(1)
	}
	
	// Parse the message body
	parseMessageBody(targetMessage)
	
	// Pretty print the message
	output, err := json.MarshalIndent(targetMessage, "", "  ")
	if err != nil {
		logger.Printf("Failed to marshal message: %v", err)
		os.Exit(1)
	}
	
	fmt.Println(string(output))
}

func handleRedrive() {
	messageId := getRequiredArg("--message-id")
	cfg := parseConfigForRedrive()
	
	ctx := context.Background()
	ctx = obs.EnsureCorrelationID(ctx)
	logger := obs.LogWithCorrelationID(ctx, log.Default())
	
	if cfg.DryRun {
		fmt.Println("DRY RUN MODE - No actual operations will be performed")
	}
	
	awsCfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(cfg.Region))
	if err != nil {
		logger.Printf("Failed to load AWS config: %v", err)
		os.Exit(1)
	}
	
	sqsClient := sqs.NewFromConfig(awsCfg)
	
	messages, err := receiveDLQMessages(ctx, sqsClient, cfg)
	if err != nil {
		logger.Printf("Failed to receive DLQ messages: %v", err)
		os.Exit(1)
	}
	
	var targetMessage *DLQMessage
	for _, msg := range messages {
		if msg.MessageId == messageId {
			targetMessage = &msg
			break
		}
	}
	
	if targetMessage == nil {
		fmt.Printf("Message with ID %s not found in DLQ\n", messageId)
		os.Exit(1)
	}
	
	if err := redriveMessage(ctx, sqsClient, cfg, *targetMessage, logger); err != nil {
		logger.Printf("Failed to redrive message: %v", err)
		os.Exit(1)
	}
	
	fmt.Printf("Successfully redriven message %s\n", messageId)
}

func handleRedriveAll() {
	cfg := parseConfigForRedrive()
	
	ctx := context.Background()
	ctx = obs.EnsureCorrelationID(ctx)
	logger := obs.LogWithCorrelationID(ctx, log.Default())
	
	if cfg.DryRun {
		fmt.Println("DRY RUN MODE - No actual operations will be performed")
	}
	
	awsCfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(cfg.Region))
	if err != nil {
		logger.Printf("Failed to load AWS config: %v", err)
		os.Exit(1)
	}
	
	sqsClient := sqs.NewFromConfig(awsCfg)
	
	messages, err := receiveDLQMessages(ctx, sqsClient, cfg)
	if err != nil {
		logger.Printf("Failed to receive DLQ messages: %v", err)
		os.Exit(1)
	}
	
	if len(messages) == 0 {
		fmt.Println("No messages found in DLQ")
		return
	}
	
	fmt.Printf("Found %d message(s) to redrive\n", len(messages))
	
	redriveCount := 0
	errorCount := 0
	
	for _, msg := range messages {
		fmt.Printf("Redriving message %s (correlation_id: %s)...", msg.MessageId, msg.CorrelationID)
		
		if err := redriveMessage(ctx, sqsClient, cfg, msg, logger); err != nil {
			fmt.Printf(" ERROR: %v\n", err)
			errorCount++
		} else {
			fmt.Printf(" SUCCESS\n")
			redriveCount++
		}
	}
	
	fmt.Printf("\nCompleted: %d successful, %d errors\n", redriveCount, errorCount)
}

func receiveDLQMessages(ctx context.Context, sqsClient *sqs.Client, cfg Config) ([]DLQMessage, error) {
	input := &sqs.ReceiveMessageInput{
		QueueUrl:            &cfg.DLQUrl,
		MaxNumberOfMessages: cfg.MaxMessages,
		WaitTimeSeconds:     cfg.WaitTime,
		MessageAttributeNames: []string{"All"},
		AttributeNames: []types.QueueAttributeName{
			types.QueueAttributeNameAll,
		},
	}
	
	result, err := sqsClient.ReceiveMessage(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to receive messages: %w", err)
	}
	
	var messages []DLQMessage
	for _, msg := range result.Messages {
		dlqMsg := DLQMessage{
			ReceiptHandle: *msg.ReceiptHandle,
			MessageId:     *msg.MessageId,
			Body:          *msg.Body,
			Attributes:    make(map[string]string),
		}
		
		// Extract attributes
		for name, value := range msg.Attributes {
			dlqMsg.Attributes[name] = value
		}
		
		// Extract correlation_id from message attributes
		dlqMsg.CorrelationID = obs.ReadCorrelationIDFromSQS(&msg)
		
		messages = append(messages, dlqMsg)
	}
	
	return messages, nil
}

func redriveMessage(ctx context.Context, sqsClient *sqs.Client, cfg Config, msg DLQMessage, logger *obs.CorrelationLogger) error {
	// Validate message body can be parsed
	if err := parseMessageBody(&msg); err != nil {
		return fmt.Errorf("message validation failed: %w", err)
	}
	
	// Check for duplicate protection - ensure correlation_id exists
	if msg.CorrelationID == "" {
		return fmt.Errorf("message missing correlation_id, cannot safely redrive")
	}
	
	if cfg.DryRun {
		logger.Printf("DRY RUN: Would redrive message %s with correlation_id %s", msg.MessageId, msg.CorrelationID)
		return nil
	}
	
	// Re-enqueue to frontier with correlation_id
	ctx = obs.WithCorrelationID(ctx, msg.CorrelationID)
	
	_, err := obs.SQSPublishWithCorrelationID(ctx, sqsClient, cfg.FrontierUrl, msg.Body, nil)
	if err != nil {
		return fmt.Errorf("failed to enqueue to frontier: %w", err)
	}
	
	// Delete from DLQ only after successful enqueue
	_, err = sqsClient.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      &cfg.DLQUrl,
		ReceiptHandle: &msg.ReceiptHandle,
	})
	if err != nil {
		return fmt.Errorf("failed to delete from DLQ (message was redriven): %w", err)
	}
	
	// Emit metrics
	obs.CountCall(ctx, "dlq_redrive", "redrive", "success", "")
	
	return nil
}

func parseMessageBody(msg *DLQMessage) error {
	// Try to parse as frontier message
	var envelope frontier.Envelope
	if err := json.Unmarshal([]byte(msg.Body), &envelope); err != nil {
		msg.Error = fmt.Sprintf("JSON parse error: %v", err)
		return err
	}
	
	// Validate the message type and parse accordingly
	switch envelope.Type {
	case "maps":
		var mapsMsg frontier.MapsMessage
		if err := json.Unmarshal([]byte(msg.Body), &mapsMsg); err != nil {
			msg.Error = fmt.Sprintf("Maps message parse error: %v", err)
			return err
		}
		if err := mapsMsg.Validate(); err != nil {
			msg.Error = fmt.Sprintf("Maps message validation error: %v", err)
			return err
		}
		msg.ParsedBody = mapsMsg
	case "web":
		var webMsg frontier.WebMessage
		if err := json.Unmarshal([]byte(msg.Body), &webMsg); err != nil {
			msg.Error = fmt.Sprintf("Web message parse error: %v", err)
			return err
		}
		if err := webMsg.Validate(); err != nil {
			msg.Error = fmt.Sprintf("Web message validation error: %v", err)
			return err
		}
		msg.ParsedBody = webMsg
	default:
		msg.Error = fmt.Sprintf("Unknown message type: %s", envelope.Type)
		return fmt.Errorf("unknown message type: %s", envelope.Type)
	}
	
	return nil
}

func parseConfigForList() Config {
	cfg := Config{
		DLQUrl:      getEnv("DLQ_URL", ""),
		Region:      getEnv("AWS_REGION", "us-east-1"),
		MaxMessages: int32(getIntArg("--max-messages", DefaultMaxMessages)),
		WaitTime:    DefaultWaitTime,
	}
	
	if dlqUrl := getOptionalArg("--dlq-url"); dlqUrl != "" {
		cfg.DLQUrl = dlqUrl
	}
	
	if cfg.DLQUrl == "" {
		fmt.Println("Error: DLQ URL is required (use --dlq-url or DLQ_URL env var)")
		os.Exit(1)
	}
	
	return cfg
}

func parseConfigForRedrive() Config {
	cfg := parseConfigForList()
	cfg.FrontierUrl = getEnv("FRONTIER_URL", "")
	cfg.DryRun = hasArg("--dry-run")
	
	if frontierUrl := getOptionalArg("--frontier-url"); frontierUrl != "" {
		cfg.FrontierUrl = frontierUrl
	}
	
	if !cfg.DryRun && cfg.FrontierUrl == "" {
		fmt.Println("Error: Frontier URL is required for redrive operations (use --frontier-url or FRONTIER_URL env var)")
		os.Exit(1)
	}
	
	return cfg
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getRequiredArg(name string) string {
	for i, arg := range os.Args {
		if arg == name && i+1 < len(os.Args) {
			return os.Args[i+1]
		}
	}
	fmt.Printf("Error: %s is required\n", name)
	os.Exit(1)
	return ""
}

func getOptionalArg(name string) string {
	for i, arg := range os.Args {
		if arg == name && i+1 < len(os.Args) {
			return os.Args[i+1]
		}
	}
	return ""
}

func getIntArg(name string, defaultValue int) int {
	for i, arg := range os.Args {
		if arg == name && i+1 < len(os.Args) {
			if val, err := strconv.Atoi(os.Args[i+1]); err == nil {
				return val
			}
		}
	}
	return defaultValue
}

func hasArg(name string) bool {
	for _, arg := range os.Args {
		if arg == name {
			return true
		}
	}
	return false
}
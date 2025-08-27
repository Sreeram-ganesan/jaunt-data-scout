# DLQ Re-drive Tool

A command-line tool for managing Dead Letter Queue (DLQ) messages in Jaunt Data Scout. This tool provides safe operations for inspecting, validating, and re-driving failed messages back to the frontier queue.

## Features

- **List DLQ Messages**: View all messages currently in the DLQ with summary information
- **Inspect Messages**: Get detailed information about specific messages including parsing errors
- **Safe Re-drive**: Re-enqueue messages back to the frontier queue with safeguards
- **Batch Operations**: Re-drive multiple messages with safety limits  
- **Dry-run Mode**: Test operations before executing them
- **Correlation ID Protection**: Prevents duplicate processing by validating correlation IDs
- **Schema Validation**: Only re-drives valid frontier messages (maps/web types)

## Installation

```bash
# Build the tool
cd epics/orchestration-step-fns/go
go build -o dlq-redrive ./cmd/dlq-redrive

# Or build to specific location
go build -o /usr/local/bin/dlq-redrive ./cmd/dlq-redrive
```

## Configuration

The tool requires AWS credentials and SQS queue URLs. Configure using environment variables:

```bash
export DLQ_URL="https://sqs.us-east-1.amazonaws.com/123456789012/jaunt-dev-frontier-dlq"
export FRONTIER_URL="https://sqs.us-east-1.amazonaws.com/123456789012/jaunt-dev-frontier"
export AWS_REGION="us-east-1"  # Optional, defaults to us-east-1
```

## Usage

### List Messages in DLQ
```bash
# List up to 10 messages (default)
./dlq-redrive list

# List up to 50 messages
./dlq-redrive list --max-messages 50

# Use specific DLQ URL
./dlq-redrive list --dlq-url "https://sqs.us-east-1.amazonaws.com/account/queue-name"
```

### Inspect a Specific Message
```bash
# Get detailed message information
./dlq-redrive inspect --message-id "12345-abcde-67890"

# Example output:
{
  "message_id": "12345-abcde-67890",
  "correlation_id": "uuid-correlation-id",
  "body": "{\"type\":\"maps\",\"city\":\"edinburgh\"...}",
  "parsed_body": {
    "type": "maps",
    "city": "edinburgh",
    "lat": 55.9533,
    "lng": -3.1883
  }
}
```

### Re-drive a Single Message
```bash
# Always test with dry-run first
./dlq-redrive redrive --message-id "12345-abcde-67890" --dry-run

# Execute the re-drive
./dlq-redrive redrive --message-id "12345-abcde-67890"
```

### Bulk Re-drive Operations
```bash
# Re-drive all messages in DLQ (dry-run first)
./dlq-redrive redrive-all --dry-run

# Re-drive up to 10 messages
./dlq-redrive redrive-all --max-messages 10

# Use specific URLs
./dlq-redrive redrive-all \
    --dlq-url "https://sqs.us-east-1.amazonaws.com/account/dlq" \
    --frontier-url "https://sqs.us-east-1.amazonaws.com/account/frontier"
```

## Safety Features

### Built-in Safeguards
- **Correlation ID Validation**: Messages without correlation_id are rejected
- **Message Schema Validation**: Only valid frontier messages are processed
- **Atomic Operations**: Messages are deleted from DLQ only after successful re-enqueue
- **Error Reporting**: Clear feedback on validation failures
- **Dry-run Mode**: Test operations without making changes

### Message Validation
The tool validates that messages conform to expected frontier message schemas:

**Maps Messages**: Must have `type: "maps"`, city, correlation_id, lat, lng, radius  
**Web Messages**: Must have `type: "web"`, city, correlation_id, source_url, source_type

Invalid messages are reported but not re-driven to prevent system errors.

## Error Handling

### Common Error Messages

**"message missing correlation_id, cannot safely redrive"**  
- Message lacks correlation ID, re-driving could create duplicates
- Investigate message source to fix correlation ID propagation

**"Maps message validation error"**  
- Required fields missing or invalid (city, lat/lng, radius)
- Check data source and input validation

**"Unknown message type"**  
- Message type is not "maps" or "web"
- May indicate corrupted data or schema changes

**"failed to enqueue to frontier"**  
- Network or permissions issue with frontier queue
- Check AWS credentials and queue permissions

## Integration with Monitoring

### CloudWatch Metrics
The tool emits metrics for monitoring re-drive operations:
- `dlq_redrive.redrive.success`: Successful re-drive operations
- Includes correlation_id in logs for traceability

### Recommended Dashboards
Monitor these metrics after re-drive operations:
- DLQ depth (should decrease)
- Frontier queue processing (should increase)
- Step Functions execution success rate
- New DLQ messages (indicating re-drive failures)

## Testing

Run the test suite to validate functionality:
```bash
go test ./cmd/dlq-redrive -v
```

The tests cover:
- Message parsing for both maps and web types
- Error handling for invalid messages
- Configuration validation
- Safety checks for correlation IDs

## Troubleshooting

### Permission Issues
Ensure your AWS credentials have these permissions:
- `sqs:ReceiveMessage` on the DLQ
- `sqs:DeleteMessage` on the DLQ  
- `sqs:SendMessage` on the frontier queue
- `sqs:GetQueueAttributes` on both queues

### No Messages Visible
Messages may be temporarily invisible due to:
- Another process is receiving messages (60s visibility timeout)
- Messages are being processed by Step Functions
- Wrong queue URL or region

Wait 60 seconds and retry, or check queue attributes:
```bash
aws sqs get-queue-attributes --queue-url $DLQ_URL --attribute-names ApproximateNumberOfMessages
```

### Re-drive Fails
Common issues:
- Frontier queue permissions
- Step Functions not processing frontier queue  
- Network connectivity issues
- Invalid message format

Check CloudWatch logs and AWS service status.

## Best Practices

### Operational
1. **Always use dry-run first** for any operation
2. **Start with single messages** before bulk operations  
3. **Monitor after re-drive** to ensure processing success
4. **Document actions taken** for incident response records
5. **Investigate patterns** in failed messages to prevent recurrence

### Development
1. **Test message formats** before deploying producers
2. **Always include correlation_id** in message attributes
3. **Validate schemas** in CI/CD pipeline
4. **Monitor DLQ depth** proactively with alarms

## See Also

- [Incident Response Runbook](../../docs/runbooks/incident-response.md)
- [DLQ Operations Guide](../../docs/runbooks/dlq-operations.md)
- [Observability Guide](../../docs/observability.md)
# DLQ Re-drive Operations Guide

This guide covers day-to-day operations for managing the Dead Letter Queue (DLQ) in Jaunt Data Scout, including message inspection and re-driving procedures.

## Quick Reference

### Environment Setup
```bash
# Set up environment variables for your target environment
export DLQ_URL="https://sqs.us-east-1.amazonaws.com/ACCOUNT/jaunt-dev-frontier-dlq"
export FRONTIER_URL="https://sqs.us-east-1.amazonaws.com/ACCOUNT/jaunt-dev-frontier"  
export AWS_REGION="us-east-1"

# Build the tool (if not already available)
cd epics/orchestration-step-fns/go
go build -o dlq-redrive ./cmd/dlq-redrive
```

### Common Operations

#### 1. Check DLQ Status
```bash
# List current messages in DLQ (default 10, max 50)
./dlq-redrive list

# List more messages if needed
./dlq-redrive list --max-messages 50
```

#### 2. Investigate Failed Messages
```bash
# Get detailed information about a specific message
./dlq-redrive inspect --message-id <message-id>

# Example output shows message type, correlation ID, parsing errors, etc.
```

#### 3. Re-drive Messages Safely
```bash
# Always test with dry-run first
./dlq-redrive redrive --message-id <message-id> --dry-run

# Execute the re-drive after confirming
./dlq-redrive redrive --message-id <message-id>

# For bulk operations (use with caution)
./dlq-redrive redrive-all --dry-run --max-messages 5
./dlq-redrive redrive-all --max-messages 5
```

## Safety Features

### Built-in Safeguards
- **Correlation ID Validation**: Messages without correlation_id are rejected to prevent duplicates
- **Message Schema Validation**: Only valid frontier messages (maps/web) are re-driven
- **Dry-run Mode**: Test operations before executing them
- **Batch Limits**: Prevent accidental mass re-drive operations
- **Error Reporting**: Clear feedback on why messages fail validation

### Best Practices
- **Always use dry-run first** for any re-drive operation
- **Start with single messages** before bulk operations
- **Inspect message content** to understand failure patterns
- **Monitor CloudWatch metrics** after re-drive operations
- **Document actions** taken for future reference

## Message Types and Validation

### Maps Messages
```json
{
  "type": "maps",
  "city": "edinburgh", 
  "correlation_id": "uuid-v4",
  "lat": 55.9533,
  "lng": -3.1883,
  "radius": 1000.0,
  "category": "restaurant",
  "enqueued_at": 1699123456
}
```

### Web Messages
```json
{
  "type": "web",
  "city": "edinburgh",
  "correlation_id": "uuid-v4", 
  "source_url": "https://example.com",
  "source_name": "example",
  "source_type": "restaurant",
  "crawl_depth": 1,
  "enqueued_at": 1699123456
}
```

## Common Failure Patterns

### 1. Invalid Message Format
**Symptom**: Parse error in inspect output  
**Cause**: Corrupted JSON or missing required fields  
**Action**: Do not re-drive; delete manually if needed

### 2. Missing Correlation ID  
**Symptom**: "message missing correlation_id, cannot safely redrive"  
**Cause**: Message created without proper observability setup  
**Action**: Do not re-drive to avoid duplicates; investigate source

### 3. Schema Validation Failures
**Symptom**: Maps/Web message validation errors  
**Cause**: Required fields missing or invalid values  
**Action**: Check data source and fix upstream issues

### 4. External Service Timeouts
**Symptom**: Multiple messages from same time period  
**Cause**: Connector timeouts or API outages  
**Action**: Safe to re-drive once external service is restored

## Monitoring After Re-drive

### CloudWatch Metrics to Watch
```bash
# Check processing success after re-drive
aws cloudwatch get-metric-statistics \
    --namespace "JauntDataScout" \
    --metric-name "Calls" \
    --dimensions Name=Service,Value=dlq_redrive Name=State,Value=redrive \
    --start-time $(date -u -d '10 minutes ago' +%Y-%m-%dT%H:%M:%S) \
    --end-time $(date -u +%Y-%m-%dT%H:%M:%S) \
    --period 300 \
    --statistics Sum

# Monitor for new DLQ messages (indicating re-drive failures)
aws sqs get-queue-attributes --queue-url $DLQ_URL --attribute-names ApproximateNumberOfMessages
```

### Step Functions Execution Status
```bash
# Check if re-driven messages are processing successfully
aws stepfunctions list-executions \
    --state-machine-arn <state-machine-arn> \
    --status-filter RUNNING \
    --max-items 20
```

## Troubleshooting

### Tool Won't Connect
```bash
# Verify AWS credentials and permissions
aws sts get-caller-identity

# Test SQS access
aws sqs get-queue-attributes --queue-url $DLQ_URL --attribute-names QueueArn

# Check environment variables
echo "DLQ_URL: $DLQ_URL"
echo "FRONTIER_URL: $FRONTIER_URL"
echo "AWS_REGION: $AWS_REGION"
```

### No Messages Visible
```bash
# Check queue URL is correct
aws sqs list-queues | grep frontier

# Verify queue has messages
aws sqs get-queue-attributes --queue-url $DLQ_URL --attribute-names ApproximateNumberOfMessages

# Messages might be in-flight (being processed by another consumer)
# Wait 60 seconds (visibility timeout) and try again
```

### Re-drive Fails
```bash
# Check destination queue permissions
aws sqs get-queue-attributes --queue-url $FRONTIER_URL --attribute-names Policy

# Verify Step Functions is processing frontier queue
aws stepfunctions list-executions --state-machine-arn <arn> --status-filter RUNNING
```

## Emergency Procedures

### Mass DLQ Cleanup (Use with Extreme Caution)
```bash
# Only if messages are definitely corrupted and cannot be recovered
# This will PERMANENTLY DELETE messages

# First, move messages to a backup location
aws sqs receive-message --queue-url $DLQ_URL --max-number-of-messages 10 > /tmp/dlq-backup-$(date +%s).json

# Then purge the queue (DESTRUCTIVE OPERATION)
aws sqs purge-queue --queue-url $DLQ_URL
```

### Bypass DLQ for Emergency Processing
```bash
# If DLQ re-drive is not working but messages need immediate processing
# Create new frontier messages with new correlation IDs

# Extract message bodies and create new messages
# This should be done by engineering team only
```

## Integration with Alerting

### CloudWatch Alarms
The DLQ operations should be performed when these alarms trigger:
- `jaunt-dlq-depth-high`: More than 10 messages in DLQ
- `jaunt-step-functions-failures`: High execution failure rate

### Operational Metrics
- Track re-drive success rate
- Monitor message age in DLQ
- Alert on repeated re-drive failures for same correlation ID

---

**Quick Commands Cheat Sheet:**
```bash
# Daily DLQ check
./dlq-redrive list

# Investigate issues  
./dlq-redrive inspect --message-id <id>

# Safe re-drive
./dlq-redrive redrive --message-id <id> --dry-run
./dlq-redrive redrive --message-id <id>

# Emergency bulk re-drive  
./dlq-redrive redrive-all --dry-run --max-messages 10
```
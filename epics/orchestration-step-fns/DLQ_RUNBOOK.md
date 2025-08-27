# Dead Letter Queue (DLQ) Re-drive Runbook

This runbook provides procedures for handling and reprocessing messages that have been sent to the Dead Letter Queue (DLQ) due to processing failures in the Step Functions workflow.

## Overview

The Step Functions workflow uses an SQS Dead Letter Queue to capture messages that fail processing after maximum retry attempts. These messages can be analyzed, fixed, and reprocessed through the workflow.

### DLQ Configuration

- **Queue Name**: `{project_prefix}-{environment}-frontier-dlq`
- **Redrive Policy**: Messages are sent to DLQ after 5 failed processing attempts
- **Message Retention**: 14 days (configurable)
- **Visibility Timeout**: 30 seconds

## When Messages End Up in DLQ

Common scenarios that cause messages to be sent to the DLQ:

1. **Lambda Function Errors**: Unhandled exceptions in processing functions
2. **Timeout Issues**: Messages taking longer than visibility timeout to process
3. **Resource Limitations**: Lambda concurrency limits, memory issues
4. **External Service Failures**: API rate limits, network issues
5. **Data Format Issues**: Invalid message formats or missing required fields
6. **Infrastructure Problems**: AWS service outages, IAM permission issues

## Monitoring and Alerting

### Check DLQ Depth

```bash
# Get current number of messages in DLQ
aws sqs get-queue-attributes \
  --queue-url "https://sqs.us-east-1.amazonaws.com/YOUR_ACCOUNT/jaunt-dev-frontier-dlq" \
  --attribute-names ApproximateNumberOfMessages ApproximateNumberOfMessagesNotVisible

# Monitor DLQ depth over time
aws cloudwatch get-metric-statistics \
  --namespace "AWS/SQS" \
  --metric-name "ApproximateNumberOfMessages" \
  --dimensions Name=QueueName,Value="jaunt-dev-frontier-dlq" \
  --start-time "2024-01-01T00:00:00Z" \
  --end-time "2024-01-02T00:00:00Z" \
  --period 3600 \
  --statistics Average,Maximum
```

### Set Up Alarms

```bash
# Create CloudWatch alarm for DLQ depth
aws cloudwatch put-metric-alarm \
  --alarm-name "jaunt-dlq-depth-alarm" \
  --alarm-description "Alert when DLQ has more than 10 messages" \
  --metric-name ApproximateNumberOfMessages \
  --namespace AWS/SQS \
  --statistic Average \
  --period 300 \
  --threshold 10 \
  --comparison-operator GreaterThanThreshold \
  --dimensions Name=QueueName,Value="jaunt-dev-frontier-dlq" \
  --evaluation-periods 2 \
  --alarm-actions "arn:aws:sns:us-east-1:YOUR_ACCOUNT:jaunt-alerts"
```

## Investigation Process

### Step 1: Analyze DLQ Messages

```bash
# Peek at messages without removing them from the queue
aws sqs receive-message \
  --queue-url "https://sqs.us-east-1.amazonaws.com/YOUR_ACCOUNT/jaunt-dev-frontier-dlq" \
  --max-number-of-messages 10 \
  --wait-time-seconds 10
```

### Step 2: Examine Message Content

Look for patterns in failed messages:

```bash
# Save messages to file for analysis
aws sqs receive-message \
  --queue-url "https://sqs.us-east-1.amazonaws.com/YOUR_ACCOUNT/jaunt-dev-frontier-dlq" \
  --max-number-of-messages 10 \
  --attribute-names All \
  --message-attribute-names All > dlq_messages.json

# Analyze message content
cat dlq_messages.json | jq '.Messages[].Body' | head -5
```

### Step 3: Check Related Logs

```bash
# Check Step Functions logs for related errors
aws logs filter-log-events \
  --log-group-name "/aws/stepfunctions/data-scout-orchestration-step-function" \
  --start-time 1640995200000 \
  --filter-pattern "ERROR"

# Check Lambda function logs
aws logs filter-log-events \
  --log-group-name "/aws/lambda/jaunt-web-fetch" \
  --start-time 1640995200000 \
  --filter-pattern "[timestamp, request_id, ERROR]"
```

### Step 4: Identify Root Cause

Common root causes and their indicators:

#### Lambda Function Errors
- **Symptoms**: Exception stack traces in logs
- **Action**: Fix code bugs, deploy updated Lambda function

#### Rate Limiting
- **Symptoms**: HTTP 429 responses, API quota exceeded errors
- **Action**: Implement exponential backoff, increase API limits

#### Resource Exhaustion  
- **Symptoms**: Lambda timeout/memory errors
- **Action**: Increase Lambda timeout/memory, optimize code

#### Data Format Issues
- **Symptoms**: JSON parsing errors, missing field errors
- **Action**: Validate message schemas, fix data producers

## Message Reprocessing

### Option 1: Automatic Redrive (Recommended)

Use SQS redrive functionality to move messages back to the main queue:

```bash
# Redrive messages from DLQ back to source queue
aws sqs start-message-move-task \
  --source-arn "arn:aws:sqs:us-east-1:YOUR_ACCOUNT:jaunt-dev-frontier-dlq" \
  --destination-arn "arn:aws:sqs:us-east-1:YOUR_ACCOUNT:jaunt-dev-frontier" \
  --max-number-of-messages-per-second 10
```

Monitor the redrive progress:

```bash
# Check redrive task status
aws sqs list-message-move-tasks \
  --source-arn "arn:aws:sqs:us-east-1:YOUR_ACCOUNT:jaunt-dev-frontier-dlq"
```

### Option 2: Manual Reprocessing

For selective reprocessing or when modifications are needed:

```bash
#!/bin/bash
# manual-dlq-reprocess.sh

DLQ_URL="https://sqs.us-east-1.amazonaws.com/YOUR_ACCOUNT/jaunt-dev-frontier-dlq"
FRONTIER_URL="https://sqs.us-east-1.amazonaws.com/YOUR_ACCOUNT/jaunt-dev-frontier"

# Process messages one by one
while true; do
  # Receive a message
  MESSAGE=$(aws sqs receive-message \
    --queue-url "$DLQ_URL" \
    --max-number-of-messages 1 \
    --wait-time-seconds 5)
  
  # Check if we got a message
  if [ "$(echo "$MESSAGE" | jq '.Messages | length')" -eq 0 ]; then
    echo "No more messages in DLQ"
    break
  fi
  
  # Extract message details
  BODY=$(echo "$MESSAGE" | jq -r '.Messages[0].Body')
  RECEIPT_HANDLE=$(echo "$MESSAGE" | jq -r '.Messages[0].ReceiptHandle')
  
  echo "Processing message: $BODY"
  
  # Optionally modify the message here
  # MODIFIED_BODY=$(echo "$BODY" | jq '. + {"retry_attempt": (.retry_attempt // 0) + 1}')
  
  # Send to frontier queue
  aws sqs send-message \
    --queue-url "$FRONTIER_URL" \
    --message-body "$BODY"
  
  if [ $? -eq 0 ]; then
    # Delete from DLQ after successful requeue
    aws sqs delete-message \
      --queue-url "$DLQ_URL" \
      --receipt-handle "$RECEIPT_HANDLE"
    echo "Successfully reprocessed and removed from DLQ"
  else
    echo "Failed to requeue message"
  fi
  
  # Rate limiting
  sleep 1
done
```

### Option 3: Step Functions Re-execution

For messages that require full workflow restart:

```bash
#!/bin/bash
# dlq-to-stepfunctions.sh

DLQ_URL="https://sqs.us-east-1.amazonaws.com/YOUR_ACCOUNT/jaunt-dev-frontier-dlq"
STATE_MACHINE_ARN="arn:aws:states:us-east-1:YOUR_ACCOUNT:stateMachine:data-scout-orchestration-step-function"

# Process DLQ messages as new Step Functions executions
aws sqs receive-message \
  --queue-url "$DLQ_URL" \
  --max-number-of-messages 10 | \
jq -r '.Messages[].Body' | \
while read -r message; do
  # Extract job information from message
  JOB_ID=$(echo "$message" | jq -r '.job_id // "reprocessed-" + now')
  
  # Create Step Functions input from DLQ message
  SF_INPUT=$(echo "$message" | jq '. + {"reprocessed": true, "original_failure_time": now}')
  
  # Start new execution
  aws stepfunctions start-execution \
    --state-machine-arn "$STATE_MACHINE_ARN" \
    --name "dlq-reprocess-$JOB_ID-$(date +%s)" \
    --input "$SF_INPUT"
done
```

## Using the DLQ Helper CLI

The `tools/dlq-helper.sh` script provides convenient commands for DLQ management:

### Setup

```bash
# Make the script executable
chmod +x epics/orchestration-step-fns/tools/dlq-helper.sh

# Set environment variables (optional)
export AWS_REGION=us-east-1
export AWS_PROFILE=default
export ENVIRONMENT=dev
```

### Commands

```bash
# Check DLQ status
./tools/dlq-helper.sh status

# Peek at messages without consuming them
./tools/dlq-helper.sh peek --count 5

# Analyze message patterns
./tools/dlq-helper.sh analyze

# Move all messages back to frontier queue
./tools/dlq-helper.sh redrive-all

# Move specific number of messages back
./tools/dlq-helper.sh redrive-batch --count 10

# Monitor DLQ in real-time
./tools/dlq-helper.sh monitor --interval 30
```

## Prevention Strategies

### 1. Improve Lambda Function Reliability

- **Error Handling**: Implement comprehensive try-catch blocks
- **Validation**: Validate input data before processing
- **Retries**: Implement exponential backoff for external API calls
- **Timeouts**: Set appropriate timeouts for external calls

### 2. Monitor Resource Usage

- **Memory**: Monitor Lambda memory usage and adjust as needed
- **Timeout**: Set realistic timeout values based on processing requirements
- **Concurrency**: Monitor and adjust Lambda concurrency limits

### 3. Implement Circuit Breakers

- **API Rate Limits**: Implement circuit breakers for external API calls
- **Service Health**: Check downstream service health before processing
- **Graceful Degradation**: Implement fallback mechanisms

### 4. Data Quality

- **Schema Validation**: Validate message schemas at ingestion
- **Data Sanitization**: Clean and normalize input data
- **Required Fields**: Ensure all required fields are present

## Escalation Procedures

### Level 1: High DLQ Volume (>50 messages)

1. Run DLQ analysis: `./tools/dlq-helper.sh analyze`
2. Check recent deployments and changes
3. Review CloudWatch logs for common error patterns
4. If identified issue is fixed, redrive messages

### Level 2: Critical System Failure (>200 messages)

1. Stop Step Functions executions if necessary
2. Investigate root cause in logs and metrics
3. Fix critical issues (code bugs, infrastructure problems)
4. Coordinate with development team for fixes
5. Plan phased message reprocessing

### Level 3: Extended Outage

1. Enable incident response procedures
2. Communicate with stakeholders
3. Implement temporary workarounds if possible
4. Coordinate cross-team effort for resolution
5. Plan comprehensive message recovery strategy

## Recovery Validation

After reprocessing DLQ messages:

### 1. Verify Message Processing

```bash
# Check that messages are being processed from frontier queue
aws sqs get-queue-attributes \
  --queue-url "https://sqs.us-east-1.amazonaws.com/YOUR_ACCOUNT/jaunt-dev-frontier" \
  --attribute-names ApproximateNumberOfMessages

# Monitor Step Functions execution success rate
aws cloudwatch get-metric-statistics \
  --namespace "AWS/States" \
  --metric-name "ExecutionsSucceeded" \
  --dimensions Name=StateMachineArn,Value="arn:aws:states:us-east-1:YOUR_ACCOUNT:stateMachine:data-scout-orchestration-step-function" \
  --start-time "$(date -u -d '1 hour ago' +%Y-%m-%dT%H:%M:%S)" \
  --end-time "$(date -u +%Y-%m-%dT%H:%M:%S)" \
  --period 300 \
  --statistics Sum
```

### 2. Validate Data Quality

```bash
# Check S3 for expected outputs
aws s3 ls s3://jaunt-dev-data-scout-raw-YOUR_ACCOUNT/cities/ --recursive | tail -10

# Verify database updates (if applicable)
# Check application metrics and business logic validation
```

### 3. Monitor for Recurring Issues

```bash
# Continue monitoring DLQ depth
./tools/dlq-helper.sh monitor --interval 60

# Set up alerts for future DLQ accumulation
# Review and adjust Lambda configuration if needed
```

## Documentation and Knowledge Sharing

### Post-Incident Review

After major DLQ incidents:

1. Document root cause and resolution steps
2. Update runbooks with lessons learned
3. Improve monitoring and alerting based on incident
4. Share knowledge with team members
5. Consider system improvements to prevent recurrence

### Knowledge Base Updates

- Update this runbook with new patterns and solutions
- Document common error messages and their resolutions
- Maintain list of escalation contacts and procedures
- Keep AWS resource ARNs and configuration up to date

## Related Documentation

- [Integration Testing Guide](../INTEGRATION_TESTING.md)
- [Feature Flags Documentation](../terraform/FEATURE_FLAGS.md)
- [Step Functions Workflow Overview](../README.md)
- [Observability Guide](../../docs/observability.md)
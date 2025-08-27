# Integration Testing Guide - Step Functions City Discovery Workflow

This guide provides comprehensive instructions for integration testing the Step Functions workflow for city discovery, including both independent testing of the Step Function and complete end-to-end integration testing.

## Overview

The Step Functions city discovery workflow orchestrates the entire data pipeline:
- **Discovery Phase**: DiscoverWebSources → DiscoverTargets → SeedPrimaries → ExpandNeighbors → TileSweep
- **Processing Phase**: WebFetch → ExtractWithLLM → GeocodeValidate
- **Finalization Phase**: DedupeCanonicalize → Persist → Rank → Finalize

## Prerequisites

- AWS CLI configured with appropriate permissions
- Terraform >= 1.5
- Go >= 1.19 (for helper tools)
- Access to AWS account with permissions for:
  - Step Functions (create/execute state machines)
  - Lambda (create/invoke functions)
  - SQS (create/send/receive messages)
  - S3 (create buckets, read/write objects)
  - IAM (create roles/policies)
  - CloudWatch (logs and metrics)

## Quick Start - Complete Integration Test

### 1. Deploy Mock Infrastructure

```bash
cd epics/orchestration-step-fns/terraform

# Initialize Terraform
make init ENV=mock

# Deploy infrastructure with mock Lambda
make plan ENV=mock
make apply ENV=mock

# Get resource ARNs/URLs
make outputs ENV=mock
```

### 2. Deploy Mock Lambda Function

```bash
cd lambdas/mock-go

# Build and deploy the mock Lambda
make deploy ARCH=arm64 REGION=us-east-1

# Capture the ARN for wiring
make arn

# Set environment variable for state identification
make set-env STATE_NAME=Integration-Test
```

### 3. Update Mock Configuration

Update `terraform/envs/mock.tfvars` with your mock Lambda ARN:

```hcl
lambda_discover_web_sources_arn = "arn:aws:lambda:us-east-1:YOUR_ACCOUNT:function:jaunt-mock"
lambda_discover_targets_arn     = "arn:aws:lambda:us-east-1:YOUR_ACCOUNT:function:jaunt-mock"
# ... (repeat for all lambda_*_arn variables)
```

Then re-apply:

```bash
cd epics/orchestration-step-fns/terraform
make apply ENV=mock
```

### 4. Execute Integration Test

```bash
cd epics/orchestration-step-fns/terraform

# Start execution with Edinburgh test data
make start-exec \
  STATE_MACHINE_ARN="arn:aws:states:us-east-1:YOUR_ACCOUNT:stateMachine:data-scout-orchestration-step-function" \
  INPUT="../examples/input.edinburgh.json" \
  NAME="integration-test-$(date +%s)"
```

### 5. Monitor Execution

```bash
# Check execution status
aws stepfunctions describe-execution \
  --execution-arn "arn:aws:states:us-east-1:YOUR_ACCOUNT:execution:data-scout-orchestration-step-function:integration-test-TIMESTAMP"

# View execution history
aws stepfunctions get-execution-history \
  --execution-arn "arn:aws:states:us-east-1:YOUR_ACCOUNT:execution:data-scout-orchestration-step-function:integration-test-TIMESTAMP"

# Check CloudWatch logs
aws logs describe-log-groups --log-group-name-prefix "/aws/stepfunctions/"
```

## Independent Step Function Testing

### Testing State Machine Definition

```bash
cd epics/orchestration-step-fns/terraform

# Validate ASL definition
aws stepfunctions validate-state-machine-definition \
  --definition file://sfn/definition.asl.json

# Validate Terraform configuration
make validate
```

### Testing Individual States

You can test individual states by modifying the mock Lambda environment:

```bash
cd lambdas/mock-go

# Set state name to identify which state is being tested
make set-env STATE_NAME=DiscoverWebSources

# Test with sample input
make invoke EVENT=../../epics/orchestration-step-fns/examples/input.edinburgh.json

# Check response
cat response.json
```

### State-by-State Testing

1. **Initialize and Budget States** (Pass states - no Lambda required)
2. **DiscoverWebSources**: Tests web source discovery logic
3. **DiscoverTargets**: Tests target identification
4. **SeedPrimaries**: Tests primary location seeding
5. **ExpandNeighbors**: Tests neighbor expansion logic
6. **TileSweep**: Tests tile-based area coverage
7. **WebFetch**: Tests web content fetching
8. **ExtractWithLLM**: Tests LLM-based data extraction
9. **GeocodeValidate**: Tests geocoding and validation
10. **DedupeCanonicalize**: Tests deduplication logic
11. **Persist**: Tests data persistence
12. **Rank**: Tests result ranking
13. **Finalize**: Tests workflow completion

## Testing Different Scenarios

### 1. Budget Exhaustion Testing

Create a test input with very low budgets:

```json
{
  "job_id": "budget-test-001",
  "city": "Edinburgh",
  "budgets": {
    "google": { "tokens_per_min": 1, "max_tokens_total": 10 },
    "web": { "bytes_per_min": 1000, "max_bytes_total": 5000 }
  },
  "early_stop": { "min_new_unique_rate": 0.95, "window_calls": 10 }
}
```

### 2. Early Stop Testing

Test early stopping when discovery rate drops:

```json
{
  "job_id": "early-stop-test-001",
  "city": "Edinburgh",
  "early_stop": { "min_new_unique_rate": 0.8, "window_calls": 50 }
}
```

### 3. Kill Switch Testing

Test various kill switches:

```json
{
  "job_id": "kill-switch-test-001",
  "city": "Edinburgh",
  "kill_switches": { "google": true, "web": false, "llm": true, "open_data": false }
}
```

### 4. Error Handling Testing

Test error scenarios by modifying the mock Lambda to return errors:

```bash
cd lambdas/mock-go
make set-env STATE_NAME=ErrorTest SHOULD_ERROR=true
```

## SQS and DLQ Testing

### Monitor Queue Depth

```bash
# Check frontier queue
aws sqs get-queue-attributes \
  --queue-url "https://sqs.us-east-1.amazonaws.com/YOUR_ACCOUNT/jaunt-dev-frontier" \
  --attribute-names ApproximateNumberOfMessages

# Check DLQ
aws sqs get-queue-attributes \
  --queue-url "https://sqs.us-east-1.amazonaws.com/YOUR_ACCOUNT/jaunt-dev-frontier-dlq" \
  --attribute-names ApproximateNumberOfMessages
```

### Test Message Flow

```bash
# Send test message to frontier queue
aws sqs send-message \
  --queue-url "https://sqs.us-east-1.amazonaws.com/YOUR_ACCOUNT/jaunt-dev-frontier" \
  --message-body '{"type":"web","url":"https://example.com","correlation_id":"test-123"}'

# Receive messages
aws sqs receive-message \
  --queue-url "https://sqs.us-east-1.amazonaws.com/YOUR_ACCOUNT/jaunt-dev-frontier"
```

## S3 Testing

### Verify Raw Cache Structure

```bash
# List raw cache contents
aws s3 ls s3://jaunt-dev-data-scout-raw-YOUR_ACCOUNT/cities/edinburgh/ --recursive

# Check manifest structure
aws s3 ls s3://jaunt-dev-data-scout-raw-YOUR_ACCOUNT/manifests/ --recursive
```

## Performance Testing

### Load Testing

Create multiple concurrent executions:

```bash
#!/bin/bash
for i in {1..10}; do
  aws stepfunctions start-execution \
    --state-machine-arn "arn:aws:states:us-east-1:YOUR_ACCOUNT:stateMachine:data-scout-orchestration-step-function" \
    --name "load-test-${i}-$(date +%s)" \
    --input file://examples/input.edinburgh.json &
done
wait
```

### Monitoring Performance

```bash
# Check CloudWatch metrics
aws cloudwatch get-metric-statistics \
  --namespace "AWS/States" \
  --metric-name "ExecutionsSucceeded" \
  --dimensions Name=StateMachineArn,Value="arn:aws:states:us-east-1:YOUR_ACCOUNT:stateMachine:data-scout-orchestration-step-function" \
  --start-time "2024-01-01T00:00:00Z" \
  --end-time "2024-01-02T00:00:00Z" \
  --period 3600 \
  --statistics Sum
```

## Troubleshooting

### Common Issues

1. **Lambda Timeout**: Increase timeout in mock Lambda configuration
2. **Permission Denied**: Check IAM roles and policies
3. **State Machine Not Found**: Verify deployment completed successfully
4. **Queue Access Denied**: Check SQS permissions in IAM role

### Debug Execution

```bash
# Get execution details with error information
aws stepfunctions describe-execution \
  --execution-arn "EXECUTION_ARN" \
  --query 'status'

# Get detailed execution history
aws stepfunctions get-execution-history \
  --execution-arn "EXECUTION_ARN" \
  --query 'events[?type==`ExecutionFailed` || type==`TaskFailed`]'
```

### Log Analysis

```bash
# Stream Step Functions logs
aws logs tail "/aws/stepfunctions/data-scout-orchestration-step-function" --follow

# Filter for errors
aws logs filter-log-events \
  --log-group-name "/aws/stepfunctions/data-scout-orchestration-step-function" \
  --filter-pattern "ERROR"
```

## Cleanup

```bash
cd epics/orchestration-step-fns/terraform

# Destroy all resources
make destroy ENV=mock

# Clean up Lambda
cd ../../lambdas/mock-go
make clean
```

## Next Steps

After successful integration testing:

1. Replace mock Lambda functions with real implementations
2. Configure production budgets and timeouts
3. Set up monitoring and alerting
4. Implement DLQ reprocessing workflows
5. Configure backup and disaster recovery procedures

## Validation Checklist

- [ ] State machine deploys successfully
- [ ] Mock Lambda functions deploy and respond correctly
- [ ] End-to-end execution completes without errors
- [ ] All state transitions logged to CloudWatch
- [ ] SQS messages flow correctly between states
- [ ] S3 raw cache receives data with correct structure
- [ ] Budget controls trigger early termination when limits reached
- [ ] Kill switches properly disable specified connectors
- [ ] DLQ captures failed messages for reprocessing
- [ ] Performance meets expected thresholds
- [ ] Error handling routes failures appropriately
- [ ] Cleanup removes all resources successfully
# Observability Guide

This document describes the observability setup for Jaunt Data Scout, including logging, tracing, metrics, and monitoring.

## Overview

The observability baseline includes:
- Step Functions logging and X-Ray tracing
- AWS Embedded Metric Format (EMF) metrics emission
- Correlation ID propagation across SQS and Lambda functions
- CloudWatch dashboards and alarms

## Step Functions Logging & Tracing

### Configuration
- **Logging Level**: `ALL` - captures all execution details including state transitions and input/output data
- **Execution Data**: Included in logs for full visibility
- **X-Ray Tracing**: Enabled for distributed tracing across state executions
- **Log Group**: `/aws/stepfunctions/{state_machine_name}` with 14-day retention

### IAM Permissions
The Step Functions execution role includes permissions for:
- CloudWatch Logs (create log groups, streams, and put events)
- X-Ray (put trace segments and telemetry records)

## EMF Metrics

### Namespace
All metrics are emitted under the `JauntDataScout` namespace.

### Common Dimensions
Every metric includes these dimensions for filtering and aggregation:
- `Service`: The service emitting the metric (e.g., `orchestrator`, `web_fetch`, `llm`)
- `State`: The current state or operation (e.g., `initialize`, `process`, `complete`)
- `Connector`: The data connector being used (e.g., `google`, `tavily`, `web`)
- `City`: The city being processed (e.g., `edinburgh`)
- `RunID`: Unique identifier for the execution run
- `Split`: Processing split (e.g., `primary`, `secondary`)
- `CorrelationID`: Request correlation identifier (when available)

### Metric Types

#### Core Metrics
- **Calls** (Count): Number of API calls or operations
- **Errors** (Count): Number of errors encountered
- **Duration** (Milliseconds): Operation duration

#### Resource Usage Metrics
- **HTTPBytesIn** (Bytes): Bytes received from HTTP requests
- **TokensIn/TokensOut** (Count): LLM token consumption
- **TokenCostEstimate** (None): Estimated cost of token usage

#### Business Metrics
- **NewUniqueRate** (Percent): Rate of new unique discoveries
- **BudgetCapUtilization** (Percent): Budget utilization percentage

### Usage Examples

```go
import obs "github.com/Sreeram-ganesan/jaunt-data-scout/epics/orchestration-step-fns/go/internal/observability"

// Count API calls
obs.CountCall(ctx, "web_fetch", "scrape", "tavily", "edinburgh")

// Record errors
obs.CountError(ctx, "llm", "extract", "openai", "edinburgh")

// Record operation duration
obs.RecordDurationMS(ctx, "geocode", "validate", "nominatim", "edinburgh", 250.5)

// Record token usage
obs.RecordTokensInOut(ctx, "llm", "extract", "openai", "edinburgh", 150.0, 75.0)

// Record budget utilization
obs.BudgetCapGauge(ctx, "orchestrator", "budget_check", "llm", "edinburgh", 0.85)
```

## Correlation ID Propagation

### Context Management
Correlation IDs are managed through Go context:

```go
import obs "github.com/Sreeram-ganesan/jaunt-data-scout/epics/orchestration-step-fns/go/internal/observability"

// Ensure correlation ID exists
ctx = obs.EnsureCorrelationID(ctx)

// Get correlation ID
correlationID := obs.FromContext(ctx)

// Add correlation ID to context
ctx = obs.WithCorrelationID(ctx, "custom-correlation-id")
```

### SQS Message Integration

#### Reading from SQS
```go
// Extract correlation ID from SQS message and add to context
ctx = obs.ContextFromSQSMessage(ctx, sqsMessage)
```

#### Writing to SQS
```go
// Send message with correlation ID
_, err := obs.SQSPublishWithCorrelationID(ctx, sqsClient, queueURL, messageBody, nil)
```

### Logging with Correlation ID
```go
// Create logger that includes correlation ID in all messages
logger := obs.LogWithCorrelationID(ctx, log.Default())
logger.Printf("Processing request") // Output: [correlation_id=abc-123] Processing request
```

## CloudWatch Dashboards

### Main Dashboard Widgets

1. **Step Functions Execution Status**
   - ExecutionsFailed, ExecutionThrottled, ExecutionsTimedOut, ExecutionsSucceeded
   - Time series view with 5-minute periods

2. **SQS Queue Depth**
   - ApproximateNumberOfMessagesVisible for frontier and DLQ queues
   - Helps monitor backlog and processing bottlenecks

3. **Budget Cap Utilization**
   - BudgetCapUtilization metrics by service (orchestrator, llm, tavily, web_fetch)
   - 0-100% scale for easy monitoring

4. **API Calls & Errors**
   - Total calls and errors from JauntDataScout metrics
   - Helps track overall system health

5. **New Unique Rate**
   - NewUniqueRate metric showing discovery effectiveness
   - 0-100% scale

6. **LLM Token Usage & Costs**
   - TokensIn, TokensOut, and TokenCostEstimate
   - Sum statistics to track total usage

7. **Web Fetch Metrics**
   - HTTPBytesIn and Duration for web scraping operations
   - Monitor bandwidth and performance

### Deployment

```hcl
module "observability" {
  source = "./infra/observability"
  
  dashboard_name      = "JauntDataScout-Production"
  state_machine_arn   = module.step_functions.state_machine_arn
  frontier_queue_name = aws_sqs_queue.frontier.name
  dlq_queue_name      = aws_sqs_queue.frontier_dlq.name
  aws_region          = "us-east-1"
  
  # Optional: Email notifications
  alarm_email = "alerts@example.com"
  
  # Alarm thresholds
  dlq_depth_threshold         = 20
  execution_failure_threshold = 10
  error_rate_threshold        = 15
  budget_cap_threshold        = 95
}
```

## CloudWatch Alarms

### Pre-configured Alarms

1. **DLQ Depth High** (`jaunt-dlq-depth-high`)
   - Triggers when DLQ has more than threshold messages
   - Default: 10 messages, 2 evaluation periods

2. **Step Functions Failures** (`jaunt-step-functions-failures`)
   - Triggers on execution failure spikes
   - Default: >5 failures, 2 evaluation periods

3. **Error Rate Spike** (`jaunt-error-rate-spike`)
   - Calculates error rate as (Errors/Calls)*100
   - Default: >10% error rate, 2 evaluation periods

4. **Budget Cap Nearing** (`jaunt-budget-cap-nearing`)
   - Triggers when budget utilization is high
   - Default: >90% utilization, 1 evaluation period

### Customizing Thresholds

Adjust thresholds based on your operational requirements:
- **Development**: Higher error rate tolerance, lower DLQ thresholds
- **Production**: Lower error rate tolerance, higher reliability requirements

## Troubleshooting

### Common Issues

1. **Missing Correlation IDs**
   - Ensure `obs.EnsureCorrelationID(ctx)` is called early in request processing
   - Check SQS message attributes contain correlation_id

2. **EMF Metrics Not Appearing**
   - Verify JSON format by checking stdout logs
   - Ensure CloudWatch Logs agent is processing the log stream
   - Check IAM permissions for metrics:PutData

3. **Dashboard Not Loading**
   - Verify all metric names and dimensions match emitted metrics
   - Check AWS region consistency
   - Validate JSON syntax in dashboard template

### Log Analysis

Search CloudWatch Logs for correlation IDs:
```
fields @timestamp, @message
| filter @message like /correlation_id=abc-123/
| sort @timestamp desc
```

Search for specific metric emissions:
```
fields @timestamp, @message
| filter @message like /"JauntDataScout"/
| filter @message like /"Calls"/
```

## Best Practices

1. **Always propagate correlation IDs** across service boundaries
2. **Use structured logging** with consistent field names
3. **Emit metrics at key decision points** in your workflow
4. **Set appropriate alarm thresholds** based on SLAs
5. **Review dashboard regularly** to identify patterns and optimization opportunities
6. **Use metric math expressions** for calculated metrics like error rates
7. **Tag all resources** consistently for cost allocation and filtering
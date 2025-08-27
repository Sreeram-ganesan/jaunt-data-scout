# Incident Runbook: Jaunt Data Scout City Jobs

This runbook provides step-by-step procedures for handling common failures in Jaunt Data Scout city discovery jobs, including DLQ management, throttling, and other operational issues.

## Overview

The Jaunt Data Scout system orchestrates city discovery jobs using AWS Step Functions, SQS queues, and various data connectors. This runbook covers:

- **DLQ Management**: Dead Letter Queue monitoring and message recovery
- **Step Functions Failures**: State machine execution issues
- **Throttling**: API rate limit and connector throttling
- **Budget Overruns**: Cost and usage limit exceeded scenarios
- **Circuit Breaker**: Emergency stop and rollback procedures

## Emergency Contacts

- **On-Call Engineer**: [Your team's on-call rotation]
- **Engineering Lead**: [Team lead contact]
- **DevOps/SRE**: [Infrastructure team contact]

## Prerequisites

Before starting incident response:
1. Ensure you have AWS CLI access with appropriate permissions
2. Have the DLQ re-drive tool available: `go build -o dlq-redrive ./cmd/dlq-redrive`
3. Access to CloudWatch logs and dashboards
4. Knowledge of the current city job execution ARN and correlation IDs

## Common Failure Scenarios

### 1. DLQ Message Buildup

**Symptoms:**
- CloudWatch alarm: `jaunt-dlq-depth-high` triggered
- Multiple failed messages in DLQ
- Step Functions execution failures

**Immediate Actions:**
```bash
# 1. List DLQ messages to assess the issue
export DLQ_URL="https://sqs.us-east-1.amazonaws.com/ACCOUNT/jaunt-dev-frontier-dlq"
export FRONTIER_URL="https://sqs.us-east-1.amazonaws.com/ACCOUNT/jaunt-dev-frontier"

./dlq-redrive list --max-messages 50

# 2. Inspect a few messages for patterns
./dlq-redrive inspect --message-id <message-id>
```

**Root Cause Investigation:**
1. **Check message patterns**: Look for common cities, connectors, or error types
2. **Review CloudWatch logs**: Check Step Functions and Lambda logs for the time period
3. **Validate connector health**: Check if external APIs (Google, Tavily, etc.) are responding

**Resolution Steps:**
```bash
# Option 1: Re-drive individual messages (safer)
./dlq-redrive redrive --message-id <message-id> --dry-run  # Test first
./dlq-redrive redrive --message-id <message-id>            # Execute

# Option 2: Bulk re-drive (use with caution)
./dlq-redrive redrive-all --dry-run --max-messages 10      # Test with small batch
./dlq-redrive redrive-all --max-messages 10                # Execute small batch

# Option 3: If messages are corrupted, delete without re-driving
# Use AWS CLI to delete messages if they cannot be parsed or are malformed
aws sqs receive-message --queue-url $DLQ_URL --max-number-of-messages 10
aws sqs delete-message --queue-url $DLQ_URL --receipt-handle <receipt-handle>
```

**Prevention:**
- Review and improve input validation
- Add retry logic for transient failures
- Monitor connector health proactively

### 2. Step Functions Execution Failures

**Symptoms:**
- CloudWatch alarm: `jaunt-step-functions-failures` triggered
- Multiple execution failures in AWS Console
- High error rate in EMF metrics

**Immediate Actions:**
```bash
# 1. Get recent failed executions
aws stepfunctions list-executions \
    --state-machine-arn <state-machine-arn> \
    --status-filter FAILED \
    --max-items 10

# 2. Inspect specific execution details
aws stepfunctions describe-execution \
    --execution-arn <execution-arn>

# 3. Get execution history for failure analysis
aws stepfunctions get-execution-history \
    --execution-arn <execution-arn> \
    --max-items 50
```

**Root Cause Analysis:**
1. **State-specific failures**: Check which state is failing most frequently
2. **Input validation**: Verify execution input format and required fields
3. **Lambda timeouts**: Check if individual Lambda functions are timing out
4. **Connector outages**: Verify external service availability

**Resolution Steps:**
```bash
# For timeout issues - restart execution with reduced batch sizes
# Create new execution with smaller input set

# For connector issues - enable circuit breaker
# Update feature flags to use mock connectors temporarily
terraform apply -var-file=envs/mock.tfvars

# For corrupted state - restart from last good checkpoint
# Use Step Functions re-execution capability if available
```

**Escalation Triggers:**
- More than 5 executions failing within 1 hour
- Same state failing across multiple cities
- External connector completely unavailable

### 3. API Throttling and Rate Limits

**Symptoms:**
- HTTP 429 (Too Many Requests) in logs
- Increased latency in external connector calls
- Budget cap nearing alerts for API usage

**Immediate Actions:**
```bash
# 1. Check current API usage metrics
aws cloudwatch get-metric-statistics \
    --namespace "JauntDataScout" \
    --metric-name "Calls" \
    --dimensions Name=Connector,Value=google \
    --start-time $(date -u -d '1 hour ago' +%Y-%m-%dT%H:%M:%S) \
    --end-time $(date -u +%Y-%m-%dT%H:%M:%S) \
    --period 300 \
    --statistics Sum

# 2. Review current budget consumption
# Check EMF metrics for token usage and HTTP bytes
```

**Mitigation Steps:**
```bash
# 1. Implement temporary throttling
# Update Lambda environment variables to reduce concurrency
aws lambda update-function-configuration \
    --function-name jaunt-web-fetch \
    --reserved-concurrent-executions 5

# 2. Enable circuit breaker for affected connector
# Use feature flags to disable high-volume connectors temporarily

# 3. Pause new executions if critical
# Update Step Functions input to skip resource-intensive states
```

**Recovery Process:**
1. **Monitor rate limit reset**: Track when API quotas reset (usually hourly/daily)
2. **Gradual ramp-up**: Slowly increase concurrency after limits reset
3. **Budget rebalancing**: Redistribute API calls across time periods

### 4. Budget Overruns

**Symptoms:**
- CloudWatch alarm: `jaunt-budget-cap-nearing` triggered
- Early-stop choice states activated
- High token costs or API usage

**Immediate Response:**
```bash
# 1. Check current budget utilization
# Review EMF metrics for cost estimates and usage

# 2. Identify top consumers
aws cloudwatch get-metric-statistics \
    --namespace "JauntDataScout" \
    --metric-name "TokenCostEstimate" \
    --dimensions Name=Connector,Value=llm \
    --start-time $(date -u -d '1 day ago' +%Y-%m-%dT%H:%M:%S) \
    --end-time $(date -u +%Y-%m-%dT%H:%M:%S) \
    --period 3600 \
    --statistics Sum
```

**Cost Control Measures:**
```bash
# 1. Emergency stop for new executions
# Pause Step Functions execution starts

# 2. Reduce LLM usage
# Switch to simpler extraction models or disable LLM extraction temporarily

# 3. Implement aggressive early-stop
# Lower minimum unique rate thresholds to stop processing sooner
```

**Budget Recovery:**
1. **Daily reset**: Wait for daily budget reset if applicable
2. **Manual budget increase**: Request budget increase if justified
3. **Optimization**: Implement more efficient processing logic

### 5. Circuit Breaker and Emergency Stop

**When to Activate:**
- Critical system failures affecting multiple cities
- Security incident requiring immediate shutdown
- Cost overrun exceeding emergency thresholds
- External compliance requirement

**Emergency Stop Procedure:**
```bash
# 1. Stop all new Step Functions executions
aws stepfunctions list-state-machines | jq -r '.stateMachines[] | select(.name | contains("jaunt")) | .stateMachineArn' | \
while read arn; do
    echo "Stopping executions for $arn"
    aws stepfunctions list-executions --state-machine-arn "$arn" --status-filter RUNNING | \
    jq -r '.executions[].executionArn' | \
    while read exec_arn; do
        aws stepfunctions stop-execution --execution-arn "$exec_arn" --error "EmergencyStop" --cause "Manual emergency stop initiated"
    done
done

# 2. Switch all connectors to mock mode
terraform apply -var-file=envs/emergency-stop.tfvars

# 3. Clear frontier queues if necessary
aws sqs purge-queue --queue-url $FRONTIER_URL
```

**System Recovery:**
```bash
# 1. Validate all systems are healthy
# Run health checks on all components

# 2. Clear DLQ of emergency-stop messages
./dlq-redrive list | grep "EmergencyStop" # Identify emergency messages
# Delete emergency messages or re-drive based on assessment

# 3. Gradual restart with single city
# Start with a single city execution to validate system health

# 4. Full operational resume
terraform apply -var-file=envs/prod.tfvars  # Restore normal config
```

## Monitoring and Alerting

### Key Metrics to Monitor
- **DLQ Depth**: Should remain < 10 messages
- **Execution Failure Rate**: Should be < 5% of total executions
- **API Error Rate**: Should be < 2% per connector
- **Budget Utilization**: Monitor approaching 90% of limits
- **Processing Duration**: Track execution time vs. expected baselines

### Alert Thresholds
- **DLQ Depth High**: > 10 messages for > 5 minutes
- **Execution Failures**: > 5 failures in 1 hour
- **Error Rate Spike**: > 10% error rate for > 10 minutes
- **Budget Cap Nearing**: > 90% utilization
- **State Timeout**: Individual state > 15 minutes

## Post-Incident Actions

### Immediate (within 24 hours)
1. **Document incident timeline**: Record all actions taken
2. **Validate data integrity**: Ensure no data corruption occurred
3. **Check financial impact**: Review any unexpected costs
4. **Notify stakeholders**: Update relevant teams on resolution

### Short-term (within 1 week)
1. **Root cause analysis**: Thorough investigation of failure cause
2. **Implement monitoring improvements**: Add new alerts if gaps identified
3. **Update runbook**: Add any new procedures discovered during incident
4. **Test recovery procedures**: Validate that recovery steps work as expected

### Long-term (within 1 month)
1. **System hardening**: Implement preventive measures
2. **Process improvements**: Update deployment and monitoring processes  
3. **Training updates**: Share learnings with team
4. **Capacity planning**: Review if resource limits need adjustment

## Useful Commands Reference

### DLQ Management
```bash
# Environment setup
export DLQ_URL="https://sqs.us-east-1.amazonaws.com/ACCOUNT/jaunt-ENV-frontier-dlq"
export FRONTIER_URL="https://sqs.us-east-1.amazonaws.com/ACCOUNT/jaunt-ENV-frontier"
export AWS_REGION="us-east-1"

# List DLQ messages
./dlq-redrive list --max-messages 20

# Inspect specific message
./dlq-redrive inspect --message-id <id>

# Re-drive with dry run first
./dlq-redrive redrive --message-id <id> --dry-run
./dlq-redrive redrive --message-id <id>

# Bulk re-drive (use caution)
./dlq-redrive redrive-all --dry-run --max-messages 10
./dlq-redrive redrive-all --max-messages 10
```

### Step Functions Management
```bash
# List executions
aws stepfunctions list-executions --state-machine-arn <arn> --status-filter FAILED

# Get execution details
aws stepfunctions describe-execution --execution-arn <exec-arn>

# Stop running execution
aws stepfunctions stop-execution --execution-arn <exec-arn> --error "Manual" --cause "Incident response"

# Start new execution
aws stepfunctions start-execution --state-machine-arn <sm-arn> --name "recovery-$(date +%s)" --input file://input.json
```

### CloudWatch Monitoring
```bash
# Get recent error metrics
aws cloudwatch get-metric-statistics \
    --namespace "JauntDataScout" \
    --metric-name "Errors" \
    --start-time $(date -u -d '1 hour ago' +%Y-%m-%dT%H:%M:%S) \
    --end-time $(date -u +%Y-%m-%dT%H:%M:%S) \
    --period 300 \
    --statistics Sum

# Check alarm state
aws cloudwatch describe-alarms --alarm-names "jaunt-dlq-depth-high" "jaunt-step-functions-failures"
```

### System Health Checks
```bash
# SQS queue attributes
aws sqs get-queue-attributes --queue-url $DLQ_URL --attribute-names All
aws sqs get-queue-attributes --queue-url $FRONTIER_URL --attribute-names All

# Lambda function status
aws lambda list-functions --function-version ALL | jq '.Functions[] | select(.FunctionName | contains("jaunt"))'

# Step Functions state machine status  
aws stepfunctions describe-state-machine --state-machine-arn <arn>
```

---

**Document Version**: 1.0  
**Last Updated**: $(date +%Y-%m-%d)  
**Next Review**: $(date -d '+3 months' +%Y-%m-%d)  

For questions or updates to this runbook, please contact the Data Scout team.
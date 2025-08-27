# DiscoverWebSources Deployment Guide

This guide explains how to deploy and test the DiscoverWebSources Lambda function.

## Prerequisites

1. AWS credentials configured with appropriate permissions
2. Tavily API key (sign up at https://tavily.com)
3. Terraform installed
4. Go 1.22+ installed (for local testing)

## Deployment Steps

### 1. Deploy Infrastructure with Terraform

```bash
cd epics/orchestration-step-fns/terraform

# Initialize Terraform
terraform init

# Plan deployment with DiscoverWebSources enabled
terraform plan -var-file=envs/dev-web-discover.tfvars

# Apply infrastructure (creates Lambda, secrets, SQS, Step Functions)
terraform apply -var-file=envs/dev-web-discover.tfvars
```

### 2. Set the Tavily API Key

After Terraform deployment, set your actual Tavily API key:

```bash
# Get the secret name from Terraform output
SECRET_NAME=$(terraform output -raw tavily_secret_arn | cut -d':' -f6)

# Set the API key
aws secretsmanager put-secret-value \
  --secret-id "$SECRET_NAME" \
  --secret-string '{"api_key":"YOUR_TAVILY_API_KEY_HERE"}'
```

### 3. Deploy the Lambda Code

```bash
cd ../../lambdas/discover-web-sources

# Build and package the Lambda
make package

# Update the Lambda function code
aws lambda update-function-code \
  --function-name jaunt-dev-web-discover-discover-web-sources \
  --zip-file fileb://function.zip
```

## Testing

### 1. Unit Tests

```bash
cd lambdas/discover-web-sources

# Run unit and contract tests
go test -v

# Run integration tests (requires deployed infrastructure)
go test -tags=integration -v
```

### 2. Manual Lambda Test

```bash
# Test the Lambda function directly
aws lambda invoke \
  --function-name jaunt-dev-web-discover-discover-web-sources \
  --payload fileb://test-input.json \
  --log-type Tail \
  --query 'LogResult' \
  --output text \
  response.json | base64 --decode

# Check the response
cat response.json
```

### 3. Step Functions Test

```bash
# Start a Step Functions execution
aws stepfunctions start-execution \
  --state-machine-arn $(terraform output -raw state_machine_arn) \
  --input file://test-sfn-input.json \
  --name test-discover-web-sources-$(date +%s)
```

## Feature Flag Configuration

The DiscoverWebSources Lambda can be toggled between mock and real implementation:

### Use Real DiscoverWebSources (recommended for testing)

```hcl
# In your .tfvars file
state_implementations = {
  discover_web_sources = "real"    # Use real Tavily integration
  discover_targets     = "mock"    # Keep others as mock for now
  web_fetch           = "mock"
  extract_with_llm    = "mock"
  # ... etc
}
```

### Use Mock DiscoverWebSources (for infrastructure testing)

```hcl
# In your .tfvars file  
state_implementations = {
  discover_web_sources = "mock"    # Use mock implementation
  # ... etc
}
```

## Monitoring

### CloudWatch Metrics

The Lambda emits these metrics to CloudWatch:

- `jaunt/DiscoverWebSources/tavily.calls` - Number of Tavily API calls
- `jaunt/DiscoverWebSources/tavily.errors` - Number of API errors
- `jaunt/DiscoverWebSources/urls.discovered` - URLs found by Tavily  
- `jaunt/DiscoverWebSources/urls.enqueued` - URLs successfully queued

### CloudWatch Logs

Check Lambda logs:

```bash
aws logs tail /aws/lambda/jaunt-dev-web-discover-discover-web-sources --follow
```

### SQS Monitoring

Check the frontier queue for enqueued messages:

```bash
# Get queue attributes
aws sqs get-queue-attributes \
  --queue-url $(terraform output -raw frontier_queue_url) \
  --attribute-names ApproximateNumberOfMessages

# Receive a sample message (non-destructive)
aws sqs receive-message \
  --queue-url $(terraform output -raw frontier_queue_url) \
  --max-number-of-messages 1
```

## Troubleshooting

### Common Issues

1. **"TAVILY_SECRET_ARN environment variable is required"**
   - Ensure the secret was created by Terraform and the Lambda has the correct environment variable

2. **"Failed to get Tavily API key"**
   - Verify the API key is set in Secrets Manager
   - Check IAM permissions for secretsmanager:GetSecretValue

3. **"Failed to enqueue frontier messages"**
   - Check SQS queue exists and Lambda has send permissions
   - Verify FRONTIER_QUEUE_URL environment variable

4. **"Tavily API returned status 401"**
   - Verify your Tavily API key is valid and not expired

5. **"No URLs discovered"**
   - Try different cities or check Tavily query templates
   - Enable debug logging to see API responses

### Debug Mode

For detailed debugging, check the CloudWatch logs. The Lambda outputs structured JSON logs with correlation_id and run_id for tracing.

## Cost Considerations

- Tavily API: ~$0.002 per search query (10 queries per city by default)
- Lambda: ~$0.000017 per invocation + compute time
- SQS: ~$0.0000004 per message
- Secrets Manager: ~$0.40 per secret per month

Typical cost per city discovery run: $0.02-0.05
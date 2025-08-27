# DiscoverWebSources Lambda

This Lambda function discovers high-signal web pages for a target city using the Tavily search API and enqueues them to the web frontier SQS queue for further processing.

## Configuration

### Required Environment Variables

- `TAVILY_SECRET_ARN`: ARN of the AWS Secrets Manager secret containing the Tavily API key
- `FRONTIER_QUEUE_URL`: URL of the SQS frontier queue
- `PROJECT_PREFIX`: Project prefix for metrics and naming (default: "jaunt")  
- `ENVIRONMENT`: Deployment environment (default: "dev")

### Secret Format

The Tavily secret should be stored as JSON in AWS Secrets Manager:
```json
{
  "api_key": "your-tavily-api-key"
}
```

## Input Schema

```json
{
  "run_id": "string",
  "city": {
    "name": "Edinburgh",
    "country_code": "GB"
  },
  "orchestrator": {
    "correlation_id": "string",
    "fail_fast": false
  },
  "budgets": {
    "max_urls": 100,
    "max_per_domain": 10
  }
}
```

## Output Schema

```json
{
  "discovered_count": 85,
  "enqueued_count": 85,
  "top_domains": [
    {"domain": "visitscotland.com", "count": 5},
    {"domain": "tripadvisor.com", "count": 4}
  ],
  "sample_urls": [
    "https://www.visitscotland.com/destinations/highlands/edinburgh/",
    "https://www.tripadvisor.com/Attractions-g186525-Activities-Edinburgh_Scotland.html"
  ]
}
```

## SQS Frontier Message Schema

Each discovered URL is enqueued as a separate message with this schema:

```json
{
  "type": "web",
  "run_id": "string",
  "correlation_id": "string", 
  "url": "https://example.com/path",
  "source": "tavily",
  "budget_token": "tavily.api",
  "city": {
    "name": "Edinburgh", 
    "country_code": "GB"
  }
}
```

## Features

### URL Discovery
- Generates city-scoped search queries using configurable templates
- Calls Tavily Search API to discover relevant web pages
- Handles pagination and respects QPS limits

### URL Processing  
- Normalizes URLs (strips fragments, tracking parameters)
- Deduplicates by normalized URL
- Applies per-domain limits to avoid over-concentration
- Sorts by relevance score from Tavily

### Observability
- Structured JSON logging with run_id and correlation_id
- CloudWatch EMF metrics:
  - `tavily.calls`: Number of API calls made
  - `tavily.errors`: Number of API errors  
  - `urls.discovered`: Number of URLs found
  - `urls.enqueued`: Number of URLs successfully queued

### Error Handling
- Retries Tavily API calls with exponential backoff
- Handles partial SQS batch failures with retry
- Circuit breaker protections against excessive API usage

## Building and Testing

```bash
# Build the Lambda
make build

# Run unit tests  
make test

# Run integration tests (requires AWS credentials and TAVILY_SECRET_ARN)
go test -tags=integration -v

# Package for deployment
make package
```

## Query Templates

The function uses these default query templates, substituting `{city}` and `{country}`:

- `{city} {country} restaurants`
- `{city} {country} cafes coffee`  
- `{city} {country} bars pubs nightlife`
- `{city} {country} things to do attractions`
- `{city} {country} events venues`
- `{city} {country} tourism tourist information`
- `{city} {country} official city website`
- `{city} {country} government site`
- `{city} {country} hotels accommodation`
- `{city} {country} museums galleries`

## IAM Permissions Required

The Lambda execution role needs:

- `secretsmanager:GetSecretValue` on the Tavily secret
- `sqs:SendMessage` and `sqs:SendMessageBatch` on the frontier queue
- `logs:CreateLogGroup`, `logs:CreateLogStream`, `logs:PutLogEvents`
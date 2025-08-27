# Feature Flags System - Mock vs Real State Implementations

The Step Functions workflow supports feature flags to control which states use mock implementations versus real Lambda functions. This enables gradual rollout, testing individual components, and mixed deployments.

## Configuration

### Variables

Two new Terraform variables control the feature flags:

1. **`mock_lambda_arn`** - The ARN of the mock Lambda function to use for states configured as "mock"
2. **`state_implementations`** - An object specifying "mock" or "real" for each state

```hcl
variable "mock_lambda_arn" {
  description = "ARN of the mock Lambda function to use for states configured as mock"
  type        = string
  default     = ""
}

variable "state_implementations" {
  description = "Configure each state to use 'mock' or 'real' implementation"
  type = object({
    discover_web_sources   = optional(string, "mock")
    discover_targets       = optional(string, "mock")
    seed_primaries        = optional(string, "mock")
    expand_neighbors      = optional(string, "mock")
    tile_sweep           = optional(string, "mock")
    web_fetch            = optional(string, "mock")
    extract_with_llm     = optional(string, "mock")
    geocode_validate     = optional(string, "mock")
    dedupe_canonicalize  = optional(string, "mock")
    persist              = optional(string, "mock")
    rank                 = optional(string, "real")
  })
  default = {}
}
```

### Supported States

The following states can be individually configured:

- `discover_web_sources` - Web source discovery
- `discover_targets` - Target identification
- `seed_primaries` - Primary location seeding
- `expand_neighbors` - Neighbor expansion
- `tile_sweep` - Tile-based area coverage
- `web_fetch` - Web content fetching
- `extract_with_llm` - LLM-based data extraction
- `geocode_validate` - Geocoding and validation
- `dedupe_canonicalize` - Deduplication and canonicalization
- `persist` - Data persistence
- `rank` - Result ranking

## Usage Examples

### All Mock Configuration

Perfect for initial testing and development:

```hcl
# envs/all-mock.tfvars
mock_lambda_arn = "arn:aws:lambda:us-east-1:123456789012:function:jaunt-mock"

state_implementations = {
  discover_web_sources   = "mock"
  discover_targets       = "mock"
  seed_primaries        = "mock"
  expand_neighbors      = "mock"
  tile_sweep           = "mock"
  web_fetch            = "mock"
  extract_with_llm     = "mock"
  geocode_validate     = "mock"
  dedupe_canonicalize  = "mock"
  persist              = "mock"
  rank                 = "mock"
}

# Set all real Lambda ARNs to placeholders since they won't be used
lambda_discover_web_sources_arn = "arn:aws:lambda:us-east-1:123456789012:function:placeholder"
# ... (repeat for all lambda_*_arn variables)
```

### Mixed Configuration

Gradually introduce real implementations:

```hcl
# envs/mixed-mock-real.tfvars
mock_lambda_arn = "arn:aws:lambda:us-east-1:123456789012:function:jaunt-mock"

state_implementations = {
  discover_web_sources   = "real"    # Use real implementation
  discover_targets       = "real"    # Use real implementation
  seed_primaries        = "mock"    # Use mock for now
  expand_neighbors      = "mock"    # Use mock for now
  tile_sweep           = "mock"    # Use mock for now
  web_fetch            = "real"    # Use real implementation
  extract_with_llm     = "real"    # Use real implementation
  geocode_validate     = "mock"    # Use mock for now
  dedupe_canonicalize  = "mock"    # Use mock for now
  persist              = "real"    # Use real implementation
  rank                 = "mock"    # Use mock for now
}

# Set real Lambda ARNs for states configured as "real"
lambda_discover_web_sources_arn = "arn:aws:lambda:us-east-1:123456789012:function:jaunt-discover-web-sources"
lambda_discover_targets_arn     = "arn:aws:lambda:us-east-1:123456789012:function:jaunt-discover-targets"
lambda_web_fetch_arn            = "arn:aws:lambda:us-east-1:123456789012:function:jaunt-web-fetch"
lambda_extract_with_llm_arn     = "arn:aws:lambda:us-east-1:123456789012:function:jaunt-extract-with-llm"
lambda_persist_arn              = "arn:aws:lambda:us-east-1:123456789012:function:jaunt-persist"

# Mock ARNs can be used for states configured as "mock"
lambda_seed_primaries_arn       = "arn:aws:lambda:us-east-1:123456789012:function:placeholder"
# ... (other mock states)
```

### All Real Configuration

Production deployment with all real implementations:

```hcl
# envs/all-real.tfvars
mock_lambda_arn = "arn:aws:lambda:us-east-1:123456789012:function:unused"

state_implementations = {
  discover_web_sources   = "real"
  discover_targets       = "real"
  seed_primaries        = "real"
  expand_neighbors      = "real"
  tile_sweep           = "real"
  web_fetch            = "real"
  extract_with_llm     = "real"
  geocode_validate     = "real"
  dedupe_canonicalize  = "real"
  persist              = "real"
  rank                 = "real"
}

# All real Lambda ARNs required
lambda_discover_web_sources_arn = "arn:aws:lambda:us-east-1:123456789012:function:jaunt-prod-discover-web-sources"
# ... (all production Lambda functions)
```

## Deployment Process

### 1. Deploy with Feature Flags

```bash
cd epics/orchestration-step-fns/terraform

# Initialize and plan with your configuration
make init ENV=mixed
make plan ENV=mixed

# Apply changes
make apply ENV=mixed
```

### 2. Validate Configuration

```bash
# Check the deployed state machine definition
aws stepfunctions describe-state-machine \
  --state-machine-arn "arn:aws:states:us-east-1:YOUR_ACCOUNT:stateMachine:data-scout-orchestration-step-function" \
  --query 'definition' \
  --output text | jq .
```

### 3. Test Individual State Changes

```bash
# Run a test execution
make start-exec \
  STATE_MACHINE_ARN="arn:aws:states:us-east-1:YOUR_ACCOUNT:stateMachine:data-scout-orchestration-step-function" \
  INPUT="../examples/input.edinburgh.json" \
  NAME="feature-flag-test-$(date +%s)"
```

## Toggling Procedure

To change a state from mock to real (or vice versa):

### Step 1: Update Configuration

Modify your tfvars file:

```hcl
state_implementations = {
  # ... other states
  web_fetch = "real"  # Changed from "mock" to "real"
  # ... other states
}

# Ensure the real Lambda ARN is correct
lambda_web_fetch_arn = "arn:aws:lambda:us-east-1:123456789012:function:jaunt-web-fetch"
```

### Step 2: Apply Changes

```bash
make plan ENV=your-env
make apply ENV=your-env
```

### Step 3: Validate

```bash
# Start a test execution to validate the change
make start-exec \
  STATE_MACHINE_ARN="your-state-machine-arn" \
  INPUT="../examples/input.edinburgh.json" \
  NAME="toggle-test-$(date +%s)"

# Monitor the execution
aws stepfunctions describe-execution \
  --execution-arn "your-execution-arn"
```

## Testing Strategy

### 1. Mock Development Phase

- All states set to "mock"
- Focus on workflow logic and state transitions
- Test budget controls and early stopping
- Validate SQS message flow and S3 caching

### 2. Gradual Real Integration

- Start with foundational states (discover_web_sources, discover_targets)
- Add processing states (web_fetch, extract_with_llm)
- Finally integrate storage states (persist, rank)

### 3. Production Readiness

- All states set to "real"
- Full end-to-end testing
- Performance and load testing
- Monitoring and alerting validation

## Monitoring Feature Flag Usage

### CloudWatch Logs

The mock Lambda function can be configured with different `STATE_NAME` environment variables to identify which state is executing:

```bash
# Set state name for identification in logs
aws lambda update-function-configuration \
  --function-name jaunt-mock \
  --environment Variables="{STATE_NAME=WebFetch}"
```

### Step Functions Execution Tracing

Enable detailed logging to see which Lambda functions are being invoked:

```hcl
# In your tfvars file
sfn_log_level              = "ALL"
sfn_include_execution_data = true
enable_sfn_tracing         = true
```

## Best Practices

1. **Start with All Mock**: Begin development with all states mocked
2. **Gradual Migration**: Migrate one state at a time to real implementations
3. **Test Each Change**: Run integration tests after each feature flag change
4. **Document State**: Keep track of which states are mock vs real in each environment
5. **Environment Isolation**: Use different feature flag configurations per environment
6. **Rollback Plan**: Keep previous tfvars configuration for quick rollback

## Troubleshooting

### Invalid Implementation Value

```
Error: All state implementations must be either 'mock' or 'real'.
```

**Solution**: Check that all values in `state_implementations` are exactly "mock" or "real".

### Missing Mock Lambda ARN

```
Error: mock_lambda_arn is required when any state is configured as 'mock'
```

**Solution**: Ensure `mock_lambda_arn` is set when any state uses "mock" implementation.

### Lambda Function Not Found

```
Error: Lambda function not found
```

**Solution**: Verify all Lambda ARNs exist and are accessible from the Step Functions execution role.

### Permission Denied

```
Error: User: ... is not authorized to perform: lambda:InvokeFunction
```

**Solution**: Check IAM permissions for the Step Functions execution role include lambda:InvokeFunction for all referenced Lambda functions.

## Advanced Usage

### Environment-Specific Defaults

You can set different defaults for different environments:

```hcl
# variables.tf
locals {
  default_implementations = var.environment == "prod" ? {
    # Production defaults to real implementations
    discover_web_sources = "real"
    discover_targets     = "real"
    # ... all real
  } : {
    # Development defaults to mock implementations
    discover_web_sources = "mock"
    discover_targets     = "mock"
    # ... all mock
  }
}

state_implementations = merge(local.default_implementations, var.state_implementations)
```

### Conditional Lambda Deployment

Skip deploying real Lambda functions when not needed:

```hcl
# Only deploy real Lambda if any state uses real implementation
resource "aws_lambda_function" "web_fetch" {
  count = contains(values(var.state_implementations), "real") ? 1 : 0
  
  function_name = "jaunt-web-fetch"
  # ... other configuration
}
```
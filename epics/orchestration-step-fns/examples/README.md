# Execution Input Examples and Golden Datasets

This directory contains predefined execution input files (golden datasets) for testing the Step Functions workflow with different cities, scenarios, and configurations.

## Usage

These input files can be used with:
- Step Functions executions via AWS CLI
- Terraform Makefile: `make start-exec INPUT=examples/input.CITY.json`
- Integration testing scripts
- Load testing and performance validation

## Available Input Files

### City-Specific Inputs

#### Edinburgh (Scotland)
- **File**: `input.edinburgh.json`
- **Center**: 55.9533°N, 3.1883°W  
- **Radius**: 10km
- **Use Case**: Primary development and testing city
- **Characteristics**: Medium-sized city, good test coverage

#### London (England)  
- **File**: `input.london.json`
- **Center**: 51.5074°N, 0.1278°W
- **Radius**: 15km  
- **Use Case**: Large metropolitan area testing
- **Characteristics**: High density, large data volume

#### Tokyo (Japan)
- **File**: `input.tokyo.json`
- **Center**: 35.6762°N, 139.6503°E
- **Radius**: 12km
- **Use Case**: International city testing
- **Characteristics**: Very high density, non-English content

#### New York (USA)
- **File**: `input.new-york.json` 
- **Center**: 40.7128°N, 74.0060°W
- **Radius**: 20km
- **Use Case**: Large North American city testing
- **Characteristics**: Very large area, high data volume

### Test Scenario Inputs

#### Budget Test (Small)
- **File**: `input.budget-test-small.json`
- **Use Case**: Test budget exhaustion and early termination
- **Characteristics**: Very low budget limits to trigger early stops
- **Expected Behavior**: Should terminate due to budget constraints

#### Early Stop Test
- **File**: `input.early-stop-test.json`
- **Use Case**: Test early stopping when discovery rate drops
- **Characteristics**: High early stop threshold (80% new unique rate)
- **Expected Behavior**: Should stop when discovery becomes inefficient

#### Kill Switch Test
- **File**: `input.kill-switch-test.json`
- **Use Case**: Test kill switch functionality
- **Characteristics**: Some services disabled (Google and LLM killed)
- **Expected Behavior**: Should skip disabled service integrations

## Input Structure

All input files follow this schema:

```json
{
  "job_id": "unique-job-identifier",
  "city": "City Name",
  "seed": {
    "type": "map",
    "center": { "lat": 0.0, "lng": 0.0 },
    "radius_km": 10
  },
  "budgets": {
    "google": { "tokens_per_min": 300, "max_tokens_total": 20000 },
    "web": { "bytes_per_min": 1000000, "max_bytes_total": 200000000 },
    "llm": { "tokens_per_min": 100, "max_tokens_total": 10000 },
    "open_data": { "calls_per_min": 100, "max_calls_total": 10000 }
  },
  "kill_switches": { 
    "google": false, 
    "web": false, 
    "llm": false, 
    "open_data": false 
  },
  "early_stop": { 
    "min_new_unique_rate": 0.05, 
    "window_calls": 200 
  },
  "timeouts": { 
    "wall_clock_seconds": 3600 
  },
  "s3_prefix": "cities/city-name/test-run/"
}
```

## Quick Start Examples

### Basic City Discovery Test

```bash
cd epics/orchestration-step-fns/terraform

# Start Edinburgh discovery
make start-exec \
  STATE_MACHINE_ARN="your-state-machine-arn" \
  INPUT="../examples/input.edinburgh.json" \
  NAME="test-edinburgh-$(date +%s)"
```

### Budget Constraint Testing

```bash
# Test budget limits
make start-exec \
  STATE_MACHINE_ARN="your-state-machine-arn" \
  INPUT="../examples/input.budget-test-small.json" \
  NAME="budget-test-$(date +%s)"
```

### Kill Switch Testing

```bash
# Test with certain services disabled
make start-exec \
  STATE_MACHINE_ARN="your-state-machine-arn" \
  INPUT="../examples/input.kill-switch-test.json" \
  NAME="kill-switch-test-$(date +%s)"
```

## Testing Recommendations

### Development Phase
1. Start with `input.edinburgh.json` for basic functionality
2. Use `input.budget-test-small.json` for budget logic validation
3. Test kill switches with `input.kill-switch-test.json`

### Integration Testing
1. Test multiple cities: Edinburgh → London → Tokyo
2. Validate early stopping with `input.early-stop-test.json`
3. Run concurrent executions with different inputs

### Load Testing  
1. Use large cities (London, New York) for volume testing
2. Run multiple concurrent executions
3. Monitor resource utilization and performance

## Customizing Inputs

### Creating New City Inputs

1. Copy an existing input file
2. Update the `job_id`, `city`, and `seed` coordinates
3. Adjust budgets based on expected city size
4. Set appropriate `s3_prefix` for data organization

### Adjusting Budgets

Budget values should be scaled based on:
- **City size**: Larger cities need higher limits
- **Test duration**: Longer tests need higher total budgets
- **Connector types**: Different services have different rate limits

### Testing Specific Scenarios

- **Timeout testing**: Reduce `timeouts.wall_clock_seconds`
- **Early stop testing**: Increase `early_stop.min_new_unique_rate`
- **Service isolation**: Use `kill_switches` to disable specific connectors
- **Rate limiting**: Reduce `*_per_min` values in budgets

## Integration with Testing Tools

### End-to-End Integration Test

```bash
# Run full integration test
./tools/e2e-integration-test.sh --city edinburgh --env mock --verbose

# Test with different cities
./tools/e2e-integration-test.sh --city london --timeout 3600
./tools/e2e-integration-test.sh --city tokyo --env dev
```

### AWS CLI Direct Execution

```bash
# Direct AWS CLI execution
aws stepfunctions start-execution \
  --state-machine-arn "arn:aws:states:us-east-1:ACCOUNT:stateMachine:data-scout-orchestration-step-function" \
  --name "manual-test-$(date +%s)" \
  --input file://examples/input.edinburgh.json
```

## Monitoring Execution

### Check Execution Status

```bash
# Get execution status
aws stepfunctions describe-execution \
  --execution-arn "your-execution-arn"

# Get execution history
aws stepfunctions get-execution-history \
  --execution-arn "your-execution-arn" \
  --max-results 10
```

### Monitor Resource Usage

```bash
# Check SQS queue depth
aws sqs get-queue-attributes \
  --queue-url "your-frontier-queue-url" \
  --attribute-names ApproximateNumberOfMessages

# Check S3 objects created
aws s3 ls s3://your-s3-bucket/cities/ --recursive
```

## Validation and Quality Assurance

### Input Validation

All input files are validated for:
- Valid JSON structure
- Required field presence
- Coordinate validity (latitude/longitude ranges)
- Budget value sanity (positive numbers)
- Timeout reasonableness

### Expected Outcomes

| Input File | Expected Duration | Expected States | Expected Outputs |
|------------|------------------|-----------------|------------------|
| `input.edinburgh.json` | 15-30 minutes | All states execute | ~1000-5000 candidates |
| `input.budget-test-small.json` | 2-5 minutes | Early termination | Budget exhaustion logs |
| `input.early-stop-test.json` | 5-15 minutes | Early stopping | Low unique rate logs |
| `input.kill-switch-test.json` | 10-20 minutes | Skip disabled states | Reduced data volume |

## Troubleshooting

### Common Issues

1. **Invalid coordinates**: Ensure latitude is -90 to 90, longitude is -180 to 180
2. **Budget too low**: Increase budget values if execution terminates immediately  
3. **Radius too large**: Large radius values may cause timeouts or high costs
4. **S3 prefix conflicts**: Use unique prefixes to avoid data collisions

### Debug Commands

```bash
# Validate JSON structure
jq . examples/input.edinburgh.json

# Check coordinate validity
echo '{"lat": 55.9533, "lng": -3.1883}' | jq 'if .lat >= -90 and .lat <= 90 and .lng >= -180 and .lng <= 180 then "valid" else "invalid" end'

# Preview execution input
cat examples/input.edinburgh.json | jq .
```

## Contributing New Inputs

When adding new input files:

1. Follow the naming convention: `input.{identifier}.json`
2. Validate JSON structure with `jq`
3. Test the input with a small execution first
4. Document the input's purpose and characteristics
5. Update this README with the new input details

## Related Documentation

- [Integration Testing Guide](../INTEGRATION_TESTING.md)
- [DLQ Runbook](../DLQ_RUNBOOK.md)
- [Feature Flags Documentation](../terraform/FEATURE_FLAGS.md)
- [Step Functions Workflow](../README.md)
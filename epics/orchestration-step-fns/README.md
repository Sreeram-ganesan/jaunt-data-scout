# Epic: Orchestration — Step Functions workflow for city discovery

This directory contains the complete implementation of the Step Functions orchestration workflow for city discovery, including infrastructure as code, testing tools, and comprehensive documentation.

## Overview

The Step Functions workflow orchestrates the entire data pipeline:
- **Discovery Phase**: DiscoverWebSources → DiscoverTargets → SeedPrimaries → ExpandNeighbors → TileSweep
- **Processing Phase**: WebFetch → ExtractWithLLM → GeocodeValidate
- **Finalization Phase**: DedupeCanonicalize → Persist → Rank → Finalize

## Directory Structure

```
epics/orchestration-step-fns/
├── terraform/              # Infrastructure as Code
│   ├── sfn/                # Step Functions module
│   ├── envs/               # Environment configurations
│   └── FEATURE_FLAGS.md    # Feature flags documentation
├── go/                     # Go test framework and utilities
├── examples/               # Golden input datasets for testing
├── tools/                  # Helper scripts and utilities
│   ├── dlq-helper.sh      # DLQ management tool
│   └── e2e-integration-test.sh # End-to-end testing
├── INTEGRATION_TESTING.md  # Complete testing guide
├── DLQ_RUNBOOK.md          # DLQ troubleshooting guide
└── README.md               # This file
```

## Quick Start

### 1. Deploy Infrastructure

```bash
cd terraform

# Initialize Terraform
make init ENV=mock

# Deploy with mock configuration
make plan ENV=mock
make apply ENV=mock

# Get deployment outputs
make outputs ENV=mock
```

### 2. Deploy Mock Lambda (for testing)

```bash
cd ../../lambdas/mock-go

# Build and deploy mock Lambda
make deploy ARCH=arm64 REGION=us-east-1
make arn  # Get the Lambda ARN for configuration
```

### 3. Run Integration Test

```bash
cd ../epics/orchestration-step-fns

# Run complete end-to-end test
./tools/e2e-integration-test.sh --city edinburgh --env mock --verbose
```

## Key Features

### ✅ Complete Infrastructure
- Step Functions state machine with all workflow states
- SQS frontier queue + Dead Letter Queue with redrive policy
- S3 storage for raw cache, extracted data, and manifests
- IAM roles and policies with least privilege access
- CloudWatch logging and X-Ray tracing (configurable)

### ✅ Feature Flags System
- Toggle individual states between mock and real implementations
- Support for gradual rollout and mixed deployments
- Environment-specific configurations
- Easy state-by-state migration path

### ✅ Comprehensive Testing
- Golden input datasets for multiple cities and scenarios
- End-to-end integration testing scripts
- Budget exhaustion and early stopping tests
- Kill switch and error handling validation
- DLQ monitoring and reprocessing tools

### ✅ Production Ready
- Budget controls and early stopping logic
- Error handling and retry mechanisms
- Observability with metrics and tracing
- Dead Letter Queue management
- Comprehensive documentation and runbooks

## Integration Testing

### Quick Integration Test

```bash
# Deploy infrastructure and run basic test
cd terraform && make apply ENV=mock
cd .. && ./tools/e2e-integration-test.sh --city edinburgh
```

### Advanced Testing Scenarios

```bash
# Test budget constraints
./tools/e2e-integration-test.sh --city budget-test-small --verbose

# Test early stopping
./tools/e2e-integration-test.sh --city early-stop-test --timeout 3600

# Test kill switches
./tools/e2e-integration-test.sh --city kill-switch-test --env dev
```

## Feature Flags Usage

### All Mock Configuration

```bash
cd terraform
make apply ENV=all-mock  # Uses envs/all-mock.tfvars
```

### Mixed Mock/Real Configuration

```bash
# Edit envs/mixed-mock-real.tfvars to configure which states use real vs mock
cd terraform
make apply ENV=mixed-mock-real
```

### All Real (Production) Configuration

```bash
cd terraform
make apply ENV=all-real  # Uses envs/all-real.tfvars
```

## DLQ Management

### Monitor DLQ

```bash
# Check DLQ status
./tools/dlq-helper.sh status

# Peek at failed messages
./tools/dlq-helper.sh peek --count 10

# Analyze failure patterns
./tools/dlq-helper.sh analyze
```

### Reprocess Failed Messages

```bash
# Redrive all messages back to frontier queue
./tools/dlq-helper.sh redrive-all

# Redrive specific number of messages
./tools/dlq-helper.sh redrive-batch --count 50

# Monitor DLQ in real-time
./tools/dlq-helper.sh monitor --interval 30
```

## Available Test Datasets

| Dataset | Use Case | Duration | Characteristics |
|---------|----------|----------|-----------------|
| `input.edinburgh.json` | Primary development | 15-30 min | Medium city, comprehensive test |
| `input.london.json` | Large city testing | 30-60 min | High density, large data volume |
| `input.tokyo.json` | International testing | 30-45 min | Very high density, non-English |
| `input.new-york.json` | Large scale testing | 45-90 min | Very large area, high volume |
| `input.budget-test-small.json` | Budget validation | 2-5 min | Low budgets, early termination |
| `input.early-stop-test.json` | Early stop logic | 5-15 min | High unique rate threshold |
| `input.kill-switch-test.json` | Service isolation | 10-20 min | Some services disabled |

## Deployment Environments

### Mock Environment
- **Purpose**: Development and basic testing
- **Configuration**: All states use mock Lambda
- **Cost**: Minimal (only mock Lambda invocations)
- **Use**: Feature development, workflow testing

### Mixed Environment  
- **Purpose**: Gradual real service integration
- **Configuration**: Some states real, others mock
- **Cost**: Variable based on real services used
- **Use**: Progressive rollout, service-specific testing

### Production Environment
- **Purpose**: Full production deployment
- **Configuration**: All states use real implementations
- **Cost**: Full operational cost
- **Use**: Production workloads, performance testing

## Monitoring and Observability

### Step Functions Monitoring

```bash
# View execution status
aws stepfunctions describe-execution --execution-arn "EXECUTION_ARN"

# Get execution history
aws stepfunctions get-execution-history --execution-arn "EXECUTION_ARN"

# Monitor via CloudWatch
aws logs tail "/aws/stepfunctions/data-scout-orchestration-step-function" --follow
```

### Queue Monitoring

```bash
# Check queue depths
aws sqs get-queue-attributes \
  --queue-url "FRONTIER_QUEUE_URL" \
  --attribute-names ApproximateNumberOfMessages

# Monitor DLQ
./tools/dlq-helper.sh status
```

### S3 Output Monitoring

```bash
# Check S3 outputs
aws s3 ls s3://YOUR-BUCKET/cities/ --recursive

# Monitor data creation
aws s3 ls s3://YOUR-BUCKET/manifests/ --recursive
```

## Troubleshooting

### Common Issues

1. **Execution Fails Immediately**
   - Check Lambda function permissions
   - Verify Lambda ARNs in configuration
   - Review Step Functions execution role

2. **Messages in DLQ**
   - Use DLQ analysis tools: `./tools/dlq-helper.sh analyze`
   - Check Lambda function logs
   - Review message format and schemas

3. **Budget Exhaustion**
   - Review budget settings in input files
   - Monitor actual vs expected usage
   - Adjust budgets based on city size

4. **Timeout Issues**
   - Increase wall clock timeout in input
   - Check individual Lambda timeouts
   - Monitor execution duration patterns

### Debug Commands

```bash
# Validate configuration
cd terraform && make validate

# Test Go utilities
cd go && make test

# Check AWS connectivity
aws sts get-caller-identity

# Validate JSON inputs
jq . examples/input.edinburgh.json
```

## Performance Guidelines

### Execution Times (Approximate)

| City Size | Mock Mode | Mixed Mode | Full Real |
|-----------|-----------|------------|-----------|
| Small (Edinburgh) | 5-10 min | 15-30 min | 30-60 min |
| Medium (Tokyo) | 7-15 min | 20-40 min | 45-90 min |
| Large (London/NYC) | 10-20 min | 30-60 min | 60-120 min |

### Resource Usage

- **Lambda Concurrency**: 100-1000 concurrent executions
- **SQS Throughput**: 1000-10000 messages/minute
- **S3 Storage**: 1-100 GB per city execution
- **Step Functions**: 1 execution per city job

## Security Considerations

- IAM roles follow least privilege principle
- All data encrypted in transit and at rest
- No secrets in configuration files
- VPC endpoints for enhanced security (optional)
- CloudTrail logging for audit trails

## Cost Optimization

- Use mock mode for development to minimize costs
- Implement appropriate budget controls
- Monitor and alert on unexpected usage
- Use S3 lifecycle policies for data retention
- Right-size Lambda memory and timeout settings

## Contributing

### Adding New Test Datasets

1. Create new input file in `examples/` directory
2. Follow naming convention: `input.{identifier}.json`
3. Validate JSON structure and coordinate ranges
4. Test with small execution first
5. Update documentation

### Extending Infrastructure

1. Add new Terraform resources in appropriate modules
2. Update variable definitions and documentation
3. Test with multiple environments
4. Ensure backwards compatibility

### Adding New Tools

1. Create scripts in `tools/` directory
2. Follow existing patterns for CLI interfaces
3. Include comprehensive help and error handling
4. Make scripts executable and well-documented

## References

- **Epic Issue**: https://github.com/Sreeram-ganesan/jaunt-data-scout/issues/3
- **Integration Testing Guide**: [INTEGRATION_TESTING.md](INTEGRATION_TESTING.md)
- **DLQ Management**: [DLQ_RUNBOOK.md](DLQ_RUNBOOK.md)
- **Feature Flags**: [terraform/FEATURE_FLAGS.md](terraform/FEATURE_FLAGS.md)
- **Project Overview**: [../../Project.md](../../Project.md)

## Status

✅ **Epic Complete**: All acceptance criteria met
- [x] Complete Step Functions workflow implementation
- [x] Feature flags for mock↔real state toggling
- [x] Comprehensive integration testing framework
- [x] DLQ management tools and runbooks
- [x] Golden input datasets for multiple test scenarios
- [x] Production-ready deployment configurations
- [x] Full documentation and operational guides

This implementation provides a complete, production-ready orchestration system with comprehensive testing, monitoring, and operational capabilities.
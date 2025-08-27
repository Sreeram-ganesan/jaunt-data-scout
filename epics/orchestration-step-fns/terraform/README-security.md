# Security Configuration - Secrets, IAM, and Egress Control

This directory contains the security infrastructure for the Jaunt Data Scout project, including secrets management, IAM roles, and network egress controls.

## Files Overview

### Terraform Infrastructure
- `secrets.tf` - AWS Secrets Manager configuration for API keys and credentials
- `iam_lambda.tf` - Least-privilege IAM roles for Lambda functions
- `vpc.tf` - VPC configuration for network egress control (optional)
- `s3.tf` - Enhanced S3 security with access logging and encryption

### Documentation
- `docs/secrets-management.md` - Comprehensive secrets management and rotation policy

## Quick Start

### 1. Deploy Basic Security (No VPC)
```bash
# Deploy with default settings (no VPC egress control)
terraform plan -var-file=envs/dev.tfvars
terraform apply -var-file=envs/dev.tfvars
```

### 2. Deploy with VPC Egress Control
```bash
# Deploy with VPC-based egress filtering
terraform plan -var-file=envs/dev.tfvars -var="enable_vpc_egress_control=true"
terraform apply -var-file=envs/dev.tfvars -var="enable_vpc_egress_control=true"
```

### 3. Deploy with KMS Encryption
```bash
# Deploy with KMS encryption for S3
terraform plan -var-file=envs/dev.tfvars -var="enable_kms_encryption=true"
terraform apply -var-file=envs/dev.tfvars -var="enable_kms_encryption=true"
```

## Security Features Implemented

### ✅ Secrets Management
- **AWS Secrets Manager**: All API keys stored securely
- **Automatic Rotation**: 30-day rotation for API keys, 90-day for database
- **Least Privilege Access**: Each Lambda role accesses only required secrets
- **Encryption**: All secrets encrypted with KMS

### ✅ IAM Roles (Least Privilege)
- **Web Fetch Role**: S3 raw/html access + Tavily API key
- **Extract LLM Role**: S3 extracted access + LLM API keys + Bedrock permissions  
- **Google Places Role**: S3 google raw access + Google API key
- **Open Data Role**: S3 open-data raw access + OTM API key
- **Database Role**: S3 manifests/extracted access + Database credentials
- **Discovery Role**: S3 read-only access for manifests
- **Geocode Role**: Google API key for geocoding validation

### ✅ S3 Security Enhancements
- **Server-Side Encryption**: AES256 by default, optional KMS
- **Access Logging**: All S3 access logged to separate bucket
- **Public Access Blocked**: All buckets have public access blocked
- **Lifecycle Policies**: Automatic cleanup based on compliance requirements

### ✅ VPC Egress Control (Optional)
- **Private Subnets**: Lambda functions isolated in private subnets
- **NAT Gateway**: Controlled outbound internet access
- **Security Groups**: Egress allowlists for HTTPS, HTTP (robots.txt), DNS
- **VPC Endpoints**: Direct access to AWS services (S3, Secrets Manager)
- **DNS Filtering**: Only approved domains accessible

## Compliance Features

### Google Maps Platform
- ✅ API key rotation policy
- ✅ Field masks enforcement (implemented in application code)
- ✅ 30-day retention for raw Google data
- ✅ Attribution capture requirements

### Web Crawling
- ✅ Robots.txt compliance (enforced in fetcher code)
- ✅ Per-domain rate limiting
- ✅ User-Agent headers
- ✅ 90-day HTML retention limit

### LLM Safety
- ✅ API key rotation and scoping
- ✅ Token/cost budget enforcement
- ✅ Multiple provider support
- ✅ No secrets in training data

## Configuration Variables

### Security Options
```hcl
# Enable KMS encryption for S3 (default: AES256)
enable_kms_encryption = true
kms_key_id = "alias/data-scout-kms-key"  # optional

# Enable VPC egress control (default: false)
enable_vpc_egress_control = true
vpc_cidr = "10.0.0.0/16"  # optional
```

### Environment Files
- `envs/dev.tfvars` - Development environment
- `envs/prod.tfvars` - Production environment (use VPC + KMS)
- `envs/off-by-default.tfvars` - Minimal security for testing

## Secrets Population

After infrastructure deployment, populate secrets:

```bash
# Google Maps API Key
aws secretsmanager update-secret --secret-id jaunt-dev-google-api-key \
  --secret-string '{"api_key":"your-google-api-key","service":"google-maps-platform","scopes":["places","geocoding"]}'

# Tavily API Key  
aws secretsmanager update-secret --secret-id jaunt-dev-tavily-api-key \
  --secret-string '{"api_key":"your-tavily-api-key","service":"tavily-search"}'

# Database Credentials
aws secretsmanager update-secret --secret-id jaunt-dev-database-credentials \
  --secret-string '{"host":"db.example.com","port":5432,"username":"dbuser","password":"secure-password","database":"jaunt","ssl_mode":"require"}'

# Continue for other secrets...
```

## Monitoring and Alerting

### CloudWatch Metrics
- Secret rotation success/failure rates
- Unauthorized access attempts  
- Lambda function secret access patterns
- S3 access patterns and anomalies

### Security Scanning
- Regular IAM policy audits
- Secret rotation compliance checks
- VPC security group rule validation
- S3 bucket policy analysis

## Testing

### Security Validation
```bash
# Test secret access from Lambda
aws lambda invoke --function-name test-function response.json

# Validate IAM permissions
aws sts get-caller-identity
aws secretsmanager describe-secret --secret-id jaunt-dev-google-api-key

# Test VPC connectivity (if enabled)
aws ec2 describe-security-groups --group-names "jaunt-dev-lambda-egress*"
```

### Rotation Testing
```bash
# Trigger manual rotation
aws secretsmanager rotate-secret --secret-id jaunt-dev-google-api-key

# Verify Lambda functions still work after rotation
# Run integration tests
```

## Troubleshooting

### Common Issues

1. **Secret Access Denied**
   - Verify Lambda role has `secretsmanager:GetSecretValue` permission
   - Check secret ARN in IAM policy matches actual secret

2. **VPC Connectivity Issues**
   - Ensure NAT Gateway is properly configured
   - Verify security group egress rules allow HTTPS
   - Check VPC endpoints are created correctly

3. **S3 Access Issues**
   - Verify bucket policy allows Lambda role access
   - Check S3 VPC endpoint configuration if using VPC

### Debug Commands
```bash
# Check Lambda role permissions
aws iam get-role-policy --role-name jaunt-dev-lambda-web-fetch-role --policy-name web-fetch-policy

# Test secret retrieval
aws secretsmanager get-secret-value --secret-id jaunt-dev-google-api-key

# Check VPC configuration
aws ec2 describe-vpcs --filters "Name=tag:Name,Values=jaunt-dev-lambda-vpc"
```

## Next Steps

1. **Deploy Infrastructure**: Use Terraform to deploy security components
2. **Populate Secrets**: Add real API keys and credentials to Secrets Manager
3. **Configure Lambda Functions**: Update Lambda functions to use new IAM roles
4. **Enable Monitoring**: Set up CloudWatch alarms and dashboards
5. **Test End-to-End**: Validate full workflow with security enabled

## Related Documentation

- [Secrets Management Policy](../docs/secrets-management.md) - Detailed rotation and compliance procedures
- [Observability Guide](../docs/observability.md) - Monitoring and alerting setup
- [Configuration Guide](../docs/configuration.md) - Budget and configuration management
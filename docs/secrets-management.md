# Secrets Management and Rotation Policy

## Overview

This document describes the secrets management strategy for the Jaunt Data Scout project, including rotation policies, access patterns, and security best practices.

## Secrets Inventory

The system manages the following secrets in AWS Secrets Manager:

### API Keys and Credentials

1. **Google Maps API Key** (`jaunt-{env}-google-api-key`)
   - Purpose: Google Places API (Text, Nearby, Details) and Geocoding
   - Rotation: 30 days (automatic)
   - Scopes: Places API, Geocoding API
   - Access: Google Places Lambda functions only

2. **Tavily API Key** (`jaunt-{env}-tavily-api-key`)
   - Purpose: Web search and discovery services
   - Rotation: 30 days (automatic)
   - Access: WebFetch Lambda functions only

3. **OpenTripMap API Key** (`jaunt-{env}-otm-api-key`)
   - Purpose: Points of interest data from OpenTripMap
   - Rotation: 30 days (automatic)
   - Access: Open Data Lambda functions only

4. **Database Credentials** (`jaunt-{env}-database-credentials`)
   - Purpose: PostgreSQL database connection
   - Rotation: 90 days (manual process)
   - Contains: host, port, username, password, database name, SSL config
   - Access: Database Lambda functions (Persist, Rank, Finalize)

5. **LLM API Keys** (`jaunt-{env}-llm-api-keys`)
   - Purpose: LLM providers (OpenAI, Anthropic, Bedrock)
   - Rotation: 30 days (automatic where supported)
   - Access: Extract with LLM Lambda functions only

## Rotation Schedule and Procedures

### Automatic Rotation (30 days)
- **API Keys**: Google, Tavily, OTM, LLM providers
- **Process**: AWS Secrets Manager automatic rotation
- **Monitoring**: CloudWatch alarms on rotation failures
- **Rollback**: Previous version retained for 24 hours

### Manual Rotation (90 days)
- **Database Credentials**: Requires coordinated update
- **Process**:
  1. Create new credentials in database
  2. Update secret in AWS Secrets Manager
  3. Test connectivity with new credentials
  4. Remove old credentials from database
  5. Verify all Lambda functions can connect

## Access Control and IAM Policies

### Least Privilege Access
Each Lambda function role has access only to the secrets it requires:

- **Web Fetch Functions**: Tavily API key only
- **LLM Extract Functions**: LLM API keys only  
- **Google Places Functions**: Google API key only
- **Open Data Functions**: OTM API key only
- **Database Functions**: Database credentials only

### IAM Policy Structure
```hcl
# Example: Lambda function can only read specific secrets
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "secretsmanager:GetSecretValue",
        "secretsmanager:DescribeSecret"
      ],
      "Resource": "arn:aws:secretsmanager:region:account:secret:jaunt-env-service-*"
    }
  ]
}
```

## Security Best Practices

### Encryption
- **At Rest**: All secrets encrypted with AWS KMS
- **In Transit**: TLS 1.2+ for all API calls
- **In Memory**: Secrets cached for minimal time, cleared after use

### Network Security
- **VPC Endpoints**: Secrets Manager accessed via VPC endpoint when VPC is enabled
- **Egress Control**: Lambda functions in VPC with restricted outbound access
- **DNS Filtering**: Only approved domains accessible for API calls

### Monitoring and Alerting
- **Failed Rotations**: CloudWatch alarms on rotation failures
- **Unauthorized Access**: CloudTrail logging of all secret access
- **Usage Patterns**: Monitoring for unusual access patterns

## Compliance Features

### Google Maps Platform Compliance
- **Field Masks**: Enforced in Google API calls
- **Attribution**: Captured and stored with responses
- **Retention**: Raw Google data deleted after 30 days
- **Rate Limiting**: API key usage within Google's terms

### Web Crawling Compliance
- **Robots.txt**: Checked before fetching any URL
- **User-Agent**: Proper identification in HTTP headers
- **Rate Limiting**: Per-domain crawl delays enforced
- **ToS Compliance**: Headers and behavior respect website terms

### LLM Provider Compliance
- **Token Limits**: Per-request and per-day caps enforced
- **Content Filtering**: No storage of raw user secrets in prompts
- **Provider Rotation**: Support for multiple LLM providers for redundancy

## Emergency Procedures

### Compromised API Key
1. **Immediate**: Rotate secret in AWS Secrets Manager
2. **Verify**: Check CloudTrail logs for unauthorized usage
3. **Audit**: Review all recent API calls with old key
4. **Report**: Document incident and update security measures

### Database Credential Compromise
1. **Immediate**: Change database password
2. **Update**: Update secret in AWS Secrets Manager
3. **Test**: Verify all Lambda functions can reconnect
4. **Audit**: Review database access logs
5. **Rotate**: Force rotation of all other database users

### Service Provider Breach
1. **Assess**: Determine scope of potential exposure
2. **Rotate**: Immediately rotate all affected API keys
3. **Monitor**: Enhanced monitoring for unusual activity
4. **Document**: Update incident response procedures

## Monitoring and Metrics

### CloudWatch Metrics
- `SecretRotationSuccess`: Successful rotations per secret
- `SecretRotationFailure`: Failed rotations requiring intervention
- `SecretAccessCount`: Number of times each secret is accessed
- `UnauthorizedSecretAccess`: Failed attempts to access secrets

### CloudTrail Events
- All `GetSecretValue` calls logged
- Failed authentication attempts tracked
- Secret rotation events recorded
- Policy changes monitored

## Cost Management

### Optimization
- **Secret Consolidation**: Related keys stored in single secret where appropriate
- **Access Patterns**: Monitor usage to optimize rotation frequency
- **Regional**: Secrets stored in same region as Lambda functions

### Budget Alerts
- Monthly cost tracking for Secrets Manager usage
- Alerts when rotation frequency impacts costs
- Regular review of secret inventory for unused entries

## Testing and Validation

### Rotation Testing
- Automated tests for each rotation scenario
- Validation of Lambda function connectivity after rotation
- Rollback procedures tested quarterly

### Security Testing
- Regular penetration testing of secrets access
- IAM policy validation with least privilege principle
- Network security testing for VPC-enabled functions

## Implementation Notes

### Terraform Configuration
- All secrets created with automatic rotation enabled
- IAM policies follow least privilege principle
- KMS encryption enabled for all secrets
- Lifecycle management prevents accidental deletion

### Lambda Integration
- Secrets cached for minimal time in Lambda memory
- Error handling for secret retrieval failures
- Graceful degradation when secrets are rotating

### Deployment
- Secrets created before Lambda functions
- Initial secret values set to placeholders
- Production secrets populated through secure deployment process
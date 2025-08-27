# Observability feature flags (OFF by default)

The following variables gate observability resources. Defaults are safe to apply in all environments:

- enable_observability = false
- enable_cloudwatch_dashboards = false
- enable_cloudwatch_alarms = false
- sfn_log_level = "OFF" (options: OFF, ERROR, ALL)
- sfn_include_execution_data = false
- enable_sfn_tracing = false
- sfn_log_group_arn = "" (set when sfn_log_level != OFF)

## Enabling Step Functions Logging and Tracing

To enable Step Functions logging and tracing, set in your tfvars:

```hcl
sfn_log_level                  = "ERROR"
sfn_include_execution_data     = true
enable_sfn_tracing             = true
sfn_log_group_arn              = "arn:aws:logs:REGION:ACCOUNT_ID:log-group:/aws/states/your-sfn-log-group:*"
```

Note: When `sfn_log_level != "OFF"`, a CloudWatch log group will be automatically created and used if `sfn_log_group_arn` is not specified.

## Enabling Dashboards and Alarms

Dashboards and alarms remain disabled unless you explicitly set:

```hcl
enable_observability         = true
enable_cloudwatch_dashboards = true
enable_cloudwatch_alarms     = true
```

**Note:** The observability module (dashboards/alarms) has pre-existing configuration issues and is currently commented out in main.tf. The core observability features (SFN logging, tracing, and IAM permissions) are fully functional.

## Example Files

- `envs/off-by-default.tfvars`: Safe defaults (all observability OFF)  
- `envs/observability-enabled.tfvars`: Example with observability features enabled

## Implementation Details

- **CloudWatch Log Group**: Only created when `sfn_log_level != "OFF"`
- **SFN Logging Configuration**: Only added to state machine when `sfn_log_level != "OFF"`
- **SFN Tracing Configuration**: Controlled by `enable_sfn_tracing` (defaults to false)
- **IAM Policies**: Logs and X-Ray permissions only attached when respective features are enabled
- **Backward Compatibility**: Existing functionality unchanged; only defaults modified to be OFF
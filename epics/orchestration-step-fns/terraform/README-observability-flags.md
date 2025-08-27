# Observability feature flags (OFF by default)

The following variables gate observability resources. Defaults are safe to apply in all environments:

- enable_observability = false
- enable_cloudwatch_dashboards = false
- enable_cloudwatch_alarms = false
- sfn_log_level = "OFF" (options: OFF, ERROR, ALL)
- sfn_include_execution_data = false
- enable_sfn_tracing = false
- sfn_log_group_arn = "" (set when sfn_log_level != OFF)

To enable Step Functions logging and tracing, set in your tfvars:

```hcl
sfn_log_level                  = "ERROR"
sfn_include_execution_data     = true
enable_sfn_tracing             = true
sfn_log_group_arn              = "arn:aws:logs:REGION:ACCOUNT_ID:log-group:/aws/states/your-sfn-log-group:*"
```

Dashboards and alarms remain disabled unless you explicitly set:

```hcl
enable_observability         = true
enable_cloudwatch_dashboards = true
enable_cloudwatch_alarms     = true
```
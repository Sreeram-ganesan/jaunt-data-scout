# Safe defaults: no logging, no tracing, no dashboards, no alarms
enable_observability         = false
enable_cloudwatch_dashboards = false
enable_cloudwatch_alarms     = false
sfn_log_level                = "OFF"
sfn_include_execution_data   = false
enable_sfn_tracing           = false

# When you later enable logging, set this to a real Log Group ARN, e.g.:
# sfn_log_group_arn = "arn:aws:logs:us-east-1:123456789012:log-group:/aws/states/jaunt-orchestrator:*"
# Example: observability features ENABLED
enable_observability         = true
enable_cloudwatch_dashboards = true
enable_cloudwatch_alarms     = true
sfn_log_level                = "ERROR"
sfn_include_execution_data   = true
enable_sfn_tracing           = true

# Mock Lambda ARNs for testing (replace with your actual ARNs)
lambda_discover_web_sources_arn = "arn:aws:lambda:us-east-1:123456789012:function:mock"
lambda_discover_targets_arn     = "arn:aws:lambda:us-east-1:123456789012:function:mock"
lambda_seed_primaries_arn       = "arn:aws:lambda:us-east-1:123456789012:function:mock"
lambda_expand_neighbors_arn     = "arn:aws:lambda:us-east-1:123456789012:function:mock"
lambda_tile_sweep_arn           = "arn:aws:lambda:us-east-1:123456789012:function:mock"
lambda_web_fetch_arn            = "arn:aws:lambda:us-east-1:123456789012:function:mock"
lambda_extract_with_llm_arn     = "arn:aws:lambda:us-east-1:123456789012:function:mock"
lambda_geocode_validate_arn     = "arn:aws:lambda:us-east-1:123456789012:function:mock"
lambda_dedupe_canonicalize_arn  = "arn:aws:lambda:us-east-1:123456789012:function:mock"
lambda_persist_arn              = "arn:aws:lambda:us-east-1:123456789012:function:mock"
lambda_rank_arn                 = "arn:aws:lambda:us-east-1:123456789012:function:mock"

# Note: sfn_log_group_arn will be automatically set to the CloudWatch log group created by terraform
# You can also override it with a custom log group ARN if needed:
# sfn_log_group_arn = "arn:aws:logs:us-east-1:123456789012:log-group:/aws/states/custom-log-group:*"
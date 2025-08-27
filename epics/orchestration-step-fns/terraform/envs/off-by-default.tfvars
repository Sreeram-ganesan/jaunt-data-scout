# Safe defaults: no logging, no tracing, no dashboards, no alarms
enable_observability         = false
enable_cloudwatch_dashboards = false
enable_cloudwatch_alarms     = false
sfn_log_level                = "OFF"
sfn_include_execution_data   = false
enable_sfn_tracing           = false

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

# When you later enable logging, set this to a real Log Group ARN, e.g.:
# sfn_log_group_arn = "arn:aws:logs:us-east-1:123456789012:log-group:/aws/states/jaunt-orchestrator:*"
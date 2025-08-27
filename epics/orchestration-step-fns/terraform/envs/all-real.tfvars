# Example: All states use real implementations
# Production-ready configuration with all real Lambda functions

# Mock Lambda ARN (not used but required by schema)
mock_lambda_arn = "arn:aws:lambda:us-east-1:YOUR_ACCOUNT:function:unused-mock"

# All states configured to use real implementations
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

# Real Lambda ARNs for production deployment
lambda_discover_web_sources_arn = "arn:aws:lambda:us-east-1:YOUR_ACCOUNT:function:jaunt-prod-discover-web-sources"
lambda_discover_targets_arn     = "arn:aws:lambda:us-east-1:YOUR_ACCOUNT:function:jaunt-prod-discover-targets"
lambda_seed_primaries_arn       = "arn:aws:lambda:us-east-1:YOUR_ACCOUNT:function:jaunt-prod-seed-primaries"
lambda_expand_neighbors_arn     = "arn:aws:lambda:us-east-1:YOUR_ACCOUNT:function:jaunt-prod-expand-neighbors"
lambda_tile_sweep_arn           = "arn:aws:lambda:us-east-1:YOUR_ACCOUNT:function:jaunt-prod-tile-sweep"
lambda_web_fetch_arn            = "arn:aws:lambda:us-east-1:YOUR_ACCOUNT:function:jaunt-prod-web-fetch"
lambda_extract_with_llm_arn     = "arn:aws:lambda:us-east-1:YOUR_ACCOUNT:function:jaunt-prod-extract-with-llm"
lambda_geocode_validate_arn     = "arn:aws:lambda:us-east-1:YOUR_ACCOUNT:function:jaunt-prod-geocode-validate"
lambda_dedupe_canonicalize_arn  = "arn:aws:lambda:us-east-1:YOUR_ACCOUNT:function:jaunt-prod-dedupe-canonicalize"
lambda_persist_arn              = "arn:aws:lambda:us-east-1:YOUR_ACCOUNT:function:jaunt-prod-persist"
lambda_rank_arn                 = "arn:aws:lambda:us-east-1:YOUR_ACCOUNT:function:jaunt-prod-rank"

# Production configuration
project_prefix = "jaunt"
environment    = "prod"
aws_region     = "us-east-1"

# Enable observability for production
sfn_log_level              = "ERROR"
sfn_include_execution_data = true
enable_sfn_tracing         = true

tags = {
  Owner       = "sreeram"
  environment = "prod"
  project     = "data-scout"
}
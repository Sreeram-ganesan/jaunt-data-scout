# Point all states to a single mock Lambda ARN via feature flags
mock_lambda_arn = "arn:aws:lambda:us-east-1:253490763512:function:jaunt-mock"

# Configure all states to use mock implementation
state_implementations = {
  discover_web_sources   = "mock"
  discover_targets       = "mock"
  seed_primaries        = "mock"
  expand_neighbors      = "mock"
  tile_sweep           = "mock"
  web_fetch            = "mock"
  extract_with_llm     = "mock"
  geocode_validate     = "mock"
  dedupe_canonicalize  = "mock"
  persist              = "mock"
  rank                 = "mock"
}

# Real Lambda ARNs (placeholders since all states use mock)
lambda_discover_web_sources_arn = "arn:aws:lambda:us-east-1:253490763512:function:placeholder"
lambda_discover_targets_arn     = "arn:aws:lambda:us-east-1:253490763512:function:placeholder"
lambda_seed_primaries_arn       = "arn:aws:lambda:us-east-1:253490763512:function:placeholder"
lambda_expand_neighbors_arn     = "arn:aws:lambda:us-east-1:253490763512:function:placeholder"
lambda_tile_sweep_arn           = "arn:aws:lambda:us-east-1:253490763512:function:placeholder"
lambda_web_fetch_arn            = "arn:aws:lambda:us-east-1:253490763512:function:placeholder"
lambda_extract_with_llm_arn     = "arn:aws:lambda:us-east-1:253490763512:function:placeholder"
lambda_geocode_validate_arn     = "arn:aws:lambda:us-east-1:253490763512:function:placeholder"
lambda_dedupe_canonicalize_arn  = "arn:aws:lambda:us-east-1:253490763512:function:placeholder"
lambda_persist_arn              = "arn:aws:lambda:us-east-1:253490763512:function:placeholder"
lambda_rank_arn                 = "arn:aws:lambda:us-east-1:253490763512:function:placeholder"

# Typical dev settings (override as needed)
project_prefix = "jaunt"
environment    = "dev"
aws_region     = "us-east-1"

# Optional: override names to avoid collisions in your account
# s3_raw_cache_bucket_name = "jaunt-data-scout-raw-cache-dev"
# sqs_frontier_name        = "jaunt-orchestration-frontier-dev"
# sqs_dlq_name             = "jaunt-orchestration-frontier-dlq-dev"

tags = {
  Owner       = "sreeram"
  environment = "dev"
  project     = "data-scout"
}
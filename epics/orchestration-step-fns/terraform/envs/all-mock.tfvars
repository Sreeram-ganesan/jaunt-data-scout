# Example: All states use mock implementation
# This configuration routes ALL states to a single mock Lambda function

# Mock Lambda ARN (replace with your deployed mock Lambda)
mock_lambda_arn = "arn:aws:lambda:us-east-1:YOUR_ACCOUNT:function:jaunt-mock"

# All states configured as mock
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

# Real Lambda ARNs (required but not used when all states are mock)
lambda_discover_web_sources_arn = "arn:aws:lambda:us-east-1:YOUR_ACCOUNT:function:placeholder"
lambda_discover_targets_arn     = "arn:aws:lambda:us-east-1:YOUR_ACCOUNT:function:placeholder"
lambda_seed_primaries_arn       = "arn:aws:lambda:us-east-1:YOUR_ACCOUNT:function:placeholder"
lambda_expand_neighbors_arn     = "arn:aws:lambda:us-east-1:YOUR_ACCOUNT:function:placeholder"
lambda_tile_sweep_arn           = "arn:aws:lambda:us-east-1:YOUR_ACCOUNT:function:placeholder"
lambda_web_fetch_arn            = "arn:aws:lambda:us-east-1:YOUR_ACCOUNT:function:placeholder"
lambda_extract_with_llm_arn     = "arn:aws:lambda:us-east-1:YOUR_ACCOUNT:function:placeholder"
lambda_geocode_validate_arn     = "arn:aws:lambda:us-east-1:YOUR_ACCOUNT:function:placeholder"
lambda_dedupe_canonicalize_arn  = "arn:aws:lambda:us-east-1:YOUR_ACCOUNT:function:placeholder"
lambda_persist_arn              = "arn:aws:lambda:us-east-1:YOUR_ACCOUNT:function:placeholder"
lambda_rank_arn                 = "arn:aws:lambda:us-east-1:YOUR_ACCOUNT:function:placeholder"

# Standard configuration
project_prefix = "jaunt"
environment    = "all-mock"
aws_region     = "us-east-1"

tags = {
  Owner       = "sreeram"
  environment = "all-mock"
  project     = "data-scout"
}
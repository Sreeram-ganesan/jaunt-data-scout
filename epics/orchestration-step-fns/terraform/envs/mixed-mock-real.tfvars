# Example: Mixed mock and real implementations
# This configuration demonstrates selective use of real vs mock implementations

# Mock Lambda ARN for states configured as mock
mock_lambda_arn = "arn:aws:lambda:us-east-1:YOUR_ACCOUNT:function:jaunt-mock"

# Mixed configuration: some states use real implementations, others use mock
state_implementations = {
  discover_web_sources   = "real"    # Use real web source discovery
  discover_targets       = "real"    # Use real target discovery
  seed_primaries        = "real"    # Use real primary seeding
  expand_neighbors      = "mock"    # Mock neighbor expansion
  tile_sweep           = "mock"    # Mock tile sweep
  web_fetch            = "real"    # Use real web fetching
  extract_with_llm     = "real"    # Use real LLM extraction
  geocode_validate     = "real"    # Use real geocoding
  dedupe_canonicalize  = "mock"    # Mock deduplication
  persist              = "real"    # Use real persistence
  rank                 = "mock"    # Mock ranking
}

# Real Lambda ARNs (used for states configured as "real")
lambda_discover_web_sources_arn = "arn:aws:lambda:us-east-1:YOUR_ACCOUNT:function:jaunt-discover-web-sources"
lambda_discover_targets_arn     = "arn:aws:lambda:us-east-1:YOUR_ACCOUNT:function:jaunt-discover-targets"
lambda_seed_primaries_arn       = "arn:aws:lambda:us-east-1:YOUR_ACCOUNT:function:jaunt-seed-primaries"
lambda_expand_neighbors_arn     = "arn:aws:lambda:us-east-1:YOUR_ACCOUNT:function:jaunt-expand-neighbors"
lambda_tile_sweep_arn           = "arn:aws:lambda:us-east-1:YOUR_ACCOUNT:function:jaunt-tile-sweep"
lambda_web_fetch_arn            = "arn:aws:lambda:us-east-1:YOUR_ACCOUNT:function:jaunt-web-fetch"
lambda_extract_with_llm_arn     = "arn:aws:lambda:us-east-1:YOUR_ACCOUNT:function:jaunt-extract-with-llm"
lambda_geocode_validate_arn     = "arn:aws:lambda:us-east-1:YOUR_ACCOUNT:function:jaunt-geocode-validate"
lambda_dedupe_canonicalize_arn  = "arn:aws:lambda:us-east-1:YOUR_ACCOUNT:function:jaunt-dedupe-canonicalize"
lambda_persist_arn              = "arn:aws:lambda:us-east-1:YOUR_ACCOUNT:function:jaunt-persist"
lambda_rank_arn                 = "arn:aws:lambda:us-east-1:YOUR_ACCOUNT:function:jaunt-rank"

# Standard configuration
project_prefix = "jaunt"
environment    = "mixed"
aws_region     = "us-east-1"

tags = {
  Owner       = "sreeram"
  environment = "mixed"
  project     = "data-scout"
}
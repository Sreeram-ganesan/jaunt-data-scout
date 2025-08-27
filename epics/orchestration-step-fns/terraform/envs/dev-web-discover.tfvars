# Configuration with DiscoverWebSources real, others mock
# This demonstrates selective use of the new DiscoverWebSources Lambda

# Mock Lambda ARN for states configured as mock
mock_lambda_arn = "arn:aws:lambda:us-east-1:YOUR_ACCOUNT:function:jaunt-mock"

# Mixed configuration: DiscoverWebSources real, others mock
state_implementations = {
  discover_web_sources   = "real"    # Use real web source discovery
  discover_targets       = "mock"    # Mock target discovery
  seed_primaries        = "mock"    # Mock primary seeding
  expand_neighbors      = "mock"    # Mock neighbor expansion
  tile_sweep           = "mock"    # Mock tile sweep
  web_fetch            = "mock"    # Mock web fetching
  extract_with_llm     = "mock"    # Mock LLM extraction
  geocode_validate     = "mock"    # Mock geocoding
  dedupe_canonicalize  = "mock"    # Mock deduplication
  persist              = "mock"    # Mock persistence
  rank                 = "mock"    # Mock ranking
}

# Leave DiscoverWebSources ARN empty to use local Lambda
# lambda_discover_web_sources_arn = ""  # Will use local aws_lambda_function.discover_web_sources

# Placeholder ARNs for other Lambdas (not used when mock)
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
environment    = "dev-web-discover"
aws_region     = "us-east-1"

tags = {
  Owner       = "sreeram"
  environment = "dev-web-discover"
  project     = "data-scout"
}
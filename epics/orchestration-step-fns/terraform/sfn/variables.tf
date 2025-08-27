variable "state_machine_name" { type = string }
variable "sfn_role_arn" { type = string }

variable "lambda_discover_web_sources_arn" { type = string }
variable "lambda_discover_targets_arn" { type = string }
variable "lambda_seed_primaries_arn" { type = string }
variable "lambda_expand_neighbors_arn" { type = string }
variable "lambda_tile_sweep_arn" { type = string }
variable "lambda_web_fetch_arn" { type = string }
variable "lambda_extract_with_llm_arn" { type = string }
variable "lambda_geocode_validate_arn" { type = string }
variable "lambda_dedupe_canonicalize_arn" { type = string }
variable "lambda_persist_arn" { type = string }
variable "lambda_rank_arn" { type = string }

variable "frontier_dlq_url" { type = string }

# Feature flag variables
variable "mock_lambda_arn" {
  description = "ARN of the mock Lambda function to use for states configured as mock"
  type        = string
}

variable "state_implementations" {
  description = "Configure each state to use 'mock' or 'real' implementation"
  type = object({
    discover_web_sources   = string
    discover_targets       = string
    seed_primaries        = string
    expand_neighbors      = string
    tile_sweep           = string
    web_fetch            = string
    extract_with_llm     = string
    geocode_validate     = string
    dedupe_canonicalize  = string
    persist              = string
    rank                 = string
  })
}
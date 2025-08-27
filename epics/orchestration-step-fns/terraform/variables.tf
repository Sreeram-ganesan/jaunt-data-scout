variable "project_prefix" {
  description = "Project prefix for naming."
  type        = string
  default     = "jaunt"
}

variable "environment" {
  description = "Deployment environment (e.g., dev, stage, prod)."
  type        = string
  default     = "dev"
}

variable "aws_region" {
  description = "AWS region."
  type        = string
  default     = "us-east-1"
}

variable "tags" {
  description = "Common tags to apply to resources."
  type        = map(string)
  default = {
    environment   = "dev"
    creation_date = "2025-08-24"
    author        = "sreeram"
    project       = "data-scout"
  }
}

variable "s3_raw_cache_bucket_name" {
  description = "S3 bucket for raw cache storage."
  type        = string
  default     = "data-scout-raw-cache"
}

variable "sqs_frontier_name" {
  description = "data-scout-orchestration-step-function-frontier"
  default     = "data-scout-orchestration-step-function-frontier"
  type        = string
}

variable "sqs_dlq_name" {
  description = "data-scout-orchestration-step-function-dlq"
  default     = "data-scout-orchestration-step-function-dlq"
  type        = string
}

variable "step_function_name" {
  description = "data-scout-orchestration-step-function"
  type        = string
  default     = "data-scout-orchestration-step-function"
}

variable "lambda_discover_web_sources_arn" {
  type = string
}

variable "lambda_discover_targets_arn" {
  type = string
}

variable "lambda_seed_primaries_arn" {
  type = string
}

variable "lambda_expand_neighbors_arn" {
  type = string
}

variable "lambda_tile_sweep_arn" {
  type = string
}

variable "lambda_web_fetch_arn" {
  type = string
}

variable "lambda_extract_with_llm_arn" {
  type = string
}

variable "lambda_geocode_validate_arn" {
  type = string
}

variable "lambda_dedupe_canonicalize_arn" {
  type = string
}

variable "lambda_persist_arn" {
  type = string
}

variable "lambda_rank_arn" {
  type = string
}

# Feature flags to control mock vs real implementations
variable "mock_lambda_arn" {
  description = "ARN of the mock Lambda function to use for states configured as mock"
  type        = string
  default     = ""
}

variable "state_implementations" {
  description = "Configure each state to use 'mock' or 'real' implementation"
  type = object({
    discover_web_sources   = optional(string, "mock")
    discover_targets       = optional(string, "mock")
    seed_primaries        = optional(string, "mock")
    expand_neighbors      = optional(string, "mock")
    tile_sweep           = optional(string, "mock")
    web_fetch            = optional(string, "mock")
    extract_with_llm     = optional(string, "mock")
    geocode_validate     = optional(string, "mock")
    dedupe_canonicalize  = optional(string, "mock")
    persist              = optional(string, "mock")
    rank                 = optional(string, "mock")
  })
  default = {}
  
  validation {
    condition = alltrue([
      for impl in values(var.state_implementations) : contains(["mock", "real"], impl)
    ])
    error_message = "All state implementations must be either 'mock' or 'real'."
  }
}
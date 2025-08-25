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
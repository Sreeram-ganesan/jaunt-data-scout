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
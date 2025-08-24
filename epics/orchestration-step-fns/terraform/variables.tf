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
  default     = {}
}

variable "s3_raw_cache_bucket_name" {
  description = "Name of the S3 bucket for raw cache."
  type        = string
  default     = null
}

variable "sqs_frontier_name" {
  description = "Name of the frontier SQS queue."
  type        = string
  default     = null
}

variable "sqs_dlq_name" {
  description = "Name of the DLQ."
  type        = string
  default     = null
}

variable "step_function_name" {
  description = "Name of the Step Functions state machine."
  type        = string
  default     = null
}
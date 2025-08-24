locals {
  name_base = "${var.project_prefix}-${var.environment}-orchestration"

  s3_raw_cache_bucket_name = coalesce(var.s3_raw_cache_bucket_name, "${local.name_base}-raw-cache")
  sqs_frontier_name        = coalesce(var.sqs_frontier_name, "${local.name_base}-frontier")
  sqs_dlq_name             = coalesce(var.sqs_dlq_name, "${local.name_base}-dlq")
  step_function_name       = coalesce(var.step_function_name, "${local.name_base}-state-machine")

  tags = merge(
    {
      Project     = var.project_prefix
      Environment = var.environment
      ManagedBy   = "terraform"
    },
    var.tags
  )
}
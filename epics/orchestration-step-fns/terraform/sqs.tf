variable "project" { type = string }
variable "environment" { type = string }

locals {
  name_prefix = "${var.project}-${var.environment}"
}

resource "aws_sqs_queue" "frontier_dlq" {
  name                       = "${local.name_prefix}-frontier-dlq"
  message_retention_seconds  = 1209600
  visibility_timeout_seconds = 60
  sqs_managed_sse_enabled    = true
}

resource "aws_sqs_queue" "frontier" {
  name                       = "${local.name_prefix}-frontier"
  visibility_timeout_seconds = 120
  sqs_managed_sse_enabled    = true
  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.frontier_dlq.arn
    maxReceiveCount     = 5
  })
  tags = {
    Project     = var.project
    Environment = var.environment
  }
}

output "frontier_queue_url" {
  value = aws_sqs_queue.frontier.id
}

output "frontier_dlq_url" {
  value = aws_sqs_queue.frontier_dlq.id
}
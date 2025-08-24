resource "aws_sqs_queue" "dlq" {
  name                      = local.sqs_dlq_name
  message_retention_seconds = 1209600 # 14 days
  tags                      = local.tags
}

resource "aws_sqs_queue" "frontier" {
  name                       = local.sqs_frontier_name
  visibility_timeout_seconds = 60
  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.dlq.arn
    maxReceiveCount     = 5
  })
  tags = local.tags
}
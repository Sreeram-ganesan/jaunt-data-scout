output "s3_raw_cache_bucket" {
  value = aws_s3_bucket.raw.bucket
}

output "sqs_frontier_queue_url" {
  value = aws_sqs_queue.frontier.id
}

output "sqs_dlq_queue_url" {
  value = aws_sqs_queue.frontier_dlq.id
}

output "state_machine_arn" {
  value = aws_sfn_state_machine.orchestration.arn
}
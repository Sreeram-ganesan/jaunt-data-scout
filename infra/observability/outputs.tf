output "dashboard_url" {
  description = "URL to the CloudWatch dashboard"
  value       = "https://${var.aws_region}.console.aws.amazon.com/cloudwatch/home?region=${var.aws_region}#dashboards:name=${var.dashboard_name}"
}

output "sns_topic_arn" {
  description = "ARN of the SNS topic for alarms (if email is configured)"
  value       = var.alarm_email != "" ? aws_sns_topic.alarms[0].arn : null
}

output "alarm_names" {
  description = "Names of created CloudWatch alarms"
  value = [
    aws_cloudwatch_metric_alarm.dlq_depth.alarm_name,
    aws_cloudwatch_metric_alarm.execution_failures.alarm_name,
    aws_cloudwatch_metric_alarm.error_rate_spike.alarm_name,
    aws_cloudwatch_metric_alarm.budget_cap_nearing.alarm_name
  ]
}
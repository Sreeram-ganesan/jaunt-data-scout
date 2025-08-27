variable "dashboard_name" {
  description = "Name for the CloudWatch dashboard"
  type        = string
  default     = "JauntDataScout-Observability"
}

variable "state_machine_arn" {
  description = "ARN of the Step Functions state machine"
  type        = string
}

variable "frontier_queue_name" {
  description = "Name of the frontier SQS queue"
  type        = string
}

variable "dlq_queue_name" {
  description = "Name of the DLQ SQS queue"
  type        = string
}

variable "aws_region" {
  description = "AWS region for the dashboard"
  type        = string
  default     = "us-east-1"
}

variable "alarm_email" {
  description = "Email address for alarm notifications"
  type        = string
  default     = ""
}

# Alarm thresholds
variable "dlq_depth_threshold" {
  description = "DLQ depth threshold for alarms"
  type        = number
  default     = 10
}

variable "execution_failure_threshold" {
  description = "Execution failure threshold for alarms"
  type        = number
  default     = 5
}

variable "error_rate_threshold" {
  description = "Error rate threshold (percentage)"
  type        = number
  default     = 10
}

variable "budget_cap_threshold" {
  description = "Budget cap utilization threshold (percentage)"
  type        = number
  default     = 90
}
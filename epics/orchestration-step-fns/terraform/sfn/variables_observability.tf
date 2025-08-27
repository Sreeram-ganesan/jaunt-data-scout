variable "sfn_log_level" {
  type        = string
  default     = "OFF"
  description = "OFF disables logging configuration; otherwise ERROR or ALL."
}

variable "sfn_include_execution_data" {
  type        = bool
  default     = false
  description = "Include execution data in Step Functions logs when logging is enabled."
}

variable "enable_sfn_tracing" {
  type        = bool
  default     = false
  description = "Enable X-Ray tracing on the state machine."
}

variable "sfn_log_group_arn" {
  type        = string
  default     = ""
  description = "CloudWatch Log Group ARN for Step Functions logging (required when sfn_log_level != OFF)."
}
variable "enable_observability" {
  description = "Top-level toggle; if false, all observability features remain disabled unless explicitly enabled."
  type        = bool
  default     = false
}

variable "enable_cloudwatch_dashboards" {
  description = "Create CloudWatch dashboards when true."
  type        = bool
  default     = false
}

variable "enable_cloudwatch_alarms" {
  description = "Create CloudWatch metric alarms when true."
  type        = bool
  default     = false
}

variable "sfn_log_level" {
  description = "Step Functions logging level. OFF disables logging entirely."
  type        = string
  default     = "OFF"
  validation {
    condition     = contains(["OFF", "ERROR", "ALL"], var.sfn_log_level)
    error_message = "sfn_log_level must be one of: OFF, ERROR, ALL"
  }
}

variable "sfn_include_execution_data" {
  description = "Whether Step Functions logs should include execution data when logging is enabled."
  type        = bool
  default     = false
}

variable "enable_sfn_tracing" {
  description = "Enable X-Ray tracing on the Step Functions state machine."
  type        = bool
  default     = false
}

variable "sfn_log_group_arn" {
  description = "CloudWatch Log Group ARN for Step Functions logging; required only when sfn_log_level != OFF."
  type        = string
  default     = ""
}
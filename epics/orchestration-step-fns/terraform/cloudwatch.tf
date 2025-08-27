resource "aws_cloudwatch_log_group" "sfn" {
  count             = var.sfn_log_level != "OFF" ? 1 : 0
  name              = "/aws/states/${local.step_function_name}"
  retention_in_days = 30
  tags              = local.tags
}
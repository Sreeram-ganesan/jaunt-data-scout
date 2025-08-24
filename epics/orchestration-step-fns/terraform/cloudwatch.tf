resource "aws_cloudwatch_log_group" "sfn" {
  name              = "/aws/states/${local.step_function_name}"
  retention_in_days = 30
  tags              = local.tags
}
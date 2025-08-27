# Optional: SFN → CloudWatch Logs permissions (attached only when logging is enabled)
data "aws_iam_policy_document" "sfn_logs_optional" {
  statement {
    sid    = "AllowLogs"
    effect = "Allow"
    actions = [
      "logs:CreateLogDelivery",
      "logs:GetLogDelivery",
      "logs:UpdateLogDelivery",
      "logs:DeleteLogDelivery",
      "logs:ListLogDeliveries",
      "logs:PutLogEvents",
      "logs:CreateLogStream",
      "logs:CreateLogGroup",
    ]
    resources = ["*"]
  }
}

resource "aws_iam_role_policy" "sfn_logs_optional" {
  count  = var.sfn_log_level != "OFF" ? 1 : 0
  name   = "${local.step_function_name}-logs"
  role   = aws_iam_role.sfn_role.id
  policy = data.aws_iam_policy_document.sfn_logs_optional.json
}

# Optional: SFN → X-Ray permissions (attached only when tracing is enabled)
data "aws_iam_policy_document" "sfn_xray_optional" {
  statement {
    sid       = "AllowXRay"
    effect    = "Allow"
    actions   = ["xray:PutTraceSegments", "xray:PutTelemetryRecords"]
    resources = ["*"]
  }
}

resource "aws_iam_role_policy" "sfn_xray_optional" {
  count  = var.enable_sfn_tracing ? 1 : 0
  name   = "${local.step_function_name}-xray"
  role   = aws_iam_role.sfn_role.id
  policy = data.aws_iam_policy_document.sfn_xray_optional.json
}
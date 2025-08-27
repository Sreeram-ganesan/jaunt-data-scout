
data "aws_region" "current" {}

# The Step Functions state machine assumes this role (as seen in cloud logs)
data "aws_iam_role" "sfn_role" {
  name = "data-scout-orchestration-step-function-role"
}

data "aws_iam_policy_document" "sfn_invoke_lambda_doc" {
  statement {
    sid     = "AllowInvokeLambdaForProjectPrefix"
    effect  = "Allow"
    actions = [
      "lambda:InvokeFunction"
    ]
    resources = [
      "arn:aws:lambda:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:function:${var.project_prefix}-*"
    ]
  }
}

resource "aws_iam_policy" "sfn_invoke_lambda" {
  name        = "${var.project_prefix}-${var.environment}-sfn-invoke-lambda"
  description = "Allow Step Functions to invoke Lambda functions for ${var.project_prefix} (${var.environment})"
  policy      = data.aws_iam_policy_document.sfn_invoke_lambda_doc.json
}

resource "aws_iam_role_policy_attachment" "attach_sfn_invoke_lambda" {
  role       = data.aws_iam_role.sfn_role.name
  policy_arn = aws_iam_policy.sfn_invoke_lambda.arn
}

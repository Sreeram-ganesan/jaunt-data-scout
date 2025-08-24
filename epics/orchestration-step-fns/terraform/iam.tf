data "aws_iam_policy_document" "sfn_assume" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["states.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "sfn_role" {
  name               = "${local.step_function_name}-role"
  assume_role_policy = data.aws_iam_policy_document.sfn_assume.json
  tags               = local.tags
}

data "aws_iam_policy_document" "sfn_policy" {
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
      "logs:CreateLogGroup"
    ]
    resources = ["*"]
  }

  statement {
    sid     = "AllowSQS"
    effect  = "Allow"
    actions = ["sqs:SendMessage", "sqs:ReceiveMessage", "sqs:DeleteMessage", "sqs:GetQueueAttributes"]
    resources = [
      aws_sqs_queue.frontier.arn,
      aws_sqs_queue.dlq.arn
    ]
  }

  statement {
    sid     = "AllowS3RawCache"
    effect  = "Allow"
    actions = ["s3:PutObject", "s3:GetObject", "s3:ListBucket"]
    resources = [
      aws_s3_bucket.raw_cache.arn,
      "${aws_s3_bucket.raw_cache.arn}/*"
    ]
  }
}

resource "aws_iam_role_policy" "sfn_inline" {
  name   = "${local.step_function_name}-inline"
  role   = aws_iam_role.sfn_role.id
  policy = data.aws_iam_policy_document.sfn_policy.json
}
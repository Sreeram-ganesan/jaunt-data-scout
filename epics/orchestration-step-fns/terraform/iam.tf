locals {
  common_tags = {
    Project     = "data-scout"
    Environment = var.environment
  }
  # New: list of Lambda ARNs SFN needs to invoke
  lambda_arns = [
    var.lambda_discover_web_sources_arn,
    var.lambda_discover_targets_arn,
    var.lambda_seed_primaries_arn,
    var.lambda_expand_neighbors_arn,
    var.lambda_tile_sweep_arn,
    var.lambda_web_fetch_arn,
    var.lambda_extract_with_llm_arn,
    var.lambda_geocode_validate_arn,
    var.lambda_dedupe_canonicalize_arn,
    var.lambda_persist_arn,
    var.lambda_rank_arn,
  ]
}

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
  tags = merge(
    local.common_tags,
    {
      Name = "data-scout-orchestration-step-function-role"
    }
  )
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
    actions = [
      "sqs:SendMessage",
      "sqs:ReceiveMessage",
      "sqs:DeleteMessage",
      "sqs:GetQueueAttributes",
      "sqs:GetQueueUrl"
    ]
    resources = [
      aws_sqs_queue.frontier_dlq.arn
    ]
  }

  statement {
    sid     = "AllowS3RawCache"
    effect  = "Allow"
    actions = [
      "s3:GetObject",
      "s3:PutObject",
      "s3:ListBucket"
    ]
    resources = [
      aws_s3_bucket.raw.arn,
      "${aws_s3_bucket.raw.arn}/*"
    ]
  }

  # New: allow SFN role to invoke provided Lambdas (base + version/alias ARNs)
  statement {
    sid     = "AllowInvokeLambda"
    effect  = "Allow"
    actions = ["lambda:InvokeFunction"]
    resources = concat(
      local.lambda_arns,
      [for arn in local.lambda_arns : "${arn}:*"]
    )
  }
}

resource "aws_iam_role_policy" "sfn_inline" {
  name   = "${local.step_function_name}-inline"
  role   = aws_iam_role.sfn_role.id
  policy = data.aws_iam_policy_document.sfn_policy.json
}
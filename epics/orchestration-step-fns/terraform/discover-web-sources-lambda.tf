# Lambda function for DiscoverWebSources
resource "aws_lambda_function" "discover_web_sources" {
  function_name = "${local.name_prefix}-discover-web-sources"
  role          = aws_iam_role.discover_web_sources_role.arn
  
  # Placeholder - will be updated with actual deployment package
  filename      = "placeholder.zip"
  handler       = "bootstrap"
  runtime       = "provided.al2"
  architectures = ["x86_64"]
  
  timeout     = 300  # 5 minutes
  memory_size = 512
  
  environment {
    variables = {
      TAVILY_SECRET_ARN    = aws_secretsmanager_secret.tavily_api_key.arn
      FRONTIER_QUEUE_URL   = aws_sqs_queue.frontier.id
      PROJECT_PREFIX       = var.project_prefix
      ENVIRONMENT          = var.environment
    }
  }

  tags = merge(
    local.common_tags,
    {
      Name = "${local.name_prefix}-discover-web-sources"
    }
  )
  
  lifecycle {
    ignore_changes = [filename, source_code_hash]
  }
}

# IAM role for the DiscoverWebSources Lambda
resource "aws_iam_role" "discover_web_sources_role" {
  name = "${local.name_prefix}-discover-web-sources-role"
  
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "lambda.amazonaws.com"
        }
      }
    ]
  })

  tags = merge(
    local.common_tags,
    {
      Name = "${local.name_prefix}-discover-web-sources-role"
    }
  )
}

# IAM policy for DiscoverWebSources Lambda
resource "aws_iam_role_policy" "discover_web_sources_policy" {
  name = "${local.name_prefix}-discover-web-sources-policy"
  role = aws_iam_role.discover_web_sources_role.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "logs:CreateLogGroup",
          "logs:CreateLogStream", 
          "logs:PutLogEvents"
        ]
        Resource = "arn:aws:logs:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:*"
      },
      {
        Effect = "Allow"
        Action = [
          "secretsmanager:GetSecretValue"
        ]
        Resource = aws_secretsmanager_secret.tavily_api_key.arn
      },
      {
        Effect = "Allow"
        Action = [
          "sqs:SendMessage",
          "sqs:SendMessageBatch"
        ]
        Resource = aws_sqs_queue.frontier.arn
      }
    ]
  })
}

# Output the Lambda ARN
output "discover_web_sources_lambda_arn" {
  value       = aws_lambda_function.discover_web_sources.arn
  description = "ARN of the DiscoverWebSources Lambda function"
}
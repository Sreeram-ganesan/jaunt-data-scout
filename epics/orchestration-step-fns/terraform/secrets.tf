# AWS Secrets Manager resources for API keys and credentials
# This file implements secrets management for Google, Tavily, OTM, DB, and LLM services

# Google Maps API Key
resource "aws_secretsmanager_secret" "google_api_key" {
  name        = "${var.project_prefix}-${var.environment}-google-api-key"
  description = "Google Maps Platform API key for Places API (Text, Nearby, Details)"
  
  # Enable automatic rotation (30 days)
  rotation_rules {
    automatically_after_days = 30
  }

  tags = merge(var.tags, {
    Name        = "Google Maps API Key"
    Service     = "google-maps"
    Rotation    = "enabled"
    Environment = var.environment
  })
}

# Google API Key Version - placeholder, will be populated externally
resource "aws_secretsmanager_secret_version" "google_api_key" {
  secret_id = aws_secretsmanager_secret.google_api_key.id
  secret_string = jsonencode({
    api_key = "PLACEHOLDER_GOOGLE_API_KEY"
    service = "google-maps-platform"
    scopes  = ["places", "geocoding"]
  })

  lifecycle {
    ignore_changes = [secret_string]
  }
}

# Tavily API Key
resource "aws_secretsmanager_secret" "tavily_api_key" {
  name        = "${var.project_prefix}-${var.environment}-tavily-api-key"
  description = "Tavily API key for web search and discovery"
  
  rotation_rules {
    automatically_after_days = 30
  }

  tags = merge(var.tags, {
    Name        = "Tavily API Key"
    Service     = "tavily"
    Rotation    = "enabled"
    Environment = var.environment
  })
}

resource "aws_secretsmanager_secret_version" "tavily_api_key" {
  secret_id = aws_secretsmanager_secret.tavily_api_key.id
  secret_string = jsonencode({
    api_key = "PLACEHOLDER_TAVILY_API_KEY"
    service = "tavily-search"
  })

  lifecycle {
    ignore_changes = [secret_string]
  }
}

# OpenTripMap API Key
resource "aws_secretsmanager_secret" "otm_api_key" {
  name        = "${var.project_prefix}-${var.environment}-otm-api-key"
  description = "OpenTripMap API key for places of interest data"
  
  rotation_rules {
    automatically_after_days = 30
  }

  tags = merge(var.tags, {
    Name        = "OpenTripMap API Key"
    Service     = "opentripmap"
    Rotation    = "enabled"
    Environment = var.environment
  })
}

resource "aws_secretsmanager_secret_version" "otm_api_key" {
  secret_id = aws_secretsmanager_secret.otm_api_key.id
  secret_string = jsonencode({
    api_key = "PLACEHOLDER_OTM_API_KEY"
    service = "opentripmap"
  })

  lifecycle {
    ignore_changes = [secret_string]
  }
}

# Database Connection String
resource "aws_secretsmanager_secret" "database_credentials" {
  name        = "${var.project_prefix}-${var.environment}-database-credentials"
  description = "Database connection credentials for Postgres"
  
  rotation_rules {
    automatically_after_days = 90
  }

  tags = merge(var.tags, {
    Name        = "Database Credentials"
    Service     = "postgres"
    Rotation    = "enabled"
    Environment = var.environment
  })
}

resource "aws_secretsmanager_secret_version" "database_credentials" {
  secret_id = aws_secretsmanager_secret.database_credentials.id
  secret_string = jsonencode({
    host     = "PLACEHOLDER_DB_HOST"
    port     = 5432
    username = "PLACEHOLDER_DB_USER"
    password = "PLACEHOLDER_DB_PASSWORD"
    database = "PLACEHOLDER_DB_NAME"
    ssl_mode = "require"
  })

  lifecycle {
    ignore_changes = [secret_string]
  }
}

# LLM Provider API Keys (supports multiple providers)
resource "aws_secretsmanager_secret" "llm_api_keys" {
  name        = "${var.project_prefix}-${var.environment}-llm-api-keys"
  description = "LLM provider API keys (Bedrock, OpenAI, Gemini)"
  
  rotation_rules {
    automatically_after_days = 30
  }

  tags = merge(var.tags, {
    Name        = "LLM API Keys"
    Service     = "llm-providers"
    Rotation    = "enabled"
    Environment = var.environment
  })
}

resource "aws_secretsmanager_secret_version" "llm_api_keys" {
  secret_id = aws_secretsmanager_secret.llm_api_keys.id
  secret_string = jsonencode({
    openai_api_key = "PLACEHOLDER_OPENAI_API_KEY"
    anthropic_api_key = "PLACEHOLDER_ANTHROPIC_API_KEY"
    bedrock_region = "us-east-1"
    default_provider = "bedrock"
  })

  lifecycle {
    ignore_changes = [secret_string]
  }
}

# IAM policy for Lambda functions to access secrets
data "aws_iam_policy_document" "lambda_secrets_policy" {
  statement {
    sid    = "AllowSecretsManagerRead"
    effect = "Allow"
    actions = [
      "secretsmanager:GetSecretValue",
      "secretsmanager:DescribeSecret"
    ]
    resources = [
      aws_secretsmanager_secret.google_api_key.arn,
      aws_secretsmanager_secret.tavily_api_key.arn,
      aws_secretsmanager_secret.otm_api_key.arn,
      aws_secretsmanager_secret.database_credentials.arn,
      aws_secretsmanager_secret.llm_api_keys.arn
    ]
  }

  statement {
    sid    = "AllowKMSDecrypt"
    effect = "Allow"
    actions = [
      "kms:Decrypt",
      "kms:GenerateDataKey"
    ]
    resources = ["*"]
    condition {
      test     = "StringEquals"
      variable = "kms:ViaService"
      values   = ["secretsmanager.${var.aws_region}.amazonaws.com"]
    }
  }
}

resource "aws_iam_policy" "lambda_secrets_policy" {
  name        = "${var.project_prefix}-${var.environment}-lambda-secrets-policy"
  description = "Policy allowing Lambda functions to access secrets"
  policy      = data.aws_iam_policy_document.lambda_secrets_policy.json

  tags = merge(var.tags, {
    Name    = "Lambda Secrets Access Policy"
    Purpose = "secrets-management"
  })
}

# Outputs for use by other modules
output "secrets" {
  description = "Map of secret ARNs and names"
  value = {
    google_api_key = {
      arn  = aws_secretsmanager_secret.google_api_key.arn
      name = aws_secretsmanager_secret.google_api_key.name
    }
    tavily_api_key = {
      arn  = aws_secretsmanager_secret.tavily_api_key.arn
      name = aws_secretsmanager_secret.tavily_api_key.name
    }
    otm_api_key = {
      arn  = aws_secretsmanager_secret.otm_api_key.arn
      name = aws_secretsmanager_secret.otm_api_key.name
    }
    database_credentials = {
      arn  = aws_secretsmanager_secret.database_credentials.arn
      name = aws_secretsmanager_secret.database_credentials.name
    }
    llm_api_keys = {
      arn  = aws_secretsmanager_secret.llm_api_keys.arn
      name = aws_secretsmanager_secret.llm_api_keys.name
    }
  }
  sensitive = true
}

output "lambda_secrets_policy_arn" {
  description = "ARN of the IAM policy for Lambda secrets access"
  value       = aws_iam_policy.lambda_secrets_policy.arn
}
# Secrets Manager for Tavily API key
resource "aws_secretsmanager_secret" "tavily_api_key" {
  name        = "${local.name_prefix}-tavily-api-key"
  description = "Tavily API key for web source discovery"
  
  tags = merge(
    local.common_tags,
    {
      Name = "${local.name_prefix}-tavily-api-key"
    }
  )
}

resource "aws_secretsmanager_secret_version" "tavily_api_key" {
  secret_id = aws_secretsmanager_secret.tavily_api_key.id
  secret_string = jsonencode({
    api_key = "PLACEHOLDER-SET-VIA-AWS-CLI"
  })
  
  lifecycle {
    ignore_changes = [secret_string]
  }
}

# Output the secret ARN for use by Lambda
output "tavily_secret_arn" {
  value       = aws_secretsmanager_secret.tavily_api_key.arn
  description = "ARN of the Tavily API key secret"
}
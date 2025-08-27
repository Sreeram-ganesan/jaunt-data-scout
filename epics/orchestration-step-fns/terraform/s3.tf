data "aws_caller_identity" "current" {}

locals {
  raw_bucket_name         = lower("jaunt-${var.environment}-data-scout-raw-${data.aws_caller_identity.current.account_id}")
  access_logs_bucket_name = lower("jaunt-${var.environment}-data-scout-access-logs-${data.aws_caller_identity.current.account_id}")
}

# S3 bucket for access logs
resource "aws_s3_bucket" "access_logs" {
  bucket = local.access_logs_bucket_name

  tags = merge(var.tags, {
    Name    = "Data Scout Access Logs"
    Purpose = "access-logging"
  })
}

resource "aws_s3_bucket_server_side_encryption_configuration" "access_logs" {
  bucket = aws_s3_bucket.access_logs.id
  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

resource "aws_s3_bucket_lifecycle_configuration" "access_logs" {
  bucket = aws_s3_bucket.access_logs.id

  rule {
    id     = "access-logs-retention"
    status = "Enabled"
    expiration { days = 90 }
  }
}

resource "aws_s3_bucket_public_access_block" "access_logs" {
  bucket = aws_s3_bucket.access_logs.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# Main raw data bucket
resource "aws_s3_bucket" "raw" {
  bucket = local.raw_bucket_name

  tags = merge(var.tags, {
    Name    = "Data Scout Raw Data"
    Purpose = "raw-data-storage"
  })
}

# Enhanced server-side encryption with optional KMS
resource "aws_s3_bucket_server_side_encryption_configuration" "raw" {
  bucket = aws_s3_bucket.raw.id
  
  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm     = var.enable_kms_encryption ? "aws:kms" : "AES256"
      kms_master_key_id = var.enable_kms_encryption ? var.kms_key_id : null
    }
    bucket_key_enabled = var.enable_kms_encryption
  }
}

# Enable access logging
resource "aws_s3_bucket_logging" "raw" {
  bucket = aws_s3_bucket.raw.id

  target_bucket = aws_s3_bucket.access_logs.id
  target_prefix = "access-logs/raw-bucket/"
}

# Block all public access
resource "aws_s3_bucket_public_access_block" "raw" {
  bucket = aws_s3_bucket.raw.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_s3_bucket_lifecycle_configuration" "raw" {
  bucket = aws_s3_bucket.raw.id

  rule {
    id     = "raw-json-google-30d"
    status = "Enabled"
    filter {
      prefix = "raw/json/google/"
    }
    expiration { days = 30 }
  }

  rule {
    id     = "open-data-json-90d"
    status = "Enabled"
    filter {
      prefix = "raw/json/open-data/"
    }
    expiration { days = 90 }
  }

  rule {
    id     = "raw-html-90d"
    status = "Enabled"
    filter { prefix = "raw/html/" }
    expiration { days = 90 }
  }

  rule {
    id     = "extracted-json-90d"
    status = "Enabled"
    filter { prefix = "extracted/" }
    expiration { days = 90 }
  }
}

output "s3_bucket_name" { 
  value = aws_s3_bucket.raw.bucket 
}

output "s3_access_logs_bucket_name" {
  value = aws_s3_bucket.access_logs.bucket
}
variable "s3_bucket_name" { type = string }

data "aws_caller_identity" "current" {}

locals {
  raw_bucket_name = lower("jaunt-${var.environment}-data-scout-raw-${data.aws_caller_identity.current.account_id}")
}

resource "aws_s3_bucket" "raw" {
  bucket = local.raw_bucket_name
}

resource "aws_s3_bucket_server_side_encryption_configuration" "raw" {
  bucket = aws_s3_bucket.raw.id
  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
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

output "s3_bucket_name" { value = aws_s3_bucket.raw.bucket }
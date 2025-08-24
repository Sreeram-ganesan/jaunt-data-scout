variable "s3_bucket_name" { type = string }

resource "aws_s3_bucket" "raw" {
  bucket = var.s3_bucket_name
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
resource "aws_s3_bucket" "raw_cache" {
  bucket = local.s3_raw_cache_bucket_name
  tags   = local.tags
}

resource "aws_s3_bucket_versioning" "raw_cache" {
  bucket = aws_s3_bucket.raw_cache.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "raw_cache" {
  bucket = aws_s3_bucket.raw_cache.id
  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}
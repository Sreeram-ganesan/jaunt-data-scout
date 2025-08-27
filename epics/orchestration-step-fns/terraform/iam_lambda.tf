# Least-privilege IAM roles for Lambda functions
# Each Lambda function gets its own role with minimal required permissions

# Base Lambda execution role policy
data "aws_iam_policy_document" "lambda_assume_role" {
  statement {
    effect = "Allow"
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
  }
}

# Common Lambda permissions for all functions
data "aws_iam_policy_document" "lambda_base_policy" {
  statement {
    sid    = "CloudWatchLogs"
    effect = "Allow"
    actions = [
      "logs:CreateLogGroup",
      "logs:CreateLogStream",
      "logs:PutLogEvents"
    ]
    resources = ["arn:aws:logs:${var.aws_region}:*:*"]
  }

  statement {
    sid    = "XRayTracing"
    effect = "Allow"
    actions = [
      "xray:PutTraceSegments",
      "xray:PutTelemetryRecords"
    ]
    resources = ["*"]
  }

  statement {
    sid    = "SQSFrontierAccess"
    effect = "Allow"
    actions = [
      "sqs:ReceiveMessage",
      "sqs:DeleteMessage",
      "sqs:SendMessage",
      "sqs:GetQueueAttributes"
    ]
    resources = [
      aws_sqs_queue.frontier.arn,
      aws_sqs_queue.frontier_dlq.arn
    ]
  }
}

# Web Fetch Lambda Role - needs S3, secrets, and outbound internet access
resource "aws_iam_role" "lambda_web_fetch_role" {
  name               = "${var.project_prefix}-${var.environment}-lambda-web-fetch-role"
  assume_role_policy = data.aws_iam_policy_document.lambda_assume_role.json

  tags = merge(var.tags, {
    Name     = "Web Fetch Lambda Role"
    Function = "web-fetch"
  })
}

data "aws_iam_policy_document" "lambda_web_fetch_policy" {
  source_policy_documents = [data.aws_iam_policy_document.lambda_base_policy.json]

  statement {
    sid    = "S3RawHTMLAccess"
    effect = "Allow"
    actions = [
      "s3:PutObject",
      "s3:GetObject"
    ]
    resources = ["${aws_s3_bucket.raw.arn}/raw/html/*"]
  }

  statement {
    sid    = "SecretsForWebFetch"
    effect = "Allow"
    actions = [
      "secretsmanager:GetSecretValue"
    ]
    resources = [
      aws_secretsmanager_secret.tavily_api_key.arn
    ]
  }
}

resource "aws_iam_role_policy" "lambda_web_fetch_policy" {
  name   = "web-fetch-policy"
  role   = aws_iam_role.lambda_web_fetch_role.id
  policy = data.aws_iam_policy_document.lambda_web_fetch_policy.json
}

resource "aws_iam_role_policy_attachment" "lambda_web_fetch_secrets" {
  role       = aws_iam_role.lambda_web_fetch_role.name
  policy_arn = aws_iam_policy.lambda_secrets_policy.arn
}

# Extract with LLM Lambda Role - needs S3, secrets for LLM providers
resource "aws_iam_role" "lambda_extract_llm_role" {
  name               = "${var.project_prefix}-${var.environment}-lambda-extract-llm-role"
  assume_role_policy = data.aws_iam_policy_document.lambda_assume_role.json

  tags = merge(var.tags, {
    Name     = "Extract LLM Lambda Role"
    Function = "extract-llm"
  })
}

data "aws_iam_policy_document" "lambda_extract_llm_policy" {
  source_policy_documents = [data.aws_iam_policy_document.lambda_base_policy.json]

  statement {
    sid    = "S3ExtractedAccess"
    effect = "Allow"
    actions = [
      "s3:PutObject",
      "s3:GetObject"
    ]
    resources = [
      "${aws_s3_bucket.raw.arn}/raw/html/*",
      "${aws_s3_bucket.raw.arn}/extracted/*"
    ]
  }

  statement {
    sid    = "BedrockAccess"
    effect = "Allow"
    actions = [
      "bedrock:InvokeModel",
      "bedrock:InvokeModelWithResponseStream"
    ]
    resources = ["*"]
  }

  statement {
    sid    = "SecretsForLLM"
    effect = "Allow"
    actions = [
      "secretsmanager:GetSecretValue"
    ]
    resources = [
      aws_secretsmanager_secret.llm_api_keys.arn
    ]
  }
}

resource "aws_iam_role_policy" "lambda_extract_llm_policy" {
  name   = "extract-llm-policy"
  role   = aws_iam_role.lambda_extract_llm_role.id
  policy = data.aws_iam_policy_document.lambda_extract_llm_policy.json
}

# Google Places Lambda Roles (Text, Nearby, Details)
resource "aws_iam_role" "lambda_google_places_role" {
  name               = "${var.project_prefix}-${var.environment}-lambda-google-places-role"
  assume_role_policy = data.aws_iam_policy_document.lambda_assume_role.json

  tags = merge(var.tags, {
    Name     = "Google Places Lambda Role"
    Function = "google-places"
  })
}

data "aws_iam_policy_document" "lambda_google_places_policy" {
  source_policy_documents = [data.aws_iam_policy_document.lambda_base_policy.json]

  statement {
    sid    = "S3GoogleRawAccess"
    effect = "Allow"
    actions = [
      "s3:PutObject",
      "s3:GetObject"
    ]
    resources = ["${aws_s3_bucket.raw.arn}/raw/json/google/*"]
  }

  statement {
    sid    = "SecretsForGoogle"
    effect = "Allow"
    actions = [
      "secretsmanager:GetSecretValue"
    ]
    resources = [
      aws_secretsmanager_secret.google_api_key.arn
    ]
  }
}

resource "aws_iam_role_policy" "lambda_google_places_policy" {
  name   = "google-places-policy"
  role   = aws_iam_role.lambda_google_places_role.id
  policy = data.aws_iam_policy_document.lambda_google_places_policy.json
}

# Open Data Lambda Role (OTM, OSM, Wiki)
resource "aws_iam_role" "lambda_open_data_role" {
  name               = "${var.project_prefix}-${var.environment}-lambda-open-data-role"
  assume_role_policy = data.aws_iam_policy_document.lambda_assume_role.json

  tags = merge(var.tags, {
    Name     = "Open Data Lambda Role"
    Function = "open-data"
  })
}

data "aws_iam_policy_document" "lambda_open_data_policy" {
  source_policy_documents = [data.aws_iam_policy_document.lambda_base_policy.json]

  statement {
    sid    = "S3OpenDataAccess"
    effect = "Allow"
    actions = [
      "s3:PutObject",
      "s3:GetObject"
    ]
    resources = ["${aws_s3_bucket.raw.arn}/raw/json/open-data/*"]
  }

  statement {
    sid    = "SecretsForOpenData"
    effect = "Allow"
    actions = [
      "secretsmanager:GetSecretValue"
    ]
    resources = [
      aws_secretsmanager_secret.otm_api_key.arn
    ]
  }
}

resource "aws_iam_role_policy" "lambda_open_data_policy" {
  name   = "open-data-policy"
  role   = aws_iam_role.lambda_open_data_role.id
  policy = data.aws_iam_policy_document.lambda_open_data_policy.json
}

# Database Lambda Role (Persist, Rank, Finalize)
resource "aws_iam_role" "lambda_database_role" {
  name               = "${var.project_prefix}-${var.environment}-lambda-database-role"
  assume_role_policy = data.aws_iam_policy_document.lambda_assume_role.json

  tags = merge(var.tags, {
    Name     = "Database Lambda Role"
    Function = "database-access"
  })
}

data "aws_iam_policy_document" "lambda_database_policy" {
  source_policy_documents = [data.aws_iam_policy_document.lambda_base_policy.json]

  statement {
    sid    = "S3ManifestAccess"
    effect = "Allow"
    actions = [
      "s3:PutObject",
      "s3:GetObject",
      "s3:ListBucket"
    ]
    resources = [
      aws_s3_bucket.raw.arn,
      "${aws_s3_bucket.raw.arn}/manifests/*",
      "${aws_s3_bucket.raw.arn}/extracted/*"
    ]
  }

  statement {
    sid    = "SecretsForDatabase"
    effect = "Allow"
    actions = [
      "secretsmanager:GetSecretValue"
    ]
    resources = [
      aws_secretsmanager_secret.database_credentials.arn
    ]
  }
}

resource "aws_iam_role_policy" "lambda_database_policy" {
  name   = "database-policy"
  role   = aws_iam_role.lambda_database_role.id
  policy = data.aws_iam_policy_document.lambda_database_policy.json
}

# Discovery Lambda Role (Discover Web Sources, Targets, Seed, Expand, Tile Sweep)
resource "aws_iam_role" "lambda_discovery_role" {
  name               = "${var.project_prefix}-${var.environment}-lambda-discovery-role"
  assume_role_policy = data.aws_iam_policy_document.lambda_assume_role.json

  tags = merge(var.tags, {
    Name     = "Discovery Lambda Role"
    Function = "discovery"
  })
}

data "aws_iam_policy_document" "lambda_discovery_policy" {
  source_policy_documents = [data.aws_iam_policy_document.lambda_base_policy.json]

  statement {
    sid    = "S3ManifestRead"
    effect = "Allow"
    actions = [
      "s3:GetObject",
      "s3:ListBucket"
    ]
    resources = [
      aws_s3_bucket.raw.arn,
      "${aws_s3_bucket.raw.arn}/manifests/*"
    ]
  }
}

resource "aws_iam_role_policy" "lambda_discovery_policy" {
  name   = "discovery-policy"
  role   = aws_iam_role.lambda_discovery_role.id
  policy = data.aws_iam_policy_document.lambda_discovery_policy.json
}

# Geocode Validate Lambda Role
resource "aws_iam_role" "lambda_geocode_role" {
  name               = "${var.project_prefix}-${var.environment}-lambda-geocode-role"
  assume_role_policy = data.aws_iam_policy_document.lambda_assume_role.json

  tags = merge(var.tags, {
    Name     = "Geocode Lambda Role"
    Function = "geocode-validate"
  })
}

data "aws_iam_policy_document" "lambda_geocode_policy" {
  source_policy_documents = [data.aws_iam_policy_document.lambda_base_policy.json]

  statement {
    sid    = "SecretsForGeocode"
    effect = "Allow"
    actions = [
      "secretsmanager:GetSecretValue"
    ]
    resources = [
      aws_secretsmanager_secret.google_api_key.arn
    ]
  }
}

resource "aws_iam_role_policy" "lambda_geocode_policy" {
  name   = "geocode-policy"
  role   = aws_iam_role.lambda_geocode_role.id
  policy = data.aws_iam_policy_document.lambda_geocode_policy.json
}

# Outputs for Lambda role ARNs
output "lambda_iam_roles" {
  description = "Map of Lambda function IAM role ARNs"
  value = {
    web_fetch_role          = aws_iam_role.lambda_web_fetch_role.arn
    extract_llm_role        = aws_iam_role.lambda_extract_llm_role.arn
    google_places_role      = aws_iam_role.lambda_google_places_role.arn
    open_data_role          = aws_iam_role.lambda_open_data_role.arn
    database_role           = aws_iam_role.lambda_database_role.arn
    discovery_role          = aws_iam_role.lambda_discovery_role.arn
    geocode_role            = aws_iam_role.lambda_geocode_role.arn
  }
}
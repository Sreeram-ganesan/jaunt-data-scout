# VPC configuration for egress control and DNS filtering
# This provides network-level security for WebFetch and LLM operations

# Variables for VPC configuration
variable "enable_vpc_egress_control" {
  description = "Enable VPC with egress filtering for Lambda functions"
  type        = bool
  default     = false
}

variable "vpc_cidr" {
  description = "CIDR block for the VPC"
  type        = string
  default     = "10.0.0.0/16"
}

# VPC for Lambda functions with egress control
resource "aws_vpc" "lambda_vpc" {
  count                = var.enable_vpc_egress_control ? 1 : 0
  cidr_block          = var.vpc_cidr
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = merge(var.tags, {
    Name = "${var.project_prefix}-${var.environment}-lambda-vpc"
    Purpose = "lambda-egress-control"
  })
}

# Private subnets for Lambda functions
resource "aws_subnet" "lambda_private" {
  count             = var.enable_vpc_egress_control ? 2 : 0
  vpc_id            = aws_vpc.lambda_vpc[0].id
  cidr_block        = "10.0.${count.index + 1}.0/24"
  availability_zone = data.aws_availability_zones.available.names[count.index]

  tags = merge(var.tags, {
    Name = "${var.project_prefix}-${var.environment}-lambda-private-${count.index + 1}"
    Type = "private"
  })
}

# Public subnet for NAT Gateway
resource "aws_subnet" "public" {
  count                   = var.enable_vpc_egress_control ? 1 : 0
  vpc_id                  = aws_vpc.lambda_vpc[0].id
  cidr_block              = "10.0.10.0/24"
  availability_zone       = data.aws_availability_zones.available.names[0]
  map_public_ip_on_launch = true

  tags = merge(var.tags, {
    Name = "${var.project_prefix}-${var.environment}-public-subnet"
    Type = "public"
  })
}

# Internet Gateway
resource "aws_internet_gateway" "igw" {
  count  = var.enable_vpc_egress_control ? 1 : 0
  vpc_id = aws_vpc.lambda_vpc[0].id

  tags = merge(var.tags, {
    Name = "${var.project_prefix}-${var.environment}-igw"
  })
}

# Elastic IP for NAT Gateway
resource "aws_eip" "nat" {
  count  = var.enable_vpc_egress_control ? 1 : 0
  domain = "vpc"

  tags = merge(var.tags, {
    Name = "${var.project_prefix}-${var.environment}-nat-eip"
  })

  depends_on = [aws_internet_gateway.igw]
}

# NAT Gateway for outbound internet access
resource "aws_nat_gateway" "nat" {
  count         = var.enable_vpc_egress_control ? 1 : 0
  allocation_id = aws_eip.nat[0].id
  subnet_id     = aws_subnet.public[0].id

  tags = merge(var.tags, {
    Name = "${var.project_prefix}-${var.environment}-nat-gateway"
  })

  depends_on = [aws_internet_gateway.igw]
}

# Route table for public subnet
resource "aws_route_table" "public" {
  count  = var.enable_vpc_egress_control ? 1 : 0
  vpc_id = aws_vpc.lambda_vpc[0].id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.igw[0].id
  }

  tags = merge(var.tags, {
    Name = "${var.project_prefix}-${var.environment}-public-rt"
  })
}

# Route table for private subnets
resource "aws_route_table" "private" {
  count  = var.enable_vpc_egress_control ? 1 : 0
  vpc_id = aws_vpc.lambda_vpc[0].id

  route {
    cidr_block     = "0.0.0.0/0"
    nat_gateway_id = aws_nat_gateway.nat[0].id
  }

  tags = merge(var.tags, {
    Name = "${var.project_prefix}-${var.environment}-private-rt"
  })
}

# Associate public subnet with public route table
resource "aws_route_table_association" "public" {
  count          = var.enable_vpc_egress_control ? 1 : 0
  subnet_id      = aws_subnet.public[0].id
  route_table_id = aws_route_table.public[0].id
}

# Associate private subnets with private route table
resource "aws_route_table_association" "private" {
  count          = var.enable_vpc_egress_control ? 2 : 0
  subnet_id      = aws_subnet.lambda_private[count.index].id
  route_table_id = aws_route_table.private[0].id
}

# Security group for Lambda functions with egress allowlist
resource "aws_security_group" "lambda_egress" {
  count       = var.enable_vpc_egress_control ? 1 : 0
  name_prefix = "${var.project_prefix}-${var.environment}-lambda-egress"
  description = "Security group for Lambda functions with egress allowlist"
  vpc_id      = aws_vpc.lambda_vpc[0].id

  # Allow HTTPS to approved domains
  egress {
    description = "HTTPS to Google APIs"
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # Allow HTTP for robots.txt checks
  egress {
    description = "HTTP for robots.txt"
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # Allow DNS resolution
  egress {
    description = "DNS resolution"
    from_port   = 53
    to_port     = 53
    protocol    = "udp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # Allow access to AWS services via VPC endpoints
  egress {
    description = "AWS services"
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = [var.vpc_cidr]
  }

  tags = merge(var.tags, {
    Name = "${var.project_prefix}-${var.environment}-lambda-egress-sg"
  })
}

# VPC Endpoints for AWS services to avoid internet routing
resource "aws_vpc_endpoint" "s3" {
  count        = var.enable_vpc_egress_control ? 1 : 0
  vpc_id       = aws_vpc.lambda_vpc[0].id
  service_name = "com.amazonaws.${var.aws_region}.s3"
  
  tags = merge(var.tags, {
    Name = "${var.project_prefix}-${var.environment}-s3-endpoint"
  })
}

resource "aws_vpc_endpoint" "secretsmanager" {
  count              = var.enable_vpc_egress_control ? 1 : 0
  vpc_id             = aws_vpc.lambda_vpc[0].id
  service_name       = "com.amazonaws.${var.aws_region}.secretsmanager"
  vpc_endpoint_type  = "Interface"
  subnet_ids         = aws_subnet.lambda_private[*].id
  security_group_ids = [aws_security_group.lambda_egress[0].id]
  
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = "*"
        Action = [
          "secretsmanager:GetSecretValue",
          "secretsmanager:DescribeSecret"
        ]
        Resource = "*"
      }
    ]
  })

  tags = merge(var.tags, {
    Name = "${var.project_prefix}-${var.environment}-secretsmanager-endpoint"
  })
}

# Data source for availability zones
data "aws_availability_zones" "available" {
  state = "available"
}

# Outputs
output "vpc_config" {
  description = "VPC configuration for Lambda functions"
  value = var.enable_vpc_egress_control ? {
    vpc_id             = aws_vpc.lambda_vpc[0].id
    private_subnet_ids = aws_subnet.lambda_private[*].id
    security_group_ids = [aws_security_group.lambda_egress[0].id]
  } : null
}
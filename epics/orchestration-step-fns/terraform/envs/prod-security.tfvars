# Production environment with enhanced security
project_prefix = "jaunt"
environment    = "prod"
aws_region     = "us-east-1"

# Enhanced security options for production
enable_kms_encryption      = true
enable_vpc_egress_control = true
vpc_cidr                  = "10.0.0.0/16"

# Tags for production resources
tags = {
  Environment   = "prod"
  Owner        = "sreeram"
  Project      = "data-scout"
  Compliance   = "required"
  Security     = "enhanced"
  CostCenter   = "engineering"
}
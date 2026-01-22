terraform {
  required_version = ">= 1.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region     = var.aws_region
  access_key = var.aws_access_key
  secret_key = var.aws_secret_key

  assume_role {
    role_arn     = "arn:aws:iam::123456789012:role/TerraformRole"
    session_name = "terraform-session"
  }

  default_tags {
    tags = {
      Environment = var.environment
      Project     = "ComplexConfig"
      ManagedBy   = "Terraform"
      Owner       = "DevOps Team"
    }
  }

  ignore_tags {
    keys = ["IgnoreMe"]
  }
}

# Alias provider for different region
provider "aws" {
  alias  = "tokyo"
  region = "ap-northeast-1"
}

# Alias provider for different account
provider "aws" {
  alias  = "backup"
  region = "us-east-1"

  assume_role {
    role_arn = "arn:aws:iam::987654321098:role/BackupRole"
  }
}

variable "aws_region" {
  description = "Primary AWS region"
  type        = string
  default     = "us-west-2"
}

variable "aws_access_key" {
  description = "AWS access key"
  type        = string
  sensitive   = true
}

variable "aws_secret_key" {
  description = "AWS secret key"
  type        = string
  sensitive   = true
}

variable "environment" {
  description = "Environment name"
  type        = string
  default     = "production"
}

resource "aws_s3_bucket" "primary" {
  bucket = "primary-bucket-${var.environment}"
}

resource "aws_s3_bucket" "tokyo" {
  provider = aws.tokyo
  bucket   = "tokyo-bucket-${var.environment}"
}

resource "aws_s3_bucket" "backup" {
  provider = aws.backup
  bucket   = "backup-bucket-${var.environment}"
}

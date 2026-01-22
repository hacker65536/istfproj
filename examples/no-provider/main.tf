terraform {
  required_version = ">= 1.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

# This file has no provider block - useful for testing --strict mode

variable "region" {
  description = "AWS region"
  type        = string
  default     = "us-east-1"
}

resource "aws_s3_bucket" "data" {
  bucket = "data-bucket"
}

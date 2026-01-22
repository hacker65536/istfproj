terraform {
  required_version = ">= 1.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 3.0"
    }
  }
}

provider "aws" {
  region = "us-west-2"

  default_tags {
    tags = {
      Environment = "Production"
      ManagedBy   = "Terraform"
    }
  }
}

provider "google" {
  project = "my-gcp-project"
  region  = "asia-northeast1"
}

provider "azurerm" {
  features {}
  subscription_id = "00000000-0000-0000-0000-000000000000"
}

resource "aws_s3_bucket" "aws_bucket" {
  bucket = "multi-cloud-aws-bucket"
}

resource "google_storage_bucket" "gcp_bucket" {
  name     = "multi-cloud-gcp-bucket"
  location = "ASIA"
}

resource "azurerm_resource_group" "azure_rg" {
  name     = "multi-cloud-rg"
  location = "Japan East"
}

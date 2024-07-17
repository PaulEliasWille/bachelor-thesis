terraform {
  required_providers {
    aws = {
      source = "hashicorp/aws"
      version = "5.58.0"
    }

    random = {
      source = "hashicorp/random"
      version = "3.6.2"
    }

    archive = {
      source = "hashicorp/archive"
      version = "2.4.2"
    }
  }
}

provider "aws" {
  region = "eu-central-1"

  default_tags {
    tags = {
      "Managed-by" = "Terraform"
      "Environment" = "prod"
    }
  }
}

provider "random" {
}

provider "archive" {
}
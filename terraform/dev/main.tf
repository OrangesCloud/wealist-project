# ./terraform/dev/main.tf

# Terraform 백엔드 설정 (상태 파일 S3 저장)
# - S3 Bucket과 DynamoDB Table은 수동으로 미리 생성되어야 합니다.
terraform {
  backend "s3" {
    bucket         = "wealist-terraform-state-dev" # S3 버킷 이름 (유니크해야 함)
    key            = "dev/wealist.tfstate"
    region         = "ap-northeast-2"
    encrypt        = true
  }
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

# AWS Provider 설정
provider "aws" {
  region = var.aws_region
}
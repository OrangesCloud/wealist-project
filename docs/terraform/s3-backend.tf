# Terraform S3 Backend Configuration
# State 파일을 안전하게 S3에 저장 + DynamoDB로 잠금

# ============================================
# 주의사항
# ============================================
#
# 1. 이 파일은 backend 설정 전에 먼저 apply해서 S3 버킷과 DynamoDB 테이블을 생성해야 합니다
# 2. S3 버킷과 DynamoDB 테이블 생성 후 parameter-store.tf에 backend 블록 추가
# 3. terraform init -migrate-state 로 기존 로컬 state를 S3로 마이그레이션

# ============================================
# Variables
# ============================================

variable "state_bucket_name" {
  description = "S3 bucket name for Terraform state"
  type        = string
  default     = "wealist-terraform-state"
}

variable "dynamodb_table_name" {
  description = "DynamoDB table name for state locking"
  type        = string
  default     = "wealist-terraform-locks"
}

# ============================================
# S3 Bucket for Terraform State
# ============================================

resource "aws_s3_bucket" "terraform_state" {
  bucket = var.state_bucket_name

  # 실수로 삭제 방지
  lifecycle {
    prevent_destroy = true
  }

  tags = {
    Name        = "terraform-state-bucket"
    Project     = "wealist"
    Purpose     = "terraform-state"
    ManagedBy   = "terraform"
  }
}

# S3 버킷 버저닝 활성화 (롤백 가능)
resource "aws_s3_bucket_versioning" "terraform_state" {
  bucket = aws_s3_bucket.terraform_state.id

  versioning_configuration {
    status = "Enabled"
  }
}

# S3 버킷 암호화 (AES256)
resource "aws_s3_bucket_server_side_encryption_configuration" "terraform_state" {
  bucket = aws_s3_bucket.terraform_state.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

# S3 버킷 Public Access 차단
resource "aws_s3_bucket_public_access_block" "terraform_state" {
  bucket = aws_s3_bucket.terraform_state.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# S3 버킷 수명 주기 정책 (오래된 버전 삭제)
resource "aws_s3_bucket_lifecycle_configuration" "terraform_state" {
  bucket = aws_s3_bucket.terraform_state.id

  rule {
    id     = "delete-old-versions"
    status = "Enabled"

    noncurrent_version_expiration {
      noncurrent_days = 90  # 90일 이상 된 버전 삭제
    }
  }

  rule {
    id     = "abort-incomplete-uploads"
    status = "Enabled"

    abort_incomplete_multipart_upload {
      days_after_initiation = 7
    }
  }
}

# ============================================
# DynamoDB Table for State Locking
# ============================================

resource "aws_dynamodb_table" "terraform_locks" {
  name         = var.dynamodb_table_name
  billing_mode = "PAY_PER_REQUEST"  # 사용량에 따라 과금 (예측 가능)
  hash_key     = "LockID"

  attribute {
    name = "LockID"
    type = "S"
  }

  # Point-in-time recovery (백업)
  point_in_time_recovery {
    enabled = true
  }

  # 삭제 방지
  lifecycle {
    prevent_destroy = true
  }

  tags = {
    Name        = "terraform-state-lock"
    Project     = "wealist"
    Purpose     = "terraform-locking"
    ManagedBy   = "terraform"
  }
}

# ============================================
# Outputs
# ============================================

output "state_bucket_name" {
  description = "Name of the S3 bucket for Terraform state"
  value       = aws_s3_bucket.terraform_state.bucket
}

output "state_bucket_arn" {
  description = "ARN of the S3 bucket"
  value       = aws_s3_bucket.terraform_state.arn
}

output "dynamodb_table_name" {
  description = "Name of the DynamoDB table for state locking"
  value       = aws_dynamodb_table.terraform_locks.name
}

output "dynamodb_table_arn" {
  description = "ARN of the DynamoDB table"
  value       = aws_dynamodb_table.terraform_locks.arn
}

# ============================================
# 사용 방법 (단계별)
# ============================================

# Step 1: S3 버킷과 DynamoDB 테이블 생성
# ----------------------------------------
# cd docs/terraform
# terraform init
# terraform apply -target=aws_s3_bucket.terraform_state -target=aws_dynamodb_table.terraform_locks
#
# 출력:
# state_bucket_name = "wealist-terraform-state"
# dynamodb_table_name = "wealist-terraform-locks"

# Step 2: parameter-store.tf에 backend 블록 추가
# ----------------------------------------
# parameter-store.tf 파일 상단에 다음 추가:
#
# terraform {
#   backend "s3" {
#     bucket         = "wealist-terraform-state"
#     key            = "parameter-store/dev/terraform.tfstate"
#     region         = "ap-northeast-2"
#     encrypt        = true
#     dynamodb_table = "wealist-terraform-locks"
#   }
# }

# Step 3: 기존 로컬 state를 S3로 마이그레이션
# ----------------------------------------
# terraform init -migrate-state
#
# 확인 메시지:
# Do you want to copy existing state to the new backend?
#   Enter a value: yes
#
# 로컬 terraform.tfstate 파일은 백업 후 삭제 가능

# Step 4: 확인
# ----------------------------------------
# aws s3 ls s3://wealist-terraform-state/parameter-store/dev/
# 출력: terraform.tfstate
#
# terraform state list
# 정상적으로 리소스 목록이 보이면 성공

# ============================================
# 팀 작업 시 주의사항
# ============================================

# 1. S3 버킷 접근 권한 설정
# ----------------------------------------
# 팀원 IAM 사용자/Role에 다음 정책 연결:
#
# {
#   "Version": "2012-10-17",
#   "Statement": [
#     {
#       "Effect": "Allow",
#       "Action": [
#         "s3:ListBucket",
#         "s3:GetObject",
#         "s3:PutObject"
#       ],
#       "Resource": [
#         "arn:aws:s3:::wealist-terraform-state",
#         "arn:aws:s3:::wealist-terraform-state/*"
#       ]
#     },
#     {
#       "Effect": "Allow",
#       "Action": [
#         "dynamodb:GetItem",
#         "dynamodb:PutItem",
#         "dynamodb:DeleteItem"
#       ],
#       "Resource": "arn:aws:dynamodb:ap-northeast-2:*:table/wealist-terraform-locks"
#     }
#   ]
# }

# 2. 동시 작업 시
# ----------------------------------------
# DynamoDB State Lock이 자동으로 처리하므로 걱정 없음
# 한 명이 terraform apply 중이면 다른 팀원은 대기

# 3. State 롤백
# ----------------------------------------
# S3 버전 관리로 이전 버전 복구 가능:
# aws s3api list-object-versions --bucket wealist-terraform-state --prefix parameter-store/dev/terraform.tfstate
# aws s3api get-object --bucket wealist-terraform-state --key parameter-store/dev/terraform.tfstate --version-id VERSION_ID terraform.tfstate.backup

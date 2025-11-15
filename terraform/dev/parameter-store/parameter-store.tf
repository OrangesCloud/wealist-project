# wealist Dev Environment - Parameter Store
# Terraform을 사용한 안전한 비밀 정보 관리

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
  region = var.aws_region
}

# ============================================
# Variables (실제 값은 terraform.tfvars에 저장)
# ============================================

variable "aws_region" {
  description = "AWS Region"
  type        = string
  default     = "ap-northeast-2"
}

variable "postgres_superuser_password" {
  description = "PostgreSQL superuser password"
  type        = string
  sensitive   = true
}

variable "user_db_password" {
  description = "User service database password"
  type        = string
  sensitive   = true
}

variable "board_db_password" {
  description = "Board service database password"
  type        = string
  sensitive   = true
}

variable "redis_password" {
  description = "Redis password"
  type        = string
  sensitive   = true
}

variable "jwt_secret" {
  description = "JWT signing secret (64+ characters for HS512)"
  type        = string
  sensitive   = true
}

variable "google_client_id" {
  description = "Google OAuth Client ID"
  type        = string
}

variable "google_client_secret" {
  description = "Google OAuth Client Secret"
  type        = string
  sensitive   = true
}

variable "grafana_admin_password" {
  description = "Grafana admin password"
  type        = string
  sensitive   = true
}

# ============================================
# Additional Variables (Usernames, DB Names)
# ============================================

variable "postgres_superuser" {
  description = "PostgreSQL superuser username"
  type        = string
  default     = "postgres"
}

variable "user_db_name" {
  description = "User service database name"
  type        = string
  default     = "wealist_user_db"
}

variable "user_db_user" {
  description = "User service database username"
  type        = string
  default     = "wealist_user"
}

variable "board_db_name" {
  description = "Board service database name"
  type        = string
  default     = "wealist_board_db"
}

variable "board_db_user" {
  description = "Board service database username"
  type        = string
  default     = "wealist_board"
}

variable "grafana_admin_user" {
  description = "Grafana admin username"
  type        = string
  default     = "admin"
}

# ============================================
# Parameter Store - PostgreSQL
# ============================================

resource "aws_ssm_parameter" "postgres_superuser_password" {
  name        = "/wealist/dev/postgres/superuser-password"
  description = "PostgreSQL superuser password for wealist dev"
  type        = "SecureString"
  value       = var.postgres_superuser_password

  tags = {
    Project     = "wealist"
    Environment = "dev"
    ManagedBy   = "terraform"
  }
}

# ============================================
# Parameter Store - User Service DB
# ============================================

resource "aws_ssm_parameter" "user_db_password" {
  name        = "/wealist/dev/db/user-password"
  description = "User service database password"
  type        = "SecureString"
  value       = var.user_db_password

  tags = {
    Project     = "wealist"
    Environment = "dev"
    Service     = "user-service"
    ManagedBy   = "terraform"
  }
}

# ============================================
# Parameter Store - Board Service DB
# ============================================

resource "aws_ssm_parameter" "board_db_password" {
  name        = "/wealist/dev/db/board-password"
  description = "Board service database password"
  type        = "SecureString"
  value       = var.board_db_password

  tags = {
    Project     = "wealist"
    Environment = "dev"
    Service     = "board-service"
    ManagedBy   = "terraform"
  }
}

# ============================================
# Parameter Store - Redis
# ============================================

resource "aws_ssm_parameter" "redis_password" {
  name        = "/wealist/dev/redis/password"
  description = "Redis password"
  type        = "SecureString"
  value       = var.redis_password

  tags = {
    Project     = "wealist"
    Environment = "dev"
    ManagedBy   = "terraform"
  }
}

# ============================================
# Parameter Store - JWT
# ============================================

resource "aws_ssm_parameter" "jwt_secret" {
  name        = "/wealist/dev/jwt/secret"
  description = "JWT signing secret"
  type        = "SecureString"
  value       = var.jwt_secret

  tags = {
    Project     = "wealist"
    Environment = "dev"
    ManagedBy   = "terraform"
  }
}

# ============================================
# Parameter Store - Google OAuth
# ============================================

resource "aws_ssm_parameter" "google_client_id" {
  name        = "/wealist/dev/oauth/google-client-id"
  description = "Google OAuth Client ID"
  type        = "String"  # 민감하지 않으므로 String
  value       = var.google_client_id

  tags = {
    Project     = "wealist"
    Environment = "dev"
    ManagedBy   = "terraform"
  }
}

resource "aws_ssm_parameter" "google_client_secret" {
  name        = "/wealist/dev/oauth/google-client-secret"
  description = "Google OAuth Client Secret"
  type        = "SecureString"
  value       = var.google_client_secret

  tags = {
    Project     = "wealist"
    Environment = "dev"
    ManagedBy   = "terraform"
  }
}

# ============================================
# Parameter Store - Grafana
# ============================================

resource "aws_ssm_parameter" "grafana_admin_password" {
  name        = "/wealist/dev/grafana/admin-password"
  description = "Grafana admin password"
  type        = "SecureString"
  value       = var.grafana_admin_password

  tags = {
    Project     = "wealist"
    Environment = "dev"
    ManagedBy   = "terraform"
  }
}

resource "aws_ssm_parameter" "grafana_admin_user" {
  name        = "/wealist/dev/grafana/admin-user"
  description = "Grafana admin username"
  type        = "String"
  value       = var.grafana_admin_user

  tags = {
    Project     = "wealist"
    Environment = "dev"
    ManagedBy   = "terraform"
  }
}

# ============================================
# Parameter Store - Database Usernames & Names
# ============================================

resource "aws_ssm_parameter" "postgres_superuser" {
  name        = "/wealist/dev/postgres/superuser"
  description = "PostgreSQL superuser username"
  type        = "String"
  value       = var.postgres_superuser

  tags = {
    Project     = "wealist"
    Environment = "dev"
    ManagedBy   = "terraform"
  }
}

resource "aws_ssm_parameter" "user_db_name" {
  name        = "/wealist/dev/db/user-name"
  description = "User service database name"
  type        = "String"
  value       = var.user_db_name

  tags = {
    Project     = "wealist"
    Environment = "dev"
    Service     = "user-service"
    ManagedBy   = "terraform"
  }
}

resource "aws_ssm_parameter" "user_db_user" {
  name        = "/wealist/dev/db/user-username"
  description = "User service database username"
  type        = "String"
  value       = var.user_db_user

  tags = {
    Project     = "wealist"
    Environment = "dev"
    Service     = "user-service"
    ManagedBy   = "terraform"
  }
}

resource "aws_ssm_parameter" "board_db_name" {
  name        = "/wealist/dev/db/board-name"
  description = "Board service database name"
  type        = "String"
  value       = var.board_db_name

  tags = {
    Project     = "wealist"
    Environment = "dev"
    Service     = "board-service"
    ManagedBy   = "terraform"
  }
}

resource "aws_ssm_parameter" "board_db_user" {
  name        = "/wealist/dev/db/board-username"
  description = "Board service database username"
  type        = "String"
  value       = var.board_db_user

  tags = {
    Project     = "wealist"
    Environment = "dev"
    Service     = "board-service"
    ManagedBy   = "terraform"
  }
}

# ============================================
# Outputs
# ============================================

output "parameter_names" {
  description = "List of created Parameter Store parameter names"
  value = [
    # Passwords & Secrets
    aws_ssm_parameter.postgres_superuser_password.name,
    aws_ssm_parameter.user_db_password.name,
    aws_ssm_parameter.board_db_password.name,
    aws_ssm_parameter.redis_password.name,
    aws_ssm_parameter.jwt_secret.name,
    aws_ssm_parameter.google_client_id.name,
    aws_ssm_parameter.google_client_secret.name,
    aws_ssm_parameter.grafana_admin_password.name,
    # Usernames & DB Names
    aws_ssm_parameter.postgres_superuser.name,
    aws_ssm_parameter.user_db_name.name,
    aws_ssm_parameter.user_db_user.name,
    aws_ssm_parameter.board_db_name.name,
    aws_ssm_parameter.board_db_user.name,
    aws_ssm_parameter.grafana_admin_user.name,
  ]
}

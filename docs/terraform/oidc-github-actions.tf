# GitHub Actions OIDC Provider for AWS
# 장기 자격증명(Access Key) 없이 GitHub Actions에서 AWS 접근

# ============================================
# Variables
# ============================================

variable "github_org" {
  description = "GitHub Organization or Username"
  type        = string
  default     = "OrangesCloud"  # 실제 GitHub Org/User로 변경
}

variable "github_repo" {
  description = "GitHub Repository name"
  type        = string
  default     = "wealist-project"
}

variable "github_branch" {
  description = "Allowed GitHub branch"
  type        = string
  default     = "deploy-dev"
}

# ============================================
# OIDC Provider
# ============================================

# GitHub OIDC Provider 생성 (AWS 계정당 1번만)
resource "aws_iam_openid_connect_provider" "github_actions" {
  url = "https://token.actions.githubusercontent.com"

  # GitHub Actions의 OIDC 토큰을 신뢰
  client_id_list = [
    "sts.amazonaws.com"
  ]

  # GitHub의 공개 키 지문 (2025년 기준)
  thumbprint_list = [
    "6938fd4d98bab03faadb97b34396831e3780aea1",
    "1c58a3a8518e8759bf075b76b750d4f2df264fcd"
  ]

  tags = {
    Name        = "github-actions-oidc"
    Project     = "wealist"
    ManagedBy   = "terraform"
  }
}

# ============================================
# IAM Role for GitHub Actions
# ============================================

# GitHub Actions가 assume할 수 있는 IAM Role
resource "aws_iam_role" "github_actions_deploy" {
  name = "GitHubActionsDeployRole"

  # Trust Policy: GitHub Actions만 assume 가능
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Federated = aws_iam_openid_connect_provider.github_actions.arn
        }
        Action = "sts:AssumeRoleWithWebIdentity"
        Condition = {
          StringEquals = {
            # OIDC Audience 검증
            "token.actions.githubusercontent.com:aud" = "sts.amazonaws.com"
          }
          StringLike = {
            # 특정 Repository + Branch만 허용
            "token.actions.githubusercontent.com:sub" = "repo:${var.github_org}/${var.github_repo}:ref:refs/heads/${var.github_branch}"
          }
        }
      }
    ]
  })

  tags = {
    Name        = "github-actions-deploy-role"
    Project     = "wealist"
    Environment = "dev"
    ManagedBy   = "terraform"
  }
}

# ============================================
# IAM Policies
# ============================================

# ECR 권한 (이미지 Pull/Push)
resource "aws_iam_policy" "ecr_access" {
  name        = "WealistGitHubActionsECRAccess"
  description = "Allow GitHub Actions to push/pull ECR images"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "ECRAuthentication"
        Effect = "Allow"
        Action = [
          "ecr:GetAuthorizationToken"
        ]
        Resource = "*"
      },
      {
        Sid    = "ECRImageManagement"
        Effect = "Allow"
        Action = [
          "ecr:BatchCheckLayerAvailability",
          "ecr:GetDownloadUrlForLayer",
          "ecr:BatchGetImage",
          "ecr:PutImage",
          "ecr:InitiateLayerUpload",
          "ecr:UploadLayerPart",
          "ecr:CompleteLayerUpload",
          "ecr:DescribeImages",
          "ecr:DescribeRepositories"
        ]
        Resource = [
          "arn:aws:ecr:ap-northeast-2:*:repository/wealist-dev-user-service",
          "arn:aws:ecr:ap-northeast-2:*:repository/wealist-dev-board-service"
        ]
      }
    ]
  })

  tags = {
    Name        = "github-actions-ecr-access"
    Project     = "wealist"
    ManagedBy   = "terraform"
  }
}

# SSM 권한 (EC2 접속 및 명령 실행)
resource "aws_iam_policy" "ssm_access" {
  name        = "WealistGitHubActionsSSMAccess"
  description = "Allow GitHub Actions to execute commands via SSM"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "SSMCommandExecution"
        Effect = "Allow"
        Action = [
          "ssm:SendCommand",
          "ssm:GetCommandInvocation",
          "ssm:DescribeInstanceInformation"
        ]
        Resource = [
          "arn:aws:ssm:ap-northeast-2:*:document/AWS-RunShellScript",
          "arn:aws:ec2:ap-northeast-2:*:instance/*"  # 특정 인스턴스 ID로 제한 권장
        ]
      }
    ]
  })

  tags = {
    Name        = "github-actions-ssm-access"
    Project     = "wealist"
    ManagedBy   = "terraform"
  }
}

# S3 권한 (배포 스크립트 업로드)
resource "aws_iam_policy" "s3_access" {
  name        = "WealistGitHubActionsS3Access"
  description = "Allow GitHub Actions to upload deployment scripts to S3"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "S3BucketManagement"
        Effect = "Allow"
        Action = [
          "s3:CreateBucket",
          "s3:ListBucket"
        ]
        Resource = "arn:aws:s3:::wealist-deploy-scripts"
      },
      {
        Sid    = "S3ObjectManagement"
        Effect = "Allow"
        Action = [
          "s3:PutObject",
          "s3:GetObject"
        ]
        Resource = "arn:aws:s3:::wealist-deploy-scripts/*"
      }
    ]
  })

  tags = {
    Name        = "github-actions-s3-access"
    Project     = "wealist"
    ManagedBy   = "terraform"
  }
}

# ============================================
# Attach Policies to Role
# ============================================

resource "aws_iam_role_policy_attachment" "ecr" {
  role       = aws_iam_role.github_actions_deploy.name
  policy_arn = aws_iam_policy.ecr_access.arn
}

resource "aws_iam_role_policy_attachment" "ssm" {
  role       = aws_iam_role.github_actions_deploy.name
  policy_arn = aws_iam_policy.ssm_access.arn
}

resource "aws_iam_role_policy_attachment" "s3" {
  role       = aws_iam_role.github_actions_deploy.name
  policy_arn = aws_iam_policy.s3_access.arn
}

# ============================================
# Outputs
# ============================================

output "oidc_provider_arn" {
  description = "ARN of the GitHub OIDC Provider"
  value       = aws_iam_openid_connect_provider.github_actions.arn
}

output "github_actions_role_arn" {
  description = "ARN of the GitHub Actions IAM Role (use this in workflows)"
  value       = aws_iam_role.github_actions_deploy.arn
}

output "github_actions_role_name" {
  description = "Name of the GitHub Actions IAM Role"
  value       = aws_iam_role.github_actions_deploy.name
}

# ============================================
# 사용 방법
# ============================================

# 1. Terraform Apply:
#    terraform apply
#
# 2. Output에서 Role ARN 복사:
#    arn:aws:iam::290008131187:role/GitHubActionsDeployRole
#
# 3. GitHub Actions 워크플로우에서 사용:
#
# permissions:
#   id-token: write
#   contents: read
#
# - name: Configure AWS Credentials
#   uses: aws-actions/configure-aws-credentials@v4
#   with:
#     role-to-assume: arn:aws:iam::290008131187:role/GitHubActionsDeployRole
#     aws-region: ap-northeast-2
#
# 4. GitHub Secrets에서 삭제 가능:
#    - WEALIST_DEV_AWS_ACCESS_KEY_ID (삭제!)
#    - WEALIST_DEV_AWS_SECRET_ACCESS_KEY (삭제!)

#!/bin/bash
# =============================================================================
# CodeDeploy Hook: ApplicationStart
# SSM Parameter Store에서 Prod 환경 변수를 로드하고 Docker Compose를 실행합니다.
# =============================================================================

set -euo pipefail

# 1. 상수 정의
# CodeDeploy Agent는 /opt/codedeploy-agent/deployment-root/deployment-group-id/.../deployment-id/deployment-archive 에 파일을 복사합니다.
# 하지만 appspec.yml에서 /home/ubuntu/wealist/로 복사했으므로 그 경로를 사용합니다.
PROJECT_ROOT="/home/ubuntu/wealist"
COMPOSE_FILE="${PROJECT_ROOT}/docker/compose/docker-compose.ec2-prod.yml"
SERVICE_NAME="user-service" # 배포할 서비스 이름

# SSM 경로 접두사 (Terraform에서 정의된)
PARAMETER_BASE_PATH="/wealist/prod"
AWS_REGION="ap-northeast-2" # SSM 호출을 위해 리전 하드코딩 (또는 Instance Metadata 사용)

echo "🚀 User Service Production Deployment Start"
echo "Project Root: ${PROJECT_ROOT}"

# 2. SSM Parameter 로드 함수 정의
# String 타입 로드
load_param() {
    local name="$1"
    aws ssm get-parameter --name "${PARAMETER_BASE_PATH}/${name}" --query 'Parameter.Value' --output text --region "${AWS_REGION}"
}

# SecureString 타입 로드
load_secret() {
    local name="$1"
    aws ssm get-parameter --name "${PARAMETER_BASE_PATH}/${name}" --with-decryption --query 'Parameter.Value' --output text --region "${AWS_REGION}"
}

# 3. 인프라 및 시크릿 환경 변수 로드 및 Export
echo "🔑 Loading secrets and endpoints from SSM Parameter Store..."

# --- 인프라 및 DB 정보 (String) ---
export AWS_ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
export AWS_REGION="${AWS_REGION}" # for Docker Image reference

# DB/Cache 엔드포인트
export RDS_HOST=$(load_param "db/rds_host")
export REDIS_HOST=$(load_param "cache/redis_host")

# DB 이름 및 사용자
export POSTGRES_SUPERUSER=$(load_param "db/rds_master_username") # RDS 마스터 사용자 이름
export USER_DB_NAME=$(load_param "db/user_db_name")
export BOARD_DB_NAME=$(load_param "db/board_db_name")

# --- 시크릿 정보 (SecureString) ---
export JWT_SECRET=$(load_secret "jwt/jwt_secret")
export POSTGRES_SUPERUSER_PASSWORD=$(load_secret "db/rds_master_password")
export REDIS_PASSWORD=$(load_secret "cache/redis_auth_token")

# User Service DB 접속 시크릿
export USER_DB_USER="wealist_user" # SSM에 저장된 경우 로드 필요
export USER_DB_PASSWORD=$(load_secret "db/user_db_password") 

# Board Service DB 접속 시크릿 (User Service에는 필요 없지만 Compose 파일 전체 실행을 위해 로드)
export BOARD_DB_USER="board_service" # SSM에 저장된 경우 로드 필요
export BOARD_DB_PASSWORD=$(load_secret "db/board_db_password")

# OAuth 및 S3 설정
export GOOGLE_CLIENT_ID=$(load_param "oauth/google_client_id")
export GOOGLE_CLIENT_SECRET=$(load_secret "oauth/google-client-secret")
export OAUTH2_CLIENT_REDIRECT_URI=$(load_param "url/oauth2_client_redirect_base")/api/users/login/oauth2/code/google # Base URL로 조합
export OAUTH2_REDIRECT_URL_ENV=$(load_param "url/oauth2_redirect_url")
export S3_BUCKET=$(load_param "s3/bucket")
export S3_REGION="${AWS_REGION}"

# --- Exporter Ports (하드코딩된 경우 SSM 로드 필요 없음) ---
export POSTGRES_EXPORTER_PORT=9187
export REDIS_EXPORTER_PORT=9121
export NODE_EXPORTER_PORT=9100

# 4. 이미지 버전 환경 변수 설정
# CodeDeploy는 배포 시점에 아티팩트 내의 appspec.yml 및 스크립트를 사용합니다.
# Docker Compose는 pull을 사용하여 최신/특정 태그를 가져와야 합니다.
# CodeDeploy는 SHA 태그를 사용하므로, 스크립트에서 SHA 태그를 가져와 사용해야 합니다.
# 🚨 CodeDeploy 아티팩트 내에 이미지 태그 정보를 포함시키는 것이 가장 확실합니다.
# 여기서는 편의상 "latest" 태그를 사용하지만, 실제 Production에서는 SHA 태그를 사용해야 합니다.
export USER_SERVICE_VERSION="latest" # CodeDeploy가 사용하는 SHA Tag로 대체되어야 함
export BOARD_SERVICE_VERSION="latest" # CodeDeploy가 사용하는 SHA Tag로 대체되어야 함


echo "✅ Parameters loaded. Starting Docker Compose..."
echo "RDS Host: ${RDS_HOST}"
echo "Redis Host: ${REDIS_HOST}"
# echo "JWT Secret: ${JWT_SECRET}" # 시크릿 값은 출력하지 않습니다.

# 5. Docker Compose 실행
# Docker Compose 명령어 결정
if docker compose version &> /dev/null; then
  COMPOSE_CMD="docker compose"
else
  COMPOSE_CMD="docker-compose"
fi

# ECR 로그인 (EC2 인스턴스 IAM Role에 권한이 있어야 함)
aws ecr get-login-password --region ${AWS_REGION} | \
  docker login --username AWS --password-stdin ${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com

# 6. 최신 이미지 Pull (user-service만)
echo "🐳 Pulling latest image for ${SERVICE_NAME}..."
$COMPOSE_CMD -f "${COMPOSE_FILE}" pull "${SERVICE_NAME}"

# 7. Docker Compose 실행 (user-service와 Exporter들 재시작)
# Prod 환경에서는 user-service, board-service, exporter들을 모두 관리해야 합니다.
# user-service와 board-service는 종속성(depends_on)을 통해 순서대로 재시작됩니다.
echo "🔄 Starting services defined in ${COMPOSE_FILE}..."
$COMPOSE_CMD -f "${COMPOSE_FILE}" up -d --no-deps --force-recreate "${SERVICE_NAME}" # user-service
$COMPOSE_CMD -f "${COMPOSE_FILE}" up -d --no-deps --force-recreate "postgres-exporter" 
$COMPOSE_CMD -f "${COMPOSE_FILE}" up -d --no-deps --force-recreate "redis-exporter" 
$COMPOSE_CMD -f "${COMPOSE_FILE}" up -d --no-deps --force-recreate "node-exporter" 

echo "✅ Deployment initiated. CodeDeploy will now run ValidateService."

# Note: 이 스크립트가 성공적으로 종료되면 CodeDeploy는 ValidateService Hook을 실행합니다.
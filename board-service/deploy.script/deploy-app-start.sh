#!/bin/bash
# =============================================================================
# CodeDeploy Hook: ApplicationStart (Board Service)
# SSM Parameter Store에서 Prod 환경 변수를 로드하고 Docker Compose를 실행합니다.
# =============================================================================

set -euo pipefail

# 1. 상수 정의
PROJECT_ROOT="/home/ubuntu/wealist"
COMPOSE_FILE="${PROJECT_ROOT}/docker-compose.ec2-prod.yml" # appspec에서 루트에 복사했으므로 경로 수정
SERVICE_NAME="board-service" # 배포할 서비스 이름

# SSM 경로 접두사 및 리전
PARAMETER_BASE_PATH="/wealist/prod"
AWS_REGION="ap-northeast-2"

echo "🚀 Board Service Production Deployment Start"
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

# 3. 인프라 및 시크릿 환경 변수 로드 및 Export (Compose 파일 실행을 위해 모든 변수 필요)
echo "🔑 Loading secrets and endpoints from SSM Parameter Store..."

# --- 인프라 및 DB 정보 (String) ---
export AWS_ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
export AWS_REGION="${AWS_REGION}" 

# DB/Cache 엔드포인트
export RDS_HOST=$(load_param "db/rds_host")
export REDIS_HOST=$(load_param "cache/redis_host")

# DB 이름 및 사용자
export POSTGRES_SUPERUSER=$(load_param "db/rds_master_username")
export USER_DB_NAME=$(load_param "db/user_db_name")
export BOARD_DB_NAME=$(load_param "db/board_db_name")

# --- 시크릿 정보 (SecureString) ---
export JWT_SECRET=$(load_secret "jwt/jwt_secret")
export POSTGRES_SUPERUSER_PASSWORD=$(load_secret "db/rds_master_password")
export REDIS_PASSWORD=$(load_secret "cache/redis_auth_token")

# User Service DB 접속 시크릿
export USER_DB_USER="wealist_user"
export USER_DB_PASSWORD=$(load_secret "db/user_db_password") 

# Board Service DB 접속 시크릿
export BOARD_DB_USER="board_service"
export BOARD_DB_PASSWORD=$(load_secret "db/board_db_password")

# OAuth 및 S3 설정
export GOOGLE_CLIENT_ID=$(load_param "oauth/google_client_id")
export GOOGLE_CLIENT_SECRET=$(load_secret "oauth/google-client-secret")
export OAUTH2_CLIENT_REDIRECT_URI=$(load_param "url/oauth2_client_redirect_base")/api/users/login/oauth2/code/google 
export OAUTH2_REDIRECT_URL_ENV=$(load_param "url/oauth2_redirect_url")
export S3_BUCKET=$(load_param "s3/bucket")
export S3_REGION="${AWS_REGION}"

# --- Exporter Ports ---
export POSTGRES_EXPORTER_PORT=9187
export REDIS_EXPORTER_PORT=9121
export NODE_EXPORTER_PORT=9100

# 4. 이미지 버전 환경 변수 설정
# 🚨 CodeDeploy 아티팩트 내에 실제 SHA 태그가 포함되어야 하지만, 현재는 latest로 가정합니다.
export USER_SERVICE_VERSION="latest" 
export BOARD_SERVICE_VERSION="latest" # Board Service의 배포 태그를 사용

echo "✅ Parameters loaded. Starting Docker Compose..."

# 5. Docker Compose 명령어 결정 및 ECR 로그인
if docker compose version &> /dev/null; then
  COMPOSE_CMD="docker compose"
else
  COMPOSE_CMD="docker-compose"
fi

aws ecr get-login-password --region ${AWS_REGION} | \
  docker login --username AWS --password-stdin ${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com

# 6. 최신 이미지 Pull (board-service만)
echo "🐳 Pulling latest image for ${SERVICE_NAME}..."
$COMPOSE_CMD -f "${COMPOSE_FILE}" pull "${SERVICE_NAME}"

# 7. Docker Compose 실행 (board-service와 Exporter들 재시작)
# user-service는 board-service의 depends_on 조건에 의해 영향을 받지 않도록 --no-deps를 사용합니다.
echo "🔄 Starting services defined in ${COMPOSE_FILE}..."
$COMPOSE_CMD -f "${COMPOSE_FILE}" up -d --no-deps --force-recreate "${SERVICE_NAME}" # board-service
$COMPOSE_CMD -f "${COMPOSE_FILE}" up -d --no-deps --force-recreate "postgres-exporter" 
$COMPOSE_CMD -f "${COMPOSE_FILE}" up -d --no-deps --force-recreate "redis-exporter" 
$COMPOSE_CMD -f "${COMPOSE_FILE}" up -d --no-deps --force-recreate "node-exporter" 

echo "✅ Deployment initiated. CodeDeploy will now run ValidateService."
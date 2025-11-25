# wealist-project2/user-service/deploy.script/deploy-app-start.sh
#!/bin/bash

# =============================================================================
# CodeDeploy ApplicationStart Hook Script (User Service)
# 역할: Parameter Store에서 환경 변수 로드, ECR 로그인, 인프라 및 서비스 재시작
# =============================================================================

# 에러 발생 시 즉시 중단
set -euo pipefail

# -----------------------------------------------------------------------------
# 1. 초기 설정 및 경로 정의
# -----------------------------------------------------------------------------
PROJECT_ROOT="/home/ubuntu/wealist-app"
COMPOSE_FILE="${PROJECT_ROOT}/docker/compose/docker-compose.ec2-prod.yml"
PARAMETER_PREFIX="/wealist/dev" 
AWS_REGION="ap-northeast-2" 
CONTAINER_NAME="wealist-user-service"

echo "🚀 Starting User Service Deployment via CodeDeploy..."
echo "📅 Started at: $(date '+%Y-%m-%d %H:%M:%S')"

# -----------------------------------------------------------------------------
# 2. SSM 파라미터 로드 함수 (EC2 인스턴스 IAM 역할을 사용)
# -----------------------------------------------------------------------------
load_param() {
    local param_name="$1"
    local full_param_path="${PARAMETER_PREFIX}/${param_name}"
    local value
    
    value=$(aws ssm get-parameter \
      --name "${full_param_path}" \
      --query 'Parameter.Value' \
      --output text \
      --region ${AWS_REGION} 2>/dev/null)
    local exit_code=$?
    
    if [ $exit_code -ne 0 ] || [ -z "$value" ] || [ "$value" = "None" ]; then
      echo "❌ Failed to load parameter: ${full_param_path}" >&2
      exit 1
    fi
    echo "$value"
}

load_secret() {
    local param_name="$1"
    local full_param_path="${PARAMETER_PREFIX}/${param_name}"
    local value
    
    value=$(aws ssm get-parameter \
      --name "${full_param_path}" \
      --with-decryption \
      --query 'Parameter.Value' \
      --output text \
      --region ${AWS_REGION} 2>/dev/null)
    local exit_code=$?
    
    if [ $exit_code -ne 0 ] || [ -z "$value" ] || [ "$value" = "None" ]; then
      echo "❌ Failed to load secret parameter: ${full_param_path}" >&2
      exit 1
    fi
    echo "$value"
}

# -----------------------------------------------------------------------------
# 3. 환경 변수 로드 및 Export
# -----------------------------------------------------------------------------
echo "📥 Loading environment variables for User Service..."

# ECR 이미지 정보
export AWS_ACCOUNT_ID=$(load_param "ci/aws_account_id")
export ECR_REPOSITORY_USER=$(load_param "ci/ecr_repository_user")
export AWS_REGION="${AWS_REGION}"
export VERSION="latest" 
export JPA_DDL_AUTO="update"

# DB 및 기타 비밀 값 로드 (기존 CD 스크립트와 동일)
export POSTGRES_SUPERUSER=$(load_param "db/postgres_superuser")
export POSTGRES_SUPERUSER_PASSWORD=$(load_secret "db/postgres_superuser_password")

export USER_DB_NAME=$(load_param "db/user_db_name")
export USER_DB_USER=$(load_param "db/user_db_user")
export USER_DB_PASSWORD=$(load_secret "db/user_db_password")

export BOARD_DB_NAME=$(load_param "db/board_db_name")
export BOARD_DB_USER=$(load_param "db/board_db_user")
export BOARD_DB_PASSWORD=$(load_secret "db/board_db_password")

export REDIS_PASSWORD=$(load_secret "cache/redis_password")
export JWT_SECRET=$(load_secret "jwt/jwt_secret")
export JWT_ACCESS_TOKEN_EXPIRATION_MS="1800000"
export JWT_REFRESH_TOKEN_EXPIRATION_MS="604800000"

export GOOGLE_CLIENT_ID=$(load_secret "oauth/google_client_id")
export GOOGLE_CLIENT_SECRET=$(load_secret "oauth/google-client-secret")
export OAUTH2_CLIENT_REDIRECT_BASE=$(load_param "url/oauth2_client_redirect_base")
export OAUTH2_REDIRECT_URL_ENV=$(load_param "url/oauth2_redirect_url")
export OAUTH2_CLIENT_REDIRECT_URI="${OAUTH2_CLIENT_REDIRECT_BASE}/api/users/login/oauth2/code/google"
export USER_SERVICE_URL=$(load_param "service/user_service_url")
export S3_BUCKET=$(load_param "s3/bucket")
export S3_REGION=$(load_param "s3/region")
export LOG_LEVEL="info"
export ENVIRONMENT="dev"
export CORS_ORIGINS="*" 
export APP_NAME="wealist"

echo "✅ Environment variables loaded successfully"
echo "📋 Loaded variables summary (Redacted):"
echo "   - AWS_ACCOUNT_ID: ${AWS_ACCOUNT_ID}"
echo "   - POSTGRES_SUPERUSER: ${POSTGRES_SUPERUSER}"
echo "   - USER_DB_NAME: ${USER_DB_NAME}"
echo "   - BOARD_DB_NAME: ${BOARD_DB_NAME}"
echo "   - S3_BUCKET: ${S3_BUCKET}"

# -----------------------------------------------------------------------------
# 4. ECR 로그인 및 최신 이미지 Pull (user-service)
# -----------------------------------------------------------------------------
ECR_REGISTRY="${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com"
USER_IMAGE_NAME="${ECR_REGISTRY}/${ECR_REPOSITORY_USER}:latest"

echo "🔑 Logging into Amazon ECR..."
aws ecr get-login-password --region ${AWS_REGION} | \
  docker login --username AWS --password-stdin ${ECR_REGISTRY}

echo "📥 Pulling latest user-service image: ${USER_IMAGE_NAME}..."
docker pull ${USER_IMAGE_NAME} 
echo "✅ Image pulled successfully"

# -----------------------------------------------------------------------------
# 5. Docker Compose 명령어 감지 및 인프라 서비스 확인/시작
# -----------------------------------------------------------------------------
if command -v docker-compose &> /dev/null; then
  COMPOSE_CMD="docker-compose"
elif docker compose version &> /dev/null 2>&1; then
  COMPOSE_CMD="docker compose"
else
  echo "  ❌ Docker Compose not found" >&2
  exit 1
fi

echo "🔍 Ensuring infrastructure services (Postgres, Redis) are running..."
$COMPOSE_CMD --env-file <(printenv) -f ${COMPOSE_FILE} up -d postgres redis

# PostgreSQL 헬스체크
echo "🏥 Checking PostgreSQL health..."
POSTGRES_READY=false
for i in {1..6}; do
  if docker exec wealist-postgres pg_isready -U ${POSTGRES_SUPERUSER} > /dev/null 2>&1; then
    echo "  ✅ PostgreSQL is ready (attempt $i/6)"
    POSTGRES_READY=true
    break
  fi
  sleep 5
done
[ "$POSTGRES_READY" = false ] && { echo "  ❌ PostgreSQL failed to become ready" >&2; exit 1; }

# Redis 헬스체크
echo "🏥 Checking Redis health..."
REDIS_READY=false
for i in {1..6}; do
  if docker exec wealist-redis redis-cli -a "${REDIS_PASSWORD}" ping > /dev/null 2>&1; then
    echo "  ✅ Redis is ready (attempt $i/6)"
    REDIS_READY=true
    break
  fi
  sleep 5
done
[ "$REDIS_READY" = false ] && { echo "  ❌ Redis failed to become ready" >&2; exit 1; }
echo "✅ Infrastructure services are ready"

# -----------------------------------------------------------------------------
# 6. 데이터베이스 유저 및 데이터베이스 생성 
# -----------------------------------------------------------------------------
echo "🗄️  Setting up database users and databases..."

# Board DB 유저 생성 (이미 존재하면 무시)
docker exec wealist-postgres psql -U ${POSTGRES_SUPERUSER} -d postgres -c "
DO \$\$
BEGIN
  IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = '${BOARD_DB_USER}') THEN
    CREATE ROLE ${BOARD_DB_USER} WITH LOGIN PASSWORD '${BOARD_DB_PASSWORD}';
  END IF;
END
\$\$;
" 2>&1

# User DB 유저 생성 (이미 존재하면 무시)
docker exec wealist-postgres psql -U ${POSTGRES_SUPERUSER} -d postgres -c "
DO \$\$
BEGIN
  IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = '${USER_DB_USER}') THEN
    CREATE ROLE ${USER_DB_USER} WITH LOGIN PASSWORD '${USER_DB_PASSWORD}';
  END IF;
END
\$\$;
" 2>&1

# Board 데이터베이스 확인 및 생성
BOARD_DB_EXISTS=$(docker exec wealist-postgres psql -U ${POSTGRES_SUPERUSER} -d postgres -tAc "SELECT 1 FROM pg_database WHERE datname='${BOARD_DB_NAME}';" 2>&1)
if [ "$BOARD_DB_EXISTS" != "1" ]; then
  docker exec wealist-postgres psql -U ${POSTGRES_SUPERUSER} -d postgres -c "CREATE DATABASE ${BOARD_DB_NAME} OWNER ${BOARD_DB_USER};" 2>&1
else
  docker exec wealist-postgres psql -U ${POSTGRES_SUPERUSER} -d postgres -c "GRANT ALL PRIVILEGES ON DATABASE ${BOARD_DB_NAME} TO ${BOARD_DB_USER};" 2>&1
fi

# User 데이터베이스 확인 및 생성
USER_DB_EXISTS=$(docker exec wealist-postgres psql -U ${POSTGRES_SUPERUSER} -d postgres -tAc "SELECT 1 FROM pg_database WHERE datname='${USER_DB_NAME}';" 2>&1)
if [ "$USER_DB_EXISTS" != "1" ]; then
  docker exec wealist-postgres psql -U ${POSTGRES_SUPERUSER} -d postgres -c "CREATE DATABASE ${USER_DB_NAME} OWNER ${USER_DB_USER};" 2>&1
else
  docker exec wealist-postgres psql -U ${POSTGRES_SUPERUSER} -d postgres -c "GRANT ALL PRIVILEGES ON DATABASE ${USER_DB_NAME} TO ${USER_DB_USER};" 2>&1
fi

echo "✅ All databases ready"

# -----------------------------------------------------------------------------
# 7. User Service 재시작 (Core Deployment)
# -----------------------------------------------------------------------------
echo "🔄 Restarting user-service..."
cd "${PROJECT_ROOT}"

if ! $COMPOSE_CMD --env-file <(printenv) -f ${COMPOSE_FILE} up -d --force-recreate user-service; then
    echo "❌ Failed to restart user-service via Docker Compose" >&2
    $COMPOSE_CMD -f ${COMPOSE_FILE} logs --tail=50 user-service 2>&1 >&2
    exit 1
fi
echo "✅ User service container recreated successfully"
echo "✅ ApplicationStart completed."
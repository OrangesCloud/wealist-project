#!/bin/bash
# =============================================================================
# CodeDeploy Hook: ApplicationStart (User Service)
# [최적화 버전] SSM 일괄 로드, sleep 단축 및 OAuth 하드코딩 (디버깅용)
# =============================================================================

set -euo pipefail

# 1. 상수 정의
PROJECT_ROOT="/home/ec2-user/wealist"
COMPOSE_FILE="${PROJECT_ROOT}/docker/compose/docker-compose.ec2-prod.yml"
SERVICE_NAME="user-service" # 배포할 서비스 이름
AWS_REGION="ap-northeast-2" 
PARAMETER_BASE_PATH="/wealist/prod"
# ⚠️ CloudFront ID를 실제 값으로 변경하세요.
CLOUDFRONT_DISTRIBUTION_ID="E1Z5D8FUFEI6KI" 
INVALIDATION_PATHS="/*" 

echo "🚀 User Service Production Deployment Start (Optimized)"
echo "Project Root: ${PROJECT_ROOT}"

export AWS_DEFAULT_REGION="${AWS_REGION}" 

# 2. SSM Parameter 전체 로드 및 파싱 함수 정의 (jq 필수)
# String 및 SecureString 파라미터 이름을 정의합니다.
STRING_PARAM_NAMES="db/rds_host,cache/redis_host,cache/redis_port,db/rds_master_username,s3/bucket,version/user_service,version/board_service"
SECRET_PARAM_NAMES="db/user_db_name,db/board_db_name,jwt/jwt_secret,db/rds_master_password,cache/redis_auth_token,db/user_db_user,db/user_db_password,db/board_db_user,db/board_db_password"

load_all_parameters() {
    local names_string="$1"
    local with_decryption_flag="$2"
    
    IFS=',' read -ra NAMES_ARRAY <<< "$names_string"
    
    local NAMES_OPTIONS=""
    for name in "${NAMES_ARRAY[@]}"; do
        NAMES_OPTIONS+=" --names ${PARAMETER_BASE_PATH}/${name}"
    done

    local SSM_OUTPUT
    # aws ssm get-parameters 호출 (API 호출 횟수 대폭 감소)
    SSM_OUTPUT=$(aws ssm get-parameters ${NAMES_OPTIONS} ${with_decryption_flag} --query 'Parameters[*].{Name:Name,Value:Value}' --output json 2>/dev/null)
    
    # JSON 파싱 및 환경 변수 설정 (jq 사용)
    if [ -n "$SSM_OUTPUT" ]; then
        echo "$SSM_OUTPUT" | jq -r '.[] | .Name as $name | .Value as $value | "\($name | split("/") | last | ascii_upcase)=\($value)"'
    fi
}

# 3. IMDS 로드 대기 및 환경 변수 로드 시작
echo "⏳ Waiting for IAM Role credentials to load via IMDS (5s delay)..."
sleep 5  # <--- 15초에서 5초로 단축 (속도 개선)

echo "🔑 Loading secrets and endpoints from SSM Parameter Store (Batch Fetch)..."

# String 파라미터 로드
echo "  - Loading String parameters..."
eval "$(load_all_parameters "${STRING_PARAM_NAMES}" "")"

# SecureString 파라미터 로드
echo "  - Loading SecureString parameters..."
eval "$(load_all_parameters "${SECRET_PARAM_NAMES}" "--with-decryption")"

# AWS_ACCOUNT_ID 추출 및 검증 (유지)
export AWS_ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text 2>/dev/null)
if [ -z "$AWS_ACCOUNT_ID" ] || [ "$AWS_ACCOUNT_ID" == "null" ]; then
    echo "❌ FATAL: Could not retrieve AWS Account ID using STS."
    exit 1
fi
export AWS_REGION="${AWS_REGION}" 
echo "✅ AWS Account ID loaded: ${AWS_ACCOUNT_ID}"

# REDIS_PASSWORD가 NONE이면 빈 문자열로 처리 (논리 유지)
if [ "${REDIS_PASSWORD}" == "NONE" ]; then
    export REDIS_PASSWORD=""
fi

# 🚨🚨🚨 OAuth 및 S3 설정 (하드코딩 - 디버깅 모드 유지) 🚨🚨🚨
# **500 에러 진단이 완료되면 SSM 로직으로 복원해야 합니다.**
export GOOGLE_CLIENT_ID="640996696843-gtht74dpnn9c4u7mb6k5craur1vojgbk.apps.googleusercontent.com"
export GOOGLE_CLIENT_SECRET=$(load_param "oauth/google_client_secret")
# 백엔드 서버의 콜백 URI (Google Console에 등록해야 하는 URI)
export OAUTH2_CLIENT_REDIRECT_URI="https://api.wealist.co.kr/api/users/login/oauth2/code/google"
# 프론트엔드로의 최종 리다이렉션 URI
export OAUTH2_REDIRECT_URL="http://wealist.co.kr/oauth/callback"

export S3_REGION="${AWS_REGION}"

# --- OAuth 디버깅 로그 (유지) ---
echo "--- OAuth Parameter Check (Hardcoded) ---"
echo "GOOGLE_CLIENT_ID (Prefix): ${GOOGLE_CLIENT_ID:0:5}..."
echo "OAUTH2_CLIENT_REDIRECT_URI (Server Callback): ${OAUTH2_CLIENT_REDIRECT_URI}"
echo "OAUTH2_REDIRECT_URL (Client Final Redirect): ${OAUTH2_REDIRECT_URL}"
echo "---------------------------------------"
# ------------------------------

# --- Exporter Ports ---
export NODE_EXPORTER_PORT=9100

# 4. 이미지 버전 환경 변수 설정 (SSM 로드 결과를 사용)
if [ -z "${USER_SERVICE_VERSION:-}" ]; then
  echo "⚠️ User Service version not found in SSM. Using latest tag."
  export USER_SERVICE_VERSION="latest";
fi
echo "User Service Tag: ${USER_SERVICE_VERSION}"

# 5. Docker Compose 실행 (유지)
if docker compose version &> /dev/null; then
  COMPOSE_CMD="docker compose"
else
  COMPOSE_CMD="docker-compose"
fi

echo "🐳 Logging into ECR..."
aws ecr get-login-password --region ${AWS_REGION} | \
  sudo docker login --username AWS --password-stdin ${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com

echo "🐳 Pulling image: ${SERVICE_NAME}:${USER_SERVICE_VERSION}"
sudo -E $COMPOSE_CMD -f "${COMPOSE_FILE}" pull "${SERVICE_NAME}"

# 7. Docker Compose 실행 (User Service만 재시작 - 간결화 로직 유지)
echo "🔄 Starting User Service defined in ${COMPOSE_FILE}..."
SERVICES_TO_RESTART="user-service"

echo "🛑 Stopping and removing existing User Service container (direct Docker commands)..."
if sudo docker ps -a | grep -q "wealist-user-service"; then
    echo "  Found existing user-service container, removing it..."
    sudo docker stop wealist-user-service 2>/dev/null || true
    sudo docker rm -f wealist-user-service 2>/dev/null || true
    echo "  ✅ Old container removed"
else
    echo "  No existing user-service container found"
fi

# 컨테이너 정리 대기
sleep 1 # <--- 2초에서 1초로 단축 (속도 개선)

echo "🚀 Starting new user-service container..."
sudo -E $COMPOSE_CMD -f "${COMPOSE_FILE}" up -d --no-deps --force-recreate ${SERVICES_TO_RESTART}

# 컨테이너 시작 대기
sleep 1 # <--- 3초에서 1초로 단축 (속도 개선)


# Node Exporter가 실행 중이 아니면 시작
echo "🔍 Ensuring node-exporter is running..."
sudo -E $COMPOSE_CMD -f "${COMPOSE_FILE}" up -d --no-deps node-exporter 2>/dev/null || true


# 8. CloudFront 캐시 무효화 (추가된 기능 - IAM 권한 필수)
echo "☁️ Initiating CloudFront Invalidation for ${CLOUDFRONT_DISTRIBUTION_ID}..."

INVALIDATION_OUTPUT=$(
    aws cloudfront create-invalidation \
        --distribution-id "${CLOUDFRONT_DISTRIBUTION_ID}" \
        --paths "${INVALIDATION_PATHS}" \
        --query 'Invalidation.Id' \
        --output text 2>/dev/null
)

if [ -n "$INVALIDATION_OUTPUT" ]; then
    echo "✅ CloudFront Invalidation started (ID: ${INVALIDATION_OUTPUT}). This may take a few minutes."
else
    echo "❌ Failed to initiate CloudFront Invalidation. Check IAM permissions and ensure the distribution ID is correct."
fi


echo "✅ Deployment initiated. CodeDeploy will now run ValidateService."
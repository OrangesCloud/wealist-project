#!/bin/bash
# =============================================================================
# CodeDeploy Hook: ApplicationStart (Board Service)
# .env 파일 기반으로 환경 변수 주입 → SSM 완전 배제 (가장 안정적!)
# =============================================================================

set -euo pipefail

echo "=== 디버깅: /home/ec2-user/wealist 디렉터리 구조 ==="
ls -la /home/ec2-user/wealist/
echo "=== docker-compose 파일 존재 여부 ==="
ls -la /home/ec2-user/wealist/docker-compose.ec2-prod.yml && echo "존재함" || echo "없음"

# 1. 상수 정의
PROJECT_ROOT="/home/ec2-user/wealist"
COMPOSE_FILE="${PROJECT_ROOT}/docker-compose.ec2-prod.yml"
SERVICE_NAME="board-service"
AWS_REGION="ap-northeast-2"

echo "Board Service Production Deployment Start"
echo "Project Root: ${PROJECT_ROOT}"

# Docker Compose 파일 존재 확인
if [ ! -f "${COMPOSE_FILE}" ]; then
    echo "ERROR: Docker Compose file not found at ${COMPOSE_FILE}"
    exit 1
fi
echo "Docker Compose file found: ${COMPOSE_FILE}"

# =============================================================================
# 2. .env 파일에서 환경 변수 로드 (이게 핵심! CI에서 정확한 태그와 함께 만들어짐)
# =============================================================================
ENV_FILE="${PROJECT_ROOT}/.env"

if [ -f "${ENV_FILE}" ]; then
    echo ".env 파일 발견 → 환경 변수 주입 시작"
    # 주석(#)으로 시작하는 줄과 빈 줄 무시하면서 로드
    export $(grep -v '^#' "${ENV_FILE}" | xargs)
    echo ".env 파일 로드 완료"
else
    echo "경고: .env 파일이 없습니다. 기본값으로 진행합니다."
fi

# =============================================================================
# 3. 필수 환경 변수 보장 (CI에서 안 넣어줬을 경우 대비)
# =============================================================================
export BOARD_SERVICE_VERSION=${BOARD_SERVICE_VERSION:-latest}
# export USER_SERVICE_VERSION=${USER_SERVICE_VERSION:-latest}

echo "최종 사용할 이미지 태그"
echo "   → Board Service : ${BOARD_SERVICE_VERSION}"
# echo "   → User Service  : ${USER_SERVICE_VERSION}"

# =============================================================================
# 4. 나머지 환경 변수들은 docker-compose.yml + CodeDeploy Hook에서 SSM으로 주입
#     (기존 방식 그대로 유지 → DB, Redis, Secret 등은 여전히 SSM에서 오게 됨)
# =============================================================================
export AWS_DEFAULT_REGION="${AWS_REGION}"

# AWS_ACCOUNT_ID 추출 (ECR 로그인용)
export AWS_ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text 2>/dev/null || echo "")
if [ -z "$AWS_ACCOUNT_ID" ] || [ "$AWS_ACCOUNT_ID" == "null" ]; then
    echo "FATAL: AWS_ACCOUNT_ID를 가져올 수 없습니다. IAM 역할 확인 필요."
    exit 1
fi
echo "AWS Account ID: ${AWS_ACCOUNT_ID}"

# =============================================================================
# 5. Docker Compose 명령어 결정
# =============================================================================
if docker compose version &> /dev/null; then
    COMPOSE_CMD="docker compose"
else
    COMPOSE_CMD="docker-compose"
fi

# =============================================================================
# 6. ECR 로그인
# =============================================================================
echo "ECR 로그인 중..."
password=$(aws ecr get-login-password --region ${AWS_REGION})
echo "${password}" | sudo docker login --username AWS --password-stdin ${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com
echo "ECR 로그인 완료"

# =============================================================================
# 7. 최신 Board Service 이미지 Pull
# =============================================================================
echo "Pulling ${SERVICE_NAME}:${BOARD_SERVICE_VERSION} ..."
sudo $COMPOSE_CMD -f "${COMPOSE_FILE}" pull "${SERVICE_NAME}"
echo "Pull 완료"

# =============================================================================
# 8. 서비스 재시작 (Board Service + Exporters만)
# =============================================================================
echo "Board Service 및 Exporters 재시작 중..."
SERVICES_TO_RESTART="board-service postgres-exporter redis-exporter node-exporter"

sudo $COMPOSE_CMD -f "${COMPOSE_FILE}" up -d --no-deps --force-recreate ${SERVICES_TO_RESTART}

echo "배포 완료! CodeDeploy가 ValidateService를 실행합니다."
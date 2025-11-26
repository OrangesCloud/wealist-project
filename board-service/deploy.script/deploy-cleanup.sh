#!/bin/bash
# CodeDeploy Hook: BeforeInstall
# 이전 배포에서 남은 Board Service 컨테이너 중지 및 이미지 정리

set +e # 오류가 나도 계속 진행 (컨테이너가 없을 수 있음)

PROJECT_ROOT="/home/ec2-user/wealist"
COMPOSE_FILE="${PROJECT_ROOT}/docker-compose.ec2-prod.yml"
echo "🧹 Cleaning up old board-service containers..."

# 1. Docker Compose 명령어가 무엇인지 확인합니다.
if docker compose version &> /dev/null; then
  COMPOSE_CMD="docker compose"
else
  COMPOSE_CMD="docker-compose"
fi

# 2. 기존 메인 서비스 컨테이너 중지 및 삭제
# [수정] 모든 Compose 명령 앞에 sudo 추가
sudo $COMPOSE_CMD -f "${COMPOSE_FILE}" stop board-service || true
sudo $COMPOSE_CMD -f "${COMPOSE_FILE}" rm -f board-service || true

# 3. 🧹 임시 헬스체크 컨테이너 정리
echo "🧹 Cleaning up temporary health check containers to prevent port conflicts..."

# User Service와 Board Service의 임시 컨테이너를 모두 정리해야 합니다.
CONTAINERS_TO_CLEANUP="temp-board-health temp-user-health" 

for CONTAINER in ${CONTAINERS_TO_CLEANUP}; do
    echo "  - Attempting to remove ${CONTAINER}..."
    sudo docker rm -f "${CONTAINER}" || true 
done

echo "✅ Cleanup complete."
set -e
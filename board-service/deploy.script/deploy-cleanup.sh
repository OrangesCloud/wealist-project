#!/bin/bash
# CodeDeploy Hook: BeforeInstall
# 이전 배포에서 남은 Board Service 컨테이너 중지 및 이미지 정리

set +e # 오류가 나도 계속 진행 (컨테이너가 없을 수 있음)

# [수정 필요] 경로를 /home/ec2-user/wealist로 변경
PROJECT_ROOT="/home/ec2-user/wealist"
# [수정 필요] docker-compose.ec2-prod.yml의 위치를 board-service 디렉토리 내로 지정
COMPOSE_FILE="${PROJECT_ROOT}/board-service/docker-compose.ec2-prod.yml"
echo "🧹 Cleaning up old board-service containers..."

if docker compose version &> /dev/null; then
  COMPOSE_CMD="docker compose"
else
  COMPOSE_CMD="docker-compose"
fi

# 1. 기존 메인 서비스 컨테이너 중지 및 삭제
$COMPOSE_CMD -f "${COMPOSE_FILE}" stop board-service || true
$COMPOSE_CMD -f "${COMPOSE_FILE}" rm -f board-service || true

# 2. 🧹 임시 헬스체크 컨테이너 정리 (추가된 부분)
echo "🧹 Cleaning up temporary health check containers to prevent port conflicts..."

CONTAINERS_TO_CLEANUP="temp-board-health"

for CONTAINER in ${CONTAINERS_TO_CLEANUP}; do
    echo "  - Attempting to remove ${CONTAINER}..."
    # -f 옵션: 컨테이너가 실행 중이든 정지 상태든 강제로 정지 및 삭제를 시도합니다.
    # || true: 해당 컨테이너가 존재하지 않아 에러가 발생해도 스크립트 실행을 계속합니다.
    docker rm -f "${CONTAINER}" || true 
done
echo "✅ Cleanup complete."
set -e
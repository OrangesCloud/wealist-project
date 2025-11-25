#!/bin/bash
# CodeDeploy Hook: BeforeInstall
# 이전 배포에서 남은 컨테이너 중지 및 이미지 정리

set +e # 오류가 나도 계속 진행 (컨테이너가 없을 수 있음)

PROJECT_ROOT="/home/ubuntu/wealist"
COMPOSE_FILE="${PROJECT_ROOT}/docker/compose/docker-compose.ec2-prod.yml"

echo "🧹 Cleaning up old containers..."

if docker compose version &> /dev/null; then
  COMPOSE_CMD="docker compose"
else
  COMPOSE_CMD="docker-compose"
fi

# 기존 서비스만 중지 (전체 다운은 하지 않음)
$COMPOSE_CMD -f "${COMPOSE_FILE}" stop user-service || true

# 컨테이너 삭제
$COMPOSE_CMD -f "${COMPOSE_FILE}" rm -f user-service || true

# 사용하지 않는 이미지 정리 (선택적)
# docker image prune -f

echo "✅ Cleanup complete."
set -e
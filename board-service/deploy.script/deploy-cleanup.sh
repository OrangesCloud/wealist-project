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

# 기존 board-service만 중지
$COMPOSE_CMD -f "${COMPOSE_FILE}" stop board-service || true

# 컨테이너 삭제
$COMPOSE_CMD -f "${COMPOSE_FILE}" rm -f board-service || true

echo "✅ Cleanup complete."
set -e
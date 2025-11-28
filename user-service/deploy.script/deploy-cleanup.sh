#!/bin/bash
# CodeDeploy Hook: BeforeInstall
# 이전 배포에서 남은 컨테이너 중지 및 이미지 정리

set +e # 오류가 나도 계속 진행 (컨테이너가 없을 수 있음)

# [수정 필요] 경로를 /home/ec2-user/wealist로 변경
PROJECT_ROOT="/home/ec2-user/wealist"
COMPOSE_FILE="${PROJECT_ROOT}/docker/compose/docker-compose.ec2-prod.yml"

echo "🧹 Cleaning up old user-service containers..."

# CodeDeploy 파일 복사 충돌 방지: 기존 docker-compose 파일 삭제
echo "🗑️ Removing existing docker-compose file to prevent CodeDeploy conflict..."
if [ -f "${COMPOSE_FILE}" ]; then
    sudo rm -f "${COMPOSE_FILE}"
    echo "  ✅ Removed existing docker-compose.ec2-prod.yml"
else
    echo "  No existing docker-compose file found"
fi

if docker compose version &> /dev/null; then
  COMPOSE_CMD="docker compose"
else
  COMPOSE_CMD="docker-compose"
fi

# 기존 User Service 컨테이너만 직접 중지 및 삭제 (다른 서비스에 영향 없음)
echo "🛑 Stopping user-service container only..."
if sudo docker ps -a | grep -q "wealist-user-service"; then
    sudo docker stop wealist-user-service 2>/dev/null || true
    sudo docker rm -f wealist-user-service 2>/dev/null || true
    echo "  ✅ User service container removed"
else
    echo "  No existing user-service container found"
fi

# ⚠️ docker-compose 파일 삭제 제거: 다른 서비스가 사용 중일 수 있음
# CodeDeploy가 자동으로 덮어쓰기 함


# 2. 🧹 User Service 관련 임시 헬스체크 컨테이너만 정리
echo "🧹 Cleaning up user-service temporary health check containers..."

# User Service 관련 임시 컨테이너만 정리
sudo docker rm -f temp-user-health 2>/dev/null || true

# 8080 포트(User Service)만 정리 - Board Service(8000)는 건드리지 않음
echo "  - Checking port 8080 (user-service)..."
PID=$(sudo lsof -ti:8080 2>/dev/null || true)
if [ -n "$PID" ]; then
    echo "    Found process ${PID} using port 8080, killing..."
    sudo kill -9 $PID 2>/dev/null || true
fi


echo "✅ Cleanup complete."
set -e
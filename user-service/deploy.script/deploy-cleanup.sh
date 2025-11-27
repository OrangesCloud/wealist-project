#!/bin/bash
# CodeDeploy Hook: BeforeInstall
# 이전 배포에서 남은 컨테이너 중지 및 이미지 정리

set +e # 오류가 나도 계속 진행 (컨테이너가 없을 수 있음)

# [수정 필요] 경로를 /home/ec2-user/wealist로 변경
PROJECT_ROOT="/home/ec2-user/wealist"
COMPOSE_FILE="${PROJECT_ROOT}/docker/compose/docker-compose.ec2-prod.yml"

echo "🧹 Cleaning up old user-service containers..."

if docker compose version &> /dev/null; then
  COMPOSE_CMD="docker compose"
else
  COMPOSE_CMD="docker-compose"
fi

# 기존 서비스만 중지 (전체 다운은 하지 않음)
$COMPOSE_CMD -f "${COMPOSE_FILE}" stop user-service || true
$COMPOSE_CMD -f "${COMPOSE_FILE}" rm -f user-service || true

# 기존 docker-compose 파일 삭제 (CodeDeploy 파일 복사를 위해)
echo "🧹 Removing old docker-compose file to allow new deployment..."
rm -f "${COMPOSE_FILE}" || true


# 2. 🧹 임시 헬스체크 컨테이너 및 포트 충돌 방지 (추가된 부분)
echo "🧹 Cleaning up temporary health check containers and processes using ports 8080, 8000..."

# 컨테이너 이름으로 정리
CONTAINERS_TO_CLEANUP="temp-user-health temp-board-health"

for CONTAINER in ${CONTAINERS_TO_CLEANUP}; do
    echo "  - Attempting to remove ${CONTAINER}..."
    sudo docker rm -f "${CONTAINER}" 2>/dev/null || true 
done

# 8080, 8000 포트를 사용하는 프로세스 정리
for PORT in 8080 8000; do
    echo "  - Checking port ${PORT}..."
    PID=$(sudo lsof -ti:${PORT} 2>/dev/null || true)
    if [ -n "$PID" ]; then
        echo "    Found process ${PID} using port ${PORT}, killing..."
        sudo kill -9 $PID 2>/dev/null || true
    fi
done

# 8080, 8000 포트를 사용하는 Docker 컨테이너 정리
for PORT in 8080 8000; do
    CONTAINER_ID=$(sudo docker ps -q --filter "publish=${PORT}" 2>/dev/null || true)
    if [ -n "$CONTAINER_ID" ]; then
        echo "  - Found container ${CONTAINER_ID} using port ${PORT}, removing..."
        sudo docker rm -f ${CONTAINER_ID} 2>/dev/null || true
    fi
done


echo "✅ Cleanup complete."
set -e
#!/bin/bash
# CodeDeploy Hook: ValidateService
# 새로 배포된 서비스의 헬스 체크
set -euo pipefail

SERVICE_PORT=8080
HEALTH_ENDPOINT="/api/users/actuator/health"
MAX_ATTEMPTS=15
WAIT_SECONDS=5

echo "🏥 Starting health check for user-service on port ${SERVICE_PORT}..."

for i in $(seq 1 $MAX_ATTEMPTS); do
    STATUS_CODE=$(curl -s -o /dev/null -w "%{http_code}" "http://localhost:${SERVICE_PORT}${HEALTH_ENDPOINT}" || true)
    
    if [ "$STATUS_CODE" -eq 200 ]; then
        echo "✅ Health check succeeded! (HTTP 200)"
        exit 0
    elif [ "$STATUS_CODE" -eq 000 ]; then
        echo "⏳ Attempt $i/$MAX_ATTEMPTS: Service not yet reachable. Waiting ${WAIT_SECONDS}s..."
    else
        echo "⚠️ Attempt $i/$MAX_ATTEMPTS: Service responded with HTTP ${STATUS_CODE}. Waiting ${WAIT_SECONDS}s..."
    fi
    
    sleep $WAIT_SECONDS
done

echo "❌ Health check failed after ${MAX_ATTEMPTS} attempts."
exit 1
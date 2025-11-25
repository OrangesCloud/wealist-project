# wealist-project2/board-service/scripts/deploy-health-check.sh
#!/bin/bash

# =============================================================================
# CodeDeploy ValidateService Hook Script
# =============================================================================

set -euo pipefail

echo "🏥 Performing service health check..."

MAX_RETRIES=12
HEALTH_ENDPOINT="http://localhost:8000/health" 
echo "  📍 Health endpoint: ${HEALTH_ENDPOINT}"

for i in $(seq 1 $MAX_RETRIES); do
  # curl -f: 4xx, 5xx 에러 발생 시 즉시 실패 (스크립트 종료 코드 1)
  if curl -f -s ${HEALTH_ENDPOINT} > /dev/null; then
    echo "  ✅ Health check passed on attempt $i!"
    echo "✅ Board Service is healthy!"
    exit 0
  fi
  echo "  ⏳ Waiting for service... ($i/$MAX_RETRIES)"
  sleep 5
done

echo "  ❌ Health check failed after ${MAX_RETRIES} attempts!" >&2
echo "  📋 Container logs:" >&2
docker logs wealist-board-service --tail=50 2>&1 >&2
exit 1
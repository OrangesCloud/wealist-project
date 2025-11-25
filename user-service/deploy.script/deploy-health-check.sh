# wealist-project2/user-service/deploy.script/deploy-health-check.sh
#!/bin/bash

# =============================================================================
# CodeDeploy ValidateService Hook Script (User Service)
# =============================================================================

set -euo pipefail

echo "🏥 Performing user-service health check..."

MAX_RETRIES=18
HEALTH_ENDPOINT="http://localhost:8080/actuator/health" 
echo "  📍 Health endpoint: ${HEALTH_ENDPOINT}"

for i in $(seq 1 $MAX_RETRIES); do
  if curl -f -s ${HEALTH_ENDPOINT} > /dev/null; then
    echo "  ✅ Health check passed on attempt $i!"
    echo "✅ User Service is healthy!"
    exit 0
  fi
  echo "  ⏳ Waiting for user-service... ($i/$MAX_RETRIES)"
  sleep 5
done

echo "  ❌ Health check failed after ${MAX_RETRIES} attempts!" >&2
echo "  📋 Container logs:" >&2
docker logs wealist-user-service --tail=50 2>&1 >&2
exit 1
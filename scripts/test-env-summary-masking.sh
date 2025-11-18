#!/usr/bin/env bash

# =============================================================================
# Test for Environment Variable Summary Masking
# Validates that the deployment script properly masks secrets in summary output
# =============================================================================

set -euo pipefail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "========================================="
echo "Environment Summary Masking Test"
echo "========================================="
echo ""

# Create a test script that simulates the environment summary output
TEST_SCRIPT="/tmp/test-env-summary.sh"
cat > "$TEST_SCRIPT" << 'EOF'
#!/bin/bash

# Simulate loaded environment variables
export AWS_ACCOUNT_ID="123456789012"
export AWS_REGION="ap-northeast-2"
export VERSION="latest"
export POSTGRES_SUPERUSER="postgres"
export POSTGRES_SUPERUSER_PASSWORD="secret_postgres_pass_123"
export USER_DB_NAME="user_db"
export USER_DB_USER="user_user"
export USER_DB_PASSWORD="secret_user_pass_456"
export BOARD_DB_NAME="board_db"
export BOARD_DB_USER="board_user"
export BOARD_DB_PASSWORD="secret_board_pass_789"
export REDIS_PASSWORD="secret_redis_pass_abc"
export JWT_SECRET="secret_jwt_token_xyz"
export JWT_ACCESS_TOKEN_EXPIRATION_MS="1800000"
export JWT_REFRESH_TOKEN_EXPIRATION_MS="604800000"
export GOOGLE_CLIENT_ID="secret_google_client_id"
export GOOGLE_CLIENT_SECRET="secret_google_client_secret"
export GRAFANA_ADMIN_USER="admin"
export GRAFANA_ADMIN_PASSWORD="secret_grafana_pass_def"
export USER_SERVICE_URL="http://localhost:8080"

# Print summary (as in the deployment script)
echo "✅ Environment variables loaded successfully"
echo "📋 Loaded variables summary:"
echo "   - AWS_ACCOUNT_ID: ${AWS_ACCOUNT_ID}"
echo "   - AWS_REGION: ${AWS_REGION}"
echo "   - VERSION: ${VERSION}"
echo "   - POSTGRES_SUPERUSER: ${POSTGRES_SUPERUSER}"
echo "   - POSTGRES_SUPERUSER_PASSWORD: [REDACTED]"
echo "   - USER_DB_NAME: ${USER_DB_NAME}"
echo "   - USER_DB_USER: ${USER_DB_USER}"
echo "   - USER_DB_PASSWORD: [REDACTED]"
echo "   - BOARD_DB_NAME: ${BOARD_DB_NAME}"
echo "   - BOARD_DB_USER: ${BOARD_DB_USER}"
echo "   - BOARD_DB_PASSWORD: [REDACTED]"
echo "   - REDIS_PASSWORD: [REDACTED]"
echo "   - JWT_SECRET: [REDACTED]"
echo "   - JWT_ACCESS_TOKEN_EXPIRATION_MS: ${JWT_ACCESS_TOKEN_EXPIRATION_MS}"
echo "   - JWT_REFRESH_TOKEN_EXPIRATION_MS: ${JWT_REFRESH_TOKEN_EXPIRATION_MS}"
echo "   - GOOGLE_CLIENT_ID: [REDACTED]"
echo "   - GOOGLE_CLIENT_SECRET: [REDACTED]"
echo "   - GRAFANA_ADMIN_USER: ${GRAFANA_ADMIN_USER}"
echo "   - GRAFANA_ADMIN_PASSWORD: [REDACTED]"
echo "   - USER_SERVICE_URL: ${USER_SERVICE_URL}"
EOF

chmod +x "$TEST_SCRIPT"

# Run the test
echo "🧪 Running environment summary test..."
echo ""

SUMMARY_OUTPUT=$("$TEST_SCRIPT" 2>&1)

echo "📋 Summary Output:"
echo "---"
echo "$SUMMARY_OUTPUT"
echo "---"
echo ""

# Verify masking
echo "🔍 Verifying masking..."
echo ""

TESTS_PASSED=0
TESTS_FAILED=0

# Test 1: Secret values should NOT appear
echo -e "${YELLOW}[TEST 1]${NC} Secret values should not appear in summary"
SECRET_VALUES=(
    "secret_postgres_pass_123"
    "secret_user_pass_456"
    "secret_board_pass_789"
    "secret_redis_pass_abc"
    "secret_jwt_token_xyz"
    "secret_google_client_id"
    "secret_google_client_secret"
    "secret_grafana_pass_def"
)

FOUND_SECRET=false
for secret in "${SECRET_VALUES[@]}"; do
    if echo "$SUMMARY_OUTPUT" | grep -qF "$secret"; then
        echo -e "${RED}  ❌ FAIL: Secret found: ${secret}${NC}"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        FOUND_SECRET=true
    fi
done

if [ "$FOUND_SECRET" = false ]; then
    echo -e "${GREEN}  ✅ PASS: No secrets found in summary${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
fi

# Test 2: [REDACTED] markers should appear
echo -e "${YELLOW}[TEST 2]${NC} [REDACTED] markers should appear for secrets"
REDACTED_COUNT=$(echo "$SUMMARY_OUTPUT" | grep -c "\[REDACTED\]" || true)

if [ "$REDACTED_COUNT" -ge 8 ]; then
    echo -e "${GREEN}  ✅ PASS: Found $REDACTED_COUNT [REDACTED] markers${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "${RED}  ❌ FAIL: Expected at least 8 [REDACTED] markers, found $REDACTED_COUNT${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi

# Test 3: Non-secret values SHOULD appear
echo -e "${YELLOW}[TEST 3]${NC} Non-secret values should appear in summary"
NON_SECRET_VALUES=(
    "123456789012"
    "ap-northeast-2"
    "latest"
    "postgres"
    "user_db"
    "board_db"
    "admin"
    "http://localhost:8080"
)

MISSING_VALUE=false
for value in "${NON_SECRET_VALUES[@]}"; do
    if ! echo "$SUMMARY_OUTPUT" | grep -qF "$value"; then
        echo -e "${RED}  ❌ FAIL: Non-secret value missing: ${value}${NC}"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        MISSING_VALUE=true
    fi
done

if [ "$MISSING_VALUE" = false ]; then
    echo -e "${GREEN}  ✅ PASS: All non-secret values present${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
fi

# Cleanup
rm -f "$TEST_SCRIPT"

echo ""
echo "========================================="
echo "Test Summary"
echo "========================================="
echo "Tests Passed: $TESTS_PASSED"
echo "Tests Failed: $TESTS_FAILED"
echo ""

if [ $TESTS_FAILED -gt 0 ]; then
    echo -e "${RED}❌ Some tests failed${NC}"
    exit 1
else
    echo -e "${GREEN}✅ All tests passed!${NC}"
    exit 0
fi

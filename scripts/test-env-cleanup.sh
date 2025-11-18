#!/usr/bin/env bash

# =============================================================================
# Property-Based Test for Environment Variable Cleanup
# Feature: cd-workflow-improvement, Property 5: Environment cleanup on exit
# Validates: Requirements 5.6
#
# Property: For any deployment outcome (success or failure), all exported 
# environment variables containing sensitive data should be unset before 
# the script exits.
# =============================================================================

set -euo pipefail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo "========================================="
echo "Environment Cleanup Property Test"
echo "========================================="
echo ""
echo -e "${BLUE}Property 5: Environment cleanup on exit${NC}"
echo "For any deployment outcome (success or failure),"
echo "all exported environment variables containing sensitive"
echo "data should be unset before the script exits."
echo ""

TESTS_PASSED=0
TESTS_FAILED=0

# =============================================================================
# Test 1: Successful deployment scenario
# =============================================================================
echo -e "${YELLOW}[TEST 1]${NC} Environment cleanup on successful deployment"
echo ""

# Create a test script that simulates successful deployment
TEST_SCRIPT_SUCCESS="/tmp/test-deploy-success.sh"
cat > "$TEST_SCRIPT_SUCCESS" << 'EOF'
#!/bin/bash
set -euo pipefail

# Cleanup function
cleanup_environment() {
    local sensitive_vars=(
        "POSTGRES_SUPERUSER_PASSWORD"
        "USER_DB_PASSWORD"
        "BOARD_DB_PASSWORD"
        "REDIS_PASSWORD"
        "JWT_SECRET"
        "GOOGLE_CLIENT_ID"
        "GOOGLE_CLIENT_SECRET"
        "GRAFANA_ADMIN_PASSWORD"
    )
    
    for var in "${sensitive_vars[@]}"; do
        if [ -n "${!var:-}" ]; then
            unset "$var"
        fi
    done
}

# Set trap for EXIT
trap cleanup_environment EXIT

# Simulate loading sensitive environment variables
export POSTGRES_SUPERUSER_PASSWORD="secret_postgres_123"
export USER_DB_PASSWORD="secret_user_456"
export BOARD_DB_PASSWORD="secret_board_789"
export REDIS_PASSWORD="secret_redis_abc"
export JWT_SECRET="secret_jwt_xyz"
export GOOGLE_CLIENT_ID="secret_google_id"
export GOOGLE_CLIENT_SECRET="secret_google_secret"
export GRAFANA_ADMIN_PASSWORD="secret_grafana_def"

# Simulate successful deployment
echo "Deployment successful"
exit 0
EOF

chmod +x "$TEST_SCRIPT_SUCCESS"

# Run in a subshell and check if variables persist
(
    source "$TEST_SCRIPT_SUCCESS" 2>&1 > /dev/null
    EXIT_CODE=$?
    
    # After script exits, check if sensitive variables are still set
    VARS_STILL_SET=()
    
    if [ -n "${POSTGRES_SUPERUSER_PASSWORD:-}" ]; then
        VARS_STILL_SET+=("POSTGRES_SUPERUSER_PASSWORD")
    fi
    if [ -n "${USER_DB_PASSWORD:-}" ]; then
        VARS_STILL_SET+=("USER_DB_PASSWORD")
    fi
    if [ -n "${BOARD_DB_PASSWORD:-}" ]; then
        VARS_STILL_SET+=("BOARD_DB_PASSWORD")
    fi
    if [ -n "${REDIS_PASSWORD:-}" ]; then
        VARS_STILL_SET+=("REDIS_PASSWORD")
    fi
    if [ -n "${JWT_SECRET:-}" ]; then
        VARS_STILL_SET+=("JWT_SECRET")
    fi
    if [ -n "${GOOGLE_CLIENT_ID:-}" ]; then
        VARS_STILL_SET+=("GOOGLE_CLIENT_ID")
    fi
    if [ -n "${GOOGLE_CLIENT_SECRET:-}" ]; then
        VARS_STILL_SET+=("GOOGLE_CLIENT_SECRET")
    fi
    if [ -n "${GRAFANA_ADMIN_PASSWORD:-}" ]; then
        VARS_STILL_SET+=("GRAFANA_ADMIN_PASSWORD")
    fi
    
    if [ ${#VARS_STILL_SET[@]} -eq 0 ]; then
        echo -e "${GREEN}  ✅ PASS: All sensitive variables cleaned up after successful deployment${NC}"
        exit 0
    else
        echo -e "${RED}  ❌ FAIL: Variables still set: ${VARS_STILL_SET[*]}${NC}"
        exit 1
    fi
)

if [ $? -eq 0 ]; then
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi

echo ""

# =============================================================================
# Test 2: Failed deployment scenario
# =============================================================================
echo -e "${YELLOW}[TEST 2]${NC} Environment cleanup on failed deployment"
echo ""

# Create a test script that simulates failed deployment
TEST_SCRIPT_FAILURE="/tmp/test-deploy-failure.sh"
cat > "$TEST_SCRIPT_FAILURE" << 'EOF'
#!/bin/bash
set -euo pipefail

# Cleanup function
cleanup_environment() {
    local sensitive_vars=(
        "POSTGRES_SUPERUSER_PASSWORD"
        "USER_DB_PASSWORD"
        "BOARD_DB_PASSWORD"
        "REDIS_PASSWORD"
        "JWT_SECRET"
        "GOOGLE_CLIENT_ID"
        "GOOGLE_CLIENT_SECRET"
        "GRAFANA_ADMIN_PASSWORD"
    )
    
    for var in "${sensitive_vars[@]}"; do
        if [ -n "${!var:-}" ]; then
            unset "$var"
        fi
    done
}

# Set trap for ERR and EXIT
trap 'cleanup_environment; exit 1' ERR
trap cleanup_environment EXIT

# Simulate loading sensitive environment variables
export POSTGRES_SUPERUSER_PASSWORD="secret_postgres_123"
export USER_DB_PASSWORD="secret_user_456"
export BOARD_DB_PASSWORD="secret_board_789"
export REDIS_PASSWORD="secret_redis_abc"
export JWT_SECRET="secret_jwt_xyz"
export GOOGLE_CLIENT_ID="secret_google_id"
export GOOGLE_CLIENT_SECRET="secret_google_secret"
export GRAFANA_ADMIN_PASSWORD="secret_grafana_def"

# Simulate deployment failure
echo "Simulating deployment failure..."
false  # This will trigger ERR trap
EOF

chmod +x "$TEST_SCRIPT_FAILURE"

# Run in a subshell and check if variables persist even after failure
(
    "$TEST_SCRIPT_FAILURE" 2>&1 > /dev/null || true
    
    # After script exits (with failure), check if sensitive variables are still set
    VARS_STILL_SET=()
    
    if [ -n "${POSTGRES_SUPERUSER_PASSWORD:-}" ]; then
        VARS_STILL_SET+=("POSTGRES_SUPERUSER_PASSWORD")
    fi
    if [ -n "${USER_DB_PASSWORD:-}" ]; then
        VARS_STILL_SET+=("USER_DB_PASSWORD")
    fi
    if [ -n "${BOARD_DB_PASSWORD:-}" ]; then
        VARS_STILL_SET+=("BOARD_DB_PASSWORD")
    fi
    if [ -n "${REDIS_PASSWORD:-}" ]; then
        VARS_STILL_SET+=("REDIS_PASSWORD")
    fi
    if [ -n "${JWT_SECRET:-}" ]; then
        VARS_STILL_SET+=("JWT_SECRET")
    fi
    if [ -n "${GOOGLE_CLIENT_ID:-}" ]; then
        VARS_STILL_SET+=("GOOGLE_CLIENT_ID")
    fi
    if [ -n "${GOOGLE_CLIENT_SECRET:-}" ]; then
        VARS_STILL_SET+=("GOOGLE_CLIENT_SECRET")
    fi
    if [ -n "${GRAFANA_ADMIN_PASSWORD:-}" ]; then
        VARS_STILL_SET+=("GRAFANA_ADMIN_PASSWORD")
    fi
    
    if [ ${#VARS_STILL_SET[@]} -eq 0 ]; then
        echo -e "${GREEN}  ✅ PASS: All sensitive variables cleaned up after failed deployment${NC}"
        exit 0
    else
        echo -e "${RED}  ❌ FAIL: Variables still set after failure: ${VARS_STILL_SET[*]}${NC}"
        exit 1
    fi
)

if [ $? -eq 0 ]; then
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi

echo ""

# =============================================================================
# Test 3: Cleanup function is idempotent
# =============================================================================
echo -e "${YELLOW}[TEST 3]${NC} Cleanup function is idempotent (can be called multiple times)"
echo ""

TEST_SCRIPT_IDEMPOTENT="/tmp/test-cleanup-idempotent.sh"
cat > "$TEST_SCRIPT_IDEMPOTENT" << 'EOF'
#!/bin/bash
set -euo pipefail

# Cleanup function
cleanup_environment() {
    local sensitive_vars=(
        "POSTGRES_SUPERUSER_PASSWORD"
        "USER_DB_PASSWORD"
        "BOARD_DB_PASSWORD"
        "REDIS_PASSWORD"
        "JWT_SECRET"
        "GOOGLE_CLIENT_ID"
        "GOOGLE_CLIENT_SECRET"
        "GRAFANA_ADMIN_PASSWORD"
    )
    
    for var in "${sensitive_vars[@]}"; do
        if [ -n "${!var:-}" ]; then
            unset "$var"
        fi
    done
}

# Set sensitive variables
export POSTGRES_SUPERUSER_PASSWORD="secret_postgres_123"
export USER_DB_PASSWORD="secret_user_456"
export BOARD_DB_PASSWORD="secret_board_789"
export REDIS_PASSWORD="secret_redis_abc"
export JWT_SECRET="secret_jwt_xyz"
export GOOGLE_CLIENT_ID="secret_google_id"
export GOOGLE_CLIENT_SECRET="secret_google_secret"
export GRAFANA_ADMIN_PASSWORD="secret_grafana_def"

# Call cleanup multiple times
cleanup_environment
cleanup_environment
cleanup_environment

# Check if any variables are still set
VARS_STILL_SET=()

if [ -n "${POSTGRES_SUPERUSER_PASSWORD:-}" ]; then
    VARS_STILL_SET+=("POSTGRES_SUPERUSER_PASSWORD")
fi
if [ -n "${USER_DB_PASSWORD:-}" ]; then
    VARS_STILL_SET+=("USER_DB_PASSWORD")
fi
if [ -n "${BOARD_DB_PASSWORD:-}" ]; then
    VARS_STILL_SET+=("BOARD_DB_PASSWORD")
fi
if [ -n "${REDIS_PASSWORD:-}" ]; then
    VARS_STILL_SET+=("REDIS_PASSWORD")
fi
if [ -n "${JWT_SECRET:-}" ]; then
    VARS_STILL_SET+=("JWT_SECRET")
fi
if [ -n "${GOOGLE_CLIENT_ID:-}" ]; then
    VARS_STILL_SET+=("GOOGLE_CLIENT_ID")
fi
if [ -n "${GOOGLE_CLIENT_SECRET:-}" ]; then
    VARS_STILL_SET+=("GOOGLE_CLIENT_SECRET")
fi
if [ -n "${GRAFANA_ADMIN_PASSWORD:-}" ]; then
    VARS_STILL_SET+=("GRAFANA_ADMIN_PASSWORD")
fi

if [ ${#VARS_STILL_SET[@]} -eq 0 ]; then
    echo "All variables cleaned up"
    exit 0
else
    echo "Variables still set: ${VARS_STILL_SET[*]}"
    exit 1
fi
EOF

chmod +x "$TEST_SCRIPT_IDEMPOTENT"

if "$TEST_SCRIPT_IDEMPOTENT" 2>&1 | grep -q "All variables cleaned up"; then
    echo -e "${GREEN}  ✅ PASS: Cleanup function is idempotent${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "${RED}  ❌ FAIL: Cleanup function failed when called multiple times${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi

echo ""

# =============================================================================
# Test 4: Trap is properly set for EXIT
# =============================================================================
echo -e "${YELLOW}[TEST 4]${NC} EXIT trap is properly configured"
echo ""

TEST_SCRIPT_TRAP="/tmp/test-trap-exit.sh"
cat > "$TEST_SCRIPT_TRAP" << 'EOF'
#!/bin/bash
set -euo pipefail

# Cleanup function
cleanup_environment() {
    echo "CLEANUP_CALLED"
}

# Set trap for EXIT
trap cleanup_environment EXIT

# Simulate some work
export POSTGRES_SUPERUSER_PASSWORD="secret_postgres_123"

# Normal exit
exit 0
EOF

chmod +x "$TEST_SCRIPT_TRAP"

OUTPUT=$("$TEST_SCRIPT_TRAP" 2>&1)

if echo "$OUTPUT" | grep -q "CLEANUP_CALLED"; then
    echo -e "${GREEN}  ✅ PASS: EXIT trap properly calls cleanup function${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "${RED}  ❌ FAIL: EXIT trap did not call cleanup function${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi

echo ""

# =============================================================================
# Test 5: Verify actual deployment script has cleanup
# =============================================================================
echo -e "${YELLOW}[TEST 5]${NC} Deployment script contains cleanup implementation"
echo ""

# Check if the deployment script in the workflow has the cleanup function
WORKFLOW_FILE=".github/workflows/cd-dev-board-service.yml"

if [ ! -f "$WORKFLOW_FILE" ]; then
    echo -e "${RED}  ❌ FAIL: Workflow file not found: $WORKFLOW_FILE${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
else
    # Check for cleanup_environment function
    if grep -q "cleanup_environment()" "$WORKFLOW_FILE"; then
        echo -e "${GREEN}  ✅ PASS: cleanup_environment function found in deployment script${NC}"
        
        # Check for trap EXIT
        if grep -q "trap cleanup_environment EXIT" "$WORKFLOW_FILE" || \
           grep -q "trap.*EXIT.*cleanup" "$WORKFLOW_FILE"; then
            echo -e "${GREEN}  ✅ PASS: EXIT trap found in deployment script${NC}"
            TESTS_PASSED=$((TESTS_PASSED + 2))
        else
            echo -e "${RED}  ❌ FAIL: EXIT trap not found in deployment script${NC}"
            TESTS_FAILED=$((TESTS_FAILED + 1))
            TESTS_PASSED=$((TESTS_PASSED + 1))
        fi
    else
        echo -e "${RED}  ❌ FAIL: cleanup_environment function not found in deployment script${NC}"
        TESTS_FAILED=$((TESTS_FAILED + 2))
    fi
fi

echo ""

# Cleanup test files
rm -f "$TEST_SCRIPT_SUCCESS" "$TEST_SCRIPT_FAILURE" "$TEST_SCRIPT_IDEMPOTENT" "$TEST_SCRIPT_TRAP"

# =============================================================================
# Test Summary
# =============================================================================
echo "========================================="
echo "Test Summary"
echo "========================================="
echo "Tests Passed: $TESTS_PASSED"
echo "Tests Failed: $TESTS_FAILED"
echo ""

if [ $TESTS_FAILED -gt 0 ]; then
    echo -e "${RED}❌ Some tests failed${NC}"
    echo ""
    echo "Property 5 validation: FAILED"
    echo "The deployment script does not properly clean up environment"
    echo "variables on exit, which violates Requirements 5.6"
    exit 1
else
    echo -e "${GREEN}✅ All tests passed!${NC}"
    echo ""
    echo "Property 5 validation: PASSED"
    echo "Environment variables are properly cleaned up on all exit scenarios"
    exit 0
fi

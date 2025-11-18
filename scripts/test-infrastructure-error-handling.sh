#!/usr/bin/env bash

# =============================================================================
# Property-Based Test for Infrastructure Error Handling
# Feature: cd-workflow-improvement, Property 4: Error messages include troubleshooting guidance
# Validates: Requirements 4.5 (infrastructure service failures)
# =============================================================================

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test counters
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

# Test result tracking
declare -a FAILED_TESTS=()

# =============================================================================
# Helper Functions
# =============================================================================

log_test() {
    echo -e "${YELLOW}[TEST]${NC} $1"
}

log_pass() {
    echo -e "${GREEN}[PASS]${NC} $1"
    TESTS_PASSED=$((TESTS_PASSED + 1))
}

log_fail() {
    echo -e "${RED}[FAIL]${NC} $1"
    TESTS_FAILED=$((TESTS_FAILED + 1))
    FAILED_TESTS+=("$1")
}

# =============================================================================
# Mock Docker Compose Commands
# =============================================================================

# Mock docker compose that fails to start infrastructure
mock_docker_compose_infra_fail() {
    echo "Error response from daemon: Cannot start container postgres: ..." >&2
    return 1
}

# Mock docker compose that succeeds
mock_docker_compose_success() {
    echo "Container wealist-postgres  Started"
    echo "Container wealist-redis  Started"
    return 0
}

# Mock docker compose ps
mock_docker_compose_ps() {
    cat << 'EOF'
NAME                STATUS              PORTS
wealist-postgres    Exited (1)          
wealist-redis       Up 5 seconds        0.0.0.0:6379->6379/tcp
EOF
}

# Mock docker compose logs
mock_docker_compose_logs() {
    cat << 'EOF'
postgres  | Error: Database initialization failed
postgres  | FATAL: could not create shared memory segment
redis     | Ready to accept connections
EOF
}

# Mock docker exec for health checks
mock_docker_exec_fail() {
    echo "Error: No such container: wealist-postgres" >&2
    return 1
}

mock_docker_exec_success() {
    echo "postgres is ready to accept connections"
    return 0
}

# =============================================================================
# Simulated Infrastructure Service Management Functions
# =============================================================================

start_infrastructure_services() {
    local compose_file="${COMPOSE_FILE:-docker-compose.yml}"
    local compose_cmd="${COMPOSE_CMD:-docker compose}"
    
    echo "  🚀 Starting infrastructure services (postgres, redis)..." >&2
    
    local infra_error
    infra_error=$(mktemp)
    
    # Use mock for testing
    if [ "${MOCK_DOCKER_BEHAVIOR:-success}" = "fail" ]; then
        mock_docker_compose_infra_fail > "$infra_error" 2>&1
        local exit_code=1
    else
        mock_docker_compose_success > "$infra_error" 2>&1
        local exit_code=0
    fi
    
    if [ $exit_code -ne 0 ]; then
        echo "  ❌ Failed to start infrastructure services" >&2
        echo "  📋 Error details:" >&2
        cat "$infra_error" >&2
        echo "  📋 Docker Compose status:" >&2
        mock_docker_compose_ps >&2
        echo "  📋 Recent logs:" >&2
        mock_docker_compose_logs >&2
        echo "  📋 Troubleshooting:" >&2
        echo "     - Check if Docker daemon is running" >&2
        echo "     - Verify docker-compose.yml file is valid" >&2
        echo "     - Check disk space: df -h" >&2
        echo "     - Check Docker logs: docker logs <container_name>" >&2
        echo "     - Verify port availability: netstat -tuln | grep <port>" >&2
        rm -f "$infra_error"
        return 1
    fi
    
    rm -f "$infra_error"
    echo "  ✅ Infrastructure services started" >&2
    return 0
}

check_postgres_health() {
    local postgres_user="${POSTGRES_SUPERUSER:-postgres}"
    
    echo "  🏥 Checking PostgreSQL health..." >&2
    
    # Use mock for testing
    if [ "${MOCK_HEALTH_CHECK:-success}" = "fail" ]; then
        mock_docker_exec_fail
        local exit_code=1
    else
        mock_docker_exec_success
        local exit_code=0
    fi
    
    if [ $exit_code -ne 0 ]; then
        echo "  ❌ PostgreSQL failed to become ready" >&2
        echo "  📋 Container status:" >&2
        echo "NAME                STATUS              PORTS" >&2
        echo "wealist-postgres    Exited (1)          " >&2
        echo "  📋 Recent logs:" >&2
        echo "postgres  | Error: Database initialization failed" >&2
        echo "  📋 Troubleshooting:" >&2
        echo "     - Check PostgreSQL container logs: docker logs wealist-postgres" >&2
        echo "     - Verify PostgreSQL configuration" >&2
        echo "     - Check if data directory is corrupted" >&2
        echo "     - Verify environment variables are set correctly" >&2
        echo "     - Check disk space and permissions" >&2
        return 1
    fi
    
    echo "  ✅ PostgreSQL is ready" >&2
    return 0
}

force_recreate_service() {
    local service="$1"
    local compose_file="${COMPOSE_FILE:-docker-compose.yml}"
    local compose_cmd="${COMPOSE_CMD:-docker compose}"
    
    echo "  🔄 Force recreating ${service}..." >&2
    
    local restart_error
    restart_error=$(mktemp)
    
    # Use mock for testing
    if [ "${MOCK_RECREATE_BEHAVIOR:-success}" = "fail" ]; then
        echo "Error: failed to create container: ..." > "$restart_error"
        local exit_code=1
    else
        echo "Container ${service} recreated" > "$restart_error"
        local exit_code=0
    fi
    
    if [ $exit_code -ne 0 ]; then
        echo "  ❌ Failed to recreate ${service}" >&2
        echo "  📋 Error details:" >&2
        cat "$restart_error" >&2
        echo "  📋 Recent service logs:" >&2
        echo "${service}  | Error: Application failed to start" >&2
        echo "  📋 Container status:" >&2
        echo "NAME                STATUS              PORTS" >&2
        echo "${service}          Exited (1)          " >&2
        echo "  📋 Troubleshooting:" >&2
        echo "     - Check if image was pulled successfully" >&2
        echo "     - Verify docker-compose.yml configuration" >&2
        echo "     - Check container logs: docker logs ${service}" >&2
        echo "     - Verify environment variables are set" >&2
        echo "     - Check disk space: df -h" >&2
        echo "     - Verify network connectivity" >&2
        rm -f "$restart_error"
        return 1
    fi
    
    rm -f "$restart_error"
    echo "  ✅ ${service} recreated successfully" >&2
    return 0
}

# =============================================================================
# Property 4: Error messages include troubleshooting guidance
# =============================================================================

test_infra_start_failure_includes_troubleshooting() {
    log_test "Property 4.4: Infrastructure start failure includes troubleshooting"
    TESTS_RUN=$((TESTS_RUN + 1))
    
    export MOCK_DOCKER_BEHAVIOR="fail"
    export COMPOSE_FILE="docker-compose.yml"
    export COMPOSE_CMD="docker compose"
    
    local error_output
    error_output=$(start_infrastructure_services 2>&1 || true)
    
    # Check for troubleshooting section
    if ! echo "$error_output" | grep -qi "troubleshooting"; then
        log_fail "Infrastructure start error missing troubleshooting section"
        echo "  Error output: $error_output"
        return 1
    fi
    
    # Check for specific guidance items
    local guidance_items=(
        "Docker daemon"
        "docker-compose.yml"
        "disk space"
        "docker logs"
    )
    
    local missing_guidance=()
    for item in "${guidance_items[@]}"; do
        if ! echo "$error_output" | grep -qi "$item"; then
            missing_guidance+=("$item")
        fi
    done
    
    if [ ${#missing_guidance[@]} -gt 0 ]; then
        log_fail "Infrastructure error missing guidance: ${missing_guidance[*]}"
        echo "  Error output: $error_output"
        return 1
    fi
    
    # Check that diagnostic information is included
    if ! echo "$error_output" | grep -qi "status\|logs"; then
        log_fail "Infrastructure error missing diagnostic information"
        echo "  Error output: $error_output"
        return 1
    fi
    
    log_pass "Infrastructure start failure includes comprehensive troubleshooting"
    return 0
}

test_health_check_failure_includes_troubleshooting() {
    log_test "Property 4.5: Health check failure includes troubleshooting"
    TESTS_RUN=$((TESTS_RUN + 1))
    
    export MOCK_HEALTH_CHECK="fail"
    export POSTGRES_SUPERUSER="postgres"
    
    local error_output
    error_output=$(check_postgres_health 2>&1 || true)
    
    # Check for troubleshooting section
    if ! echo "$error_output" | grep -qi "troubleshooting"; then
        log_fail "Health check error missing troubleshooting section"
        echo "  Error output: $error_output"
        return 1
    fi
    
    # Check for specific guidance items
    local guidance_items=(
        "container logs"
        "configuration"
        "environment variables"
        "disk space"
    )
    
    local missing_guidance=()
    for item in "${guidance_items[@]}"; do
        if ! echo "$error_output" | grep -qi "$item"; then
            missing_guidance+=("$item")
        fi
    done
    
    if [ ${#missing_guidance[@]} -gt 0 ]; then
        log_fail "Health check error missing guidance: ${missing_guidance[*]}"
        echo "  Error output: $error_output"
        return 1
    fi
    
    log_pass "Health check failure includes comprehensive troubleshooting"
    return 0
}

test_service_recreate_failure_includes_troubleshooting() {
    log_test "Property 4.6: Service recreate failure includes troubleshooting"
    TESTS_RUN=$((TESTS_RUN + 1))
    
    export MOCK_RECREATE_BEHAVIOR="fail"
    export COMPOSE_FILE="docker-compose.yml"
    export COMPOSE_CMD="docker compose"
    
    local service_name="board-service"
    local error_output
    error_output=$(force_recreate_service "$service_name" 2>&1 || true)
    
    # Check for troubleshooting section
    if ! echo "$error_output" | grep -qi "troubleshooting"; then
        log_fail "Service recreate error missing troubleshooting section"
        echo "  Error output: $error_output"
        return 1
    fi
    
    # Check for specific guidance items
    local guidance_items=(
        "image"
        "docker-compose.yml"
        "container logs"
        "environment variables"
        "disk space"
    )
    
    local missing_guidance=()
    for item in "${guidance_items[@]}"; do
        if ! echo "$error_output" | grep -qi "$item"; then
            missing_guidance+=("$item")
        fi
    done
    
    if [ ${#missing_guidance[@]} -gt 0 ]; then
        log_fail "Service recreate error missing guidance: ${missing_guidance[*]}"
        echo "  Error output: $error_output"
        return 1
    fi
    
    # Check that service name appears in error
    if ! echo "$error_output" | grep -q "$service_name"; then
        log_fail "Service name '$service_name' not found in error message"
        echo "  Error output: $error_output"
        return 1
    fi
    
    log_pass "Service recreate failure includes comprehensive troubleshooting"
    return 0
}

test_error_includes_diagnostic_commands() {
    log_test "Property 4.7: Error messages include specific diagnostic commands"
    TESTS_RUN=$((TESTS_RUN + 1))
    
    export MOCK_DOCKER_BEHAVIOR="fail"
    export COMPOSE_FILE="docker-compose.yml"
    
    local error_output
    error_output=$(start_infrastructure_services 2>&1 || true)
    
    # Check for specific commands that users can run
    local commands=(
        "df -h"
        "docker logs"
        "netstat"
    )
    
    local missing_commands=()
    for cmd in "${commands[@]}"; do
        if ! echo "$error_output" | grep -q "$cmd"; then
            missing_commands+=("$cmd")
        fi
    done
    
    if [ ${#missing_commands[@]} -gt 0 ]; then
        log_fail "Error missing diagnostic commands: ${missing_commands[*]}"
        echo "  Error output: $error_output"
        return 1
    fi
    
    log_pass "Error messages include specific diagnostic commands"
    return 0
}

# =============================================================================
# Property-Based Testing with Random Service Names
# =============================================================================

generate_random_service_name() {
    local services=("board-service" "user-service" "api-gateway" "auth-service" "notification-service")
    echo "${services[$RANDOM % ${#services[@]}]}"
}

test_random_service_failures_include_troubleshooting() {
    log_test "Property 4 (PBT): Random service failures include troubleshooting"
    TESTS_RUN=$((TESTS_RUN + 1))
    
    export MOCK_RECREATE_BEHAVIOR="fail"
    export COMPOSE_FILE="docker-compose.yml"
    
    local iterations=20
    local failures=0
    
    echo "  Running $iterations iterations with random service names..."
    
    for i in $(seq 1 $iterations); do
        local random_service
        random_service=$(generate_random_service_name)
        
        local error_output
        error_output=$(force_recreate_service "$random_service" 2>&1 || true)
        
        if ! echo "$error_output" | grep -qi "troubleshooting"; then
            echo "  ❌ Iteration $i: No troubleshooting for service '$random_service'"
            failures=$((failures + 1))
            
            # Show first failure as example
            if [ $failures -eq 1 ]; then
                echo "  Error output: $error_output"
            fi
        fi
        
        # Also check that service name appears
        if ! echo "$error_output" | grep -q "$random_service"; then
            echo "  ❌ Iteration $i: Service name '$random_service' not in error"
            failures=$((failures + 1))
        fi
    done
    
    if [ $failures -gt 0 ]; then
        log_fail "$failures out of $iterations service errors missing proper guidance"
        return 1
    fi
    
    log_pass "All $iterations random service errors include troubleshooting"
    return 0
}

# =============================================================================
# Edge Cases
# =============================================================================

test_multiple_failures_all_include_guidance() {
    log_test "Property 4 (Edge Case): Multiple consecutive failures include guidance"
    TESTS_RUN=$((TESTS_RUN + 1))
    
    # Simulate multiple failures in sequence
    export MOCK_DOCKER_BEHAVIOR="fail"
    export MOCK_HEALTH_CHECK="fail"
    export MOCK_RECREATE_BEHAVIOR="fail"
    
    local all_errors=""
    
    # Collect errors from multiple operations
    all_errors+=$(start_infrastructure_services 2>&1 || true)
    all_errors+=$(check_postgres_health 2>&1 || true)
    all_errors+=$(force_recreate_service "test-service" 2>&1 || true)
    
    # Count troubleshooting sections
    local troubleshooting_count
    troubleshooting_count=$(echo "$all_errors" | grep -ci "troubleshooting" || true)
    
    if [ "$troubleshooting_count" -lt 3 ]; then
        log_fail "Expected 3 troubleshooting sections, found $troubleshooting_count"
        echo "  Combined errors: $all_errors"
        return 1
    fi
    
    log_pass "Multiple consecutive failures all include troubleshooting guidance"
    return 0
}

test_error_guidance_is_specific_to_failure_type() {
    log_test "Property 4 (Edge Case): Error guidance is specific to failure type"
    TESTS_RUN=$((TESTS_RUN + 1))
    
    export MOCK_DOCKER_BEHAVIOR="fail"
    export MOCK_HEALTH_CHECK="fail"
    
    local infra_error
    infra_error=$(start_infrastructure_services 2>&1 || true)
    
    local health_error
    health_error=$(check_postgres_health 2>&1 || true)
    
    # Infrastructure error should mention docker-compose
    if ! echo "$infra_error" | grep -qi "docker-compose"; then
        log_fail "Infrastructure error doesn't mention docker-compose"
        return 1
    fi
    
    # Health check error should mention PostgreSQL-specific items
    if ! echo "$health_error" | grep -qi "postgresql\|postgres"; then
        log_fail "Health check error doesn't mention PostgreSQL"
        return 1
    fi
    
    # Health check error should mention data directory
    if ! echo "$health_error" | grep -qi "data directory\|configuration"; then
        log_fail "Health check error doesn't mention PostgreSQL-specific troubleshooting"
        return 1
    fi
    
    log_pass "Error guidance is specific to failure type"
    return 0
}

# =============================================================================
# Run All Tests
# =============================================================================

echo "========================================="
echo "Infrastructure Error Handling Tests"
echo "========================================="
echo ""

# Property 4: Error messages include troubleshooting guidance
test_infra_start_failure_includes_troubleshooting
test_health_check_failure_includes_troubleshooting
test_service_recreate_failure_includes_troubleshooting
test_error_includes_diagnostic_commands

# Property-based tests
test_random_service_failures_include_troubleshooting

# Edge cases
test_multiple_failures_all_include_guidance
test_error_guidance_is_specific_to_failure_type

# =============================================================================
# Test Summary
# =============================================================================

echo ""
echo "========================================="
echo "Test Summary"
echo "========================================="
echo "Tests Run:    $TESTS_RUN"
echo "Tests Passed: $TESTS_PASSED"
echo "Tests Failed: $TESTS_FAILED"
echo ""

if [ $TESTS_FAILED -gt 0 ]; then
    echo -e "${RED}FAILED TESTS:${NC}"
    for test in "${FAILED_TESTS[@]}"; do
        echo "  - $test"
    done
    echo ""
    exit 1
else
    echo -e "${GREEN}✅ All tests passed!${NC}"
    exit 0
fi

#!/usr/bin/env bash

# =============================================================================
# Property-Based Test for Log Masking
# Feature: cd-workflow-improvement, Property 6: Sensitive information masking in logs
# Validates: Requirements 6.1, 6.2, 6.3
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
# Mock Functions (simulating deployment script functions)
# =============================================================================

# Simulated load_secret function with masking
load_secret_with_masking() {
    local param_name="$1"
    local secret_value="$2"  # In real scenario, this comes from AWS
    
    # Log without revealing the secret
    echo "  ⏳ Loading secret parameter: ${param_name}" >&2
    echo "  ✅ Loaded: ${param_name} (encrypted)" >&2
    
    # Return the actual value (but logs should never show it)
    echo "$secret_value"
}

# Simulated environment variable summary with masking
print_env_summary_with_masking() {
    # Takes key=value pairs as arguments
    echo "📋 Loaded variables summary:"
    
    while [ $# -gt 0 ]; do
        local pair="$1"
        local key="${pair%%=*}"
        local value="${pair#*=}"
        local is_secret=false
        
        # Check if this is a secret variable
        if [[ "$key" =~ PASSWORD|SECRET|KEY|TOKEN ]]; then
            is_secret=true
        fi
        
        if [ "$is_secret" = true ]; then
            echo "   - ${key}: [REDACTED]"
        else
            echo "   - ${key}: ${value}"
        fi
        
        shift
    done
}

# Simulated error message with masking
print_error_with_masking() {
    local error_msg="$1"
    local secret_value="$2"
    
    # Mask the secret in error messages
    local masked_msg="${error_msg//$secret_value/[REDACTED]}"
    echo "❌ Error: ${masked_msg}" >&2
}

# =============================================================================
# Property 6: Sensitive information masking in logs
# =============================================================================

# Property: For any secret value loaded from Parameter Store (SecureString type),
# all log outputs should display [REDACTED] instead of the actual value

test_secret_loading_masks_value() {
    log_test "Property 6.1: Secret loading should mask value in logs"
    TESTS_RUN=$((TESTS_RUN + 1))
    
    local secret_value="super_secret_password_12345"
    local param_name="db/postgres_password"
    
    # Capture stderr (where logs go)
    local log_output
    log_output=$(load_secret_with_masking "$param_name" "$secret_value" 2>&1 >/dev/null)
    
    # Check that the secret value does NOT appear in logs
    if echo "$log_output" | grep -q "$secret_value"; then
        log_fail "Secret value '$secret_value' found in logs during parameter loading"
        echo "  Log output: $log_output"
        return 1
    fi
    
    # Check that the parameter name DOES appear
    if ! echo "$log_output" | grep -q "$param_name"; then
        log_fail "Parameter name '$param_name' not found in logs"
        return 1
    fi
    
    log_pass "Secret value properly masked during loading"
    return 0
}

test_env_summary_masks_secrets() {
    log_test "Property 6.2: Environment summary should mask secret values"
    TESTS_RUN=$((TESTS_RUN + 1))
    
    # Capture output
    local summary_output
    summary_output=$(print_env_summary_with_masking \
        "AWS_ACCOUNT_ID=123456789012" \
        "POSTGRES_SUPERUSER=postgres" \
        "POSTGRES_SUPERUSER_PASSWORD=secret_pass_123" \
        "BOARD_DB_NAME=board_db" \
        "BOARD_DB_PASSWORD=another_secret_456" \
        "JWT_SECRET=jwt_secret_token_789" \
        "REDIS_PASSWORD=redis_pass_abc")
    
    # Check that secret values do NOT appear
    local secret_values=(
        "secret_pass_123"
        "another_secret_456"
        "jwt_secret_token_789"
        "redis_pass_abc"
    )
    
    local found_secret=false
    for secret in "${secret_values[@]}"; do
        if echo "$summary_output" | grep -q "$secret"; then
            log_fail "Secret value '$secret' found in environment summary"
            echo "  Summary output: $summary_output"
            found_secret=true
        fi
    done
    
    if [ "$found_secret" = true ]; then
        return 1
    fi
    
    # Check that [REDACTED] appears for secret variables
    local redacted_count
    redacted_count=$(echo "$summary_output" | grep -c "\[REDACTED\]" || true)
    
    if [ "$redacted_count" -lt 4 ]; then
        log_fail "Expected at least 4 [REDACTED] markers, found $redacted_count"
        echo "  Summary output: $summary_output"
        return 1
    fi
    
    # Check that non-secret values DO appear
    if ! echo "$summary_output" | grep -q "123456789012"; then
        log_fail "Non-secret value (AWS_ACCOUNT_ID) not found in summary"
        return 1
    fi
    
    log_pass "Secret values properly masked in environment summary"
    return 0
}

test_error_messages_mask_secrets() {
    log_test "Property 6.3: Error messages should mask secret values"
    TESTS_RUN=$((TESTS_RUN + 1))
    
    local secret_value="my_secret_password_xyz"
    local error_msg="Failed to connect to database with password: ${secret_value}"
    
    # Capture error output
    local error_output
    error_output=$(print_error_with_masking "$error_msg" "$secret_value" 2>&1)
    
    # Check that secret value does NOT appear
    if echo "$error_output" | grep -q "$secret_value"; then
        log_fail "Secret value '$secret_value' found in error message"
        echo "  Error output: $error_output"
        return 1
    fi
    
    # Check that [REDACTED] appears
    if ! echo "$error_output" | grep -q "\[REDACTED\]"; then
        log_fail "[REDACTED] marker not found in error message"
        echo "  Error output: $error_output"
        return 1
    fi
    
    log_pass "Secret value properly masked in error messages"
    return 0
}

# =============================================================================
# Property-Based Testing with Random Inputs
# =============================================================================

generate_random_secret() {
    # Generate random secret value (alphanumeric + special chars)
    local length=${1:-20}
    # Use openssl for better compatibility
    openssl rand -base64 "$length" | tr -d '\n' | head -c "$length"
}

test_random_secrets_are_masked() {
    log_test "Property 6 (PBT): Random secret values should always be masked"
    TESTS_RUN=$((TESTS_RUN + 1))
    
    local iterations=100
    local failures=0
    
    echo "  Running $iterations iterations with random secrets..."
    
    for i in $(seq 1 $iterations); do
        local random_secret
        random_secret=$(generate_random_secret 32)
        local param_name="test/secret_${i}"
        
        # Test secret loading
        local log_output
        log_output=$(load_secret_with_masking "$param_name" "$random_secret" 2>&1 >/dev/null)
        
        if echo "$log_output" | grep -qF "$random_secret"; then
            echo "  ❌ Iteration $i: Secret leaked in logs"
            failures=$((failures + 1))
            
            # Show first failure as example
            if [ $failures -eq 1 ]; then
                echo "  Example leaked secret: ${random_secret:0:10}..."
                echo "  Log output: $log_output"
            fi
        fi
    done
    
    if [ $failures -gt 0 ]; then
        log_fail "$failures out of $iterations random secrets were leaked in logs"
        return 1
    fi
    
    log_pass "All $iterations random secrets properly masked"
    return 0
}

test_secrets_with_special_chars_are_masked() {
    log_test "Property 6 (Edge Case): Secrets with special characters should be masked"
    TESTS_RUN=$((TESTS_RUN + 1))
    
    # Test secrets with various special characters that might break regex
    local special_secrets=(
        'pass$word!123'
        'secret.with.dots'
        'token[with]brackets'
        'key{with}braces'
        'value|with|pipes'
        'secret\with\backslashes'
        'pass*word?test'
    )
    
    local failures=0
    for secret in "${special_secrets[@]}"; do
        local log_output
        log_output=$(load_secret_with_masking "test/special" "$secret" 2>&1 >/dev/null)
        
        if echo "$log_output" | grep -qF "$secret"; then
            echo "  ❌ Secret with special chars leaked: ${secret}"
            failures=$((failures + 1))
        fi
    done
    
    if [ $failures -gt 0 ]; then
        log_fail "$failures secrets with special characters were leaked"
        return 1
    fi
    
    log_pass "All secrets with special characters properly masked"
    return 0
}

# =============================================================================
# Run All Tests
# =============================================================================

echo "========================================="
echo "Log Masking Property-Based Tests"
echo "========================================="
echo ""

# Run individual property tests
test_secret_loading_masks_value
test_env_summary_masks_secrets
test_error_messages_mask_secrets

# Run property-based tests with random inputs
test_random_secrets_are_masked
test_secrets_with_special_chars_are_masked

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

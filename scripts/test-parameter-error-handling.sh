#!/usr/bin/env bash

# =============================================================================
# Property-Based Test for Parameter Error Handling
# Feature: cd-workflow-improvement, Property 2: Missing parameter error handling
# Feature: cd-workflow-improvement, Property 4: Error messages include troubleshooting guidance
# Validates: Requirements 2.5, 4.5
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
# Mock AWS CLI for Testing
# =============================================================================

# Mock AWS CLI that simulates parameter not found
mock_aws_parameter_not_found() {
    local param_name="$1"
    echo "An error occurred (ParameterNotFound) when calling the GetParameter operation: Parameter ${param_name} not found." >&2
    return 1
}

# Mock AWS CLI that simulates access denied
mock_aws_access_denied() {
    local param_name="$1"
    echo "An error occurred (AccessDeniedException) when calling the GetParameter operation: User is not authorized to perform: ssm:GetParameter on resource: ${param_name}" >&2
    return 1
}

# Mock AWS CLI that simulates success
mock_aws_success() {
    echo "test-value"
    return 0
}

# =============================================================================
# Simulated load_param function (from deployment script)
# =============================================================================

load_param() {
    local param_name="$1"
    local error_output
    local value
    local full_param_path="${PARAMETER_PREFIX}/${param_name}"
    
    echo "  ⏳ Loading parameter: ${param_name}" >&2
    
    error_output=$(mktemp)
    
    # Use mock AWS CLI for testing
    if [ "${MOCK_AWS_BEHAVIOR:-success}" = "not_found" ]; then
        mock_aws_parameter_not_found "$full_param_path" > "$error_output" 2>&1
        local exit_code=1
        value=""
    elif [ "${MOCK_AWS_BEHAVIOR:-success}" = "access_denied" ]; then
        mock_aws_access_denied "$full_param_path" > "$error_output" 2>&1
        local exit_code=1
        value=""
    else
        value=$(mock_aws_success)
        local exit_code=0
    fi
    
    if [ $exit_code -ne 0 ] || [ -z "$value" ] || [ "$value" = "None" ]; then
        echo "  ❌ Failed to load parameter: ${param_name}" >&2
        echo "     Full path: ${full_param_path}" >&2
        echo "  📋 Troubleshooting:" >&2
        echo "     - Check if parameter exists in Parameter Store" >&2
        echo "     - Verify IAM role has ssm:GetParameter permission" >&2
        echo "     - Confirm AWS region is correct: ${AWS_REGION}" >&2
        rm -f "$error_output"
        return 1
    fi
    
    rm -f "$error_output"
    echo "  ✅ Loaded: ${param_name}" >&2
    echo "$value"
}

load_secret() {
    local param_name="$1"
    local error_output
    local value
    local full_param_path="${PARAMETER_PREFIX}/${param_name}"
    
    echo "  ⏳ Loading secret parameter: ${param_name}" >&2
    
    error_output=$(mktemp)
    
    # Use mock AWS CLI for testing
    if [ "${MOCK_AWS_BEHAVIOR:-success}" = "not_found" ]; then
        mock_aws_parameter_not_found "$full_param_path" > "$error_output" 2>&1
        local exit_code=1
        value=""
    elif [ "${MOCK_AWS_BEHAVIOR:-success}" = "access_denied" ]; then
        mock_aws_access_denied "$full_param_path" > "$error_output" 2>&1
        local exit_code=1
        value=""
    else
        value=$(mock_aws_success)
        local exit_code=0
    fi
    
    if [ $exit_code -ne 0 ] || [ -z "$value" ] || [ "$value" = "None" ]; then
        echo "  ❌ Failed to load secret parameter: ${param_name}" >&2
        echo "     Full path: ${full_param_path}" >&2
        echo "  📋 Troubleshooting:" >&2
        echo "     - Check if parameter exists in Parameter Store" >&2
        echo "     - Verify IAM role has ssm:GetParameter permission" >&2
        echo "     - Confirm parameter type is SecureString" >&2
        echo "     - Verify KMS key permissions for decryption" >&2
        echo "     - Confirm AWS region is correct: ${AWS_REGION}" >&2
        rm -f "$error_output"
        return 1
    fi
    
    rm -f "$error_output"
    echo "  ✅ Loaded: ${param_name} (SecureString)" >&2
    echo "$value"
}

# =============================================================================
# Property 2: Missing parameter error handling
# =============================================================================

# Property: For any required parameter name, when that parameter is missing
# or inaccessible from Parameter Store, the system should fail with an error
# message that includes the parameter name.

test_missing_parameter_includes_name() {
    log_test "Property 2.1: Missing parameter error includes parameter name"
    TESTS_RUN=$((TESTS_RUN + 1))
    
    export PARAMETER_PREFIX="/wealist/dev"
    export AWS_REGION="ap-northeast-2"
    export MOCK_AWS_BEHAVIOR="not_found"
    
    local param_name="db/postgres_password"
    local error_output
    
    # Capture stderr
    error_output=$(load_param "$param_name" 2>&1 >/dev/null || true)
    
    # Check that parameter name appears in error message
    if ! echo "$error_output" | grep -q "$param_name"; then
        log_fail "Parameter name '$param_name' not found in error message"
        echo "  Error output: $error_output"
        return 1
    fi
    
    # Check that error message indicates failure
    if ! echo "$error_output" | grep -qi "failed"; then
        log_fail "Error message does not indicate failure"
        echo "  Error output: $error_output"
        return 1
    fi
    
    log_pass "Missing parameter error includes parameter name"
    return 0
}

test_missing_secret_includes_name() {
    log_test "Property 2.2: Missing secret parameter error includes parameter name"
    TESTS_RUN=$((TESTS_RUN + 1))
    
    export PARAMETER_PREFIX="/wealist/dev"
    export AWS_REGION="ap-northeast-2"
    export MOCK_AWS_BEHAVIOR="not_found"
    
    local param_name="jwt/jwt_secret"
    local error_output
    
    # Capture stderr
    error_output=$(load_secret "$param_name" 2>&1 >/dev/null || true)
    
    # Check that parameter name appears in error message
    if ! echo "$error_output" | grep -q "$param_name"; then
        log_fail "Secret parameter name '$param_name' not found in error message"
        echo "  Error output: $error_output"
        return 1
    fi
    
    # Check that error message indicates failure
    if ! echo "$error_output" | grep -qi "failed"; then
        log_fail "Error message does not indicate failure"
        echo "  Error output: $error_output"
        return 1
    fi
    
    log_pass "Missing secret parameter error includes parameter name"
    return 0
}

test_access_denied_includes_name() {
    log_test "Property 2.3: Access denied error includes parameter name"
    TESTS_RUN=$((TESTS_RUN + 1))
    
    export PARAMETER_PREFIX="/wealist/dev"
    export AWS_REGION="ap-northeast-2"
    export MOCK_AWS_BEHAVIOR="access_denied"
    
    local param_name="db/board_db_password"
    local error_output
    
    # Capture stderr
    error_output=$(load_secret "$param_name" 2>&1 >/dev/null || true)
    
    # Check that parameter name appears in error message
    if ! echo "$error_output" | grep -q "$param_name"; then
        log_fail "Parameter name '$param_name' not found in access denied error"
        echo "  Error output: $error_output"
        return 1
    fi
    
    log_pass "Access denied error includes parameter name"
    return 0
}

# =============================================================================
# Property 4: Error messages include troubleshooting guidance
# =============================================================================

# Property: For any parameter-related error, the error message should include
# troubleshooting guidance (e.g., "Check if parameter exists", "Verify IAM permissions").

test_error_includes_troubleshooting_guidance() {
    log_test "Property 4.1: Error message includes troubleshooting guidance"
    TESTS_RUN=$((TESTS_RUN + 1))
    
    export PARAMETER_PREFIX="/wealist/dev"
    export AWS_REGION="ap-northeast-2"
    export MOCK_AWS_BEHAVIOR="not_found"
    
    local param_name="db/user_db_password"
    local error_output
    
    # Capture stderr
    error_output=$(load_param "$param_name" 2>&1 >/dev/null || true)
    
    # Check for troubleshooting section
    if ! echo "$error_output" | grep -qi "troubleshooting"; then
        log_fail "Error message does not include 'Troubleshooting' section"
        echo "  Error output: $error_output"
        return 1
    fi
    
    # Check for specific guidance items
    local guidance_items=(
        "Check if parameter exists"
        "IAM role"
        "ssm:GetParameter"
        "AWS region"
    )
    
    local missing_guidance=()
    for item in "${guidance_items[@]}"; do
        if ! echo "$error_output" | grep -qi "$item"; then
            missing_guidance+=("$item")
        fi
    done
    
    if [ ${#missing_guidance[@]} -gt 0 ]; then
        log_fail "Error message missing guidance items: ${missing_guidance[*]}"
        echo "  Error output: $error_output"
        return 1
    fi
    
    log_pass "Error message includes comprehensive troubleshooting guidance"
    return 0
}

test_secret_error_includes_kms_guidance() {
    log_test "Property 4.2: Secret parameter error includes KMS guidance"
    TESTS_RUN=$((TESTS_RUN + 1))
    
    export PARAMETER_PREFIX="/wealist/dev"
    export AWS_REGION="ap-northeast-2"
    export MOCK_AWS_BEHAVIOR="not_found"
    
    local param_name="jwt/jwt_secret"
    local error_output
    
    # Capture stderr
    error_output=$(load_secret "$param_name" 2>&1 >/dev/null || true)
    
    # Check for KMS-specific guidance
    if ! echo "$error_output" | grep -qi "KMS"; then
        log_fail "Secret parameter error does not mention KMS"
        echo "  Error output: $error_output"
        return 1
    fi
    
    # Check for SecureString mention
    if ! echo "$error_output" | grep -qi "SecureString"; then
        log_fail "Secret parameter error does not mention SecureString"
        echo "  Error output: $error_output"
        return 1
    fi
    
    log_pass "Secret parameter error includes KMS and SecureString guidance"
    return 0
}

test_error_includes_full_parameter_path() {
    log_test "Property 4.3: Error message includes full parameter path"
    TESTS_RUN=$((TESTS_RUN + 1))
    
    export PARAMETER_PREFIX="/wealist/dev"
    export AWS_REGION="ap-northeast-2"
    export MOCK_AWS_BEHAVIOR="not_found"
    
    local param_name="db/board_db_name"
    local expected_full_path="${PARAMETER_PREFIX}/${param_name}"
    local error_output
    
    # Capture stderr
    error_output=$(load_param "$param_name" 2>&1 >/dev/null || true)
    
    # Check that full path appears in error message
    if ! echo "$error_output" | grep -q "$expected_full_path"; then
        log_fail "Full parameter path '$expected_full_path' not found in error message"
        echo "  Error output: $error_output"
        return 1
    fi
    
    log_pass "Error message includes full parameter path"
    return 0
}

# =============================================================================
# Property-Based Testing with Random Parameter Names
# =============================================================================

generate_random_param_name() {
    # Generate random parameter name in format: category/param_name
    local categories=("db" "cache" "jwt" "oauth" "monitor" "service")
    local category=${categories[$RANDOM % ${#categories[@]}]}
    local param="param_$(openssl rand -hex 4)"
    echo "${category}/${param}"
}

test_random_missing_parameters_include_names() {
    log_test "Property 2 (PBT): Random missing parameters always include names in errors"
    TESTS_RUN=$((TESTS_RUN + 1))
    
    export PARAMETER_PREFIX="/wealist/dev"
    export AWS_REGION="ap-northeast-2"
    export MOCK_AWS_BEHAVIOR="not_found"
    
    local iterations=50
    local failures=0
    
    echo "  Running $iterations iterations with random parameter names..."
    
    for i in $(seq 1 $iterations); do
        local random_param
        random_param=$(generate_random_param_name)
        
        local error_output
        error_output=$(load_param "$random_param" 2>&1 >/dev/null || true)
        
        if ! echo "$error_output" | grep -q "$random_param"; then
            echo "  ❌ Iteration $i: Parameter name '$random_param' not in error"
            failures=$((failures + 1))
            
            # Show first failure as example
            if [ $failures -eq 1 ]; then
                echo "  Error output: $error_output"
            fi
        fi
    done
    
    if [ $failures -gt 0 ]; then
        log_fail "$failures out of $iterations random parameters missing from error messages"
        return 1
    fi
    
    log_pass "All $iterations random parameter names included in error messages"
    return 0
}

test_random_parameters_include_troubleshooting() {
    log_test "Property 4 (PBT): Random parameter errors always include troubleshooting"
    TESTS_RUN=$((TESTS_RUN + 1))
    
    export PARAMETER_PREFIX="/wealist/dev"
    export AWS_REGION="ap-northeast-2"
    export MOCK_AWS_BEHAVIOR="not_found"
    
    local iterations=50
    local failures=0
    
    echo "  Running $iterations iterations with random parameter names..."
    
    for i in $(seq 1 $iterations); do
        local random_param
        random_param=$(generate_random_param_name)
        
        local error_output
        error_output=$(load_param "$random_param" 2>&1 >/dev/null || true)
        
        if ! echo "$error_output" | grep -qi "troubleshooting"; then
            echo "  ❌ Iteration $i: No troubleshooting guidance for '$random_param'"
            failures=$((failures + 1))
            
            # Show first failure as example
            if [ $failures -eq 1 ]; then
                echo "  Error output: $error_output"
            fi
        fi
    done
    
    if [ $failures -gt 0 ]; then
        log_fail "$failures out of $iterations errors missing troubleshooting guidance"
        return 1
    fi
    
    log_pass "All $iterations random parameter errors include troubleshooting guidance"
    return 0
}

# =============================================================================
# Edge Cases
# =============================================================================

test_empty_parameter_name() {
    log_test "Property 2 (Edge Case): Empty parameter name handling"
    TESTS_RUN=$((TESTS_RUN + 1))
    
    export PARAMETER_PREFIX="/wealist/dev"
    export AWS_REGION="ap-northeast-2"
    export MOCK_AWS_BEHAVIOR="not_found"
    
    local param_name=""
    local error_output
    
    # Capture stderr
    error_output=$(load_param "$param_name" 2>&1 >/dev/null || true)
    
    # Should still fail gracefully with error message
    if ! echo "$error_output" | grep -qi "failed"; then
        log_fail "Empty parameter name did not produce error"
        echo "  Error output: $error_output"
        return 1
    fi
    
    log_pass "Empty parameter name handled gracefully"
    return 0
}

test_parameter_with_special_chars() {
    log_test "Property 2 (Edge Case): Parameter names with special characters"
    TESTS_RUN=$((TESTS_RUN + 1))
    
    export PARAMETER_PREFIX="/wealist/dev"
    export AWS_REGION="ap-northeast-2"
    export MOCK_AWS_BEHAVIOR="not_found"
    
    local special_params=(
        "db/param-with-dashes"
        "db/param_with_underscores"
        "db/param.with.dots"
    )
    
    local failures=0
    for param in "${special_params[@]}"; do
        local error_output
        error_output=$(load_param "$param" 2>&1 >/dev/null || true)
        
        if ! echo "$error_output" | grep -q "$param"; then
            echo "  ❌ Parameter '$param' not found in error message"
            failures=$((failures + 1))
        fi
    done
    
    if [ $failures -gt 0 ]; then
        log_fail "$failures parameters with special characters not handled correctly"
        return 1
    fi
    
    log_pass "Parameters with special characters handled correctly"
    return 0
}

# =============================================================================
# Run All Tests
# =============================================================================

echo "========================================="
echo "Parameter Error Handling Property Tests"
echo "========================================="
echo ""

# Property 2: Missing parameter error handling
test_missing_parameter_includes_name
test_missing_secret_includes_name
test_access_denied_includes_name

# Property 4: Error messages include troubleshooting guidance
test_error_includes_troubleshooting_guidance
test_secret_error_includes_kms_guidance
test_error_includes_full_parameter_path

# Property-based tests with random inputs
test_random_missing_parameters_include_names
test_random_parameters_include_troubleshooting

# Edge cases
test_empty_parameter_name
test_parameter_with_special_chars

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

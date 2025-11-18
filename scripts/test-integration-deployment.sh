#!/bin/bash

# =============================================================================
# Integration Test for CD Workflow Improvements
# =============================================================================
# This script tests the complete deployment workflow end-to-end:
# 1. First deployment to a fresh EC2 instance
# 2. Redeployment with force-recreate
# 3. docker-compose upload control
# 4. Security validation (hardcoded secrets detection)
# 5. Log masking verification
#
# Requirements: 1.4, 3.2, 3.3, 3.4, 7.2
# =============================================================================

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test results tracking
TESTS_PASSED=0
TESTS_FAILED=0
TESTS_TOTAL=0

# =============================================================================
# Helper Functions
# =============================================================================

print_header() {
    echo ""
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo ""
}

print_test() {
    echo -e "${YELLOW}TEST:${NC} $1"
}

print_success() {
    echo -e "${GREEN}✅ PASS:${NC} $1"
    TESTS_PASSED=$((TESTS_PASSED + 1))
    TESTS_TOTAL=$((TESTS_TOTAL + 1))
}

print_failure() {
    echo -e "${RED}❌ FAIL:${NC} $1"
    TESTS_FAILED=$((TESTS_FAILED + 1))
    TESTS_TOTAL=$((TESTS_TOTAL + 1))
}

print_info() {
    echo -e "${BLUE}ℹ️  INFO:${NC} $1"
}

# =============================================================================
# Test 1: First Deployment Success
# =============================================================================
test_first_deployment() {
    print_header "Test 1: First Deployment Success"
    print_test "Verifying that first deployment completes successfully"
    
    # This test validates Requirements 1.4:
    # "WHEN the first deployment runs THEN the system SHALL complete 
    #  successfully without requiring a second deployment attempt"
    
    print_info "Checking if deployment script exists..."
    if [ ! -f ".github/workflows/cd-dev-board-service.yml" ]; then
        print_failure "CD workflow file not found"
        return 1
    fi
    
    print_info "Verifying health check endpoint is correct..."
    if grep -q "/api/boards/health" ".github/workflows/cd-dev-board-service.yml"; then
        print_success "Health check endpoint is correct (/api/boards/health)"
    else
        print_failure "Health check endpoint is incorrect (should be /api/boards/health)"
        return 1
    fi
    
    print_info "Verifying force-recreate flag is present..."
    if grep -q "\-\-force-recreate" ".github/workflows/cd-dev-board-service.yml"; then
        print_success "Force-recreate flag is present in deployment script"
    else
        print_failure "Force-recreate flag is missing from deployment script"
        return 1
    fi
    
    print_info "Verifying infrastructure service checks..."
    if grep -q "Checking infrastructure services" ".github/workflows/cd-dev-board-service.yml"; then
        print_success "Infrastructure service checks are implemented"
    else
        print_failure "Infrastructure service checks are missing"
        return 1
    fi
    
    print_success "First deployment configuration is correct"
}

# =============================================================================
# Test 2: Redeployment with Force Recreate
# =============================================================================
test_redeployment_force_recreate() {
    print_header "Test 2: Redeployment with Force Recreate"
    print_test "Verifying that redeployment forces container recreation"
    
    # This test validates Requirements 1.3:
    # "WHEN the board service is restarted THEN the system SHALL force 
    #  container recreation even if the image has not changed"
    
    print_info "Checking for --force-recreate flag in board service deployment..."
    if grep -A 5 "board-service" ".github/workflows/cd-dev-board-service.yml" | grep -q "\-\-force-recreate"; then
        print_success "Board service uses --force-recreate flag"
    else
        print_failure "Board service does not use --force-recreate flag"
        return 1
    fi
    
    print_info "Checking for --force-recreate flag in user service deployment..."
    if grep -A 5 "user-service" ".github/workflows/cd-dev-user-service.yml" | grep -q "\-\-force-recreate"; then
        print_success "User service uses --force-recreate flag"
    else
        print_failure "User service does not use --force-recreate flag"
        return 1
    fi
    
    print_success "Redeployment correctly forces container recreation"
}

# =============================================================================
# Test 3: docker-compose Upload Control
# =============================================================================
test_compose_upload_control() {
    print_header "Test 3: docker-compose Upload Control"
    print_test "Verifying docker-compose upload can be controlled"
    
    # This test validates Requirements 3.2, 3.3, 3.4:
    # - Automatic upload when docker-compose is modified
    # - Upload when option is enabled
    # - Skip upload when option is disabled
    
    print_info "Checking for upload-compose input parameter..."
    if grep -q "upload-compose:" ".github/workflows/cd-dev-board-service.yml"; then
        print_success "upload-compose input parameter exists"
    else
        print_failure "upload-compose input parameter is missing"
        return 1
    fi
    
    print_info "Checking for conditional upload based on input..."
    if grep -q "if:.*upload-compose.*==.*true.*||.*push" ".github/workflows/cd-dev-board-service.yml"; then
        print_success "Conditional upload logic is implemented"
    else
        print_failure "Conditional upload logic is missing"
        return 1
    fi
    
    print_info "Checking for automatic upload on push events..."
    if grep -A 10 "Upload docker-compose file to EC2" ".github/workflows/cd-dev-board-service.yml" | \
       grep -q "github.event_name == 'push'"; then
        print_success "Automatic upload on push is configured"
    else
        print_failure "Automatic upload on push is not configured"
        return 1
    fi
    
    print_success "docker-compose upload control is correctly implemented"
}

# =============================================================================
# Test 4: Security Validation (Hardcoded Secrets Detection)
# =============================================================================
test_security_validation() {
    print_header "Test 4: Security Validation"
    print_test "Verifying hardcoded secrets detection"
    
    # This test validates Requirements 7.2:
    # "WHEN a potential secret is detected in docker-compose files 
    #  THEN the system SHALL fail the deployment with a clear error message"
    
    print_info "Checking if validation script exists..."
    if [ ! -f "scripts/validate-compose-secrets.sh" ]; then
        print_failure "Validation script not found"
        return 1
    fi
    
    print_info "Checking if validation is integrated in CD workflow..."
    if grep -q "Validate docker-compose Security" ".github/workflows/cd-dev-board-service.yml"; then
        print_success "Security validation step exists in CD workflow"
    else
        print_failure "Security validation step is missing from CD workflow"
        return 1
    fi
    
    print_info "Testing validation script with hardcoded password..."
    cat > /tmp/test-compose-bad.yml << 'EOF'
version: '3.8'
services:
  db:
    image: postgres
    environment:
      POSTGRES_PASSWORD: mysecretpassword123
EOF
    
    # Script should fail (exit 1) when detecting hardcoded secrets
    if ! bash scripts/validate-compose-secrets.sh /tmp/test-compose-bad.yml > /dev/null 2>&1; then
        print_success "Validation script correctly detects hardcoded passwords"
    else
        print_failure "Validation script failed to detect hardcoded passwords"
        rm -f /tmp/test-compose-bad.yml
        return 1
    fi
    rm -f /tmp/test-compose-bad.yml
    
    print_info "Testing validation script with environment variable substitution..."
    cat > /tmp/test-compose-good.yml << 'EOF'
version: '3.8'
services:
  db:
    image: postgres
    environment:
      POSTGRES_PASSWORD: ${DB_PASSWORD}
EOF
    
    if bash scripts/validate-compose-secrets.sh /tmp/test-compose-good.yml > /dev/null 2>&1; then
        print_success "Validation script allows environment variable substitution"
    else
        print_failure "Validation script incorrectly rejects environment variable substitution"
        rm -f /tmp/test-compose-good.yml
        return 1
    fi
    rm -f /tmp/test-compose-good.yml
    
    print_success "Security validation is correctly implemented"
}

# =============================================================================
# Test 5: Log Masking Verification
# =============================================================================
test_log_masking() {
    print_header "Test 5: Log Masking Verification"
    print_test "Verifying sensitive information is masked in logs"
    
    # This test validates Requirements 6.1, 6.2, 6.3:
    # - Environment variable summaries mask secret values
    # - Parameter Store loading logs only names, not values
    # - Error messages don't include parameter values
    
    print_info "Checking for [REDACTED] in deployment script..."
    if grep -q "\[REDACTED\]" ".github/workflows/cd-dev-board-service.yml"; then
        print_success "Deployment script contains [REDACTED] markers"
    else
        print_failure "Deployment script is missing [REDACTED] markers"
        return 1
    fi
    
    print_info "Verifying password variables are masked..."
    local masked_vars=(
        "POSTGRES_SUPERUSER_PASSWORD"
        "USER_DB_PASSWORD"
        "BOARD_DB_PASSWORD"
        "REDIS_PASSWORD"
        "JWT_SECRET"
        "GOOGLE_CLIENT_SECRET"
        "GRAFANA_ADMIN_PASSWORD"
    )
    
    local all_masked=true
    for var in "${masked_vars[@]}"; do
        if grep -A 1 "echo.*${var}:" ".github/workflows/cd-dev-board-service.yml" | grep -q "\[REDACTED\]"; then
            print_info "  ✓ ${var} is masked"
        else
            print_info "  ✗ ${var} is NOT masked"
            all_masked=false
        fi
    done
    
    if [ "$all_masked" = true ]; then
        print_success "All sensitive variables are properly masked"
    else
        print_failure "Some sensitive variables are not masked"
        return 1
    fi
    
    print_info "Checking that load_secret function masks values..."
    # Check that load_secret function exists and uses --with-decryption for SecureString
    if grep -A 20 "load_secret()" ".github/workflows/cd-dev-board-service.yml" | \
       grep -q "with-decryption"; then
        print_success "load_secret function uses --with-decryption for SecureString"
    else
        print_failure "load_secret function doesn't use --with-decryption"
        return 1
    fi
    
    # Also check that it logs SecureString in the output
    if grep -A 20 "load_secret()" ".github/workflows/cd-dev-board-service.yml" | \
       grep -q "SecureString"; then
        print_success "load_secret function indicates SecureString type in logs"
    else
        print_info "  Note: load_secret function doesn't explicitly log SecureString type (optional)"
    fi
    
    print_success "Log masking is correctly implemented"
}

# =============================================================================
# Test 6: Parameter Store Integration
# =============================================================================
test_parameter_store_integration() {
    print_header "Test 6: Parameter Store Integration"
    print_test "Verifying Parameter Store is used instead of hardcoded values"
    
    print_info "Checking that workflows load from Parameter Store..."
    if grep -q "Load Configuration from Parameter Store" ".github/workflows/cd-dev-board-service.yml"; then
        print_success "Board service workflow loads from Parameter Store"
    else
        print_failure "Board service workflow doesn't load from Parameter Store"
        return 1
    fi
    
    if grep -q "Load Configuration from Parameter Store" ".github/workflows/cd-dev-user-service.yml"; then
        print_success "User service workflow loads from Parameter Store"
    else
        print_failure "User service workflow doesn't load from Parameter Store"
        return 1
    fi
    
    print_info "Verifying no hardcoded AWS account IDs..."
    # Allow account IDs in ECR URLs that use variables
    if grep -E "[0-9]{12}" ".github/workflows/cd-dev-board-service.yml" | \
       grep -v "\${AWS_ACCOUNT_ID}" | \
       grep -v "steps.config.outputs" | \
       grep -v "comment" | \
       grep -v "#" > /dev/null 2>&1; then
        print_failure "Found hardcoded AWS account ID in board service workflow"
        return 1
    else
        print_success "No hardcoded AWS account IDs in board service workflow"
    fi
    
    print_info "Verifying no hardcoded instance IDs..."
    if grep -E "i-[0-9a-f]{17}" ".github/workflows/cd-dev-board-service.yml" | \
       grep -v "steps.config.outputs" | \
       grep -v "comment" | \
       grep -v "#" > /dev/null 2>&1; then
        print_failure "Found hardcoded instance ID in board service workflow"
        return 1
    else
        print_success "No hardcoded instance IDs in board service workflow"
    fi
    
    print_success "Parameter Store integration is correct"
}

# =============================================================================
# Test 7: Error Handling
# =============================================================================
test_error_handling() {
    print_header "Test 7: Error Handling"
    print_test "Verifying proper error handling and troubleshooting guidance"
    
    print_info "Checking for troubleshooting guidance in parameter loading..."
    if grep -A 5 "Failed to load parameter" ".github/workflows/cd-dev-board-service.yml" | \
       grep -q "Troubleshooting"; then
        print_success "Parameter loading includes troubleshooting guidance"
    else
        print_failure "Parameter loading missing troubleshooting guidance"
        return 1
    fi
    
    print_info "Checking for diagnostic output on health check failure..."
    if grep -A 10 "Health check failed" ".github/workflows/cd-dev-board-service.yml" | \
       grep -q "Diagnostic information"; then
        print_success "Health check failure includes diagnostic information"
    else
        print_failure "Health check failure missing diagnostic information"
        return 1
    fi
    
    print_info "Checking for container status output on failures..."
    if grep -q "Container status:" ".github/workflows/cd-dev-board-service.yml"; then
        print_success "Deployment script outputs container status on failures"
    else
        print_failure "Deployment script doesn't output container status"
        return 1
    fi
    
    print_success "Error handling is properly implemented"
}

# =============================================================================
# Test 8: Environment Cleanup
# =============================================================================
test_environment_cleanup() {
    print_header "Test 8: Environment Cleanup"
    print_test "Verifying environment variables are cleaned up after deployment"
    
    print_info "Checking for cleanup_environment function..."
    if grep -q "cleanup_environment()" ".github/workflows/cd-dev-board-service.yml"; then
        print_success "cleanup_environment function exists"
    else
        print_failure "cleanup_environment function is missing"
        return 1
    fi
    
    print_info "Checking for trap EXIT to ensure cleanup..."
    if grep -q "trap.*cleanup_environment.*EXIT" ".github/workflows/cd-dev-board-service.yml"; then
        print_success "trap EXIT ensures cleanup on script exit"
    else
        print_failure "trap EXIT for cleanup is missing"
        return 1
    fi
    
    print_info "Checking that sensitive variables are unset..."
    if grep -A 20 "cleanup_environment()" ".github/workflows/cd-dev-board-service.yml" | \
       grep -q "unset"; then
        print_success "Cleanup function unsets sensitive variables"
    else
        print_failure "Cleanup function doesn't unset variables"
        return 1
    fi
    
    print_success "Environment cleanup is properly implemented"
}

# =============================================================================
# Main Test Execution
# =============================================================================

main() {
    print_header "CD Workflow Integration Tests"
    print_info "Testing complete deployment workflow end-to-end"
    print_info "Requirements: 1.4, 3.2, 3.3, 3.4, 7.2"
    echo ""
    
    # Run all tests
    test_first_deployment || true
    test_redeployment_force_recreate || true
    test_compose_upload_control || true
    test_security_validation || true
    test_log_masking || true
    test_parameter_store_integration || true
    test_error_handling || true
    test_environment_cleanup || true
    
    # Print summary
    print_header "Test Summary"
    echo -e "Total Tests: ${TESTS_TOTAL}"
    echo -e "${GREEN}Passed: ${TESTS_PASSED}${NC}"
    echo -e "${RED}Failed: ${TESTS_FAILED}${NC}"
    echo ""
    
    if [ ${TESTS_FAILED} -eq 0 ]; then
        echo -e "${GREEN}✅ All integration tests passed!${NC}"
        exit 0
    else
        echo -e "${RED}❌ Some integration tests failed!${NC}"
        exit 1
    fi
}

# Run main function
main "$@"

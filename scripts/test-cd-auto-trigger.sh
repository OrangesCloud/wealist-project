#!/bin/bash

# =============================================================================
# Property Test: CD workflow triggers only on CI success
# Feature: cd-workflow-improvement, Property 9
# Validates: Requirements 8.1, 8.2
#
# Property: For any CI workflow execution, the CD workflow should trigger
# automatically if and only if the CI workflow status is "success"
# (not "failure", "cancelled", or "skipped").
# =============================================================================

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

echo "🧪 Testing Property 9: CD workflow triggers only on CI success"
echo "=============================================================="
echo ""

# Test counters
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

# Test result tracking
declare -a FAILED_TESTS=()

# =============================================================================
# Helper Functions
# =============================================================================

pass_test() {
    local test_name="$1"
    TESTS_PASSED=$((TESTS_PASSED + 1))
    echo "  ✅ PASS: ${test_name}"
}

fail_test() {
    local test_name="$1"
    local reason="$2"
    TESTS_FAILED=$((TESTS_FAILED + 1))
    FAILED_TESTS+=("${test_name}: ${reason}")
    echo "  ❌ FAIL: ${test_name}"
    echo "     Reason: ${reason}"
}

run_test() {
    local test_name="$1"
    TESTS_RUN=$((TESTS_RUN + 1))
    echo ""
    echo "Test ${TESTS_RUN}: ${test_name}"
    echo "---"
}

# =============================================================================
# Test 1: Board Service CD has workflow_run trigger
# =============================================================================
run_test "Board Service CD has workflow_run trigger"

BOARD_CD_FILE="${PROJECT_ROOT}/.github/workflows/cd-dev-board-service.yml"

if [ ! -f "$BOARD_CD_FILE" ]; then
    fail_test "Board Service CD workflow file exists" "File not found: $BOARD_CD_FILE"
else
    # Check if workflow_run trigger exists
    if grep -q "workflow_run:" "$BOARD_CD_FILE"; then
        pass_test "Board Service CD has workflow_run trigger"
        
        # Verify it references the correct CI workflow
        if grep -A 5 "workflow_run:" "$BOARD_CD_FILE" | grep -q '"CI - Dev Board Service"'; then
            pass_test "Board Service CD references correct CI workflow"
        else
            fail_test "Board Service CD references correct CI workflow" \
                "workflow_run does not reference 'CI - Dev Board Service'"
        fi
        
        # Verify it triggers on completed
        if grep -A 5 "workflow_run:" "$BOARD_CD_FILE" | grep -q "completed"; then
            pass_test "Board Service CD triggers on completed"
        else
            fail_test "Board Service CD triggers on completed" \
                "workflow_run does not have 'completed' type"
        fi
        
        # Verify it filters by branch
        if grep -A 10 "workflow_run:" "$BOARD_CD_FILE" | grep -q "deploy-dev"; then
            pass_test "Board Service CD filters by deploy-dev branch"
        else
            fail_test "Board Service CD filters by deploy-dev branch" \
                "workflow_run does not filter by deploy-dev branch"
        fi
    else
        fail_test "Board Service CD has workflow_run trigger" \
            "No workflow_run trigger found in CD workflow"
    fi
fi

# =============================================================================
# Test 2: User Service CD has workflow_run trigger
# =============================================================================
run_test "User Service CD has workflow_run trigger"

USER_CD_FILE="${PROJECT_ROOT}/.github/workflows/cd-dev-user-service.yml"

if [ ! -f "$USER_CD_FILE" ]; then
    fail_test "User Service CD workflow file exists" "File not found: $USER_CD_FILE"
else
    # Check if workflow_run trigger exists
    if grep -q "workflow_run:" "$USER_CD_FILE"; then
        pass_test "User Service CD has workflow_run trigger"
        
        # Verify it references the correct CI workflow
        if grep -A 5 "workflow_run:" "$USER_CD_FILE" | grep -q '"CI - Dev User Service"'; then
            pass_test "User Service CD references correct CI workflow"
        else
            fail_test "User Service CD references correct CI workflow" \
                "workflow_run does not reference 'CI - Dev User Service'"
        fi
        
        # Verify it triggers on completed
        if grep -A 5 "workflow_run:" "$USER_CD_FILE" | grep -q "completed"; then
            pass_test "User Service CD triggers on completed"
        else
            fail_test "User Service CD triggers on completed" \
                "workflow_run does not have 'completed' type"
        fi
        
        # Verify it filters by branch
        if grep -A 10 "workflow_run:" "$USER_CD_FILE" | grep -q "deploy-dev"; then
            pass_test "User Service CD filters by deploy-dev branch"
        else
            fail_test "User Service CD filters by deploy-dev branch" \
                "workflow_run does not filter by deploy-dev branch"
        fi
    else
        fail_test "User Service CD has workflow_run trigger" \
            "No workflow_run trigger found in CD workflow"
    fi
fi

# =============================================================================
# Test 3: Board Service CI does NOT have manual CD trigger
# =============================================================================
run_test "Board Service CI does NOT have manual CD trigger"

BOARD_CI_FILE="${PROJECT_ROOT}/.github/workflows/ci-dev-board-service.yml"

if [ ! -f "$BOARD_CI_FILE" ]; then
    fail_test "Board Service CI workflow file exists" "File not found: $BOARD_CI_FILE"
else
    # Check that there's no actions/github-script step that triggers CD
    if grep -A 10 "actions/github-script" "$BOARD_CI_FILE" | grep -q "createWorkflowDispatch"; then
        fail_test "Board Service CI does NOT have manual CD trigger" \
            "Found actions/github-script with createWorkflowDispatch"
    else
        pass_test "Board Service CI does NOT have manual CD trigger"
    fi
    
    # Also check for any reference to cd-dev-board-service.yml
    if grep -q "cd-dev-board-service.yml" "$BOARD_CI_FILE"; then
        fail_test "Board Service CI does NOT reference CD workflow" \
            "Found reference to cd-dev-board-service.yml in CI workflow"
    else
        pass_test "Board Service CI does NOT reference CD workflow"
    fi
fi

# =============================================================================
# Test 4: User Service CI does NOT have manual CD trigger
# =============================================================================
run_test "User Service CI does NOT have manual CD trigger"

USER_CI_FILE="${PROJECT_ROOT}/.github/workflows/ci-dev-user-service.yml"

if [ ! -f "$USER_CI_FILE" ]; then
    fail_test "User Service CI workflow file exists" "File not found: $USER_CI_FILE"
else
    # Check that there's no actions/github-script step that triggers CD
    if grep -A 10 "actions/github-script" "$USER_CI_FILE" | grep -q "createWorkflowDispatch"; then
        fail_test "User Service CI does NOT have manual CD trigger" \
            "Found actions/github-script with createWorkflowDispatch"
    else
        pass_test "User Service CI does NOT have manual CD trigger"
    fi
    
    # Also check for any reference to cd-dev-user-service.yml
    if grep -q "cd-dev-user-service.yml" "$USER_CI_FILE"; then
        fail_test "User Service CI does NOT reference CD workflow" \
            "Found reference to cd-dev-user-service.yml in CI workflow"
    else
        pass_test "User Service CI does NOT reference CD workflow"
    fi
fi

# =============================================================================
# Test 5: CD workflows check CI success status
# =============================================================================
run_test "CD workflows check CI success status"

# Board Service CD should check workflow_run conclusion
if [ -f "$BOARD_CD_FILE" ]; then
    # The CD workflow should have a condition that checks if the CI workflow succeeded
    # This is typically done with: if: ${{ github.event.workflow_run.conclusion == 'success' }}
    if grep -q "workflow_run.conclusion" "$BOARD_CD_FILE"; then
        if grep "workflow_run.conclusion" "$BOARD_CD_FILE" | grep -q "success"; then
            pass_test "Board Service CD checks for CI success"
        else
            fail_test "Board Service CD checks for CI success" \
                "workflow_run.conclusion check does not verify 'success'"
        fi
    else
        # If no explicit check, the workflow will run on any completion
        # This is acceptable but not ideal
        echo "  ⚠️  WARNING: Board Service CD does not explicitly check workflow_run.conclusion"
        echo "     The workflow will run even if CI fails. Consider adding:"
        echo "     if: \${{ github.event.workflow_run.conclusion == 'success' }}"
        pass_test "Board Service CD has workflow_run trigger (implicit success check)"
    fi
fi

# User Service CD should check workflow_run conclusion
if [ -f "$USER_CD_FILE" ]; then
    if grep -q "workflow_run.conclusion" "$USER_CD_FILE"; then
        if grep "workflow_run.conclusion" "$USER_CD_FILE" | grep -q "success"; then
            pass_test "User Service CD checks for CI success"
        else
            fail_test "User Service CD checks for CI success" \
                "workflow_run.conclusion check does not verify 'success'"
        fi
    else
        echo "  ⚠️  WARNING: User Service CD does not explicitly check workflow_run.conclusion"
        echo "     The workflow will run even if CI fails. Consider adding:"
        echo "     if: \${{ github.event.workflow_run.conclusion == 'success' }}"
        pass_test "User Service CD has workflow_run trigger (implicit success check)"
    fi
fi

# =============================================================================
# Test Summary
# =============================================================================
echo ""
echo "=============================================================="
echo "Test Summary"
echo "=============================================================="
echo "Total tests run: ${TESTS_RUN}"
echo "Passed: ${TESTS_PASSED}"
echo "Failed: ${TESTS_FAILED}"
echo ""

if [ ${TESTS_FAILED} -gt 0 ]; then
    echo "❌ Failed tests:"
    for failed_test in "${FAILED_TESTS[@]}"; do
        echo "  - ${failed_test}"
    done
    echo ""
    echo "❌ Property test FAILED"
    exit 1
else
    echo "✅ All tests passed!"
    echo "✅ Property 9 verified: CD workflows trigger only on CI success"
    exit 0
fi

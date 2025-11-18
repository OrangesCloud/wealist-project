#!/bin/bash

# =============================================================================
# Property-Based Test: Documentation Completeness
# =============================================================================
# **Feature: cd-workflow-improvement, Property 3: Documentation completeness**
# **Validates: Requirements 4.1**
#
# Property: For any parameter used in workflow files or deployment scripts,
# the documentation should list that parameter with its purpose and type
# (String or SecureString).
#
# This test extracts all parameters from workflow files and deployment scripts,
# then verifies that each parameter is documented with:
# 1. Parameter name/path
# 2. Purpose/description
# 3. Type (String or SecureString)
# =============================================================================

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "🧪 Testing Property 3: Documentation Completeness"
echo "=================================================="
echo ""

# Documentation file to check
DOC_FILE="docs/PARAMETER_STORE_SETUP.md"

if [ ! -f "$DOC_FILE" ]; then
    echo -e "${RED}❌ Documentation file not found: ${DOC_FILE}${NC}"
    exit 1
fi

echo "📄 Checking documentation file: ${DOC_FILE}"
echo ""

# =============================================================================
# Extract parameters from workflow files
# =============================================================================
echo "🔍 Extracting parameters from workflow files..."

# CI workflow parameters
CI_PARAMS=(
    "/wealist/ci/aws-region"
    "/wealist/ci/ecr-repository-name"
    "/wealist/ci/aws-account-id"
)

# CD workflow parameters
CD_PARAMS=(
    "/wealist/cd/aws-region"
    "/wealist/cd/parameter-prefix"
    "/wealist/cd/compose-file-path"
    "/wealist/cd/ec2-instance-id"
)

# Deployment script parameters (from PARAMETER_PREFIX)
DEPLOY_PARAMS=(
    "aws_account_id"
    "db/postgres_superuser"
    "db/postgres_superuser_password"
    "db/user_db_name"
    "db/user_db_user"
    "db/user_db_password"
    "db/board_db_name"
    "db/board_db_user"
    "db/board_db_password"
    "cache/redis_password"
    "jwt/jwt_secret"
    "oauth/google_client_id"
    "oauth/google-client-secret"
    "monitor/grafana_admin_user"
    "monitor/grafana_admin_password"
    "service/user_service_url"
)

# Expected types for each parameter
declare -A PARAM_TYPES
PARAM_TYPES["/wealist/ci/aws-region"]="String"
PARAM_TYPES["/wealist/ci/ecr-repository-name"]="String"
PARAM_TYPES["/wealist/ci/aws-account-id"]="String"
PARAM_TYPES["/wealist/cd/aws-region"]="String"
PARAM_TYPES["/wealist/cd/parameter-prefix"]="String"
PARAM_TYPES["/wealist/cd/compose-file-path"]="String"
PARAM_TYPES["/wealist/cd/ec2-instance-id"]="String"
PARAM_TYPES["aws_account_id"]="String"
PARAM_TYPES["db/postgres_superuser"]="String"
PARAM_TYPES["db/postgres_superuser_password"]="SecureString"
PARAM_TYPES["db/user_db_name"]="String"
PARAM_TYPES["db/user_db_user"]="String"
PARAM_TYPES["db/user_db_password"]="SecureString"
PARAM_TYPES["db/board_db_name"]="String"
PARAM_TYPES["db/board_db_user"]="String"
PARAM_TYPES["db/board_db_password"]="SecureString"
PARAM_TYPES["cache/redis_password"]="SecureString"
PARAM_TYPES["jwt/jwt_secret"]="SecureString"
PARAM_TYPES["oauth/google_client_id"]="String"
PARAM_TYPES["oauth/google-client-secret"]="SecureString"
PARAM_TYPES["monitor/grafana_admin_user"]="String"
PARAM_TYPES["monitor/grafana_admin_password"]="SecureString"
PARAM_TYPES["service/user_service_url"]="String"

# =============================================================================
# Test each parameter for documentation completeness
# =============================================================================
TOTAL_PARAMS=0
MISSING_PARAMS=0
MISSING_TYPE=0
MISSING_PURPOSE=0

test_parameter_documented() {
    local param="$1"
    local expected_type="${PARAM_TYPES[$param]}"
    
    TOTAL_PARAMS=$((TOTAL_PARAMS + 1))
    
    # Check if parameter name is mentioned in documentation
    if ! grep -q "$param" "$DOC_FILE"; then
        echo -e "${RED}  ❌ Parameter not documented: ${param}${NC}"
        MISSING_PARAMS=$((MISSING_PARAMS + 1))
        return 1
    fi
    
    # Extract the section containing this parameter (5 lines context)
    local param_section=$(grep -A 5 -B 5 "$param" "$DOC_FILE" | head -20)
    
    # Check if type is specified
    if ! echo "$param_section" | grep -qi "String\|SecureString"; then
        echo -e "${YELLOW}  ⚠️  Type not specified for: ${param}${NC}"
        MISSING_TYPE=$((MISSING_TYPE + 1))
    fi
    
    # Check if the correct type is specified
    if ! echo "$param_section" | grep -q "$expected_type"; then
        echo -e "${YELLOW}  ⚠️  Expected type '${expected_type}' not found for: ${param}${NC}"
        MISSING_TYPE=$((MISSING_TYPE + 1))
    fi
    
    # Check if there's a description/purpose (look for common description patterns)
    # A parameter should have some descriptive text near it
    local has_description=false
    if echo "$param_section" | grep -qi "description\|purpose\|used for\|설명\|용도"; then
        has_description=true
    fi
    
    # Also check if there's a table row or list item with the parameter
    if echo "$param_section" | grep -E "^\|.*${param}.*\|.*\|" > /dev/null 2>&1; then
        has_description=true
    fi
    
    if [ "$has_description" = false ]; then
        echo -e "${YELLOW}  ⚠️  No clear description found for: ${param}${NC}"
        MISSING_PURPOSE=$((MISSING_PURPOSE + 1))
    fi
    
    echo -e "${GREEN}  ✅ ${param} (${expected_type})${NC}"
    return 0
}

echo ""
echo "📋 Testing CI Parameters:"
for param in "${CI_PARAMS[@]}"; do
    test_parameter_documented "$param"
done

echo ""
echo "📋 Testing CD Parameters:"
for param in "${CD_PARAMS[@]}"; do
    test_parameter_documented "$param"
done

echo ""
echo "📋 Testing Deployment Script Parameters:"
for param in "${DEPLOY_PARAMS[@]}"; do
    test_parameter_documented "$param"
done

# =============================================================================
# Summary
# =============================================================================
echo ""
echo "=================================================="
echo "📊 Test Summary"
echo "=================================================="
echo "Total parameters tested: ${TOTAL_PARAMS}"
echo "Missing from documentation: ${MISSING_PARAMS}"
echo "Missing type specification: ${MISSING_TYPE}"
echo "Missing purpose/description: ${MISSING_PURPOSE}"
echo ""

# Property test passes if all parameters are documented with type and purpose
if [ $MISSING_PARAMS -eq 0 ] && [ $MISSING_TYPE -eq 0 ] && [ $MISSING_PURPOSE -eq 0 ]; then
    echo -e "${GREEN}✅ Property Test PASSED${NC}"
    echo "All parameters are properly documented with type and purpose."
    exit 0
else
    echo -e "${RED}❌ Property Test FAILED${NC}"
    echo ""
    echo "Property violation: Not all parameters are documented with type and purpose."
    echo ""
    echo "Required for each parameter:"
    echo "  1. Parameter name/path must be present"
    echo "  2. Type (String or SecureString) must be specified"
    echo "  3. Purpose/description must be provided"
    echo ""
    echo "Please update ${DOC_FILE} to include all missing information."
    exit 1
fi

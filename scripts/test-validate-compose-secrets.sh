#!/bin/bash

# Property-Based Tests for docker-compose security validation
# Feature: cd-workflow-improvement, Property 7: Secret pattern detection in docker-compose files
# Feature: cd-workflow-improvement, Property 8: Environment variable substitution allowed
# Validates: Requirements 7.2, 7.3, 7.5

# Note: Do not use 'set -e' because we need to capture exit codes from validation script

# 색상 코드
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 테스트 카운터
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# 임시 디렉토리 생성
TEST_DIR=$(mktemp -d)
trap "rm -rf $TEST_DIR" EXIT

# 테스트 결과 출력
test_result() {
    local test_name="$1"
    local expected="$2"
    local actual="$3"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    if [ "$expected" = "$actual" ]; then
        echo -e "${GREEN}✅ PASS${NC}: $test_name"
        PASSED_TESTS=$((PASSED_TESTS + 1))
        return 0
    else
        echo -e "${RED}❌ FAIL${NC}: $test_name"
        echo -e "   Expected: $expected, Got: $actual"
        FAILED_TESTS=$((FAILED_TESTS + 1))
        return 1
    fi
}

echo "=========================================="
echo "Property-Based Tests for Compose Validation"
echo "=========================================="
echo ""

# Property 7: Secret pattern detection in docker-compose files
echo "🧪 Property 7: Secret pattern detection"
echo "   For any docker-compose file containing hardcoded secret patterns,"
echo "   the validation function should detect and reject the file."
echo ""

# Test 1: 하드코딩된 비밀번호 감지
echo "Test 1: Detect hardcoded password"
cat > "$TEST_DIR/compose-hardcoded-password.yml" << 'EOF'
version: '3.8'
services:
  db:
    image: postgres:14
    environment:
      POSTGRES_PASSWORD: mysecretpassword123
EOF

./scripts/validate-compose-secrets.sh "$TEST_DIR/compose-hardcoded-password.yml" > /dev/null 2>&1
test_result "Hardcoded password detection" "1" "$?"

# Test 2: AWS Access Key 감지
echo "Test 2: Detect AWS Access Key"
cat > "$TEST_DIR/compose-aws-key.yml" << 'EOF'
version: '3.8'
services:
  app:
    image: myapp:latest
    environment:
      AWS_ACCESS_KEY_ID: AKIAIOSFODNN7EXAMPLE
EOF

./scripts/validate-compose-secrets.sh "$TEST_DIR/compose-aws-key.yml" > /dev/null 2>&1
test_result "AWS Access Key detection" "1" "$?"

# Test 3: JWT Secret 감지
echo "Test 3: Detect hardcoded JWT secret"
cat > "$TEST_DIR/compose-jwt-secret.yml" << 'EOF'
version: '3.8'
services:
  api:
    image: api:latest
    environment:
      JWT_SECRET: my-super-secret-jwt-key-12345
EOF

./scripts/validate-compose-secrets.sh "$TEST_DIR/compose-jwt-secret.yml" > /dev/null 2>&1
test_result "JWT secret detection" "1" "$?"

# Test 4: API Key 감지
echo "Test 4: Detect hardcoded API key"
cat > "$TEST_DIR/compose-api-key.yml" << 'EOF'
version: '3.8'
services:
  service:
    image: service:latest
    environment:
      API_KEY: sk-1234567890abcdef
EOF

./scripts/validate-compose-secrets.sh "$TEST_DIR/compose-api-key.yml" > /dev/null 2>&1
test_result "API key detection" "1" "$?"

# Test 5: 여러 시크릿 패턴 동시 감지
echo "Test 5: Detect multiple secret patterns"
cat > "$TEST_DIR/compose-multiple-secrets.yml" << 'EOF'
version: '3.8'
services:
  db:
    image: postgres:14
    environment:
      POSTGRES_PASSWORD: dbpassword123
  redis:
    image: redis:7
    environment:
      REDIS_PASSWORD: redispass456
EOF

./scripts/validate-compose-secrets.sh "$TEST_DIR/compose-multiple-secrets.yml" > /dev/null 2>&1
test_result "Multiple secrets detection" "1" "$?"

echo ""
echo "=========================================="
echo "🧪 Property 8: Environment variable substitution allowed"
echo "   For any docker-compose file using environment variable substitution,"
echo "   the validation function should allow the file to pass."
echo ""

# Test 6: 환경변수 치환 (${VAR}) 허용
echo "Test 6: Allow environment variable substitution with \${VAR}"
cat > "$TEST_DIR/compose-env-var-braces.yml" << 'EOF'
version: '3.8'
services:
  db:
    image: postgres:14
    environment:
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
      REDIS_PASSWORD: ${REDIS_PASSWORD}
EOF

./scripts/validate-compose-secrets.sh "$TEST_DIR/compose-env-var-braces.yml" > /dev/null 2>&1
test_result "Environment variable substitution \${VAR}" "0" "$?"

# Test 7: 환경변수 치환 ($VAR) 허용
echo "Test 7: Allow environment variable substitution with \$VAR"
cat > "$TEST_DIR/compose-env-var-dollar.yml" << 'EOF'
version: '3.8'
services:
  db:
    image: postgres:14
    environment:
      POSTGRES_PASSWORD: $POSTGRES_PASSWORD
      JWT_SECRET: $JWT_SECRET
EOF

./scripts/validate-compose-secrets.sh "$TEST_DIR/compose-env-var-dollar.yml" > /dev/null 2>&1
test_result "Environment variable substitution \$VAR" "0" "$?"

# Test 8: 혼합 사용 (환경변수 + 일반 값)
echo "Test 8: Mixed usage - env vars for secrets, plain values for non-secrets"
cat > "$TEST_DIR/compose-mixed.yml" << 'EOF'
version: '3.8'
services:
  db:
    image: postgres:14
    environment:
      POSTGRES_USER: postgres
      POSTGRES_DB: mydb
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
  app:
    image: myapp:latest
    environment:
      APP_NAME: MyApplication
      APP_PORT: 8080
      JWT_SECRET: ${JWT_SECRET}
EOF

./scripts/validate-compose-secrets.sh "$TEST_DIR/compose-mixed.yml" > /dev/null 2>&1
test_result "Mixed usage (env vars + plain values)" "0" "$?"

# Test 9: 모든 환경변수 치환 사용
echo "Test 9: All environment variables using substitution"
cat > "$TEST_DIR/compose-all-env-vars.yml" << 'EOF'
version: '3.8'
services:
  db:
    image: postgres:14
    environment:
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
  redis:
    image: redis:7
    environment:
      REDIS_PASSWORD: ${REDIS_PASSWORD}
  api:
    image: api:latest
    environment:
      JWT_SECRET: ${JWT_SECRET}
      API_KEY: ${API_KEY}
EOF

./scripts/validate-compose-secrets.sh "$TEST_DIR/compose-all-env-vars.yml" > /dev/null 2>&1
test_result "All secrets using env var substitution" "0" "$?"

# Test 10: 빈 파일 (시크릿 없음)
echo "Test 10: Empty compose file (no secrets)"
cat > "$TEST_DIR/compose-empty.yml" << 'EOF'
version: '3.8'
services:
  app:
    image: nginx:latest
    ports:
      - "80:80"
EOF

./scripts/validate-compose-secrets.sh "$TEST_DIR/compose-empty.yml" > /dev/null 2>&1
test_result "Empty compose file (no secrets)" "0" "$?"

# Property-based test: 랜덤 시크릿 값 생성 및 테스트
echo ""
echo "=========================================="
echo "🧪 Property-Based Test: Random secret values"
echo "   Generate random secret values and verify detection"
echo ""

# Test 11-15: 랜덤 비밀번호 생성 및 감지
for i in {1..5}; do
    RANDOM_PASSWORD=$(openssl rand -base64 16 | tr -d '\n')
    echo "Test $((10 + i)): Random password detection (iteration $i)"
    
    cat > "$TEST_DIR/compose-random-$i.yml" << EOF
version: '3.8'
services:
  db:
    image: postgres:14
    environment:
      POSTGRES_PASSWORD: $RANDOM_PASSWORD
EOF
    
    ./scripts/validate-compose-secrets.sh "$TEST_DIR/compose-random-$i.yml" > /dev/null 2>&1
    test_result "Random password detection (iteration $i)" "1" "$?"
done

# Test 16-20: 랜덤 환경변수 이름 생성 및 허용
echo ""
for i in {1..5}; do
    RANDOM_VAR_NAME="SECRET_$(openssl rand -hex 4 | tr '[:lower:]' '[:upper:]')"
    echo "Test $((15 + i)): Random env var name (iteration $i): \${$RANDOM_VAR_NAME}"
    
    cat > "$TEST_DIR/compose-random-env-$i.yml" << EOF
version: '3.8'
services:
  db:
    image: postgres:14
    environment:
      POSTGRES_PASSWORD: \${${RANDOM_VAR_NAME}}
EOF
    
    ./scripts/validate-compose-secrets.sh "$TEST_DIR/compose-random-env-$i.yml" > /dev/null 2>&1
    test_result "Random env var name allowed (iteration $i)" "0" "$?"
done

# 최종 결과 출력
echo ""
echo "=========================================="
echo "Test Summary"
echo "=========================================="
echo -e "Total Tests:  $TOTAL_TESTS"
echo -e "${GREEN}Passed:       $PASSED_TESTS${NC}"
echo -e "${RED}Failed:       $FAILED_TESTS${NC}"
echo ""

if [ $FAILED_TESTS -eq 0 ]; then
    echo -e "${GREEN}✅ All property-based tests passed!${NC}"
    exit 0
else
    echo -e "${RED}❌ Some tests failed!${NC}"
    exit 1
fi

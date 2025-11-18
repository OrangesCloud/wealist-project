#!/bin/bash

# docker-compose 파일에서 하드코딩된 시크릿을 감지하는 스크립트
# Feature: cd-workflow-improvement, Property 7: Secret pattern detection in docker-compose files
# Feature: cd-workflow-improvement, Property 8: Environment variable substitution allowed

set -e

# 색상 코드
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 사용법 출력
usage() {
    echo "Usage: $0 <docker-compose-file>"
    echo ""
    echo "Validates docker-compose files for hardcoded secrets."
    echo "Allows environment variable substitution syntax: \${VARIABLE_NAME}"
    echo ""
    echo "Example:"
    echo "  $0 docker-compose.yml"
    exit 1
}

# 인자 확인
if [ $# -ne 1 ]; then
    usage
fi

COMPOSE_FILE="$1"

# 파일 존재 확인
if [ ! -f "$COMPOSE_FILE" ]; then
    echo -e "${RED}❌ Error: File not found: $COMPOSE_FILE${NC}" >&2
    exit 1
fi

echo "🔍 Scanning docker-compose file for hardcoded secrets..."
echo "📄 File: $COMPOSE_FILE"
echo ""

found_secrets=0

# 1. 하드코딩된 비밀번호 패턴 감지
# 패턴: password: <값> (환경변수 치환이 아닌 경우)
# 허용: password: ${PASSWORD}, password: $PASSWORD
# 불허: password: mysecret123, password: "mysecret123"
echo "🔎 Checking for hardcoded passwords..."
if grep -n -E '^\s*(POSTGRES_PASSWORD|REDIS_PASSWORD|DB_PASSWORD|password):\s*[^$#]' "$COMPOSE_FILE" | grep -v -E ':\s*\$\{' | grep -v -E ':\s*\$[A-Z_]'; then
    echo -e "${RED}❌ Found hardcoded password(s)${NC}" >&2
    echo -e "${YELLOW}💡 Use environment variable substitution: \${PASSWORD}${NC}" >&2
    found_secrets=1
else
    echo -e "${GREEN}✅ No hardcoded passwords found${NC}"
fi
echo ""

# 2. AWS Access Key 패턴 감지
# 패턴: AKIA로 시작하는 20자 문자열
echo "🔎 Checking for AWS Access Keys..."
if grep -n -E 'AKIA[0-9A-Z]{16}' "$COMPOSE_FILE"; then
    echo -e "${RED}❌ Found AWS Access Key(s)${NC}" >&2
    echo -e "${YELLOW}💡 Use environment variable substitution: \${AWS_ACCESS_KEY_ID}${NC}" >&2
    found_secrets=1
else
    echo -e "${GREEN}✅ No AWS Access Keys found${NC}"
fi
echo ""

# 3. JWT Secret 패턴 감지
# 패턴: jwt*secret: <값> (환경변수 치환이 아닌 경우)
echo "🔎 Checking for hardcoded JWT secrets..."
if grep -n -i -E '(jwt.*secret|secret.*jwt):\s*[^$#]' "$COMPOSE_FILE" | grep -v -E ':\s*\$\{' | grep -v -E ':\s*\$[A-Z_]'; then
    echo -e "${RED}❌ Found hardcoded JWT secret(s)${NC}" >&2
    echo -e "${YELLOW}💡 Use environment variable substitution: \${JWT_SECRET}${NC}" >&2
    found_secrets=1
else
    echo -e "${GREEN}✅ No hardcoded JWT secrets found${NC}"
fi
echo ""

# 4. API Key 패턴 감지
# 패턴: api*key: <값> (환경변수 치환이 아닌 경우)
echo "🔎 Checking for hardcoded API keys..."
if grep -n -i -E '(api.*key|key.*api):\s*[^$#]' "$COMPOSE_FILE" | grep -v -E ':\s*\$\{' | grep -v -E ':\s*\$[A-Z_]'; then
    echo -e "${RED}❌ Found hardcoded API key(s)${NC}" >&2
    echo -e "${YELLOW}💡 Use environment variable substitution: \${API_KEY}${NC}" >&2
    found_secrets=1
else
    echo -e "${GREEN}✅ No hardcoded API keys found${NC}"
fi
echo ""

# 결과 출력
if [ $found_secrets -eq 1 ]; then
    echo -e "${RED}❌ Security validation failed!${NC}" >&2
    echo -e "${RED}   Remove hardcoded secrets and use environment variable substitution.${NC}" >&2
    exit 1
fi

echo -e "${GREEN}✅ Security validation passed!${NC}"
echo -e "${GREEN}   No hardcoded secrets found in docker-compose file.${NC}"
exit 0

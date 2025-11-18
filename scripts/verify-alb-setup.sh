#!/bin/bash

# ALB 설정 검증 스크립트
# 이 스크립트는 ALB를 통한 API 호출이 정상적으로 동작하는지 확인합니다.

set -e

# 색상 정의
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 설정
ALB_URL="${ALB_URL:-https://api.wealist.co.kr}"
VERBOSE="${VERBOSE:-false}"

echo "=========================================="
echo "ALB 설정 검증 스크립트"
echo "=========================================="
echo "ALB URL: $ALB_URL"
echo ""

# 함수: HTTP 상태 코드 확인
check_endpoint() {
    local url=$1
    local expected_status=$2
    local description=$3
    
    echo -n "테스트: $description ... "
    
    if [ "$VERBOSE" = "true" ]; then
        echo ""
        echo "URL: $url"
    fi
    
    # HTTP 상태 코드 가져오기
    status_code=$(curl -s -o /dev/null -w "%{http_code}" "$url" 2>/dev/null || echo "000")
    
    if [ "$status_code" = "$expected_status" ]; then
        echo -e "${GREEN}✓ PASS${NC} (HTTP $status_code)"
        return 0
    else
        echo -e "${RED}✗ FAIL${NC} (Expected: $expected_status, Got: $status_code)"
        return 1
    fi
}

# 함수: 상세 응답 확인
check_endpoint_verbose() {
    local url=$1
    local description=$2
    
    echo ""
    echo "=========================================="
    echo "상세 테스트: $description"
    echo "=========================================="
    echo "URL: $url"
    echo ""
    
    response=$(curl -s -w "\n\nHTTP Status: %{http_code}\nTime Total: %{time_total}s\n" "$url" 2>/dev/null || echo "Error: Connection failed")
    
    echo "$response"
    echo ""
}

# 테스트 카운터
total_tests=0
passed_tests=0
failed_tests=0

# 1. User Service Health Check
echo "=========================================="
echo "1. User Service 검증"
echo "=========================================="

total_tests=$((total_tests + 1))
if check_endpoint "$ALB_URL/api/users/actuator/health" "200" "User Service Health Check"; then
    passed_tests=$((passed_tests + 1))
else
    failed_tests=$((failed_tests + 1))
fi

# 2. Board Service Health Check
echo ""
echo "=========================================="
echo "2. Board Service 검증"
echo "=========================================="

total_tests=$((total_tests + 1))
if check_endpoint "$ALB_URL/api/boards/health" "200" "Board Service Health Check"; then
    passed_tests=$((passed_tests + 1))
else
    failed_tests=$((failed_tests + 1))
fi

# 3. 에러 케이스 테스트
echo ""
echo "=========================================="
echo "3. 에러 케이스 검증"
echo "=========================================="

total_tests=$((total_tests + 1))
if check_endpoint "$ALB_URL/api/nonexistent/path" "404" "존재하지 않는 경로 (404 예상)"; then
    passed_tests=$((passed_tests + 1))
else
    failed_tests=$((failed_tests + 1))
    echo -e "${YELLOW}참고: ALB 기본 동작에 따라 다른 상태 코드가 반환될 수 있습니다${NC}"
fi

# 4. 상세 응답 확인 (선택사항)
if [ "$VERBOSE" = "true" ]; then
    check_endpoint_verbose "$ALB_URL/api/users/actuator/health" "User Service Health Check"
    check_endpoint_verbose "$ALB_URL/api/boards/health" "Board Service Health Check"
fi

# 결과 요약
echo ""
echo "=========================================="
echo "검증 결과 요약"
echo "=========================================="
echo -e "총 테스트: $total_tests"
echo -e "${GREEN}성공: $passed_tests${NC}"
echo -e "${RED}실패: $failed_tests${NC}"
echo ""

if [ $failed_tests -eq 0 ]; then
    echo -e "${GREEN}✓ 모든 테스트 통과!${NC}"
    echo ""
    echo "다음 단계:"
    echo "1. AWS Console에서 Listener Rules 확인"
    echo "2. Target Group Health 상태 확인"
    echo "3. 실제 API 호출 테스트 (인증 포함)"
    echo ""
    echo "자세한 내용은 docs/ALB_VERIFICATION_GUIDE.md를 참조하세요."
    exit 0
else
    echo -e "${RED}✗ 일부 테스트 실패${NC}"
    echo ""
    echo "문제 해결:"
    echo "1. 서비스가 정상적으로 실행 중인지 확인"
    echo "2. ALB Listener Rules가 올바르게 설정되었는지 확인"
    echo "3. Target Group Health가 healthy 상태인지 확인"
    echo "4. Health check 경로가 올바른지 확인"
    echo ""
    echo "자세한 내용은 docs/ALB_VERIFICATION_GUIDE.md를 참조하세요."
    exit 1
fi

#!/bin/bash

# performance_test_auto.sh
# Board Service 자동 성능 테스트 스크립트 (수정됨)

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

# Configuration
BASE_URL="http://localhost:8000"
USER_SERVICE_URL="http://localhost:8080"

# Test user credentials
TEST_EMAIL="perftest@example.com"
TEST_PASSWORD="password123"
TEST_NAME="Performance Test User"

#성능테스트 - 아래 경우에 따라 수정
# 1. 현재 설정으로 기본 성능 확인
#TEST_REQUESTS=20, CONCURRENT_USERS=3

# 2. 중간 부하로 안정성 확인
#TEST_REQUESTS=100, CONCURRENT_USERS=10

# 3. 높은 부하에서 한계점 찾기
#TEST_REQUESTS=300, CONCURRENT_USERS=30

# 4. Redis 캐시 효과 확인
#TEST_REQUESTS=1000, CONCURRENT_USERS=1  # 단일 사용자, 많은 요청

# Performance test settings
WARMUP_REQUESTS=5
TEST_REQUESTS=500
CONCURRENT_USERS=20

# Global variables
TOKEN=""
PROJECT_ID=""
WORKSPACE_ID=""

echo -e "${BLUE}============================================${NC}"
echo -e "${BLUE} Board Service 자동 성능 테스트${NC}"
echo -e "${BLUE}============================================${NC}"

# Function to print step
print_step() {
    echo -e "\n${YELLOW}📋 $1${NC}"
}

# Function to print success
print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

# Function to print error
print_error() {
    echo -e "${RED}❌ $1${NC}"
    exit 1
}

# Function to make API call
api_call() {
    local method=$1
    local url=$2
    local data=$3
    local headers=$4

    if [ -n "$data" ]; then
        if [ -n "$headers" ]; then
            curl -s -X $method "$url" -H "Content-Type: application/json" -H "$headers" -d "$data"
        else
            curl -s -X $method "$url" -H "Content-Type: application/json" -d "$data"
        fi
    else
        if [ -n "$headers" ]; then
            curl -s -X $method "$url" -H "$headers"
        else
            curl -s -X $method "$url"
        fi
    fi
}

# Setup: Create test environment
setup_test_environment() {
    print_step "🔧 테스트 환경 설정 중..."

    # Health check (수정됨)
    response=$(api_call "GET" "$BASE_URL/health")
    if ! echo "$response" | grep -q '"status":"healthy"'; then
        print_error "Health check failed: $response"
    fi

    # Register/Login user (수정됨)
    register_data="{\"email\":\"$TEST_EMAIL\",\"password\":\"$TEST_PASSWORD\",\"name\":\"$TEST_NAME\"}"
    api_call "POST" "$USER_SERVICE_URL/api/auth/signup" "$register_data" > /dev/null 2>&1 || true

    login_data="{\"email\":\"$TEST_EMAIL\",\"password\":\"$TEST_PASSWORD\"}"
    login_response=$(api_call "POST" "$USER_SERVICE_URL/api/auth/login" "$login_data")

    if echo "$login_response" | grep -q "accessToken"; then
        TOKEN=$(echo "$login_response" | grep -o '"accessToken":"[^"]*"' | cut -d'"' -f4)
        print_success "로그인 성공"
    else
        print_error "로그인 실패: $login_response"
    fi

    # Create workspace (수정됨)
    workspace_data="{\"name\":\"Performance Test Workspace\",\"description\":\"성능 테스트용\"}"
    workspace_response=$(api_call "POST" "$BASE_URL/api/workspaces" "$workspace_data" "Authorization: Bearer $TOKEN")

    if echo "$workspace_response" | grep -q '"data"'; then
        WORKSPACE_ID=$(echo "$workspace_response" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
        print_success "Workspace 생성 완료"
    else
        print_error "Workspace 생성 실패: $workspace_response"
    fi

    # Create project with kanbans (수정됨)
    project_data="{\"workspaceId\":\"$WORKSPACE_ID\",\"name\":\"Performance Test Project\",\"description\":\"성능 테스트용\"}"
    project_response=$(api_call "POST" "$BASE_URL/api/projects" "$project_data" "Authorization: Bearer $TOKEN")

    if echo "$project_response" | grep -q '"data"'; then
        PROJECT_ID=$(echo "$project_response" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
        print_success "Project 생성 완료"
    else
        print_error "Project 생성 실패: $project_response"
    fi

    # Create some test kanbans (수정됨)
    stages_response=$(api_call "GET" "$BASE_URL/api/custom-fields/projects/$PROJECT_ID/stages" "" "Authorization: Bearer $TOKEN")
    STAGE_ID=$(echo "$stages_response" | grep -o '"id":"[^"]*"' | sed -n '2p' | cut -d'"' -f4)

    roles_response=$(api_call "GET" "$BASE_URL/api/custom-fields/projects/$PROJECT_ID/roles" "" "Authorization: Bearer $TOKEN")
    ROLE_ID=$(echo "$roles_response" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)

    importance_response=$(api_call "GET" "$BASE_URL/api/custom-fields/projects/$PROJECT_ID/importance" "" "Authorization: Bearer $TOKEN")
    IMPORTANCE_ID=$(echo "$importance_response" | grep -o '"id":"[^"]*"' | sed -n '2p' | cut -d'"' -f4)

    # Create multiple kanbans for realistic test
    for i in {1..10}; do
        kanban_data="{\"projectId\":\"$PROJECT_ID\",\"title\":\"성능 테스트 칸반 $i\",\"roleIds\":[\"$ROLE_ID\"],\"stageId\":\"$STAGE_ID\",\"importanceId\":\"$IMPORTANCE_ID\"}"
        api_call "POST" "$BASE_URL/api/kanbans" "$kanban_data" "Authorization: Bearer $TOKEN" > /dev/null
    done

    print_success "테스트 데이터 생성 완료 (칸반 10개)"
}

# Performance test function
run_performance_test() {
    local test_name=$1
    local url=$2
    local method=${3:-GET}
    local data=${4:-""}

    print_step "🚀 $test_name 성능 테스트"

    # Warmup
    echo -e "${CYAN}워밍업 중... ($WARMUP_REQUESTS 요청)${NC}"
    for i in $(seq 1 $WARMUP_REQUESTS); do
        if [ -n "$data" ]; then
            curl -s -X $method "$url" -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" -d "$data" > /dev/null
        else
            curl -s -H "Authorization: Bearer $TOKEN" "$url" > /dev/null
        fi
    done

    # Actual test
    echo -e "${CYAN}성능 테스트 실행 중... ($TEST_REQUESTS 요청)${NC}"

    total_time=0
    min_time=999999
    max_time=0
    failed_requests=0

    # Results array
    declare -a response_times

    for i in $(seq 1 $TEST_REQUESTS); do
        start_time=$(date +%s%N)

        if [ -n "$data" ]; then
            response=$(curl -s -w "%{http_code}" -X $method "$url" -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" -d "$data")
        else
            response=$(curl -s -w "%{http_code}" -H "Authorization: Bearer $TOKEN" "$url")
        fi

        end_time=$(date +%s%N)
        duration=$(((end_time - start_time) / 1000000))  # Convert to milliseconds

        # Check if request was successful
        http_code="${response: -3}"
        if [[ "$http_code" =~ ^2[0-9][0-9]$ ]]; then
            response_times[$i]=$duration
            total_time=$((total_time + duration))

            if [ $duration -lt $min_time ]; then
                min_time=$duration
            fi

            if [ $duration -gt $max_time ]; then
                max_time=$duration
            fi

            echo -ne "\rRequest $i/$TEST_REQUESTS: ${duration}ms (HTTP $http_code)"
        else
            failed_requests=$((failed_requests + 1))
            echo -ne "\rRequest $i/$TEST_REQUESTS: FAILED (HTTP $http_code)"
        fi
    done

    echo # New line

    # Calculate statistics
    successful_requests=$((TEST_REQUESTS - failed_requests))
    if [ $successful_requests -gt 0 ]; then
        avg_time=$((total_time / successful_requests))

        # Calculate percentiles (simple approach)
        sorted_times=($(printf '%s\n' "${response_times[@]}" | sort -n))
        p50_index=$(((successful_requests * 50) / 100))
        p95_index=$(((successful_requests * 95) / 100))
        p99_index=$(((successful_requests * 99) / 100))

        p50=${sorted_times[$p50_index]:-0}
        p95=${sorted_times[$p95_index]:-0}
        p99=${sorted_times[$p99_index]:-0}

        # Calculate throughput
        throughput=$(echo "scale=2; $successful_requests / ($total_time / 1000)" | bc -l 2>/dev/null || echo "N/A")

        # Print results
        echo -e "\n${GREEN}📊 $test_name 성능 결과:${NC}"
        echo -e "   성공한 요청: $successful_requests/$TEST_REQUESTS"
        echo -e "   실패한 요청: $failed_requests"
        echo -e "   평균 응답시간: ${avg_time}ms"
        echo -e "   최소 응답시간: ${min_time}ms"
        echo -e "   최대 응답시간: ${max_time}ms"
        echo -e "   50th percentile: ${p50}ms"
        echo -e "   95th percentile: ${p95}ms"
        echo -e "   99th percentile: ${p99}ms"
        if [ "$throughput" != "N/A" ]; then
            echo -e "   처리량: ${throughput} req/sec"
        fi

        # Performance evaluation
        if [ $avg_time -lt 100 ]; then
            echo -e "   ${GREEN}✅ 성능 평가: 우수${NC}"
        elif [ $avg_time -lt 500 ]; then
            echo -e "   ${YELLOW}⚠️  성능 평가: 보통${NC}"
        else
            echo -e "   ${RED}❌ 성능 평가: 개선 필요${NC}"
        fi
    else
        echo -e "${RED}❌ 모든 요청이 실패했습니다.${NC}"
    fi
}

# Concurrent test function
run_concurrent_test() {
    local test_name=$1
    local url=$2

    print_step "🔄 $test_name 동시성 테스트 ($CONCURRENT_USERS명 동시 사용자)"

    # Create background processes
    pids=()

    start_time=$(date +%s)

    for user in $(seq 1 $CONCURRENT_USERS); do
        {
            for req in $(seq 1 $((TEST_REQUESTS / CONCURRENT_USERS))); do
                curl -s -H "Authorization: Bearer $TOKEN" "$url" > /dev/null
            done
        } &
        pids+=($!)
    done

    # Wait for all background processes
    for pid in "${pids[@]}"; do
        wait $pid
    done

    end_time=$(date +%s)
    total_duration=$((end_time - start_time))

    echo -e "${GREEN}📊 동시성 테스트 결과:${NC}"
    echo -e "   동시 사용자: $CONCURRENT_USERS명"
    echo -e "   총 요청: $TEST_REQUESTS개"
    echo -e "   총 소요시간: ${total_duration}초"
    echo -e "   평균 처리량: $(echo "scale=2; $TEST_REQUESTS / $total_duration" | bc -l) req/sec"
}

# Cleanup function
cleanup() {
    print_step "🧹 테스트 데이터 정리 중..."

    if [ -n "$WORKSPACE_ID" ] && [ -n "$TOKEN" ]; then
        api_call "DELETE" "$BASE_URL/api/workspaces/$WORKSPACE_ID" "" "Authorization: Bearer $TOKEN" > /dev/null 2>&1 || true
        print_success "테스트 데이터 정리 완료"
    fi
}

# Main execution
main() {
    # Setup
    setup_test_environment

    echo -e "\n${BLUE}============================================${NC}"
    echo -e "${BLUE} 성능 테스트 시작${NC}"
    echo -e "${BLUE}============================================${NC}"

    # Test different endpoints
    run_performance_test "Stage-based Board View" "$BASE_URL/api/projects/$PROJECT_ID/orders/stage-board"

    run_performance_test "Role-based Board View" "$BASE_URL/api/projects/$PROJECT_ID/orders/role-board"

    run_performance_test "Project 상세 조회" "$BASE_URL/api/projects/$PROJECT_ID"

    run_performance_test "칸반 목록 조회" "$BASE_URL/api/kanbans?projectId=$PROJECT_ID"

    run_performance_test "커스텀 필드 조회" "$BASE_URL/api/custom-fields/projects/$PROJECT_ID/stages"

    # Concurrent tests
    run_concurrent_test "Stage Board 동시 접근" "$BASE_URL/api/projects/$PROJECT_ID/orders/stage-board"

    # Summary
    echo -e "\n${BLUE}============================================${NC}"
    echo -e "${BLUE} 성능 테스트 완료${NC}"
    echo -e "${BLUE}============================================${NC}"
    print_success "모든 성능 테스트가 완료되었습니다!"

    # Cleanup
    cleanup
}

# Trap to ensure cleanup on exit
trap cleanup EXIT

# Run main function
main

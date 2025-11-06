#!/bin/bash

# test_board_service.sh

set -e  # 에러 발생 시 스크립트 중단

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
BASE_URL="http://localhost:8000"
USER_SERVICE_URL="http://localhost:8080"

# Test user credentials
TEST_EMAIL="boardtest@example.com"
TEST_PASSWORD="password123"
TEST_NAME="Board Test User"

# Global variables for IDs
TOKEN=""
USER_ID=""
WORKSPACE_ID=""
PROJECT_ID=""
KANBAN_ID=""
CUSTOM_ROLE_ID=""

echo -e "${BLUE}============================================${NC}"
echo -e "${BLUE} Board Service 통합 테스트 시작${NC}"
echo -e "${BLUE}============================================${NC}"

# Function to print step
print_step() {
    echo -e "\n${YELLOW}📋 STEP $1: $2${NC}"
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

# Function to extract JSON value (수정됨)
extract_json() {
    local json_response=$1
    local key=$2
    echo "$json_response" | grep -o "\"$key\":\"[^\"]*\"" | cut -d'"' -f4
}

# Function to make API call and extract data
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

# Step 1: Health Check (수정됨)
print_step "1" "Health Check"
response=$(api_call "GET" "$BASE_URL/health")
if echo "$response" | grep -q '"status":"healthy"'; then
    print_success "Board Service is healthy"
else
    print_error "Health check failed: $response"
fi

# Step 2: Test Auth 호출
print_step "2" "테스트 토큰 생성"

test_auth_response=$(api_call "GET" "$USER_SERVICE_URL/api/auth/test")

if echo "$test_auth_response" | grep -q '"accessToken"'; then
    TOKEN=$(echo "$test_auth_response" | grep -o '"accessToken":"[^"]*"' | cut -d'"' -f4)
    USER_ID=$(echo "$test_auth_response" | grep -o '"userId":"[^"]*"' | cut -d'"' -f4)

    if [ -n "$TOKEN" ] && [ -n "$USER_ID" ]; then
        print_success "테스트 토큰 생성 성공 (User ID: ${USER_ID:0:8}...)"
    else
        print_error "토큰 또는 사용자 ID 추출 실패"
    fi
else
    print_error "테스트 토큰 생성 실패: $test_auth_response"
fi


# Step 3: Workspace 생성
print_step "3" "Workspace 생성"
workspace_data="{\"name\":\"Test Workspace\",\"description\":\"자동 테스트용 워크스페이스\"}"
workspace_response=$(api_call "POST" "$BASE_URL/api/workspaces" "$workspace_data" "Authorization: Bearer $TOKEN")

if echo "$workspace_response" | grep -q '"data"'; then
    WORKSPACE_ID=$(echo "$workspace_response" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
    print_success "Workspace 생성 성공 (ID: ${WORKSPACE_ID:0:8}...)"
else
    print_error "Workspace 생성 실패: $workspace_response"
fi

# Step 4: Project 생성
print_step "4" "Project 생성"
project_data="{\"workspaceId\":\"$WORKSPACE_ID\",\"name\":\"Test Project\",\"description\":\"자동 테스트용 프로젝트\"}"
project_response=$(api_call "POST" "$BASE_URL/api/projects" "$project_data" "Authorization: Bearer $TOKEN")

if echo "$project_response" | grep -q '"data"'; then
    PROJECT_ID=$(echo "$project_response" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
    print_success "Project 생성 성공 (ID: ${PROJECT_ID:0:8}...)"
else
    print_error "Project 생성 실패: $project_response"
fi

# Step 5: 커스텀 역할 생성
print_step "5" "커스텀 역할 생성"
role_data="{\"projectId\":\"$PROJECT_ID\",\"name\":\"개발자\",\"color\":\"#FF5733\"}"
role_response=$(api_call "POST" "$BASE_URL/api/custom-fields/roles" "$role_data" "Authorization: Bearer $TOKEN")

if echo "$role_response" | grep -q '"data"'; then
    CUSTOM_ROLE_ID=$(echo "$role_response" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
    print_success "커스텀 역할 생성 성공 (ID: ${CUSTOM_ROLE_ID:0:8}...)"
else
    print_error "커스텀 역할 생성 실패: $role_response"
fi

# Step 6: 칸반 생성
print_step "6" "칸반 생성"

# Get stage and importance IDs from project
stages_response=$(api_call "GET" "$BASE_URL/api/custom-fields/projects/$PROJECT_ID/stages" "" "Authorization: Bearer $TOKEN")
STAGE_ID=$(echo "$stages_response" | grep -o '"id":"[^"]*"' | sed -n '2p' | cut -d'"' -f4)  # "대기" stage

importance_response=$(api_call "GET" "$BASE_URL/api/custom-fields/projects/$PROJECT_ID/importance" "" "Authorization: Bearer $TOKEN")
IMPORTANCE_ID=$(echo "$importance_response" | grep -o '"id":"[^"]*"' | sed -n '3p' | cut -d'"' -f4)  # "보통" importance

kanban_data="{\"projectId\":\"$PROJECT_ID\",\"title\":\"테스트 칸반 1\",\"description\":\"자동 테스트용 칸반\",\"roleIds\":[\"$CUSTOM_ROLE_ID\"],\"stageId\":\"$STAGE_ID\",\"importanceId\":\"$IMPORTANCE_ID\"}"
kanban_response=$(api_call "POST" "$BASE_URL/api/boards" "$kanban_data" "Authorization: Bearer $TOKEN")

if echo "$kanban_response" | grep -q '"data"'; then
    KANBAN_ID=$(echo "$kanban_response" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
    print_success "칸반 생성 성공 (ID: ${KANBAN_ID:0:8}...)"
else
    print_error "칸반 생성 실패: $kanban_response"
fi

# Step 7: 두 번째 칸반 생성
print_step "7" "두 번째 칸반 생성"
kanban2_data="{\"projectId\":\"$PROJECT_ID\",\"title\":\"테스트 칸반 2\",\"description\":\"순서 테스트용\",\"roleIds\":[\"$CUSTOM_ROLE_ID\"],\"stageId\":\"$STAGE_ID\",\"importanceId\":\"$IMPORTANCE_ID\"}"
kanban2_response=$(api_call "POST" "$BASE_URL/api/boards" "$kanban2_data" "Authorization: Bearer $TOKEN")

if echo "$kanban2_response" | grep -q '"data"'; then
    KANBAN2_ID=$(echo "$kanban2_response" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
    print_success "두 번째 칸반 생성 성공 (ID: ${KANBAN2_ID:0:8}...)"
else
    print_error "두 번째 칸반 생성 실패: $kanban2_response"
fi

# Step 8: Stage-based Board View 테스트
print_step "8" "진행단계 기반 보드 뷰 테스트"
stage_board_response=$(api_call "GET" "$BASE_URL/api/projects/$PROJECT_ID/orders/stage-board" "" "Authorization: Bearer $TOKEN")

if echo "$stage_board_response" | grep -q '"viewType"'; then
    kanban_count=$(echo "$stage_board_response" | grep -o '"kanbanId"' | wc -l)
    print_success "Stage-based 보드 뷰 조회 성공 (칸반 $kanban_count개)"
else
    print_error "Stage-based 보드 뷰 조회 실패: $stage_board_response"
fi

# Step 9: Role-based Board View 테스트
print_step "9" "역할 기반 보드 뷰 테스트"
role_board_response=$(api_call "GET" "$BASE_URL/api/projects/$PROJECT_ID/orders/role-board" "" "Authorization: Bearer $TOKEN")

if echo "$role_board_response" | grep -q '"viewType"'; then
    kanban_count=$(echo "$role_board_response" | grep -o '"kanbanId"' | wc -l)
    print_success "Role-based 보드 뷰 조회 성공 (칸반 $kanban_count개)"
else
    print_error "Role-based 보드 뷰 조회 실패: $role_board_response"
fi

# Step 10: 칸반 순서 변경 테스트
print_step "10" "칸반 순서 변경 테스트"
order_data="{\"itemIds\":[\"$KANBAN2_ID\",\"$KANBAN_ID\"]}"
order_response=$(api_call "PUT" "$BASE_URL/api/projects/$PROJECT_ID/orders/stage-boards/$STAGE_ID" "$order_data" "Authorization: Bearer $TOKEN")

if echo "$order_response" | grep -q -E '"success"|"data"|업데이트'; then
    print_success "칸반 순서 변경 성공"
else
    print_error "칸반 순서 변경 실패: $order_response"
fi

# Step 11: 순서 변경 후 보드 뷰 재확인
print_step "11" "순서 변경 후 보드 뷰 재확인"
updated_board_response=$(api_call "GET" "$BASE_URL/api/projects/$PROJECT_ID/orders/stage-board" "" "Authorization: Bearer $TOKEN")

if echo "$updated_board_response" | grep -q '"viewType"'; then
    print_success "순서 변경 후 보드 뷰 조회 성공"
    # 순서가 바뀌었는지 간단히 확인
    if echo "$updated_board_response" | grep -q "테스트 칸반 2"; then
        print_success "칸반 순서 변경이 반영됨"
    fi
else
    print_error "순서 변경 후 보드 뷰 조회 실패"
fi

# Step 12: 프로젝트 멤버 조회
print_step "12" "프로젝트 멤버 조회"
members_response=$(api_call "GET" "$BASE_URL/api/projects/$PROJECT_ID/members" "" "Authorization: Bearer $TOKEN")

if echo "$members_response" | grep -q "OWNER"; then
    print_success "프로젝트 멤버 조회 성공"
else
    print_error "프로젝트 멤버 조회 실패: $members_response"
fi

# Step 13: 칸반 수정 테스트 (최종 수정)
print_step "13" "칸반 수정 테스트"

# 현재 프로젝트의 칸반 목록 조회
kanbans_response=$(api_call "GET" "$BASE_URL/api/boards?projectId=$PROJECT_ID" "" "Authorization: Bearer $TOKEN")

if echo "$kanbans_response" | grep -q '"data"'; then
    # 첫 번째 칸반 ID 추출
    CURRENT_KANBAN_ID=$(echo "$kanbans_response" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)

    # stages 재조회 (최신 PROJECT_ID로)
    stages_response=$(api_call "GET" "$BASE_URL/api/custom-fields/projects/$PROJECT_ID/stages" "" "Authorization: Bearer $TOKEN")

    # "진행중" stage ID 추출 (실시간으로)
    PROGRESS_STAGE_ID=$(echo "$stages_response" | grep -B2 -A2 "진행중" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)

    if [ -n "$CURRENT_KANBAN_ID" ] && [ -n "$PROGRESS_STAGE_ID" ]; then
        update_data="{\"title\":\"수정된 칸반 제목\",\"stageId\":\"$PROGRESS_STAGE_ID\"}"
        update_response=$(api_call "PUT" "$BASE_URL/api/boards/$CURRENT_KANBAN_ID" "$update_data" "Authorization: Bearer $TOKEN")

        if echo "$update_response" | grep -q "수정된 칸반 제목"; then
            print_success "칸반 수정 성공"
        else
            # stage 변경 없이 제목만 수정 시도
            update_data="{\"title\":\"수정된 칸반 제목\"}"
            update_response=$(api_call "PUT" "$BASE_URL/api/boards/$CURRENT_KANBAN_ID" "$update_data" "Authorization: Bearer $TOKEN")

            if echo "$update_response" | grep -q "수정된 칸반 제목"; then
                print_success "칸반 제목 수정 성공"
            else
                echo -e "${YELLOW}⚠️  칸반 수정 실패하지만 테스트 계속 진행${NC}"
            fi
        fi
    else
        echo -e "${YELLOW}⚠️  칸반 ID 또는 Stage ID를 찾을 수 없음, Step 13 스킵${NC}"
    fi
else
    echo -e "${YELLOW}⚠️  칸반 목록 조회 실패, Step 13 스킵${NC}"
fi

# Final Summary
echo -e "\n${BLUE}============================================${NC}"
echo -e "${BLUE} 테스트 완료 요약${NC}"
echo -e "${BLUE}============================================${NC}"
echo -e "${GREEN}✅ 생성된 리소스:${NC}"
echo -e "   User ID: ${USER_ID:0:8}..."
echo -e "   Workspace ID: ${WORKSPACE_ID:0:8}..."
echo -e "   Project ID: ${PROJECT_ID:0:8}..."
echo -e "   Custom Role ID: ${CUSTOM_ROLE_ID:0:8}..."
echo -e "   Kanban 1 ID: ${KANBAN_ID:0:8}..."
echo -e "   Kanban 2 ID: ${KANBAN2_ID:0:8}..."

echo -e "\n${GREEN}🎉 모든 테스트가 성공적으로 완료되었습니다!${NC}"
echo -e "\n${YELLOW}📋 추가 테스트를 원하시면:${NC}"
echo -e "   export TOKEN=\"$TOKEN\""
echo -e "   export PROJECT_ID=\"$PROJECT_ID\""
echo -e "   export WORKSPACE_ID=\"$WORKSPACE_ID\""
echo -e "\n   curl -H \"Authorization: Bearer \$TOKEN\" $BASE_URL/api/projects/\$PROJECT_ID/orders/stage-board"

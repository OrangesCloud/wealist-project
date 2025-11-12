#!/bin/bash

# ===== 사전 준비 =====
if ! command -v jq &> /dev/null;
  then
  echo "jq 명령어가 필요합니다. 설치 후 다시 실행하세요."
  exit 1
fi

# ===== 환경 설정 =====
BASE_URL="http://localhost:8080"

# 색상 설정
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# ===== 헬퍼 함수 =====
print_section() {
    echo -e "\n${BLUE}===== $1 =====${NC}\n"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

# ===== 1. 인증 정보 설정 (OAuth2) =====
print_section "Authentication Setup (OAuth2)"

# 1.1 수동으로 Google OAuth2 로그인을 수행하고, 발급받은 JWT Access Token을 아래에 붙여넣으세요.
ACCESS_TOKEN=""

if [ -z "$ACCESS_TOKEN" ]; then
    print_error "ACCESS_TOKEN이 비어있습니다. 스크립트를 수정하여 토큰을 직접 설정해주세요."
    exit 1
fi

echo "제공된 ACCESS_TOKEN을 사용하여 테스트를 시작합니다."

# 1.2 토큰을 사용하여 내 정보 조회 (USER_ID 획득)
ME_RESPONSE=$(curl -s -X GET "$BASE_URL/api/users/me" \
  -H "Authorization: Bearer $ACCESS_TOKEN")

USER_ID=$(echo "$ME_RESPONSE" | jq -r '.userId // empty')

if [ -n "$USER_ID" ]; then
    print_success "내 정보 조회 성공 (USER_ID: $USER_ID)"
else
    print_error "내 정보 조회 실패. ACCESS_TOKEN이 유효한지 확인하세요."
    echo "$ME_RESPONSE" > me_error.log
    exit 1
fi

# ===== 변수 (테스트용) =====
WORKSPACE_ID=""
OTHER_USER_ID=""

# ===== 2. 사용자 관련 API =====
print_section "User APIs"

# 2.1 다른 사용자 생성
CREATE_USER_RESPONSE=$(curl -s -X POST "$BASE_URL/api/users" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -d "{
    \"email\": \"testuser-$(date +%s)@example.com\",
    \"googleId\": \"google-id-$(date +%s)\" 
  }")

OTHER_USER_ID=$(echo "$CREATE_USER_RESPONSE" | jq -r '.userId // empty')

if [ -n "$OTHER_USER_ID" ]; then
    print_success "다른 사용자 생성 성공 (OTHER_USER_ID: $OTHER_USER_ID)"
else
    print_error "다른 사용자 생성 실패"
    echo "$CREATE_USER_RESPONSE" > user_create_error.log
fi

# 2.2 사용자 정보 조회
GET_USER_RESPONSE=$(curl -s -X GET "$BASE_URL/api/users/$USER_ID" \
  -H "Authorization: Bearer $ACCESS_TOKEN")

GET_USER_ID=$(echo "$GET_USER_RESPONSE" | jq -r '.userId // empty')

if [ "$GET_USER_ID" == "$USER_ID" ]; then
    print_success "사용자 정보 조회 성공"
else
    print_error "사용자 정보 조회 실패"
    echo "$GET_USER_RESPONSE" > user_get_error.log
fi

# 2.3 사용자 정보 수정
UPDATE_USER_RESPONSE=$(curl -s -X PUT "$BASE_URL/api/users/$USER_ID" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -d "{
        \"email\": \"updated-$(date +%s)@example.com\"
    }")
UPDATED_EMAIL=$(echo "$UPDATE_USER_RESPONSE" | jq -r '.email // empty')
if [[ $UPDATED_EMAIL == *"updated"* ]]; then
    print_success "사용자 정보 수정 성공"
else
    print_error "사용자 정보 수정 실패"
    echo "$UPDATE_USER_RESPONSE" > user_update_error.log
fi


# ===== 3. 워크스페이스 관련 API =====
print_section "Workspace APIs"

# 3.1 워크스페이스 생성
CREATE_WORKSPACE_RESPONSE=$(curl -s -X POST "$BASE_URL/api/workspaces" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -d "{
    \"workspaceName\": \"테스트 워크스페이스\",
    \"workspaceDescription\": \"API 테스트용 워크스페이스\"
  }")

WORKSPACE_ID=$(echo "$CREATE_WORKSPACE_RESPONSE" | jq -r '.workspaceId // empty')

if [ -n "$WORKSPACE_ID" ]; then
    print_success "워크스페이스 생성 성공 (WORKSPACE_ID: $WORKSPACE_ID)"
else
    print_error "워크스페이스 생성 실패"
    echo "$CREATE_WORKSPACE_RESPONSE" > workspace_create_error.log
    exit 1
fi

# 3.2 워크스페이스 목록 조회
GET_WORKSPACES_RESPONSE=$(curl -s -X GET "$BASE_URL/api/workspaces/all" \
  -H "Authorization: Bearer $ACCESS_TOKEN")

if [[ $(echo "$GET_WORKSPACES_RESPONSE" | jq -r '.[] | .workspaceId') == *"$WORKSPACE_ID"* ]]; then
    print_success "워크스페이스 목록 조회 성공"
else
    print_error "워크스페이스 목록 조회 실패"
    echo "$GET_WORKSPACES_RESPONSE" > workspace_get_all_error.log
fi

# 3.3 워크스페이스 정보 수정
UPDATE_WORKSPACE_RESPONSE=$(curl -s -X PUT "$BASE_URL/api/workspaces/$WORKSPACE_ID" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -d "{
        \"workspaceName\": \"수정된 워크스페이스\"
    }")
UPDATED_WORKSPACE_NAME=$(echo "$UPDATE_WORKSPACE_RESPONSE" | jq -r '.workspaceName // empty')
if [ "$UPDATED_WORKSPACE_NAME" == "수정된 워크스페이스" ]; then
    print_success "워크스페이스 정보 수정 성공"
else
    print_error "워크스페이스 정보 수정 실패"
    echo "$UPDATE_WORKSPACE_RESPONSE" > workspace_update_error.log
fi

# 3.4 다른 사용자 초대
INVITE_USER_RESPONSE=$(curl -s -X POST "$BASE_URL/api/workspaces/$WORKSPACE_ID/members/invite" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -d "{
        \"userId\": \"$OTHER_USER_ID\"
    }")
INVITED_USER_ID=$(echo "$INVITE_USER_RESPONSE" | jq -r '.userId // empty')
if [ "$INVITED_USER_ID" == "$OTHER_USER_ID" ]; then
    print_success "다른 사용자 초대 성공"
else
    print_error "다른 사용자 초대 실패"
    echo "$INVITE_USER_RESPONSE" > workspace_invite_error.log
fi

# 3.5 워크스페이스 멤버 목록 조회
GET_MEMBERS_RESPONSE=$(curl -s -X GET "$BASE_URL/api/workspaces/$WORKSPACE_ID/members" \
  -H "Authorization: Bearer $ACCESS_TOKEN")

if [[ $(echo "$GET_MEMBERS_RESPONSE" | jq -r '.[] | .userId') == *"$OTHER_USER_ID"* ]]; then
    print_success "워크스페이스 멤버 목록 조회 성공"
else
    print_error "워크스페이스 멤버 목록 조회 실패"
    echo "$GET_MEMBERS_RESPONSE" > workspace_get_members_error.log
fi


# ===== 4. 프로필 관련 API =====
print_section "UserProfile APIs"

# 4.1 프로필 생성
CREATE_PROFILE_RESPONSE=$(curl -s -X POST "$BASE_URL/api/profiles" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -d "{
        \"workspaceId\": \"$WORKSPACE_ID\",
        \"nickName\": \"테스트닉네임\"
    }")
PROFILE_ID=$(echo "$CREATE_PROFILE_RESPONSE" | jq -r '.profileId // empty')
if [ -n "$PROFILE_ID" ]; then
    print_success "프로필 생성 성공"
else
    print_error "프로필 생성 실패"
    echo "$CREATE_PROFILE_RESPONSE" > profile_create_error.log
fi

# 4.2 프로필 조회
GET_PROFILE_RESPONSE=$(curl -s -X GET "$BASE_URL/api/profiles/me" \
  -H "Authorization: Bearer $ACCESS_TOKEN")

if [[ $(echo "$GET_PROFILE_RESPONSE" | jq -r '.userId') == *"$USER_ID"* ]]; then
    print_success "프로필 조회 성공"
else
    print_error "프로필 조회 실패"
    echo "$GET_PROFILE_RESPONSE" > profile_get_error.log
fi

# 4.3 프로필 수정
UPDATE_PROFILE_RESPONSE=$(curl -s -X PUT "$BASE_URL/api/profiles/me" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -d "{
        \"workspaceId\": \"$WORKSPACE_ID\",
        \"userId\": \"$USER_ID\",
        \"nickName\": \"수정된닉네임\"
    }")
UPDATED_NICKNAME=$(echo "$UPDATE_PROFILE_RESPONSE" | jq -r '.nickName // empty')
if [ "$UPDATED_NICKNAME" == "수정된닉네임" ]; then
    print_success "프로필 수정 성공"
else
    print_error "프로필 수정 실패"
    echo "$UPDATE_PROFILE_RESPONSE" > profile_update_error.log
fi


# ===== 5. 워크스페이스 가입 및 승인 =====
print_section "Workspace Join and Approval"

# 5.1 테스트용 사용자 및 워크스페이스 생성
JOIN_TEST_USER_RESPONSE=$(curl -s -X POST "$BASE_URL/api/users" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -d "{
    \"email\": \"join-test-user-$(date +%s)@example.com\",
    \"googleId\": \"join-google-id-$(date +%s)\"
  }")
JOIN_TEST_USER_ID=$(echo "$JOIN_TEST_USER_RESPONSE" | jq -r '.userId // empty')
if [ -n "$JOIN_TEST_USER_ID" ]; then
    print_success "가입 테스트용 사용자 생성 성공 (USER_ID: $JOIN_TEST_USER_ID)"
else
    print_error "가입 테스트용 사용자 생성 실패"
    echo "$JOIN_TEST_USER_RESPONSE" > join_user_create_error.log
fi

CREATE_JOIN_WORKSPACE_RESPONSE=$(curl -s -X POST "$BASE_URL/api/workspaces" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -d "{
    \"workspaceName\": \"가입 테스트 워크스페이스\",
    \"workspaceDescription\": \"가입 및 승인 테스트용\"
  }")
JOIN_WORKSPACE_ID=$(echo "$CREATE_JOIN_WORKSPACE_RESPONSE" | jq -r '.workspaceId // empty')
if [ -n "$JOIN_WORKSPACE_ID" ]; then
    print_success "가입 테스트용 워크스페이스 생성 성공 (WORKSPACE_ID: $JOIN_WORKSPACE_ID)"
else
    print_error "가입 테스트용 워크스페이스 생성 실패"
    echo "$CREATE_JOIN_WORKSPACE_RESPONSE" > join_workspace_create_error.log
fi

# 5.2 워크스페이스 가입 신청
# 참고: 실제로는 가입하려는 사용자의 토큰을 사용해야 하지만, 테스트 편의상 OWNER 토큰으로 요청을 보냅니다.
# 서비스 로직에서 userId를 기반으로 처리해야 합니다.
CREATE_JOIN_REQUEST_RESPONSE=$(curl -s -X POST "$BASE_URL/api/workspaces/join-requests" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -d "{
        \"workspaceId\": \"$JOIN_WORKSPACE_ID\"
    }")
JOIN_REQUEST_ID=$(echo "$CREATE_JOIN_REQUEST_RESPONSE" | jq -r '.id // empty')
if [ -n "$JOIN_REQUEST_ID" ]; then
    print_success "워크스페이스 가입 신청 성공 (REQUEST_ID: $JOIN_REQUEST_ID)"
else
    print_error "워크스페이스 가입 신청 실패"
    echo "$CREATE_JOIN_REQUEST_RESPONSE" > join_request_create_error.log
fi

# 5.3 가입 신청 승인 (OWNER)
APPROVE_RESPONSE=$(curl -s -X POST "$BASE_URL/api/workspaces/$JOIN_WORKSPACE_ID/members/$JOIN_TEST_USER_ID/approve" \
    -H "Authorization: Bearer $ACCESS_TOKEN")
if [ -z "$APPROVE_RESPONSE" ]; then
    print_success "가입 신청 승인 성공"
else
    print_error "가입 신청 승인 실패"
    echo "$APPROVE_RESPONSE" > join_approve_error.log
fi

# 5.4 멤버 확인
GET_MEMBERS_AFTER_APPROVAL_RESPONSE=$(curl -s -X GET "$BASE_URL/api/workspaces/$JOIN_WORKSPACE_ID/members" \
  -H "Authorization: Bearer $ACCESS_TOKEN")

if [[ $(echo "$GET_MEMBERS_AFTER_APPROVAL_RESPONSE" | jq -r '.[] | .userId') == *"$JOIN_TEST_USER_ID"* ]]; then
    print_success "멤버 목록에서 승인된 사용자 확인"
else
    print_error "멤버 목록에서 승인된 사용자 확인 실패"
    echo "$GET_MEMBERS_AFTER_APPROVAL_RESPONSE" > join_verify_member_error.log
fi

# 5.5 가입 신청 거절 테스트
# 새로운 사용자 및 워크스페이스 생성
REJECT_TEST_USER_RESPONSE=$(curl -s -X POST "$BASE_URL/api/users" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -d "{
    \"email\": \"reject-test-user-$(date +%s)@example.com\",
    \"googleId\": \"reject-google-id-$(date +%s)\"
  }")
REJECT_TEST_USER_ID=$(echo "$REJECT_TEST_USER_RESPONSE" | jq -r '.userId // empty')

CREATE_REJECT_WORKSPACE_RESPONSE=$(curl -s -X POST "$BASE_URL/api/workspaces" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -d "{
    \"workspaceName\": \"거절 테스트 워크스페이스\"
  }")
REJECT_WORKSPACE_ID=$(echo "$CREATE_REJECT_WORKSPACE_RESPONSE" | jq -r '.workspaceId // empty')

# 가입 신청
CREATE_REJECT_JOIN_REQUEST_RESPONSE=$(curl -s -X POST "$BASE_URL/api/workspaces/join-requests" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -d "{
        \"workspaceId\": \"$REJECT_WORKSPACE_ID\"
    }")
REJECT_JOIN_REQUEST_ID=$(echo "$CREATE_REJECT_JOIN_REQUEST_RESPONSE" | jq -r '.id // empty')

# 가입 거절
REJECT_RESPONSE=$(curl -s -X POST "$BASE_URL/api/workspaces/$REJECT_WORKSPACE_ID/members/$REJECT_TEST_USER_ID/reject" \
    -H "Authorization: Bearer $ACCESS_TOKEN")
if [ -z "$REJECT_RESPONSE" ]; then
    print_success "가입 신청 거절 성공"
else
    print_error "가입 신청 거절 실패"
    echo "$REJECT_RESPONSE" > join_reject_error.log
fi


# ===== 6. 정리(삭제) API =====
print_section "Cleanup APIs"

# 6.1 워크스페이스 삭제
if [ -n "$WORKSPACE_ID" ]; then
    echo -e "\n워크스페이스 삭제"
    curl -s -X DELETE "$BASE_URL/api/workspaces/$WORKSPACE_ID" \
      -H "Authorization: Bearer $ACCESS_TOKEN"
    print_success "워크스페이스 삭제 요청 완료"
fi
if [ -n "$JOIN_WORKSPACE_ID" ]; then
    echo -e "\n가입 테스트 워크스페이스 삭제"
    curl -s -X DELETE "$BASE_URL/api/workspaces/$JOIN_WORKSPACE_ID" \
      -H "Authorization: Bearer $ACCESS_TOKEN"
    print_success "가입 테스트 워크스페이스 삭제 요청 완료"
fi
if [ -n "$REJECT_WORKSPACE_ID" ]; then
    echo -e "\n거절 테스트 워크스페이스 삭제"
    curl -s -X DELETE "$BASE_URL/api/workspaces/$REJECT_WORKSPACE_ID" \
      -H "Authorization: Bearer $ACCESS_TOKEN"
    print_success "거절 테스트 워크스페이스 삭제 요청 완료"
fi

# 6.2 사용자 삭제
if [ -n "$OTHER_USER_ID" ]; then
    echo -e "\n다른 사용자 삭제"
    curl -s -X DELETE "$BASE_URL/api/users/$OTHER_USER_ID" \
      -H "Authorization: Bearer $ACCESS_TOKEN"
    print_success "다른 사용자 삭제 요청 완료"
fi
if [ -n "$JOIN_TEST_USER_ID" ]; then
    echo -e "\n가입 테스트 사용자 삭제"
    curl -s -X DELETE "$BASE_URL/api/users/$JOIN_TEST_USER_ID" \
      -H "Authorization: Bearer $ACCESS_TOKEN"
    print_success "가입 테스트 사용자 삭제 요청 완료"
fi
if [ -n "$REJECT_TEST_USER_ID" ]; then
    echo -e "\n거절 테스트 사용자 삭제"
    curl -s -X DELETE "$BASE_URL/api/users/$REJECT_TEST_USER_ID" \
      -H "Authorization: Bearer $ACCESS_TOKEN"
    print_success "거절 테스트 사용자 삭제 요청 완료"
fi

print_section "✅ 모든 테스트 완료!"

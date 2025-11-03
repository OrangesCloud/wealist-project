#!/bin/bash

# ===== 사전 준비 =====
if ! command -v jq &> /dev/null; then
  echo "jq 명령어가 필요합니다. 설치 후 다시 실행하세요."
  exit 1
fi

# ===== 환경 설정 =====
BASE_URL="http://localhost:8080"
# BASE_URL="https://api.orangecloud.com"  # 프로덕션 서버

# 색상 설정
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# ===== 변수 =====
ACCESS_TOKEN=""
REFRESH_TOKEN=""
USER_ID=""
GROUP_ID=""
TEAM_ID=""
EMAIL="test-$(date +%s)@example.com"
PASSWORD="password123"
COMPANY_NAME="test-company-$(date +%s%N)"

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

# ===== 1. 인증 관련 API =====
print_section "Authentication APIs"

# 1.1 회원가입
echo "1.1 회원가입 테스트"
SIGNUP_RESPONSE=$(curl -s -X POST "$BASE_URL/api/auth/signup" \
  -H "Content-Type: application/json" \
  -d "{
    \"name\": \"홍길동\",
    \"email\": \"$EMAIL\",
    \"password\": \"$PASSWORD\" 
  }")

USER_ID_SIGNUP=$(echo "$SIGNUP_RESPONSE" | jq -r '.userId // empty')

if [ -n "$USER_ID_SIGNUP" ]; then
    print_success "회원가입 성공 (USER_ID: $USER_ID_SIGNUP)"
else
    print_error "회원가입 실패"
    echo "$SIGNUP_RESPONSE" > signup_error.log
    exit 1
fi

# 1.2 로그인
echo -e "\n1.2 로그인 테스트"
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/api/auth/login" \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"$EMAIL\",
    \"password\": \"$PASSWORD\" 
  }")

ACCESS_TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.accessToken // empty')
REFRESH_TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.refreshToken // empty')
USER_ID=$(echo "$LOGIN_RESPONSE" | jq -r '.userId // empty')

if [ -n "$ACCESS_TOKEN" ] && [ -n "$USER_ID" ]; then
    print_success "로그인 성공 및 토큰/ID 추출 완료"
    echo "ACCESS_TOKEN: $ACCESS_TOKEN"
    echo "USER_ID: $USER_ID"
else
    print_error "로그인 실패 또는 토큰 추출 실패"
    echo "$LOGIN_RESPONSE" > login_error.log
    exit 1
fi

# 1.3 토큰 재발급
echo -e "\n1.3 토큰 재발급 테스트"
REFRESH_RESPONSE=$(curl -s -X POST "$BASE_URL/api/auth/refresh" \
  -H "Content-Type: application/json" \
  -d "{
    \"refreshToken\": \"$REFRESH_TOKEN\" 
  }")

ACCESS_TOKEN=$(echo "$REFRESH_RESPONSE" | jq -r '.accessToken // empty')

if [ -n "$ACCESS_TOKEN" ]; then
    print_success "토큰 재발급 성공"
else
    print_error "토큰 재발급 실패"
fi

# ===== 2. 그룹 관련 API =====
print_section "Group APIs (Prerequisite)"

# 2.1 그룹 생성
CREATE_GROUP_RESPONSE=$(curl -s -X POST "$BASE_URL/api/groups" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -d "{
    \"name\": \"테스트 그룹\",
    \"companyName\": \"$COMPANY_NAME\" 
  }")

GROUP_ID=$(echo "$CREATE_GROUP_RESPONSE" | jq -r '.data.groupId // empty')

if [ -n "$GROUP_ID" ]; then
    print_success "그룹 생성 성공 (GROUP_ID: $GROUP_ID)"
else
    print_error "그룹 생성 실패"
    echo "$CREATE_GROUP_RESPONSE" > group_error.log
    exit 1
fi

# ===== 3. 사용자 상세 정보 =====
print_section "User Info APIs"

CREATE_USER_INFO_RESPONSE=$(curl -s -X POST "$BASE_URL/api/userinfo" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -d "{
    \"userId\": \"$USER_ID\",
    \"groupId\": \"$GROUP_ID\",
    \"role\": \"USER\" 
  }")

USER_INFO_ID=$(echo "$CREATE_USER_INFO_RESPONSE" | jq -r '.userId // empty')

if [ -n "$USER_INFO_ID" ]; then
    print_success "사용자 상세 정보 생성 성공"
else
    print_error "사용자 상세 정보 생성 실패"
    echo "$CREATE_USER_INFO_RESPONSE" > userinfo_error.log
fi

# ===== 4. 사용자 관련 API =====
print_section "User APIs"

# 4.1 사용자 정보 수정
UPDATE_USER_RESPONSE=$(curl -s -X PUT "$BASE_URL/api/users/$USER_ID" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -d "{
    \"name\": \"홍길동_수정\",
    \"email\": \"$EMAIL\" 
  }")

UPDATED_USER_ID=$(echo "$UPDATE_USER_RESPONSE" | jq -r '.userId // empty')
if [ -n "$UPDATED_USER_ID" ]; then
    print_success "사용자 정보 수정 성공"
else
    print_error "사용자 정보 수정 실패"
fi

# 4.2 비밀번호 변경
CHANGE_PW_RESPONSE=$(curl -s -X PATCH "$BASE_URL/api/users/$USER_ID/password" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -d "{
    \"currentPassword\": \"$PASSWORD\",
    \"newPassword\": \"newPassword123\" 
  }")
PASSWORD="newPassword123"

print_success "비밀번호 변경 요청 완료"
echo "Response: $CHANGE_PW_RESPONSE"

# ===== 5. 팀 관련 API =====
print_section "Team APIs"

# 5.1 팀 생성
CREATE_TEAM_RESPONSE=$(curl -s -X POST "$BASE_URL/api/teams" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -d "{
    \"teamName\": \"테스트 팀\",
    \"companyName\": \"$COMPANY_NAME\",
    \"leaderId\": \"$USER_ID\",
    \"description\": \"API 테스트용 팀\" 
  }")

TEAM_ID=$(echo "$CREATE_TEAM_RESPONSE" | jq -r '.data.team.teamId // empty')

if [ -n "$TEAM_ID" ] && [ "$TEAM_ID" != "null" ]; then
    print_success "팀 생성 성공 (TEAM_ID: $TEAM_ID)"
else
    print_error "팀 생성 실패"
    echo "$CREATE_TEAM_RESPONSE" > team_error.log
fi

# 5.2 팀에 멤버 추가
ADD_USER_EMAIL="add-user-$(date +%s%N)@example.com"
ADD_USER_PW="password123"

ADD_USER_JSON=$(printf '{"name":"추가멤버","email":"%s","password":"%s"}' "$ADD_USER_EMAIL" "$ADD_USER_PW")
ADD_USER_RESPONSE=$(curl -s -X POST "$BASE_URL/api/auth/signup" -H "Content-Type: application/json" -d "$ADD_USER_JSON")
ADD_USER_ID=$(echo "$ADD_USER_RESPONSE" | jq -r '.userId // empty')

if [ -n "$ADD_USER_ID" ]; then
    print_success "추가 멤버 생성 성공 (USER_ID: $ADD_USER_ID)"
else
    print_error "추가 멤버 생성 실패"
    echo "$ADD_USER_RESPONSE" > add_user_error.log
fi

USER_INFO_JSON=$(printf '{"userId":"%s","groupId":"%s","role":"USER"}' "$ADD_USER_ID" "$GROUP_ID")
curl -s -X POST "$BASE_URL/api/userinfo" -H "Content-Type: application/json" -H "Authorization: Bearer $ACCESS_TOKEN" -d "$USER_INFO_JSON"

ADD_MEMBER_RESPONSE=$(curl -s -X POST "$BASE_URL/api/teams/$TEAM_ID/members?requesterId=$USER_ID&userId=$ADD_USER_ID&role=MEMBER" \
  -H "Authorization: Bearer $ACCESS_TOKEN")
echo "Response: $ADD_MEMBER_RESPONSE"

# ===== 6. 정리(삭제) API =====
print_section "Cleanup APIs"

# 6.1 팀 삭제
DELETE_TEAM_RESPONSE=$(curl -s -X DELETE "$BASE_URL/api/teams/$TEAM_ID?requesterId=$USER_ID" \
  -H "Authorization: Bearer $ACCESS_TOKEN")
echo "Response: $DELETE_TEAM_RESPONSE"

# 6.2 그룹 내 사용자 정보 삭제
DELETE_USERINFO_RESPONSE=$(curl -s -X DELETE "$BASE_URL/api/userinfo/$USER_INFO_ID" \
  -H "Authorization: Bearer $ACCESS_TOKEN")
echo "Response: $DELETE_USERINFO_RESPONSE"

# 6.3 그룹 삭제
DELETE_GROUP_RESPONSE=$(curl -s -X DELETE "$BASE_URL/api/groups/$GROUP_ID" \
  -H "Authorization: Bearer $ACCESS_TOKEN")
echo "Response: $DELETE_GROUP_RESPONSE"

# 6.4 사용자 삭제
DELETE_USER_RESPONSE=$(curl -s -X DELETE "$BASE_URL/api/users/$USER_ID" \
  -H "Authorization: Bearer $ACCESS_TOKEN")
echo "Response: $DELETE_USER_RESPONSE"

# 6.5 추가 멤버 삭제
curl -s -X DELETE "$BASE_URL/api/users/$ADD_USER_ID" -H "Authorization: Bearer $ACCESS_TOKEN"

print_section "✅ 모든 테스트 완료!"

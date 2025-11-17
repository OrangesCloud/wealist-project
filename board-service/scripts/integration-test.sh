#!/bin/bash
# =============================================================================
# Board Service 통합 테스트 스크립트
# =============================================================================
# 실제 user-service API를 사용하여 board-service를 테스트합니다.
# 
# 사용법:
#   ./board-service/scripts/integration-test.sh
#
# 전제 조건:
#   - Docker Compose로 user-service와 board-service가 실행 중이어야 함
#   - DataInitializer가 테스트 데이터를 생성했어야 함
# =============================================================================

set -e

# 색상 정의
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# 설정
USER_API_URL="${USER_API_URL:-http://localhost:8080}"
BOARD_API_URL="${BOARD_API_URL:-http://localhost:8000}"
TEST_USER_EMAIL="${TEST_USER_EMAIL:-user1@example.com}"
POSTGRES_CONTAINER="${POSTGRES_CONTAINER:-wealist-postgres}"
DB_USER="${DB_USER:-wealist_user}"
DB_NAME="${DB_NAME:-wealist_user_db}"

# 타임아웃 설정 (초)
HEALTH_CHECK_TIMEOUT=30
HEALTH_CHECK_INTERVAL=2

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Board Service 통합 테스트${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo -e "${CYAN}설정:${NC}"
echo -e "  User API:    ${USER_API_URL}"
echo -e "  Board API:   ${BOARD_API_URL}"
echo -e "  테스트 사용자: ${TEST_USER_EMAIL}"
echo ""

# =============================================================================
# 함수: 서비스 Health Check
# =============================================================================
check_service_health() {
  local service_name=$1
  local health_url=$2
  local timeout=$3
  
  echo -e "${YELLOW}${service_name} 상태 확인 중...${NC}"
  
  local elapsed=0
  while [ $elapsed -lt $timeout ]; do
    HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" "${health_url}" 2>/dev/null || echo "000")
    
    if [ "$HTTP_CODE" = "200" ]; then
      echo -e "${GREEN}✅ ${service_name} 정상 작동 중 (HTTP 200)${NC}"
      return 0
    fi
    
    echo -e "${CYAN}   대기 중... (${elapsed}/${timeout}초)${NC}"
    sleep $HEALTH_CHECK_INTERVAL
    elapsed=$((elapsed + HEALTH_CHECK_INTERVAL))
  done
  
  echo -e "${RED}❌ ${service_name} 응답 없음 (타임아웃)${NC}"
  echo -e "${YELLOW}💡 다음 명령어로 서비스를 시작하세요:${NC}"
  echo -e "   cd docker && ./scripts/dev.sh up${NC}"
  return 1
}

# =============================================================================
# Step 1: 서비스 Health Check
# =============================================================================
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Step 1: 서비스 상태 확인${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# User-service health check
if ! check_service_health "user-service" "${USER_API_URL}/actuator/health" $HEALTH_CHECK_TIMEOUT; then
  exit 1
fi
echo ""

# Board-service health check
if ! check_service_health "board-service" "${BOARD_API_URL}/health" $HEALTH_CHECK_TIMEOUT; then
  exit 1
fi
echo ""

# =============================================================================
# Step 2: 테스트 사용자 정보 조회
# =============================================================================
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Step 2: 테스트 사용자 정보 조회${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

echo -e "${YELLOW}데이터베이스에서 사용자 ID 조회 중...${NC}"

# PostgreSQL에서 사용자 ID 조회
USER_ID=$(docker exec $POSTGRES_CONTAINER psql -U $DB_USER -d $DB_NAME -t -c \
  "SELECT user_id FROM users WHERE email = '${TEST_USER_EMAIL}' LIMIT 1;" 2>/dev/null | tr -d ' \n')

if [ -z "$USER_ID" ]; then
  echo -e "${RED}❌ 사용자를 찾을 수 없습니다: ${TEST_USER_EMAIL}${NC}"
  echo -e "${YELLOW}💡 다음을 확인하세요:${NC}"
  echo -e "   1. PostgreSQL 컨테이너가 실행 중인지 확인: docker ps | grep postgres${NC}"
  echo -e "   2. DataInitializer가 실행되었는지 확인: docker logs user-service | grep 'Dummy data'${NC}"
  echo -e "   3. 데이터베이스에 사용자가 있는지 확인:${NC}"
  echo -e "      docker exec $POSTGRES_CONTAINER psql -U $DB_USER -d $DB_NAME -c 'SELECT email FROM users LIMIT 5;'${NC}"
  exit 1
fi

echo -e "${GREEN}✅ 사용자 정보 조회 성공${NC}"
echo -e "   User ID: ${USER_ID}"
echo -e "   Email:   ${TEST_USER_EMAIL}"
echo ""

# =============================================================================
# Step 3: 테스트 토큰 생성
# =============================================================================
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Step 3: 테스트 토큰 생성${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

echo -e "${YELLOW}user-service에서 테스트 토큰 생성 중...${NC}"

# 테스트 토큰 엔드포인트 호출
TOKEN_RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" "${USER_API_URL}/api/users/test/${USER_ID}")
HTTP_STATUS=$(echo "$TOKEN_RESPONSE" | grep "HTTP_STATUS" | cut -d':' -f2)
ACCESS_TOKEN=$(echo "$TOKEN_RESPONSE" | sed '/HTTP_STATUS/d' | tr -d '\n')

if [ "$HTTP_STATUS" != "200" ] || [ -z "$ACCESS_TOKEN" ]; then
  echo -e "${RED}❌ 토큰 생성 실패 (HTTP ${HTTP_STATUS})${NC}"
  echo -e "${YELLOW}응답: ${ACCESS_TOKEN}${NC}"
  echo -e "${YELLOW}💡 user-service 로그를 확인하세요:${NC}"
  echo -e "   docker logs user-service -f${NC}"
  exit 1
fi

echo -e "${GREEN}✅ 테스트 토큰 생성 성공${NC}"
echo -e "   토큰 길이: ${#ACCESS_TOKEN} 문자"
echo -e "   토큰 앞부분: ${ACCESS_TOKEN:0:50}...${NC}"
echo ""

# =============================================================================
# Step 4: 워크스페이스 정보 조회
# =============================================================================
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Step 4: 워크스페이스 정보 조회${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

echo -e "${YELLOW}사용자의 기본 워크스페이스 조회 중...${NC}"

# 데이터베이스에서 워크스페이스 ID 조회
WORKSPACE_ID=$(docker exec $POSTGRES_CONTAINER psql -U $DB_USER -d $DB_NAME -t -c \
  "SELECT workspace_id FROM workspace_members WHERE user_id = '${USER_ID}' AND is_default = true LIMIT 1;" \
  2>/dev/null | tr -d ' \n')

# 기본 워크스페이스가 없으면 첫 번째 워크스페이스 사용
if [ -z "$WORKSPACE_ID" ]; then
  echo -e "${YELLOW}   기본 워크스페이스가 없습니다. 첫 번째 워크스페이스를 사용합니다.${NC}"
  WORKSPACE_ID=$(docker exec $POSTGRES_CONTAINER psql -U $DB_USER -d $DB_NAME -t -c \
    "SELECT workspace_id FROM workspace_members WHERE user_id = '${USER_ID}' LIMIT 1;" \
    2>/dev/null | tr -d ' \n')
fi

if [ -z "$WORKSPACE_ID" ]; then
  echo -e "${RED}❌ 워크스페이스를 찾을 수 없습니다${NC}"
  echo -e "${YELLOW}💡 사용자가 워크스페이스에 속해 있는지 확인하세요:${NC}"
  echo -e "   docker exec $POSTGRES_CONTAINER psql -U $DB_USER -d $DB_NAME -c \\${NC}"
  echo -e "     \"SELECT * FROM workspace_members WHERE user_id = '${USER_ID}';\"${NC}"
  exit 1
fi

# 워크스페이스 이름 조회
WORKSPACE_NAME=$(docker exec $POSTGRES_CONTAINER psql -U $DB_USER -d $DB_NAME -t -c \
  "SELECT workspace_name FROM workspaces WHERE workspace_id = '${WORKSPACE_ID}';" \
  2>/dev/null | tr -d '\n' | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')

echo -e "${GREEN}✅ 워크스페이스 정보 조회 성공${NC}"
echo -e "   Workspace ID:   ${WORKSPACE_ID}"
echo -e "   Workspace Name: ${WORKSPACE_NAME}"
echo ""

# =============================================================================
# Step 5: Board-service API 테스트
# =============================================================================
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Step 5: Board-service API 테스트${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# 테스트 결과 카운터
TESTS_PASSED=0
TESTS_FAILED=0

# 테스트 1: GET /api/projects (프로젝트 목록 조회)
echo -e "${YELLOW}테스트 1: GET /api/projects?workspaceId=${WORKSPACE_ID}${NC}"
PROJECTS_RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" \
  "${BOARD_API_URL}/api/projects?workspaceId=${WORKSPACE_ID}" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}")

HTTP_STATUS=$(echo "$PROJECTS_RESPONSE" | grep "HTTP_STATUS" | cut -d':' -f2)
RESPONSE_BODY=$(echo "$PROJECTS_RESPONSE" | sed '/HTTP_STATUS/d')

echo "   HTTP 상태: $HTTP_STATUS"
echo "   응답: ${RESPONSE_BODY:0:200}..."

if [ "$HTTP_STATUS" = "200" ]; then
  echo -e "${GREEN}   ✅ 성공: 프로젝트 목록 조회${NC}"
  TESTS_PASSED=$((TESTS_PASSED + 1))
else
  echo -e "${RED}   ❌ 실패: 예상 200, 실제 ${HTTP_STATUS}${NC}"
  TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo ""

# 테스트 2: POST /api/projects (프로젝트 생성)
echo -e "${YELLOW}테스트 2: POST /api/projects (프로젝트 생성)${NC}"
PROJECT_NAME="통합테스트 프로젝트 $(date +%s)"
CREATE_PROJECT_RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X POST \
  "${BOARD_API_URL}/api/projects" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}" \
  -d "{
    \"workspaceId\": \"${WORKSPACE_ID}\",
    \"name\": \"${PROJECT_NAME}\",
    \"description\": \"통합 테스트로 생성된 프로젝트\"
  }")

CREATE_HTTP_STATUS=$(echo "$CREATE_PROJECT_RESPONSE" | grep "HTTP_STATUS" | cut -d':' -f2)
CREATE_RESPONSE_BODY=$(echo "$CREATE_PROJECT_RESPONSE" | sed '/HTTP_STATUS/d')

echo "   HTTP 상태: $CREATE_HTTP_STATUS"
echo "   응답: ${CREATE_RESPONSE_BODY:0:200}..."

if [ "$CREATE_HTTP_STATUS" = "201" ] || [ "$CREATE_HTTP_STATUS" = "200" ]; then
  echo -e "${GREEN}   ✅ 성공: 프로젝트 생성${NC}"
  TESTS_PASSED=$((TESTS_PASSED + 1))
  
  # 프로젝트 ID 추출
  PROJECT_ID=$(echo "$CREATE_RESPONSE_BODY" | grep -o '"projectId":"[^"]*' | head -1 | cut -d'"' -f4)
  if [ -z "$PROJECT_ID" ]; then
    PROJECT_ID=$(echo "$CREATE_RESPONSE_BODY" | grep -o '"project_id":"[^"]*' | head -1 | cut -d'"' -f4)
  fi
  
  if [ -n "$PROJECT_ID" ]; then
    echo -e "   프로젝트 ID: ${PROJECT_ID}"
  fi
else
  echo -e "${RED}   ❌ 실패: 예상 201/200, 실제 ${CREATE_HTTP_STATUS}${NC}"
  TESTS_FAILED=$((TESTS_FAILED + 1))
  PROJECT_ID=""
fi
echo ""

# 테스트 3: GET /api/projects/{projectId} (프로젝트 상세 조회)
if [ -n "$PROJECT_ID" ]; then
  echo -e "${YELLOW}테스트 3: GET /api/projects/${PROJECT_ID} (프로젝트 상세 조회)${NC}"
  GET_PROJECT_RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" \
    "${BOARD_API_URL}/api/projects/${PROJECT_ID}" \
    -H "Authorization: Bearer ${ACCESS_TOKEN}")

  GET_HTTP_STATUS=$(echo "$GET_PROJECT_RESPONSE" | grep "HTTP_STATUS" | cut -d':' -f2)
  GET_RESPONSE_BODY=$(echo "$GET_PROJECT_RESPONSE" | sed '/HTTP_STATUS/d')

  echo "   HTTP 상태: $GET_HTTP_STATUS"
  echo "   응답: ${GET_RESPONSE_BODY:0:200}..."

  if [ "$GET_HTTP_STATUS" = "200" ]; then
    echo -e "${GREEN}   ✅ 성공: 프로젝트 상세 조회${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
  else
    echo -e "${RED}   ❌ 실패: 예상 200, 실제 ${GET_HTTP_STATUS}${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
  fi
  echo ""
fi

# 테스트 4: GET /api/boards?projectId={projectId} (보드 목록 조회)
if [ -n "$PROJECT_ID" ]; then
  echo -e "${YELLOW}테스트 4: GET /api/boards?projectId=${PROJECT_ID} (보드 목록 조회)${NC}"
  BOARDS_RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" \
    "${BOARD_API_URL}/api/boards?projectId=${PROJECT_ID}" \
    -H "Authorization: Bearer ${ACCESS_TOKEN}")

  BOARDS_HTTP_STATUS=$(echo "$BOARDS_RESPONSE" | grep "HTTP_STATUS" | cut -d':' -f2)
  BOARDS_RESPONSE_BODY=$(echo "$BOARDS_RESPONSE" | sed '/HTTP_STATUS/d')

  echo "   HTTP 상태: $BOARDS_HTTP_STATUS"
  echo "   응답: ${BOARDS_RESPONSE_BODY:0:200}..."

  if [ "$BOARDS_HTTP_STATUS" = "200" ]; then
    echo -e "${GREEN}   ✅ 성공: 보드 목록 조회${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
  else
    echo -e "${RED}   ❌ 실패: 예상 200, 실제 ${BOARDS_HTTP_STATUS}${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
  fi
  echo ""
fi

# 테스트 5: POST /api/boards (보드 생성)
if [ -n "$PROJECT_ID" ]; then
  echo -e "${YELLOW}테스트 5: POST /api/boards (보드 생성)${NC}"
  BOARD_NAME="통합테스트 보드 $(date +%s)"
  CREATE_BOARD_RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X POST \
    "${BOARD_API_URL}/api/boards" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer ${ACCESS_TOKEN}" \
    -d "{
      \"projectId\": \"${PROJECT_ID}\",
      \"title\": \"${BOARD_NAME}\",
      \"content\": \"통합 테스트로 생성된 보드\"
    }")

  CREATE_BOARD_HTTP_STATUS=$(echo "$CREATE_BOARD_RESPONSE" | grep "HTTP_STATUS" | cut -d':' -f2)
  CREATE_BOARD_RESPONSE_BODY=$(echo "$CREATE_BOARD_RESPONSE" | sed '/HTTP_STATUS/d')

  echo "   HTTP 상태: $CREATE_BOARD_HTTP_STATUS"
  echo "   응답: ${CREATE_BOARD_RESPONSE_BODY:0:200}..."

  if [ "$CREATE_BOARD_HTTP_STATUS" = "201" ] || [ "$CREATE_BOARD_HTTP_STATUS" = "200" ]; then
    echo -e "${GREEN}   ✅ 성공: 보드 생성${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
    
    # 보드 ID 추출
    BOARD_ID=$(echo "$CREATE_BOARD_RESPONSE_BODY" | grep -o '"boardId":"[^"]*' | head -1 | cut -d'"' -f4)
    if [ -z "$BOARD_ID" ]; then
      BOARD_ID=$(echo "$CREATE_BOARD_RESPONSE_BODY" | grep -o '"board_id":"[^"]*' | head -1 | cut -d'"' -f4)
    fi
    
    if [ -n "$BOARD_ID" ]; then
      echo -e "   보드 ID: ${BOARD_ID}"
    fi
  else
    echo -e "${RED}   ❌ 실패: 예상 201/200, 실제 ${CREATE_BOARD_HTTP_STATUS}${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
    BOARD_ID=""
  fi
  echo ""
fi

# 테스트 6: GET /api/boards/{boardId} (보드 상세 조회)
if [ -n "$BOARD_ID" ]; then
  echo -e "${YELLOW}테스트 6: GET /api/boards/${BOARD_ID} (보드 상세 조회)${NC}"
  GET_BOARD_RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" \
    "${BOARD_API_URL}/api/boards/${BOARD_ID}" \
    -H "Authorization: Bearer ${ACCESS_TOKEN}")

  GET_BOARD_HTTP_STATUS=$(echo "$GET_BOARD_RESPONSE" | grep "HTTP_STATUS" | cut -d':' -f2)
  GET_BOARD_RESPONSE_BODY=$(echo "$GET_BOARD_RESPONSE" | sed '/HTTP_STATUS/d')

  echo "   HTTP 상태: $GET_BOARD_HTTP_STATUS"
  echo "   응답: ${GET_BOARD_RESPONSE_BODY:0:200}..."

  if [ "$GET_BOARD_HTTP_STATUS" = "200" ]; then
    echo -e "${GREEN}   ✅ 성공: 보드 상세 조회${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
  else
    echo -e "${RED}   ❌ 실패: 예상 200, 실제 ${GET_BOARD_HTTP_STATUS}${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
  fi
  echo ""
fi

# 테스트 7: PUT /api/boards/{boardId} (보드 수정)
if [ -n "$BOARD_ID" ]; then
  echo -e "${YELLOW}테스트 7: PUT /api/boards/${BOARD_ID} (보드 수정)${NC}"
  UPDATE_BOARD_RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X PUT \
    "${BOARD_API_URL}/api/boards/${BOARD_ID}" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer ${ACCESS_TOKEN}" \
    -d "{
      \"title\": \"수정된 보드 제목\",
      \"content\": \"수정된 보드 내용\"
    }")

  UPDATE_BOARD_HTTP_STATUS=$(echo "$UPDATE_BOARD_RESPONSE" | grep "HTTP_STATUS" | cut -d':' -f2)
  UPDATE_BOARD_RESPONSE_BODY=$(echo "$UPDATE_BOARD_RESPONSE" | sed '/HTTP_STATUS/d')

  echo "   HTTP 상태: $UPDATE_BOARD_HTTP_STATUS"
  echo "   응답: ${UPDATE_BOARD_RESPONSE_BODY:0:200}..."

  if [ "$UPDATE_BOARD_HTTP_STATUS" = "200" ]; then
    echo -e "${GREEN}   ✅ 성공: 보드 수정${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
  else
    echo -e "${RED}   ❌ 실패: 예상 200, 실제 ${UPDATE_BOARD_HTTP_STATUS}${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
  fi
  echo ""
fi

# 테스트 8: POST /api/participants (참가자 추가)
if [ -n "$BOARD_ID" ]; then
  echo -e "${YELLOW}테스트 8: POST /api/participants (참가자 추가)${NC}"
  ADD_PARTICIPANT_RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X POST \
    "${BOARD_API_URL}/api/participants" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer ${ACCESS_TOKEN}" \
    -d "{
      \"boardId\": \"${BOARD_ID}\",
      \"userId\": \"${USER_ID}\"
    }")

  ADD_PARTICIPANT_HTTP_STATUS=$(echo "$ADD_PARTICIPANT_RESPONSE" | grep "HTTP_STATUS" | cut -d':' -f2)
  ADD_PARTICIPANT_RESPONSE_BODY=$(echo "$ADD_PARTICIPANT_RESPONSE" | sed '/HTTP_STATUS/d')

  echo "   HTTP 상태: $ADD_PARTICIPANT_HTTP_STATUS"
  echo "   응답: ${ADD_PARTICIPANT_RESPONSE_BODY:0:200}..."

  if [ "$ADD_PARTICIPANT_HTTP_STATUS" = "201" ] || [ "$ADD_PARTICIPANT_HTTP_STATUS" = "200" ]; then
    echo -e "${GREEN}   ✅ 성공: 참가자 추가${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
  else
    echo -e "${RED}   ❌ 실패: 예상 201/200, 실제 ${ADD_PARTICIPANT_HTTP_STATUS}${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
  fi
  echo ""
fi

# 테스트 9: GET /api/participants/board/{boardId} (참가자 목록 조회)
if [ -n "$BOARD_ID" ]; then
  echo -e "${YELLOW}테스트 9: GET /api/participants/board/${BOARD_ID} (참가자 목록 조회)${NC}"
  GET_PARTICIPANTS_RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" \
    "${BOARD_API_URL}/api/participants/board/${BOARD_ID}" \
    -H "Authorization: Bearer ${ACCESS_TOKEN}")

  GET_PARTICIPANTS_HTTP_STATUS=$(echo "$GET_PARTICIPANTS_RESPONSE" | grep "HTTP_STATUS" | cut -d':' -f2)
  GET_PARTICIPANTS_RESPONSE_BODY=$(echo "$GET_PARTICIPANTS_RESPONSE" | sed '/HTTP_STATUS/d')

  echo "   HTTP 상태: $GET_PARTICIPANTS_HTTP_STATUS"
  echo "   응답: ${GET_PARTICIPANTS_RESPONSE_BODY:0:200}..."

  if [ "$GET_PARTICIPANTS_HTTP_STATUS" = "200" ]; then
    echo -e "${GREEN}   ✅ 성공: 참가자 목록 조회${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
  else
    echo -e "${RED}   ❌ 실패: 예상 200, 실제 ${GET_PARTICIPANTS_HTTP_STATUS}${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
  fi
  echo ""
fi

# 테스트 10: POST /api/comments (댓글 생성)
if [ -n "$BOARD_ID" ]; then
  echo -e "${YELLOW}테스트 10: POST /api/comments (댓글 생성)${NC}"
  CREATE_COMMENT_RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X POST \
    "${BOARD_API_URL}/api/comments" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer ${ACCESS_TOKEN}" \
    -d "{
      \"boardId\": \"${BOARD_ID}\",
      \"content\": \"통합 테스트 댓글\"
    }")

  CREATE_COMMENT_HTTP_STATUS=$(echo "$CREATE_COMMENT_RESPONSE" | grep "HTTP_STATUS" | cut -d':' -f2)
  CREATE_COMMENT_RESPONSE_BODY=$(echo "$CREATE_COMMENT_RESPONSE" | sed '/HTTP_STATUS/d')

  echo "   HTTP 상태: $CREATE_COMMENT_HTTP_STATUS"
  echo "   응답: ${CREATE_COMMENT_RESPONSE_BODY:0:200}..."

  if [ "$CREATE_COMMENT_HTTP_STATUS" = "201" ] || [ "$CREATE_COMMENT_HTTP_STATUS" = "200" ]; then
    echo -e "${GREEN}   ✅ 성공: 댓글 생성${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
    
    # 댓글 ID 추출
    COMMENT_ID=$(echo "$CREATE_COMMENT_RESPONSE_BODY" | grep -o '"commentId":"[^"]*' | head -1 | cut -d'"' -f4)
    if [ -z "$COMMENT_ID" ]; then
      COMMENT_ID=$(echo "$CREATE_COMMENT_RESPONSE_BODY" | grep -o '"comment_id":"[^"]*' | head -1 | cut -d'"' -f4)
    fi
    
    if [ -n "$COMMENT_ID" ]; then
      echo -e "   댓글 ID: ${COMMENT_ID}"
    fi
  else
    echo -e "${RED}   ❌ 실패: 예상 201/200, 실제 ${CREATE_COMMENT_HTTP_STATUS}${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
    COMMENT_ID=""
  fi
  echo ""
fi

# 테스트 11: GET /api/comments/board/{boardId} (댓글 목록 조회)
if [ -n "$BOARD_ID" ]; then
  echo -e "${YELLOW}테스트 11: GET /api/comments/board/${BOARD_ID} (댓글 목록 조회)${NC}"
  GET_COMMENTS_RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" \
    "${BOARD_API_URL}/api/comments/board/${BOARD_ID}" \
    -H "Authorization: Bearer ${ACCESS_TOKEN}")

  GET_COMMENTS_HTTP_STATUS=$(echo "$GET_COMMENTS_RESPONSE" | grep "HTTP_STATUS" | cut -d':' -f2)
  GET_COMMENTS_RESPONSE_BODY=$(echo "$GET_COMMENTS_RESPONSE" | sed '/HTTP_STATUS/d')

  echo "   HTTP 상태: $GET_COMMENTS_HTTP_STATUS"
  echo "   응답: ${GET_COMMENTS_RESPONSE_BODY:0:200}..."

  if [ "$GET_COMMENTS_HTTP_STATUS" = "200" ]; then
    echo -e "${GREEN}   ✅ 성공: 댓글 목록 조회${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
  else
    echo -e "${RED}   ❌ 실패: 예상 200, 실제 ${GET_COMMENTS_HTTP_STATUS}${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
  fi
  echo ""
fi

# 테스트 12: PUT /api/comments/{commentId} (댓글 수정)
if [ -n "$COMMENT_ID" ]; then
  echo -e "${YELLOW}테스트 12: PUT /api/comments/${COMMENT_ID} (댓글 수정)${NC}"
  UPDATE_COMMENT_RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X PUT \
    "${BOARD_API_URL}/api/comments/${COMMENT_ID}" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer ${ACCESS_TOKEN}" \
    -d "{
      \"content\": \"수정된 댓글 내용\"
    }")

  UPDATE_COMMENT_HTTP_STATUS=$(echo "$UPDATE_COMMENT_RESPONSE" | grep "HTTP_STATUS" | cut -d':' -f2)
  UPDATE_COMMENT_RESPONSE_BODY=$(echo "$UPDATE_COMMENT_RESPONSE" | sed '/HTTP_STATUS/d')

  echo "   HTTP 상태: $UPDATE_COMMENT_HTTP_STATUS"
  echo "   응답: ${UPDATE_COMMENT_RESPONSE_BODY:0:200}..."

  if [ "$UPDATE_COMMENT_HTTP_STATUS" = "200" ]; then
    echo -e "${GREEN}   ✅ 성공: 댓글 수정${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
  else
    echo -e "${RED}   ❌ 실패: 예상 200, 실제 ${UPDATE_COMMENT_HTTP_STATUS}${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
  fi
  echo ""
fi

# 테스트 13: GET /api/field-options (필드 옵션 조회)
if [ -n "$PROJECT_ID" ]; then
  echo -e "${YELLOW}테스트 13: GET /api/field-options?projectId=${PROJECT_ID}&fieldType=stage (필드 옵션 조회)${NC}"
  GET_FIELD_OPTIONS_RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" \
    "${BOARD_API_URL}/api/field-options?projectId=${PROJECT_ID}&fieldType=stage" \
    -H "Authorization: Bearer ${ACCESS_TOKEN}")

  GET_FIELD_OPTIONS_HTTP_STATUS=$(echo "$GET_FIELD_OPTIONS_RESPONSE" | grep "HTTP_STATUS" | cut -d':' -f2)
  GET_FIELD_OPTIONS_RESPONSE_BODY=$(echo "$GET_FIELD_OPTIONS_RESPONSE" | sed '/HTTP_STATUS/d')

  echo "   HTTP 상태: $GET_FIELD_OPTIONS_HTTP_STATUS"
  echo "   응답: ${GET_FIELD_OPTIONS_RESPONSE_BODY:0:200}..."

  if [ "$GET_FIELD_OPTIONS_HTTP_STATUS" = "200" ]; then
    echo -e "${GREEN}   ✅ 성공: 필드 옵션 조회${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
  else
    echo -e "${RED}   ❌ 실패: 예상 200, 실제 ${GET_FIELD_OPTIONS_HTTP_STATUS}${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
  fi
  echo ""
fi

# 테스트 14: DELETE /api/comments/{commentId} (댓글 삭제)
if [ -n "$COMMENT_ID" ]; then
  echo -e "${YELLOW}테스트 14: DELETE /api/comments/${COMMENT_ID} (댓글 삭제)${NC}"
  DELETE_COMMENT_RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X DELETE \
    "${BOARD_API_URL}/api/comments/${COMMENT_ID}" \
    -H "Authorization: Bearer ${ACCESS_TOKEN}")

  DELETE_COMMENT_HTTP_STATUS=$(echo "$DELETE_COMMENT_RESPONSE" | grep "HTTP_STATUS" | cut -d':' -f2)
  DELETE_COMMENT_RESPONSE_BODY=$(echo "$DELETE_COMMENT_RESPONSE" | sed '/HTTP_STATUS/d')

  echo "   HTTP 상태: $DELETE_COMMENT_HTTP_STATUS"
  echo "   응답: ${DELETE_COMMENT_RESPONSE_BODY:0:200}..."

  if [ "$DELETE_COMMENT_HTTP_STATUS" = "200" ] || [ "$DELETE_COMMENT_HTTP_STATUS" = "204" ]; then
    echo -e "${GREEN}   ✅ 성공: 댓글 삭제${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
  else
    echo -e "${RED}   ❌ 실패: 예상 200/204, 실제 ${DELETE_COMMENT_HTTP_STATUS}${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
  fi
  echo ""
fi

# 테스트 15: DELETE /api/boards/{boardId} (보드 삭제)
if [ -n "$BOARD_ID" ]; then
  echo -e "${YELLOW}테스트 15: DELETE /api/boards/${BOARD_ID} (보드 삭제)${NC}"
  DELETE_BOARD_RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X DELETE \
    "${BOARD_API_URL}/api/boards/${BOARD_ID}" \
    -H "Authorization: Bearer ${ACCESS_TOKEN}")

  DELETE_BOARD_HTTP_STATUS=$(echo "$DELETE_BOARD_RESPONSE" | grep "HTTP_STATUS" | cut -d':' -f2)
  DELETE_BOARD_RESPONSE_BODY=$(echo "$DELETE_BOARD_RESPONSE" | sed '/HTTP_STATUS/d')

  echo "   HTTP 상태: $DELETE_BOARD_HTTP_STATUS"
  echo "   응답: ${DELETE_BOARD_RESPONSE_BODY:0:200}..."

  if [ "$DELETE_BOARD_HTTP_STATUS" = "200" ] || [ "$DELETE_BOARD_HTTP_STATUS" = "204" ]; then
    echo -e "${GREEN}   ✅ 성공: 보드 삭제${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
  else
    echo -e "${RED}   ❌ 실패: 예상 200/204, 실제 ${DELETE_BOARD_HTTP_STATUS}${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
  fi
  echo ""
fi

# 테스트 16: 서비스 간 통신 검증
echo -e "${YELLOW}테스트 16: 서비스 간 통신 검증${NC}"
echo -e "   board-service가 user-service와 통신할 수 있는지 확인 중...${NC}"

# 프로젝트 목록 조회는 내부적으로 user-service를 호출함
VERIFY_RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" \
  "${BOARD_API_URL}/api/projects?workspaceId=${WORKSPACE_ID}" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}")

VERIFY_HTTP_STATUS=$(echo "$VERIFY_RESPONSE" | grep "HTTP_STATUS" | cut -d':' -f2)

if [ "$VERIFY_HTTP_STATUS" = "200" ]; then
  echo -e "${GREEN}   ✅ 성공: board-service ↔ user-service 통신 정상${NC}"
  TESTS_PASSED=$((TESTS_PASSED + 1))
else
  echo -e "${RED}   ❌ 실패: 서비스 간 통신 문제 감지${NC}"
  echo -e "${YELLOW}   💡 Docker 네트워크 설정을 확인하세요:${NC}"
  echo -e "      docker network inspect wealist-backend-net${NC}"
  TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo ""

# =============================================================================
# 테스트 결과 요약
# =============================================================================
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}테스트 결과 요약${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

TOTAL_TESTS=$((TESTS_PASSED + TESTS_FAILED))
echo -e "${CYAN}총 테스트: ${TOTAL_TESTS}${NC}"
echo -e "${GREEN}성공: ${TESTS_PASSED}${NC}"
echo -e "${RED}실패: ${TESTS_FAILED}${NC}"
echo ""

if [ $TESTS_FAILED -eq 0 ]; then
  echo -e "${GREEN}🎉 모든 테스트가 성공했습니다!${NC}"
  EXIT_CODE=0
else
  echo -e "${RED}⚠️  일부 테스트가 실패했습니다.${NC}"
  EXIT_CODE=1
fi

echo ""
echo -e "${CYAN}테스트 정보:${NC}"
echo -e "  사용자 Email:     ${TEST_USER_EMAIL}"
echo -e "  사용자 ID:        ${USER_ID}"
echo -e "  워크스페이스 ID:  ${WORKSPACE_ID}"
echo -e "  워크스페이스 이름: ${WORKSPACE_NAME}"
if [ -n "$PROJECT_ID" ]; then
  echo -e "  생성된 프로젝트 ID: ${PROJECT_ID}"
fi
if [ -n "$BOARD_ID" ]; then
  echo -e "  생성된 보드 ID:     ${BOARD_ID}"
fi
if [ -n "$COMMENT_ID" ]; then
  echo -e "  생성된 댓글 ID:     ${COMMENT_ID}"
fi

echo ""
echo -e "${CYAN}유용한 명령어:${NC}"
echo -e "  board-service 로그 확인:"
echo -e "    ${YELLOW}docker logs board-service -f${NC}"
echo ""
echo -e "  user-service 로그 확인:"
echo -e "    ${YELLOW}docker logs user-service -f${NC}"
echo ""
echo -e "  네트워크 연결 테스트:"
echo -e "    ${YELLOW}docker exec board-service ping -c 3 user-service${NC}"
echo ""
echo -e "  데이터베이스 직접 조회:"
echo -e "    ${YELLOW}docker exec $POSTGRES_CONTAINER psql -U $DB_USER -d $DB_NAME${NC}"
echo ""

exit $EXIT_CODE

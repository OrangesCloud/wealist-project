# Kanban Service API 테스트 가이드

## 📋 사전 준비

Kanban Service API를 테스트하려면 **실제 User Service에서 발급받은 JWT 토큰**이 필요합니다.

---

## 🔑 1. JWT 토큰 발급받기

### 1-1. User Service 회원가입

```bash
curl -X POST http://localhost:8080/api/auth/signup \
  -H "Content-Type: application/json" \
  -d '{
    "email": "testuser@example.com",
    "password": "YourSecurePassword123!",
    "name": "Test User"
  }'
```

**응답 예시:**
```json
{
  "accessToken": "eyJhbGciOiJIUzUxMiJ9...",
  "refreshToken": "eyJhbGciOiJIUzUxMiJ9...",
  "userId": "22639bd0-7fac-4493-8298-3cbb02bae220",
  "name": "Test User",
  "email": "testuser@example.com",
  "tokenType": "Bearer"
}
```

### 1-2. 로그인 (기존 사용자)

```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "testuser@example.com",
    "password": "YourSecurePassword123!"
  }'
```

### 1-3. 토큰 저장

```bash
# 환경 변수에 저장
export ACCESS_TOKEN="eyJhbGciOiJIUzUxMiJ9..."

# 또는 변수에 저장
TOKEN="eyJhbGciOiJIUzUxMiJ9..."
```

---

## 🧪 2. API 테스트 (전체 플로우)

### 2-1. Workspace 생성

```bash
curl -X POST http://localhost:8000/api/workspaces/ \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -d '{
    "name": "My Workspace",
    "description": "테스트용 워크스페이스"
  }'
```

**응답에서 `workspace_id` 추출**

### 2-2. Project 생성

```bash
WORKSPACE_ID="a1cbfdb9-e3e2-4381-a92d-e36d19a479af"

curl -X POST http://localhost:8000/api/projects/ \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -d "{
    \"name\": \"My Project\",
    \"description\": \"프로젝트 설명\",
    \"workspace_id\": \"$WORKSPACE_ID\"
  }"
```

**응답에서 `project_id` 추출**

### 2-3. Ticket Type 생성

```bash
PROJECT_ID="a561921d-ef21-4784-b07e-5a178a6e87f6"

curl -X POST http://localhost:8000/api/projects/$PROJECT_ID/ticket-types/ \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -d '{
    "type_name": "Bug",
    "description": "버그 리포트",
    "color": "#FF0000",
    "icon": "bug",
    "display_order": 1
  }'
```

### 2-4. Ticket 생성

```bash
curl -X POST http://localhost:8000/api/tickets/ \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -d "{
    \"title\": \"로그인 버그 수정\",
    \"description\": \"구글 로그인 시 오류 발생\",
    \"status\": \"OPEN\",
    \"priority\": \"HIGH\",
    \"project_id\": \"$PROJECT_ID\",
    \"assignee_id\": null
  }"
```

**응답에서 `ticket_id` 추출**

### 2-5. Task 생성

```bash
TICKET_ID="3283c2aa-5a8a-4404-8d6e-84bd545522ee"

curl -X POST http://localhost:8000/api/tasks/ \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -d "{
    \"title\": \"OAuth 설정 확인\",
    \"description\": \"Google OAuth 2.0 설정 검토\",
    \"status\": \"TODO\",
    \"ticket_id\": \"$TICKET_ID\",
    \"assignee_id\": null
  }"
```

### 2-6. Task 완료 처리

```bash
TASK_ID="9d44411f-aec0-4965-b54d-94f018fb0bf3"

curl -X PATCH http://localhost:8000/api/tasks/$TASK_ID/complete \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

---

## 🌐 3. Swagger UI 사용

### 3-1. Swagger UI 접속

브라우저에서 http://localhost:8000/docs 접속

### 3-2. 인증 설정

1. 우측 상단 **"Authorize"** 버튼 클릭
2. **Value** 필드에 토큰 입력 (Bearer 제외)
   ```
   eyJhbGciOiJIUzUxMiJ9...
   ```
3. **"Authorize"** 버튼 클릭
4. **"Close"** 버튼 클릭

### 3-3. API 테스트

이후 모든 API 요청에 자동으로 토큰이 포함됩니다.

---

## ⚠️ 주의사항

### 1. 토큰 만료

- Access Token 기본 만료 시간: **7일**
- 만료 시 User Service에서 재발급 필요

### 2. 실제 사용자만 사용 가능

- **중요**: Kanban Service는 User Service와 연동되어 있습니다
- 임의의 `user_id`로 토큰을 생성하면 **에러가 발생**합니다
- 반드시 **User Service에서 회원가입한 실제 사용자**의 토큰을 사용하세요

### 3. 토큰 검증 실패 시

**에러 예시:**
```json
{
  "detail": "Could not validate credentials"
}
```

**해결 방법:**
1. User Service에서 새 토큰 발급
2. JWT_SECRET이 User Service와 동일한지 확인

---

## 🔍 4. 토큰 디버깅

### 토큰 페이로드 확인 (jwt.io)

1. https://jwt.io 접속
2. 토큰을 Encoded 필드에 붙여넣기
3. Payload 확인:
   ```json
   {
     "sub": "22639bd0-7fac-4493-8298-3cbb02bae220",  // user_id (UUID)
     "iat": 1761968876,                              // 발급 시간
     "exp": 1762573676                               // 만료 시간
   }
   ```

---

## 📚 추가 리소스

- **API 문서**: http://localhost:8000/docs
- **ReDoc**: http://localhost:8000/redoc
- **Health Check**: http://localhost:8000/health
- **User Service API**: http://localhost:8080/swagger-ui.html

---

**마지막 업데이트**: 2025-11-01

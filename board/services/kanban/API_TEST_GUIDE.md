# Kanban API 테스트 가이드

## 문제 해결 완료

모든 API 엔드포인트가 정상 작동합니다. "거부" 문제는 **JWT 인증 토큰이 없어서** 발생한 것입니다.

## 빠른 시작

### 1. 테스트용 토큰 생성

```bash
docker exec kanban-api python generate_test_token.py --user-id 1
```

### 2. API 호출 예시

생성된 토큰을 복사하여 사용하세요:

```bash
# 토큰 변수 설정
TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxIiwiZXhwIjoxNzYxMTM5MDk3LCJpYXQiOjE3NjEwNTI2OTd9.C2DhEEJtnmHJqq7xy2kxUGPWj4YFvEjzmVoDRxYfLqI"

# 워크스페이스 목록 조회
curl -H "Authorization: Bearer $TOKEN" http://localhost:8000/api/workspaces/

# 프로젝트 목록 조회
curl -H "Authorization: Bearer $TOKEN" http://localhost:8000/api/projects/

# 티켓 목록 조회
curl -H "Authorization: Bearer $TOKEN" http://localhost:8000/api/tickets/

# 작업 목록 조회
curl -H "Authorization: Bearer $TOKEN" http://localhost:8000/api/tasks/
```

## 주요 API 엔드포인트

### 중요: Trailing Slash 필수

FastAPI의 특성상 모든 엔드포인트에 **마지막 슬래시(/)가 필수**입니다:
- ❌ `/api/workspaces` → 307 Redirect
- ✅ `/api/workspaces/` → 정상 작동

### Workspaces

| 메서드 | 엔드포인트 | 설명 | 인증 |
|--------|-----------|------|------|
| GET | `/api/workspaces/` | 워크스페이스 목록 | Required |
| POST | `/api/workspaces/` | 워크스페이스 생성 | Required |
| GET | `/api/workspaces/{id}/` | 특정 워크스페이스 조회 | Required |
| PATCH | `/api/workspaces/{id}/` | 워크스페이스 수정 | Required |
| DELETE | `/api/workspaces/{id}/` | 워크스페이스 삭제 | Required |

### Projects

| 메서드 | 엔드포인트 | 설명 | 인증 |
|--------|-----------|------|------|
| GET | `/api/projects/` | 프로젝트 목록 | Required |
| POST | `/api/projects/` | 프로젝트 생성 | Required |
| GET | `/api/projects/{id}/` | 특정 프로젝트 조회 | Required |
| PATCH | `/api/projects/{id}/` | 프로젝트 수정 | Required |
| DELETE | `/api/projects/{id}/` | 프로젝트 삭제 | Required |

### Tickets

| 메서드 | 엔드포인트 | 설명 | 인증 |
|--------|-----------|------|------|
| GET | `/api/tickets/` | 티켓 목록 | Required |
| POST | `/api/tickets/` | 티켓 생성 | Required |
| GET | `/api/tickets/{id}/` | 특정 티켓 조회 | Required |
| PATCH | `/api/tickets/{id}/` | 티켓 수정 | Required |
| DELETE | `/api/tickets/{id}/` | 티켓 삭제 | Required |

### Tasks

| 메서드 | 엔드포인트 | 설명 | 인증 |
|--------|-----------|------|------|
| GET | `/api/tasks/` | 작업 목록 | Required |
| POST | `/api/tasks/` | 작업 생성 | Required |
| GET | `/api/tasks/{id}/` | 특정 작업 조회 | Required |
| PATCH | `/api/tasks/{id}/` | 작업 수정 | Required |
| DELETE | `/api/tasks/{id}/` | 작업 삭제 | Required |

## 프론트엔드에서 사용

### JavaScript/TypeScript 예시

```javascript
// 토큰 설정
const API_TOKEN = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxIiwiZXhwIjoxNzYxMTM5MDk3LCJpYXQiOjE3NjEwNTI2OTd9.C2DhEEJtnmHJqq7xy2kxUGPWj4YFvEjzmVoDRxYfLqI";

// Fetch API 사용
async function getWorkspaces() {
  const response = await fetch('http://localhost:8000/api/workspaces/', {
    headers: {
      'Authorization': `Bearer ${API_TOKEN}`,
      'Content-Type': 'application/json'
    }
  });

  if (!response.ok) {
    throw new Error(`HTTP error! status: ${response.status}`);
  }

  const data = await response.json();
  return data;
}

// Axios 사용
import axios from 'axios';

const api = axios.create({
  baseURL: 'http://localhost:8000',
  headers: {
    'Authorization': `Bearer ${API_TOKEN}`,
    'Content-Type': 'application/json'
  }
});

// 워크스페이스 목록 조회
const workspaces = await api.get('/api/workspaces/');

// 워크스페이스 생성
const newWorkspace = await api.post('/api/workspaces/', {
  name: 'My Workspace',
  description: 'Test workspace'
});
```

### React Query 예시

```typescript
import { useQuery, useMutation } from '@tanstack/react-query';

const API_TOKEN = "your-token-here";

// 워크스페이스 목록 조회
function useWorkspaces() {
  return useQuery({
    queryKey: ['workspaces'],
    queryFn: async () => {
      const response = await fetch('http://localhost:8000/api/workspaces/', {
        headers: {
          'Authorization': `Bearer ${API_TOKEN}`
        }
      });
      return response.json();
    }
  });
}

// 워크스페이스 생성
function useCreateWorkspace() {
  return useMutation({
    mutationFn: async (data: { name: string; description?: string }) => {
      const response = await fetch('http://localhost:8000/api/workspaces/', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${API_TOKEN}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(data)
      });
      return response.json();
    }
  });
}
```

## POST 요청 예시

### 워크스페이스 생성

```bash
curl -X POST http://localhost:8000/api/workspaces/ \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "My Workspace",
    "description": "Test workspace for development"
  }'
```

### 프로젝트 생성

```bash
curl -X POST http://localhost:8000/api/projects/ \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "workspace_id": 1,
    "name": "My Project",
    "description": "Test project",
    "status": "active"
  }'
```

### 티켓 생성

```bash
curl -X POST http://localhost:8000/api/tickets/ \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "project_id": 1,
    "title": "Sample Ticket",
    "description": "This is a test ticket",
    "status": "todo",
    "priority": "medium"
  }'
```

## 토큰 생성 옵션

```bash
# 기본 사용 (user_id=1, 24시간)
docker exec kanban-api python generate_test_token.py

# 다른 사용자 ID로 생성
docker exec kanban-api python generate_test_token.py --user-id 5

# 만료 시간 설정 (48시간)
docker exec kanban-api python generate_test_token.py --user-id 1 --expires 48

# Bearer 형식으로 출력
docker exec kanban-api python generate_test_token.py --bearer
```

## Swagger UI 사용

개발 환경에서는 Swagger UI를 통해 API를 테스트할 수 있습니다:

1. 브라우저에서 http://localhost:8000/docs 접속
2. 우측 상단 "Authorize" 버튼 클릭
3. Bearer 토큰 입력: `eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...`
4. "Authorize" 클릭
5. 이제 모든 API를 Swagger UI에서 테스트 가능

## 페이지네이션

목록 조회 API는 페이지네이션을 지원합니다:

```bash
# 기본 (limit=20, offset=0)
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8000/api/workspaces/"

# 페이지 크기 지정
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8000/api/workspaces/?limit=10&offset=0"

# 두 번째 페이지
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8000/api/workspaces/?limit=10&offset=10"
```

응답 형식:
```json
{
  "total": 25,
  "items": [...],
  "limit": 10,
  "offset": 0
}
```

## 헬스 체크

인증 없이 사용 가능한 엔드포인트:

```bash
# 기본 헬스 체크
curl http://localhost:8000/health

# Liveness probe
curl http://localhost:8000/health/live

# Readiness probe
curl http://localhost:8000/health/ready
```

## 문제 해결

### "Not authenticated" 에러
- JWT 토큰이 없거나 만료됨
- 새 토큰 생성: `docker exec kanban-api python generate_test_token.py`

### 307 Redirect
- URL 끝에 슬래시(/) 추가 필요
- ❌ `/api/workspaces` → ✅ `/api/workspaces/`

### CORS 에러
- 현재 설정: `http://localhost:3000`, `http://localhost:8000`
- 다른 도메인 추가 필요 시 `.env` 파일의 `CORS_ORIGINS` 수정

## 참고사항

- 모든 시간은 UTC 기준
- 토큰 기본 만료 시간: 24시간
- SECRET_KEY는 Member 서비스와 동일하게 설정 필요
- 개발 환경에서만 Swagger UI 사용 가능 (`DEBUG=True`)

# API Migration Guide

## Overview
이 문서는 Board Service API의 표준화 작업에 따른 변경 사항을 설명합니다. 모든 API 엔드포인트와 데이터 구조가 camelCase 네이밍 컨벤션으로 통일되었습니다.

## 주요 변경사항

### 1. Base Path 변경
- **변경 전**: `/api/v1`
- **변경 후**: `/api`

모든 API 엔드포인트의 기본 경로가 `/api/v1`에서 `/api`로 변경되었습니다.

### 2. 경로 파라미터 표준화 (snake_case → camelCase)

모든 URL 경로 파라미터가 camelCase로 변경되었습니다:

| 변경 전 | 변경 후 |
|---------|---------|
| `board_id` | `boardId` |
| `project_id` | `projectId` |
| `workspace_id` | `workspaceId` |
| `comment_id` | `commentId` |
| `user_id` | `userId` |

### 3. API 엔드포인트 변경

#### Boards API

| 변경 전 | 변경 후 |
|---------|---------|
| `GET /api/v1/boards/{board_id}` | `GET /api/boards/{boardId}` |
| `GET /api/v1/boards/project/{project_id}` | `GET /api/boards/project/{projectId}` |
| `PUT /api/v1/boards/{board_id}` | `PUT /api/boards/{boardId}` |
| `DELETE /api/v1/boards/{board_id}` | `DELETE /api/boards/{boardId}` |
| `POST /api/v1/boards` | `POST /api/boards` |

#### Projects API

| 변경 전 | 변경 후 |
|---------|---------|
| `GET /api/v1/projects/workspace/{workspace_id}` | `GET /api/projects/workspace/{workspaceId}` |
| `GET /api/v1/projects/workspace/{workspace_id}/default` | `GET /api/projects/workspace/{workspaceId}/default` |
| `POST /api/v1/projects` | `POST /api/projects` |

#### Comments API

| 변경 전 | 변경 후 |
|---------|---------|
| `GET /api/v1/comments/board/{board_id}` | `GET /api/comments/board/{boardId}` |
| `PUT /api/v1/comments/{comment_id}` | `PUT /api/comments/{commentId}` |
| `DELETE /api/v1/comments/{comment_id}` | `DELETE /api/comments/{commentId}` |
| `POST /api/v1/comments` | `POST /api/comments` |

#### Participants API

| 변경 전 | 변경 후 |
|---------|---------|
| `GET /api/v1/participants/board/{board_id}` | `GET /api/participants/board/{boardId}` |
| `DELETE /api/v1/participants/board/{board_id}/user/{user_id}` | `DELETE /api/participants/board/{boardId}/user/{userId}` |
| `POST /api/v1/participants` | `POST /api/participants` |

### 4. JSON 필드명 변경 (snake_case → camelCase)

#### Participant 요청/응답

**AddParticipantRequest:**
```json
// 변경 전
{
  "board_id": "550e8400-e29b-41d4-a716-446655440000",
  "user_id": "660e8400-e29b-41d4-a716-446655440000"
}

// 변경 후
{
  "boardId": "550e8400-e29b-41d4-a716-446655440000",
  "userId": "660e8400-e29b-41d4-a716-446655440000"
}
```

**ParticipantResponse:**
```json
// 변경 전
{
  "id": "770e8400-e29b-41d4-a716-446655440000",
  "board_id": "550e8400-e29b-41d4-a716-446655440000",
  "user_id": "660e8400-e29b-41d4-a716-446655440000",
  "created_at": "2024-01-15T10:30:00Z"
}

// 변경 후
{
  "id": "770e8400-e29b-41d4-a716-446655440000",
  "boardId": "550e8400-e29b-41d4-a716-446655440000",
  "userId": "660e8400-e29b-41d4-a716-446655440000",
  "createdAt": "2024-01-15T10:30:00Z"
}
```

## 프론트엔드 마이그레이션 가이드

### 1. API 호출 URL 업데이트

모든 API 호출에서 base path와 경로 파라미터를 변경해야 합니다.

**JavaScript/TypeScript 예제:**

```javascript
// 변경 전
const response = await fetch('/api/v1/boards/123');
const board = await fetch(`/api/v1/boards/${board_id}`);
const participants = await fetch(`/api/v1/participants/board/${board_id}`);

// 변경 후
const response = await fetch('/api/boards/123');
const board = await fetch(`/api/boards/${boardId}`);
const participants = await fetch(`/api/participants/board/${boardId}`);
```

**React 예제:**

```javascript
// 변경 전
const fetchBoard = async (board_id) => {
  const response = await fetch(`/api/v1/boards/${board_id}`);
  return response.json();
};

// 변경 후
const fetchBoard = async (boardId) => {
  const response = await fetch(`/api/boards/${boardId}`);
  return response.json();
};
```

**Axios 예제:**

```javascript
// 변경 전
axios.get(`/api/v1/boards/${board_id}`)
axios.delete(`/api/v1/participants/board/${board_id}/user/${user_id}`)

// 변경 후
axios.get(`/api/boards/${boardId}`)
axios.delete(`/api/participants/board/${boardId}/user/${userId}`)
```

### 2. 요청 Body 필드명 업데이트

API 요청 시 전송하는 JSON 필드명을 camelCase로 변경해야 합니다.

**참여자 추가 예제:**

```javascript
// 변경 전
const addParticipant = async (board_id, user_id) => {
  await fetch('/api/v1/participants', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      board_id: board_id,
      user_id: user_id
    })
  });
};

// 변경 후
const addParticipant = async (boardId, userId) => {
  await fetch('/api/participants', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      boardId: boardId,
      userId: userId
    })
  });
};
```

### 3. 응답 데이터 필드명 업데이트

API 응답에서 받은 데이터의 필드명이 변경되었으므로 코드를 업데이트해야 합니다.

**참여자 응답 처리 예제:**

```javascript
// 변경 전
const participants = await response.json();
participants.forEach(p => {
  console.log(p.board_id, p.user_id, p.created_at);
});

// 변경 후
const participants = await response.json();
participants.forEach(p => {
  console.log(p.boardId, p.userId, p.createdAt);
});
```

**React State 예제:**

```javascript
// 변경 전
const [participant, setParticipant] = useState({
  board_id: '',
  user_id: '',
  created_at: ''
});

// 변경 후
const [participant, setParticipant] = useState({
  boardId: '',
  userId: '',
  createdAt: ''
});
```

### 4. TypeScript 타입 정의 업데이트

TypeScript를 사용하는 경우 인터페이스를 업데이트해야 합니다.

```typescript
// 변경 전
interface AddParticipantRequest {
  board_id: string;
  user_id: string;
}

interface ParticipantResponse {
  id: string;
  board_id: string;
  user_id: string;
  created_at: string;
}

// 변경 후
interface AddParticipantRequest {
  boardId: string;
  userId: string;
}

interface ParticipantResponse {
  id: string;
  boardId: string;
  userId: string;
  createdAt: string;
}
```

### 5. API 클라이언트 라이브러리 업데이트

API 클라이언트 클래스나 서비스를 사용하는 경우:

```typescript
// 변경 전
class BoardAPI {
  async getBoard(board_id: string) {
    return axios.get(`/api/v1/boards/${board_id}`);
  }
  
  async getBoardsByProject(project_id: string) {
    return axios.get(`/api/v1/boards/project/${project_id}`);
  }
}

// 변경 후
class BoardAPI {
  async getBoard(boardId: string) {
    return axios.get(`/api/boards/${boardId}`);
  }
  
  async getBoardsByProject(projectId: string) {
    return axios.get(`/api/boards/project/${projectId}`);
  }
}
```

## Breaking Changes 요약

### ⚠️ 중요: 하위 호환성 없음

이번 변경은 **하위 호환성을 제공하지 않습니다**. 이전 API 경로는 더 이상 작동하지 않으므로, 프론트엔드 코드를 반드시 업데이트해야 합니다.

### 변경 사항 체크리스트

- [ ] Base path를 `/api/v1`에서 `/api`로 변경
- [ ] 모든 경로 파라미터를 camelCase로 변경 (`board_id` → `boardId` 등)
- [ ] 요청 body의 필드명을 camelCase로 변경
- [ ] 응답 데이터 처리 코드에서 필드명을 camelCase로 변경
- [ ] TypeScript 타입 정의 업데이트
- [ ] API 클라이언트 라이브러리 업데이트
- [ ] 테스트 코드 업데이트

## 테스트 방법

### 1. Swagger UI에서 확인

서버 실행 후 Swagger UI에서 변경된 API를 확인할 수 있습니다:

```
http://localhost:8000/swagger/index.html
```

### 2. curl로 테스트

```bash
# 이전 경로 (404 에러 예상)
curl http://localhost:8000/api/v1/boards/123

# 새 경로 (정상 동작)
curl http://localhost:8000/api/boards/123

# 참여자 추가 (새 필드명)
curl -X POST http://localhost:8000/api/participants \
  -H "Content-Type: application/json" \
  -d '{
    "boardId": "550e8400-e29b-41d4-a716-446655440000",
    "userId": "660e8400-e29b-41d4-a716-446655440000"
  }'
```

## 롤백 계획

변경 사항을 되돌려야 하는 경우:

1. Git에서 이전 커밋으로 복원
2. Swagger 문서 재생성: `swag init -g cmd/api/main.go -o docs`
3. 서비스 재시작

## 지원

문제가 발생하거나 질문이 있는 경우:
- GitHub Issues에 문의
- Swagger 문서 참조: `http://localhost:8000/swagger/index.html`

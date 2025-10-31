# weAlist Kanban Service API Documentation

## 📋 목차
- [개요](#개요)
- [인증](#인증)
- [공통 응답 형식](#공통-응답-형식)
- [API 엔드포인트](#api-엔드포인트)
  - [1. 워크스페이스 (Workspaces)](#1-워크스페이스-workspaces)
  - [2. 프로젝트 (Projects)](#2-프로젝트-projects)
  - [3. 티켓 (Tickets)](#3-티켓-tickets)
  - [4. 태스크 (Tasks)](#4-태스크-tasks)
  - [5. 알림 (Notifications)](#5-알림-notifications)
  - [6. 티켓 타입 (Ticket Types)](#6-티켓-타입-ticket-types)
  - [7. 헬스 체크 (Health Check)](#7-헬스-체크-health-check)
- [데이터 타입 및 Enum](#데이터-타입-및-enum)
- [에러 코드](#에러-코드)

---

## 개요

**Base URL**: `http://localhost:8000`

**Content-Type**: `application/json`

**Swagger UI**: `http://localhost:8000/docs`

**ReDoc**: `http://localhost:8000/redoc`

---

## 인증

모든 API는 JWT 토큰 인증이 필요합니다.

### 인증 헤더 형식
```
Authorization: Bearer {access_token}
```

### 테스트 토큰 생성
```bash
# 기본 사용자 ID로 토큰 생성
docker exec wealist-kanban-service python scripts/generate_test_token.py

# 특정 사용자 ID로 토큰 생성
docker exec wealist-kanban-service python scripts/generate_test_token.py --user-id {UUID}

# 만료 기간 설정 (기본 24시간)
docker exec wealist-kanban-service python scripts/generate_test_token.py --expire-days 30
```

### 인증이 필요 없는 엔드포인트
- `GET /health` - 기본 헬스 체크
- `GET /health/live` - Liveness probe
- `GET /health/ready` - Readiness probe

---

## 공통 응답 형식

### 성공 응답 (단일 리소스)
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "워크스페이스 이름",
  "description": "설명",
  "created_at": "2025-10-26T00:00:00",
  "updated_at": "2025-10-26T00:00:00",
  "created_by": "660e8400-e29b-41d4-a716-446655440000",
  "updated_by": null
}
```

### 성공 응답 (목록 - 페이지네이션)
```json
{
  "total": 100,
  "items": [...],
  "limit": 20,
  "offset": 0
}
```

### 실패 응답
```json
{
  "detail": "오류 메시지"
}
```

---

## API 엔드포인트

## 1. 워크스페이스 (Workspaces)

워크스페이스는 프로젝트를 그룹화하는 최상위 컨테이너입니다.

### 1.1 워크스페이스 생성
```
POST /api/workspaces/
```

**Headers**: `Authorization: Bearer {token}`

**Request Body**:
```json
{
  "name": "개발팀 워크스페이스",
  "description": "개발 관련 프로젝트를 관리하는 워크스페이스"
}
```

**Response** (201 Created):
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "개발팀 워크스페이스",
  "description": "개발 관련 프로젝트를 관리하는 워크스페이스",
  "created_at": "2025-10-26T00:00:00",
  "updated_at": "2025-10-26T00:00:00",
  "created_by": "660e8400-e29b-41d4-a716-446655440000",
  "updated_by": null
}
```

**에러**:
- `409 Conflict` - 동일한 이름의 워크스페이스가 이미 존재

---

### 1.2 워크스페이스 목록 조회
```
GET /api/workspaces/
```

**Headers**: `Authorization: Bearer {token}`

**Query Parameters**:
- `limit` (int, optional): 페이지 크기 (기본값: 20, 최대: 100)
- `offset` (int, optional): 시작 위치 (기본값: 0)

**Response** (200 OK):
```json
{
  "total": 5,
  "items": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "개발팀 워크스페이스",
      "description": "개발 관련 프로젝트",
      "created_at": "2025-10-26T00:00:00",
      "updated_at": "2025-10-26T00:00:00",
      "created_by": "660e8400-e29b-41d4-a716-446655440000",
      "updated_by": null
    }
  ],
  "limit": 20,
  "offset": 0
}
```

---

### 1.3 특정 워크스페이스 조회
```
GET /api/workspaces/{workspace_id}
```

**Headers**: `Authorization: Bearer {token}`

**Path Parameters**:
- `workspace_id` (UUID): 워크스페이스 ID

**Response** (200 OK): 1.1과 동일한 워크스페이스 객체

**에러**:
- `404 Not Found` - 워크스페이스를 찾을 수 없음

---

### 1.4 워크스페이스 수정
```
PATCH /api/workspaces/{workspace_id}
```

**Headers**: `Authorization: Bearer {token}`

**Request Body** (모든 필드 optional):
```json
{
  "name": "신규 개발팀",
  "description": "업데이트된 설명"
}
```

**Response** (200 OK): 수정된 워크스페이스 객체

---

### 1.5 워크스페이스 삭제
```
DELETE /api/workspaces/{workspace_id}
```

**Headers**: `Authorization: Bearer {token}`

**주의**: 워크스페이스 삭제 시 연결된 모든 프로젝트, 티켓, 태스크가 **애플리케이션 레벨 CASCADE**로 삭제됩니다.

**Response** (204 No Content)

---

## 2. 프로젝트 (Projects)

프로젝트는 워크스페이스 내에서 티켓을 관리하는 단위입니다.

### 2.1 프로젝트 생성
```
POST /api/projects/
```

**Headers**: `Authorization: Bearer {token}`

**Request Body**:
```json
{
  "workspace_id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "백엔드 개발 프로젝트",
  "description": "API 서버 개발",
  "status": "ACTIVE",
  "priority": "HIGH"
}
```

**Field 설명**:
- `workspace_id` (UUID, required): 워크스페이스 ID
- `name` (string, required): 프로젝트 이름 (1-200자)
- `description` (string, optional): 설명 (최대 1000자)
- `status` (string, optional): 프로젝트 상태 (기본값: "PLANNING")
- `priority` (string, optional): 우선순위 (기본값: "MEDIUM")

**Response** (201 Created):
```json
{
  "id": "770e8400-e29b-41d4-a716-446655440000",
  "workspace_id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "백엔드 개발 프로젝트",
  "description": "API 서버 개발",
  "status": "ACTIVE",
  "priority": "HIGH",
  "created_at": "2025-10-26T00:00:00",
  "updated_at": "2025-10-26T00:00:00",
  "created_by": "660e8400-e29b-41d4-a716-446655440000",
  "updated_by": null
}
```

**에러**:
- `404 Not Found` - 워크스페이스를 찾을 수 없음

---

### 2.2 프로젝트 목록 조회 (필터링)
```
GET /api/projects/
```

**Headers**: `Authorization: Bearer {token}`

**Query Parameters**:
- `workspace_id` (UUID, optional): 워크스페이스별 필터링
- `status` (string, optional): 상태별 필터링
- `priority` (string, optional): 우선순위별 필터링
- `limit` (int, optional): 페이지 크기 (기본값: 20, 최대: 100)
- `offset` (int, optional): 시작 위치 (기본값: 0)

**예시**:
```
GET /api/projects/?workspace_id=550e8400-e29b-41d4-a716-446655440000&status=ACTIVE&limit=10
```

**Response** (200 OK):
```json
{
  "total": 3,
  "items": [
    {
      "id": "770e8400-e29b-41d4-a716-446655440000",
      "workspace_id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "백엔드 개발 프로젝트",
      "status": "ACTIVE",
      "priority": "HIGH",
      "created_at": "2025-10-26T00:00:00"
    }
  ],
  "limit": 10,
  "offset": 0
}
```

---

### 2.3 특정 프로젝트 조회
```
GET /api/projects/{project_id}
```

**Headers**: `Authorization: Bearer {token}`

**Response** (200 OK): 2.1과 동일한 프로젝트 객체

---

### 2.4 프로젝트 수정
```
PATCH /api/projects/{project_id}
```

**Headers**: `Authorization: Bearer {token}`

**Request Body** (모든 필드 optional):
```json
{
  "name": "신규 프로젝트명",
  "description": "업데이트된 설명",
  "status": "COMPLETED",
  "priority": "LOW"
}
```

---

### 2.5 프로젝트 삭제
```
DELETE /api/projects/{project_id}
```

**Headers**: `Authorization: Bearer {token}`

**주의**: 프로젝트 삭제 시 연결된 모든 티켓과 태스크가 **애플리케이션 레벨 CASCADE**로 삭제됩니다.

**Response** (204 No Content)

---

## 3. 티켓 (Tickets)

티켓은 프로젝트 내에서 작업 항목을 나타냅니다.

### 3.1 티켓 생성
```
POST /api/tickets/
```

**Headers**: `Authorization: Bearer {token}`

**Request Body**:
```json
{
  "project_id": "770e8400-e29b-41d4-a716-446655440000",
  "title": "사용자 인증 API 구현",
  "description": "JWT 기반 인증 시스템 개발",
  "status": "OPEN",
  "priority": "HIGH",
  "ticket_type_id": "880e8400-e29b-41d4-a716-446655440000",
  "assignee_id": "660e8400-e29b-41d4-a716-446655440000"
}
```

**Field 설명**:
- `project_id` (UUID, required): 프로젝트 ID
- `title` (string, required): 티켓 제목
- `description` (string, optional): 상세 설명
- `status` (string, optional): 티켓 상태 (기본값: "OPEN")
- `priority` (string, optional): 우선순위 (기본값: "MEDIUM")
- `ticket_type_id` (UUID, optional): 티켓 타입 ID
- `assignee_id` (UUID, optional): 담당자 ID

**Response** (201 Created)

---

### 3.2 티켓 목록 조회 (필터링)
```
GET /api/tickets/
```

**Headers**: `Authorization: Bearer {token}`

**Query Parameters**:
- `project_id` (UUID, optional): 프로젝트별 필터링
- `status` (string, optional): 상태별 필터링
- `priority` (string, optional): 우선순위별 필터링
- `limit` (int, optional): 페이지 크기
- `offset` (int, optional): 시작 위치

**예시**:
```
GET /api/tickets/?project_id=770e8400-e29b-41d4-a716-446655440000&status=IN_PROGRESS
```

---

### 3.3 특정 티켓 조회
```
GET /api/tickets/{ticket_id}
```

---

### 3.4 티켓 수정
```
PATCH /api/tickets/{ticket_id}
```

**Request Body** (모든 필드 optional):
```json
{
  "title": "업데이트된 제목",
  "status": "IN_PROGRESS",
  "priority": "URGENT",
  "assignee_id": "990e8400-e29b-41d4-a716-446655440000"
}
```

---

### 3.5 티켓 삭제
```
DELETE /api/tickets/{ticket_id}
```

**주의**: 티켓 삭제 시 연결된 모든 태스크가 **애플리케이션 레벨 CASCADE**로 삭제됩니다.

**Response** (204 No Content)

---

## 4. 태스크 (Tasks)

태스크는 티켓을 작은 단위로 나눈 작업 항목입니다.

### 4.1 태스크 생성
```
POST /api/tasks/
```

**Headers**: `Authorization: Bearer {token}`

**Request Body**:
```json
{
  "ticket_id": "aa0e8400-e29b-41d4-a716-446655440000",
  "title": "데이터베이스 스키마 설계",
  "description": "사용자 테이블 설계",
  "status": "TODO",
  "assignee_id": "660e8400-e29b-41d4-a716-446655440000"
}
```

**Response** (201 Created)

---

### 4.2 태스크 목록 조회 (필터링)
```
GET /api/tasks/
```

**Headers**: `Authorization: Bearer {token}`

**Query Parameters**:
- `ticket_id` (UUID, optional): 티켓별 필터링
- `status` (string, optional): 상태별 필터링
- `limit` (int, optional): 페이지 크기
- `offset` (int, optional): 시작 위치

---

### 4.3 특정 태스크 조회
```
GET /api/tasks/{task_id}
```

---

### 4.4 태스크 수정
```
PATCH /api/tasks/{task_id}
```

**Request Body** (모든 필드 optional):
```json
{
  "title": "업데이트된 제목",
  "status": "IN_PROGRESS",
  "assignee_id": "990e8400-e29b-41d4-a716-446655440000"
}
```

---

### 4.5 태스크 완료 처리
```
PATCH /api/tasks/{task_id}/complete
```

**Headers**: `Authorization: Bearer {token}`

**설명**: 태스크 상태를 "DONE"으로 변경하고 완료 시간을 기록합니다.

**Response** (200 OK):
```json
{
  "id": "bb0e8400-e29b-41d4-a716-446655440000",
  "title": "데이터베이스 스키마 설계",
  "status": "DONE",
  "completed_at": "2025-10-26T12:30:00"
}
```

**에러**:
- `400 Bad Request` - 이미 완료된 태스크

---

### 4.6 태스크 삭제
```
DELETE /api/tasks/{task_id}
```

**Response** (204 No Content)

---

## 5. 알림 (Notifications)

사용자별 알림을 관리합니다.

### 5.1 알림 생성
```
POST /api/notifications/
```

**Headers**: `Authorization: Bearer {token}`

**Request Body**:
```json
{
  "user_id": "660e8400-e29b-41d4-a716-446655440000",
  "notification_type": "TICKET_ASSIGNED",
  "title": "새로운 티켓이 할당되었습니다",
  "message": "백엔드 개발 프로젝트에 새 티켓이 할당되었습니다",
  "target_type": "TICKET",
  "target_id": "aa0e8400-e29b-41d4-a716-446655440000"
}
```

**Response** (201 Created)

---

### 5.2 알림 목록 조회
```
GET /api/notifications/
```

**Headers**: `Authorization: Bearer {token}`

**Query Parameters**:
- `is_read` (boolean, optional): 읽음 여부 필터 (null=전체, true=읽음, false=안읽음)
- `notification_type` (string, optional): 알림 타입 필터
- `limit` (int, optional): 페이지 크기
- `offset` (int, optional): 시작 위치

**Response** (200 OK):
```json
{
  "total": 15,
  "unread_count": 5,
  "items": [
    {
      "id": "cc0e8400-e29b-41d4-a716-446655440000",
      "user_id": "660e8400-e29b-41d4-a716-446655440000",
      "notification_type": "TICKET_ASSIGNED",
      "title": "새로운 티켓이 할당되었습니다",
      "message": "...",
      "is_read": false,
      "created_at": "2025-10-26T10:00:00"
    }
  ],
  "limit": 20,
  "offset": 0
}
```

---

### 5.3 읽지 않은 알림 개수 조회
```
GET /api/notifications/unread-count
```

**Headers**: `Authorization: Bearer {token}`

**Response** (200 OK):
```json
{
  "unread_count": 5
}
```

---

### 5.4 특정 알림 조회
```
GET /api/notifications/{notification_id}
```

**Headers**: `Authorization: Bearer {token}`

---

### 5.5 알림 읽음 처리
```
PATCH /api/notifications/{notification_id}/read
```

**Headers**: `Authorization: Bearer {token}`

**Response** (200 OK):
```json
{
  "id": "cc0e8400-e29b-41d4-a716-446655440000",
  "is_read": true,
  "read_at": "2025-10-26T12:00:00"
}
```

---

### 5.6 모든 알림 읽음 처리
```
POST /api/notifications/mark-all-read
```

**Headers**: `Authorization: Bearer {token}`

**Response** (200 OK):
```json
{
  "message": "5 notifications marked as read",
  "count": 5
}
```

---

### 5.7 알림 삭제
```
DELETE /api/notifications/{notification_id}
```

**Headers**: `Authorization: Bearer {token}`

**주의**: 알림은 소프트 삭제가 아닌 영구 삭제됩니다.

**Response** (204 No Content)

---

## 6. 티켓 타입 (Ticket Types)

프로젝트별로 사용자 정의 티켓 타입을 관리합니다.

### 6.1 티켓 타입 생성
```
POST /api/projects/{project_id}/ticket-types/
```

**Headers**: `Authorization: Bearer {token}`

**Path Parameters**:
- `project_id` (UUID): 프로젝트 ID

**Request Body**:
```json
{
  "type_name": "개발 작업",
  "description": "개발 관련 작업",
  "color": "#3498db",
  "icon": "code",
  "display_order": 1
}
```

**Field 설명**:
- `type_name` (string, required): 타입 이름
- `description` (string, optional): 설명
- `color` (string, optional): HEX 색상 코드 (예: #FF5733)
- `icon` (string, optional): 아이콘 이름 (예: "code", "bug", "palette")
- `display_order` (int, optional): 표시 순서

**Response** (201 Created)

**에러**:
- `404 Not Found` - 프로젝트를 찾을 수 없음
- `400 Bad Request` - 동일한 프로젝트에 같은 이름의 타입이 이미 존재

---

### 6.2 티켓 타입 목록 조회
```
GET /api/projects/{project_id}/ticket-types/
```

**Headers**: `Authorization: Bearer {token}`

**Query Parameters**:
- `include_deleted` (boolean, optional): 삭제된 항목 포함 여부 (기본값: false)
- `limit` (int, optional): 페이지 크기 (기본값: 50)
- `offset` (int, optional): 시작 위치

**Response** (200 OK):
```json
{
  "total": 4,
  "items": [
    {
      "id": "dd0e8400-e29b-41d4-a716-446655440000",
      "project_id": "770e8400-e29b-41d4-a716-446655440000",
      "type_name": "개발 작업",
      "description": "개발 관련 작업",
      "color": "#3498db",
      "icon": "code",
      "display_order": 1,
      "is_deleted": false,
      "created_at": "2025-10-26T00:00:00"
    }
  ],
  "limit": 50,
  "offset": 0
}
```

---

### 6.3 특정 티켓 타입 조회
```
GET /api/projects/{project_id}/ticket-types/{type_id}
```

**Headers**: `Authorization: Bearer {token}`

---

### 6.4 티켓 타입 수정
```
PATCH /api/projects/{project_id}/ticket-types/{type_id}
```

**Headers**: `Authorization: Bearer {token}`

**Request Body** (모든 필드 optional):
```json
{
  "type_name": "백엔드 개발",
  "color": "#2ecc71",
  "display_order": 2
}
```

**에러**:
- `400 Bad Request` - 타입 이름이 프로젝트 내에서 중복됨

---

### 6.5 티켓 타입 삭제 (소프트 삭제)
```
DELETE /api/projects/{project_id}/ticket-types/{type_id}
```

**Headers**: `Authorization: Bearer {token}`

**설명**: `is_deleted` 플래그를 True로 설정합니다. 이미 이 타입을 사용하는 티켓들은 영향받지 않습니다.

**Response** (204 No Content)

**에러**:
- `400 Bad Request` - 이미 삭제된 타입

---

## 7. 헬스 체크 (Health Check)

### 7.1 기본 헬스 체크
```
GET /health
```

**인증 불필요**

**Response** (200 OK):
```json
{
  "status": "healthy",
  "timestamp": "2025-10-26T12:00:00",
  "service": "kanban-service"
}
```

---

### 7.2 Liveness Probe
```
GET /health/live
```

**인증 불필요**

**Response** (200 OK)

---

### 7.3 Readiness Probe
```
GET /health/ready
```

**인증 불필요**

**Response** (200 OK): 데이터베이스 연결 및 Redis 연결 확인

---

## 데이터 타입 및 Enum

### UUID
모든 ID 필드는 UUID 형식입니다.
```
예시: "550e8400-e29b-41d4-a716-446655440000"
```

### 날짜/시간
ISO 8601 형식 (UTC)을 사용합니다.
```
예시: "2025-10-26T12:30:00"
```

### ProjectStatus (프로젝트 상태)
- `PLANNING` - 계획 단계
- `ACTIVE` - 활성화
- `COMPLETED` - 완료
- `ON_HOLD` - 보류
- `CANCELLED` - 취소

### TicketStatus (티켓 상태)
- `OPEN` - 시작 전
- `IN_PROGRESS` - 진행 중
- `REVIEW` - 리뷰 중
- `TESTING` - 테스트 중
- `DONE` - 완료
- `CLOSED` - 종료
- `BLOCKED` - 차단됨

### TaskStatus (태스크 상태)
- `TODO` - 할 일
- `IN_PROGRESS` - 진행 중
- `REVIEW` - 리뷰 중
- `DONE` - 완료

### Priority (우선순위)
- `LOW` - 낮음
- `MEDIUM` - 보통
- `HIGH` - 높음
- `URGENT` - 긴급

### NotificationType (알림 타입)
- `TICKET_CREATED` - 티켓 생성
- `TICKET_UPDATED` - 티켓 업데이트
- `COMMENT_ADDED` - 댓글 추가
- `TICKET_ASSIGNED` - 티켓 할당
- `DUE_DATE_REMINDER` - 마감일 알림
- `MENTION` - 멘션

---

## 에러 코드

| HTTP 상태 코드 | 설명 |
|----------------|------|
| 200 | 성공 |
| 201 | 생성 성공 |
| 204 | 성공 (응답 본문 없음) |
| 400 | 잘못된 요청 (유효성 검증 실패, 중복 등) |
| 401 | 인증 실패 (토큰 없음 또는 만료) |
| 403 | 권한 없음 |
| 404 | 리소스를 찾을 수 없음 |
| 409 | 충돌 (리소스 중복) |
| 500 | 서버 내부 오류 |

---

## 페이지네이션

모든 목록 조회 API는 페이지네이션을 지원합니다.

**Query Parameters**:
- `limit` (int): 한 페이지당 항목 수 (기본값: 20, 최대: 100)
- `offset` (int): 시작 위치 (기본값: 0)

**예시**:
```
# 첫 번째 페이지 (20개)
GET /api/projects/?limit=20&offset=0

# 두 번째 페이지 (20개)
GET /api/projects/?limit=20&offset=20

# 세 번째 페이지 (10개)
GET /api/projects/?limit=10&offset=20
```

---

## 중요 사항

### Trailing Slash 필수
FastAPI의 특성상 모든 엔드포인트에 **마지막 슬래시(/)가 필수**입니다:
- ❌ `/api/workspaces` → 307 Redirect
- ✅ `/api/workspaces/` → 정상 작동

### CASCADE 삭제
- Workspace 삭제 → 모든 Project, Ticket, Task 삭제
- Project 삭제 → 모든 Ticket, Task 삭제
- Ticket 삭제 → 모든 Task 삭제

**주의**: 모든 CASCADE는 **애플리케이션 레벨**에서 처리됩니다 (FK 없음).

### 소프트 삭제 vs 영구 삭제
- **소프트 삭제**: Ticket Types (is_deleted 플래그)
- **영구 삭제**: Workspace, Project, Ticket, Task, Notification

### JWT 토큰
- User Service와 동일한 `JWT_SECRET` 사용
- 토큰은 User Service에서 발급, Kanban Service에서 검증만 수행
- 토큰 만료 시 User Service에서 재발급 필요

---

## 개발 팁

1. **Swagger UI 활용**: `http://localhost:8000/docs`에서 API를 직접 테스트할 수 있습니다.
   - 우측 상단 "Authorize" 버튼 클릭
   - Bearer 토큰 입력 (Bearer 키워드 제외)
   - 모든 API 테스트 가능

2. **토큰 생성**: 로컬 테스트 시 `generate_test_token.py` 스크립트 활용

3. **CORS 설정**:
   - 기본 허용 도메인: `http://localhost:3000`, `http://localhost:8000`
   - 추가 도메인 필요 시 `.env` 파일의 `CORS_ORIGINS` 수정

4. **데이터베이스 접근**:
   ```bash
   docker exec -it wealist-kanban-db psql -U wealist_kanban_user -d wealist_kanban_db
   ```

5. **로그 확인**:
   ```bash
   docker logs -f wealist-kanban-service
   ```

---

## 참고 문서

- **프로젝트 가이드**: `/CLAUDE.md`
- **프로젝트 컨텍스트**: `/.claude/context.md`
- **User Service API**: `/weAlist-User/API_DOCUMENTATION.md`
- **테스트 가이드**: `/kanban-service/services/kanban/API_TEST_GUIDE.md`

---

## 문의

문제가 발생하거나 질문이 있으시면 백엔드 팀에 문의해주세요.

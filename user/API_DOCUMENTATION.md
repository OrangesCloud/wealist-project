# weAlist User Service API Documentation

## 📋 목차
- [개요](#개요)
- [인증](#인증)
- [공통 응답 형식](#공통-응답-형식)
- [API 엔드포인트](#api-엔드포인트)
  - [1. 인증 (Authentication)](#1-인증-authentication)
  - [2. 사용자 (Users)](#2-사용자-users)
  - [3. 그룹 (Groups)](#3-그룹-groups)
  - [4. 팀 (Teams)](#4-팀-teams)
  - [5. 사용자 정보 (UserInfo)](#5-사용자-정보-userinfo)
- [에러 코드](#에러-코드)

---

## 개요

**Base URL**: `http://localhost:8081`

**Content-Type**: `application/json`

**Swagger UI**: `http://localhost:8081/swagger-ui.html`

---

## 인증

대부분의 API는 JWT 토큰 인증이 필요합니다.

### 인증 헤더 형식
```
Authorization: Bearer {access_token}
```

### JWT 인증 에러 응답

토큰 인증 실패 시, 클라이언트는 HTTP `401 Unauthorized` 상태 코드와 함께 다음 형식의 JSON 응답을 받게 됩니다. 프론트엔드에서는 `code` 필드를 사용하여 에러의 원인을 파악하고 후속 조치(예: 토큰 재발급 요청, 로그인 페이지로 리디렉션)를 수행할 수 있습니다.

**응답 형식 예시 (`J002: 토큰 만료`):**
```json
{
  "status": 401,
  "code": "J002",
  "message": "Expired JWT token"
}
```

**주요 JWT 에러 코드:**

| 코드 (code) | 메시지 (message) | 설명 및 클라이언트 조치 방안 |
| :--- | :--- | :--- |
| `J001` | Invalid JWT token | 유효하지 않은 토큰 형식입니다. 토큰이 잘못되었거나 손상되었을 수 있습니다. | 
| `J002` | Expired JWT token | 토큰이 만료되었습니다. 클라이언트는 `POST /api/auth/refresh`를 통해 토큰 재발급을 시도해야 합니다. | 
| `J003` | Unsupported JWT token | 지원되지 않는 형식의 토큰입니다. | 
| `J004` | Malformed JWT token | 구조가 잘못된 토큰입니다. | 
| `J006` | JWT signature is invalid | 서명이 유효하지 않은, 변조되었을 가능성이 있는 토큰입니다. | 
| `J007` | JWT token is blacklisted | 로그아웃 처리된 토큰입니다. 사용자는 다시 로그인해야 합니다. | 

### 인증이 필요 없는 엔드포인트
- `POST /api/auth/signup` - 회원가입
- `POST /api/auth/login` - 로그인
- `POST /api/auth/refresh` - 토큰 갱신
- `GET /health` - 헬스 체크

---

## 공통 응답 형식

### 성공 응답
```json
{
  "success": true,
  "message": "작업이 성공적으로 완료되었습니다.",
  "data": { ... }
}
```

### 실패 응답
```json
{
  "success": false,
  "message": "오류 메시지",
  "data": null
}
```

---

## API 엔드포인트

## 1. 인증 (Authentication)

### 1.1 회원가입
```
POST /api/auth/signup
```

**Request Body**:
```json
{
  "email": "user@example.com",
  "password": "password123",
  "name": "홍길동"
}
```

**Response**:
```json
{
  "success": true,
  "message": "회원가입이 완료되었습니다.",
  "data": {
    "accessToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "tokenType": "Bearer",
    "expiresIn": 1800000
  }
}
```

---

### 1.2 로그인
```
POST /api/auth/login
```

**Request Body**:
```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

**Response**: 회원가입과 동일

---

### 1.3 로그아웃
```
POST /api/auth/logout
```

**Headers**: `Authorization: Bearer {token}`

**Response**:
```json
{
  "success": true,
  "message": "로그아웃이 완료되었습니다."
}
```

---

### 1.4 토큰 갱신
```
POST /api/auth/refresh
```

**Request Body**:
```json
{
  "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Response**:
```json
{
  "success": true,
  "message": "토큰이 갱신되었습니다.",
  "data": {
    "accessToken": "new_access_token...",
    "refreshToken": "new_refresh_token...",
    "tokenType": "Bearer",
    "expiresIn": 1800000
  }
}
```

---

### 1.5 내 정보 조회
```
GET /api/auth/me
```

**Headers**: `Authorization: Bearer {token}`

**Response**:
```json
{
  "success": true,
  "message": "사용자 정보 조회 성공",
  "data": {
    "userId": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "name": "홍길동",
    "groupId": "660e8400-e29b-41d4-a716-446655440000",
    "role": "개발자"
  }
}
```

---

## 2. 사용자 (Users)

### 2.1 활성화된 사용자 목록 조회
```
GET /api/users
```

**Response**:
```json
[
  {
    "userId": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "name": "홍길동",
    "isActive": true,
    "createdAt": "2025-01-01T00:00:00",
    "updatedAt": "2025-01-01T00:00:00"
  }
]
```

---

### 2.2 특정 사용자 조회
```
GET /api/users/{userId}
```

**Path Parameters**:
- `userId` (UUID): 사용자 ID

**Response**: 2.1과 동일한 사용자 객체

---

### 2.3 이메일로 사용자 조회
```
GET /api/users/email/{email}
```

**Path Parameters**:
- `email` (string): 사용자 이메일

---

### 2.4 이름으로 사용자 검색
```
GET /api/users/search?name={name}
```

**Query Parameters**:
- `name` (string): 검색할 이름

**Response**: 사용자 배열

---

### 2.5 이메일 중복 체크
```
GET /api/users/check-email?email={email}
```

**Query Parameters**:
- `email` (string): 확인할 이메일

**Response**:
```json
{
  "success": true,
  "message": "사용 가능한 이메일입니다.",
  "data": true
}
```

---

### 2.6 활성화된 사용자 수 조회
```
GET /api/users/count
```

**Response**:
```json
{
  "count": 150,
  "message": "활성화된 사용자 수"
}
```

---

### 2.7 사용자 정보 수정
```
PUT /api/users/{userId}
```

**Path Parameters**:
- `userId` (UUID): 사용자 ID

**Request Body**:
```json
{
  "name": "김철수",
  "email": "newmail@example.com"
}
```

---

### 2.8 비밀번호 변경
```
PATCH /api/users/{userId}/password
```

**Request Body**:
```json
{
  "currentPassword": "old_password",
  "newPassword": "new_password"
}
```

---

### 2.9 사용자 재활성화
```
PUT /api/users/{userId}/reactivate
```

---

### 2.10 사용자 삭제 (Soft Delete)
```
DELETE /api/users/{userId}
```

---

## 3. 그룹 (Groups)

### 3.1 그룹 생성
```
POST /api/groups
```

**Request Body**:
```json
{
  "name": "개발팀",
  "companyName": "테크컴퍼니"
}
```

**Response**:
```json
{
  "success": true,
  "message": "'테크컴퍼니' 회사의 첫 번째 그룹 '개발팀'이(가) 성공적으로 생성되었습니다.",
  "data": {
    "groupId": "660e8400-e29b-41d4-a716-446655440000",
    "name": "개발팀",
    "companyName": "테크컴퍼니",
    "isDeleted": false,
    "createdAt": "2025-01-01T00:00:00"
  }
}
```

---

### 3.2 활성화된 그룹 목록 조회
```
GET /api/groups
```

**Response**:
```json
{
  "success": true,
  "message": "활성화된 그룹 목록을 성공적으로 조회했습니다.",
  "data": [
    {
      "groupId": "660e8400-e29b-41d4-a716-446655440000",
      "name": "개발팀",
      "companyName": "테크컴퍼니",
      "isDeleted": false
    }
  ]
}
```

---

### 3.3 특정 그룹 조회
```
GET /api/groups/{groupId}
```

---

### 3.4 회사명으로 그룹 조회
```
GET /api/groups/company/{companyName}
```

**Path Parameters**:
- `companyName` (string): 회사명

---

### 3.5 회사별 모든 그룹 조회
```
GET /api/groups/company/{companyName}/all
```

**Response**:
```json
{
  "success": true,
  "message": "'테크컴퍼니' 회사의 그룹 3개를 성공적으로 조회했습니다.",
  "data": [...]
}
```

---

### 3.6 회사명 중복 체크
```
GET /api/groups/check-company/{companyName}
```

**Response**:
```json
{
  "success": true,
  "message": "'테크컴퍼니' 회사명의 그룹이 이미 존재합니다.",
  "data": true
}
```

---

### 3.7 그룹명으로 검색
```
GET /api/groups/search/name?name={name}
```

---

### 3.8 그룹 정보 수정
```
PUT /api/groups/{groupId}
```

**Request Body**:
```json
{
  "name": "신규 개발팀",
  "companyName": "테크컴퍼니"
}
```

---

### 3.9 그룹 삭제 (Soft Delete)
```
DELETE /api/groups/{groupId}
```

---

### 3.10 그룹 재활성화
```
PUT /api/groups/{groupId}/reactivate
```

---

### 3.11 활성화된 그룹 수 조회
```
GET /api/groups/count
```

---

### 3.12 비활성화된 그룹 조회 (관리자)
```
GET /api/groups/inactive
```

---

## 4. 팀 (Teams)

### 4.1 팀 생성
```
POST /api/teams
```

**Request Body**:
```json
{
  "teamName": "백엔드팀",
  "description": "백엔드 개발 담당 팀",
  "groupId": "660e8400-e29b-41d4-a716-446655440000",
  "leaderId": "550e8400-e29b-41d4-a716-446655440000",
  "memberIds": [
    "770e8400-e29b-41d4-a716-446655440000",
    "880e8400-e29b-41d4-a716-446655440000"
  ]
}
```

**Response**:
```json
{
  "success": true,
  "message": "팀이 성공적으로 생성되었습니다.",
  "data": {
    "team": {
      "teamId": "990e8400-e29b-41d4-a716-446655440000",
      "teamName": "백엔드팀",
      "description": "백엔드 개발 담당 팀",
      "groupId": "660e8400-e29b-41d4-a716-446655440000",
      "leaderId": "550e8400-e29b-41d4-a716-446655440000"
    },
    "members": [...]
  }
}
```

---

### 4.2 활성화된 팀 목록 조회
```
GET /api/teams
```

---

### 4.3 특정 팀 조회
```
GET /api/teams/{teamId}
```

---

### 4.4 그룹별 팀 조회
```
GET /api/teams/group/{groupId}
```

---

### 4.5 팀장별 팀 조회
```
GET /api/teams/leader/{leaderId}
```

---

### 4.6 사용자가 속한 팀 조회
```
GET /api/teams/user/{userId}
```

---

### 4.7 사용자가 팀장인 팀 조회
```
GET /api/teams/user/{userId}/led-teams
```

---

### 4.8 팀 멤버 목록 조회
```
GET /api/teams/{teamId}/members
```

**Response**:
```json
{
  "success": true,
  "message": "팀 멤버 목록을 성공적으로 조회했습니다.",
  "data": [
    {
      "userId": "550e8400-e29b-41d4-a716-446655440000",
      "groupId": "660e8400-e29b-41d4-a716-446655440000",
      "role": "개발자"
    }
  ]
}
```

---

### 4.9 팀 멤버 수 조회
```
GET /api/teams/{teamId}/members/count
```

---

### 4.10 팀에 할당 가능한 사용자 목록
```
GET /api/teams/available-users/{groupId}
```

---

### 4.11 팀장 권한 확인
```
GET /api/teams/{teamId}/leader/check?userId={userId}
```

**Query Parameters**:
- `userId` (UUID): 확인할 사용자 ID

**Response**:
```json
{
  "success": true,
  "message": "사용자가 해당 팀의 팀장입니다.",
  "data": true
}
```

---

### 4.12 팀에 멤버 추가
```
POST /api/teams/{teamId}/members?requesterId={requesterId}&userId={userId}&role={role}
```

**Query Parameters**:
- `requesterId` (UUID): 요청자 ID (팀장이어야 함)
- `userId` (UUID): 추가할 사용자 ID
- `role` (string, optional): 역할 (기본값: "팀원")

---

### 4.13 팀 정보 업데이트
```
PUT /api/teams/{teamId}?requesterId={requesterId}&teamName={teamName}&description={description}
```

**Query Parameters**:
- `requesterId` (UUID): 요청자 ID (팀장이어야 함)
- `teamName` (string, optional): 새 팀명
- `description` (string, optional): 새 설명

---

### 4.14 팀장 변경
```
PUT /api/teams/{teamId}/leader?currentLeaderId={currentLeaderId}&newLeaderId={newLeaderId}
```

**Query Parameters**:
- `currentLeaderId` (UUID): 현재 팀장 ID
- `newLeaderId` (UUID): 새 팀장 ID

---

### 4.15 팀원 제거
```
DELETE /api/teams/{teamId}/members/{userId}?requesterId={requesterId}
```

**Query Parameters**:
- `requesterId` (UUID): 요청자 ID (팀장이어야 함)

---

### 4.16 팀 삭제
```
DELETE /api/teams/{teamId}?requesterId={requesterId}
```

---

### 4.17 팀 해산
```
DELETE /api/teams/{teamId}/disband?requesterId={requesterId}
```

---

## 5. 사용자 정보 (UserInfo)

### 5.1 사용자 정보 생성
```
POST /api/userinfo
```

**Request Body**:
```json
{
  "userId": "550e8400-e29b-41d4-a716-446655440000",
  "groupId": "660e8400-e29b-41d4-a716-446655440000",
  "role": "개발자"
}
```

---

### 5.2 활성화된 사용자 정보 목록
```
GET /api/userinfo
```

---

### 5.3 그룹별 사용자 조회
```
GET /api/userinfo/group/{groupId}
```

---

### 5.4 역할별 사용자 조회
```
GET /api/userinfo/role/{role}
```

**Path Parameters**:
- `role` (string): 역할 (예: "개발자", "디자이너", "PM")

---

### 5.5 사용자 정보 수정
```
PUT /api/userinfo/{userId}
```

**Request Body**:
```json
{
  "groupId": "660e8400-e29b-41d4-a716-446655440000",
  "role": "시니어 개발자"
}
```

---

### 5.6 사용자 정보 재활성화
```
PUT /api/userinfo/{userId}/reactivate
```

---

### 5.7 사용자 정보 삭제 (Soft Delete)
```
DELETE /api/userinfo/{userId}
```

---

### 5.8 그룹의 모든 사용자 비활성화
```
DELETE /api/userinfo/group/{groupId}
```

---

### 5.9 활성화된 사용자 수 조회
```
GET /api/userinfo/count
```

---

## 에러 코드

| HTTP 상태 코드 | 설명 |
|----------------|------|
| 200 | 성공 |
| 400 | 잘못된 요청 (유효성 검증 실패) |
| 401 | 인증 실패 (토큰 없음 또는 만료) |
| 403 | 권한 없음 |
| 404 | 리소스를 찾을 수 없음 |
| 500 | 서버 내부 오류 |

---

## 헬스 체크

### 서비스 상태 확인
```
GET /health
```

**Response**:
```json
{
  "checks": {
    "database": "UP"
  },
  "service": "UserRepo",
  "message": "Health check completed",
  "status": "UP",
  "timestamp": "2025-10-25T21:43:07.392604384"
}
```

---

## 주요 데이터 타입

### UUID
모든 ID 필드는 UUID 형식입니다.
```
예시: "550e8400-e29b-41d4-a716-446655440000"
```

### 날짜/시간
ISO 8601 형식을 사용합니다.
```
예시: "2025-01-01T00:00:00"
```

---

## 개발 팁

1. **Swagger UI 활용**: `http://localhost:8081/swagger-ui.html`에서 API를 직접 테스트할 수 있습니다.

2. **토큰 관리**: Access Token은 30분, Refresh Token은 24시간 유효합니다. Refresh Token을 사용해 새 Access Token을 발급받으세요.

3. **Soft Delete**: 삭제된 리소스는 `isDeleted=true`로 표시되며 물리적으로 삭제되지 않습니다.

4. **권한 관리**:
   - 팀 관련 작업은 팀장 권한 필요
   - 사용자 정보 수정은 본인 또는 관리자만 가능

---

## 문의

문제가 발생하거나 질문이 있으시면 백엔드 팀에 문의해주세요.

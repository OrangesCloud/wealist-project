# Swagger API 문서 가이드

## 개요

이 프로젝트는 Swagger/OpenAPI 3.0을 사용하여 API 문서를 자동 생성합니다.

## Swagger UI 접속

서버 실행 후 브라우저에서 다음 URL에 접속하세요:

```
http://localhost:8080/swagger/index.html
```

## 주요 기능

### 1. API 엔드포인트 탐색
- 모든 API 엔드포인트를 카테고리별로 확인
- 각 엔드포인트의 HTTP 메서드, 경로, 설명 확인

### 2. 요청/응답 스키마
- 요청 바디 구조 확인
- 응답 데이터 구조 확인
- 필수/선택 필드 구분

### 3. API 테스트
- "Try it out" 버튼으로 직접 API 호출
- 요청 파라미터 입력
- 실시간 응답 확인

### 4. 예시 데이터
- 각 엔드포인트의 요청/응답 예시 제공
- 실제 사용 가능한 데이터 형식 확인

## API 카테고리

### Projects
- Workspace별 프로젝트 관리
- 프로젝트 생성, 조회, 수정, 삭제
- 프로젝트 검색 (이름/설명)
- 프로젝트 초기 설정 조회
- 기본 프로젝트 조회

### Project Members
- 프로젝트 멤버 목록 조회
- 멤버 제거 (OWNER/ADMIN)
- 멤버 역할 변경 (OWNER)

### Project Join Requests
- 프로젝트 가입 요청 생성
- 가입 요청 목록 조회 (OWNER/ADMIN)
- 가입 요청 승인/거부 (OWNER/ADMIN)

### Boards
- Board CRUD 작업
- Stage, Importance, Role 관리

### Participants
- Board 참여자 관리
- 참여자 추가/제거

### Comments
- Board 댓글 기능
- 댓글 작성/수정/삭제

## Swagger 문서 업데이트

코드 변경 후 Swagger 문서를 재생성하려면:

```bash
make swagger
```

또는

```bash
swag init -g cmd/api/main.go -o docs
```

## Swagger 주석 작성 가이드

핸들러 함수에 다음과 같은 주석을 추가하세요:

```go
// CreateBoard godoc
// @Summary      Board 생성
// @Description  새로운 Board를 생성합니다
// @Tags         boards
// @Accept       json
// @Produce      json
// @Param        request body dto.CreateBoardRequest true "Board 생성 요청"
// @Success      201 {object} response.SuccessResponse{data=dto.BoardResponse}
// @Failure      400 {object} response.ErrorResponse
// @Router       /boards [post]
func (h *BoardHandler) CreateBoard(c *gin.Context) {
    // ...
}
```

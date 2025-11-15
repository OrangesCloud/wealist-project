# 구현 TODO 리스트

## 완료된 작업 ✅

1. ✅ 마이그레이션 파일 생성 (`migrations/002_add_project_members_and_board_fields.sql`)
2. ✅ DTO 업데이트 (board_dto.go, project_dto.go, comment_dto.go)
3. ✅ Domain 모델 업데이트 (board.go, project.go)
4. ✅ Project Repository 업데이트 (멤버 관리, 가입 요청 메서드 추가)
5. ✅ Response 구조 변경 (message → requestId)
6. ✅ Request ID 미들웨어 생성
7. ✅ 마이그레이션 가이드 작성

## 남은 작업 ⏳

### 1. Board Repository 업데이트
**파일**: `internal/repository/board_repository.go`

추가할 메서드:
```go
// 페이지네이션 지원
FindByProjectIDWithPagination(ctx context.Context, projectID uuid.UUID, page, limit int) ([]*domain.Board, int64, error)

// 필드 값 업데이트
UpdateField(ctx context.Context, boardID uuid.UUID, field string, value string) error
```

### 2. Project Service 업데이트
**파일**: `internal/service/project_service.go`

추가할 메서드:
```go
// 검색
Search(ctx context.Context, workspaceID uuid.UUID, query string, page, limit int) (*dto.PaginatedProjectsResponse, error)

// 멤버 관리
GetMembers(ctx context.Context, projectID uuid.UUID) ([]dto.ProjectMemberResponse, error)
UpdateMemberRole(ctx context.Context, projectID, memberID, requesterID uuid.UUID, role string) error
RemoveMember(ctx context.Context, projectID, memberID, requesterID uuid.UUID) error

// 가입 요청
CreateJoinRequest(ctx context.Context, projectID, userID uuid.UUID) (*dto.ProjectJoinRequestResponse, error)
GetJoinRequests(ctx context.Context, projectID, requesterID uuid.UUID, status *string) ([]dto.ProjectJoinRequestResponse, error)
UpdateJoinRequest(ctx context.Context, requestID, requesterID uuid.UUID, status string) (*dto.ProjectJoinRequestResponse, error)
```

권한 체크 로직:
- OWNER: 모든 권한
- ADMIN: 멤버 추가/제거 가능
- MEMBER: 읽기만 가능

### 3. Board Service 업데이트
**파일**: `internal/service/board_service.go`

수정할 메서드:
```go
// Create 메서드에 authorID 추가
Create(ctx context.Context, req *dto.CreateBoardRequest, authorID uuid.UUID) (*dto.BoardResponse, error)

// GetByProjectID에 페이지네이션 추가
GetByProjectID(ctx context.Context, projectID uuid.UUID, page, limit int) (*dto.PaginatedBoardsResponse, error)
```

추가할 메서드:
```go
// 필드 값 변경
UpdateField(ctx context.Context, boardID uuid.UUID, field, value string) error
```

### 4. Project Handler 업데이트
**파일**: `internal/handler/project_handler.go`

추가할 핸들러:
```go
// @Summary Search projects
// @Router /api/projects/search [get]
func (h *ProjectHandler) SearchProjects(c *gin.Context)

// @Summary Get project members
// @Router /api/projects/{projectId}/members [get]
func (h *ProjectHandler) GetMembers(c *gin.Context)

// @Summary Update member role
// @Router /api/projects/{projectId}/members/{memberId}/role [put]
func (h *ProjectHandler) UpdateMemberRole(c *gin.Context)

// @Summary Remove member
// @Router /api/projects/{projectId}/members/{memberId} [delete]
func (h *ProjectHandler) RemoveMember(c *gin.Context)

// @Summary Create join request
// @Router /api/projects/join-requests [post]
func (h *ProjectHandler) CreateJoinRequest(c *gin.Context)

// @Summary Get join requests
// @Router /api/projects/{projectId}/join-requests [get]
func (h *ProjectHandler) GetJoinRequests(c *gin.Context)

// @Summary Update join request
// @Router /api/projects/join-requests/{requestId} [put]
func (h *ProjectHandler) UpdateJoinRequest(c *gin.Context)
```

### 5. Board Handler 업데이트
**파일**: `internal/handler/board_handler.go`

수정할 핸들러:
```go
// CreateBoard - authorID를 컨텍스트에서 가져오기
// GetBoardsByProject - 페이지네이션 파라미터 추가
```

추가할 핸들러:
```go
// @Summary Update board field
// @Router /api/boards/{boardId}/field [put]
func (h *BoardHandler) UpdateField(c *gin.Context)
```

### 6. Response 호출 변경 (모든 Handler)
**참고**: `RESPONSE_MIGRATION.md` 파일 참조

모든 handler 파일에서:
```go
// 변경 전
response.SendSuccess(c, 200, data, "성공 메시지")
response.SendError(c, 404, "CODE", "메시지", "상세")

// 변경 후
response.SendSuccess(c, 200, data)
response.SendError(c, 404, "CODE", "메시지")
```

**파일 목록**:
- `internal/handler/board_handler.go`
- `internal/handler/project_handler.go`
- `internal/handler/comment_handler.go`
- `internal/handler/participant_handler.go`

### 7. Router 업데이트
**파일**: `cmd/api/main.go` 또는 router 파일

Request ID 미들웨어 추가:
```go
r.Use(middleware.RequestID())
```

Base path 변경:
```go
// 변경 전
api := r.Group("/api/v1")

// 변경 후
api := r.Group("/api")
```

새로운 라우트 추가:
```go
// Projects
projects.GET("/search", projectHandler.SearchProjects)
projects.GET("/:projectId/members", projectHandler.GetMembers)
projects.PUT("/:projectId/members/:memberId/role", projectHandler.UpdateMemberRole)
projects.DELETE("/:projectId/members/:memberId", projectHandler.RemoveMember)
projects.POST("/join-requests", projectHandler.CreateJoinRequest)
projects.GET("/:projectId/join-requests", projectHandler.GetJoinRequests)
projects.PUT("/join-requests/:requestId", projectHandler.UpdateJoinRequest)

// Boards
boards.PUT("/:boardId/field", boardHandler.UpdateField)

// Comments - 경로 변경
comments.GET("", commentHandler.GetCommentsByBoard) // query param: boardId
```

### 8. 테스트 업데이트

모든 테스트 파일에서:
1. JSON 필드명을 camelCase로 변경
2. 새로운 필드 추가 (authorId, assigneeId, ownerId 등)
3. Response 구조 변경 (message → requestId)
4. 새로운 엔드포인트 테스트 추가

**파일 목록**:
- `internal/handler/board_handler_test.go`
- `internal/handler/project_handler_test.go`
- `internal/handler/comment_handler_test.go`
- `internal/service/board_service_test.go`
- `internal/service/project_service_test.go`
- `internal/service/comment_service_test.go`

### 9. Swagger 문서 재생성

Swagger 주석에서 response 구조 변경:
```go
// @Success 200 {object} response.SuccessResponse{data=dto.BoardResponse}
```

```bash
# Swagger 주석 업데이트 후
swag init -g cmd/api/main.go -o docs
```

## 우선순위

1. **High Priority** (핵심 기능):
   - ⚠️ **Response 호출 변경** (모든 handler에서 필수!)
   - Router에 Request ID 미들웨어 추가
   - Board Repository 페이지네이션
   - Project Service 멤버 관리
   - Project Service 가입 요청
   - Project Handler 새 엔드포인트
   - Router base path 변경

2. **Medium Priority**:
   - Board Service 필드 업데이트
   - Board Handler 필드 업데이트
   - Project Service 검색

3. **Low Priority**:
   - 테스트 업데이트
   - Swagger 문서 재생성

## 실행 순서

```bash
# 1. 마이그레이션 실행
psql -U your_user -d your_database -f migrations/002_add_project_members_and_board_fields.sql

# 2. 코드 변경 후 빌드
go build -o bin/api cmd/api/main.go

# 3. 테스트 실행
go test ./...

# 4. Swagger 재생성
swag init -g cmd/api/main.go -o docs

# 5. 서버 실행
./bin/api
```

## 참고사항

- 모든 새로운 API는 인증 미들웨어 적용 필요
- 권한 체크는 service 레이어에서 수행
- 에러 메시지는 명확하게 작성
- 페이지네이션 기본값: page=1, limit=20

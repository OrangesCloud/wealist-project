# Response 구조 변경 가이드

## 변경 사항

### 1. Response 구조 변경

**변경 전**:
```json
{
  "data": { ... },
  "message": "성공 메시지"
}
```

**변경 후**:
```json
{
  "data": { ... },
  "requestId": "550e8400-e29b-41d4-a716-446655440000"
}
```

### 2. Error Response 변경

**변경 전**:
```json
{
  "error": {
    "code": "BOARD_NOT_FOUND",
    "message": "Board를 찾을 수 없습니다",
    "details": "상세 정보"
  },
  "message": "Board를 찾을 수 없습니다"
}
```

**변경 후**:
```json
{
  "error": {
    "code": "BOARD_NOT_FOUND",
    "message": "Board를 찾을 수 없습니다"
  },
  "requestId": "550e8400-e29b-41d4-a716-446655440000"
}
```

## 코드 변경 방법

### Handler에서 Response 호출 변경

**변경 전**:
```go
response.SendSuccess(c, http.StatusOK, board, "Board 조회 성공")
response.SendError(c, http.StatusNotFound, "BOARD_NOT_FOUND", "Board를 찾을 수 없습니다", "")
```

**변경 후**:
```go
// 방법 1: 새로운 방식 (권장)
response.SendSuccess(c, http.StatusOK, board)
response.SendError(c, http.StatusNotFound, "BOARD_NOT_FOUND", "Board를 찾을 수 없습니다")

// 방법 2: 메시지를 data에 포함 (삭제 성공 등)
response.SendSuccessMessage(c, http.StatusOK, nil, "Board 삭제 성공")
// 결과: {"data": {"message": "Board 삭제 성공"}, "requestId": "..."}
```

### 모든 Handler 파일에서 변경 필요

1. **board_handler.go**
   - `CreateBoard`: `SendSuccess(c, 201, board, "Board 생성 성공")` → `SendSuccess(c, 201, board)`
   - `GetBoard`: `SendSuccess(c, 200, board, "Board 조회 성공")` → `SendSuccess(c, 200, board)`
   - `UpdateBoard`: `SendSuccess(c, 200, board, "Board 수정 성공")` → `SendSuccess(c, 200, board)`
   - `DeleteBoard`: `SendSuccess(c, 200, nil, "Board 삭제 성공")` → `SendSuccessMessage(c, 200, nil, "Board 삭제 성공")`
   - 모든 `SendError` 호출에서 마지막 `details` 파라미터 제거

2. **project_handler.go**
   - 동일한 패턴으로 변경

3. **comment_handler.go**
   - 동일한 패턴으로 변경

4. **participant_handler.go**
   - 동일한 패턴으로 변경

### Request ID 미들웨어 추가

**main.go 또는 router 설정**:
```go
import "project-board-api/internal/middleware"

func main() {
    r := gin.Default()
    
    // Request ID 미들웨어 추가 (가장 먼저)
    r.Use(middleware.RequestID())
    
    // 나머지 미들웨어...
    r.Use(middleware.CORS())
    // ...
}
```

## 자동 변경 스크립트 (참고용)

모든 handler 파일에서 일괄 변경이 필요합니다:

```bash
# SendSuccess 호출 변경 (message 파라미터 제거)
# 수동으로 확인하면서 변경하는 것을 권장합니다

# SendError 호출 변경 (details 파라미터 제거)
# 마찬가지로 수동 확인 권장
```

## 테스트 업데이트

모든 테스트에서 응답 구조 변경:

**변경 전**:
```go
assert.Equal(t, "Board 생성 성공", response["message"])
```

**변경 후**:
```go
assert.NotEmpty(t, response["requestId"])
// message는 더 이상 최상위에 없음
```

## 체크리스트

- [ ] `internal/response/response.go` 업데이트 ✅
- [ ] `internal/middleware/request_id.go` 생성 ✅
- [ ] `cmd/api/main.go`에 RequestID 미들웨어 추가
- [ ] `internal/handler/board_handler.go` 업데이트
- [ ] `internal/handler/project_handler.go` 업데이트
- [ ] `internal/handler/comment_handler.go` 업데이트
- [ ] `internal/handler/participant_handler.go` 업데이트
- [ ] 모든 테스트 파일 업데이트
- [ ] Swagger 주석 업데이트

## 예시: board_handler.go 변경

**변경 전**:
```go
func (h *BoardHandler) CreateBoard(c *gin.Context) {
    // ... 로직 ...
    response.SendSuccess(c, http.StatusCreated, boardResp, "Board 생성 성공")
}

func (h *BoardHandler) GetBoard(c *gin.Context) {
    // ... 로직 ...
    if board == nil {
        response.SendError(c, http.StatusNotFound, "BOARD_NOT_FOUND", "Board를 찾을 수 없습니다", "")
        return
    }
    response.SendSuccess(c, http.StatusOK, board, "Board 조회 성공")
}
```

**변경 후**:
```go
func (h *BoardHandler) CreateBoard(c *gin.Context) {
    // ... 로직 ...
    response.SendSuccess(c, http.StatusCreated, boardResp)
}

func (h *BoardHandler) GetBoard(c *gin.Context) {
    // ... 로직 ...
    if board == nil {
        response.SendError(c, http.StatusNotFound, "BOARD_NOT_FOUND", "Board를 찾을 수 없습니다")
        return
    }
    response.SendSuccess(c, http.StatusOK, board)
}
```

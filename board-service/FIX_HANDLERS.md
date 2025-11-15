# Handler 파일 수정 가이드

## 문제
response.SendError와 SendSuccess 호출에서 파라미터 개수가 변경되었습니다.

## 수정 방법

### SendError 수정
**변경 전**:
```go
response.SendError(c, http.StatusBadRequest, response.ErrCodeValidation, "Invalid request body", err.Error())
```

**변경 후**:
```go
response.SendError(c, http.StatusBadRequest, response.ErrCodeValidation, "Invalid request body")
```

### SendSuccess 수정
**변경 전**:
```go
response.SendSuccess(c, http.StatusOK, data, "Success message")
```

**변경 후**:
```go
response.SendSuccess(c, http.StatusOK, data)
```

### 삭제 성공 등 메시지만 필요한 경우
**변경 후**:
```go
response.SendSuccessMessage(c, http.StatusOK, nil, "Board deleted successfully")
```

## 수정이 필요한 파일들

1. internal/handler/board_handler.go
2. internal/handler/project_handler.go
3. internal/handler/comment_handler.go
4. internal/handler/participant_handler.go
5. internal/handler/error_handler.go ✅ (완료)

## 자동 수정 스크립트 (주의해서 사용)

```bash
# 백업 먼저!
cp -r internal/handler internal/handler.backup

# SendError 수정 (마지막 파라미터 제거)
for file in internal/handler/*.go; do
  # err.Error() 파라미터 제거
  perl -i -pe 's/response\.SendError\(([^,]+),\s*([^,]+),\s*([^,]+),\s*([^,]+),\s*err\.Error\(\)\)/response.SendError($1, $2, $3, $4)/g' "$file"
  
  # 빈 문자열 파라미터 제거
  perl -i -pe 's/response\.SendError\(([^,]+),\s*([^,]+),\s*([^,]+),\s*([^,]+),\s*""\)/response.SendError($1, $2, $3, $4)/g' "$file"
done

# SendSuccess 수정 (마지막 파라미터 제거)
for file in internal/handler/*.go; do
  perl -i -pe 's/response\.SendSuccess\(([^,]+),\s*([^,]+),\s*([^,]+),\s*"[^"]*"\)/response.SendSuccess($1, $2, $3)/g' "$file"
done
```

## 수동 수정 예시

### board_handler.go 라인 39
```go
// 변경 전
response.SendError(c, http.StatusBadRequest, response.ErrCodeValidation, "Invalid request body", err.Error())

// 변경 후
response.SendError(c, http.StatusBadRequest, response.ErrCodeValidation, "Invalid request body")
```

### board_handler.go 라인 49
```go
// 변경 전
response.SendSuccess(c, http.StatusCreated, board, "Board created successfully")

// 변경 후
response.SendSuccess(c, http.StatusCreated, board)
```

### board_handler.go 라인 165 (DeleteBoard)
```go
// 변경 전
response.SendSuccess(c, http.StatusOK, nil, "Board deleted successfully")

// 변경 후
response.SendSuccessMessage(c, http.StatusOK, nil, "Board deleted successfully")
```

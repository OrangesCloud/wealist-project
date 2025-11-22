package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"project-board-api/internal/database"
	"project-board-api/internal/dto"
	"project-board-api/internal/response"
	"project-board-api/internal/service"
)

type BoardHandler struct {
	boardService service.BoardService
}

func NewBoardHandler(boardService service.BoardService) *BoardHandler {
	return &BoardHandler{
		boardService: boardService,
	}
}

// CreateBoard godoc
// @Summary      Board 생성
// @Description  새로운 Board를 생성합니다
// @Description  customFields는 value 기반 인터페이스를 사용합니다 (UUID 아님)
// @Description  유효한 필드 타입: stage, role, importance
// @Description  예시 값: stage="in_progress", role="developer", importance="high"
// @Description  잘못된 field value 제공 시 400 에러 반환
// @Description  assigneeId가 제공되지 않으면 자동으로 authorId로 설정됩니다
// @Description  startDate와 dueDate는 선택 사항이며, startDate는 dueDate보다 이전이어야 합니다
// @Tags         boards
// @Accept       json
// @Produce      json
// @Param        request body dto.CreateBoardRequest true "Board 생성 요청"
// @Success      201 {object} response.SuccessResponse{data=dto.BoardResponse} "Board 생성 성공"
// @Failure      400 {object} response.ErrorResponse "잘못된 요청 또는 유효하지 않은 field value"
// @Failure      404 {object} response.ErrorResponse "Project를 찾을 수 없음"
// @Failure      500 {object} response.ErrorResponse "서버 에러"
// @Router       /boards [post]
func (h *BoardHandler) CreateBoard(c *gin.Context) {
	var req dto.CreateBoardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.SendError(c, http.StatusBadRequest, response.ErrCodeValidation, "Invalid request body")
		return
	}

	ctx := c.Request.Context()
	if userID, exists := c.Get("user_id"); exists {
		ctx = context.WithValue(ctx, "user_id", userID)
	}

	board, err := h.boardService.CreateBoard(ctx, &req)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	// 💡 [수정] 응답을 먼저 보낸 후 브로드캐스트
	response.SendSuccess(c, http.StatusCreated, board)

	// 💡 [추가] 브로드캐스트
	event := WSEvent{
		Type:    "BOARD_CREATED",
		BoardID: board.ID.String(),
		Payload: board,
	}
	BroadcastEvent(req.ProjectID.String(), event)
}

// GetBoard godoc
// @Summary      Board 상세 조회
// @Description  Board ID로 상세 정보를 조회합니다 (참여자, 댓글, 첨부파일 포함)
// @Description  응답의 customFields는 value 기반 (UUID가 아닌 문자열 값)
// @Description  예시: {"importance": "high", "role": "developer", "stage": "in_progress"}
// @Description  participantIds는 보드에 참여하는 사용자 ID 배열입니다
// @Description  attachments는 보드에 첨부된 파일 메타데이터 배열입니다
// @Description  startDate와 dueDate는 설정된 경우에만 포함됩니다
// @Tags         boards
// @Produce      json
// @Param        boardId path string true "Board ID (UUID)"
// @Success      200 {object} response.SuccessResponse{data=dto.BoardDetailResponse} "Board 조회 성공"
// @Failure      400 {object} response.ErrorResponse "잘못된 Board ID"
// @Failure      404 {object} response.ErrorResponse "Board를 찾을 수 없음"
// @Failure      500 {object} response.ErrorResponse "서버 에러"
// @Router       /boards/{boardId} [get]
func (h *BoardHandler) GetBoard(c *gin.Context) {
	boardIDStr := c.Param("boardId")
	boardID, err := uuid.Parse(boardIDStr)
	if err != nil {
		response.SendError(c, http.StatusBadRequest, response.ErrCodeValidation, "Invalid board ID")
		return
	}

	board, err := h.boardService.GetBoard(c.Request.Context(), boardID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response.SendSuccess(c, http.StatusOK, board)
}

// GetBoardsByProject godoc
// @Summary      Project의 Board 목록 조회
// @Description  특정 Project에 속한 모든 Board를 조회합니다. customFields 파라미터로 필터링 가능 (JSON 형식)
// @Description  응답의 customFields는 value 기반 (UUID가 아닌 문자열 값)
// @Description  예시: {"importance": "high", "role": "developer", "stage": "in_progress"}
// @Description  각 보드는 participantIds (참여자 ID 배열)와 attachments (첨부파일 메타데이터 배열)를 포함합니다
// @Description  startDate와 dueDate는 설정된 경우에만 포함됩니다
// @Tags         boards
// @Produce      json
// @Param        projectId    path      string  true   "Project ID (UUID)"
// @Param        customFields query     string  false  "Custom Fields 필터 JSON 객체. 예시: {\"importance\":\"high\",\"stage\":\"in_progress\"}"
// @Success      200 {object} response.SuccessResponse{data=[]dto.BoardResponse} "Board 목록 조회 성공"
// @Failure      400 {object} response.ErrorResponse "잘못된 Project ID 또는 필터 파라미터"
// @Failure      404 {object} response.ErrorResponse "Project를 찾을 수 없음"
// @Failure      500 {object} response.ErrorResponse "서버 에러"
// @Router       /boards/project/{projectId} [get]
func (h *BoardHandler) GetBoardsByProject(c *gin.Context) {
	projectIDStr := c.Param("projectId")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		response.SendError(c, http.StatusBadRequest, response.ErrCodeValidation, "Invalid project ID")
		return
	}

	filters := &dto.BoardFilters{}
	customFieldsStr := c.Query("customFields")

	if customFieldsStr != "" {
		var customFields map[string]interface{}
		if err := json.Unmarshal([]byte(customFieldsStr), &customFields); err != nil {
			response.SendError(c, http.StatusBadRequest, response.ErrCodeValidation, "Invalid customFields format: must be valid JSON")
			return
		}
		filters.CustomFields = customFields
	}

	boards, err := h.boardService.GetBoardsByProject(c.Request.Context(), projectID, filters)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response.SendSuccess(c, http.StatusOK, boards)
}

// GetBoardsByProjectQuery godoc
// @Summary      Project의 Board 목록 조회 (쿼리 파라미터 방식)
// @Description  특정 Project에 속한 모든 Board를 조회합니다. 프론트엔드 호환용 엔드포인트
// @Description  응답의 customFields는 value 기반 (UUID가 아닌 문자열 값)
// @Description  예시: {"importance": "high", "role": "developer", "stage": "in_progress"}
// @Description  각 보드는 participantIds (참여자 ID 배열)와 attachments (첨부파일 메타데이터 배열)를 포함합니다
// @Description  startDate와 dueDate는 설정된 경우에만 포함됩니다
// @Tags         boards
// @Produce      json
// @Param        projectId    query     string  true   "Project ID (UUID)"
// @Param        customFields query     string  false  "Custom Fields 필터 JSON 객체. 예시: {\"importance\":\"high\",\"stage\":\"in_progress\"}"
// @Success      200 {object} response.SuccessResponse{data=[]dto.BoardResponse} "Board 목록 조회 성공"
// @Failure      400 {object} response.ErrorResponse "잘못된 Project ID 또는 필터 파라미터"
// @Failure      404 {object} response.ErrorResponse "Project를 찾을 수 없음"
// @Failure      500 {object} response.ErrorResponse "서버 에러"
// @Router       /boards [get]
func (h *BoardHandler) GetBoardsByProjectQuery(c *gin.Context) {
	projectIDStr := c.Query("projectId")
	if projectIDStr == "" {
		response.SendError(c, http.StatusBadRequest, response.ErrCodeValidation, "Project ID is required")
		return
	}

	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		response.SendError(c, http.StatusBadRequest, response.ErrCodeValidation, "Invalid project ID")
		return
	}

	filters := &dto.BoardFilters{}
	customFieldsStr := c.Query("customFields")

	if customFieldsStr != "" {
		var customFields map[string]interface{}
		if err := json.Unmarshal([]byte(customFieldsStr), &customFields); err != nil {
			response.SendError(c, http.StatusBadRequest, response.ErrCodeValidation, "Invalid customFields format: must be valid JSON")
			return
		}
		filters.CustomFields = customFields
	}

	boards, err := h.boardService.GetBoardsByProject(c.Request.Context(), projectID, filters)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response.SendSuccess(c, http.StatusOK, boards)
}

// UpdateBoard godoc
// @Summary      Board 수정
// @Description  Board 정보를 수정합니다 (제목, 내용, 단계, 중요도, 역할, 담당자, 날짜)
// @Description  customFields는 value 기반 인터페이스를 사용합니다 (UUID 아님)
// @Description  유효한 필드 타입: stage, role, importance
// @Description  예시 값: stage="completed", role="designer", importance="medium"
// @Description  잘못된 field value 제공 시 400 에러 반환
// @Description  startDate와 dueDate를 수정할 수 있으며, startDate는 dueDate보다 이전이어야 합니다
// @Tags         boards
// @Accept       json
// @Produce      json
// @Param        boardId path string true "Board ID (UUID)"
// @Param        request body dto.UpdateBoardRequest true "Board 수정 요청"
// @Success      200 {object} response.SuccessResponse{data=dto.BoardResponse} "Board 수정 성공"
// @Failure      400 {object} response.ErrorResponse "잘못된 요청 또는 유효하지 않은 field value"
// @Failure      404 {object} response.ErrorResponse "Board를 찾을 수 없음"
// @Failure      500 {object} response.ErrorResponse "서버 에러"
// @Router       /boards/{boardId} [put]
func (h *BoardHandler) UpdateBoard(c *gin.Context) {
	boardIDStr := c.Param("boardId")
	boardID, err := uuid.Parse(boardIDStr)
	if err != nil {
		response.SendError(c, http.StatusBadRequest, response.ErrCodeValidation, "Invalid board ID")
		return
	}

	var req dto.UpdateBoardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.SendError(c, http.StatusBadRequest, response.ErrCodeValidation, "Invalid request body")
		return
	}

	board, err := h.boardService.UpdateBoard(c.Request.Context(), boardID, &req)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	// 💡 [수정] 응답을 먼저 보낸 후 브로드캐스트
	response.SendSuccess(c, http.StatusOK, board)

	// 💡 [추가] 브로드캐스트
	event := WSEvent{
		Type:    "BOARD_UPDATED",
		BoardID: boardID.String(),
		Payload: board,
	}
	BroadcastEvent(board.ProjectID.String(), event)
}

// DeleteBoard godoc
// @Summary      Board 삭제
// @Description  Board를 소프트 삭제합니다
// @Tags         boards
// @Produce      json
// @Param        boardId path string true "Board ID (UUID)"
// @Success      200 {object} response.SuccessResponse "Board 삭제 성공"
// @Failure      400 {object} response.ErrorResponse "잘못된 Board ID"
// @Failure      404 {object} response.ErrorResponse "Board를 찾을 수 없음"
// @Failure      500 {object} response.ErrorResponse "서버 에러"
// @Router       /boards/{boardId} [delete]
func (h *BoardHandler) DeleteBoard(c *gin.Context) {
	boardIDStr := c.Param("boardId")
	boardID, err := uuid.Parse(boardIDStr)
	if err != nil {
		response.SendError(c, http.StatusBadRequest, response.ErrCodeValidation, "Invalid board ID")
		return
	}

	// 💡 [수정] 삭제 전에 보드 정보 가져오기 (projectId 필요)
	board, err := h.boardService.GetBoard(c.Request.Context(), boardID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	err = h.boardService.DeleteBoard(c.Request.Context(), boardID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	// 💡 [수정] 응답을 먼저 보낸 후 브로드캐스트
	response.SendSuccess(c, http.StatusOK, nil)

	// 💡 [추가] 브로드캐스트
	event := WSEvent{
		Type:    "BOARD_DELETED",
		BoardID: boardID.String(),
		Payload: map[string]string{
			"boardId": boardID.String(),
		},
	}
	BroadcastEvent(board.ProjectID.String(), event)
}

// MoveBoard godoc
// @Summary      Board 이동 (실시간 동기화)
// @Description  Board를 다른 컬럼으로 이동합니다. WebSocket을 통해 실시간으로 다른 클라이언트에게 전파됩니다
// @Description  groupByFieldName에 해당하는 필드의 값을 newFieldValue로 변경합니다
// @Tags         boards
// @Accept       json
// @Produce      json
// @Param        boardId path string true "Board ID (UUID)"
// @Param        request body dto.MoveBoardRequest true "Board 이동 요청"
// @Success      200 {object} response.SuccessResponse{data=dto.MoveBoardResponse} "Board 이동 성공"
// @Failure      400 {object} response.ErrorResponse "잘못된 요청"
// @Failure      404 {object} response.ErrorResponse "Board를 찾을 수 없음"
// @Failure      500 {object} response.ErrorResponse "서버 에러"
// @Router       /boards/{boardId}/move [put]
func (h *BoardHandler) MoveBoard(c *gin.Context) {
	boardIDStr := c.Param("boardId")
	boardID, err := uuid.Parse(boardIDStr)
	if err != nil {
		response.SendError(c, http.StatusBadRequest, response.ErrCodeValidation, "Invalid board ID")
		return
	}

	var req dto.MoveBoardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.SendError(c, http.StatusBadRequest, response.ErrCodeValidation, "Invalid request body")
		return
	}

	ctx := c.Request.Context()

	// 🔥 [수정] req.ProjectID를 UUID로 파싱
	projectID, err := uuid.Parse(req.ProjectID)
	if err != nil {
		response.SendError(c, http.StatusBadRequest, response.ErrCodeValidation, "Invalid project ID")
		return
	}

	// 🔥 [수정] newFieldValue가 nil이면 에러
	if req.NewFieldValue == nil {
		response.SendError(c, http.StatusBadRequest, response.ErrCodeValidation, "newFieldValue is required")
		return
	}

	newFieldValue := *req.NewFieldValue // 🔥 포인터 역참조

	// 1. 기존 보드 가져오기
	board, err := h.boardService.GetBoard(ctx, boardID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	// 현재 그룹 값 추출
	oldGroupValue := ""
	if board.CustomFields != nil {
		if val, ok := board.CustomFields[req.GroupByFieldName]; ok {
			oldGroupValue = val.(string)
		}
	}

	// 🔥 [핵심 수정] 기존 customFields를 유지하면서 해당 필드만 업데이트
	existingCustomFields := make(map[string]interface{})
	if board.CustomFields != nil {
		// 기존 필드들을 모두 복사
		for k, v := range board.CustomFields {
			existingCustomFields[k] = v
		}
	}
	// 변경할 필드만 업데이트
	existingCustomFields[req.GroupByFieldName] = newFieldValue

	// 2. 필드 값 업데이트
	updateReq := &dto.UpdateBoardRequest{
		CustomFields: &existingCustomFields,
	}
	_, err = h.boardService.UpdateBoard(ctx, boardID, updateReq)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	// 3. Redis 순서 업데이트
	redisClient := database.GetRedis()
	if redisClient != nil {
		oldKey := fmt.Sprintf("kanban:project:%s:group:%s", projectID.String(), oldGroupValue)
		newKey := fmt.Sprintf("kanban:project:%s:group:%s", projectID.String(), newFieldValue)

		redisClient.ZRem(context.Background(), oldKey, boardID.String())
		redisClient.ZAdd(context.Background(), newKey, &redis.Z{
			Score:  float64(time.Now().UnixMilli()),
			Member: boardID.String(),
		})
	}

	// 4. 실시간 브로드캐스트
	event := WSEvent{
		Type:    "BOARD_MOVED",
		BoardID: boardID.String(),
		Payload: map[string]string{
			"from": oldGroupValue,
			"to":   newFieldValue,
		},
	}

	log := getLogger(c)
	log.Info("Broadcasting BOARD_MOVED event",
		zap.String("projectId", projectID.String()),
		zap.String("boardId", boardID.String()),
		zap.String("from", oldGroupValue),
		zap.String("to", newFieldValue))

	BroadcastEvent(projectID.String(), event)

	// 응답
	response.SendSuccess(c, http.StatusOK, dto.MoveBoardResponse{
		BoardID:       boardID.String(),
		NewFieldValue: newFieldValue,
		Message:       "Board moved successfully",
	})
}
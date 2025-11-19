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

// CreateBoard - Board 생성
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

// GetBoard - Board 상세 조회
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

// GetBoardsByProject - Project의 Board 목록 조회
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

// GetBoardsByProjectQuery - Project의 Board 목록 조회 (쿼리 파라미터)
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

// UpdateBoard - Board 수정
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
		Payload: board, // ✅ board 변수 사용
	}
	BroadcastEvent(board.ProjectID.String(), event)
}

// DeleteBoard - Board 삭제
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

// MoveBoard - 카드 이동 (실시간 반영)
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

	// 2. 필드 값 업데이트
	updateReq := &dto.UpdateBoardRequest{
		CustomFields: &map[string]interface{}{
			req.GroupByFieldName: req.NewFieldValue,
		},
	}
	_, err = h.boardService.UpdateBoard(ctx, boardID, updateReq)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	// 3. Redis 순서 업데이트
	redisClient := database.GetRedis()
	if redisClient != nil {
		oldKey := fmt.Sprintf("kanban:project:%s:group:%s", req.ProjectID.String(), oldGroupValue)
		newKey := fmt.Sprintf("kanban:project:%s:group:%s", req.ProjectID.String(), req.NewFieldValue)

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
			"to":   req.NewFieldValue,
		},
	}

	log := getLogger(c)
	log.Info("Broadcasting BOARD_MOVED event",
		zap.String("projectId", req.ProjectID.String()),
		zap.String("boardId", boardID.String()),
		zap.String("from", oldGroupValue),
		zap.String("to", req.NewFieldValue))

	BroadcastEvent(req.ProjectID.String(), event)

	// 응답
	response.SendSuccess(c, http.StatusOK, dto.MoveBoardResponse{
		BoardID:       boardID.String(),
		NewFieldValue: req.NewFieldValue,
		Message:       "Board moved successfully",
	})
}

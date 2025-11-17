package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

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

	// Create context with user_id from Gin context
	ctx := c.Request.Context()
	if userID, exists := c.Get("user_id"); exists {
		ctx = context.WithValue(ctx, "user_id", userID)
	}

	board, err := h.boardService.CreateBoard(ctx, &req)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response.SendSuccess(c, http.StatusCreated, board)
}

// GetBoard godoc
// @Summary      Board 상세 조회
// @Description  Board ID로 상세 정보를 조회합니다 (참여자, 댓글 포함)
// @Description  응답의 customFields는 value 기반 (UUID가 아닌 문자열 값)
// @Description  예시: {"importance": "high", "role": "developer", "stage": "in_progress"}
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
// @Tags         boards
// @Produce      json
// @Param        projectId path string true "Project ID (UUID)"
// @Param        customFields query string false "Custom Fields 필터 JSON 객체. 예시: {\"stage\":\"in_progress\",\"role\":\"developer\"}"
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

	// Parse customFields filter parameter
	filters := &dto.BoardFilters{}
	customFieldsStr := c.Query("customFields")
	
	if customFieldsStr != "" {
		// Parse JSON string to map
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
// @Tags         boards
// @Produce      json
// @Param        projectId query string true "Project ID (UUID)"
// @Param        customFields query string false "Custom Fields 필터 JSON 객체. 예시: {\"stage\":\"in_progress\",\"role\":\"developer\"}"
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

	// Parse customFields filter parameter
	filters := &dto.BoardFilters{}
	customFieldsStr := c.Query("customFields")
	
	if customFieldsStr != "" {
		// Parse JSON string to map
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
// @Description  Board 정보를 수정합니다 (제목, 내용, 단계, 중요도, 역할)
// @Description  customFields는 value 기반 인터페이스를 사용합니다 (UUID 아님)
// @Description  유효한 필드 타입: stage, role, importance
// @Description  예시 값: stage="completed", role="designer", importance="medium"
// @Description  잘못된 field value 제공 시 400 에러 반환
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

	response.SendSuccess(c, http.StatusOK, board)
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

	err = h.boardService.DeleteBoard(c.Request.Context(), boardID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response.SendSuccess(c, http.StatusOK, nil)
}

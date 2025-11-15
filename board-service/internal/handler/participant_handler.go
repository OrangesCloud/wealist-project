package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"project-board-api/internal/dto"
	"project-board-api/internal/response"
	"project-board-api/internal/service"
)

type ParticipantHandler struct {
	participantService service.ParticipantService
}

func NewParticipantHandler(participantService service.ParticipantService) *ParticipantHandler {
	return &ParticipantHandler{
		participantService: participantService,
	}
}

// AddParticipant godoc
// @Summary      Participant 추가
// @Description  Board에 참여자를 추가합니다
// @Tags         participants
// @Accept       json
// @Produce      json
// @Param        request body dto.AddParticipantRequest true "Participant 추가 요청"
// @Success      201 {object} response.SuccessResponse "Participant 추가 성공"
// @Failure      400 {object} response.ErrorResponse "잘못된 요청"
// @Failure      404 {object} response.ErrorResponse "Board를 찾을 수 없음"
// @Failure      409 {object} response.ErrorResponse "이미 참여 중인 사용자"
// @Failure      500 {object} response.ErrorResponse "서버 에러"
// @Router       /participants [post]
func (h *ParticipantHandler) AddParticipant(c *gin.Context) {
	var req dto.AddParticipantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.SendError(c, http.StatusBadRequest, response.ErrCodeValidation, "Invalid request body")
		return
	}

	err := h.participantService.AddParticipant(c.Request.Context(), &req)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response.SendSuccess(c, http.StatusCreated, nil)
}

// GetParticipants godoc
// @Summary      Board의 Participant 목록 조회
// @Description  특정 Board의 모든 참여자를 조회합니다
// @Tags         participants
// @Produce      json
// @Param        boardId path string true "Board ID (UUID)"
// @Success      200 {object} response.SuccessResponse{data=[]dto.ParticipantResponse} "Participant 목록 조회 성공"
// @Failure      400 {object} response.ErrorResponse "잘못된 Board ID"
// @Failure      404 {object} response.ErrorResponse "Board를 찾을 수 없음"
// @Failure      500 {object} response.ErrorResponse "서버 에러"
// @Router       /participants/board/{boardId} [get]
func (h *ParticipantHandler) GetParticipants(c *gin.Context) {
	boardIDStr := c.Param("boardId")
	boardID, err := uuid.Parse(boardIDStr)
	if err != nil {
		response.SendError(c, http.StatusBadRequest, response.ErrCodeValidation, "Invalid board ID")
		return
	}

	participants, err := h.participantService.GetParticipants(c.Request.Context(), boardID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response.SendSuccess(c, http.StatusOK, participants)
}

// RemoveParticipant godoc
// @Summary      Participant 제거
// @Description  Board에서 참여자를 제거합니다
// @Tags         participants
// @Produce      json
// @Param        boardId path string true "Board ID (UUID)"
// @Param        userId path string true "User ID (UUID)"
// @Success      200 {object} response.SuccessResponse "Participant 제거 성공"
// @Failure      400 {object} response.ErrorResponse "잘못된 ID"
// @Failure      404 {object} response.ErrorResponse "Board 또는 Participant를 찾을 수 없음"
// @Failure      500 {object} response.ErrorResponse "서버 에러"
// @Router       /participants/board/{boardId}/user/{userId} [delete]
func (h *ParticipantHandler) RemoveParticipant(c *gin.Context) {
	boardIDStr := c.Param("boardId")
	boardID, err := uuid.Parse(boardIDStr)
	if err != nil {
		response.SendError(c, http.StatusBadRequest, response.ErrCodeValidation, "Invalid board ID")
		return
	}

	userIDStr := c.Param("userId")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.SendError(c, http.StatusBadRequest, response.ErrCodeValidation, "Invalid user ID")
		return
	}

	err = h.participantService.RemoveParticipant(c.Request.Context(), boardID, userID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response.SendSuccess(c, http.StatusOK, nil)
}

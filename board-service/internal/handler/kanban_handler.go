package handler

import (
	"board-service/internal/apperrors"
	"board-service/internal/dto"
	"board-service/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type KanbanHandler struct {
	service service.KanbanService
}

func NewKanbanHandler(service service.KanbanService) *KanbanHandler {
	return &KanbanHandler{service: service}
}

// CreateKanban godoc
// @Summary      Create kanban
// @Description  Create a new kanban card (task/issue) in a project
// @Tags         kanbans
// @Accept       json
// @Produce      json
// @Param        request body dto.CreateKanbanRequest true "Kanban details"
// @Success      201 {object} dto.SuccessResponse{data=dto.KanbanResponse}
// @Failure      400 {object} dto.ErrorResponse
// @Failure      403 {object} dto.ErrorResponse
// @Failure      404 {object} dto.ErrorResponse
// @Router       /api/kanbans [post]
// @Security     BearerAuth
func (h *KanbanHandler) CreateKanban(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		dto.Error(c, apperrors.ErrUnauthorized)
		return
	}

	var req dto.CreateKanbanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.Error(c, apperrors.Wrap(err, apperrors.ErrCodeValidation, "입력값 검증 실패", 400))
		return
	}

	kanban, err := h.service.CreateKanban(userID, &req)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			dto.Error(c, appErr)
		} else {
			dto.Error(c, apperrors.ErrInternalServer)
		}
		return
	}

	dto.SuccessWithStatus(c, http.StatusCreated, kanban)
}

// GetKanban godoc
// @Summary      Get kanban
// @Description  Get a specific kanban by ID (project member only)
// @Tags         kanbans
// @Accept       json
// @Produce      json
// @Param        id path string true "Kanban ID"
// @Success      200 {object} dto.SuccessResponse{data=dto.KanbanResponse}
// @Failure      403 {object} dto.ErrorResponse
// @Failure      404 {object} dto.ErrorResponse
// @Router       /api/kanbans/{id} [get]
// @Security     BearerAuth
func (h *KanbanHandler) GetKanban(c *gin.Context) {
	userID := c.GetString("user_id")
	kanbanID := c.Param("id")

	kanban, err := h.service.GetKanban(kanbanID, userID)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			dto.Error(c, appErr)
		} else {
			dto.Error(c, apperrors.ErrInternalServer)
		}
		return
	}

	dto.Success(c, kanban)
}

// GetKanbans godoc
// @Summary      Get kanbans
// @Description  Get kanbans for a project with optional filters
// @Tags         kanbans
// @Accept       json
// @Produce      json
// @Param        projectId query string true "Project ID"
// @Param        stageId query string false "Filter by Stage ID"
// @Param        roleId query string false "Filter by Role ID"
// @Param        importanceId query string false "Filter by Importance ID"
// @Param        assigneeId query string false "Filter by Assignee ID"
// @Param        authorId query string false "Filter by Author ID"
// @Param        page query int false "Page number (default: 1)"
// @Param        limit query int false "Items per page (default: 20, max: 100)"
// @Success      200 {object} dto.SuccessResponse{data=dto.PaginatedKanbansResponse}
// @Failure      400 {object} dto.ErrorResponse
// @Failure      403 {object} dto.ErrorResponse
// @Router       /api/kanbans [get]
// @Security     BearerAuth
func (h *KanbanHandler) GetKanbans(c *gin.Context) {
	userID := c.GetString("user_id")

	var req dto.GetKanbansRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		dto.Error(c, apperrors.Wrap(err, apperrors.ErrCodeValidation, "입력값 검증 실패", 400))
		return
	}

	// Default pagination
	if req.Page == 0 {
		req.Page = 1
	}
	if req.Limit == 0 {
		req.Limit = 20
	}

	kanbans, err := h.service.GetKanbans(userID, &req)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			dto.Error(c, appErr)
		} else {
			dto.Error(c, apperrors.ErrInternalServer)
		}
		return
	}

	dto.Success(c, kanbans)
}

// UpdateKanban godoc
// @Summary      Update kanban
// @Description  Update a kanban (author or ADMIN+ only)
// @Tags         kanbans
// @Accept       json
// @Produce      json
// @Param        id path string true "Kanban ID"
// @Param        request body dto.UpdateKanbanRequest true "Kanban updates"
// @Success      200 {object} dto.SuccessResponse{data=dto.KanbanResponse}
// @Failure      400 {object} dto.ErrorResponse
// @Failure      403 {object} dto.ErrorResponse
// @Failure      404 {object} dto.ErrorResponse
// @Router       /api/kanbans/{id} [put]
// @Security     BearerAuth
func (h *KanbanHandler) UpdateKanban(c *gin.Context) {
	userID := c.GetString("user_id")
	kanbanID := c.Param("id")

	var req dto.UpdateKanbanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.Error(c, apperrors.Wrap(err, apperrors.ErrCodeValidation, "입력값 검증 실패", 400))
		return
	}

	kanban, err := h.service.UpdateKanban(kanbanID, userID, &req)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			dto.Error(c, appErr)
		} else {
			dto.Error(c, apperrors.ErrInternalServer)
		}
		return
	}

	dto.Success(c, kanban)
}

// DeleteKanban godoc
// @Summary      Delete kanban
// @Description  Delete a kanban (soft delete, author or ADMIN+ only)
// @Tags         kanbans
// @Accept       json
// @Produce      json
// @Param        id path string true "Kanban ID"
// @Success      200 {object} dto.SuccessResponse
// @Failure      403 {object} dto.ErrorResponse
// @Failure      404 {object} dto.ErrorResponse
// @Router       /api/kanbans/{id} [delete]
// @Security     BearerAuth
func (h *KanbanHandler) DeleteKanban(c *gin.Context) {
	userID := c.GetString("user_id")
	kanbanID := c.Param("id")

	if err := h.service.DeleteKanban(kanbanID, userID); err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			dto.Error(c, appErr)
		} else {
			dto.Error(c, apperrors.ErrInternalServer)
		}
		return
	}

	dto.Success(c, gin.H{"message": "칸반이 삭제되었습니다"})
}

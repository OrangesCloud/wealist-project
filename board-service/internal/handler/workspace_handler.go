package handler

import (
	"board-service/internal/apperrors"
	"board-service/internal/dto"
	"board-service/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type WorkspaceHandler struct {
	service service.WorkspaceService
}

func NewWorkspaceHandler(service service.WorkspaceService) *WorkspaceHandler {
	return &WorkspaceHandler{service: service}
}

// CreateWorkspace godoc
// @Summary      Create workspace
// @Description  Create a new workspace (OWNER role automatically assigned)
// @Tags         workspaces
// @Accept       json
// @Produce      json
// @Param        request body dto.CreateWorkspaceRequest true "Workspace details"
// @Success      201 {object} dto.SuccessResponse{data=dto.WorkspaceResponse}
// @Failure      400 {object} dto.ErrorResponse
// @Failure      401 {object} dto.ErrorResponse
// @Failure      500 {object} dto.ErrorResponse
// @Router       /api/workspaces [post]
// @Security     BearerAuth
func (h *WorkspaceHandler) CreateWorkspace(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		dto.Error(c, apperrors.ErrUnauthorized)
		return
	}

	var req dto.CreateWorkspaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.Error(c, apperrors.Wrap(err, apperrors.ErrCodeValidation, "입력값 검증 실패", 400))
		return
	}

	workspace, err := h.service.CreateWorkspace(userID, &req)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			dto.Error(c, appErr)
		} else {
			dto.Error(c, apperrors.ErrInternalServer)
		}
		return
	}

	dto.SuccessWithStatus(c, http.StatusCreated, workspace)
}

// GetWorkspace godoc
// @Summary      Get workspace
// @Description  Get workspace details (member only)
// @Tags         workspaces
// @Accept       json
// @Produce      json
// @Param        id path string true "Workspace ID"
// @Success      200 {object} dto.SuccessResponse{data=dto.WorkspaceResponse}
// @Failure      403 {object} dto.ErrorResponse
// @Failure      404 {object} dto.ErrorResponse
// @Router       /api/workspaces/{id} [get]
// @Security     BearerAuth
func (h *WorkspaceHandler) GetWorkspace(c *gin.Context) {
	userID := c.GetString("user_id")
	workspaceID := c.Param("id")

	workspace, err := h.service.GetWorkspace(workspaceID, userID)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			dto.Error(c, appErr)
		} else {
			dto.Error(c, apperrors.ErrInternalServer)
		}
		return
	}

	dto.Success(c, workspace)
}

// UpdateWorkspace godoc
// @Summary      Update workspace
// @Description  Update workspace details (OWNER only)
// @Tags         workspaces
// @Accept       json
// @Produce      json
// @Param        id path string true "Workspace ID"
// @Param        request body dto.UpdateWorkspaceRequest true "Updated workspace details"
// @Success      200 {object} dto.SuccessResponse{data=dto.WorkspaceResponse}
// @Failure      400 {object} dto.ErrorResponse
// @Failure      403 {object} dto.ErrorResponse
// @Failure      404 {object} dto.ErrorResponse
// @Router       /api/workspaces/{id} [put]
// @Security     BearerAuth
func (h *WorkspaceHandler) UpdateWorkspace(c *gin.Context) {
	userID := c.GetString("user_id")
	workspaceID := c.Param("id")

	var req dto.UpdateWorkspaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.Error(c, apperrors.Wrap(err, apperrors.ErrCodeValidation, "입력값 검증 실패", 400))
		return
	}

	workspace, err := h.service.UpdateWorkspace(workspaceID, userID, &req)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			dto.Error(c, appErr)
		} else {
			dto.Error(c, apperrors.ErrInternalServer)
		}
		return
	}

	dto.Success(c, workspace)
}

// DeleteWorkspace godoc
// @Summary      Delete workspace
// @Description  Soft delete a workspace (OWNER only)
// @Tags         workspaces
// @Accept       json
// @Produce      json
// @Param        id path string true "Workspace ID"
// @Success      200 {object} dto.SuccessResponse{data=object{message=string}}
// @Failure      403 {object} dto.ErrorResponse
// @Failure      404 {object} dto.ErrorResponse
// @Router       /api/workspaces/{id} [delete]
// @Security     BearerAuth
func (h *WorkspaceHandler) DeleteWorkspace(c *gin.Context) {
	userID := c.GetString("user_id")
	workspaceID := c.Param("id")

	if err := h.service.DeleteWorkspace(workspaceID, userID); err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			dto.Error(c, appErr)
		} else {
			dto.Error(c, apperrors.ErrInternalServer)
		}
		return
	}

	dto.Success(c, map[string]string{"message": "워크스페이스가 삭제되었습니다"})
}

// SearchWorkspaces godoc
// @Summary      Search workspaces
// @Description  Search workspaces by name or description
// @Tags         workspaces
// @Accept       json
// @Produce      json
// @Param        query query string true "Search query"
// @Param        page query int false "Page number (default: 1)"
// @Param        limit query int false "Page size (default: 10, max: 100)"
// @Success      200 {object} dto.SuccessResponse{data=dto.PaginatedWorkspacesResponse}
// @Failure      400 {object} dto.ErrorResponse
// @Router       /api/workspaces/search [get]
// @Security     BearerAuth
func (h *WorkspaceHandler) SearchWorkspaces(c *gin.Context) {
	var req dto.SearchWorkspacesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		dto.Error(c, apperrors.Wrap(err, apperrors.ErrCodeValidation, "입력값 검증 실패", 400))
		return
	}

	result, err := h.service.SearchWorkspaces(&req)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			dto.Error(c, appErr)
		} else {
			dto.Error(c, apperrors.ErrInternalServer)
		}
		return
	}

	dto.Success(c, result)
}

// CreateJoinRequest godoc
// @Summary      Create join request
// @Description  Request to join a workspace
// @Tags         workspaces
// @Accept       json
// @Produce      json
// @Param        request body dto.CreateJoinRequestRequest true "Join request details"
// @Success      201 {object} dto.SuccessResponse{data=dto.JoinRequestResponse}
// @Failure      400 {object} dto.ErrorResponse
// @Failure      404 {object} dto.ErrorResponse
// @Failure      409 {object} dto.ErrorResponse
// @Router       /api/workspaces/join-requests [post]
// @Security     BearerAuth
func (h *WorkspaceHandler) CreateJoinRequest(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		dto.Error(c, apperrors.ErrUnauthorized)
		return
	}

	var req dto.CreateJoinRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.Error(c, apperrors.Wrap(err, apperrors.ErrCodeValidation, "입력값 검증 실패", 400))
		return
	}

	joinReq, err := h.service.CreateJoinRequest(userID, &req)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			dto.Error(c, appErr)
		} else {
			dto.Error(c, apperrors.ErrInternalServer)
		}
		return
	}

	dto.SuccessWithStatus(c, http.StatusCreated, joinReq)
}

// GetJoinRequests godoc
// @Summary      Get join requests
// @Description  Get join requests for a workspace (OWNER/ADMIN only)
// @Tags         workspaces
// @Accept       json
// @Produce      json
// @Param        id path string true "Workspace ID"
// @Param        status query string false "Filter by status (PENDING/APPROVED/REJECTED)"
// @Success      200 {object} dto.SuccessResponse{data=[]dto.JoinRequestResponse}
// @Failure      403 {object} dto.ErrorResponse
// @Failure      404 {object} dto.ErrorResponse
// @Router       /api/workspaces/{id}/join-requests [get]
// @Security     BearerAuth
func (h *WorkspaceHandler) GetJoinRequests(c *gin.Context) {
	userID := c.GetString("user_id")
	workspaceID := c.Param("id")
	status := c.Query("status")

	requests, err := h.service.GetJoinRequests(workspaceID, userID, status)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			dto.Error(c, appErr)
		} else {
			dto.Error(c, apperrors.ErrInternalServer)
		}
		return
	}

	dto.Success(c, requests)
}

// UpdateJoinRequest godoc
// @Summary      Update join request
// @Description  Approve or reject a join request (OWNER/ADMIN only)
// @Tags         workspaces
// @Accept       json
// @Produce      json
// @Param        id path string true "Join Request ID"
// @Param        request body dto.UpdateJoinRequestRequest true "Status update"
// @Success      200 {object} dto.SuccessResponse{data=dto.JoinRequestResponse}
// @Failure      400 {object} dto.ErrorResponse
// @Failure      403 {object} dto.ErrorResponse
// @Failure      404 {object} dto.ErrorResponse
// @Router       /api/workspaces/join-requests/{id} [put]
// @Security     BearerAuth
func (h *WorkspaceHandler) UpdateJoinRequest(c *gin.Context) {
	userID := c.GetString("user_id")
	requestID := c.Param("id")

	var req dto.UpdateJoinRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.Error(c, apperrors.Wrap(err, apperrors.ErrCodeValidation, "입력값 검증 실패", 400))
		return
	}

	joinReq, err := h.service.UpdateJoinRequest(requestID, userID, &req)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			dto.Error(c, appErr)
		} else {
			dto.Error(c, apperrors.ErrInternalServer)
		}
		return
	}

	dto.Success(c, joinReq)
}

// GetWorkspaceMembers godoc
// @Summary      Get workspace members
// @Description  Get all members of a workspace (member only)
// @Tags         workspaces
// @Accept       json
// @Produce      json
// @Param        id path string true "Workspace ID"
// @Success      200 {object} dto.SuccessResponse{data=[]dto.WorkspaceMemberResponse}
// @Failure      403 {object} dto.ErrorResponse
// @Failure      404 {object} dto.ErrorResponse
// @Router       /api/workspaces/{id}/members [get]
// @Security     BearerAuth
func (h *WorkspaceHandler) GetWorkspaceMembers(c *gin.Context) {
	userID := c.GetString("user_id")
	workspaceID := c.Param("id")

	members, err := h.service.GetWorkspaceMembers(workspaceID, userID)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			dto.Error(c, appErr)
		} else {
			dto.Error(c, apperrors.ErrInternalServer)
		}
		return
	}

	dto.Success(c, members)
}

// UpdateMemberRole godoc
// @Summary      Update member role
// @Description  Update a member's role in workspace (OWNER only)
// @Tags         workspaces
// @Accept       json
// @Produce      json
// @Param        id path string true "Workspace ID"
// @Param        memberId path string true "Member ID"
// @Param        request body dto.UpdateMemberRoleRequest true "New role"
// @Success      200 {object} dto.SuccessResponse{data=dto.WorkspaceMemberResponse}
// @Failure      400 {object} dto.ErrorResponse
// @Failure      403 {object} dto.ErrorResponse
// @Failure      404 {object} dto.ErrorResponse
// @Router       /api/workspaces/{id}/members/{memberId}/role [put]
// @Security     BearerAuth
func (h *WorkspaceHandler) UpdateMemberRole(c *gin.Context) {
	userID := c.GetString("user_id")
	workspaceID := c.Param("id")
	memberID := c.Param("memberId")

	var req dto.UpdateMemberRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.Error(c, apperrors.Wrap(err, apperrors.ErrCodeValidation, "입력값 검증 실패", 400))
		return
	}

	member, err := h.service.UpdateMemberRole(workspaceID, memberID, userID, &req)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			dto.Error(c, appErr)
		} else {
			dto.Error(c, apperrors.ErrInternalServer)
		}
		return
	}

	dto.Success(c, member)
}

// RemoveMember godoc
// @Summary      Remove member
// @Description  Remove a member from workspace (OWNER/ADMIN only, cannot remove OWNER or self)
// @Tags         workspaces
// @Accept       json
// @Produce      json
// @Param        id path string true "Workspace ID"
// @Param        memberId path string true "Member ID"
// @Success      200 {object} dto.SuccessResponse{data=object{message=string}}
// @Failure      400 {object} dto.ErrorResponse
// @Failure      403 {object} dto.ErrorResponse
// @Failure      404 {object} dto.ErrorResponse
// @Router       /api/workspaces/{id}/members/{memberId} [delete]
// @Security     BearerAuth
func (h *WorkspaceHandler) RemoveMember(c *gin.Context) {
	userID := c.GetString("user_id")
	workspaceID := c.Param("id")
	memberID := c.Param("memberId")

	if err := h.service.RemoveMember(workspaceID, memberID, userID); err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			dto.Error(c, appErr)
		} else {
			dto.Error(c, apperrors.ErrInternalServer)
		}
		return
	}

	dto.Success(c, map[string]string{"message": "멤버가 삭제되었습니다"})
}

// SetDefaultWorkspace godoc
// @Summary      Set default workspace
// @Description  Set a workspace as default for the current user
// @Tags         workspaces
// @Accept       json
// @Produce      json
// @Param        request body dto.SetDefaultWorkspaceRequest true "Workspace ID"
// @Success      200 {object} dto.SuccessResponse{data=object{message=string}}
// @Failure      400 {object} dto.ErrorResponse
// @Failure      403 {object} dto.ErrorResponse
// @Router       /api/workspaces/default [post]
// @Security     BearerAuth
func (h *WorkspaceHandler) SetDefaultWorkspace(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		dto.Error(c, apperrors.ErrUnauthorized)
		return
	}

	var req dto.SetDefaultWorkspaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.Error(c, apperrors.Wrap(err, apperrors.ErrCodeValidation, "입력값 검증 실패", 400))
		return
	}

	if err := h.service.SetDefaultWorkspace(userID, &req); err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			dto.Error(c, appErr)
		} else {
			dto.Error(c, apperrors.ErrInternalServer)
		}
		return
	}

	dto.Success(c, map[string]string{"message": "대표 워크스페이스가 설정되었습니다"})
}

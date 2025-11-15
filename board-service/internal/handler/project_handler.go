package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"project-board-api/internal/dto"
	"project-board-api/internal/response"
	"project-board-api/internal/service"
)

type ProjectHandler struct {
	projectService service.ProjectService
}

func NewProjectHandler(projectService service.ProjectService) *ProjectHandler {
	return &ProjectHandler{
		projectService: projectService,
	}
}

// CreateProject godoc
// @Summary      Project 생성
// @Description  새로운 Project를 생성합니다
// @Tags         projects
// @Accept       json
// @Produce      json
// @Param        request body dto.CreateProjectRequest true "Project 생성 요청"
// @Success      201 {object} response.SuccessResponse{data=dto.ProjectResponse} "Project 생성 성공"
// @Failure      400 {object} response.ErrorResponse "잘못된 요청"
// @Failure      500 {object} response.ErrorResponse "서버 에러"
// @Router       /projects [post]
func (h *ProjectHandler) CreateProject(c *gin.Context) {
	var req dto.CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.SendError(c, http.StatusBadRequest, response.ErrCodeValidation, "Invalid request body")
		return
	}

	// Extract user ID from context (set by Auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		response.SendError(c, http.StatusUnauthorized, response.ErrCodeUnauthorized, "User ID not found in context")
		return
	}
	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		response.SendError(c, http.StatusUnauthorized, response.ErrCodeUnauthorized, "Invalid user ID format")
		return
	}

	// Extract JWT token from context (set by Auth middleware)
	token, exists := c.Get("jwtToken")
	if !exists {
		response.SendError(c, http.StatusUnauthorized, response.ErrCodeUnauthorized, "JWT token not found in context")
		return
	}
	tokenStr, ok := token.(string)
	if !ok {
		response.SendError(c, http.StatusUnauthorized, response.ErrCodeUnauthorized, "Invalid token format")
		return
	}

	project, err := h.projectService.CreateProject(c.Request.Context(), &req, userUUID, tokenStr)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response.SendSuccess(c, http.StatusCreated, project)
}

// GetProjectsByWorkspace godoc
// @Summary      Workspace의 Project 목록 조회
// @Description  특정 Workspace에 속한 모든 Project를 조회합니다
// @Tags         projects
// @Produce      json
// @Param        workspaceId path string true "Workspace ID (UUID)"
// @Success      200 {object} response.SuccessResponse{data=[]dto.ProjectResponse} "Project 목록 조회 성공"
// @Failure      400 {object} response.ErrorResponse "잘못된 Workspace ID"
// @Failure      500 {object} response.ErrorResponse "서버 에러"
// @Router       /projects/workspace/{workspaceId} [get]
func (h *ProjectHandler) GetProjectsByWorkspace(c *gin.Context) {
	workspaceIDStr := c.Param("workspaceId")
	workspaceID, err := uuid.Parse(workspaceIDStr)
	if err != nil {
		response.SendError(c, http.StatusBadRequest, response.ErrCodeValidation, "Invalid workspace ID")
		return
	}

	// Extract user ID from context (set by Auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		response.SendError(c, http.StatusUnauthorized, response.ErrCodeUnauthorized, "User ID not found in context")
		return
	}
	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		response.SendError(c, http.StatusUnauthorized, response.ErrCodeUnauthorized, "Invalid user ID format")
		return
	}

	// Extract JWT token from context (set by Auth middleware)
	token, exists := c.Get("jwtToken")
	if !exists {
		response.SendError(c, http.StatusUnauthorized, response.ErrCodeUnauthorized, "JWT token not found in context")
		return
	}
	tokenStr, ok := token.(string)
	if !ok {
		response.SendError(c, http.StatusUnauthorized, response.ErrCodeUnauthorized, "Invalid token format")
		return
	}

	projects, err := h.projectService.GetProjectsByWorkspace(c.Request.Context(), workspaceID, userUUID, tokenStr)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response.SendSuccess(c, http.StatusOK, projects)
}

// GetDefaultProject godoc
// @Summary      Workspace의 기본 Project 조회
// @Description  특정 Workspace의 기본(default) Project를 조회합니다
// @Tags         projects
// @Produce      json
// @Param        workspaceId path string true "Workspace ID (UUID)"
// @Success      200 {object} response.SuccessResponse{data=dto.ProjectResponse} "기본 Project 조회 성공"
// @Failure      400 {object} response.ErrorResponse "잘못된 Workspace ID"
// @Failure      404 {object} response.ErrorResponse "기본 Project를 찾을 수 없음"
// @Failure      500 {object} response.ErrorResponse "서버 에러"
// @Router       /projects/workspace/{workspaceId}/default [get]
func (h *ProjectHandler) GetDefaultProject(c *gin.Context) {
	workspaceIDStr := c.Param("workspaceId")
	workspaceID, err := uuid.Parse(workspaceIDStr)
	if err != nil {
		response.SendError(c, http.StatusBadRequest, response.ErrCodeValidation, "Invalid workspace ID")
		return
	}

	// Extract user ID from context (set by Auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		response.SendError(c, http.StatusUnauthorized, response.ErrCodeUnauthorized, "User ID not found in context")
		return
	}
	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		response.SendError(c, http.StatusUnauthorized, response.ErrCodeUnauthorized, "Invalid user ID format")
		return
	}

	// Extract JWT token from context (set by Auth middleware)
	token, exists := c.Get("jwtToken")
	if !exists {
		response.SendError(c, http.StatusUnauthorized, response.ErrCodeUnauthorized, "JWT token not found in context")
		return
	}
	tokenStr, ok := token.(string)
	if !ok {
		response.SendError(c, http.StatusUnauthorized, response.ErrCodeUnauthorized, "Invalid token format")
		return
	}

	project, err := h.projectService.GetDefaultProject(c.Request.Context(), workspaceID, userUUID, tokenStr)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response.SendSuccess(c, http.StatusOK, project)
}

package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"project-board-api/internal/client"
	"project-board-api/internal/domain"
	"project-board-api/internal/dto"
	"project-board-api/internal/repository"
	"project-board-api/internal/response"
)

// ProjectService defines the interface for project business logic
type ProjectService interface {
	CreateProject(ctx context.Context, req *dto.CreateProjectRequest, userID uuid.UUID, token string) (*dto.ProjectResponse, error)
	GetProjectsByWorkspace(ctx context.Context, workspaceID, userID uuid.UUID, token string) ([]*dto.ProjectResponse, error)
	GetDefaultProject(ctx context.Context, workspaceID, userID uuid.UUID, token string) (*dto.ProjectResponse, error)
}

// projectServiceImpl is the implementation of ProjectService
type projectServiceImpl struct {
	projectRepo repository.ProjectRepository
	userClient  client.UserClient
}

// NewProjectService creates a new instance of ProjectService
func NewProjectService(projectRepo repository.ProjectRepository, userClient client.UserClient) ProjectService {
	return &projectServiceImpl{
		projectRepo: projectRepo,
		userClient:  userClient,
	}
}

// CreateProject creates a new project
func (s *projectServiceImpl) CreateProject(ctx context.Context, req *dto.CreateProjectRequest, userID uuid.UUID, token string) (*dto.ProjectResponse, error) {
	// Validate workspace membership
	isValid, err := s.userClient.ValidateWorkspaceMember(ctx, req.WorkspaceID, userID, token)
	if err != nil {
		// Log error but continue with graceful degradation
		// Return forbidden error if validation explicitly fails
		return nil, response.NewAppError(response.ErrCodeForbidden, "You are not a member of this workspace", "")
	}
	if !isValid {
		return nil, response.NewAppError(response.ErrCodeForbidden, "You are not a member of this workspace", "")
	}

	// Create domain model from request
	project := &domain.Project{
		WorkspaceID: req.WorkspaceID,
		OwnerID:     userID,
		Name:        req.Name,
		Description: req.Description,
		IsDefault:   false, // Default to false, can be changed later
		IsPublic:    false, // Default to private
	}

	// Save to repository
	if err := s.projectRepo.Create(ctx, project); err != nil {
		return nil, response.NewAppError(response.ErrCodeInternal, "Failed to create project", err.Error())
	}

	// Convert to response DTO
	return s.toProjectResponse(project), nil
}

// GetProjectsByWorkspace retrieves all projects for a workspace
func (s *projectServiceImpl) GetProjectsByWorkspace(ctx context.Context, workspaceID, userID uuid.UUID, token string) ([]*dto.ProjectResponse, error) {
	// Validate workspace membership
	isValid, err := s.userClient.ValidateWorkspaceMember(ctx, workspaceID, userID, token)
	if err != nil {
		// Log error but continue with graceful degradation
		// Return forbidden error if validation explicitly fails
		return nil, response.NewAppError(response.ErrCodeForbidden, "You are not a member of this workspace", "")
	}
	if !isValid {
		return nil, response.NewAppError(response.ErrCodeForbidden, "You are not a member of this workspace", "")
	}

	// Fetch projects from repository
	projects, err := s.projectRepo.FindByWorkspaceID(ctx, workspaceID)
	if err != nil {
		return nil, response.NewAppError(response.ErrCodeInternal, "Failed to fetch projects", err.Error())
	}

	// Convert to response DTOs with owner profile information
	responses := make([]*dto.ProjectResponse, len(projects))
	for i, project := range projects {
		responses[i] = s.toProjectResponseWithProfile(ctx, project, token)
	}

	return responses, nil
}

// GetDefaultProject retrieves the default project for a workspace
func (s *projectServiceImpl) GetDefaultProject(ctx context.Context, workspaceID, userID uuid.UUID, token string) (*dto.ProjectResponse, error) {
	// Validate workspace membership
	isValid, err := s.userClient.ValidateWorkspaceMember(ctx, workspaceID, userID, token)
	if err != nil {
		// Log error but continue with graceful degradation
		// Return forbidden error if validation explicitly fails
		return nil, response.NewAppError(response.ErrCodeForbidden, "You are not a member of this workspace", "")
	}
	if !isValid {
		return nil, response.NewAppError(response.ErrCodeForbidden, "You are not a member of this workspace", "")
	}

	// Fetch default project from repository
	project, err := s.projectRepo.FindDefaultByWorkspaceID(ctx, workspaceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, response.NewAppError(response.ErrCodeNotFound, "Default project not found", "")
		}
		return nil, response.NewAppError(response.ErrCodeInternal, "Failed to fetch default project", err.Error())
	}

	// Convert to response DTO with owner profile information
	return s.toProjectResponseWithProfile(ctx, project, token), nil
}

// toProjectResponse converts domain.Project to dto.ProjectResponse
func (s *projectServiceImpl) toProjectResponse(project *domain.Project) *dto.ProjectResponse {
	return &dto.ProjectResponse{
		ID:          project.ID,
		WorkspaceID: project.WorkspaceID,
		OwnerID:     project.OwnerID,
		Name:        project.Name,
		Description: project.Description,
		IsPublic:    project.IsPublic,
		CreatedAt:   project.CreatedAt,
		UpdatedAt:   project.UpdatedAt,
	}
}

// toProjectResponseWithProfile converts domain.Project to dto.ProjectResponse with owner profile
func (s *projectServiceImpl) toProjectResponseWithProfile(ctx context.Context, project *domain.Project, token string) *dto.ProjectResponse {
	response := s.toProjectResponse(project)

	// Fetch workspace profile for owner
	profile, err := s.userClient.GetWorkspaceProfile(ctx, project.WorkspaceID, project.OwnerID, token)
	if err == nil && profile != nil {
		// Include profile information if available
		response.OwnerEmail = profile.Email
		response.OwnerName = profile.NickName
	}
	// Graceful degradation: if profile fetch fails, return response without owner details

	return response
}

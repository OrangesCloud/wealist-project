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
	GetProject(ctx context.Context, projectID, userID uuid.UUID, token string) (*dto.ProjectResponse, error)
	UpdateProject(ctx context.Context, projectID, userID uuid.UUID, req *dto.UpdateProjectRequest) (*dto.ProjectResponse, error)
	DeleteProject(ctx context.Context, projectID, userID uuid.UUID) error
	SearchProjects(ctx context.Context, workspaceID, userID uuid.UUID, query string, page, limit int, token string) (*dto.PaginatedProjectsResponse, error)
	GetProjectInitSettings(ctx context.Context, projectID, userID uuid.UUID, token string) (*dto.ProjectInitSettingsResponse, error)
}

// projectServiceImpl is the implementation of ProjectService
type projectServiceImpl struct {
	projectRepo       repository.ProjectRepository
	fieldOptionRepo   repository.FieldOptionRepository
	userClient        client.UserClient
}

// NewProjectService creates a new instance of ProjectService
func NewProjectService(projectRepo repository.ProjectRepository, fieldOptionRepo repository.FieldOptionRepository, userClient client.UserClient) ProjectService {
	return &projectServiceImpl{
		projectRepo:     projectRepo,
		fieldOptionRepo: fieldOptionRepo,
		userClient:      userClient,
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

	// Add creator as OWNER member
	member := &domain.ProjectMember{
		ProjectID: project.ID,
		UserID:    userID,
		RoleName:  domain.ProjectRoleOwner,
	}
	if err := s.projectRepo.AddMember(ctx, member); err != nil {
		return nil, response.NewAppError(response.ErrCodeInternal, "Failed to add project owner", err.Error())
	}

	// Create default field options for the project
	if err := s.createDefaultFieldOptions(ctx, project.ID); err != nil {
		// Log error but don't fail project creation
		// The field options can be created manually later
		// TODO: Add proper logging
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

	// 빈 배열 명시적 처리 - nil이거나 길이가 0이면 빈 배열 반환
	if projects == nil || len(projects) == 0 {
		return []*dto.ProjectResponse{}, nil
	}

	// Convert to response DTOs with owner profile information
	// 동적으로 append하여 개별 프로젝트 변환 실패 시 전체 실패 방지
	responses := make([]*dto.ProjectResponse, 0, len(projects))
	for i, project := range projects {
		// nil 프로젝트 스킵
		if project == nil {
			continue
		}
		
		// 개별 변환 실패 시 해당 프로젝트만 스킵
		projectResp := s.toProjectResponseWithProfile(ctx, project, token)
		if projectResp != nil {
			responses = append(responses, projectResp)
		} else {
			// Log when a project response is nil to help debugging
			// This should not happen in normal operation
			_ = i // Avoid unused variable warning
		}
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
	// nil 체크 - project가 nil이면 nil 반환
	if project == nil {
		return nil
	}
	
	response := s.toProjectResponse(project)
	// response가 nil이면 nil 반환
	if response == nil {
		return nil
	}

	// Fetch workspace profile for owner - graceful degradation
	profile, err := s.userClient.GetWorkspaceProfile(ctx, project.WorkspaceID, project.OwnerID, token)
	if err != nil {
		// 에러 발생 시 owner 정보 없이 반환 (graceful degradation)
		return response
	}
	
	// profile이 nil이 아닐 때만 정보 추가
	if profile != nil {
		response.OwnerEmail = profile.Email
		response.OwnerName = profile.NickName
	}

	return response
}

// GetProject retrieves a project by ID with membership validation
func (s *projectServiceImpl) GetProject(ctx context.Context, projectID, userID uuid.UUID, token string) (*dto.ProjectResponse, error) {
	// Fetch project from repository
	project, err := s.projectRepo.FindByID(ctx, projectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, response.NewNotFoundError("Project not found", "")
		}
		return nil, response.NewAppError(response.ErrCodeInternal, "Failed to fetch project", err.Error())
	}

	// Check if user is a project member
	isMember, err := s.projectRepo.IsProjectMember(ctx, projectID, userID)
	if err != nil {
		return nil, response.NewAppError(response.ErrCodeInternal, "Failed to check membership", err.Error())
	}
	if !isMember {
		return nil, response.NewForbiddenError("You are not a member of this project", "")
	}

	// Convert to response DTO with owner profile information
	return s.toProjectResponseWithProfile(ctx, project, token), nil
}

// UpdateProject updates a project (OWNER only)
func (s *projectServiceImpl) UpdateProject(ctx context.Context, projectID, userID uuid.UUID, req *dto.UpdateProjectRequest) (*dto.ProjectResponse, error) {
	// Fetch project from repository
	project, err := s.projectRepo.FindByID(ctx, projectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, response.NewNotFoundError("Project not found", "")
		}
		return nil, response.NewAppError(response.ErrCodeInternal, "Failed to fetch project", err.Error())
	}

	// Check if user is the project owner
	member, err := s.projectRepo.FindMemberByProjectAndUser(ctx, projectID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, response.NewForbiddenError("You are not a member of this project", "")
		}
		return nil, response.NewAppError(response.ErrCodeInternal, "Failed to check membership", err.Error())
	}
	if member.RoleName != domain.ProjectRoleOwner {
		return nil, response.NewForbiddenError("Only project owner can update project", "")
	}

	// Update fields if provided
	if req.Name != nil {
		project.Name = *req.Name
	}
	if req.Description != nil {
		project.Description = *req.Description
	}

	// Save to repository
	if err := s.projectRepo.Update(ctx, project); err != nil {
		return nil, response.NewAppError(response.ErrCodeInternal, "Failed to update project", err.Error())
	}

	// Convert to response DTO
	return s.toProjectResponse(project), nil
}

// DeleteProject soft deletes a project (OWNER only)
func (s *projectServiceImpl) DeleteProject(ctx context.Context, projectID, userID uuid.UUID) error {
	// Fetch project from repository
	project, err := s.projectRepo.FindByID(ctx, projectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.NewNotFoundError("Project not found", "")
		}
		return response.NewAppError(response.ErrCodeInternal, "Failed to fetch project", err.Error())
	}

	// Check if user is the project owner
	member, err := s.projectRepo.FindMemberByProjectAndUser(ctx, projectID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.NewForbiddenError("You are not a member of this project", "")
		}
		return response.NewAppError(response.ErrCodeInternal, "Failed to check membership", err.Error())
	}
	if member.RoleName != domain.ProjectRoleOwner {
		return response.NewForbiddenError("Only project owner can delete project", "")
	}

	// Delete from repository
	if err := s.projectRepo.Delete(ctx, project.ID); err != nil {
		return response.NewAppError(response.ErrCodeInternal, "Failed to delete project", err.Error())
	}

	return nil
}

// SearchProjects searches projects by name or description with workspace membership validation
func (s *projectServiceImpl) SearchProjects(ctx context.Context, workspaceID, userID uuid.UUID, query string, page, limit int, token string) (*dto.PaginatedProjectsResponse, error) {
	// Validate workspace membership
	isValid, err := s.userClient.ValidateWorkspaceMember(ctx, workspaceID, userID, token)
	if err != nil {
		return nil, response.NewAppError(response.ErrCodeForbidden, "You are not a member of this workspace", "")
	}
	if !isValid {
		return nil, response.NewAppError(response.ErrCodeForbidden, "You are not a member of this workspace", "")
	}

	// Validate query parameter
	if query == "" {
		return nil, response.NewValidationError("Search query cannot be empty", "")
	}

	// Set default pagination values
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	// Search projects from repository
	projects, total, err := s.projectRepo.Search(ctx, workspaceID, query, page, limit)
	if err != nil {
		return nil, response.NewAppError(response.ErrCodeInternal, "Failed to search projects", err.Error())
	}

	// Convert to response DTOs with owner profile information
	responses := make([]dto.ProjectResponse, len(projects))
	for i, project := range projects {
		responses[i] = *s.toProjectResponseWithProfile(ctx, project, token)
	}

	return &dto.PaginatedProjectsResponse{
		Projects: responses,
		Total:    total,
		Page:     page,
		Limit:    limit,
	}, nil
}

// createDefaultFieldOptions creates default field options for a new project
// by copying system default options
func (s *projectServiceImpl) createDefaultFieldOptions(ctx context.Context, projectID uuid.UUID) error {
	// Fetch system default field options
	systemDefaults, err := s.fieldOptionRepo.FindSystemDefaults(ctx)
	if err != nil {
		return err
	}

	if len(systemDefaults) == 0 {
		// No system defaults found, skip
		return nil
	}

	// Create project-specific copies of system defaults
	projectOptions := make([]*domain.FieldOption, len(systemDefaults))
	for i, defaultOption := range systemDefaults {
		projectOptions[i] = &domain.FieldOption{
			ProjectID:       &projectID,
			FieldType:       defaultOption.FieldType,
			Value:           defaultOption.Value,
			Label:           defaultOption.Label,
			Color:           defaultOption.Color,
			DisplayOrder:    defaultOption.DisplayOrder,
			IsSystemDefault: false, // Project-specific options are not system defaults
		}
	}

	// Batch create all project options
	if err := s.fieldOptionRepo.CreateBatch(ctx, projectOptions); err != nil {
		return err
	}

	return nil
}

// GetProjectInitSettings retrieves initial settings for a project including field definitions
func (s *projectServiceImpl) GetProjectInitSettings(ctx context.Context, projectID, userID uuid.UUID, token string) (*dto.ProjectInitSettingsResponse, error) {
	// Fetch project from repository
	project, err := s.projectRepo.FindByID(ctx, projectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, response.NewNotFoundError("Project not found", "")
		}
		return nil, response.NewAppError(response.ErrCodeInternal, "Failed to fetch project", err.Error())
	}

	// Check if user is a project member
	isMember, err := s.projectRepo.IsProjectMember(ctx, projectID, userID)
	if err != nil {
		return nil, response.NewAppError(response.ErrCodeInternal, "Failed to check membership", err.Error())
	}
	if !isMember {
		return nil, response.NewForbiddenError("You are not a member of this project", "")
	}

	// Fetch workspace information
	workspace, err := s.userClient.GetWorkspace(ctx, project.WorkspaceID, token)
	if err != nil {
		// Log error but continue with graceful degradation
		workspace = &client.Workspace{
			ID:   project.WorkspaceID,
			Name: "",
		}
	}

	// Build project basic info with workspace details
	projectInfo := dto.ProjectBasicInfo{
		ProjectID:        project.ID,
		WorkspaceID:      project.WorkspaceID,
		WorkspaceName:    workspace.Name,
		WorkspaceEmail:   workspace.OwnerEmail,
		Name:             project.Name,
		Description:      project.Description,
		OwnerID:          project.OwnerID,
		IsPublic:         project.IsPublic,
		CreatedAt:        project.CreatedAt,
		UpdatedAt:        project.UpdatedAt,
	}

	// Fetch project-specific field options from database
	stageOptions, err := s.fieldOptionRepo.FindByProjectAndFieldType(ctx, projectID, domain.FieldTypeStage)
	if err != nil {
		return nil, response.NewAppError(response.ErrCodeInternal, "Failed to fetch stage options", err.Error())
	}

	roleOptions, err := s.fieldOptionRepo.FindByProjectAndFieldType(ctx, projectID, domain.FieldTypeRole)
	if err != nil {
		return nil, response.NewAppError(response.ErrCodeInternal, "Failed to fetch role options", err.Error())
	}

	importanceOptions, err := s.fieldOptionRepo.FindByProjectAndFieldType(ctx, projectID, domain.FieldTypeImportance)
	if err != nil {
		return nil, response.NewAppError(response.ErrCodeInternal, "Failed to fetch importance options", err.Error())
	}

	// Convert field options to DTO format
	stageFieldOptions := make([]dto.FieldOption, len(stageOptions))
	for i, opt := range stageOptions {
		stageFieldOptions[i] = dto.FieldOption{
			OptionID:     opt.ID.String(),
			OptionLabel:  opt.Label,
			OptionValue:  opt.Value,
			Color:        opt.Color,
			DisplayOrder: opt.DisplayOrder,
			FieldID:      "stage",
		}
	}

	roleFieldOptions := make([]dto.FieldOption, len(roleOptions))
	for i, opt := range roleOptions {
		roleFieldOptions[i] = dto.FieldOption{
			OptionID:     opt.ID.String(),
			OptionLabel:  opt.Label,
			OptionValue:  opt.Value,
			Color:        opt.Color,
			DisplayOrder: opt.DisplayOrder,
			FieldID:      "role",
		}
	}

	importanceFieldOptions := make([]dto.FieldOption, len(importanceOptions))
	for i, opt := range importanceOptions {
		importanceFieldOptions[i] = dto.FieldOption{
			OptionID:     opt.ID.String(),
			OptionLabel:  opt.Label,
			OptionValue:  opt.Value,
			Color:        opt.Color,
			DisplayOrder: opt.DisplayOrder,
			FieldID:      "importance",
		}
	}

	// Define field definitions with options from database
	fields := []dto.FieldWithOptionsResponse{
		{
			FieldID:     "stage",
			FieldName:   "Stage",
			FieldType:   "select",
			IsRequired:  true,
			Description: "Current stage of the board",
			Options:     stageFieldOptions,
		},
		{
			FieldID:     "importance",
			FieldName:   "Importance",
			FieldType:   "select",
			IsRequired:  true,
			Description: "Priority level of the board",
			Options:     importanceFieldOptions,
		},
		{
			FieldID:     "role",
			FieldName:   "Role",
			FieldType:   "select",
			IsRequired:  true,
			Description: "Role responsible for the board",
			Options:     roleFieldOptions,
		},
	}

	// Define field types
	fieldTypes := []dto.FieldTypeInfo{
		{
			TypeID:      "select",
			TypeName:    "Select",
			Description: "Single selection from predefined options",
		},
		{
			TypeID:      "text",
			TypeName:    "Text",
			Description: "Free text input",
		},
		{
			TypeID:      "date",
			TypeName:    "Date",
			Description: "Date selection",
		},
		{
			TypeID:      "user",
			TypeName:    "User",
			Description: "User selection",
		},
	}

	return &dto.ProjectInitSettingsResponse{
		Project:       projectInfo,
		Fields:        fields,
		FieldTypes:    fieldTypes,
		DefaultViewID: nil, // Can be extended later to support custom views
	}, nil
}

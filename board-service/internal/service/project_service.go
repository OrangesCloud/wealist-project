package service

import (
	"board-service/internal/apperrors"
	"board-service/internal/cache"
	"board-service/internal/client"
	"board-service/internal/domain"
	"board-service/internal/dto"
	"board-service/internal/repository"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type ProjectService interface {
	CreateProject(userID string, token string, req *dto.CreateProjectRequest) (*dto.ProjectResponse, error)
	GetProject(projectID, userID string) (*dto.ProjectResponse, error)
	GetProjectsByWorkspaceID(workspaceID, userID string, token string) ([]dto.ProjectResponse, error)
	UpdateProject(projectID, userID string, req *dto.UpdateProjectRequest) (*dto.ProjectResponse, error)
	DeleteProject(projectID, userID string) error
	SearchProjects(userID string, token string, req *dto.SearchProjectsRequest) (*dto.PaginatedProjectsResponse, error)

	// Init Settings
	GetProjectInitSettings(projectID, userID string) (*dto.ProjectInitSettingsResponse, error)

	// Join Request
	CreateJoinRequest(userID string, token string, req *dto.CreateProjectJoinRequestRequest) (*dto.ProjectJoinRequestResponse, error)
	GetJoinRequests(projectID, userID string, status string) ([]dto.ProjectJoinRequestResponse, error)
	UpdateJoinRequest(requestID, userID string, req *dto.UpdateProjectJoinRequestRequest) (*dto.ProjectJoinRequestResponse, error)

	// Member
	GetProjectMembers(projectID, userID string) ([]dto.ProjectMemberResponse, error)
	UpdateMemberRole(projectID, memberID, requestUserID string, req *dto.UpdateProjectMemberRoleRequest) (*dto.ProjectMemberResponse, error)
	RemoveMember(projectID, memberID, requestUserID string) error
}

type projectService struct {
	repo             repository.ProjectRepository
	roleRepo         repository.RoleRepository
	fieldRepo        repository.FieldRepository
	boardRepo        repository.BoardRepository
	projectFieldRepo repository.ProjectFieldRepository
	fieldOptionRepo  repository.FieldOptionRepository
	boardOrderRepo   repository.BoardOrderRepository
	viewRepo         repository.ViewRepository
	userClient       client.UserClient
	workspaceCache   cache.WorkspaceCache
	userInfoCache    cache.UserInfoCache
	logger           *zap.Logger
	db               *gorm.DB
}

func NewProjectService(
	repo repository.ProjectRepository,
	roleRepo repository.RoleRepository,
	fieldRepo repository.FieldRepository,
	boardRepo repository.BoardRepository,
	projectFieldRepo repository.ProjectFieldRepository,
	fieldOptionRepo repository.FieldOptionRepository,
	boardOrderRepo repository.BoardOrderRepository,
	viewRepo repository.ViewRepository,
	userClient client.UserClient,
	workspaceCache cache.WorkspaceCache,
	userInfoCache cache.UserInfoCache,
	logger *zap.Logger,
	db *gorm.DB,
) ProjectService {
	return &projectService{
		repo:             repo,
		roleRepo:         roleRepo,
		fieldRepo:        fieldRepo,
		boardRepo:        boardRepo,
		projectFieldRepo: projectFieldRepo,
		fieldOptionRepo:  fieldOptionRepo,
		boardOrderRepo:   boardOrderRepo,
		viewRepo:         viewRepo,
		userClient:       userClient,
		workspaceCache:   workspaceCache,
		userInfoCache:    userInfoCache,
		logger:           logger,
		db:               db,
	}
}

// CreateProject creates a new project
func (s *projectService) CreateProject(userID string, token string, req *dto.CreateProjectRequest) (*dto.ProjectResponse, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	workspaceUUID, err := uuid.Parse(req.WorkspaceID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 워크스페이스 ID", 400)
	}

	// Check workspace membership with caching
	ctx := context.Background()
	if err := s.validateWorkspaceMembership(ctx, req.WorkspaceID, userID, token); err != nil {
		return nil, err
	}

	// Get OWNER role
	ownerRole, err := s.roleRepo.FindByName("OWNER")
	if err != nil {
		s.logger.Error("Failed to find OWNER role", zap.Error(err))
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "권한 조회 실패", 500)
	}

	// Create project and member in transaction
	var project *domain.Project
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// Create project
		project = &domain.Project{
			WorkspaceID: workspaceUUID,
			Name:        req.Name,
			Description: req.Description,
			OwnerID:     userUUID,
		}
		if err := s.repo.Create(project); err != nil {
			return err
		}

		// Create owner member
		member := &domain.ProjectMember{
			ProjectID: project.ID,
			UserID:    userUUID,
			RoleID:    ownerRole.ID,
			JoinedAt:  time.Now(),
		}
		if err := s.repo.CreateMember(member); err != nil {
			return err
		}

		// Create default custom fields (stage, role, importance)
		if err := s.initializeDefaultFields(project.ID); err != nil {
			s.logger.Error("Failed to initialize default fields", zap.Error(err))
			return err
		}

		return nil
	})

	if err != nil {
		s.logger.Error("Failed to create project", zap.Error(err))
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "프로젝트 생성 실패", 500)
	}

	// Fetch user info
	return s.toProjectResponse(project)
}

// GetProject retrieves a project by ID
func (s *projectService) GetProject(projectID, userID string) (*dto.ProjectResponse, error) {
	projUUID, err := uuid.Parse(projectID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 프로젝트 ID", 400)
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	// Check if user is project member
	_, err = s.repo.FindMemberByUserAndProject(userUUID, projUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.ErrCodeForbidden, "프로젝트 멤버가 아닙니다", 403)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "멤버 확인 실패", 500)
	}

	project, err := s.repo.FindByID(projUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.ErrCodeNotFound, "프로젝트를 찾을 수 없습니다", 404)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "프로젝트 조회 실패", 500)
	}

	return s.toProjectResponse(project)
}

// GetProjectsByWorkspaceID retrieves all projects in a workspace
func (s *projectService) GetProjectsByWorkspaceID(workspaceID, userID string, token string) ([]dto.ProjectResponse, error) {
	workspaceUUID, err := uuid.Parse(workspaceID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 워크스페이스 ID", 400)
	}

	// Check workspace membership with caching
	ctx := context.Background()
	if err := s.validateWorkspaceMembership(ctx, workspaceID, userID, token); err != nil {
		return nil, err
	}

	projects, err := s.repo.FindByWorkspaceID(workspaceUUID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "프로젝트 조회 실패", 500)
	}

	// Batch fetch owner info
	ownerIDs := make([]string, 0, len(projects))
	for _, proj := range projects {
		ownerIDs = append(ownerIDs, proj.OwnerID.String())
	}
	userMap := s.getUserInfoBatch(ctx, ownerIDs)

	// Convert to responses
	responses := make([]dto.ProjectResponse, 0, len(projects))
	for _, proj := range projects {
		response := &dto.ProjectResponse{
			ID:          proj.ID.String(),
			WorkspaceID: proj.WorkspaceID.String(),
			Name:        proj.Name,
			Description: proj.Description,
			OwnerID:     proj.OwnerID.String(),
			CreatedAt:   proj.CreatedAt,
			UpdatedAt:   proj.UpdatedAt,
		}

		// Add owner info from batch result
		if userInfo, ok := userMap[proj.OwnerID.String()]; ok {
			response.OwnerName = userInfo.Name
			response.OwnerEmail = userInfo.Email
		}

		responses = append(responses, *response)
	}

	return responses, nil
}

// UpdateProject updates project information
func (s *projectService) UpdateProject(projectID, userID string, req *dto.UpdateProjectRequest) (*dto.ProjectResponse, error) {
	projUUID, err := uuid.Parse(projectID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 프로젝트 ID", 400)
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	// Check if user is project OWNER
	if err := s.checkProjectOwnerPermission(userUUID, projUUID); err != nil {
		return nil, err
	}

	project, err := s.repo.FindByID(projUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.ErrCodeNotFound, "프로젝트를 찾을 수 없습니다", 404)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "프로젝트 조회 실패", 500)
	}

	// Update fields using Domain methods (Rich Domain Model)
	if req.Name != "" {
		// Domain 메서드 사용: 검증 로직이 Domain에 포함됨
		if err := project.UpdateName(req.Name); err != nil {
			// Domain 에러를 Infrastructure 에러로 변환
			return nil, apperrors.FromDomainError(err)
		}
	}
	if req.Description != "" {
		// Domain 메서드 사용: 비즈니스 로직이 Domain에 캡슐화됨
		project.UpdateDescription(req.Description)
	}

	if err := s.repo.Update(project); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "프로젝트 수정 실패", 500)
	}

	return s.toProjectResponse(project)
}

// DeleteProject soft deletes a project
func (s *projectService) DeleteProject(projectID, userID string) error {
	projUUID, err := uuid.Parse(projectID)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 프로젝트 ID", 400)
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	// Check if user is project OWNER
	if err := s.checkProjectOwnerPermission(userUUID, projUUID); err != nil {
		return err
	}

	if err := s.repo.Delete(projUUID); err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "프로젝트 삭제 실패", 500)
	}

	return nil
}

// SearchProjects searches projects in a workspace
func (s *projectService) SearchProjects(userID string, token string, req *dto.SearchProjectsRequest) (*dto.PaginatedProjectsResponse, error) {
	workspaceUUID, err := uuid.Parse(req.WorkspaceID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 워크스페이스 ID", 400)
	}

	// Check workspace membership with caching
	ctx := context.Background()
	if err := s.validateWorkspaceMembership(ctx, req.WorkspaceID, userID, token); err != nil {
		return nil, err
	}

	// Default values
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 {
		req.Limit = 10
	}

	projects, total, err := s.repo.Search(workspaceUUID, req.Query, req.Page, req.Limit)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "프로젝트 검색 실패", 500)
	}

	// Batch fetch owner info
	ownerIDs := make([]string, 0, len(projects))
	for _, proj := range projects {
		ownerIDs = append(ownerIDs, proj.OwnerID.String())
	}
	userMap := s.getUserInfoBatch(ctx, ownerIDs)

	// Convert to response DTOs
	responses := make([]dto.ProjectResponse, 0, len(projects))
	for _, proj := range projects {
		response := &dto.ProjectResponse{
			ID:          proj.ID.String(),
			WorkspaceID: proj.WorkspaceID.String(),
			Name:        proj.Name,
			Description: proj.Description,
			OwnerID:     proj.OwnerID.String(),
			CreatedAt:   proj.CreatedAt,
			UpdatedAt:   proj.UpdatedAt,
		}

		// Add owner info from batch result
		if userInfo, ok := userMap[proj.OwnerID.String()]; ok {
			response.OwnerName = userInfo.Name
			response.OwnerEmail = userInfo.Email
		}

		responses = append(responses, *response)
	}

	return &dto.PaginatedProjectsResponse{
		Projects: responses,
		Total:    total,
		Page:     req.Page,
		Limit:    req.Limit,
	}, nil
}

// CreateJoinRequest creates a join request
func (s *projectService) CreateJoinRequest(userID string, token string, req *dto.CreateProjectJoinRequestRequest) (*dto.ProjectJoinRequestResponse, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	projUUID, err := uuid.Parse(req.ProjectID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 프로젝트 ID", 400)
	}

	// Check if project exists
	project, err := s.repo.FindByID(projUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.ErrCodeNotFound, "프로젝트를 찾을 수 없습니다", 404)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "프로젝트 조회 실패", 500)
	}

	// Check workspace membership with caching
	ctx := context.Background()
	if err := s.validateWorkspaceMembership(ctx, project.WorkspaceID.String(), userID, token); err != nil {
		return nil, err
	}

	// Check if already a member
	_, err = s.repo.FindMemberByUserAndProject(userUUID, projUUID)
	if err == nil {
		return nil, apperrors.New(apperrors.ErrCodeConflict, "이미 프로젝트 멤버입니다", 409)
	}

	// Check if request already exists
	existingReq, err := s.repo.FindJoinRequestByUserAndProject(userUUID, projUUID)
	if err == nil && existingReq.Status == domain.ProjectJoinRequestPending {
		return nil, apperrors.New(apperrors.ErrCodeConflict, "이미 참여 신청이 있습니다", 409)
	}

	// Create join request
	joinReq := &domain.ProjectJoinRequest{
		ProjectID:   projUUID,
		UserID:      userUUID,
		Status:      domain.ProjectJoinRequestPending,
		RequestedAt: time.Now(),
	}

	if err := s.repo.CreateJoinRequest(joinReq); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "참여 신청 실패", 500)
	}

	return s.toJoinRequestResponse(joinReq)
}

// GetJoinRequests retrieves join requests for a project
func (s *projectService) GetJoinRequests(projectID, userID string, status string) ([]dto.ProjectJoinRequestResponse, error) {
	projUUID, err := uuid.Parse(projectID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 프로젝트 ID", 400)
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	// Check if user is project OWNER or ADMIN
	if err := s.checkProjectAdminPermission(userUUID, projUUID); err != nil {
		return nil, err
	}

	requests, err := s.repo.FindJoinRequestsByProject(projUUID, status)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "참여 신청 조회 실패", 500)
	}

	// Batch fetch user info
	ctx := context.Background()
	userIDs := make([]string, 0, len(requests))
	for _, req := range requests {
		userIDs = append(userIDs, req.UserID.String())
	}
	userMap := s.getUserInfoBatch(ctx, userIDs)

	// Convert to responses
	responses := make([]dto.ProjectJoinRequestResponse, 0, len(requests))
	for _, req := range requests {
		response := &dto.ProjectJoinRequestResponse{
			ID:          req.ID.String(),
			ProjectID:   req.ProjectID.String(),
			UserID:      req.UserID.String(),
			Status:      string(req.Status),
			RequestedAt: req.RequestedAt,
			UpdatedAt:   req.UpdatedAt,
		}

		// Add user info from batch result
		if userInfo, ok := userMap[req.UserID.String()]; ok {
			response.UserName = userInfo.Name
			response.UserEmail = userInfo.Email
		}

		responses = append(responses, *response)
	}

	return responses, nil
}

// UpdateJoinRequest approves or rejects a join request
func (s *projectService) UpdateJoinRequest(requestID, userID string, req *dto.UpdateProjectJoinRequestRequest) (*dto.ProjectJoinRequestResponse, error) {
	reqUUID, err := uuid.Parse(requestID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 요청 ID", 400)
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	joinReq, err := s.repo.FindJoinRequestByID(reqUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.ErrCodeNotFound, "참여 신청을 찾을 수 없습니다", 404)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "참여 신청 조회 실패", 500)
	}

	// Check if user is project OWNER or ADMIN
	if err := s.checkProjectAdminPermission(userUUID, joinReq.ProjectID); err != nil {
		return nil, err
	}

	// Update status
	joinReq.Status = domain.ProjectJoinRequestStatus(req.Status)

	// If approved, create member
	if req.Status == "APPROVED" {
		memberRole, err := s.roleRepo.FindByName("MEMBER")
		if err != nil {
			return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "권한 조회 실패", 500)
		}

		member := &domain.ProjectMember{
			ProjectID: joinReq.ProjectID,
			UserID:    joinReq.UserID,
			RoleID:    memberRole.ID,
			JoinedAt:  time.Now(),
		}

		if err := s.repo.CreateMember(member); err != nil {
			return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "멤버 생성 실패", 500)
		}
	}

	if err := s.repo.UpdateJoinRequest(joinReq); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "참여 신청 처리 실패", 500)
	}

	return s.toJoinRequestResponse(joinReq)
}

// GetProjectMembers retrieves all members of a project
func (s *projectService) GetProjectMembers(projectID, userID string) ([]dto.ProjectMemberResponse, error) {
	projUUID, err := uuid.Parse(projectID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 프로젝트 ID", 400)
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	// Check if user is project member
	_, err = s.repo.FindMemberByUserAndProject(userUUID, projUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.ErrCodeForbidden, "프로젝트 멤버가 아닙니다", 403)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "멤버 확인 실패", 500)
	}

	members, err := s.repo.FindMembersByProject(projUUID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "멤버 조회 실패", 500)
	}

	// Batch fetch user info
	ctx := context.Background()
	userIDs := make([]string, 0, len(members))
	for _, member := range members {
		userIDs = append(userIDs, member.UserID.String())
	}
	userMap := s.getUserInfoBatch(ctx, userIDs)

	// Convert to responses
	responses := make([]dto.ProjectMemberResponse, 0, len(members))
	for _, member := range members {
		// Get role name
		role, err := s.roleRepo.FindByID(member.RoleID)
		if err != nil {
			s.logger.Warn("Failed to get role", zap.Error(err))
			continue
		}

		response := &dto.ProjectMemberResponse{
			ID:        member.ID.String(),
			ProjectID: member.ProjectID.String(),
			UserID:    member.UserID.String(),
			RoleName:  role.Name,
			JoinedAt:  member.JoinedAt,
		}

		// Add user info from batch result
		if userInfo, ok := userMap[member.UserID.String()]; ok {
			response.UserName = userInfo.Name
			response.UserEmail = userInfo.Email
		}

		responses = append(responses, *response)
	}

	return responses, nil
}

// UpdateMemberRole updates a member's role
func (s *projectService) UpdateMemberRole(projectID, memberID, requestUserID string, req *dto.UpdateProjectMemberRoleRequest) (*dto.ProjectMemberResponse, error) {
	projUUID, err := uuid.Parse(projectID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 프로젝트 ID", 400)
	}

	memberUUID, err := uuid.Parse(memberID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 멤버 ID", 400)
	}

	reqUserUUID, err := uuid.Parse(requestUserID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	// Check if requestUser is project OWNER
	if err := s.checkProjectOwnerPermission(reqUserUUID, projUUID); err != nil {
		return nil, err
	}

	member, err := s.repo.FindMemberByID(memberUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.ErrCodeNotFound, "멤버를 찾을 수 없습니다", 404)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "멤버 조회 실패", 500)
	}

	// Cannot change own role
	if member.UserID == reqUserUUID {
		return nil, apperrors.New(apperrors.ErrCodeBadRequest, "자신의 권한은 변경할 수 없습니다", 400)
	}

	// Get new role
	newRole, err := s.roleRepo.FindByName(req.RoleName)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "권한 조회 실패", 500)
	}

	member.RoleID = newRole.ID
	if err := s.repo.UpdateMember(member); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "멤버 권한 수정 실패", 500)
	}

	return s.toMemberResponse(member)
}

// RemoveMember removes a member from project
func (s *projectService) RemoveMember(projectID, memberID, requestUserID string) error {
	projUUID, err := uuid.Parse(projectID)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 프로젝트 ID", 400)
	}

	memberUUID, err := uuid.Parse(memberID)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 멤버 ID", 400)
	}

	reqUserUUID, err := uuid.Parse(requestUserID)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	// Check if requestUser is project OWNER or ADMIN
	if err := s.checkProjectAdminPermission(reqUserUUID, projUUID); err != nil {
		return err
	}

	member, err := s.repo.FindMemberByID(memberUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.New(apperrors.ErrCodeNotFound, "멤버를 찾을 수 없습니다", 404)
		}
		return apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "멤버 조회 실패", 500)
	}

	// Cannot remove self
	if member.UserID == reqUserUUID {
		return apperrors.New(apperrors.ErrCodeBadRequest, "자신을 삭제할 수 없습니다", 400)
	}

	// Cannot remove OWNER
	role, err := s.roleRepo.FindByID(member.RoleID)
	if err == nil && role.Name == "OWNER" {
		return apperrors.New(apperrors.ErrCodeBadRequest, "OWNER는 삭제할 수 없습니다", 400)
	}

	if err := s.repo.DeleteMember(memberUUID); err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "멤버 삭제 실패", 500)
	}

	return nil
}

// Helper methods

// validateWorkspaceMembership checks workspace membership with caching
func (s *projectService) validateWorkspaceMembership(ctx context.Context, workspaceID, userID, token string) error {
	// TODO: Temporarily skip cache for debugging
	s.logger.Info("Skipping cache, calling User Service directly for debugging",
		zap.String("workspace_id", workspaceID),
		zap.String("user_id", userID))

	var isMember bool
	var err error

	// Try cache first (commented out for debugging)
	// cacheExists, isMember, err := s.workspaceCache.GetMembership(ctx, workspaceID, userID)
	// if err != nil {
	// 	s.logger.Warn("Failed to get workspace membership from cache", zap.Error(err))
	// 	// Continue to User Service call on cache error
	// }

	// if cacheExists {
	// 	// Cache hit
	// 	s.logger.Info("Cache hit for workspace membership",
	// 		zap.String("workspace_id", workspaceID),
	// 		zap.String("user_id", userID),
	// 		zap.Bool("is_member", isMember))
	// 	if !isMember {
	// 		return apperrors.New(apperrors.ErrCodeWorkspaceAccessDenied, "워크스페이스 멤버가 아닙니다", 403)
	// 	}
	// 	return nil
	// }

	// s.logger.Info("Cache miss for workspace membership, calling User Service",
	// 	zap.String("workspace_id", workspaceID),
	// 	zap.String("user_id", userID))

	// Cache miss - validate via User Service
	// First check if workspace exists
	workspaceExists, err := s.userClient.CheckWorkspaceExists(ctx, workspaceID, token)
	if err != nil {
		s.logger.Error("Failed to check workspace existence", zap.Error(err), zap.String("workspace_id", workspaceID))
		return apperrors.Wrap(err, apperrors.ErrCodeWorkspaceValidationFailed, "워크스페이스 확인 실패", 500)
	}
	if !workspaceExists {
		return apperrors.New(apperrors.ErrCodeWorkspaceNotFound, "워크스페이스를 찾을 수 없습니다", 404)
	}

	// Check membership
	s.logger.Info("Validating workspace membership via User Service",
		zap.String("workspace_id", workspaceID),
		zap.String("user_id", userID))
	isMember, err = s.userClient.ValidateWorkspaceMembership(ctx, workspaceID, userID, token)
	if err != nil {
		s.logger.Error("Failed to validate workspace membership", zap.Error(err), zap.String("workspace_id", workspaceID), zap.String("user_id", userID))
		return apperrors.Wrap(err, apperrors.ErrCodeWorkspaceValidationFailed, "워크스페이스 멤버십 확인 실패", 500)
	}
	s.logger.Info("Workspace membership validation result",
		zap.String("workspace_id", workspaceID),
		zap.String("user_id", userID),
		zap.Bool("is_member", isMember))

	// Cache the result
	if cacheErr := s.workspaceCache.SetMembership(ctx, workspaceID, userID, isMember); cacheErr != nil {
		s.logger.Warn("Failed to cache workspace membership", zap.Error(cacheErr))
		// Don't fail the request on cache write error
	}

	if !isMember {
		s.logger.Warn("User is not a member of workspace",
			zap.String("workspace_id", workspaceID),
			zap.String("user_id", userID))
		return apperrors.New(apperrors.ErrCodeWorkspaceAccessDenied, "워크스페이스 멤버가 아닙니다", 403)
	}

	return nil
}

// getUserInfoWithCache retrieves user info with caching
func (s *projectService) getUserInfoWithCache(ctx context.Context, userID string) (*dto.UserInfo, error) {
	// Try cache first
	cacheExists, cachedUser, err := s.userInfoCache.GetUserInfo(ctx, userID)
	if err != nil {
		s.logger.Warn("Failed to get user info from cache", zap.Error(err))
	}

	if cacheExists && cachedUser != nil {
		// Cache hit - convert to dto.UserInfo
		return &dto.UserInfo{
			UserID: cachedUser.UserID,
			Name:   cachedUser.Name,
			Email:  cachedUser.Email,
		}, nil
	}

	// Cache miss - fetch from User Service
	userInfo, err := s.userClient.GetUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Cache the result
	cacheUser := &cache.UserInfo{
		UserID:   userInfo.UserID,
		Name:     userInfo.Name,
		Email:    userInfo.Email,
		IsActive: userInfo.IsActive,
	}
	if cacheErr := s.userInfoCache.SetUserInfo(ctx, cacheUser); cacheErr != nil {
		s.logger.Warn("Failed to cache user info", zap.Error(cacheErr))
	}

	return &dto.UserInfo{
		UserID: userInfo.UserID,
		Name:   userInfo.Name,
		Email:  userInfo.Email,
	}, nil
}

// getUserInfoBatch retrieves multiple user infos with caching
func (s *projectService) getUserInfoBatch(ctx context.Context, userIDs []string) map[string]*dto.UserInfo {
	if len(userIDs) == 0 {
		return make(map[string]*dto.UserInfo)
	}

	// Try cache first
	cachedUsers, err := s.userInfoCache.GetSimpleUsersBatch(ctx, userIDs)
	if err != nil {
		s.logger.Warn("Failed to get users from cache", zap.Error(err))
		cachedUsers = make(map[string]*cache.SimpleUser)
	}

	// Find missing user IDs
	missingUserIDs := []string{}
	for _, userID := range userIDs {
		if _, exists := cachedUsers[userID]; !exists {
			missingUserIDs = append(missingUserIDs, userID)
		}
	}

	// Build result map
	userMap := make(map[string]*dto.UserInfo)

	// Fetch missing users from User Service
	if len(missingUserIDs) > 0 {
		users, err := s.userClient.GetUsersBatch(ctx, missingUserIDs)
		if err != nil {
			s.logger.Warn("Failed to fetch users from User Service", zap.Error(err))
		} else {
			// Cache the fetched users
			simpleUsers := make([]cache.SimpleUser, 0, len(users))
			for _, user := range users {
				userMap[user.UserID] = &dto.UserInfo{
					UserID: user.UserID,
					Name:   user.Name,
					Email:  user.Email,
				}
				simpleUsers = append(simpleUsers, cache.SimpleUser{
					ID:        user.UserID,
					Name:      user.Name,
					AvatarURL: "",
				})
			}
			if cacheErr := s.userInfoCache.SetSimpleUsersBatch(ctx, simpleUsers); cacheErr != nil {
				s.logger.Warn("Failed to cache users", zap.Error(cacheErr))
			}
		}
	}

	// Add cached users to result
	for userID, cachedUser := range cachedUsers {
		if _, exists := userMap[userID]; !exists {
			userMap[userID] = &dto.UserInfo{
				UserID: cachedUser.ID,
				Name:   cachedUser.Name,
				Email:  "", // SimpleUser doesn't have email
			}
		}
	}

	return userMap
}

func (s *projectService) checkProjectOwnerPermission(userID, projectID uuid.UUID) error {
	member, err := s.repo.FindMemberByUserAndProject(userID, projectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.New(apperrors.ErrCodeForbidden, "프로젝트 멤버가 아닙니다", 403)
		}
		return apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "멤버 확인 실패", 500)
	}

	role, err := s.roleRepo.FindByID(member.RoleID)
	if err != nil || role.Name != "OWNER" {
		return apperrors.New(apperrors.ErrCodeForbidden, "OWNER 권한이 필요합니다", 403)
	}

	return nil
}

func (s *projectService) checkProjectAdminPermission(userID, projectID uuid.UUID) error {
	member, err := s.repo.FindMemberByUserAndProject(userID, projectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.New(apperrors.ErrCodeForbidden, "프로젝트 멤버가 아닙니다", 403)
		}
		return apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "멤버 확인 실패", 500)
	}

	role, err := s.roleRepo.FindByID(member.RoleID)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "권한 조회 실패", 500)
	}

	if role.Name != "OWNER" && role.Name != "ADMIN" {
		return apperrors.New(apperrors.ErrCodeForbidden, "OWNER 또는 ADMIN 권한이 필요합니다", 403)
	}

	return nil
}

func (s *projectService) toProjectResponse(project *domain.Project) (*dto.ProjectResponse, error) {
	response := &dto.ProjectResponse{
		ID:          project.ID.String(),
		WorkspaceID: project.WorkspaceID.String(),
		Name:        project.Name,
		Description: project.Description,
		OwnerID:     project.OwnerID.String(),
		CreatedAt:   project.CreatedAt,
		UpdatedAt:   project.UpdatedAt,
	}

	// Fetch user info with caching
	ctx := context.Background()
	userInfo, err := s.getUserInfoWithCache(ctx, project.OwnerID.String())
	if err != nil {
		s.logger.Warn("Failed to fetch user info", zap.Error(err), zap.String("user_id", project.OwnerID.String()))
		// Continue without user info
	} else {
		response.OwnerName = userInfo.Name
		response.OwnerEmail = userInfo.Email
	}

	return response, nil
}

func (s *projectService) toMemberResponse(member *domain.ProjectMember) (*dto.ProjectMemberResponse, error) {
	// Get role name
	role, err := s.roleRepo.FindByID(member.RoleID)
	if err != nil {
		return nil, err
	}

	response := &dto.ProjectMemberResponse{
		ID:        member.ID.String(),
		ProjectID: member.ProjectID.String(),
		UserID:    member.UserID.String(),
		RoleName:  role.Name,
		JoinedAt:  member.JoinedAt,
	}

	// Fetch user info with caching
	ctx := context.Background()
	userInfo, err := s.getUserInfoWithCache(ctx, member.UserID.String())
	if err != nil {
		s.logger.Warn("Failed to fetch user info", zap.Error(err), zap.String("user_id", member.UserID.String()))
	} else {
		response.UserName = userInfo.Name
		response.UserEmail = userInfo.Email
	}

	return response, nil
}

func (s *projectService) toJoinRequestResponse(req *domain.ProjectJoinRequest) (*dto.ProjectJoinRequestResponse, error) {
	response := &dto.ProjectJoinRequestResponse{
		ID:          req.ID.String(),
		ProjectID:   req.ProjectID.String(),
		UserID:      req.UserID.String(),
		Status:      string(req.Status),
		RequestedAt: req.RequestedAt,
		UpdatedAt:   req.UpdatedAt,
	}

	// Fetch user info with caching
	ctx := context.Background()
	userInfo, err := s.getUserInfoWithCache(ctx, req.UserID.String())
	if err != nil {
		s.logger.Warn("Failed to fetch user info", zap.Error(err), zap.String("user_id", req.UserID.String()))
	} else {
		response.UserName = userInfo.Name
		response.UserEmail = userInfo.Email
	}

	return response, nil
}

// initializeDefaultFields creates default custom fields for a new project
func (s *projectService) initializeDefaultFields(projectID uuid.UUID) error {
	// 1. Create Stage field
	stageField := &domain.ProjectField{
		ProjectID:       projectID,
		Name:            "Stage",
		FieldType:       domain.FieldTypeSingleSelect,
		Description:     "작업 진행 단계",
		DisplayOrder:    0,
		IsRequired:      true,
		IsSystemDefault: true,
		Config:          "{}",
	}
	if err := s.fieldRepo.CreateField(stageField); err != nil {
		return err
	}

	// Create Stage options
	stageOptions := []struct {
		label string
		color string
		order int
	}{
		{"대기", "#F59E0B", 0},
		{"진행중", "#3B82F6", 1},
		{"완료", "#10B981", 2},
	}
	for _, opt := range stageOptions {
		option := &domain.FieldOption{
			FieldID:      stageField.ID,
			Label:        opt.label,
			Color:        opt.color,
			DisplayOrder: opt.order,
		}
		if err := s.fieldRepo.CreateOption(option); err != nil {
			return err
		}
	}

	// 2. Create Role field
	roleField := &domain.ProjectField{
		ProjectID:       projectID,
		Name:            "Role",
		FieldType:       domain.FieldTypeSingleSelect,
		Description:     "담당 역할",
		DisplayOrder:    1,
		IsRequired:      false,
		IsSystemDefault: true,
		Config:          "{}",
	}
	if err := s.fieldRepo.CreateField(roleField); err != nil {
		return err
	}

	// Create Role options
	roleOptions := []struct {
		label string
		color string
		order int
	}{
		{"프론트엔드", "#EC4899", 0},
		{"백엔드", "#8B5CF6", 1},
		{"디자인", "#F97316", 2},
	}
	for _, opt := range roleOptions {
		option := &domain.FieldOption{
			FieldID:      roleField.ID,
			Label:        opt.label,
			Color:        opt.color,
			DisplayOrder: opt.order,
		}
		if err := s.fieldRepo.CreateOption(option); err != nil {
			return err
		}
	}

	// 3. Create Importance field
	importanceField := &domain.ProjectField{
		ProjectID:       projectID,
		Name:            "Importance",
		FieldType:       domain.FieldTypeSingleSelect,
		Description:     "작업 중요도",
		DisplayOrder:    2,
		IsRequired:      false,
		IsSystemDefault: true,
		Config:          "{}",
	}
	if err := s.fieldRepo.CreateField(importanceField); err != nil {
		return err
	}

	// Create Importance options
	importanceOptions := []struct {
		label string
		color string
		order int
	}{
		{"낮음", "#94A3B8", 0},
		{"보통", "#FBBF24", 1},
		{"높음", "#EF4444", 2},
	}
	for _, opt := range importanceOptions {
		option := &domain.FieldOption{
			FieldID:      importanceField.ID,
			Label:        opt.label,
			Color:        opt.color,
			DisplayOrder: opt.order,
		}
		if err := s.fieldRepo.CreateOption(option); err != nil {
			return err
		}
	}

	return nil
}

// GetProjectInitSettings retrieves static configuration data needed for project initialization
func (s *projectService) GetProjectInitSettings(projectID, userID string) (*dto.ProjectInitSettingsResponse, error) {
	projectUUID, err := uuid.Parse(projectID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 프로젝트 ID", 400)
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	// 1. Fetch project information
	project, err := s.repo.FindByID(projectUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.ErrCodeNotFound, "프로젝트를 찾을 수 없습니다", 404)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "프로젝트 조회 실패", 500)
	}

	// 2. Check project membership
	_, err = s.repo.FindMemberByUserAndProject(userUUID, projectUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.ErrCodeForbidden, "프로젝트 멤버가 아닙니다", 403)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "멤버 확인 실패", 500)
	}

	// 3. Fetch default view ID (if exists)
	var defaultViewID string
	defaultView, err := s.viewRepo.FindDefault(projectUUID)
	if err == nil && defaultView != nil {
		defaultViewID = defaultView.ID.String()
	}

	// 4. Fetch fields
	fields, err := s.projectFieldRepo.FindByProject(projectUUID)
	if err != nil {
		s.logger.Error("Failed to fetch fields", zap.Error(err))
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "필드 조회 실패", 500)
	}

	// 5. Fetch options for each field and build response
	fieldsWithOptions := make([]dto.FieldWithOptionsResponse, 0, len(fields))
	for _, field := range fields {
		// Get field options
		options, err := s.fieldOptionRepo.FindByField(field.ID)
		if err != nil {
			s.logger.Error("Failed to fetch field options", zap.String("fieldID", field.ID.String()), zap.Error(err))
			return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "필드 옵션 조회 실패", 500)
		}

		// Convert options to DTO
		optionResponses := make([]dto.OptionResponse, 0, len(options))
		for _, opt := range options {
			optionResponses = append(optionResponses, dto.OptionResponse{
				OptionID:     opt.ID.String(),
				FieldID:      opt.FieldID.String(),
				Label:        opt.Label,
				Color:        opt.Color,
				Description:  opt.Description,
				DisplayOrder: opt.DisplayOrder,
				CreatedAt:    opt.CreatedAt,
				UpdatedAt:    opt.UpdatedAt,
			})
		}

		// Parse config
		var config map[string]interface{}
		if err := json.Unmarshal([]byte(field.Config), &config); err != nil {
			config = make(map[string]interface{})
		}

		// Parse canEditRoles
		var canEditRoles []string
		if field.CanEditRoles != nil && *field.CanEditRoles != "" {
			canEditRoles = strings.Split(*field.CanEditRoles, ",")
		}

		fieldsWithOptions = append(fieldsWithOptions, dto.FieldWithOptionsResponse{
			FieldID:         field.ID.String(),
			ProjectID:       field.ProjectID.String(),
			Name:            field.Name,
			FieldType:       string(field.FieldType),
			Description:     field.Description,
			DisplayOrder:    field.DisplayOrder,
			IsRequired:      field.IsRequired,
			IsSystemDefault: field.IsSystemDefault,
			Config:          config,
			CanEditRoles:    canEditRoles,
			Options:         optionResponses,
			CreatedAt:       field.CreatedAt,
			UpdatedAt:       field.UpdatedAt,
		})
	}

	// 6. Build field type info
	fieldTypes := []dto.FieldTypeInfo{
		{Type: "text", DisplayName: "텍스트", Description: "짧은 텍스트 입력", HasOptions: false},
		{Type: "number", DisplayName: "숫자", Description: "숫자 입력", HasOptions: false},
		{Type: "single_select", DisplayName: "단일 선택", Description: "하나의 옵션 선택", HasOptions: true},
		{Type: "multi_select", DisplayName: "다중 선택", Description: "여러 옵션 선택", HasOptions: true},
		{Type: "date", DisplayName: "날짜", Description: "날짜 선택", HasOptions: false},
		{Type: "datetime", DisplayName: "날짜/시간", Description: "날짜와 시간 선택", HasOptions: false},
		{Type: "single_user", DisplayName: "담당자", Description: "한 명의 사용자 지정", HasOptions: false},
		{Type: "multi_user", DisplayName: "다중 담당자", Description: "여러 사용자 지정", HasOptions: false},
		{Type: "checkbox", DisplayName: "체크박스", Description: "예/아니오 선택", HasOptions: false},
		{Type: "url", DisplayName: "URL", Description: "웹 링크", HasOptions: false},
	}

	// 7. Build response
	return &dto.ProjectInitSettingsResponse{
		Project: dto.ProjectBasicInfo{
			ProjectID:   project.ID.String(),
			Name:        project.Name,
			Description: project.Description,
			WorkspaceID: project.WorkspaceID.String(),
			OwnerID:     project.OwnerID.String(),
			IsPublic:    project.IsPublic,
			CreatedAt:   project.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   project.UpdatedAt.Format(time.RFC3339),
		},
		Fields:        fieldsWithOptions,
		FieldTypes:    fieldTypes,
		DefaultViewID: defaultViewID,
	}, nil
}

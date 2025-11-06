package service

import (
	"board-service/internal/apperrors"
	"board-service/internal/client"
	"board-service/internal/domain"
	"board-service/internal/dto"
	"board-service/internal/repository"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type WorkspaceService interface {
	CreateWorkspace(ownerID string, req *dto.CreateWorkspaceRequest) (*dto.WorkspaceResponse, error)
	GetWorkspace(workspaceID, userID string) (*dto.WorkspaceResponse, error)
	UpdateWorkspace(workspaceID, userID string, req *dto.UpdateWorkspaceRequest) (*dto.WorkspaceResponse, error)
	DeleteWorkspace(workspaceID, userID string) error
	SearchWorkspaces(req *dto.SearchWorkspacesRequest) (*dto.PaginatedWorkspacesResponse, error)

	// Join Request
	CreateJoinRequest(userID string, req *dto.CreateJoinRequestRequest) (*dto.JoinRequestResponse, error)
	GetJoinRequests(workspaceID, userID string, status string) ([]dto.JoinRequestResponse, error)
	UpdateJoinRequest(requestID, userID string, req *dto.UpdateJoinRequestRequest) (*dto.JoinRequestResponse, error)

	// Member
	GetWorkspaceMembers(workspaceID, userID string) ([]dto.WorkspaceMemberResponse, error)
	UpdateMemberRole(workspaceID, memberID, requestUserID string, req *dto.UpdateMemberRoleRequest) (*dto.WorkspaceMemberResponse, error)
	RemoveMember(workspaceID, memberID, requestUserID string) error
	SetDefaultWorkspace(userID string, req *dto.SetDefaultWorkspaceRequest) error
}

type workspaceService struct {
	repo       repository.WorkspaceRepository
	roleRepo   repository.RoleRepository
	userClient client.UserClient
	logger     *zap.Logger
	db         *gorm.DB
}

func NewWorkspaceService(
	repo repository.WorkspaceRepository,
	roleRepo repository.RoleRepository,
	userClient client.UserClient,
	logger *zap.Logger,
	db *gorm.DB,
) WorkspaceService {
	return &workspaceService{
		repo:       repo,
		roleRepo:   roleRepo,
		userClient: userClient,
		logger:     logger,
		db:         db,
	}
}

// CreateWorkspace creates a new workspace
func (s *workspaceService) CreateWorkspace(ownerID string, req *dto.CreateWorkspaceRequest) (*dto.WorkspaceResponse, error) {
	ownerUUID, err := uuid.Parse(ownerID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	// Get OWNER role
	ownerRole, err := s.roleRepo.FindByName("OWNER")
	if err != nil {
		s.logger.Error("Failed to find OWNER role", zap.Error(err))
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "권한 조회 실패", 500)
	}

	// Create workspace and member in transaction
	var workspace *domain.Workspace
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// Create workspace
		workspace = &domain.Workspace{
			Name:        req.Name,
			Description: req.Description,
			OwnerID:     ownerUUID,
			IsPublic:    false,
			IsDeleted:   false,
		}
		if err := s.repo.Create(workspace); err != nil {
			return err
		}

		// Create owner member
		member := &domain.WorkspaceMember{
			WorkspaceID: workspace.ID,
			UserID:      ownerUUID,
			RoleID:      ownerRole.ID,
			JoinedAt:    time.Now(),
			IsDefault:   true,
		}
		if err := s.repo.CreateMember(member); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		s.logger.Error("Failed to create workspace", zap.Error(err))
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "워크스페이스 생성 실패", 500)
	}

	// Fetch user info
	return s.toWorkspaceResponse(workspace)
}

// GetWorkspace retrieves a workspace by ID
func (s *workspaceService) GetWorkspace(workspaceID, userID string) (*dto.WorkspaceResponse, error) {
	wsUUID, err := uuid.Parse(workspaceID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 워크스페이스 ID", 400)
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	// Check if user is member
	_, err = s.repo.FindMemberByUserAndWorkspace(userUUID, wsUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.ErrCodeForbidden, "워크스페이스 멤버가 아닙니다", 403)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "멤버 확인 실패", 500)
	}

	workspace, err := s.repo.FindByID(wsUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.ErrCodeNotFound, "워크스페이스를 찾을 수 없습니다", 404)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "워크스페이스 조회 실패", 500)
	}

	return s.toWorkspaceResponse(workspace)
}

// UpdateWorkspace updates workspace information
func (s *workspaceService) UpdateWorkspace(workspaceID, userID string, req *dto.UpdateWorkspaceRequest) (*dto.WorkspaceResponse, error) {
	wsUUID, err := uuid.Parse(workspaceID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 워크스페이스 ID", 400)
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	// Check if user is OWNER
	if err := s.checkOwnerPermission(userUUID, wsUUID); err != nil {
		return nil, err
	}

	workspace, err := s.repo.FindByID(wsUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.ErrCodeNotFound, "워크스페이스를 찾을 수 없습니다", 404)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "워크스페이스 조회 실패", 500)
	}

	// Update fields
	if req.Name != "" {
		workspace.Name = req.Name
	}
	if req.Description != "" {
		workspace.Description = req.Description
	}

	if err := s.repo.Update(workspace); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "워크스페이스 수정 실패", 500)
	}

	return s.toWorkspaceResponse(workspace)
}

// DeleteWorkspace soft deletes a workspace
func (s *workspaceService) DeleteWorkspace(workspaceID, userID string) error {
	wsUUID, err := uuid.Parse(workspaceID)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 워크스페이스 ID", 400)
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	// Check if user is OWNER
	if err := s.checkOwnerPermission(userUUID, wsUUID); err != nil {
		return err
	}

	if err := s.repo.Delete(wsUUID); err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "워크스페이스 삭제 실패", 500)
	}

	return nil
}

// SearchWorkspaces searches workspaces
func (s *workspaceService) SearchWorkspaces(req *dto.SearchWorkspacesRequest) (*dto.PaginatedWorkspacesResponse, error) {
	// Default values
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 {
		req.Limit = 10
	}

	workspaces, total, err := s.repo.Search(req.Query, req.Page, req.Limit)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "워크스페이스 검색 실패", 500)
	}

	// Convert to response DTOs
	var responses []dto.WorkspaceResponse
	for _, ws := range workspaces {
		resp, err := s.toWorkspaceResponse(&ws)
		if err != nil {
			s.logger.Warn("Failed to convert workspace to response", zap.Error(err))
			continue
		}
		responses = append(responses, *resp)
	}

	return &dto.PaginatedWorkspacesResponse{
		Workspaces: responses,
		Total:      total,
		Page:       req.Page,
		Limit:      req.Limit,
	}, nil
}

// CreateJoinRequest creates a join request
func (s *workspaceService) CreateJoinRequest(userID string, req *dto.CreateJoinRequestRequest) (*dto.JoinRequestResponse, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	wsUUID, err := uuid.Parse(req.WorkspaceID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 워크스페이스 ID", 400)
	}

	// Check if workspace exists
	_, err = s.repo.FindByID(wsUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.ErrCodeNotFound, "워크스페이스를 찾을 수 없습니다", 404)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "워크스페이스 조회 실패", 500)
	}

	// Check if already a member
	_, err = s.repo.FindMemberByUserAndWorkspace(userUUID, wsUUID)
	if err == nil {
		return nil, apperrors.New(apperrors.ErrCodeConflict, "이미 워크스페이스 멤버입니다", 409)
	}

	// Check if request already exists
	existingReq, err := s.repo.FindJoinRequestByUserAndWorkspace(userUUID, wsUUID)
	if err == nil && existingReq.Status == domain.JoinRequestPending {
		return nil, apperrors.New(apperrors.ErrCodeConflict, "이미 참여 신청이 있습니다", 409)
	}

	// Create join request
	joinReq := &domain.WorkspaceJoinRequest{
		WorkspaceID: wsUUID,
		UserID:      userUUID,
		Status:      domain.JoinRequestPending,
		RequestedAt: time.Now(),
	}

	if err := s.repo.CreateJoinRequest(joinReq); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "참여 신청 실패", 500)
	}

	return s.toJoinRequestResponse(joinReq)
}

// GetJoinRequests retrieves join requests for a workspace
func (s *workspaceService) GetJoinRequests(workspaceID, userID string, status string) ([]dto.JoinRequestResponse, error) {
	wsUUID, err := uuid.Parse(workspaceID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 워크스페이스 ID", 400)
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	// Check if user is OWNER or ADMIN
	if err := s.checkAdminPermission(userUUID, wsUUID); err != nil {
		return nil, err
	}

	requests, err := s.repo.FindJoinRequestsByWorkspace(wsUUID, status)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "참여 신청 조회 실패", 500)
	}

	var responses []dto.JoinRequestResponse
	for _, req := range requests {
		resp, err := s.toJoinRequestResponse(&req)
		if err != nil {
			s.logger.Warn("Failed to convert join request to response", zap.Error(err))
			continue
		}
		responses = append(responses, *resp)
	}

	return responses, nil
}

// UpdateJoinRequest approves or rejects a join request
func (s *workspaceService) UpdateJoinRequest(requestID, userID string, req *dto.UpdateJoinRequestRequest) (*dto.JoinRequestResponse, error) {
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

	// Check if user is OWNER or ADMIN
	if err := s.checkAdminPermission(userUUID, joinReq.WorkspaceID); err != nil {
		return nil, err
	}

	// Update status
	now := time.Now()
	joinReq.Status = domain.JoinRequestStatus(req.Status)
	joinReq.ProcessedAt = &now
	joinReq.ProcessedBy = &userUUID

	// If approved, create member
	if req.Status == "APPROVED" {
		memberRole, err := s.roleRepo.FindByName("MEMBER")
		if err != nil {
			return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "권한 조회 실패", 500)
		}

		member := &domain.WorkspaceMember{
			WorkspaceID: joinReq.WorkspaceID,
			UserID:      joinReq.UserID,
			RoleID:      memberRole.ID,
			JoinedAt:    time.Now(),
			IsDefault:   false,
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

// GetWorkspaceMembers retrieves all members of a workspace
func (s *workspaceService) GetWorkspaceMembers(workspaceID, userID string) ([]dto.WorkspaceMemberResponse, error) {
	wsUUID, err := uuid.Parse(workspaceID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 워크스페이스 ID", 400)
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	// Check if user is member
	_, err = s.repo.FindMemberByUserAndWorkspace(userUUID, wsUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.ErrCodeForbidden, "워크스페이스 멤버가 아닙니다", 403)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "멤버 확인 실패", 500)
	}

	members, err := s.repo.FindMembersByWorkspace(wsUUID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "멤버 조회 실패", 500)
	}

	var responses []dto.WorkspaceMemberResponse
	for _, member := range members {
		resp, err := s.toMemberResponse(&member)
		if err != nil {
			s.logger.Warn("Failed to convert member to response", zap.Error(err))
			continue
		}
		responses = append(responses, *resp)
	}

	return responses, nil
}

// UpdateMemberRole updates a member's role
func (s *workspaceService) UpdateMemberRole(workspaceID, memberID, requestUserID string, req *dto.UpdateMemberRoleRequest) (*dto.WorkspaceMemberResponse, error) {
	wsUUID, err := uuid.Parse(workspaceID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 워크스페이스 ID", 400)
	}

	memberUUID, err := uuid.Parse(memberID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 멤버 ID", 400)
	}

	reqUserUUID, err := uuid.Parse(requestUserID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	// Check if requestUser is OWNER
	if err := s.checkOwnerPermission(reqUserUUID, wsUUID); err != nil {
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

// RemoveMember removes a member from workspace
func (s *workspaceService) RemoveMember(workspaceID, memberID, requestUserID string) error {
	wsUUID, err := uuid.Parse(workspaceID)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 워크스페이스 ID", 400)
	}

	memberUUID, err := uuid.Parse(memberID)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 멤버 ID", 400)
	}

	reqUserUUID, err := uuid.Parse(requestUserID)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	// Check if requestUser is OWNER or ADMIN
	if err := s.checkAdminPermission(reqUserUUID, wsUUID); err != nil {
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

// SetDefaultWorkspace sets a workspace as default for the user
func (s *workspaceService) SetDefaultWorkspace(userID string, req *dto.SetDefaultWorkspaceRequest) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	wsUUID, err := uuid.Parse(req.WorkspaceID)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 워크스페이스 ID", 400)
	}

	// Check if user is member
	_, err = s.repo.FindMemberByUserAndWorkspace(userUUID, wsUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.New(apperrors.ErrCodeForbidden, "워크스페이스 멤버가 아닙니다", 403)
		}
		return apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "멤버 확인 실패", 500)
	}

	if err := s.repo.SetDefaultWorkspace(userUUID, wsUUID); err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "대표 워크스페이스 설정 실패", 500)
	}

	return nil
}

// Helper methods

func (s *workspaceService) checkOwnerPermission(userID, workspaceID uuid.UUID) error {
	member, err := s.repo.FindMemberByUserAndWorkspace(userID, workspaceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.New(apperrors.ErrCodeForbidden, "워크스페이스 멤버가 아닙니다", 403)
		}
		return apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "멤버 확인 실패", 500)
	}

	role, err := s.roleRepo.FindByID(member.RoleID)
	if err != nil || role.Name != "OWNER" {
		return apperrors.New(apperrors.ErrCodeForbidden, "OWNER 권한이 필요합니다", 403)
	}

	return nil
}

func (s *workspaceService) checkAdminPermission(userID, workspaceID uuid.UUID) error {
	member, err := s.repo.FindMemberByUserAndWorkspace(userID, workspaceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.New(apperrors.ErrCodeForbidden, "워크스페이스 멤버가 아닙니다", 403)
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

func (s *workspaceService) toWorkspaceResponse(workspace *domain.Workspace) (*dto.WorkspaceResponse, error) {
	response := &dto.WorkspaceResponse{
		ID:          workspace.ID.String(),
		Name:        workspace.Name,
		Description: workspace.Description,
		OwnerID:     workspace.OwnerID.String(),
		CreatedAt:   workspace.CreatedAt,
		UpdatedAt:   workspace.UpdatedAt,
	}

	// Fetch user info from User Service
	ctx := context.Background()
	userInfo, err := s.userClient.GetUser(ctx, workspace.OwnerID.String())
	if err != nil {
		s.logger.Warn("Failed to fetch user info", zap.Error(err), zap.String("user_id", workspace.OwnerID.String()))
		// Continue without user info
	} else {
		response.OwnerName = userInfo.Name
		response.OwnerEmail = userInfo.Email
	}

	return response, nil
}

func (s *workspaceService) toMemberResponse(member *domain.WorkspaceMember) (*dto.WorkspaceMemberResponse, error) {
	// Get role name
	role, err := s.roleRepo.FindByID(member.RoleID)
	if err != nil {
		return nil, err
	}

	response := &dto.WorkspaceMemberResponse{
		ID:          member.ID.String(),
		WorkspaceID: member.WorkspaceID.String(),
		UserID:      member.UserID.String(),
		RoleName:    role.Name,
		IsDefault:   member.IsDefault,
		JoinedAt:    member.JoinedAt,
	}

	// Fetch user info from User Service
	ctx := context.Background()
	userInfo, err := s.userClient.GetUser(ctx, member.UserID.String())
	if err != nil {
		s.logger.Warn("Failed to fetch user info", zap.Error(err), zap.String("user_id", member.UserID.String()))
	} else {
		response.UserName = userInfo.Name
		response.UserEmail = userInfo.Email
	}

	return response, nil
}

func (s *workspaceService) toJoinRequestResponse(req *domain.WorkspaceJoinRequest) (*dto.JoinRequestResponse, error) {
	response := &dto.JoinRequestResponse{
		ID:          req.ID.String(),
		WorkspaceID: req.WorkspaceID.String(),
		UserID:      req.UserID.String(),
		Status:      string(req.Status),
		RequestedAt: req.RequestedAt,
		UpdatedAt:   req.UpdatedAt,
	}

	// Fetch user info from User Service
	ctx := context.Background()
	userInfo, err := s.userClient.GetUser(ctx, req.UserID.String())
	if err != nil {
		s.logger.Warn("Failed to fetch user info", zap.Error(err), zap.String("user_id", req.UserID.String()))
	} else {
		response.UserName = userInfo.Name
		response.UserEmail = userInfo.Email
	}

	return response, nil
}

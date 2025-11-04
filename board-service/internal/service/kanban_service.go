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

type KanbanService interface {
	CreateKanban(userID string, req *dto.CreateKanbanRequest) (*dto.KanbanResponse, error)
	GetKanban(kanbanID, userID string) (*dto.KanbanResponse, error)
	GetKanbans(userID string, req *dto.GetKanbansRequest) (*dto.PaginatedKanbansResponse, error)
	UpdateKanban(kanbanID, userID string, req *dto.UpdateKanbanRequest) (*dto.KanbanResponse, error)
	DeleteKanban(kanbanID, userID string) error
}

type kanbanService struct {
	repo            repository.KanbanRepository
	projectRepo     repository.ProjectRepository
	customFieldRepo repository.CustomFieldRepository
	roleRepo        repository.RoleRepository
	userClient      *client.UserClient
	logger          *zap.Logger
	db              *gorm.DB
}

func NewKanbanService(
	repo repository.KanbanRepository,
	projectRepo repository.ProjectRepository,
	customFieldRepo repository.CustomFieldRepository,
	roleRepo repository.RoleRepository,
	userClient *client.UserClient,
	logger *zap.Logger,
	db *gorm.DB,
) KanbanService {
	return &kanbanService{
		repo:            repo,
		projectRepo:     projectRepo,
		customFieldRepo: customFieldRepo,
		roleRepo:        roleRepo,
		userClient:      userClient,
		logger:          logger,
		db:              db,
	}
}

// ==================== Create Kanban ====================

func (s *kanbanService) CreateKanban(userID string, req *dto.CreateKanbanRequest) (*dto.KanbanResponse, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	projectUUID, err := uuid.Parse(req.ProjectID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 프로젝트 ID", 400)
	}

	// 1. Check if user is project member
	_, err = s.projectRepo.FindMemberByUserAndProject(userUUID, projectUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.ErrCodeForbidden, "프로젝트 멤버가 아닙니다", 403)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "멤버 확인 실패", 500)
	}

	// 2. Validate Stage (required)
	stageUUID, err := uuid.Parse(req.StageID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 진행단계 ID", 400)
	}

	stage, err := s.customFieldRepo.FindCustomStageByID(stageUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.ErrCodeNotFound, "진행단계를 찾을 수 없습니다", 404)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "진행단계 조회 실패", 500)
	}

	if stage.ProjectID != projectUUID {
		return nil, apperrors.New(apperrors.ErrCodeForbidden, "다른 프로젝트의 진행단계입니다", 403)
	}

	// 3. Validate Importance (optional)
	var importance *domain.CustomImportance
	var importanceUUID *uuid.UUID
	if req.ImportanceID != nil {
		parsedImportanceUUID, err := uuid.Parse(*req.ImportanceID)
		if err != nil {
			return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 중요도 ID", 400)
		}
		importanceUUID = &parsedImportanceUUID

		importance, err = s.customFieldRepo.FindCustomImportanceByID(parsedImportanceUUID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, apperrors.New(apperrors.ErrCodeNotFound, "중요도를 찾을 수 없습니다", 404)
			}
			return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "중요도 조회 실패", 500)
		}

		if importance.ProjectID != projectUUID {
			return nil, apperrors.New(apperrors.ErrCodeForbidden, "다른 프로젝트의 중요도입니다", 403)
		}
	}

	// 4. Validate Roles (required, at least 1)
	roleUUIDs := make([]uuid.UUID, 0, len(req.RoleIDs))
	roles := make([]*domain.CustomRole, 0, len(req.RoleIDs))
	for _, roleID := range req.RoleIDs {
		roleUUID, err := uuid.Parse(roleID)
		if err != nil {
			return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 역할 ID", 400)
		}

		role, err := s.customFieldRepo.FindCustomRoleByID(roleUUID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, apperrors.New(apperrors.ErrCodeNotFound, "역할을 찾을 수 없습니다", 404)
			}
			return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "역할 조회 실패", 500)
		}

		if role.ProjectID != projectUUID {
			return nil, apperrors.New(apperrors.ErrCodeForbidden, "다른 프로젝트의 역할입니다", 403)
		}

		roleUUIDs = append(roleUUIDs, roleUUID)
		roles = append(roles, role)
	}

	// 5. Validate Assignee (optional)
	var assigneeUUID *uuid.UUID
	if req.AssigneeID != nil {
		parsedAssigneeUUID, err := uuid.Parse(*req.AssigneeID)
		if err != nil {
			return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 담당자 ID", 400)
		}
		assigneeUUID = &parsedAssigneeUUID

		_, err = s.projectRepo.FindMemberByUserAndProject(parsedAssigneeUUID, projectUUID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, apperrors.New(apperrors.ErrCodeNotFound, "담당자가 프로젝트 멤버가 아닙니다", 404)
			}
			return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "담당자 확인 실패", 500)
		}
	}

	// 6. Parse DueDate (optional)
	var dueDate *time.Time
	if req.DueDate != nil {
		parsed, err := time.Parse(time.RFC3339, *req.DueDate)
		if err != nil {
			return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 마감일 형식입니다 (ISO 8601 required)", 400)
		}
		dueDate = &parsed
	}

	// 7. Create Kanban in transaction
	var kanban *domain.Kanban
	err = s.db.Transaction(func(tx *gorm.DB) error {
		kanban = &domain.Kanban{
			ProjectID:    projectUUID,
			Title:        req.Title,
			Content:      req.Content,
			StageID:      stageUUID,
			ImportanceID: importanceUUID,
			AssigneeID:   assigneeUUID,
			AuthorID:     userUUID,
			DueDate:      dueDate,
		}

		if err := s.repo.Create(kanban); err != nil {
			s.logger.Error("Failed to create kanban", zap.Error(err))
			return err
		}

		// Create kanban_roles (many-to-many)
		if err := s.repo.CreateKanbanRoles(kanban.ID, roleUUIDs); err != nil {
			s.logger.Error("Failed to create kanban roles", zap.Error(err))
			return err
		}

		return nil
	})

	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "칸반 생성 실패", 500)
	}

	// 8. Build response with user info
	return s.buildKanbanResponse(kanban, stage, importance, roles)
}

// ==================== Get Single Kanban ====================

func (s *kanbanService) GetKanban(kanbanID, userID string) (*dto.KanbanResponse, error) {
	kanbanUUID, err := uuid.Parse(kanbanID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 칸반 ID", 400)
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	// 1. Find kanban
	kanban, err := s.repo.FindByID(kanbanUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.ErrCodeNotFound, "칸반을 찾을 수 없습니다", 404)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "칸반 조회 실패", 500)
	}

	// 2. Check if user is project member
	_, err = s.projectRepo.FindMemberByUserAndProject(userUUID, kanban.ProjectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.ErrCodeForbidden, "프로젝트 멤버가 아닙니다", 403)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "멤버 확인 실패", 500)
	}

	// 3. Fetch related data
	stage, err := s.customFieldRepo.FindCustomStageByID(kanban.StageID)
	if err != nil {
		s.logger.Warn("Failed to fetch stage", zap.Error(err), zap.String("stage_id", kanban.StageID.String()))
	}

	var importance *domain.CustomImportance
	if kanban.ImportanceID != nil {
		importance, err = s.customFieldRepo.FindCustomImportanceByID(*kanban.ImportanceID)
		if err != nil {
			s.logger.Warn("Failed to fetch importance", zap.Error(err), zap.String("importance_id", kanban.ImportanceID.String()))
		}
	}

	kanbanRoles, err := s.repo.FindRolesByKanban(kanban.ID)
	if err != nil {
		s.logger.Warn("Failed to fetch kanban roles", zap.Error(err))
	}

	roles := make([]*domain.CustomRole, 0, len(kanbanRoles))
	for _, kr := range kanbanRoles {
		role, err := s.customFieldRepo.FindCustomRoleByID(kr.CustomRoleID)
		if err == nil && role != nil {
			roles = append(roles, role)
		}
	}

	// 4. Build response
	return s.buildKanbanResponse(kanban, stage, importance, roles)
}

// ==================== Get Kanbans (List with Filters) ====================

func (s *kanbanService) GetKanbans(userID string, req *dto.GetKanbansRequest) (*dto.PaginatedKanbansResponse, error) {
	projectUUID, err := uuid.Parse(req.ProjectID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 프로젝트 ID", 400)
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	// 1. Check if user is project member
	_, err = s.projectRepo.FindMemberByUserAndProject(userUUID, projectUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.ErrCodeForbidden, "프로젝트 멤버가 아닙니다", 403)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "멤버 확인 실패", 500)
	}

	// 2. Build filters
	filters := repository.KanbanFilters{}
	if req.StageID != "" {
		stageUUID, err := uuid.Parse(req.StageID)
		if err == nil {
			filters.StageID = stageUUID
		}
	}
	if req.RoleID != "" {
		roleUUID, err := uuid.Parse(req.RoleID)
		if err == nil {
			filters.RoleID = roleUUID
		}
	}
	if req.ImportanceID != "" {
		importanceUUID, err := uuid.Parse(req.ImportanceID)
		if err == nil {
			filters.ImportanceID = importanceUUID
		}
	}
	if req.AssigneeID != "" {
		assigneeUUID, err := uuid.Parse(req.AssigneeID)
		if err == nil {
			filters.AssigneeID = assigneeUUID
		}
	}
	if req.AuthorID != "" {
		authorUUID, err := uuid.Parse(req.AuthorID)
		if err == nil {
			filters.AuthorID = authorUUID
		}
	}

	// 3. Default pagination
	page := req.Page
	if page < 1 {
		page = 1
	}
	limit := req.Limit
	if limit < 1 {
		limit = 20
	}

	// 4. Fetch kanbans
	kanbans, total, err := s.repo.FindByProject(projectUUID, filters, page, limit)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "칸반 조회 실패", 500)
	}

	// 5. Build responses (batch processing)
	responses := make([]dto.KanbanResponse, 0, len(kanbans))
	for _, kanban := range kanbans {
		stage, _ := s.customFieldRepo.FindCustomStageByID(kanban.StageID)

		var importance *domain.CustomImportance
		if kanban.ImportanceID != nil {
			importance, _ = s.customFieldRepo.FindCustomImportanceByID(*kanban.ImportanceID)
		}

		kanbanRoles, _ := s.repo.FindRolesByKanban(kanban.ID)
		roles := make([]*domain.CustomRole, 0, len(kanbanRoles))
		for _, kr := range kanbanRoles {
			role, err := s.customFieldRepo.FindCustomRoleByID(kr.CustomRoleID)
			if err == nil && role != nil {
				roles = append(roles, role)
			}
		}

		response, err := s.buildKanbanResponse(&kanban, stage, importance, roles)
		if err == nil && response != nil {
			responses = append(responses, *response)
		}
	}

	return &dto.PaginatedKanbansResponse{
		Kanbans: responses,
		Total:   total,
		Page:    page,
		Limit:   limit,
	}, nil
}

// ==================== Update Kanban ====================

func (s *kanbanService) UpdateKanban(kanbanID, userID string, req *dto.UpdateKanbanRequest) (*dto.KanbanResponse, error) {
	kanbanUUID, err := uuid.Parse(kanbanID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 칸반 ID", 400)
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	// 1. Find kanban
	kanban, err := s.repo.FindByID(kanbanUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.ErrCodeNotFound, "칸반을 찾을 수 없습니다", 404)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "칸반 조회 실패", 500)
	}

	// 2. Check permission (author or ADMIN+)
	member, err := s.projectRepo.FindMemberByUserAndProject(userUUID, kanban.ProjectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.ErrCodeForbidden, "프로젝트 멤버가 아닙니다", 403)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "멤버 확인 실패", 500)
	}

	// Get member role
	role, err := s.roleRepo.FindByID(member.RoleID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "권한 조회 실패", 500)
	}

	// Check if user is author or has ADMIN+ permission
	if kanban.AuthorID != userUUID && role.Name == "MEMBER" {
		return nil, apperrors.New(apperrors.ErrCodeForbidden, "수정 권한이 없습니다", 403)
	}

	// 3. Update fields
	if req.Title != "" {
		kanban.Title = req.Title
	}
	if req.Content != "" {
		kanban.Content = req.Content
	}

	if req.StageID != "" {
		stageUUID, err := uuid.Parse(req.StageID)
		if err != nil {
			return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 진행단계 ID", 400)
		}

		stage, err := s.customFieldRepo.FindCustomStageByID(stageUUID)
		if err != nil || stage.ProjectID != kanban.ProjectID {
			return nil, apperrors.New(apperrors.ErrCodeNotFound, "진행단계를 찾을 수 없습니다", 404)
		}
		kanban.StageID = stageUUID
	}

	if req.ImportanceID != nil {
		importanceUUID, err := uuid.Parse(*req.ImportanceID)
		if err != nil {
			return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 중요도 ID", 400)
		}

		importance, err := s.customFieldRepo.FindCustomImportanceByID(importanceUUID)
		if err != nil || importance.ProjectID != kanban.ProjectID {
			return nil, apperrors.New(apperrors.ErrCodeNotFound, "중요도를 찾을 수 없습니다", 404)
		}
		kanban.ImportanceID = &importanceUUID
	}

	if req.AssigneeID != nil {
		assigneeUUID, err := uuid.Parse(*req.AssigneeID)
		if err != nil {
			return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 담당자 ID", 400)
		}

		_, err = s.projectRepo.FindMemberByUserAndProject(assigneeUUID, kanban.ProjectID)
		if err != nil {
			return nil, apperrors.New(apperrors.ErrCodeNotFound, "담당자가 프로젝트 멤버가 아닙니다", 404)
		}
		kanban.AssigneeID = &assigneeUUID
	}

	if req.DueDate != nil {
		parsed, err := time.Parse(time.RFC3339, *req.DueDate)
		if err != nil {
			return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 마감일 형식입니다 (ISO 8601 required)", 400)
		}
		kanban.DueDate = &parsed
	}

	// 4. Update roles if provided
	if len(req.RoleIDs) > 0 {
		// Validate all roles first
		roleUUIDs := make([]uuid.UUID, 0, len(req.RoleIDs))
		for _, roleID := range req.RoleIDs {
			roleUUID, err := uuid.Parse(roleID)
			if err != nil {
				return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 역할 ID", 400)
			}

			role, err := s.customFieldRepo.FindCustomRoleByID(roleUUID)
			if err != nil || role.ProjectID != kanban.ProjectID {
				return nil, apperrors.New(apperrors.ErrCodeNotFound, "역할을 찾을 수 없습니다", 404)
			}

			roleUUIDs = append(roleUUIDs, roleUUID)
		}

		// Delete existing roles and create new ones in transaction
		err = s.db.Transaction(func(tx *gorm.DB) error {
			if err := s.repo.DeleteKanbanRolesByKanban(kanban.ID); err != nil {
				return err
			}
			if err := s.repo.CreateKanbanRoles(kanban.ID, roleUUIDs); err != nil {
				return err
			}
			return nil
		})

		if err != nil {
			return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "역할 업데이트 실패", 500)
		}
	}

	// 5. Save kanban
	if err := s.repo.Update(kanban); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "칸반 수정 실패", 500)
	}

	// 6. Return updated kanban
	return s.GetKanban(kanban.ID.String(), userID)
}

// ==================== Delete Kanban (Soft) ====================

func (s *kanbanService) DeleteKanban(kanbanID, userID string) error {
	kanbanUUID, err := uuid.Parse(kanbanID)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 칸반 ID", 400)
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	// 1. Find kanban
	kanban, err := s.repo.FindByID(kanbanUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.New(apperrors.ErrCodeNotFound, "칸반을 찾을 수 없습니다", 404)
		}
		return apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "칸반 조회 실패", 500)
	}

	// 2. Check permission (author or ADMIN+)
	member, err := s.projectRepo.FindMemberByUserAndProject(userUUID, kanban.ProjectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.New(apperrors.ErrCodeForbidden, "프로젝트 멤버가 아닙니다", 403)
		}
		return apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "멤버 확인 실패", 500)
	}

	// Get member role
	role, err := s.roleRepo.FindByID(member.RoleID)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "권한 조회 실패", 500)
	}

	// Check if user is author or has ADMIN+ permission
	if kanban.AuthorID != userUUID && role.Name == "MEMBER" {
		return apperrors.New(apperrors.ErrCodeForbidden, "삭제 권한이 없습니다", 403)
	}

	// 3. Soft delete
	if err := s.repo.Delete(kanban.ID); err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "칸반 삭제 실패", 500)
	}

	return nil
}

// ==================== Helper: Build Kanban Response ====================

func (s *kanbanService) buildKanbanResponse(
	kanban *domain.Kanban,
	stage *domain.CustomStage,
	importance *domain.CustomImportance,
	roles []*domain.CustomRole,
) (*dto.KanbanResponse, error) {
	// Collect user IDs for batch query
	userIDs := []string{kanban.AuthorID.String()}
	if kanban.AssigneeID != nil {
		userIDs = append(userIDs, kanban.AssigneeID.String())
	}

	// Fetch users from User Service (batch)
	ctx := context.Background()
	users, err := s.userClient.GetUsersBatch(ctx, userIDs)
	if err != nil {
		s.logger.Warn("Failed to fetch users from User Service", zap.Error(err))
		users = []client.UserInfo{} // Continue without user info
	}

	// Convert to map for easy lookup
	userMap := make(map[string]client.UserInfo)
	for _, user := range users {
		userMap[user.UserID] = user
	}

	// Build response
	response := &dto.KanbanResponse{
		ID:        kanban.ID.String(),
		ProjectID: kanban.ProjectID.String(),
		Title:     kanban.Title,
		Content:   kanban.Content,
		DueDate:   kanban.DueDate,
		CreatedAt: kanban.CreatedAt,
		UpdatedAt: kanban.UpdatedAt,
	}

	// Stage
	if stage != nil {
		response.Stage = dto.CustomStageResponse{
			ID:              stage.ID.String(),
			ProjectID:       stage.ProjectID.String(),
			Name:            stage.Name,
			Color:           stage.Color,
			IsSystemDefault: stage.IsSystemDefault,
			DisplayOrder:    stage.DisplayOrder,
			CreatedAt:       stage.CreatedAt,
			UpdatedAt:       stage.UpdatedAt,
		}
	}

	// Importance
	if importance != nil {
		response.Importance = &dto.CustomImportanceResponse{
			ID:              importance.ID.String(),
			ProjectID:       importance.ProjectID.String(),
			Name:            importance.Name,
			Color:           importance.Color,
			IsSystemDefault: importance.IsSystemDefault,
			DisplayOrder:    importance.DisplayOrder,
			CreatedAt:       importance.CreatedAt,
			UpdatedAt:       importance.UpdatedAt,
		}
	}

	// Roles
	roleResponses := make([]dto.CustomRoleResponse, 0, len(roles))
	for _, role := range roles {
		roleResponses = append(roleResponses, dto.CustomRoleResponse{
			ID:              role.ID.String(),
			ProjectID:       role.ProjectID.String(),
			Name:            role.Name,
			Color:           role.Color,
			IsSystemDefault: role.IsSystemDefault,
			DisplayOrder:    role.DisplayOrder,
			CreatedAt:       role.CreatedAt,
			UpdatedAt:       role.UpdatedAt,
		})
	}
	response.Roles = roleResponses

	// Author
	if author, ok := userMap[kanban.AuthorID.String()]; ok {
		response.Author = dto.UserInfo{
			UserID:   author.UserID,
			Name:     author.Name,
			Email:    author.Email,
			IsActive: author.IsActive,
		}
	} else {
		// Fallback if user not found
		response.Author = dto.UserInfo{
			UserID:   kanban.AuthorID.String(),
			Name:     "Unknown User",
			Email:    "",
			IsActive: false,
		}
	}

	// Assignee
	if kanban.AssigneeID != nil {
		if assignee, ok := userMap[kanban.AssigneeID.String()]; ok {
			response.Assignee = &dto.UserInfo{
				UserID:   assignee.UserID,
				Name:     assignee.Name,
				Email:    assignee.Email,
				IsActive: assignee.IsActive,
			}
		} else {
			// Fallback if user not found
			response.Assignee = &dto.UserInfo{
				UserID:   kanban.AssigneeID.String(),
				Name:     "Unknown User",
				Email:    "",
				IsActive: false,
			}
		}
	}

	return response, nil
}

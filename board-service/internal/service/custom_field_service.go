package service

import (
	"board-service/internal/apperrors"
	"board-service/internal/domain"
	"board-service/internal/dto"
	"board-service/internal/repository"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type CustomFieldService interface {
	// Custom Roles
	CreateCustomRole(userID string, req *dto.CreateCustomRoleRequest) (*dto.CustomRoleResponse, error)
	GetCustomRoles(projectID, userID string) ([]dto.CustomRoleResponse, error)
	GetCustomRole(roleID, userID string) (*dto.CustomRoleResponse, error)
	UpdateCustomRole(roleID, userID string, req *dto.UpdateCustomRoleRequest) (*dto.CustomRoleResponse, error)
	DeleteCustomRole(roleID, userID string) error
	UpdateCustomRoleOrder(projectID, userID string, req *dto.UpdateCustomRoleOrderRequest) error

	// Custom Stages
	CreateCustomStage(userID string, req *dto.CreateCustomStageRequest) (*dto.CustomStageResponse, error)
	GetCustomStages(projectID, userID string) ([]dto.CustomStageResponse, error)
	GetCustomStage(stageID, userID string) (*dto.CustomStageResponse, error)
	UpdateCustomStage(stageID, userID string, req *dto.UpdateCustomStageRequest) (*dto.CustomStageResponse, error)
	DeleteCustomStage(stageID, userID string) error
	UpdateCustomStageOrder(projectID, userID string, req *dto.UpdateCustomStageOrderRequest) error

	// Custom Importance
	CreateCustomImportance(userID string, req *dto.CreateCustomImportanceRequest) (*dto.CustomImportanceResponse, error)
	GetCustomImportances(projectID, userID string) ([]dto.CustomImportanceResponse, error)
	GetCustomImportance(importanceID, userID string) (*dto.CustomImportanceResponse, error)
	UpdateCustomImportance(importanceID, userID string, req *dto.UpdateCustomImportanceRequest) (*dto.CustomImportanceResponse, error)
	DeleteCustomImportance(importanceID, userID string) error
	UpdateCustomImportanceOrder(projectID, userID string, req *dto.UpdateCustomImportanceOrderRequest) error

	// Default fields creation (called when creating a project)
	CreateDefaultCustomFields(projectID uuid.UUID) error
}

type customFieldService struct {
	repo        repository.CustomFieldRepository
	projectRepo repository.ProjectRepository
	roleRepo    repository.RoleRepository
	boardRepo  repository.BoardRepository
	logger      *zap.Logger
	db          *gorm.DB
}

func NewCustomFieldService(
	repo repository.CustomFieldRepository,
	projectRepo repository.ProjectRepository,
	roleRepo repository.RoleRepository,
	boardRepo repository.BoardRepository,
	logger *zap.Logger,
	db *gorm.DB,
) CustomFieldService {
	return &customFieldService{
		repo:        repo,
		projectRepo: projectRepo,
		roleRepo:    roleRepo,
		boardRepo:  boardRepo,
		logger:      logger,
		db:          db,
	}
}

// ==================== Default Fields Creation ====================

func (s *customFieldService) CreateDefaultCustomFields(projectID uuid.UUID) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 1. Custom Roles 기본값: "없음"
		defaultRole := &domain.CustomRole{
			ProjectID:       projectID,
			Name:            "없음",
			Color:           "#CCCCCC",
			IsSystemDefault: true,
			DisplayOrder:    0,
		}
		if err := s.repo.CreateCustomRole(defaultRole); err != nil {
			s.logger.Error("Failed to create default custom role", zap.Error(err))
			return err
		}

		// 2. Custom Stages 기본값: "없음", "대기", "진행중", "완료"
		defaultStages := []domain.CustomStage{
			{ProjectID: projectID, Name: "없음", Color: "#CCCCCC", IsSystemDefault: true, DisplayOrder: 0},
			{ProjectID: projectID, Name: "대기", Color: "#FFA500", IsSystemDefault: true, DisplayOrder: 1},
			{ProjectID: projectID, Name: "진행중", Color: "#4169E1", IsSystemDefault: true, DisplayOrder: 2},
			{ProjectID: projectID, Name: "완료", Color: "#32CD32", IsSystemDefault: true, DisplayOrder: 3},
		}
		for _, stage := range defaultStages {
			stageCopy := stage
			if err := s.repo.CreateCustomStage(&stageCopy); err != nil {
				s.logger.Error("Failed to create default custom stage", zap.Error(err), zap.String("name", stage.Name))
				return err
			}
		}

		// 3. Custom Importance 기본값: "없음", "낮음", "보통", "높음", "긴급"
		defaultImportances := []domain.CustomImportance{
			{ProjectID: projectID, Name: "없음", Color: "#CCCCCC", IsSystemDefault: true, DisplayOrder: 0},
			{ProjectID: projectID, Name: "낮음", Color: "#90EE90", IsSystemDefault: true, DisplayOrder: 1},
			{ProjectID: projectID, Name: "보통", Color: "#FFD700", IsSystemDefault: true, DisplayOrder: 2},
			{ProjectID: projectID, Name: "높음", Color: "#FF8C00", IsSystemDefault: true, DisplayOrder: 3},
			{ProjectID: projectID, Name: "긴급", Color: "#DC143C", IsSystemDefault: true, DisplayOrder: 4},
		}
		for _, importance := range defaultImportances {
			importanceCopy := importance
			if err := s.repo.CreateCustomImportance(&importanceCopy); err != nil {
				s.logger.Error("Failed to create default custom importance", zap.Error(err), zap.String("name", importance.Name))
				return err
			}
		}

		return nil
	})
}

// ==================== Custom Roles ====================

func (s *customFieldService) CreateCustomRole(userID string, req *dto.CreateCustomRoleRequest) (*dto.CustomRoleResponse, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	projectUUID, err := uuid.Parse(req.ProjectID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 프로젝트 ID", 400)
	}

	// Check ADMIN permission
	if err := s.checkProjectAdminPermission(userUUID, projectUUID); err != nil {
		return nil, err
	}

	// Check duplicate name
	existing, _ := s.repo.FindCustomRoleByProjectAndName(projectUUID, req.Name)
	if existing != nil {
		return nil, apperrors.New(apperrors.ErrCodeConflict, "이미 존재하는 역할 이름입니다", 409)
	}

	// Calculate display order (max + 1)
	roles, err := s.repo.FindCustomRolesByProject(projectUUID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "역할 조회 실패", 500)
	}

	maxOrder := 0
	for _, role := range roles {
		if role.DisplayOrder > maxOrder {
			maxOrder = role.DisplayOrder
		}
	}

	// Create role
	role := &domain.CustomRole{
		ProjectID:       projectUUID,
		Name:            req.Name,
		Color:           req.Color,
		IsSystemDefault: false,
		DisplayOrder:    maxOrder + 1,
	}

	if err := s.repo.CreateCustomRole(role); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "역할 생성 실패", 500)
	}

	return s.toCustomRoleResponse(role), nil
}

func (s *customFieldService) GetCustomRoles(projectID, userID string) ([]dto.CustomRoleResponse, error) {
	projectUUID, err := uuid.Parse(projectID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 프로젝트 ID", 400)
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	// Check if user is project member
	if err := s.checkProjectMember(userUUID, projectUUID); err != nil {
		return nil, err
	}

	roles, err := s.repo.FindCustomRolesByProject(projectUUID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "역할 조회 실패", 500)
	}

	var responses []dto.CustomRoleResponse
	for _, role := range roles {
		responses = append(responses, *s.toCustomRoleResponse(&role))
	}

	return responses, nil
}

func (s *customFieldService) GetCustomRole(roleID, userID string) (*dto.CustomRoleResponse, error) {
	roleUUID, err := uuid.Parse(roleID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 역할 ID", 400)
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	role, err := s.repo.FindCustomRoleByID(roleUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.ErrCodeNotFound, "역할을 찾을 수 없습니다", 404)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "역할 조회 실패", 500)
	}

	// Check if user is project member
	if err := s.checkProjectMember(userUUID, role.ProjectID); err != nil {
		return nil, err
	}

	return s.toCustomRoleResponse(role), nil
}

func (s *customFieldService) UpdateCustomRole(roleID, userID string, req *dto.UpdateCustomRoleRequest) (*dto.CustomRoleResponse, error) {
	roleUUID, err := uuid.Parse(roleID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 역할 ID", 400)
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	role, err := s.repo.FindCustomRoleByID(roleUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.ErrCodeNotFound, "역할을 찾을 수 없습니다", 404)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "역할 조회 실패", 500)
	}

	// Check ADMIN permission
	if err := s.checkProjectAdminPermission(userUUID, role.ProjectID); err != nil {
		return nil, err
	}

	// Cannot update system default
	if role.IsSystemDefault {
		return nil, apperrors.New(apperrors.ErrCodeForbidden, "시스템 기본 역할은 수정할 수 없습니다", 403)
	}

	// Update fields
	if req.Name != "" {
		// Check duplicate name
		existing, _ := s.repo.FindCustomRoleByProjectAndName(role.ProjectID, req.Name)
		if existing != nil && existing.ID != role.ID {
			return nil, apperrors.New(apperrors.ErrCodeConflict, "이미 존재하는 역할 이름입니다", 409)
		}
		role.Name = req.Name
	}
	if req.Color != "" {
		role.Color = req.Color
	}

	if err := s.repo.UpdateCustomRole(role); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "역할 수정 실패", 500)
	}

	return s.toCustomRoleResponse(role), nil
}

func (s *customFieldService) DeleteCustomRole(roleID, userID string) error {
	roleUUID, err := uuid.Parse(roleID)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 역할 ID", 400)
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	role, err := s.repo.FindCustomRoleByID(roleUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.New(apperrors.ErrCodeNotFound, "역할을 찾을 수 없습니다", 404)
		}
		return apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "역할 조회 실패", 500)
	}

	// Check ADMIN permission
	if err := s.checkProjectAdminPermission(userUUID, role.ProjectID); err != nil {
		return err
	}

	// Cannot delete system default
	if role.IsSystemDefault {
		return apperrors.New(apperrors.ErrCodeForbidden, "시스템 기본 역할은 삭제할 수 없습니다", 403)
	}

	// Phase 5 IMPLEMENTED: Update boards using this role to "없음" (default role)
	defaultRole, err := s.repo.FindCustomRoleByProjectAndName(role.ProjectID, "없음")
	if err != nil {
		s.logger.Error("Failed to find default role", zap.Error(err), zap.String("project_id", role.ProjectID.String()))
		return apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "기본 역할 조회 실패", 500)
	}

	if err := s.boardRepo.UpdateBoardsRoleToDefault(roleUUID, defaultRole.ID); err != nil {
		s.logger.Error("Failed to update boards role to default", zap.Error(err))
		return apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "칸반 역할 업데이트 실패", 500)
	}

	if err := s.repo.DeleteCustomRole(roleUUID); err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "역할 삭제 실패", 500)
	}

	return nil
}

func (s *customFieldService) UpdateCustomRoleOrder(projectID, userID string, req *dto.UpdateCustomRoleOrderRequest) error {
	projectUUID, err := uuid.Parse(projectID)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 프로젝트 ID", 400)
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	// Check ADMIN permission
	if err := s.checkProjectAdminPermission(userUUID, projectUUID); err != nil {
		return err
	}

	// Validate and prepare roles for update
	roles := make([]domain.CustomRole, 0, len(req.RoleOrders))
	for _, order := range req.RoleOrders {
		roleUUID, err := uuid.Parse(order.RoleID)
		if err != nil {
			return apperrors.Wrap(err, apperrors.ErrCodeBadRequest, fmt.Sprintf("잘못된 역할 ID: %s", order.RoleID), 400)
		}

		role, err := s.repo.FindCustomRoleByID(roleUUID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return apperrors.New(apperrors.ErrCodeNotFound, fmt.Sprintf("역할을 찾을 수 없습니다: %s", order.RoleID), 404)
			}
			return apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "역할 조회 실패", 500)
		}

		// Verify project ownership
		if role.ProjectID != projectUUID {
			return apperrors.New(apperrors.ErrCodeForbidden, "다른 프로젝트의 역할입니다", 403)
		}

		role.DisplayOrder = order.DisplayOrder
		roles = append(roles, *role)
	}

	// Update orders
	if err := s.repo.UpdateCustomRoleOrders(roles); err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "역할 순서 업데이트 실패", 500)
	}

	return nil
}

// ==================== Custom Stages ====================
// Similar implementation pattern as Custom Roles

func (s *customFieldService) CreateCustomStage(userID string, req *dto.CreateCustomStageRequest) (*dto.CustomStageResponse, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	projectUUID, err := uuid.Parse(req.ProjectID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 프로젝트 ID", 400)
	}

	if err := s.checkProjectAdminPermission(userUUID, projectUUID); err != nil {
		return nil, err
	}

	existing, _ := s.repo.FindCustomStageByProjectAndName(projectUUID, req.Name)
	if existing != nil {
		return nil, apperrors.New(apperrors.ErrCodeConflict, "이미 존재하는 단계 이름입니다", 409)
	}

	stages, err := s.repo.FindCustomStagesByProject(projectUUID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "단계 조회 실패", 500)
	}

	maxOrder := 0
	for _, stage := range stages {
		if stage.DisplayOrder > maxOrder {
			maxOrder = stage.DisplayOrder
		}
	}

	stage := &domain.CustomStage{
		ProjectID:       projectUUID,
		Name:            req.Name,
		Color:           req.Color,
		IsSystemDefault: false,
		DisplayOrder:    maxOrder + 1,
	}

	if err := s.repo.CreateCustomStage(stage); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "단계 생성 실패", 500)
	}

	return s.toCustomStageResponse(stage), nil
}

func (s *customFieldService) GetCustomStages(projectID, userID string) ([]dto.CustomStageResponse, error) {
	projectUUID, err := uuid.Parse(projectID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 프로젝트 ID", 400)
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	if err := s.checkProjectMember(userUUID, projectUUID); err != nil {
		return nil, err
	}

	stages, err := s.repo.FindCustomStagesByProject(projectUUID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "단계 조회 실패", 500)
	}

	var responses []dto.CustomStageResponse
	for _, stage := range stages {
		responses = append(responses, *s.toCustomStageResponse(&stage))
	}

	return responses, nil
}

func (s *customFieldService) GetCustomStage(stageID, userID string) (*dto.CustomStageResponse, error) {
	stageUUID, err := uuid.Parse(stageID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 단계 ID", 400)
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	stage, err := s.repo.FindCustomStageByID(stageUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.ErrCodeNotFound, "단계를 찾을 수 없습니다", 404)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "단계 조회 실패", 500)
	}

	if err := s.checkProjectMember(userUUID, stage.ProjectID); err != nil {
		return nil, err
	}

	return s.toCustomStageResponse(stage), nil
}

func (s *customFieldService) UpdateCustomStage(stageID, userID string, req *dto.UpdateCustomStageRequest) (*dto.CustomStageResponse, error) {
	stageUUID, err := uuid.Parse(stageID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 단계 ID", 400)
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	stage, err := s.repo.FindCustomStageByID(stageUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.ErrCodeNotFound, "단계를 찾을 수 없습니다", 404)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "단계 조회 실패", 500)
	}

	if err := s.checkProjectAdminPermission(userUUID, stage.ProjectID); err != nil {
		return nil, err
	}

	if stage.IsSystemDefault {
		return nil, apperrors.New(apperrors.ErrCodeForbidden, "시스템 기본 단계는 수정할 수 없습니다", 403)
	}

	if req.Name != "" {
		existing, _ := s.repo.FindCustomStageByProjectAndName(stage.ProjectID, req.Name)
		if existing != nil && existing.ID != stage.ID {
			return nil, apperrors.New(apperrors.ErrCodeConflict, "이미 존재하는 단계 이름입니다", 409)
		}
		stage.Name = req.Name
	}
	if req.Color != "" {
		stage.Color = req.Color
	}

	if err := s.repo.UpdateCustomStage(stage); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "단계 수정 실패", 500)
	}

	return s.toCustomStageResponse(stage), nil
}

func (s *customFieldService) DeleteCustomStage(stageID, userID string) error {
	stageUUID, err := uuid.Parse(stageID)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 단계 ID", 400)
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	stage, err := s.repo.FindCustomStageByID(stageUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.New(apperrors.ErrCodeNotFound, "단계를 찾을 수 없습니다", 404)
		}
		return apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "단계 조회 실패", 500)
	}

	if err := s.checkProjectAdminPermission(userUUID, stage.ProjectID); err != nil {
		return err
	}

	if stage.IsSystemDefault {
		return apperrors.New(apperrors.ErrCodeForbidden, "시스템 기본 단계는 삭제할 수 없습니다", 403)
	}

	// Phase 5 IMPLEMENTED: Update boards using this stage to "없음"
	defaultStage, err := s.repo.FindCustomStageByProjectAndName(stage.ProjectID, "없음")
	if err != nil {
		s.logger.Error("Failed to find default stage", zap.Error(err), zap.String("project_id", stage.ProjectID.String()))
		return apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "기본 단계 조회 실패", 500)
	}

	if err := s.boardRepo.UpdateBoardsStageToDefault(stageUUID, defaultStage.ID); err != nil {
		s.logger.Error("Failed to update boards stage to default", zap.Error(err))
		return apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "칸반 단계 업데이트 실패", 500)
	}

	if err := s.repo.DeleteCustomStage(stageUUID); err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "단계 삭제 실패", 500)
	}

	return nil
}

func (s *customFieldService) UpdateCustomStageOrder(projectID, userID string, req *dto.UpdateCustomStageOrderRequest) error {
	projectUUID, err := uuid.Parse(projectID)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 프로젝트 ID", 400)
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	if err := s.checkProjectAdminPermission(userUUID, projectUUID); err != nil {
		return err
	}

	stages := make([]domain.CustomStage, 0, len(req.StageOrders))
	for _, order := range req.StageOrders {
		stageUUID, err := uuid.Parse(order.StageID)
		if err != nil {
			return apperrors.Wrap(err, apperrors.ErrCodeBadRequest, fmt.Sprintf("잘못된 단계 ID: %s", order.StageID), 400)
		}

		stage, err := s.repo.FindCustomStageByID(stageUUID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return apperrors.New(apperrors.ErrCodeNotFound, fmt.Sprintf("단계를 찾을 수 없습니다: %s", order.StageID), 404)
			}
			return apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "단계 조회 실패", 500)
		}

		if stage.ProjectID != projectUUID {
			return apperrors.New(apperrors.ErrCodeForbidden, "다른 프로젝트의 단계입니다", 403)
		}

		stage.DisplayOrder = order.DisplayOrder
		stages = append(stages, *stage)
	}

	if err := s.repo.UpdateCustomStageOrders(stages); err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "단계 순서 업데이트 실패", 500)
	}

	return nil
}

// ==================== Custom Importance ====================
// Similar implementation pattern

func (s *customFieldService) CreateCustomImportance(userID string, req *dto.CreateCustomImportanceRequest) (*dto.CustomImportanceResponse, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	projectUUID, err := uuid.Parse(req.ProjectID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 프로젝트 ID", 400)
	}

	if err := s.checkProjectAdminPermission(userUUID, projectUUID); err != nil {
		return nil, err
	}

	existing, _ := s.repo.FindCustomImportanceByProjectAndName(projectUUID, req.Name)
	if existing != nil {
		return nil, apperrors.New(apperrors.ErrCodeConflict, "이미 존재하는 중요도 이름입니다", 409)
	}

	importances, err := s.repo.FindCustomImportancesByProject(projectUUID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "중요도 조회 실패", 500)
	}

	maxOrder := 0
	for _, importance := range importances {
		if importance.DisplayOrder > maxOrder {
			maxOrder = importance.DisplayOrder
		}
	}

	importance := &domain.CustomImportance{
		ProjectID:       projectUUID,
		Name:            req.Name,
		Color:           req.Color,
		IsSystemDefault: false,
		DisplayOrder:    maxOrder + 1,
	}

	if err := s.repo.CreateCustomImportance(importance); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "중요도 생성 실패", 500)
	}

	return s.toCustomImportanceResponse(importance), nil
}

func (s *customFieldService) GetCustomImportances(projectID, userID string) ([]dto.CustomImportanceResponse, error) {
	projectUUID, err := uuid.Parse(projectID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 프로젝트 ID", 400)
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	if err := s.checkProjectMember(userUUID, projectUUID); err != nil {
		return nil, err
	}

	importances, err := s.repo.FindCustomImportancesByProject(projectUUID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "중요도 조회 실패", 500)
	}

	var responses []dto.CustomImportanceResponse
	for _, importance := range importances {
		responses = append(responses, *s.toCustomImportanceResponse(&importance))
	}

	return responses, nil
}

func (s *customFieldService) GetCustomImportance(importanceID, userID string) (*dto.CustomImportanceResponse, error) {
	importanceUUID, err := uuid.Parse(importanceID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 중요도 ID", 400)
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	importance, err := s.repo.FindCustomImportanceByID(importanceUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.ErrCodeNotFound, "중요도를 찾을 수 없습니다", 404)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "중요도 조회 실패", 500)
	}

	if err := s.checkProjectMember(userUUID, importance.ProjectID); err != nil {
		return nil, err
	}

	return s.toCustomImportanceResponse(importance), nil
}

func (s *customFieldService) UpdateCustomImportance(importanceID, userID string, req *dto.UpdateCustomImportanceRequest) (*dto.CustomImportanceResponse, error) {
	importanceUUID, err := uuid.Parse(importanceID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 중요도 ID", 400)
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	importance, err := s.repo.FindCustomImportanceByID(importanceUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.ErrCodeNotFound, "중요도를 찾을 수 없습니다", 404)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "중요도 조회 실패", 500)
	}

	if err := s.checkProjectAdminPermission(userUUID, importance.ProjectID); err != nil {
		return nil, err
	}

	if importance.IsSystemDefault {
		return nil, apperrors.New(apperrors.ErrCodeForbidden, "시스템 기본 중요도는 수정할 수 없습니다", 403)
	}

	if req.Name != "" {
		existing, _ := s.repo.FindCustomImportanceByProjectAndName(importance.ProjectID, req.Name)
		if existing != nil && existing.ID != importance.ID {
			return nil, apperrors.New(apperrors.ErrCodeConflict, "이미 존재하는 중요도 이름입니다", 409)
		}
		importance.Name = req.Name
	}
	if req.Color != "" {
		importance.Color = req.Color
	}

	if err := s.repo.UpdateCustomImportance(importance); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "중요도 수정 실패", 500)
	}

	return s.toCustomImportanceResponse(importance), nil
}

func (s *customFieldService) DeleteCustomImportance(importanceID, userID string) error {
	importanceUUID, err := uuid.Parse(importanceID)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 중요도 ID", 400)
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	importance, err := s.repo.FindCustomImportanceByID(importanceUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.New(apperrors.ErrCodeNotFound, "중요도를 찾을 수 없습니다", 404)
		}
		return apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "중요도 조회 실패", 500)
	}

	if err := s.checkProjectAdminPermission(userUUID, importance.ProjectID); err != nil {
		return err
	}

	if importance.IsSystemDefault {
		return apperrors.New(apperrors.ErrCodeForbidden, "시스템 기본 중요도는 삭제할 수 없습니다", 403)
	}

	// Phase 5 IMPLEMENTED: Update boards using this importance to "없음"
	defaultImportance, err := s.repo.FindCustomImportanceByProjectAndName(importance.ProjectID, "없음")
	if err != nil {
		s.logger.Error("Failed to find default importance", zap.Error(err), zap.String("project_id", importance.ProjectID.String()))
		return apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "기본 중요도 조회 실패", 500)
	}

	if err := s.boardRepo.UpdateBoardsImportanceToDefault(importanceUUID, defaultImportance.ID); err != nil {
		s.logger.Error("Failed to update boards importance to default", zap.Error(err))
		return apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "칸반 중요도 업데이트 실패", 500)
	}

	if err := s.repo.DeleteCustomImportance(importanceUUID); err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "중요도 삭제 실패", 500)
	}

	return nil
}

func (s *customFieldService) UpdateCustomImportanceOrder(projectID, userID string, req *dto.UpdateCustomImportanceOrderRequest) error {
	projectUUID, err := uuid.Parse(projectID)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 프로젝트 ID", 400)
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	if err := s.checkProjectAdminPermission(userUUID, projectUUID); err != nil {
		return err
	}

	importances := make([]domain.CustomImportance, 0, len(req.ImportanceOrders))
	for _, order := range req.ImportanceOrders {
		importanceUUID, err := uuid.Parse(order.ImportanceID)
		if err != nil {
			return apperrors.Wrap(err, apperrors.ErrCodeBadRequest, fmt.Sprintf("잘못된 중요도 ID: %s", order.ImportanceID), 400)
		}

		importance, err := s.repo.FindCustomImportanceByID(importanceUUID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return apperrors.New(apperrors.ErrCodeNotFound, fmt.Sprintf("중요도를 찾을 수 없습니다: %s", order.ImportanceID), 404)
			}
			return apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "중요도 조회 실패", 500)
		}

		if importance.ProjectID != projectUUID {
			return apperrors.New(apperrors.ErrCodeForbidden, "다른 프로젝트의 중요도입니다", 403)
		}

		importance.DisplayOrder = order.DisplayOrder
		importances = append(importances, *importance)
	}

	if err := s.repo.UpdateCustomImportanceOrders(importances); err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "중요도 순서 업데이트 실패", 500)
	}

	return nil
}

// ==================== Helper Methods ====================

func (s *customFieldService) checkProjectMember(userID, projectID uuid.UUID) error {
	_, err := s.projectRepo.FindMemberByUserAndProject(userID, projectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.New(apperrors.ErrCodeForbidden, "프로젝트 멤버가 아닙니다", 403)
		}
		return apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "멤버 확인 실패", 500)
	}
	return nil
}

func (s *customFieldService) checkProjectAdminPermission(userID, projectID uuid.UUID) error {
	member, err := s.projectRepo.FindMemberByUserAndProject(userID, projectID)
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

func (s *customFieldService) toCustomRoleResponse(role *domain.CustomRole) *dto.CustomRoleResponse {
	return &dto.CustomRoleResponse{
		ID:              role.ID.String(),
		ProjectID:       role.ProjectID.String(),
		Name:            role.Name,
		Color:           role.Color,
		IsSystemDefault: role.IsSystemDefault,
		DisplayOrder:    role.DisplayOrder,
		CreatedAt:       role.CreatedAt,
		UpdatedAt:       role.UpdatedAt,
	}
}

func (s *customFieldService) toCustomStageResponse(stage *domain.CustomStage) *dto.CustomStageResponse {
	return &dto.CustomStageResponse{
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

func (s *customFieldService) toCustomImportanceResponse(importance *domain.CustomImportance) *dto.CustomImportanceResponse {
	return &dto.CustomImportanceResponse{
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

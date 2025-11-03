package service

import (
	"board-service/internal/apperrors"
	"board-service/internal/cache"
	"board-service/internal/domain"
	"board-service/internal/dto"
	"board-service/internal/repository"
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type UserOrderService interface {
	// Board Views
	GetRoleBasedBoardView(ctx context.Context, userID, projectID string) (*dto.RoleBasedBoardView, error)
	GetStageBasedBoardView(ctx context.Context, userID, projectID string) (*dto.StageBasedBoardView, error)

	// Update Column Orders (drag-and-drop)
	UpdateRoleColumnOrder(ctx context.Context, userID, projectID string, req *dto.UpdateOrderRequest) error
	UpdateStageColumnOrder(ctx context.Context, userID, projectID string, req *dto.UpdateOrderRequest) error

	// Update Kanban Orders (drag-and-drop within column)
	UpdateKanbanOrderInRole(ctx context.Context, userID, projectID, roleID string, req *dto.UpdateOrderRequest) error
	UpdateKanbanOrderInStage(ctx context.Context, userID, projectID, stageID string, req *dto.UpdateOrderRequest) error
}

type userOrderService struct {
	orderRepo       repository.UserOrderRepository
	projectRepo     repository.ProjectRepository
	customFieldRepo repository.CustomFieldRepository
	kanbanRepo      repository.KanbanRepository
	cache           *cache.UserOrderCache
	logger          *zap.Logger
}

func NewUserOrderService(
	orderRepo repository.UserOrderRepository,
	projectRepo repository.ProjectRepository,
	customFieldRepo repository.CustomFieldRepository,
	kanbanRepo repository.KanbanRepository,
	cache *cache.UserOrderCache,
	logger *zap.Logger,
) UserOrderService {
	return &userOrderService{
		orderRepo:       orderRepo,
		projectRepo:     projectRepo,
		customFieldRepo: customFieldRepo,
		kanbanRepo:      kanbanRepo,
		cache:           cache,
		logger:          logger,
	}
}

// ==================== Role-Based Board View ====================

func (s *userOrderService) GetRoleBasedBoardView(ctx context.Context, userID, projectID string) (*dto.RoleBasedBoardView, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	projectUUID, err := uuid.Parse(projectID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 프로젝트 ID", 400)
	}

	// Verify user is project member
	if err := s.verifyProjectMember(ctx, userUUID, projectUUID); err != nil {
		return nil, err
	}

	// Get role column orders
	roleOrders, err := s.getRoleColumnOrdersWithCache(ctx, userUUID, projectUUID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "역할 컬럼 순서 조회 실패", 500)
	}

	// Build board view
	view := &dto.RoleBasedBoardView{
		ProjectID: projectID,
		ViewType:  "role",
		Columns:   make([]dto.RoleColumnView, len(roleOrders)),
	}

	for i, roleOrder := range roleOrders {
		// Get role details
		role, err := s.customFieldRepo.FindCustomRoleByID(roleOrder.CustomRoleID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				continue // Skip deleted roles
			}
			return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "역할 조회 실패", 500)
		}

		// Get kanbans for this role
		kanbans, err := s.getKanbansForRole(ctx, userUUID, projectUUID, roleOrder.CustomRoleID)
		if err != nil {
			return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "칸반 조회 실패", 500)
		}

		view.Columns[i] = dto.RoleColumnView{
			CustomRoleID: role.ID.String(),
			RoleName:     role.Name,
			RoleColor:    role.Color,
			DisplayOrder: roleOrder.DisplayOrder,
			Kanbans:      kanbans,
		}
	}

	return view, nil
}

// ==================== Stage-Based Board View ====================

func (s *userOrderService) GetStageBasedBoardView(ctx context.Context, userID, projectID string) (*dto.StageBasedBoardView, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	projectUUID, err := uuid.Parse(projectID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 프로젝트 ID", 400)
	}

	// Verify user is project member
	if err := s.verifyProjectMember(ctx, userUUID, projectUUID); err != nil {
		return nil, err
	}

	// Get stage column orders
	stageOrders, err := s.getStageColumnOrdersWithCache(ctx, userUUID, projectUUID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "진행단계 컬럼 순서 조회 실패", 500)
	}

	// Build board view
	view := &dto.StageBasedBoardView{
		ProjectID: projectID,
		ViewType:  "stage",
		Columns:   make([]dto.StageColumnView, len(stageOrders)),
	}

	for i, stageOrder := range stageOrders {
		// Get stage details
		stage, err := s.customFieldRepo.FindCustomStageByID(stageOrder.CustomStageID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				continue // Skip deleted stages
			}
			return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "진행단계 조회 실패", 500)
		}

		// Get kanbans for this stage
		kanbans, err := s.getKanbansForStage(ctx, userUUID, projectUUID, stageOrder.CustomStageID)
		if err != nil {
			return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "칸반 조회 실패", 500)
		}

		view.Columns[i] = dto.StageColumnView{
			CustomStageID: stage.ID.String(),
			StageName:     stage.Name,
			StageColor:    stage.Color,
			DisplayOrder:  stageOrder.DisplayOrder,
			Kanbans:       kanbans,
		}
	}

	return view, nil
}

// ==================== Update Column Orders ====================

func (s *userOrderService) UpdateRoleColumnOrder(ctx context.Context, userID, projectID string, req *dto.UpdateOrderRequest) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	projectUUID, err := uuid.Parse(projectID)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 프로젝트 ID", 400)
	}

	// Verify user is project member
	if err := s.verifyProjectMember(ctx, userUUID, projectUUID); err != nil {
		return err
	}

	// Build update orders with auto-calculated displayOrder
	orders := make([]domain.UserRoleColumnOrder, len(req.ItemIds))
	for i, itemID := range req.ItemIds {
		roleID, err := uuid.Parse(itemID)
		if err != nil {
			return apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 역할 ID", 400)
		}

		orders[i] = domain.UserRoleColumnOrder{
			UserID:       userUUID,
			ProjectID:    projectUUID,
			CustomRoleID: roleID,
			DisplayOrder: i + 1, // Auto-increment order (1, 2, 3, ...)
		}
	}

	// Update in DB
	if err := s.orderRepo.UpsertRoleColumnOrder(ctx, orders); err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "역할 컬럼 순서 업데이트 실패", 500)
	}

	// Invalidate cache
	if err := s.cache.DeleteRoleColumnOrder(ctx, userID, projectID); err != nil && err != redis.Nil {
		s.logger.Warn("Failed to invalidate role column order cache", zap.Error(err))
	}

	return nil
}

func (s *userOrderService) UpdateStageColumnOrder(ctx context.Context, userID, projectID string, req *dto.UpdateOrderRequest) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	projectUUID, err := uuid.Parse(projectID)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 프로젝트 ID", 400)
	}

	// Verify user is project member
	if err := s.verifyProjectMember(ctx, userUUID, projectUUID); err != nil {
		return err
	}

	// Build update orders with auto-calculated displayOrder
	orders := make([]domain.UserStageColumnOrder, len(req.ItemIds))
	for i, itemID := range req.ItemIds {
		stageID, err := uuid.Parse(itemID)
		if err != nil {
			return apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 진행단계 ID", 400)
		}

		orders[i] = domain.UserStageColumnOrder{
			UserID:        userUUID,
			ProjectID:     projectUUID,
			CustomStageID: stageID,
			DisplayOrder:  i + 1, // Auto-increment order (1, 2, 3, ...)
		}
	}

	// Update in DB
	if err := s.orderRepo.UpsertStageColumnOrder(ctx, orders); err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "진행단계 컬럼 순서 업데이트 실패", 500)
	}

	// Invalidate cache
	if err := s.cache.DeleteStageColumnOrder(ctx, userID, projectID); err != nil && err != redis.Nil {
		s.logger.Warn("Failed to invalidate stage column order cache", zap.Error(err))
	}

	return nil
}

// ==================== Update Kanban Orders ====================

func (s *userOrderService) UpdateKanbanOrderInRole(ctx context.Context, userID, projectID, roleID string, req *dto.UpdateOrderRequest) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	projectUUID, err := uuid.Parse(projectID)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 프로젝트 ID", 400)
	}

	roleUUID, err := uuid.Parse(roleID)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 역할 ID", 400)
	}

	// Verify user is project member
	if err := s.verifyProjectMember(ctx, userUUID, projectUUID); err != nil {
		return err
	}

	// Build update orders with auto-calculated displayOrder
	orders := make([]domain.UserKanbanOrderInRole, len(req.ItemIds))
	for i, itemID := range req.ItemIds {
		kanbanID, err := uuid.Parse(itemID)
		if err != nil {
			return apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 칸반 ID", 400)
		}

		orders[i] = domain.UserKanbanOrderInRole{
			UserID:       userUUID,
			ProjectID:    projectUUID,
			CustomRoleID: roleUUID,
			KanbanID:     kanbanID,
			DisplayOrder: i + 1, // Auto-increment order (1, 2, 3, ...)
		}
	}

	// Update in DB
	if err := s.orderRepo.UpsertKanbanOrderInRole(ctx, orders); err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "역할별 칸반 순서 업데이트 실패", 500)
	}

	// Invalidate cache
	if err := s.cache.DeleteRoleKanbanOrder(ctx, userID, projectID, roleID); err != nil && err != redis.Nil {
		s.logger.Warn("Failed to invalidate role kanban order cache", zap.Error(err))
	}

	return nil
}

func (s *userOrderService) UpdateKanbanOrderInStage(ctx context.Context, userID, projectID, stageID string, req *dto.UpdateOrderRequest) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 사용자 ID", 400)
	}

	projectUUID, err := uuid.Parse(projectID)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 프로젝트 ID", 400)
	}

	stageUUID, err := uuid.Parse(stageID)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 진행단계 ID", 400)
	}

	// Verify user is project member
	if err := s.verifyProjectMember(ctx, userUUID, projectUUID); err != nil {
		return err
	}

	// Build update orders with auto-calculated displayOrder
	orders := make([]domain.UserKanbanOrderInStage, len(req.ItemIds))
	for i, itemID := range req.ItemIds {
		kanbanID, err := uuid.Parse(itemID)
		if err != nil {
			return apperrors.Wrap(err, apperrors.ErrCodeBadRequest, "잘못된 칸반 ID", 400)
		}

		orders[i] = domain.UserKanbanOrderInStage{
			UserID:        userUUID,
			ProjectID:     projectUUID,
			CustomStageID: stageUUID,
			KanbanID:      kanbanID,
			DisplayOrder:  i + 1, // Auto-increment order (1, 2, 3, ...)
		}
	}

	// Update in DB
	if err := s.orderRepo.UpsertKanbanOrderInStage(ctx, orders); err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "진행단계별 칸반 순서 업데이트 실패", 500)
	}

	// Invalidate cache
	if err := s.cache.DeleteStageKanbanOrder(ctx, userID, projectID, stageID); err != nil && err != redis.Nil {
		s.logger.Warn("Failed to invalidate stage kanban order cache", zap.Error(err))
	}

	return nil
}

// ==================== Helper Functions ====================

func (s *userOrderService) verifyProjectMember(ctx context.Context, userID, projectID uuid.UUID) error {
	_, err := s.projectRepo.FindMemberByUserAndProject(userID, projectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.New(apperrors.ErrCodeForbidden, "프로젝트 멤버가 아닙니다", 403)
		}
		return apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "멤버 확인 실패", 500)
	}
	return nil
}

func (s *userOrderService) getRoleColumnOrdersWithCache(ctx context.Context, userID, projectID uuid.UUID) ([]domain.UserRoleColumnOrder, error) {
	// Try cache first
	cachedData, err := s.cache.GetRoleColumnOrder(ctx, userID.String(), projectID.String())
	if err == nil {
		var orders []domain.UserRoleColumnOrder
		if err := json.Unmarshal(cachedData, &orders); err == nil {
			return orders, nil
		}
	}

	// Cache miss or error - fetch from DB
	orders, err := s.orderRepo.GetRoleColumnOrder(ctx, userID, projectID)
	if err != nil {
		return nil, err
	}

	// If no orders exist, initialize them
	if len(orders) == 0 {
		if err := s.orderRepo.InitializeUserOrders(ctx, userID, projectID); err != nil {
			return nil, err
		}
		orders, err = s.orderRepo.GetRoleColumnOrder(ctx, userID, projectID)
		if err != nil {
			return nil, err
		}
	}

	// Cache the result
	if err := s.cache.SetRoleColumnOrder(ctx, userID.String(), projectID.String(), orders); err != nil {
		s.logger.Warn("Failed to cache role column orders", zap.Error(err))
	}

	return orders, nil
}

func (s *userOrderService) getStageColumnOrdersWithCache(ctx context.Context, userID, projectID uuid.UUID) ([]domain.UserStageColumnOrder, error) {
	// Try cache first
	cachedData, err := s.cache.GetStageColumnOrder(ctx, userID.String(), projectID.String())
	if err == nil {
		var orders []domain.UserStageColumnOrder
		if err := json.Unmarshal(cachedData, &orders); err == nil {
			return orders, nil
		}
	}

	// Cache miss or error - fetch from DB
	orders, err := s.orderRepo.GetStageColumnOrder(ctx, userID, projectID)
	if err != nil {
		return nil, err
	}

	// If no orders exist, initialize them
	if len(orders) == 0 {
		if err := s.orderRepo.InitializeUserOrders(ctx, userID, projectID); err != nil {
			return nil, err
		}
		orders, err = s.orderRepo.GetStageColumnOrder(ctx, userID, projectID)
		if err != nil {
			return nil, err
		}
	}

	// Cache the result
	if err := s.cache.SetStageColumnOrder(ctx, userID.String(), projectID.String(), orders); err != nil {
		s.logger.Warn("Failed to cache stage column orders", zap.Error(err))
	}

	return orders, nil
}

func (s *userOrderService) getKanbansForRole(ctx context.Context, userID, projectID, roleID uuid.UUID) ([]dto.KanbanOrderResponse, error) {
	// Try cache first
	cachedData, err := s.cache.GetRoleKanbanOrder(ctx, userID.String(), projectID.String(), roleID.String())
	if err == nil {
		var kanbans []dto.KanbanOrderResponse
		if err := json.Unmarshal(cachedData, &kanbans); err == nil {
			return kanbans, nil
		}
	}

	// Get kanbans from DB (using kanban_roles join)
	kanbans, err := s.kanbanRepo.FindKanbansByRole(roleID)
	if err != nil {
		return nil, err
	}

	// Get user-specific order
	userOrders, err := s.orderRepo.GetKanbanOrderInRole(ctx, userID, projectID, roleID)
	if err != nil {
		return nil, err
	}

	// Create order map
	orderMap := make(map[uuid.UUID]int)
	for _, order := range userOrders {
		orderMap[order.KanbanID] = order.DisplayOrder
	}

	// Build response with order
	result := make([]dto.KanbanOrderResponse, 0, len(kanbans))
	for _, kanban := range kanbans {
		order, exists := orderMap[kanban.ID]
		if !exists {
			order = 999999 // Default order for new kanbans
		}

		result = append(result, dto.KanbanOrderResponse{
			KanbanID:     kanban.ID.String(),
			Title:        kanban.Title,
			DisplayOrder: order,
		})
	}

	// Cache the result
	if err := s.cache.SetRoleKanbanOrder(ctx, userID.String(), projectID.String(), roleID.String(), result); err != nil {
		s.logger.Warn("Failed to cache role kanban orders", zap.Error(err))
	}

	return result, nil
}

func (s *userOrderService) getKanbansForStage(ctx context.Context, userID, projectID, stageID uuid.UUID) ([]dto.KanbanOrderResponse, error) {
	// Try cache first
	cachedData, err := s.cache.GetStageKanbanOrder(ctx, userID.String(), projectID.String(), stageID.String())
	if err == nil {
		var kanbans []dto.KanbanOrderResponse
		if err := json.Unmarshal(cachedData, &kanbans); err == nil {
			return kanbans, nil
		}
	}

	// Get kanbans from DB (filter by stage)
	kanbans, _, err := s.kanbanRepo.FindByProject(projectID, repository.KanbanFilters{StageID: stageID}, 1, 1000)
	if err != nil {
		return nil, err
	}

	// Get user-specific order
	userOrders, err := s.orderRepo.GetKanbanOrderInStage(ctx, userID, projectID, stageID)
	if err != nil {
		return nil, err
	}

	// Create order map
	orderMap := make(map[uuid.UUID]int)
	for _, order := range userOrders {
		orderMap[order.KanbanID] = order.DisplayOrder
	}

	// Build response with order
	result := make([]dto.KanbanOrderResponse, 0, len(kanbans))
	for _, kanban := range kanbans {
		order, exists := orderMap[kanban.ID]
		if !exists {
			order = 999999 // Default order for new kanbans
		}

		result = append(result, dto.KanbanOrderResponse{
			KanbanID:     kanban.ID.String(),
			Title:        kanban.Title,
			DisplayOrder: order,
		})
	}

	// Cache the result
	if err := s.cache.SetStageKanbanOrder(ctx, userID.String(), projectID.String(), stageID.String(), result); err != nil {
		s.logger.Warn("Failed to cache stage kanban orders", zap.Error(err))
	}

	return result, nil
}

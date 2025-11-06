package repository

import (
	"board-service/internal/domain"
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type UserOrderRepository interface {
	// Role Column Order
	GetRoleColumnOrder(ctx context.Context, userID, projectID uuid.UUID) ([]domain.UserRoleColumnOrder, error)
	UpsertRoleColumnOrder(ctx context.Context, orders []domain.UserRoleColumnOrder) error

	// Stage Column Order
	GetStageColumnOrder(ctx context.Context, userID, projectID uuid.UUID) ([]domain.UserStageColumnOrder, error)
	UpsertStageColumnOrder(ctx context.Context, orders []domain.UserStageColumnOrder) error

	// Role Board Order
	GetBoardOrderInRole(ctx context.Context, userID, projectID, roleID uuid.UUID) ([]domain.UserBoardOrderInRole, error)
	UpsertBoardOrderInRole(ctx context.Context, orders []domain.UserBoardOrderInRole) error

	// Stage Board Order
	GetBoardOrderInStage(ctx context.Context, userID, projectID, stageID uuid.UUID) ([]domain.UserBoardOrderInStage, error)
	UpsertBoardOrderInStage(ctx context.Context, orders []domain.UserBoardOrderInStage) error

	// Initialize user orders when joining a project
	InitializeUserOrders(ctx context.Context, userID, projectID uuid.UUID) error

	// Get all custom field IDs for a project (helper for initialization)
	GetProjectCustomRoleIDs(ctx context.Context, projectID uuid.UUID) ([]uuid.UUID, error)
	GetProjectCustomStageIDs(ctx context.Context, projectID uuid.UUID) ([]uuid.UUID, error)
}

type userOrderRepository struct {
	db *gorm.DB
}

func NewUserOrderRepository(db *gorm.DB) UserOrderRepository {
	return &userOrderRepository{db: db}
}

// ==================== Role Column Order ====================

func (r *userOrderRepository) GetRoleColumnOrder(ctx context.Context, userID, projectID uuid.UUID) ([]domain.UserRoleColumnOrder, error) {
	var orders []domain.UserRoleColumnOrder
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND project_id = ?", userID, projectID).
		Order("display_order ASC").
		Find(&orders).Error
	return orders, err
}

func (r *userOrderRepository) UpsertRoleColumnOrder(ctx context.Context, orders []domain.UserRoleColumnOrder) error {
	if len(orders) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, order := range orders {
			if err := tx.Clauses(clause.OnConflict{
				Columns: []clause.Column{
					{Name: "user_id"},
					{Name: "project_id"},
					{Name: "custom_role_id"},
				},
				DoUpdates: clause.AssignmentColumns([]string{"display_order", "updated_at"}),
			}).Create(&order).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// ==================== Stage Column Order ====================

func (r *userOrderRepository) GetStageColumnOrder(ctx context.Context, userID, projectID uuid.UUID) ([]domain.UserStageColumnOrder, error) {
	var orders []domain.UserStageColumnOrder
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND project_id = ?", userID, projectID).
		Order("display_order ASC").
		Find(&orders).Error
	return orders, err
}

func (r *userOrderRepository) UpsertStageColumnOrder(ctx context.Context, orders []domain.UserStageColumnOrder) error {
	if len(orders) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, order := range orders {
			if err := tx.Clauses(clause.OnConflict{
				Columns: []clause.Column{
					{Name: "user_id"},
					{Name: "project_id"},
					{Name: "custom_stage_id"},
				},
				DoUpdates: clause.AssignmentColumns([]string{"display_order", "updated_at"}),
			}).Create(&order).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// ==================== Role Board Order ====================

func (r *userOrderRepository) GetBoardOrderInRole(ctx context.Context, userID, projectID, roleID uuid.UUID) ([]domain.UserBoardOrderInRole, error) {
	var orders []domain.UserBoardOrderInRole
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND project_id = ? AND custom_role_id = ?", userID, projectID, roleID).
		Order("display_order ASC").
		Find(&orders).Error
	return orders, err
}

func (r *userOrderRepository) UpsertBoardOrderInRole(ctx context.Context, orders []domain.UserBoardOrderInRole) error {
	if len(orders) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, order := range orders {
			if err := tx.Clauses(clause.OnConflict{
				Columns: []clause.Column{
					{Name: "user_id"},
					{Name: "project_id"},
					{Name: "custom_role_id"},
					{Name: "board_id"},
				},
				DoUpdates: clause.AssignmentColumns([]string{"display_order", "updated_at"}),
			}).Create(&order).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// ==================== Stage Board Order ====================

func (r *userOrderRepository) GetBoardOrderInStage(ctx context.Context, userID, projectID, stageID uuid.UUID) ([]domain.UserBoardOrderInStage, error) {
	var orders []domain.UserBoardOrderInStage
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND project_id = ? AND custom_stage_id = ?", userID, projectID, stageID).
		Order("display_order ASC").
		Find(&orders).Error
	return orders, err
}

func (r *userOrderRepository) UpsertBoardOrderInStage(ctx context.Context, orders []domain.UserBoardOrderInStage) error {
	if len(orders) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, order := range orders {
			if err := tx.Clauses(clause.OnConflict{
				Columns: []clause.Column{
					{Name: "user_id"},
					{Name: "project_id"},
					{Name: "custom_stage_id"},
					{Name: "board_id"},
				},
				DoUpdates: clause.AssignmentColumns([]string{"display_order", "updated_at"}),
			}).Create(&order).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// ==================== Initialization ====================

func (r *userOrderRepository) GetProjectCustomRoleIDs(ctx context.Context, projectID uuid.UUID) ([]uuid.UUID, error) {
	var roleIDs []uuid.UUID
	err := r.db.WithContext(ctx).
		Model(&domain.CustomRole{}).
		Where("project_id = ? AND is_deleted = false", projectID).
		Order("created_at ASC").
		Pluck("id", &roleIDs).Error
	return roleIDs, err
}

func (r *userOrderRepository) GetProjectCustomStageIDs(ctx context.Context, projectID uuid.UUID) ([]uuid.UUID, error) {
	var stageIDs []uuid.UUID
	err := r.db.WithContext(ctx).
		Model(&domain.CustomStage{}).
		Where("project_id = ? AND is_deleted = false", projectID).
		Order("created_at ASC").
		Pluck("id", &stageIDs).Error
	return stageIDs, err
}

func (r *userOrderRepository) InitializeUserOrders(ctx context.Context, userID, projectID uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Initialize role column orders
		roleIDs, err := r.GetProjectCustomRoleIDs(ctx, projectID)
		if err != nil {
			return err
		}

		roleOrders := make([]domain.UserRoleColumnOrder, len(roleIDs))
		for i, roleID := range roleIDs {
			roleOrders[i] = domain.UserRoleColumnOrder{
				UserID:       userID,
				ProjectID:    projectID,
				CustomRoleID: roleID,
				DisplayOrder: i,
			}
		}

		if len(roleOrders) > 0 {
			if err := r.UpsertRoleColumnOrder(ctx, roleOrders); err != nil {
				return err
			}
		}

		// Initialize stage column orders
		stageIDs, err := r.GetProjectCustomStageIDs(ctx, projectID)
		if err != nil {
			return err
		}

		stageOrders := make([]domain.UserStageColumnOrder, len(stageIDs))
		for i, stageID := range stageIDs {
			stageOrders[i] = domain.UserStageColumnOrder{
				UserID:        userID,
				ProjectID:     projectID,
				CustomStageID: stageID,
				DisplayOrder:  i,
			}
		}

		if len(stageOrders) > 0 {
			if err := r.UpsertStageColumnOrder(ctx, stageOrders); err != nil {
				return err
			}
		}

		return nil
	})
}

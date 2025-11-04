package repository

import (
	"board-service/internal/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type KanbanRepository interface {
	// CRUD
	Create(kanban *domain.Kanban) error
	FindByID(id uuid.UUID) (*domain.Kanban, error)
	FindByProject(projectID uuid.UUID, filters KanbanFilters, page, limit int) ([]domain.Kanban, int64, error)
	Update(kanban *domain.Kanban) error
	Delete(id uuid.UUID) error

	// Kanban Roles (many-to-many)
	CreateKanbanRoles(kanbanID uuid.UUID, roleIDs []uuid.UUID) error
	DeleteKanbanRolesByKanban(kanbanID uuid.UUID) error
	FindRolesByKanban(kanbanID uuid.UUID) ([]domain.KanbanRole, error)
	FindKanbansByRole(roleID uuid.UUID) ([]domain.Kanban, error)

	// Phase 4 TODO implementation: Update kanbans when custom fields are deleted
	UpdateKanbansRoleToDefault(oldRoleID, defaultRoleID uuid.UUID) error
	UpdateKanbansStageToDefault(oldStageID, defaultStageID uuid.UUID) error
	UpdateKanbansImportanceToDefault(oldImportanceID, defaultImportanceID uuid.UUID) error
}

type KanbanFilters struct {
	StageID      uuid.UUID
	RoleID       uuid.UUID
	ImportanceID uuid.UUID
	AssigneeID   uuid.UUID
	AuthorID     uuid.UUID
}

type kanbanRepository struct {
	db *gorm.DB
}

func NewKanbanRepository(db *gorm.DB) KanbanRepository {
	return &kanbanRepository{db: db}
}

// ==================== CRUD ====================

func (r *kanbanRepository) Create(kanban *domain.Kanban) error {
	return r.db.Create(kanban).Error
}

func (r *kanbanRepository) FindByID(id uuid.UUID) (*domain.Kanban, error) {
	var kanban domain.Kanban
	if err := r.db.Where("id = ? AND deleted_at IS NULL", id).First(&kanban).Error; err != nil {
		return nil, err
	}
	return &kanban, nil
}

func (r *kanbanRepository) FindByProject(projectID uuid.UUID, filters KanbanFilters, page, limit int) ([]domain.Kanban, int64, error) {
	var kanbans []domain.Kanban
	var total int64

	query := r.db.Model(&domain.Kanban{}).Where("project_id = ? AND deleted_at IS NULL", projectID)

	// Apply filters
	if filters.StageID != uuid.Nil {
		query = query.Where("stage_id = ?", filters.StageID)
	}
	if filters.ImportanceID != uuid.Nil {
		query = query.Where("importance_id = ?", filters.ImportanceID)
	}
	if filters.AssigneeID != uuid.Nil {
		query = query.Where("assignee_id = ?", filters.AssigneeID)
	}
	if filters.AuthorID != uuid.Nil {
		query = query.Where("author_id = ?", filters.AuthorID)
	}

	// Role filter (JOIN)
	if filters.RoleID != uuid.Nil {
		query = query.Joins("INNER JOIN kanban_roles ON kanbans.id = kanban_roles.kanban_id").
			Where("kanban_roles.custom_role_id = ?", filters.RoleID)
	}

	// Total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Pagination
	offset := (page - 1) * limit
	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&kanbans).Error; err != nil {
		return nil, 0, err
	}

	return kanbans, total, nil
}

func (r *kanbanRepository) Update(kanban *domain.Kanban) error {
	return r.db.Save(kanban).Error
}

func (r *kanbanRepository) Delete(id uuid.UUID) error {
	// Soft delete
	return r.db.Model(&domain.Kanban{}).Where("id = ?", id).Update("deleted_at", gorm.Expr("CURRENT_TIMESTAMP")).Error
}

// ==================== Kanban Roles (Many-to-Many) ====================

func (r *kanbanRepository) CreateKanbanRoles(kanbanID uuid.UUID, roleIDs []uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, roleID := range roleIDs {
			kanbanRole := &domain.KanbanRole{
				KanbanID:     kanbanID,
				CustomRoleID: roleID,
			}
			if err := tx.Create(kanbanRole).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *kanbanRepository) DeleteKanbanRolesByKanban(kanbanID uuid.UUID) error {
	return r.db.Where("kanban_id = ?", kanbanID).Delete(&domain.KanbanRole{}).Error
}

func (r *kanbanRepository) FindRolesByKanban(kanbanID uuid.UUID) ([]domain.KanbanRole, error) {
	var roles []domain.KanbanRole
	if err := r.db.Where("kanban_id = ?", kanbanID).Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *kanbanRepository) FindKanbansByRole(roleID uuid.UUID) ([]domain.Kanban, error) {
	var kanbans []domain.Kanban
	if err := r.db.Joins("INNER JOIN kanban_roles ON kanbans.id = kanban_roles.kanban_id").
		Where("kanban_roles.custom_role_id = ? AND kanbans.deleted_at IS NULL", roleID).
		Find(&kanbans).Error; err != nil {
		return nil, err
	}
	return kanbans, nil
}

// ==================== Phase 4 TODO Implementation ====================

func (r *kanbanRepository) UpdateKanbansRoleToDefault(oldRoleID, defaultRoleID uuid.UUID) error {
	// Update all kanban_roles entries using oldRoleID to defaultRoleID
	return r.db.Model(&domain.KanbanRole{}).
		Where("custom_role_id = ?", oldRoleID).
		Update("custom_role_id", defaultRoleID).Error
}

func (r *kanbanRepository) UpdateKanbansStageToDefault(oldStageID, defaultStageID uuid.UUID) error {
	// Update all kanbans using oldStageID to defaultStageID
	return r.db.Model(&domain.Kanban{}).
		Where("stage_id = ? AND deleted_at IS NULL", oldStageID).
		Update("stage_id", defaultStageID).Error
}

func (r *kanbanRepository) UpdateKanbansImportanceToDefault(oldImportanceID, defaultImportanceID uuid.UUID) error {
	// Update all kanbans using oldImportanceID to defaultImportanceID
	return r.db.Model(&domain.Kanban{}).
		Where("importance_id = ? AND deleted_at IS NULL", oldImportanceID).
		Update("importance_id", defaultImportanceID).Error
}

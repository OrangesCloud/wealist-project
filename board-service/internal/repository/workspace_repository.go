package repository

import (
	"board-service/internal/domain"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type WorkspaceRepository interface {
	// Workspace CRUD
	Create(workspace *domain.Workspace) error
	FindByID(id uuid.UUID) (*domain.Workspace, error)
	FindByOwnerID(ownerID uuid.UUID) ([]domain.Workspace, error)
	Update(workspace *domain.Workspace) error
	Delete(id uuid.UUID) error
	Search(query string, page, limit int) ([]domain.Workspace, int64, error)

	// Join Request
	CreateJoinRequest(req *domain.WorkspaceJoinRequest) error
	FindJoinRequestByID(id uuid.UUID) (*domain.WorkspaceJoinRequest, error)
	FindJoinRequestsByWorkspace(workspaceID uuid.UUID, status string) ([]domain.WorkspaceJoinRequest, error)
	FindJoinRequestByUserAndWorkspace(userID, workspaceID uuid.UUID) (*domain.WorkspaceJoinRequest, error)
	UpdateJoinRequest(req *domain.WorkspaceJoinRequest) error

	// Member
	CreateMember(member *domain.WorkspaceMember) error
	FindMemberByID(id uuid.UUID) (*domain.WorkspaceMember, error)
	FindMembersByWorkspace(workspaceID uuid.UUID) ([]domain.WorkspaceMember, error)
	FindMemberByUserAndWorkspace(userID, workspaceID uuid.UUID) (*domain.WorkspaceMember, error)
	FindMembersByUser(userID uuid.UUID) ([]domain.WorkspaceMember, error)
	UpdateMember(member *domain.WorkspaceMember) error
	DeleteMember(id uuid.UUID) error
	ClearDefaultWorkspace(userID uuid.UUID) error
	SetDefaultWorkspace(userID, workspaceID uuid.UUID) error
}

type workspaceRepository struct {
	db *gorm.DB
}

func NewWorkspaceRepository(db *gorm.DB) WorkspaceRepository {
	return &workspaceRepository{db: db}
}

// Workspace CRUD

func (r *workspaceRepository) Create(workspace *domain.Workspace) error {
	return r.db.Create(workspace).Error
}

func (r *workspaceRepository) FindByID(id uuid.UUID) (*domain.Workspace, error) {
	var workspace domain.Workspace
	if err := r.db.Where("id = ? AND is_deleted = ?", id, false).First(&workspace).Error; err != nil {
		return nil, err
	}
	return &workspace, nil
}

func (r *workspaceRepository) FindByOwnerID(ownerID uuid.UUID) ([]domain.Workspace, error) {
	var workspaces []domain.Workspace
	if err := r.db.Where("created_by = ? AND is_deleted = ?", ownerID, false).
		Order("created_at DESC").
		Find(&workspaces).Error; err != nil {
		return nil, err
	}
	return workspaces, nil
}

func (r *workspaceRepository) Update(workspace *domain.Workspace) error {
	return r.db.Save(workspace).Error
}

func (r *workspaceRepository) Delete(id uuid.UUID) error {
	// Soft delete
	return r.db.Model(&domain.Workspace{}).
		Where("id = ?", id).
		Update("is_deleted", true).Error
}

func (r *workspaceRepository) Search(query string, page, limit int) ([]domain.Workspace, int64, error) {
	var workspaces []domain.Workspace
	var total int64

	// Default pagination
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	// Build query
	db := r.db.Model(&domain.Workspace{}).Where("is_deleted = ?", false)

	if query != "" {
		searchPattern := fmt.Sprintf("%%%s%%", query)
		db = db.Where("name ILIKE ? OR description ILIKE ?", searchPattern, searchPattern)
	}

	// Count total
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	if err := db.Offset(offset).Limit(limit).
		Order("created_at DESC").
		Find(&workspaces).Error; err != nil {
		return nil, 0, err
	}

	return workspaces, total, nil
}

// Join Request

func (r *workspaceRepository) CreateJoinRequest(req *domain.WorkspaceJoinRequest) error {
	return r.db.Create(req).Error
}

func (r *workspaceRepository) FindJoinRequestByID(id uuid.UUID) (*domain.WorkspaceJoinRequest, error) {
	var req domain.WorkspaceJoinRequest
	if err := r.db.Where("id = ?", id).First(&req).Error; err != nil {
		return nil, err
	}
	return &req, nil
}

func (r *workspaceRepository) FindJoinRequestsByWorkspace(workspaceID uuid.UUID, status string) ([]domain.WorkspaceJoinRequest, error) {
	var requests []domain.WorkspaceJoinRequest

	query := r.db.Where("workspace_id = ?", workspaceID)
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Order("requested_at DESC").Find(&requests).Error; err != nil {
		return nil, err
	}
	return requests, nil
}

func (r *workspaceRepository) FindJoinRequestByUserAndWorkspace(userID, workspaceID uuid.UUID) (*domain.WorkspaceJoinRequest, error) {
	var req domain.WorkspaceJoinRequest
	if err := r.db.Where("user_id = ? AND workspace_id = ?", userID, workspaceID).
		First(&req).Error; err != nil {
		return nil, err
	}
	return &req, nil
}

func (r *workspaceRepository) UpdateJoinRequest(req *domain.WorkspaceJoinRequest) error {
	return r.db.Save(req).Error
}

// Member

func (r *workspaceRepository) CreateMember(member *domain.WorkspaceMember) error {
	return r.db.Create(member).Error
}

func (r *workspaceRepository) FindMemberByID(id uuid.UUID) (*domain.WorkspaceMember, error) {
	var member domain.WorkspaceMember
	if err := r.db.Where("id = ? AND left_at IS NULL", id).First(&member).Error; err != nil {
		return nil, err
	}
	return &member, nil
}

func (r *workspaceRepository) FindMembersByWorkspace(workspaceID uuid.UUID) ([]domain.WorkspaceMember, error) {
	var members []domain.WorkspaceMember
	if err := r.db.Where("workspace_id = ? AND left_at IS NULL", workspaceID).
		Order("joined_at ASC").
		Find(&members).Error; err != nil {
		return nil, err
	}
	return members, nil
}

func (r *workspaceRepository) FindMemberByUserAndWorkspace(userID, workspaceID uuid.UUID) (*domain.WorkspaceMember, error) {
	var member domain.WorkspaceMember
	if err := r.db.Where("user_id = ? AND workspace_id = ? AND left_at IS NULL", userID, workspaceID).
		First(&member).Error; err != nil {
		return nil, err
	}
	return &member, nil
}

func (r *workspaceRepository) FindMembersByUser(userID uuid.UUID) ([]domain.WorkspaceMember, error) {
	var members []domain.WorkspaceMember
	if err := r.db.Where("user_id = ? AND left_at IS NULL", userID).
		Find(&members).Error; err != nil {
		return nil, err
	}
	return members, nil
}

func (r *workspaceRepository) UpdateMember(member *domain.WorkspaceMember) error {
	return r.db.Save(member).Error
}

func (r *workspaceRepository) DeleteMember(id uuid.UUID) error {
	// Soft delete by setting left_at
	return r.db.Model(&domain.WorkspaceMember{}).
		Where("id = ?", id).
		Update("left_at", gorm.Expr("CURRENT_TIMESTAMP")).Error
}

func (r *workspaceRepository) ClearDefaultWorkspace(userID uuid.UUID) error {
	return r.db.Model(&domain.WorkspaceMember{}).
		Where("user_id = ? AND is_default = ?", userID, true).
		Update("is_default", false).Error
}

func (r *workspaceRepository) SetDefaultWorkspace(userID, workspaceID uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Clear all default flags for this user
		if err := tx.Model(&domain.WorkspaceMember{}).
			Where("user_id = ?", userID).
			Update("is_default", false).Error; err != nil {
			return err
		}

		// Set new default
		return tx.Model(&domain.WorkspaceMember{}).
			Where("user_id = ? AND workspace_id = ? AND left_at IS NULL", userID, workspaceID).
			Update("is_default", true).Error
	})
}

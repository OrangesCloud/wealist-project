package repository

import (
	"board-service/internal/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RoleRepository interface {
	FindByName(name string) (*domain.Role, error)
	FindByID(id uuid.UUID) (*domain.Role, error)
}

type roleRepository struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) RoleRepository {
	return &roleRepository{db: db}
}

func (r *roleRepository) FindByName(name string) (*domain.Role, error) {
	var role domain.Role
	if err := r.db.Where("name = ?", name).First(&role).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) FindByID(id uuid.UUID) (*domain.Role, error) {
	var role domain.Role
	if err := r.db.Where("id = ?", id).First(&role).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

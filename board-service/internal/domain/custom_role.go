package domain

import "github.com/google/uuid"

type CustomRole struct {
	BaseModel
	ProjectID       uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_project_role_name" json:"project_id"`
	Name            string    `gorm:"type:varchar(50);not null;uniqueIndex:idx_project_role_name" json:"name"`
	Color           string    `gorm:"type:varchar(7)" json:"color"`
	IsSystemDefault bool      `gorm:"not null;default:false;index" json:"is_system_default"`
	DisplayOrder    int       `gorm:"not null;default:0" json:"display_order"`
}

func (CustomRole) TableName() string {
	return "custom_roles"
}

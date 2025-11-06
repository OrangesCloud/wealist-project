package domain

import "github.com/google/uuid"

type Project struct {
	BaseModel
	WorkspaceID uuid.UUID `gorm:"type:uuid;not null;index" json:"workspace_id"`
	Name        string    `gorm:"type:varchar(255);not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	OwnerID     uuid.UUID `gorm:"type:uuid;not null;index" json:"owner_id"`
	IsPublic    bool      `gorm:"default:false" json:"is_public"`
	IsDeleted   bool      `gorm:"default:false;index" json:"is_deleted"`
}

func (Project) TableName() string {
	return "projects"
}

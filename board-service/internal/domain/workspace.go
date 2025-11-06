package domain

import "github.com/google/uuid"

type Workspace struct {
	BaseModel
	Name        string    `gorm:"type:varchar(255);not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	OwnerID     uuid.UUID `gorm:"type:uuid;not null" json:"owner_id"`
	IsPublic    bool      `gorm:"default:false" json:"is_public"`
	IsDeleted   bool      `gorm:"default:false" json:"is_deleted"`
}

func (Workspace) TableName() string {
	return "workspaces"
}

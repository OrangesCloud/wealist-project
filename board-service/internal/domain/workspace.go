package domain

import "github.com/google/uuid"

type Workspace struct {
	BaseModel
	Name        string    `gorm:"type:varchar(255);not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	CreatedBy   uuid.UUID `gorm:"type:uuid;not null" json:"created_by"`
	IsDeleted   bool      `gorm:"default:false" json:"is_deleted"`
}

func (Workspace) TableName() string {
	return "workspaces"
}

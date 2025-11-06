package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Board struct {
	ID                 uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ProjectID          uuid.UUID      `gorm:"type:uuid;not null;index" json:"project_id"`
	Title              string         `gorm:"type:varchar(255);not null" json:"title"`
	Description        string         `gorm:"type:text" json:"description"`
	CustomStageID      uuid.UUID      `gorm:"type:uuid;not null;index" json:"custom_stage_id"`
	CustomImportanceID *uuid.UUID     `gorm:"type:uuid;index" json:"custom_importance_id"`
	AssigneeID         *uuid.UUID     `gorm:"type:uuid;index" json:"assignee_id"`
	CreatedBy          uuid.UUID      `gorm:"type:uuid;not null;index" json:"created_by"`
	DueDate            *time.Time     `gorm:"index" json:"due_date"`
	CreatedAt          time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt          time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
}

func (Board) TableName() string {
	return "boards"
}

// BeforeCreate is a GORM hook that generates UUID before creating a record
func (b *Board) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}

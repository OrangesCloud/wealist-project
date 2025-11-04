package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type KanbanRole struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	KanbanID     uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_kanban_custom_role" json:"kanban_id"`
	CustomRoleID uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_kanban_custom_role" json:"custom_role_id"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (KanbanRole) TableName() string {
	return "kanban_roles"
}

// BeforeCreate is a GORM hook that generates UUID before creating a record
func (kr *KanbanRole) BeforeCreate(tx *gorm.DB) error {
	if kr.ID == uuid.Nil {
		kr.ID = uuid.New()
	}
	return nil
}

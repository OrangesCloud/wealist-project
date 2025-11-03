package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserRoleColumnOrder stores user-specific display order for role columns
type UserRoleColumnOrder struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID       uuid.UUID `gorm:"type:uuid;not null;index:idx_user_role_order,priority:1" json:"user_id"`
	ProjectID    uuid.UUID `gorm:"type:uuid;not null;index:idx_user_role_order,priority:2" json:"project_id"`
	CustomRoleID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_user_role_unique" json:"custom_role_id"`
	DisplayOrder int       `gorm:"not null;default:0" json:"display_order"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (UserRoleColumnOrder) TableName() string {
	return "user_role_column_order"
}

func (u *UserRoleColumnOrder) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// UserStageColumnOrder stores user-specific display order for stage columns
type UserStageColumnOrder struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID        uuid.UUID `gorm:"type:uuid;not null;index:idx_user_stage_order,priority:1" json:"user_id"`
	ProjectID     uuid.UUID `gorm:"type:uuid;not null;index:idx_user_stage_order,priority:2" json:"project_id"`
	CustomStageID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_user_stage_unique" json:"custom_stage_id"`
	DisplayOrder  int       `gorm:"not null;default:0" json:"display_order"`
	CreatedAt     time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (UserStageColumnOrder) TableName() string {
	return "user_stage_column_order"
}

func (u *UserStageColumnOrder) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// UserKanbanOrderInRole stores user-specific display order for kanbans within each role column
type UserKanbanOrderInRole struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID       uuid.UUID `gorm:"type:uuid;not null;index:idx_user_kanban_role_order,priority:1" json:"user_id"`
	ProjectID    uuid.UUID `gorm:"type:uuid;not null;index:idx_user_kanban_role_order,priority:2" json:"project_id"`
	CustomRoleID uuid.UUID `gorm:"type:uuid;not null;index:idx_user_kanban_role_order,priority:3" json:"custom_role_id"`
	KanbanID     uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_user_kanban_role_unique" json:"kanban_id"`
	DisplayOrder int       `gorm:"not null;default:0" json:"display_order"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (UserKanbanOrderInRole) TableName() string {
	return "user_kanban_order_in_role"
}

func (u *UserKanbanOrderInRole) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// UserKanbanOrderInStage stores user-specific display order for kanbans within each stage column
type UserKanbanOrderInStage struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID        uuid.UUID `gorm:"type:uuid;not null;index:idx_user_kanban_stage_order,priority:1" json:"user_id"`
	ProjectID     uuid.UUID `gorm:"type:uuid;not null;index:idx_user_kanban_stage_order,priority:2" json:"project_id"`
	CustomStageID uuid.UUID `gorm:"type:uuid;not null;index:idx_user_kanban_stage_order,priority:3" json:"custom_stage_id"`
	KanbanID      uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_user_kanban_stage_unique" json:"kanban_id"`
	DisplayOrder  int       `gorm:"not null;default:0" json:"display_order"`
	CreatedAt     time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (UserKanbanOrderInStage) TableName() string {
	return "user_kanban_order_in_stage"
}

func (u *UserKanbanOrderInStage) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

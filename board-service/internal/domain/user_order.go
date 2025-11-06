package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserRoleColumnOrder stores user-specific display order for role columns
type UserRoleColumnOrder struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID       uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_user_role_unique,priority:1" json:"user_id"`
	ProjectID    uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_user_role_unique,priority:2" json:"project_id"`
	CustomRoleID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_user_role_unique,priority:3" json:"custom_role_id"`
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
	UserID        uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_user_stage_unique,priority:1" json:"user_id"`
	ProjectID     uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_user_stage_unique,priority:2" json:"project_id"`
	CustomStageID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_user_stage_unique,priority:3" json:"custom_stage_id"`
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

// UserBoardOrderInRole stores user-specific display order for boards within each role column
type UserBoardOrderInRole struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID       uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_user_board_role_unique,priority:1" json:"user_id"`
	ProjectID    uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_user_board_role_unique,priority:2" json:"project_id"`
	CustomRoleID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_user_board_role_unique,priority:3" json:"custom_role_id"`
	BoardID      uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_user_board_role_unique,priority:4" json:"board_id"`
	DisplayOrder int       `gorm:"not null;default:0" json:"display_order"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (UserBoardOrderInRole) TableName() string {
	return "user_board_order_in_role"
}

func (u *UserBoardOrderInRole) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// UserBoardOrderInStage stores user-specific display order for boards within each stage column
type UserBoardOrderInStage struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID        uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_user_board_stage_unique,priority:1" json:"user_id"`
	ProjectID     uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_user_board_stage_unique,priority:2" json:"project_id"`
	CustomStageID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_user_board_stage_unique,priority:3" json:"custom_stage_id"`
	BoardID       uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_user_board_stage_unique,priority:4" json:"board_id"`
	DisplayOrder  int       `gorm:"not null;default:0" json:"display_order"`
	CreatedAt     time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (UserBoardOrderInStage) TableName() string {
	return "user_board_order_in_stage"
}

func (u *UserBoardOrderInStage) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

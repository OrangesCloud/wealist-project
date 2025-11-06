package domain

import (
	"github.com/google/uuid"
)

// UserRoleColumnOrder stores user-specific display order for role columns
type UserRoleColumnOrder struct {
	BaseModel
	UserID       uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_user_role_unique,priority:1" json:"user_id"`
	ProjectID    uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_user_role_unique,priority:2" json:"project_id"`
	CustomRoleID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_user_role_unique,priority:3" json:"custom_role_id"`
	DisplayOrder int       `gorm:"not null;default:0" json:"display_order"`
}

func (UserRoleColumnOrder) TableName() string {
	return "user_role_column_order"
}

// UserStageColumnOrder stores user-specific display order for stage columns
type UserStageColumnOrder struct {
	BaseModel
	UserID        uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_user_stage_unique,priority:1" json:"user_id"`
	ProjectID     uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_user_stage_unique,priority:2" json:"project_id"`
	CustomStageID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_user_stage_unique,priority:3" json:"custom_stage_id"`
	DisplayOrder  int       `gorm:"not null;default:0" json:"display_order"`
}

func (UserStageColumnOrder) TableName() string {
	return "user_stage_column_order"
}

// UserBoardOrderInRole stores user-specific display order for boards within each role column
type UserBoardOrderInRole struct {
	BaseModel
	UserID       uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_user_board_role_unique,priority:1" json:"user_id"`
	ProjectID    uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_user_board_role_unique,priority:2" json:"project_id"`
	CustomRoleID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_user_board_role_unique,priority:3" json:"custom_role_id"`
	BoardID      uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_user_board_role_unique,priority:4" json:"board_id"`
	DisplayOrder int       `gorm:"not null;default:0" json:"display_order"`
}

func (UserBoardOrderInRole) TableName() string {
	return "user_board_order_in_role"
}

// UserBoardOrderInStage stores user-specific display order for boards within each stage column
type UserBoardOrderInStage struct {
	BaseModel
	UserID        uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_user_board_stage_unique,priority:1" json:"user_id"`
	ProjectID     uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_user_board_stage_unique,priority:2" json:"project_id"`
	CustomStageID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_user_board_stage_unique,priority:3" json:"custom_stage_id"`
	BoardID       uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_user_board_stage_unique,priority:4" json:"board_id"`
	DisplayOrder  int       `gorm:"not null;default:0" json:"display_order"`
}

func (UserBoardOrderInStage) TableName() string {
	return "user_board_order_in_stage"
}

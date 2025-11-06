package domain

import (
	"github.com/google/uuid"
)

type BoardRole struct {
	BaseModel
	BoardID      uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_board_custom_role" json:"board_id"`
	CustomRoleID uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_board_custom_role" json:"custom_role_id"`
}

func (BoardRole) TableName() string {
	return "board_roles"
}

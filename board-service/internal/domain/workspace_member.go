package domain

import (
	"time"

	"github.com/google/uuid"
)

type WorkspaceMember struct {
	BaseModel
	WorkspaceID uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:idx_workspace_user" json:"workspace_id"`
	UserID      uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:idx_workspace_user" json:"user_id"`
	RoleID      uuid.UUID  `gorm:"type:uuid;not null" json:"role_id"`
	JoinedAt    time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP" json:"joined_at"`
	LeftAt      *time.Time `json:"left_at"`
	IsDefault   bool       `gorm:"default:false" json:"is_default"`
}

func (WorkspaceMember) TableName() string {
	return "workspace_members"
}

package domain

import (
	"time"

	"github.com/google/uuid"
)

type ProjectJoinRequestStatus string

const (
	ProjectJoinRequestPending  ProjectJoinRequestStatus = "PENDING"
	ProjectJoinRequestApproved ProjectJoinRequestStatus = "APPROVED"
	ProjectJoinRequestRejected ProjectJoinRequestStatus = "REJECTED"
)

type ProjectJoinRequest struct {
	BaseModel
	ProjectID   uuid.UUID                `gorm:"type:uuid;not null;index;uniqueIndex:idx_project_user_request" json:"project_id"`
	UserID      uuid.UUID                `gorm:"type:uuid;not null;index;uniqueIndex:idx_project_user_request" json:"user_id"`
	Status      ProjectJoinRequestStatus `gorm:"type:varchar(20);not null;default:'PENDING';index" json:"status"`
	RequestedAt time.Time                `gorm:"not null;default:CURRENT_TIMESTAMP" json:"requested_at"`
}

func (ProjectJoinRequest) TableName() string {
	return "project_join_requests"
}

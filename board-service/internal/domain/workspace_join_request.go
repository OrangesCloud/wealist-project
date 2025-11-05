package domain

import (
	"time"

	"github.com/google/uuid"
)

type JoinRequestStatus string

const (
	JoinRequestPending  JoinRequestStatus = "PENDING"
	JoinRequestApproved JoinRequestStatus = "APPROVED"
	JoinRequestRejected JoinRequestStatus = "REJECTED"
)

type WorkspaceJoinRequest struct {
	BaseModel
	WorkspaceID uuid.UUID         `gorm:"type:uuid;not null" json:"workspace_id"`
	UserID      uuid.UUID         `gorm:"type:uuid;not null" json:"user_id"`
	Status      JoinRequestStatus `gorm:"type:varchar(20);not null;default:'PENDING'" json:"status"`
	RequestedAt time.Time         `gorm:"not null;default:CURRENT_TIMESTAMP" json:"requested_at"`
	ProcessedAt *time.Time        `json:"processed_at"`
	ProcessedBy *uuid.UUID        `gorm:"type:uuid" json:"processed_by"`
}

func (WorkspaceJoinRequest) TableName() string {
	return "workspace_join_requests"
}

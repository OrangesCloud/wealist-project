package domain

import (
	"time"

	"github.com/google/uuid"
)

// Project represents a project entity within a workspace
type Project struct {
	BaseModel
	WorkspaceID uuid.UUID        `gorm:"type:uuid;not null;index" json:"workspace_id"`
	OwnerID     uuid.UUID        `gorm:"type:uuid;not null;index" json:"owner_id"`
	Name        string           `gorm:"type:varchar(255);not null" json:"name"`
	Description string           `gorm:"type:text" json:"description"`
	IsDefault   bool             `gorm:"default:false;index" json:"is_default"`
	IsPublic    bool             `gorm:"default:false" json:"is_public"`
	Boards      []Board          `gorm:"foreignKey:ProjectID" json:"boards,omitempty"`
	Members     []ProjectMember  `gorm:"foreignKey:ProjectID" json:"members,omitempty"`
	JoinRequests []ProjectJoinRequest `gorm:"foreignKey:ProjectID" json:"join_requests,omitempty"`
}

// ProjectRole represents the role of a project member
type ProjectRole string

const (
	ProjectRoleOwner  ProjectRole = "OWNER"
	ProjectRoleAdmin  ProjectRole = "ADMIN"
	ProjectRoleMember ProjectRole = "MEMBER"
)

// ProjectMember represents a member of a project
type ProjectMember struct {
	ID        uuid.UUID   `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ProjectID uuid.UUID   `gorm:"type:uuid;not null;index" json:"project_id"`
	UserID    uuid.UUID   `gorm:"type:uuid;not null;index" json:"user_id"`
	RoleName  ProjectRole `gorm:"type:varchar(50);not null" json:"role_name"`
	JoinedAt  time.Time   `gorm:"type:timestamp;not null;default:now()" json:"joined_at"`
	Project   Project     `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
}

// ProjectJoinRequestStatus represents the status of a join request
type ProjectJoinRequestStatus string

const (
	JoinRequestPending  ProjectJoinRequestStatus = "PENDING"
	JoinRequestApproved ProjectJoinRequestStatus = "APPROVED"
	JoinRequestRejected ProjectJoinRequestStatus = "REJECTED"
)

// ProjectJoinRequest represents a request to join a project
type ProjectJoinRequest struct {
	ID          uuid.UUID                `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ProjectID   uuid.UUID                `gorm:"type:uuid;not null;index" json:"project_id"`
	UserID      uuid.UUID                `gorm:"type:uuid;not null;index" json:"user_id"`
	Status      ProjectJoinRequestStatus `gorm:"type:varchar(50);not null;default:'PENDING'" json:"status"`
	RequestedAt time.Time                `gorm:"type:timestamp;not null;default:now()" json:"requested_at"`
	UpdatedAt   time.Time                `gorm:"type:timestamp;not null;default:now()" json:"updated_at"`
	Project     Project                  `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
}

// TableName specifies the table name for Project
func (Project) TableName() string {
	return "projects"
}

// TableName specifies the table name for ProjectMember
func (ProjectMember) TableName() string {
	return "project_members"
}

// TableName specifies the table name for ProjectJoinRequest
func (ProjectJoinRequest) TableName() string {
	return "project_join_requests"
}

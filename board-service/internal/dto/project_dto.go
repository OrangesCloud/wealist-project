package dto

import (
	"time"

	"github.com/google/uuid"
)

// CreateProjectRequest represents the request to create a new project
type CreateProjectRequest struct {
	WorkspaceID uuid.UUID `json:"workspaceId" binding:"required"`
	Name        string    `json:"name" binding:"required,min=2,max=100"`
	Description string    `json:"description" binding:"max=500"`
}

// UpdateProjectRequest represents the request to update a project
type UpdateProjectRequest struct {
	Name        *string `json:"name" binding:"omitempty,min=2,max=100"`
	Description *string `json:"description" binding:"omitempty,max=500"`
}

// ProjectResponse represents the project response
type ProjectResponse struct {
	ID          uuid.UUID `json:"projectId"`
	WorkspaceID uuid.UUID `json:"workspaceId"`
	OwnerID     uuid.UUID `json:"ownerId"`
	OwnerEmail  string    `json:"ownerEmail,omitempty"`
	OwnerName   string    `json:"ownerName,omitempty"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsPublic    bool      `json:"isPublic"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// ProjectMemberResponse represents a project member
type ProjectMemberResponse struct {
	MemberID  uuid.UUID `json:"memberId"`
	ProjectID uuid.UUID `json:"projectId"`
	UserID    uuid.UUID `json:"userId"`
	UserEmail string    `json:"userEmail,omitempty"`
	UserName  string    `json:"userName,omitempty"`
	RoleName  string    `json:"roleName"`
	JoinedAt  time.Time `json:"joinedAt"`
}

// UpdateProjectMemberRoleRequest represents the request to update member role
type UpdateProjectMemberRoleRequest struct {
	RoleName string `json:"roleName" binding:"required,oneof=OWNER ADMIN MEMBER"`
}

// ProjectJoinRequestResponse represents a join request
type ProjectJoinRequestResponse struct {
	RequestID   uuid.UUID `json:"requestId"`
	ProjectID   uuid.UUID `json:"projectId"`
	UserID      uuid.UUID `json:"userId"`
	UserEmail   string    `json:"userEmail,omitempty"`
	UserName    string    `json:"userName,omitempty"`
	Status      string    `json:"status"`
	RequestedAt time.Time `json:"requestedAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// CreateProjectJoinRequestRequest represents the request to join a project
type CreateProjectJoinRequestRequest struct {
	ProjectID uuid.UUID `json:"projectId" binding:"required"`
}

// UpdateProjectJoinRequestRequest represents the request to update join request status
type UpdateProjectJoinRequestRequest struct {
	Status string `json:"status" binding:"required,oneof=APPROVED REJECTED"`
}

// PaginatedProjectsResponse represents paginated projects response
type PaginatedProjectsResponse struct {
	Projects []ProjectResponse `json:"projects"`
	Total    int64             `json:"total"`
	Page     int               `json:"page"`
	Limit    int               `json:"limit"`
}

// ProjectInitSettingsResponse represents the initial settings for a project
type ProjectInitSettingsResponse struct {
	Project       ProjectBasicInfo           `json:"project"`
	Fields        []FieldWithOptionsResponse `json:"fields"`
	FieldTypes    []FieldTypeInfo            `json:"fieldTypes"`
	DefaultViewID *uuid.UUID                 `json:"defaultViewId,omitempty"`
}

// ProjectBasicInfo represents basic project information
type ProjectBasicInfo struct {
	ProjectID   uuid.UUID `json:"projectId"`
	WorkspaceID uuid.UUID `json:"workspaceId"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	OwnerID     uuid.UUID `json:"ownerId"`
	IsPublic    bool      `json:"isPublic"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// FieldWithOptionsResponse represents a field definition with its options
type FieldWithOptionsResponse struct {
	FieldID     string        `json:"fieldId"`
	FieldName   string        `json:"fieldName"`
	FieldType   string        `json:"fieldType"`
	IsRequired  bool          `json:"isRequired"`
	Options     []FieldOption `json:"options"`
	Description string        `json:"description,omitempty"`
}

// FieldOption represents an option for a field
type FieldOption struct {
	OptionID    string `json:"optionId"`
	OptionLabel string `json:"optionLabel"`
	OptionValue string `json:"optionValue"`
}

// FieldTypeInfo represents information about a field type
type FieldTypeInfo struct {
	TypeID      string `json:"typeId"`
	TypeName    string `json:"typeName"`
	Description string `json:"description,omitempty"`
}

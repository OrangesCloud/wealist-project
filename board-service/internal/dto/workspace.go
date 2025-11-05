package dto

import "time"

// Request DTOs

type CreateWorkspaceRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=100"`
	Description string `json:"description" binding:"max=500"`
}

type UpdateWorkspaceRequest struct {
	Name        string `json:"name" binding:"omitempty,min=2,max=100"`
	Description string `json:"description" binding:"omitempty,max=500"`
}

type SearchWorkspacesRequest struct {
	Query string `form:"query" binding:"required,min=1"`
	Page  int    `form:"page" binding:"omitempty,min=1"`
	Limit int    `form:"limit" binding:"omitempty,min=1,max=100"`
}

type CreateJoinRequestRequest struct {
	WorkspaceID string `json:"workspaceId" binding:"required,uuid"`
}

type UpdateJoinRequestRequest struct {
	Status string `json:"status" binding:"required,oneof=APPROVED REJECTED"`
}

type UpdateMemberRoleRequest struct {
	RoleName string `json:"roleName" binding:"required,oneof=OWNER ADMIN MEMBER"`
}

type SetDefaultWorkspaceRequest struct {
	WorkspaceID string `json:"workspaceId" binding:"required,uuid"`
}

// Response DTOs

type WorkspaceResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	OwnerID     string    `json:"ownerId"`
	OwnerName   string    `json:"ownerName"`
	OwnerEmail  string    `json:"ownerEmail"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type WorkspaceMemberResponse struct {
	ID          string    `json:"id"`
	WorkspaceID string    `json:"workspaceId"`
	UserID      string    `json:"userId"`
	UserName    string    `json:"userName"`
	UserEmail   string    `json:"userEmail"`
	RoleName    string    `json:"roleName"`
	IsDefault   bool      `json:"isDefault"`
	JoinedAt    time.Time `json:"joinedAt"`
}

type JoinRequestResponse struct {
	ID          string    `json:"id"`
	WorkspaceID string    `json:"workspaceId"`
	UserID      string    `json:"userId"`
	UserName    string    `json:"userName"`
	UserEmail   string    `json:"userEmail"`
	Status      string    `json:"status"`
	RequestedAt time.Time `json:"requestedAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type PaginatedWorkspacesResponse struct {
	Workspaces []WorkspaceResponse `json:"workspaces"`
	Total      int64               `json:"total"`
	Page       int                 `json:"page"`
	Limit      int                 `json:"limit"`
}

package dto

import "time"

// ==================== Custom Roles ====================

// Request DTOs
type CreateCustomRoleRequest struct {
	ProjectID string `json:"projectId" binding:"required,uuid"`
	Name      string `json:"name" binding:"required,min=1,max=50"`
	Color     string `json:"color" binding:"omitempty,len=7"` // #RRGGBB
}

type UpdateCustomRoleRequest struct {
	Name  string `json:"name" binding:"omitempty,min=1,max=50"`
	Color string `json:"color" binding:"omitempty,len=7"`
}

type UpdateCustomRoleOrderRequest struct {
	RoleOrders []RoleOrder `json:"roleOrders" binding:"required,min=1,dive"`
}

type RoleOrder struct {
	RoleID       string `json:"roleId" binding:"required,uuid"`
	DisplayOrder int    `json:"displayOrder" binding:"min=0"`
}

// Response DTOs
type CustomRoleResponse struct {
	ID              string    `json:"id"`
	ProjectID       string    `json:"projectId"`
	Name            string    `json:"name"`
	Color           string    `json:"color"`
	IsSystemDefault bool      `json:"isSystemDefault"`
	DisplayOrder    int       `json:"displayOrder"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

// ==================== Custom Stages ====================

// Request DTOs
type CreateCustomStageRequest struct {
	ProjectID string `json:"projectId" binding:"required,uuid"`
	Name      string `json:"name" binding:"required,min=1,max=50"`
	Color     string `json:"color" binding:"omitempty,len=7"`
}

type UpdateCustomStageRequest struct {
	Name  string `json:"name" binding:"omitempty,min=1,max=50"`
	Color string `json:"color" binding:"omitempty,len=7"`
}

type UpdateCustomStageOrderRequest struct {
	StageOrders []StageOrder `json:"stageOrders" binding:"required,min=1,dive"`
}

type StageOrder struct {
	StageID      string `json:"stageId" binding:"required,uuid"`
	DisplayOrder int    `json:"displayOrder" binding:"min=0"`
}

// Response DTOs
type CustomStageResponse struct {
	ID              string    `json:"id"`
	ProjectID       string    `json:"projectId"`
	Name            string    `json:"name"`
	Color           string    `json:"color"`
	IsSystemDefault bool      `json:"isSystemDefault"`
	DisplayOrder    int       `json:"displayOrder"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

// ==================== Custom Importance ====================

// Request DTOs
type CreateCustomImportanceRequest struct {
	ProjectID string `json:"projectId" binding:"required,uuid"`
	Name      string `json:"name" binding:"required,min=1,max=50"`
	Color     string `json:"color" binding:"omitempty,len=7"`
}

type UpdateCustomImportanceRequest struct {
	Name  string `json:"name" binding:"omitempty,min=1,max=50"`
	Color string `json:"color" binding:"omitempty,len=7"`
}

type UpdateCustomImportanceOrderRequest struct {
	ImportanceOrders []ImportanceOrder `json:"importanceOrders" binding:"required,min=1,dive"`
}

type ImportanceOrder struct {
	ImportanceID string `json:"importanceId" binding:"required,uuid"`
	DisplayOrder int    `json:"displayOrder" binding:"min=0"`
}

// Response DTOs
type CustomImportanceResponse struct {
	ID              string    `json:"id"`
	ProjectID       string    `json:"projectId"`
	Name            string    `json:"name"`
	Color           string    `json:"color"`
	IsSystemDefault bool      `json:"isSystemDefault"`
	DisplayOrder    int       `json:"displayOrder"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

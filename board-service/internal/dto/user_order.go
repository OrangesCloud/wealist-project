package dto

// ==================== Request DTOs ====================

// UpdateOrderRequest is used for drag-and-drop reordering
type UpdateOrderRequest struct {
	ItemIds []string `json:"itemIds" binding:"required,min=1,dive,uuid"` // 다른 핸들러와 일치하도록 수정
}

// ==================== Response DTOs ====================

// ColumnOrderResponse represents a single column's order information
type ColumnOrderResponse struct {
	ID            string `json:"id"`
	CustomRoleID  string `json:"customRoleId,omitempty"`
	CustomStageID string `json:"customStageId,omitempty"`
	DisplayOrder  int    `json:"displayOrder"`
	Name          string `json:"name"` // Retrieved from cache/DB
}

// KanbanOrderResponse represents a single kanban's order information
type KanbanOrderResponse struct {
	KanbanID     string `json:"kanbanId"`
	Title        string `json:"title"`
	DisplayOrder int    `json:"displayOrder"`
}

// RoleBasedBoardView represents the complete board view organized by roles
type RoleBasedBoardView struct {
	ProjectID string           `json:"projectId"`
	ViewType  string           `json:"viewType"` // "role"
	Columns   []RoleColumnView `json:"columns"`
}

// RoleColumnView represents a single role column with its kanbans
type RoleColumnView struct {
	CustomRoleID string                `json:"customRoleId"`
	RoleName     string                `json:"roleName"`
	RoleColor    string                `json:"roleColor,omitempty"`
	DisplayOrder int                   `json:"displayOrder"`
	Kanbans      []KanbanOrderResponse `json:"kanbans"`
}

// StageBasedBoardView represents the complete board view organized by stages
type StageBasedBoardView struct {
	ProjectID string            `json:"projectId"`
	ViewType  string            `json:"viewType"` // "stage"
	Columns   []StageColumnView `json:"columns"`
}

// StageColumnView represents a single stage column with its kanbans
type StageColumnView struct {
	CustomStageID string                `json:"customStageId"`
	StageName     string                `json:"stageName"`
	StageColor    string                `json:"stageColor,omitempty"`
	DisplayOrder  int                   `json:"displayOrder"`
	Kanbans       []KanbanOrderResponse `json:"kanbans"`
}

// ==================== Internal DTOs ====================

// RoleColumnOrderData is used internally for caching
type RoleColumnOrderData struct {
	CustomRoleID string `json:"customRoleId"`
	DisplayOrder int    `json:"displayOrder"`
	RoleName     string `json:"roleName"`
	RoleColor    string `json:"roleColor"`
}

// StageColumnOrderData is used internally for caching
type StageColumnOrderData struct {
	CustomStageID string `json:"customStageId"`
	DisplayOrder  int    `json:"displayOrder"`
	StageName     string `json:"stageName"`
	StageColor    string `json:"stageColor"`
}

// KanbanOrderData is used internally for caching
type KanbanOrderData struct {
	KanbanID     string `json:"kanbanId"`
	DisplayOrder int    `json:"displayOrder"`
	Title        string `json:"title"`
}

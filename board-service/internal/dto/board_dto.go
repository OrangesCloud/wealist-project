package dto

import (
	"time"

	"github.com/google/uuid"
)

// CreateBoardRequest represents the request to create a new board
type CreateBoardRequest struct {
	ProjectID    uuid.UUID              `json:"projectId" binding:"required"`
	Title        string                 `json:"title" binding:"required,min=1,max=200"`
	Content      string                 `json:"content" binding:"max=5000"`
	CustomFields map[string]interface{} `json:"customFields"`
	AssigneeID   *uuid.UUID             `json:"assigneeId"`
	DueDate      *time.Time             `json:"dueDate"`
}

// UpdateBoardRequest represents the request to update a board
type UpdateBoardRequest struct {
	Title        *string                 `json:"title" binding:"omitempty,min=1,max=200"`
	Content      *string                 `json:"content" binding:"omitempty,max=5000"`
	CustomFields *map[string]interface{} `json:"customFields"`
	AssigneeID   *uuid.UUID              `json:"assigneeId"`
	DueDate      *time.Time              `json:"dueDate"`
}

// UpdateBoardFieldRequest represents the request to update a single board field
type UpdateBoardFieldRequest struct {
	FieldID string `json:"fieldId" binding:"required,oneof=stage importance role"`
	Value   string `json:"value" binding:"required"`
}

// BoardResponse represents the board response
type BoardResponse struct {
	ID           uuid.UUID              `json:"boardId"`
	ProjectID    uuid.UUID              `json:"projectId"`
	AuthorID     uuid.UUID              `json:"authorId"`
	AssigneeID   *uuid.UUID             `json:"assigneeId,omitempty"`
	Title        string                 `json:"title"`
	Content      string                 `json:"content"`
	CustomFields map[string]interface{} `json:"customFields"`
	DueDate      *time.Time             `json:"dueDate,omitempty"`
	CreatedAt    time.Time              `json:"createdAt"`
	UpdatedAt    time.Time              `json:"updatedAt"`
}

// PaginatedBoardsResponse represents paginated boards response
type PaginatedBoardsResponse struct {
	Boards []BoardResponse `json:"boards"`
	Total  int64           `json:"total"`
	Page   int             `json:"page"`
	Limit  int             `json:"limit"`
}

// BoardDetailResponse represents the detailed board response with participants and comments
type BoardDetailResponse struct {
	BoardResponse
	Participants []ParticipantResponse `json:"participants"`
	Comments     []CommentResponse     `json:"comments"`
}

// BoardFilters represents the filter parameters for board queries
type BoardFilters struct {
	CustomFields map[string]interface{} `json:"customFields,omitempty"`
}

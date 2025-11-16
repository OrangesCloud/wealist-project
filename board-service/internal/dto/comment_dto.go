package dto

import (
	"time"

	"github.com/google/uuid"
)

// CreateCommentRequest represents the request to create a new comment
type CreateCommentRequest struct {
	BoardID uuid.UUID `json:"boardId" binding:"required"`
	Content string    `json:"content" binding:"required,min=1"`
}

// UpdateCommentRequest represents the request to update a comment
type UpdateCommentRequest struct {
	Content string `json:"content" binding:"required,min=1"`
}

// CommentResponse represents the comment response
type CommentResponse struct {
	CommentID uuid.UUID `json:"commentId"`
	BoardID   uuid.UUID `json:"boardId"`
	UserID    uuid.UUID `json:"userId"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

package dto

import (
	"time"

	"github.com/google/uuid"
)

// AddParticipantsRequest represents the request to add one or more participants to a board
type AddParticipantsRequest struct {
	BoardID uuid.UUID   `json:"boardId" binding:"required"`
	UserIDs []uuid.UUID `json:"userIds" binding:"required,min=1,max=50"`
}

// ParticipantResult represents the result of adding a single participant
type ParticipantResult struct {
	UserID  uuid.UUID `json:"userId"`
	Success bool      `json:"success"`
	Error   string    `json:"error,omitempty"`
}

// AddParticipantsResponse represents the response for adding participants
type AddParticipantsResponse struct {
	TotalRequested int                 `json:"totalRequested"`
	TotalSuccess   int                 `json:"totalSuccess"`
	TotalFailed    int                 `json:"totalFailed"`
	Results        []ParticipantResult `json:"results"`
}

// ParticipantResponse represents the participant response
type ParticipantResponse struct {
	ID        uuid.UUID `json:"id"`
	BoardID   uuid.UUID `json:"boardId"`
	UserID    uuid.UUID `json:"userId"`
	CreatedAt time.Time `json:"createdAt"`
}

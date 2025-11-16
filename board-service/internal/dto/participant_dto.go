package dto

import (
	"time"

	"github.com/google/uuid"
)

// AddParticipantRequest represents the request to add a participant to a board
type AddParticipantRequest struct {
	BoardID uuid.UUID `json:"boardId" binding:"required"`
	UserID  uuid.UUID `json:"userId" binding:"required"`
}

// ParticipantResponse represents the participant response
type ParticipantResponse struct {
	ID        uuid.UUID `json:"id"`
	BoardID   uuid.UUID `json:"boardId"`
	UserID    uuid.UUID `json:"userId"`
	CreatedAt time.Time `json:"createdAt"`
}

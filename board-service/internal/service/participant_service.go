package service

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"project-board-api/internal/domain"
	"project-board-api/internal/dto"
	"project-board-api/internal/repository"
	"project-board-api/internal/response"
)

// ParticipantService defines the interface for participant business logic
type ParticipantService interface {
	AddParticipant(ctx context.Context, req *dto.AddParticipantRequest) error
	GetParticipants(ctx context.Context, boardID uuid.UUID) ([]*dto.ParticipantResponse, error)
	RemoveParticipant(ctx context.Context, boardID, userID uuid.UUID) error
}

// participantServiceImpl is the implementation of ParticipantService
type participantServiceImpl struct {
	participantRepo repository.ParticipantRepository
	boardRepo       repository.BoardRepository
}

// NewParticipantService creates a new instance of ParticipantService
func NewParticipantService(participantRepo repository.ParticipantRepository, boardRepo repository.BoardRepository) ParticipantService {
	return &participantServiceImpl{
		participantRepo: participantRepo,
		boardRepo:       boardRepo,
	}
}

// AddParticipant adds a participant to a board
func (s *participantServiceImpl) AddParticipant(ctx context.Context, req *dto.AddParticipantRequest) error {
	// Verify board exists
	_, err := s.boardRepo.FindByID(ctx, req.BoardID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.NewAppError(response.ErrCodeNotFound, "Board not found", "")
		}
		return response.NewAppError(response.ErrCodeInternal, "Failed to verify board", err.Error())
	}

	// Check if participant already exists
	existing, err := s.participantRepo.FindByBoardAndUser(ctx, req.BoardID, req.UserID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return response.NewAppError(response.ErrCodeInternal, "Failed to check existing participant", err.Error())
	}
	if existing != nil {
		return response.NewAppError(response.ErrCodeAlreadyExists, "Participant already exists", "")
	}

	// Create domain model from request
	participant := &domain.Participant{
		BoardID: req.BoardID,
		UserID:  req.UserID,
	}

	// Save to repository
	if err := s.participantRepo.Create(ctx, participant); err != nil {
		// Check for unique constraint violation
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return response.NewAppError(response.ErrCodeAlreadyExists, "Participant already exists", "")
		}
		return response.NewAppError(response.ErrCodeInternal, "Failed to add participant", err.Error())
	}

	return nil
}

// GetParticipants retrieves all participants for a board
func (s *participantServiceImpl) GetParticipants(ctx context.Context, boardID uuid.UUID) ([]*dto.ParticipantResponse, error) {
	// Verify board exists
	_, err := s.boardRepo.FindByID(ctx, boardID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, response.NewAppError(response.ErrCodeNotFound, "Board not found", "")
		}
		return nil, response.NewAppError(response.ErrCodeInternal, "Failed to verify board", err.Error())
	}

	// Fetch participants from repository
	participants, err := s.participantRepo.FindByBoardID(ctx, boardID)
	if err != nil {
		return nil, response.NewAppError(response.ErrCodeInternal, "Failed to fetch participants", err.Error())
	}

	// Convert to response DTOs
	responses := make([]*dto.ParticipantResponse, len(participants))
	for i, participant := range participants {
		responses[i] = s.toParticipantResponse(participant)
	}

	return responses, nil
}

// RemoveParticipant removes a participant from a board
func (s *participantServiceImpl) RemoveParticipant(ctx context.Context, boardID, userID uuid.UUID) error {
	// Verify board exists
	_, err := s.boardRepo.FindByID(ctx, boardID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.NewAppError(response.ErrCodeNotFound, "Board not found", "")
		}
		return response.NewAppError(response.ErrCodeInternal, "Failed to verify board", err.Error())
	}

	// Check if participant exists
	_, err = s.participantRepo.FindByBoardAndUser(ctx, boardID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.NewAppError(response.ErrCodeNotFound, "Participant not found", "")
		}
		return response.NewAppError(response.ErrCodeInternal, "Failed to verify participant", err.Error())
	}

	// Delete participant
	if err := s.participantRepo.Delete(ctx, boardID, userID); err != nil {
		return response.NewAppError(response.ErrCodeInternal, "Failed to remove participant", err.Error())
	}

	return nil
}

// toParticipantResponse converts domain.Participant to dto.ParticipantResponse
func (s *participantServiceImpl) toParticipantResponse(participant *domain.Participant) *dto.ParticipantResponse {
	return &dto.ParticipantResponse{
		ID:        participant.ID,
		BoardID:   participant.BoardID,
		UserID:    participant.UserID,
		CreatedAt: participant.CreatedAt,
	}
}

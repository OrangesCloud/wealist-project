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
	AddParticipants(ctx context.Context, req *dto.AddParticipantsRequest) (*dto.AddParticipantsResponse, error)
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

// AddParticipants adds one or more participants to a board (supports single and bulk operations)
func (s *participantServiceImpl) AddParticipants(ctx context.Context, req *dto.AddParticipantsRequest) (*dto.AddParticipantsResponse, error) {
	// Verify board exists
	_, err := s.boardRepo.FindByID(ctx, req.BoardID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, response.NewAppError(response.ErrCodeNotFound, "Board not found", "")
		}
		return nil, response.NewAppError(response.ErrCodeInternal, "Failed to verify board", err.Error())
	}

	// Remove duplicates from the request
	uniqueUserIDs := removeDuplicateUUIDs(req.UserIDs)

	// Initialize response
	response := &dto.AddParticipantsResponse{
		TotalRequested: len(uniqueUserIDs),
		TotalSuccess:   0,
		TotalFailed:    0,
		Results:        make([]dto.ParticipantResult, 0, len(uniqueUserIDs)),
	}

	// Process each participant individually
	for _, userID := range uniqueUserIDs {
		result := s.addSingleParticipant(ctx, req.BoardID, userID)
		response.Results = append(response.Results, result)
		
		if result.Success {
			response.TotalSuccess++
		} else {
			response.TotalFailed++
		}
	}

	return response, nil
}

// addSingleParticipant attempts to add a single participant and returns the result
func (s *participantServiceImpl) addSingleParticipant(ctx context.Context, boardID, userID uuid.UUID) dto.ParticipantResult {
	result := dto.ParticipantResult{
		UserID:  userID,
		Success: false,
	}

	// Check if participant already exists
	existing, err := s.participantRepo.FindByBoardAndUser(ctx, boardID, userID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		result.Error = "Failed to check existing participant"
		return result
	}
	if existing != nil {
		result.Error = "Participant already exists"
		return result
	}

	// Create domain model
	participant := &domain.Participant{
		BoardID: boardID,
		UserID:  userID,
	}

	// Save to repository
	if err := s.participantRepo.Create(ctx, participant); err != nil {
		// Check for unique constraint violation
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			result.Error = "Participant already exists"
		} else {
			result.Error = "Failed to add participant"
		}
		return result
	}

	result.Success = true
	return result
}

// removeDuplicateUUIDs removes duplicate UUIDs from a slice
func removeDuplicateUUIDs(uuids []uuid.UUID) []uuid.UUID {
	seen := make(map[uuid.UUID]bool)
	result := make([]uuid.UUID, 0, len(uuids))
	
	for _, id := range uuids {
		if !seen[id] {
			seen[id] = true
			result = append(result, id)
		}
	}
	
	return result
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

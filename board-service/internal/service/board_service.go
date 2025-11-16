package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"project-board-api/internal/domain"
	"project-board-api/internal/dto"
	"project-board-api/internal/repository"
	"project-board-api/internal/response"
)

// BoardService defines the interface for board business logic
type BoardService interface {
	CreateBoard(ctx context.Context, req *dto.CreateBoardRequest) (*dto.BoardResponse, error)
	GetBoard(ctx context.Context, boardID uuid.UUID) (*dto.BoardDetailResponse, error)
	GetBoardsByProject(ctx context.Context, projectID uuid.UUID) ([]*dto.BoardResponse, error)
	UpdateBoard(ctx context.Context, boardID uuid.UUID, req *dto.UpdateBoardRequest) (*dto.BoardResponse, error)
	DeleteBoard(ctx context.Context, boardID uuid.UUID) error
}

// boardServiceImpl is the implementation of BoardService
type boardServiceImpl struct {
	boardRepo   repository.BoardRepository
	projectRepo repository.ProjectRepository
}

// NewBoardService creates a new instance of BoardService
func NewBoardService(boardRepo repository.BoardRepository, projectRepo repository.ProjectRepository) BoardService {
	return &boardServiceImpl{
		boardRepo:   boardRepo,
		projectRepo: projectRepo,
	}
}

// CreateBoard creates a new board
func (s *boardServiceImpl) CreateBoard(ctx context.Context, req *dto.CreateBoardRequest) (*dto.BoardResponse, error) {
	// Verify project exists
	_, err := s.projectRepo.FindByID(ctx, req.ProjectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, response.NewAppError(response.ErrCodeNotFound, "Project not found", "")
		}
		return nil, response.NewAppError(response.ErrCodeInternal, "Failed to verify project", err.Error())
	}

	// Create domain model from request
	board := &domain.Board{
		ProjectID:  req.ProjectID,
		Title:      req.Title,
		Content:    req.Content,
		Stage:      domain.Stage(req.Stage),
		Importance: domain.Importance(req.Importance),
		Role:       domain.Role(req.Role),
	}

	// Save to repository
	if err := s.boardRepo.Create(ctx, board); err != nil {
		return nil, response.NewAppError(response.ErrCodeInternal, "Failed to create board", err.Error())
	}

	// Convert to response DTO
	return s.toBoardResponse(board), nil
}

// GetBoard retrieves a board by ID with participants and comments
func (s *boardServiceImpl) GetBoard(ctx context.Context, boardID uuid.UUID) (*dto.BoardDetailResponse, error) {
	// Fetch board from repository
	board, err := s.boardRepo.FindByID(ctx, boardID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, response.NewAppError(response.ErrCodeNotFound, "Board not found", "")
		}
		return nil, response.NewAppError(response.ErrCodeInternal, "Failed to fetch board", err.Error())
	}

	// Convert to detailed response DTO
	return s.toBoardDetailResponse(board), nil
}

// GetBoardsByProject retrieves all boards for a project
func (s *boardServiceImpl) GetBoardsByProject(ctx context.Context, projectID uuid.UUID) ([]*dto.BoardResponse, error) {
	// Verify project exists
	_, err := s.projectRepo.FindByID(ctx, projectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, response.NewAppError(response.ErrCodeNotFound, "Project not found", "")
		}
		return nil, response.NewAppError(response.ErrCodeInternal, "Failed to verify project", err.Error())
	}

	// Fetch boards from repository
	boards, err := s.boardRepo.FindByProjectID(ctx, projectID)
	if err != nil {
		return nil, response.NewAppError(response.ErrCodeInternal, "Failed to fetch boards", err.Error())
	}

	// Convert to response DTOs
	responses := make([]*dto.BoardResponse, len(boards))
	for i, board := range boards {
		responses[i] = s.toBoardResponse(board)
	}

	return responses, nil
}

// UpdateBoard updates a board's attributes
func (s *boardServiceImpl) UpdateBoard(ctx context.Context, boardID uuid.UUID, req *dto.UpdateBoardRequest) (*dto.BoardResponse, error) {
	// Fetch existing board
	board, err := s.boardRepo.FindByID(ctx, boardID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, response.NewAppError(response.ErrCodeNotFound, "Board not found", "")
		}
		return nil, response.NewAppError(response.ErrCodeInternal, "Failed to fetch board", err.Error())
	}

	// Update fields if provided
	if req.Title != nil {
		board.Title = *req.Title
	}
	if req.Content != nil {
		board.Content = *req.Content
	}
	if req.Stage != nil {
		board.Stage = domain.Stage(*req.Stage)
	}
	if req.Importance != nil {
		board.Importance = domain.Importance(*req.Importance)
	}
	if req.Role != nil {
		board.Role = domain.Role(*req.Role)
	}

	// Save updated board
	if err := s.boardRepo.Update(ctx, board); err != nil {
		return nil, response.NewAppError(response.ErrCodeInternal, "Failed to update board", err.Error())
	}

	// Convert to response DTO
	return s.toBoardResponse(board), nil
}

// DeleteBoard soft deletes a board
func (s *boardServiceImpl) DeleteBoard(ctx context.Context, boardID uuid.UUID) error {
	// Verify board exists
	_, err := s.boardRepo.FindByID(ctx, boardID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.NewAppError(response.ErrCodeNotFound, "Board not found", "")
		}
		return response.NewAppError(response.ErrCodeInternal, "Failed to verify board", err.Error())
	}

	// Delete board
	if err := s.boardRepo.Delete(ctx, boardID); err != nil {
		return response.NewAppError(response.ErrCodeInternal, "Failed to delete board", err.Error())
	}

	return nil
}

// toBoardResponse converts domain.Board to dto.BoardResponse
func (s *boardServiceImpl) toBoardResponse(board *domain.Board) *dto.BoardResponse {
	return &dto.BoardResponse{
		ID:         board.ID,
		ProjectID:  board.ProjectID,
		Title:      board.Title,
		Content:    board.Content,
		Stage:      string(board.Stage),
		Importance: string(board.Importance),
		Role:       string(board.Role),
		CreatedAt:  board.CreatedAt,
		UpdatedAt:  board.UpdatedAt,
	}
}

// toBoardDetailResponse converts domain.Board to dto.BoardDetailResponse
func (s *boardServiceImpl) toBoardDetailResponse(board *domain.Board) *dto.BoardDetailResponse {
	// Convert participants
	participants := make([]dto.ParticipantResponse, len(board.Participants))
	for i, p := range board.Participants {
		participants[i] = dto.ParticipantResponse{
			ID:        p.ID,
			BoardID:   p.BoardID,
			UserID:    p.UserID,
			CreatedAt: p.CreatedAt,
		}
	}

	// Convert comments
	comments := make([]dto.CommentResponse, len(board.Comments))
	for i, c := range board.Comments {
		comments[i] = dto.CommentResponse{
			CommentID: c.ID,
			BoardID:   c.BoardID,
			UserID:    c.UserID,
			Content:   c.Content,
			CreatedAt: c.CreatedAt,
			UpdatedAt: c.UpdatedAt,
		}
	}

	return &dto.BoardDetailResponse{
		BoardResponse: *s.toBoardResponse(board),
		Participants:  participants,
		Comments:      comments,
	}
}

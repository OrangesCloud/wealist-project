package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"project-board-api/internal/domain"
	"project-board-api/internal/dto"
	"project-board-api/internal/metrics"
	"project-board-api/internal/repository"
	"project-board-api/internal/response"
)

// BoardService defines the interface for board business logic
type BoardService interface {
	CreateBoard(ctx context.Context, req *dto.CreateBoardRequest) (*dto.BoardResponse, error)
	GetBoard(ctx context.Context, boardID uuid.UUID) (*dto.BoardDetailResponse, error)
	GetBoardsByProject(ctx context.Context, projectID uuid.UUID, filters *dto.BoardFilters) ([]*dto.BoardResponse, error)
	UpdateBoard(ctx context.Context, boardID uuid.UUID, req *dto.UpdateBoardRequest) (*dto.BoardResponse, error)
	DeleteBoard(ctx context.Context, boardID uuid.UUID) error
}

// boardServiceImpl is the implementation of BoardService
type boardServiceImpl struct {
	boardRepo            repository.BoardRepository
	projectRepo          repository.ProjectRepository
	fieldOptionRepo      repository.FieldOptionRepository
	participantRepo      repository.ParticipantRepository
	fieldOptionConverter FieldOptionConverter
	metrics              *metrics.Metrics
	logger               *zap.Logger
}

// FieldOptionConverter handles conversion between field option values and IDs
type FieldOptionConverter interface {
	ConvertValuesToIDs(ctx context.Context, projectID uuid.UUID, customFields map[string]interface{}) (map[string]interface{}, error)
	ConvertIDsToValues(ctx context.Context, customFields map[string]interface{}) (map[string]interface{}, error)
	ConvertIDsToValuesBatch(ctx context.Context, boards []*domain.Board) error
}

// NewBoardService creates a new instance of BoardService
func NewBoardService(
	boardRepo repository.BoardRepository,
	projectRepo repository.ProjectRepository,
	fieldOptionRepo repository.FieldOptionRepository,
	participantRepo repository.ParticipantRepository,
	fieldOptionConverter FieldOptionConverter,
	m *metrics.Metrics,
	logger *zap.Logger,
) BoardService {
	return &boardServiceImpl{
		boardRepo:            boardRepo,
		projectRepo:          projectRepo,
		fieldOptionRepo:      fieldOptionRepo,
		participantRepo:      participantRepo,
		fieldOptionConverter: fieldOptionConverter,
		metrics:              m,
		logger:               logger,
	}
}

// CreateBoard creates a new board
func (s *boardServiceImpl) CreateBoard(ctx context.Context, req *dto.CreateBoardRequest) (*dto.BoardResponse, error) {
	// Extract user_id from context (set by auth middleware as uuid.UUID)
	authorID, exists := ctx.Value("user_id").(uuid.UUID)
	if !exists {
		return nil, response.NewAppError(response.ErrCodeUnauthorized, "User ID not found in context", "")
	}

	// Validate date range
	if err := validateDateRange(req.StartDate, req.DueDate); err != nil {
		return nil, err
	}

	// Verify project exists
	_, err := s.projectRepo.FindByID(ctx, req.ProjectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, response.NewAppError(response.ErrCodeNotFound, "Project not found", "")
		}
		return nil, response.NewAppError(response.ErrCodeInternal, "Failed to verify project", err.Error())
	}

	// Convert CustomFields from values to IDs, then to datatypes.JSON
	var customFieldsJSON datatypes.JSON
	if req.CustomFields != nil {
		// Convert values to IDs
		convertedFields, err := s.fieldOptionConverter.ConvertValuesToIDs(ctx, req.ProjectID, req.CustomFields)
		if err != nil {
			return nil, response.NewAppError(response.ErrCodeValidation, "Invalid custom field values", err.Error())
		}
		
		jsonBytes, err := json.Marshal(convertedFields)
		if err != nil {
			return nil, response.NewAppError(response.ErrCodeInternal, "Failed to marshal custom fields", err.Error())
		}
		customFieldsJSON = jsonBytes
	}

	// Set assigneeID: use provided value, or default to authorID if not provided
	assigneeID := req.AssigneeID
	if assigneeID == nil {
		assigneeID = &authorID
	}

	// Create domain model from request with AuthorID
	board := &domain.Board{
		ProjectID:    req.ProjectID,
		AuthorID:     authorID,
		Title:        req.Title,
		Content:      req.Content,
		CustomFields: customFieldsJSON,
		AssigneeID:   assigneeID,
		StartDate:    req.StartDate,
		DueDate:      req.DueDate,
	}

	// Save to repository
	if err := s.boardRepo.Create(ctx, board); err != nil {
		return nil, response.NewAppError(response.ErrCodeInternal, "Failed to create board", err.Error())
	}

	// Increment board creation metric
	if s.metrics != nil {
		s.metrics.IncrementBoardCreated()
	}

	// Add participants if provided
	if len(req.Participants) > 0 {
		successCount, err := s.addParticipantsInternal(ctx, board.ID, req.Participants)
		if err != nil {
			s.logger.Warn("Error occurred while adding participants during board creation",
				zap.String("board_id", board.ID.String()),
				zap.Int("success_count", successCount),
				zap.Error(err))
		}
		
		// Reload board with participants to include them in response
		reloadedBoard, err := s.boardRepo.FindByID(ctx, board.ID)
		if err != nil {
			s.logger.Warn("Failed to reload board with participants",
				zap.String("board_id", board.ID.String()),
				zap.Error(err))
			// Continue with original board if reload fails
		} else {
			board = reloadedBoard
		}
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

	// Convert IDs to values in customFields
	if err := s.convertBoardCustomFieldsToValues(ctx, board); err != nil {
		return nil, response.NewAppError(response.ErrCodeInternal, "Failed to convert custom fields", err.Error())
	}

	// Convert to detailed response DTO
	return s.toBoardDetailResponse(board), nil
}

// GetBoardsByProject retrieves all boards for a project with optional filters
func (s *boardServiceImpl) GetBoardsByProject(ctx context.Context, projectID uuid.UUID, filters *dto.BoardFilters) ([]*dto.BoardResponse, error) {
	// Verify project exists
	_, err := s.projectRepo.FindByID(ctx, projectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, response.NewAppError(response.ErrCodeNotFound, "Project not found", "")
		}
		return nil, response.NewAppError(response.ErrCodeInternal, "Failed to verify project", err.Error())
	}

	// Prepare filter parameter for repository
	var filterParam interface{}
	if filters != nil && filters.CustomFields != nil {
		filterParam = filters.CustomFields
	}

	// Fetch boards from repository with filters
	boards, err := s.boardRepo.FindByProjectID(ctx, projectID, filterParam)
	if err != nil {
		return nil, response.NewAppError(response.ErrCodeInternal, "Failed to fetch boards", err.Error())
	}

	// Convert IDs to values in batch for all boards
	if err := s.fieldOptionConverter.ConvertIDsToValuesBatch(ctx, boards); err != nil {
		return nil, response.NewAppError(response.ErrCodeInternal, "Failed to convert custom fields", err.Error())
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

	// Determine the effective start and due dates for validation
	effectiveStartDate := board.StartDate
	effectiveDueDate := board.DueDate
	
	if req.StartDate != nil {
		effectiveStartDate = req.StartDate
	}
	if req.DueDate != nil {
		effectiveDueDate = req.DueDate
	}
	
	// Validate date range with effective dates
	if err := validateDateRange(effectiveStartDate, effectiveDueDate); err != nil {
		return nil, err
	}

	// Update fields if provided
	if req.Title != nil {
		board.Title = *req.Title
	}
	if req.Content != nil {
		board.Content = *req.Content
	}
	if req.CustomFields != nil {
		// Convert values to IDs
		convertedFields, err := s.fieldOptionConverter.ConvertValuesToIDs(ctx, board.ProjectID, *req.CustomFields)
		if err != nil {
			return nil, response.NewAppError(response.ErrCodeValidation, "Invalid custom field values", err.Error())
		}
		
		// Convert CustomFields to datatypes.JSON
		jsonBytes, err := json.Marshal(convertedFields)
		if err != nil {
			return nil, response.NewAppError(response.ErrCodeInternal, "Failed to marshal custom fields", err.Error())
		}
		board.CustomFields = jsonBytes
	}
	if req.AssigneeID != nil {
		board.AssigneeID = req.AssigneeID
	}
	if req.StartDate != nil {
		board.StartDate = req.StartDate
	}
	if req.DueDate != nil {
		board.DueDate = req.DueDate
	}

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

// convertBoardCustomFieldsToValues converts a single board's customFields from IDs to values
func (s *boardServiceImpl) convertBoardCustomFieldsToValues(ctx context.Context, board *domain.Board) error {
	if board.CustomFields == nil || len(board.CustomFields) == 0 {
		return nil
	}

	var customFields map[string]interface{}
	if err := json.Unmarshal(board.CustomFields, &customFields); err != nil {
		return err
	}

	convertedFields, err := s.fieldOptionConverter.ConvertIDsToValues(ctx, customFields)
	if err != nil {
		return err
	}

	jsonBytes, err := json.Marshal(convertedFields)
	if err != nil {
		return err
	}

	board.CustomFields = jsonBytes
	return nil
}

// toBoardResponse converts domain.Board to dto.BoardResponse
func (s *boardServiceImpl) toBoardResponse(board *domain.Board) *dto.BoardResponse {
	// Convert datatypes.JSON to map[string]interface{}
	var customFields map[string]interface{}
	if len(board.CustomFields) > 0 {
		_ = json.Unmarshal(board.CustomFields, &customFields)
	}
	
	// Extract participant IDs from board participants
	participantIDs := make([]uuid.UUID, 0, len(board.Participants))
	for _, p := range board.Participants {
		participantIDs = append(participantIDs, p.UserID)
	}
	
	// Convert attachments to response DTOs
	attachments := make([]dto.AttachmentResponse, 0, len(board.Attachments))
	for _, a := range board.Attachments {
		attachments = append(attachments, dto.AttachmentResponse{
			ID:          a.ID,
			FileName:    a.FileName,
			FileURL:     a.FileURL,
			FileSize:    a.FileSize,
			ContentType: a.ContentType,
			UploadedBy:  a.UploadedBy,
			UploadedAt:  a.CreatedAt,
		})
	}
	
	return &dto.BoardResponse{
		ID:             board.ID,
		ProjectID:      board.ProjectID,
		AuthorID:       board.AuthorID,
		AssigneeID:     board.AssigneeID,
		Title:          board.Title,
		Content:        board.Content,
		CustomFields:   customFields,
		StartDate:      board.StartDate,
		DueDate:        board.DueDate,
		ParticipantIDs: participantIDs,
		Attachments:    attachments,
		CreatedAt:      board.CreatedAt,
		UpdatedAt:      board.UpdatedAt,
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

// validateDateRange validates that startDate is not after dueDate
func validateDateRange(startDate, dueDate *time.Time) error {
	if startDate != nil && dueDate != nil {
		if startDate.After(*dueDate) {
			return response.NewAppError(response.ErrCodeValidation, "Start date cannot be after due date", "")
		}
	}
	return nil
}

// addParticipantsInternal is an internal helper to add participants during board creation
// It does not verify board existence (assumes board was just created)
// Returns the number of successfully added participants and any errors
func (s *boardServiceImpl) addParticipantsInternal(ctx context.Context, boardID uuid.UUID, userIDs []uuid.UUID) (int, error) {
	// Remove duplicates from the user IDs
	uniqueUserIDs := removeDuplicateUUIDs(userIDs)
	
	successCount := 0
	var failedUserIDs []uuid.UUID
	
	// Process each participant individually
	for _, userID := range uniqueUserIDs {
		// Check if participant already exists
		existing, err := s.participantRepo.FindByBoardAndUser(ctx, boardID, userID)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Warn("Failed to check existing participant",
				zap.String("board_id", boardID.String()),
				zap.String("user_id", userID.String()),
				zap.Error(err))
			failedUserIDs = append(failedUserIDs, userID)
			continue
		}
		if existing != nil {
			// Participant already exists, skip
			continue
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
				// Participant already exists, skip
				continue
			}
			s.logger.Warn("Failed to add participant",
				zap.String("board_id", boardID.String()),
				zap.String("user_id", userID.String()),
				zap.Error(err))
			failedUserIDs = append(failedUserIDs, userID)
			continue
		}
		
		successCount++
	}
	
	// Log summary if there were failures
	if len(failedUserIDs) > 0 {
		s.logger.Warn("Some participants failed to be added during board creation",
			zap.String("board_id", boardID.String()),
			zap.Int("success_count", successCount),
			zap.Int("failed_count", len(failedUserIDs)),
			zap.Any("failed_user_ids", failedUserIDs))
	}
	
	return successCount, nil
}

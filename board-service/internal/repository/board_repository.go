package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"project-board-api/internal/domain"
)

// BoardRepository defines the interface for board data access
type BoardRepository interface {
	Create(ctx context.Context, board *domain.Board) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Board, error)
	FindByProjectID(ctx context.Context, projectID uuid.UUID, filters interface{}) ([]*domain.Board, error)
	Update(ctx context.Context, board *domain.Board) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// boardRepositoryImpl is the GORM implementation of BoardRepository
type boardRepositoryImpl struct {
	db *gorm.DB
}

// NewBoardRepository creates a new instance of BoardRepository
func NewBoardRepository(db *gorm.DB) BoardRepository {
	return &boardRepositoryImpl{db: db}
}

// Create creates a new board
func (r *boardRepositoryImpl) Create(ctx context.Context, board *domain.Board) error {
	if err := r.db.WithContext(ctx).Create(board).Error; err != nil {
		return err
	}
	return nil
}

// FindByID finds a board by ID with preloaded participants and comments
func (r *boardRepositoryImpl) FindByID(ctx context.Context, id uuid.UUID) (*domain.Board, error) {
	var board domain.Board
	if err := r.db.WithContext(ctx).
		Preload("Participants").
		Preload("Comments").
		Where("id = ?", id).
		First(&board).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, err
	}
	return &board, nil
}

// FindByProjectID finds all boards by project ID with optional filters
func (r *boardRepositoryImpl) FindByProjectID(ctx context.Context, projectID uuid.UUID, filters interface{}) ([]*domain.Board, error) {
	var boards []*domain.Board
	
	// Start building the query
	query := r.db.WithContext(ctx).Where("project_id = ?", projectID)
	
	// Apply filters if provided
	if filters != nil {
		// Type assertion to get the filters struct
		if boardFilters, ok := filters.(interface {
			GetStage() string
			GetRole() string
			GetImportance() string
		}); ok {
			// Apply stage filter
			if stage := boardFilters.GetStage(); stage != "" {
				query = query.Where("stage = ?", stage)
			}
			
			// Apply role filter
			if role := boardFilters.GetRole(); role != "" {
				query = query.Where("role = ?", role)
			}
			
			// Apply importance filter
			if importance := boardFilters.GetImportance(); importance != "" {
				query = query.Where("importance = ?", importance)
			}
		}
	}
	
	// Execute the query
	if err := query.Find(&boards).Error; err != nil {
		return nil, err
	}
	
	return boards, nil
}

// Update updates a board
func (r *boardRepositoryImpl) Update(ctx context.Context, board *domain.Board) error {
	if err := r.db.WithContext(ctx).Save(board).Error; err != nil {
		return err
	}
	return nil
}

// Delete soft deletes a board
func (r *boardRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&domain.Board{}, id).Error; err != nil {
		return err
	}
	return nil
}

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

// CommentService defines the interface for comment business logic
type CommentService interface {
	CreateComment(ctx context.Context, req *dto.CreateCommentRequest) (*dto.CommentResponse, error)
	GetComments(ctx context.Context, boardID uuid.UUID) ([]*dto.CommentResponse, error)
	UpdateComment(ctx context.Context, commentID uuid.UUID, req *dto.UpdateCommentRequest) (*dto.CommentResponse, error)
	DeleteComment(ctx context.Context, commentID uuid.UUID) error
}

// commentServiceImpl is the implementation of CommentService
type commentServiceImpl struct {
	commentRepo repository.CommentRepository
	boardRepo   repository.BoardRepository
}

// NewCommentService creates a new instance of CommentService
func NewCommentService(commentRepo repository.CommentRepository, boardRepo repository.BoardRepository) CommentService {
	return &commentServiceImpl{
		commentRepo: commentRepo,
		boardRepo:   boardRepo,
	}
}

// CreateComment creates a new comment on a board
func (s *commentServiceImpl) CreateComment(ctx context.Context, req *dto.CreateCommentRequest) (*dto.CommentResponse, error) {
	// Verify board exists
	_, err := s.boardRepo.FindByID(ctx, req.BoardID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, response.NewAppError(response.ErrCodeNotFound, "Board not found", "")
		}
		return nil, response.NewAppError(response.ErrCodeInternal, "Failed to verify board", err.Error())
	}

	// TODO: Get UserID from context (authentication middleware)
	// For now, using a placeholder UUID
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000000")
	
	// Create domain model from request
	comment := &domain.Comment{
		BoardID: req.BoardID,
		UserID:  userID,
		Content: req.Content,
	}

	// Save to repository
	if err := s.commentRepo.Create(ctx, comment); err != nil {
		return nil, response.NewAppError(response.ErrCodeInternal, "Failed to create comment", err.Error())
	}

	// Convert to response DTO
	return s.toCommentResponse(comment), nil
}

// GetComments retrieves all comments for a board
func (s *commentServiceImpl) GetComments(ctx context.Context, boardID uuid.UUID) ([]*dto.CommentResponse, error) {
	// Verify board exists
	_, err := s.boardRepo.FindByID(ctx, boardID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, response.NewAppError(response.ErrCodeNotFound, "Board not found", "")
		}
		return nil, response.NewAppError(response.ErrCodeInternal, "Failed to verify board", err.Error())
	}

	// Fetch comments from repository
	comments, err := s.commentRepo.FindByBoardID(ctx, boardID)
	if err != nil {
		return nil, response.NewAppError(response.ErrCodeInternal, "Failed to fetch comments", err.Error())
	}

	// Convert to response DTOs
	responses := make([]*dto.CommentResponse, len(comments))
	for i, comment := range comments {
		responses[i] = s.toCommentResponse(comment)
	}

	return responses, nil
}

// UpdateComment updates a comment's content
func (s *commentServiceImpl) UpdateComment(ctx context.Context, commentID uuid.UUID, req *dto.UpdateCommentRequest) (*dto.CommentResponse, error) {
	// Fetch existing comment
	comment, err := s.commentRepo.FindByID(ctx, commentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, response.NewAppError(response.ErrCodeNotFound, "Comment not found", "")
		}
		return nil, response.NewAppError(response.ErrCodeInternal, "Failed to fetch comment", err.Error())
	}

	// Update content
	comment.Content = req.Content

	// Save updated comment
	if err := s.commentRepo.Update(ctx, comment); err != nil {
		return nil, response.NewAppError(response.ErrCodeInternal, "Failed to update comment", err.Error())
	}

	// Convert to response DTO
	return s.toCommentResponse(comment), nil
}

// DeleteComment soft deletes a comment
func (s *commentServiceImpl) DeleteComment(ctx context.Context, commentID uuid.UUID) error {
	// Verify comment exists
	_, err := s.commentRepo.FindByID(ctx, commentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.NewAppError(response.ErrCodeNotFound, "Comment not found", "")
		}
		return response.NewAppError(response.ErrCodeInternal, "Failed to verify comment", err.Error())
	}

	// Delete comment
	if err := s.commentRepo.Delete(ctx, commentID); err != nil {
		return response.NewAppError(response.ErrCodeInternal, "Failed to delete comment", err.Error())
	}

	return nil
}

// toCommentResponse converts domain.Comment to dto.CommentResponse
func (s *commentServiceImpl) toCommentResponse(comment *domain.Comment) *dto.CommentResponse {
	return &dto.CommentResponse{
		CommentID: comment.ID,
		BoardID:   comment.BoardID,
		UserID:    comment.UserID,
		Content:   comment.Content,
		CreatedAt: comment.CreatedAt,
		UpdatedAt: comment.UpdatedAt,
	}
}

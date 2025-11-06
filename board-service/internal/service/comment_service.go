package service

import (
	"board-service/internal/apperrors"
	"board-service/internal/client"
	"board-service/internal/domain"
	"board-service/internal/dto"
	"board-service/internal/repository"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// CommentService defines the interface for comment business logic.
type CommentService interface {
	CreateComment(ctx context.Context, req dto.CreateCommentRequest, userID uuid.UUID) (*dto.CommentResponse, error)
	GetCommentsByBoardID(ctx context.Context, boardID uuid.UUID, userID uuid.UUID) ([]dto.CommentResponse, error)
	UpdateComment(ctx context.Context, commentID uuid.UUID, req dto.UpdateCommentRequest, userID uuid.UUID) (*dto.CommentResponse, error)
	DeleteComment(ctx context.Context, commentID uuid.UUID, userID uuid.UUID) error
}

type commentService struct {
	commentRepo repository.CommentRepository
	boardRepo  repository.BoardRepository
	projectRepo repository.ProjectRepository
	userClient  client.UserClient
	logger      *zap.Logger
	db          *gorm.DB
}

// NewCommentService creates a new instance of CommentService.
func NewCommentService(cr repository.CommentRepository, kr repository.BoardRepository, pr repository.ProjectRepository, uc client.UserClient, l *zap.Logger, db *gorm.DB) CommentService {
	return &commentService{
		commentRepo: cr,
		boardRepo:  kr,
		projectRepo: pr,
		userClient:  uc,
		logger:      l,
		db:          db,
	}
}

// CreateComment creates a new comment on a board.
func (s *commentService) CreateComment(ctx context.Context, req dto.CreateCommentRequest, userID uuid.UUID) (*dto.CommentResponse, error) {
	board, err := s.boardRepo.FindByID(req.BoardID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.ErrCodeNotFound, fmt.Sprintf("board with id %s not found", req.BoardID), 404)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "failed to find board", 500)
	}

	_, err = s.projectRepo.FindMemberByUserAndProject(userID, board.ProjectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.ErrCodeForbidden, "user is not a member of the project", 403)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "failed to check project membership", 500)
	}

	comment := &domain.Comment{
		BoardID: req.BoardID,
		UserID:   userID,
		Content:  req.Content,
	}

	if err := s.commentRepo.Create(comment); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "failed to create comment", 500)
	}

	user, err := s.userClient.GetSimpleUser(userID.String())
	if err != nil {
		// Log the error but proceed, as the comment is already created.
		s.logger.Error("Failed to get user info for comment", zap.Error(err), zap.String("userID", userID.String()))
		user = &client.SimpleUser{Name: "Unknown User", AvatarURL: ""} // Fallback user
	}

	return &dto.CommentResponse{
		ID:        comment.ID,
		UserID:    comment.UserID,
		UserName:  user.Name,
		UserAvatar: user.AvatarURL,
		Content:   comment.Content,
		CreatedAt: comment.CreatedAt,
		UpdatedAt: comment.UpdatedAt,
	}, nil
}

// GetCommentsByBoardID retrieves all comments for a given board.
func (s *commentService) GetCommentsByBoardID(ctx context.Context, boardID uuid.UUID, userID uuid.UUID) ([]dto.CommentResponse, error) {
	board, err := s.boardRepo.FindByID(boardID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.ErrCodeNotFound, fmt.Sprintf("board with id %s not found", boardID), 404)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "failed to find board", 500)
	}

	_, err = s.projectRepo.FindMemberByUserAndProject(userID, board.ProjectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.ErrCodeForbidden, "user is not a member of the project", 403)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "failed to check project membership", 500)
	}

	comments, err := s.commentRepo.FindByBoardID(boardID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "failed to get comments", 500)
	}

	// Batch fetch user info
	userIDs := make([]string, 0, len(comments))
	for _, c := range comments {
		userIDs = append(userIDs, c.UserID.String())
	}

	users, err := s.userClient.GetSimpleUsers(userIDs)
	userMap := make(map[string]client.SimpleUser)
	if err == nil {
		for _, u := range users {
			userMap[u.ID] = u
		}
	}

	responses := make([]dto.CommentResponse, len(comments))
	for i, c := range comments {
		user, ok := userMap[c.UserID.String()]
		if !ok {
			user = client.SimpleUser{Name: "Unknown User", AvatarURL: ""}
		}
		responses[i] = dto.CommentResponse{
			ID:        c.ID,
			UserID:    c.UserID,
			UserName:  user.Name,
			UserAvatar: user.AvatarURL,
			Content:   c.Content,
			CreatedAt: c.CreatedAt,
			UpdatedAt: c.UpdatedAt,
		}
	}

	return responses, nil
}

// UpdateComment updates an existing comment.
func (s *commentService) UpdateComment(ctx context.Context, commentID uuid.UUID, req dto.UpdateCommentRequest, userID uuid.UUID) (*dto.CommentResponse, error) {
	comment, err := s.commentRepo.FindByID(commentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.ErrCodeNotFound, fmt.Sprintf("comment with id %s not found", commentID), 404)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "failed to find comment", 500)
	}

	if comment.UserID != userID {
		return nil, apperrors.New(apperrors.ErrCodeForbidden, "user does not have permission to update this comment", 403)
	}

	comment.Content = req.Content
	if err := s.commentRepo.Update(comment); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "failed to update comment", 500)
	}

	user, err := s.userClient.GetSimpleUser(userID.String())
	if err != nil {
		s.logger.Error("Failed to get user info for comment", zap.Error(err), zap.String("userID", userID.String()))
		user = &client.SimpleUser{Name: "Unknown User", AvatarURL: ""} // Fallback user
	}

	return &dto.CommentResponse{
		ID:        comment.ID,
		UserID:    comment.UserID,
		UserName:  user.Name,
		UserAvatar: user.AvatarURL,
		Content:   comment.Content,
		CreatedAt: comment.CreatedAt,
		UpdatedAt: comment.UpdatedAt,
	}, nil
}

// DeleteComment deletes a comment.
func (s *commentService) DeleteComment(ctx context.Context, commentID uuid.UUID, userID uuid.UUID) error {
	comment, err := s.commentRepo.FindByID(commentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.New(apperrors.ErrCodeNotFound, fmt.Sprintf("comment with id %s not found", commentID), 404)
		}
		return apperrors.Wrap(err, apperrors.ErrCodeInternalServer, "failed to find comment", 500)
	}

	if comment.UserID != userID {
		// Also allow project owner/admin to delete?
		// For now, only the author can delete.
		return apperrors.New(apperrors.ErrCodeForbidden, "user does not have permission to delete this comment", 403)
	}

	return s.commentRepo.Delete(comment.ID)
}

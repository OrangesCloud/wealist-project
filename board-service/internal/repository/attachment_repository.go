package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"project-board-api/internal/domain"
)

// AttachmentRepository defines the interface for attachment data access
type AttachmentRepository interface {
	Create(ctx context.Context, attachment *domain.Attachment) error
	FindByEntityID(ctx context.Context, entityType domain.EntityType, entityID uuid.UUID) ([]*domain.Attachment, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// attachmentRepositoryImpl is the GORM implementation of AttachmentRepository
type attachmentRepositoryImpl struct {
	db *gorm.DB
}

// NewAttachmentRepository creates a new instance of AttachmentRepository
func NewAttachmentRepository(db *gorm.DB) AttachmentRepository {
	return &attachmentRepositoryImpl{db: db}
}

// Create creates a new attachment
func (r *attachmentRepositoryImpl) Create(ctx context.Context, attachment *domain.Attachment) error {
	if err := r.db.WithContext(ctx).Create(attachment).Error; err != nil {
		return err
	}
	return nil
}

// FindByEntityID finds all attachments by entity type and entity ID
func (r *attachmentRepositoryImpl) FindByEntityID(ctx context.Context, entityType domain.EntityType, entityID uuid.UUID) ([]*domain.Attachment, error) {
	var attachments []*domain.Attachment
	if err := r.db.WithContext(ctx).
		Where("entity_type = ? AND entity_id = ?", entityType, entityID).
		Order("created_at DESC").
		Find(&attachments).Error; err != nil {
		return nil, err
	}
	return attachments, nil
}

// Delete soft deletes an attachment by ID
func (r *attachmentRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&domain.Attachment{}, id).Error; err != nil {
		return err
	}
	return nil
}

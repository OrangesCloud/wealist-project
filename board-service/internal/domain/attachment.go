package domain

import "github.com/google/uuid"

// EntityType represents the type of entity an attachment is associated with
type EntityType string

const (
	EntityTypeBoard   EntityType = "BOARD"
	EntityTypeProject EntityType = "PROJECT"
	EntityTypeComment EntityType = "COMMENT"
)

// Attachment represents a file attachment associated with a board or project
type Attachment struct {
	BaseModel
	EntityType  EntityType `gorm:"type:varchar(50);not null;index:idx_attachments_entity,priority:1" json:"entity_type"`
	EntityID    uuid.UUID  `gorm:"type:uuid;not null;index:idx_attachments_entity,priority:2;index:idx_attachments_entity_id" json:"entity_id"`
	FileName    string     `gorm:"type:varchar(255);not null" json:"file_name"`
	FileURL     string     `gorm:"type:text;not null" json:"file_url"`
	FileSize    int64      `gorm:"not null" json:"file_size"`
	ContentType string     `gorm:"type:varchar(100);not null" json:"content_type"`
	UploadedBy  uuid.UUID  `gorm:"type:uuid;not null;index:idx_attachments_uploaded_by" json:"uploaded_by"`
}

// TableName specifies the table name for Attachment
func (Attachment) TableName() string {
	return "attachments"
}

package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// Board represents a work board entity within a project
type Board struct {
	BaseModel
	ProjectID    uuid.UUID              `gorm:"type:uuid;not null;index:idx_boards_project_id" json:"project_id"`
	AuthorID     uuid.UUID              `gorm:"type:uuid;not null;index:idx_boards_author_id" json:"author_id"`
	AssigneeID   *uuid.UUID             `gorm:"type:uuid;index:idx_boards_assignee_id" json:"assignee_id"`
	Title        string                 `gorm:"type:varchar(255);not null" json:"title"`
	Content      string                 `gorm:"type:text" json:"content"`
	CustomFields datatypes.JSON `gorm:"type:jsonb" json:"custom_fields"`
	StartDate    *time.Time             `gorm:"type:timestamp;index:idx_boards_start_date" json:"start_date"`
	DueDate      *time.Time             `gorm:"type:timestamp;index:idx_boards_due_date" json:"due_date"`
	Project      Project                `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE" json:"project,omitempty"`
	Participants []Participant          `gorm:"foreignKey:BoardID;constraint:OnDelete:CASCADE" json:"participants,omitempty"`
	Comments     []Comment              `gorm:"foreignKey:BoardID;constraint:OnDelete:CASCADE" json:"comments,omitempty"`
	Attachments  []Attachment           `gorm:"foreignKey:EntityID;constraint:OnDelete:CASCADE" json:"attachments,omitempty"`
}

// TableName specifies the table name for Board
func (Board) TableName() string {
	return "boards"
}

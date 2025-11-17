package domain

import (
	"time"

	"github.com/google/uuid"
)

// Board represents a work board entity within a project
type Board struct {
	BaseModel
	ProjectID    uuid.UUID              `gorm:"type:uuid;not null;index" json:"project_id"`
	AuthorID     uuid.UUID              `gorm:"type:uuid;not null;index" json:"author_id"`
	AssigneeID   *uuid.UUID             `gorm:"type:uuid;index" json:"assignee_id"`
	Title        string                 `gorm:"type:varchar(255);not null" json:"title"`
	Content      string                 `gorm:"type:text" json:"content"`
	CustomFields map[string]interface{} `gorm:"type:jsonb" json:"custom_fields"`
	DueDate      *time.Time             `gorm:"type:timestamp" json:"due_date"`
	Project      Project                `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
	Participants []Participant          `gorm:"foreignKey:BoardID" json:"participants,omitempty"`
	Comments     []Comment              `gorm:"foreignKey:BoardID" json:"comments,omitempty"`
}

// TableName specifies the table name for Board
func (Board) TableName() string {
	return "boards"
}

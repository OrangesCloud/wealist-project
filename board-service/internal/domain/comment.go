package domain

import "github.com/google/uuid"

// Comment represents a comment on a board
type Comment struct {
	BaseModel
	BoardID uuid.UUID `gorm:"type:uuid;not null;index" json:"board_id"`
	UserID  uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	Content string    `gorm:"type:text;not null" json:"content"`
	Board   Board     `gorm:"foreignKey:BoardID" json:"board,omitempty"`
}

// TableName specifies the table name for Comment
func (Comment) TableName() string {
	return "comments"
}

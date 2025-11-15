package domain

import (
	"time"

	"github.com/google/uuid"
)

// Stage represents the progress status of a board
type Stage string

const (
	StageInProgress Stage = "in_progress"
	StagePending    Stage = "pending"
	StageApproved   Stage = "approved"
	StageReview     Stage = "review"
)

// Importance represents the priority level of a board
type Importance string

const (
	ImportanceUrgent Importance = "urgent"
	ImportanceNormal Importance = "normal"
)

// Role represents the role of the person responsible for the board
type Role string

const (
	RoleDeveloper Role = "developer"
	RolePlanner   Role = "planner"
)

// Board represents a work board entity within a project
type Board struct {
	BaseModel
	ProjectID    uuid.UUID     `gorm:"type:uuid;not null;index" json:"project_id"`
	AuthorID     uuid.UUID     `gorm:"type:uuid;not null;index" json:"author_id"`
	AssigneeID   *uuid.UUID    `gorm:"type:uuid;index" json:"assignee_id"`
	Title        string        `gorm:"type:varchar(255);not null" json:"title"`
	Content      string        `gorm:"type:text" json:"content"`
	Stage        Stage         `gorm:"type:varchar(50);not null" json:"stage"`
	Importance   Importance    `gorm:"type:varchar(50);not null" json:"importance"`
	Role         Role          `gorm:"type:varchar(50);not null" json:"role"`
	DueDate      *time.Time    `gorm:"type:timestamp" json:"due_date"`
	Project      Project       `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
	Participants []Participant `gorm:"foreignKey:BoardID" json:"participants,omitempty"`
	Comments     []Comment     `gorm:"foreignKey:BoardID" json:"comments,omitempty"`
}

// TableName specifies the table name for Board
func (Board) TableName() string {
	return "boards"
}

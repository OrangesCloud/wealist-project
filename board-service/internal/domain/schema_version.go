package domain

import "time"

// SchemaVersion tracks applied database migrations
type SchemaVersion struct {
	Version     string    `gorm:"type:varchar(50);primaryKey" json:"version"`
	Description string    `gorm:"type:text" json:"description"`
	AppliedAt   time.Time `gorm:"autoCreateTime" json:"applied_at"`
}

func (SchemaVersion) TableName() string {
	return "schema_versions"
}

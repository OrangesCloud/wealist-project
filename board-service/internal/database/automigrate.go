package database

import (
	"fmt"

	"gorm.io/gorm"
	"project-board-api/internal/domain"
)

// AutoMigrate runs GORM auto-migration for all domain models
// It automatically creates tables, indexes, and foreign key constraints
// based on the struct definitions in the domain package
func AutoMigrate(db *gorm.DB) error {
	// List of all domain models to migrate
	models := []interface{}{
		&domain.Project{},
		&domain.ProjectMember{},
		&domain.ProjectJoinRequest{},
		&domain.Board{},
		&domain.Participant{},
		&domain.Comment{},
		&domain.FieldOption{},
	}

	// Run auto-migration for all models
	if err := db.AutoMigrate(models...); err != nil {
		return fmt.Errorf("failed to run auto-migration: %w", err)
	}

	return nil
}

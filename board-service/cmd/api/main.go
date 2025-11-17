package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"project-board-api/internal/client"
	"project-board-api/internal/config"
	"project-board-api/internal/database"
	"project-board-api/internal/logger"
	"project-board-api/internal/router"

	_ "project-board-api/docs" // Swagger docs
)

// @title           Project Board Management API
// @version         1.0
// @description     프로젝트 보드 관리 시스템 API 서버입니다.
// @description     Board, Project, Comment, Participant 관리 기능을 제공합니다.

// @contact.name   API Support
// @contact.email  support@example.com

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @BasePath  /api
// @schemes   http https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT 토큰을 입력하세요. 형식: Bearer {token}

// Additional DTO schemas for documentation
// These DTOs are defined in the codebase and should be included in swagger definitions
// even if not directly used in current handler endpoints

// BoardFilters represents filter parameters for board queries
// @Description Board filtering parameters
// @Accept json
// @Produce json
// @Success 200 {object} dto.BoardFilters

// PaginatedBoardsResponse represents paginated board list response
// @Description Paginated boards response with metadata
// @Accept json
// @Produce json
// @Success 200 {object} dto.PaginatedBoardsResponse

// UpdateBoardFieldRequest represents request to update a single board field
// @Description Update a single board field (stage, importance, or role)
// @Accept json
// @Produce json
// @Success 200 {object} dto.UpdateBoardFieldRequest

func main() {
	// Load configuration
	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	log, err := logger.New(cfg.Logger.Level, cfg.Logger.OutputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync()

	log.Info("Starting application",
		zap.String("mode", cfg.Server.Mode),
		zap.String("port", cfg.Server.Port),
	)

	// Set Gin mode
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// Connect to database
	dbConfig := database.Config{
		DSN:             cfg.Database.GetDSN(),
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		MaxIdleConns:    cfg.Database.MaxIdleConns,
		ConnMaxLifetime: cfg.Database.ConnMaxLifetime,
	}

	db, err := database.New(dbConfig)
	if err != nil {
		log.Fatal("Failed to connect to database", zap.Error(err))
	}

	log.Info("Database connection established",
		zap.String("host", cfg.Database.Host),
		zap.String("database", cfg.Database.DBName),
	)

	// Run GORM auto-migration with retry logic
	log.Info("Running GORM auto-migration with retry logic")
	if err := database.SafeAutoMigrateWithRetry(db, log.Logger, 3); err != nil {
		log.Fatal("Failed to run auto-migration",
			zap.Error(err),
			zap.String("hint", "Check database connection and schema conflicts"),
		)
	}
	log.Info("Database schema migration completed successfully")

	// Log complete User API configuration for debugging
	log.Info("User API Configuration",
		zap.String("base_url", cfg.UserAPI.BaseURL),
		zap.Duration("timeout", cfg.UserAPI.Timeout),
		zap.String("config_source", "Loaded from config.yaml and environment variables"),
	)

	// Initialize User API client
	userClient := client.NewUserClient(
		cfg.UserAPI.BaseURL,
		cfg.UserAPI.Timeout,
		log.Logger,
	)

	log.Info("User API client initialized successfully",
		zap.String("base_url", cfg.UserAPI.BaseURL),
		zap.Duration("timeout", cfg.UserAPI.Timeout),
	)
	
	// Log example endpoint URLs for verification
	log.Info("User API endpoint examples (for debugging)",
		zap.String("validate_member", cfg.UserAPI.BaseURL+"/api/workspaces/{workspaceId}/validate-member/{userId}"),
		zap.String("get_user", cfg.UserAPI.BaseURL+"/api/users/{userId}"),
		zap.String("get_workspace_profile", cfg.UserAPI.BaseURL+"/api/profiles/workspace/{workspaceId}"),
		zap.String("get_workspace", cfg.UserAPI.BaseURL+"/api/workspaces/{workspaceId}"),
	)

	// Setup router with dependency injection
	routerConfig := router.Config{
		DB:         db,
		Logger:     log.Logger,
		JWTSecret:  cfg.JWT.Secret,
		UserClient: userClient,
	}

	r := router.Setup(routerConfig)

	// Create HTTP server
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Start server in a goroutine
	go func() {
		log.Info("Server starting", zap.String("address", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	log.Info("Server started successfully", zap.String("port", cfg.Server.Port))

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	// SIGINT (Ctrl+C) and SIGTERM (kill) signals
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Block until signal is received
	sig := <-quit
	log.Info("Shutdown signal received", zap.String("signal", sig.String()))

	// Create shutdown context with timeout
	shutdownTimeout := cfg.Server.ShutdownTimeout
	if shutdownTimeout == 0 {
		shutdownTimeout = 30 * time.Second // Default timeout
	}

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	// Attempt graceful shutdown
	log.Info("Shutting down server gracefully", zap.Duration("timeout", shutdownTimeout))

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("Server forced to shutdown", zap.Error(err))
	} else {
		log.Info("Server shutdown completed, all in-flight requests completed")
	}

	// Close database connection
	log.Info("Closing database connection")
	if err := database.Close(db); err != nil {
		log.Error("Failed to close database connection", zap.Error(err))
	} else {
		log.Info("Database connection closed successfully")
	}

	log.Info("Application stopped")
}

package router

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"project-board-api/internal/client"
	"project-board-api/internal/handler"
	"project-board-api/internal/middleware"
	"project-board-api/internal/repository"
	"project-board-api/internal/service"
)

// Config holds router configuration
type Config struct {
	DB         *gorm.DB
	Logger     *zap.Logger
	JWTSecret  string
	UserClient client.UserClient
}

// Setup initializes the router with all dependencies and routes
func Setup(cfg Config) *gin.Engine {
	// Create Gin router
	router := gin.New()

	// Apply global middleware chain
	router.Use(
		middleware.Recovery(cfg.Logger), // 1. Panic recovery
		middleware.RequestID(),          // 2. Request ID tracking
		middleware.Logger(cfg.Logger),   // 3. Request logging
		middleware.CORS(),               // 4. CORS configuration
	)

	// Initialize repositories
	projectRepo := repository.NewProjectRepository(cfg.DB)
	boardRepo := repository.NewBoardRepository(cfg.DB)
	participantRepo := repository.NewParticipantRepository(cfg.DB)
	commentRepo := repository.NewCommentRepository(cfg.DB)

	// Initialize services with repository dependencies
	projectService := service.NewProjectService(projectRepo, cfg.UserClient)
	boardService := service.NewBoardService(boardRepo, projectRepo)
	participantService := service.NewParticipantService(participantRepo, boardRepo)
	commentService := service.NewCommentService(commentRepo, boardRepo)

	// Initialize handlers with service dependencies
	projectHandler := handler.NewProjectHandler(projectService)
	boardHandler := handler.NewBoardHandler(boardService)
	participantHandler := handler.NewParticipantHandler(participantService)
	commentHandler := handler.NewCommentHandler(commentService)

	// Health check endpoint
	router.GET("/health", healthCheckHandler(cfg.DB))

	// Swagger documentation endpoint
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Setup API routes
	setupRoutes(router, cfg.JWTSecret, projectHandler, boardHandler, participantHandler, commentHandler)

	return router
}

// healthCheckHandler returns a handler for the health check endpoint
func healthCheckHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check database connection
		sqlDB, err := db.DB()
		if err != nil {
			c.JSON(500, gin.H{
				"status":   "unhealthy",
				"database": "error",
				"error":    err.Error(),
			})
			return
		}

		if err := sqlDB.Ping(); err != nil {
			c.JSON(500, gin.H{
				"status":   "unhealthy",
				"database": "disconnected",
				"error":    err.Error(),
			})
			return
		}

		c.JSON(200, gin.H{
			"status":   "healthy",
			"database": "connected",
		})
	}
}

// setupRoutes configures all API routes
func setupRoutes(
	router *gin.Engine,
	jwtSecret string,
	projectHandler *handler.ProjectHandler,
	boardHandler *handler.BoardHandler,
	participantHandler *handler.ParticipantHandler,
	commentHandler *handler.CommentHandler,
) {
	// API group
	api := router.Group("/api")
	{
		// Project routes
		projects := api.Group("/projects")
		{
			projects.POST("", projectHandler.CreateProject)
			projects.GET("/workspace/:workspaceId", projectHandler.GetProjectsByWorkspace)
			projects.GET("/workspace/:workspaceId/default", projectHandler.GetDefaultProject)
		}

		// Board routes
		boards := api.Group("/boards")
		{
			boards.POST("", boardHandler.CreateBoard)
			boards.GET("/:boardId", boardHandler.GetBoard)
			boards.GET("/project/:projectId", boardHandler.GetBoardsByProject)
			boards.PUT("/:boardId", boardHandler.UpdateBoard)
			boards.DELETE("/:boardId", boardHandler.DeleteBoard)
		}

		// Participant routes
		participants := api.Group("/participants")
		{
			participants.POST("", participantHandler.AddParticipant)
			participants.GET("/board/:boardId", participantHandler.GetParticipants)
			participants.DELETE("/board/:boardId/user/:userId", participantHandler.RemoveParticipant)
		}

		// Comment routes
		comments := api.Group("/comments")
		{
			comments.POST("", commentHandler.CreateComment)
			comments.GET("/board/:boardId", commentHandler.GetComments)
			comments.PUT("/:commentId", commentHandler.UpdateComment)
			comments.DELETE("/:commentId", commentHandler.DeleteComment)
		}
	}

	// Optional: Add authenticated routes group if needed
	// authenticated := api.Group("")
	// authenticated.Use(middleware.Auth(jwtSecret))
	// {
	//     // Add routes that require authentication here
	// }
}

package router

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"project-board-api/internal/client"
	"project-board-api/internal/converter"
	"project-board-api/internal/handler"

	"project-board-api/internal/middleware"
	"project-board-api/internal/repository"
	"project-board-api/internal/service"
)

type Config struct {
	DB        *gorm.DB
	Logger    *zap.Logger
	JWTSecret string
	// 💡 [수정] 통합된 UserClient 인터페이스를 사용합니다.
	UserClient         client.UserClient
	BasePath           string
	UserServiceBaseURL string
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
	fieldOptionRepo := repository.NewFieldOptionRepository(cfg.DB)

	// Initialize converters
	fieldOptionConverter := converter.NewFieldOptionConverter(fieldOptionRepo)

	// 💡 [핵심 추가] User Service Base URL을 사용하여 AuthClient 초기화
	// cfg.UserClient가 UserClient 역할을 하고 AuthClient 인터페이스를 구현하는 것으로 가정합니다.

	// Initialize services with repository dependencies
	projectService := service.NewProjectService(projectRepo, fieldOptionRepo, cfg.UserClient, cfg.Logger)
	boardService := service.NewBoardService(boardRepo, projectRepo, fieldOptionRepo, fieldOptionConverter)
	participantService := service.NewParticipantService(participantRepo, boardRepo)
	commentService := service.NewCommentService(commentRepo, boardRepo)
	fieldOptionService := service.NewFieldOptionService(fieldOptionRepo)
	projectMemberService := service.NewProjectMemberService(projectRepo, cfg.UserClient)
	projectJoinRequestService := service.NewProjectJoinRequestService(projectRepo, cfg.UserClient)

	// Initialize handlers with service dependencies
	projectHandler := handler.NewProjectHandler(projectService)
	boardHandler := handler.NewBoardHandler(boardService)
	participantHandler := handler.NewParticipantHandler(participantService)
	commentHandler := handler.NewCommentHandler(commentService)
	fieldOptionHandler := handler.NewFieldOptionHandler(fieldOptionService)
	projectMemberHandler := handler.NewProjectMemberHandler(projectMemberService)
	projectJoinRequestHandler := handler.NewProjectJoinRequestHandler(projectJoinRequestService)

	// 💡 [추가] WebSocket Handler 초기화 (AuthClient 주입)
	wsHandler := handler.NewWSHandler(cfg.Logger, cfg.UserClient)

	// Create base path group if configured
	var baseGroup *gin.RouterGroup
	if cfg.BasePath != "" {
		baseGroup = router.Group(cfg.BasePath)
		cfg.Logger.Info("Base path configured for ALB routing", zap.String("base_path", cfg.BasePath))
	} else {
		baseGroup = router.Group("")
		cfg.Logger.Info("No base path configured, using root path")
	}

	// Health check endpoint
	baseGroup.GET("/health", healthCheckHandler(cfg.DB))

	// Swagger documentation endpoint
	baseGroup.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Setup API routes
	setupRoutes(baseGroup, cfg.JWTSecret, projectHandler, boardHandler, participantHandler, commentHandler, fieldOptionHandler, projectMemberHandler, projectJoinRequestHandler, wsHandler)

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
	baseGroup *gin.RouterGroup,
	jwtSecret string,
	projectHandler *handler.ProjectHandler,
	boardHandler *handler.BoardHandler,
	participantHandler *handler.ParticipantHandler,
	commentHandler *handler.CommentHandler,
	fieldOptionHandler *handler.FieldOptionHandler,
	projectMemberHandler *handler.ProjectMemberHandler,
	projectJoinRequestHandler *handler.ProjectJoinRequestHandler,
	// 💡 [추가] WSHandler 인스턴스를 받도록 함수 정의 수정
	wsHandler *handler.WSHandler,
) {
	// API group with authentication
	api := baseGroup.Group("/api")
	api.Use(middleware.Auth(jwtSecret))
	{
		// Project routes
		projects := api.Group("/projects")
		{
			// Frontend compatibility route (query parameter style)
			projects.GET("", projectHandler.GetProjectsByWorkspaceQuery)

			// Existing routes
			projects.POST("", projectHandler.CreateProject)
			projects.GET("/workspace/:workspaceId", projectHandler.GetProjectsByWorkspace)
			projects.GET("/workspace/:workspaceId/default", projectHandler.GetDefaultProject)

			// New project management extension routes
			projects.GET("/search", projectHandler.SearchProjects)
			projects.GET("/:projectId", projectHandler.GetProject)
			projects.PUT("/:projectId", projectHandler.UpdateProject)
			projects.DELETE("/:projectId", projectHandler.DeleteProject)
			projects.GET("/:projectId/init-settings", projectHandler.GetProjectInitSettings)

			// Project member routes
			projects.GET("/:projectId/members", projectMemberHandler.GetMembers)
			projects.DELETE("/:projectId/members/:memberId", projectMemberHandler.RemoveMember)
			projects.PUT("/:projectId/members/:memberId/role", projectMemberHandler.UpdateMemberRole)

			// Project join request routes
			projects.GET("/:projectId/join-requests", projectJoinRequestHandler.GetJoinRequests)
		}

		// Join request routes (not nested under project)
		joinRequests := api.Group("/join-requests")
		{
			joinRequests.POST("", projectJoinRequestHandler.CreateJoinRequest)
			joinRequests.PUT("/:joinRequestId", projectJoinRequestHandler.UpdateJoinRequest)
		}

		// Board routes
		boards := api.Group("/boards")
		{
			// Frontend compatibility route (query parameter style)
			boards.GET("", boardHandler.GetBoardsByProjectQuery)

			boards.POST("", boardHandler.CreateBoard)
			boards.GET("/:boardId", boardHandler.GetBoard)
			boards.GET("/project/:projectId", boardHandler.GetBoardsByProject)
			boards.PUT("/:boardId", boardHandler.UpdateBoard)
			boards.DELETE("/:boardId", boardHandler.DeleteBoard)
			// 실시간 이동 API
			boards.PUT("/:boardId/move", boardHandler.MoveBoard)
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
			// Frontend compatibility route (query parameter style)
			comments.GET("", commentHandler.GetCommentsByQuery)

			comments.POST("", commentHandler.CreateComment)
			comments.GET("/board/:boardId", commentHandler.GetComments)
			comments.PUT("/:commentId", commentHandler.UpdateComment)
			comments.DELETE("/:commentId", commentHandler.DeleteComment)
		}

		// Field option routes
		fieldOptions := api.Group("/field-options")
		{
			fieldOptions.GET("", fieldOptionHandler.GetFieldOptions)
			fieldOptions.POST("", fieldOptionHandler.CreateFieldOption)
			fieldOptions.PATCH("/:optionId", fieldOptionHandler.UpdateFieldOption)
			fieldOptions.DELETE("/:optionId", fieldOptionHandler.DeleteFieldOption)
		}

	} // 👈 HTTP API 그룹 종료

	// WebSocket 엔드포인트 그룹 (인증 미들웨어 영향 받지 않음)
	wsGroup := baseGroup.Group("/api")
	{
		// 💡 [핵심] WebSocket 실시간 엔드포인트는 핸들러 내부에서 토큰 검증
		wsGroup.GET("/ws/project/:projectId", wsHandler.HandleWebSocket)
	}
}

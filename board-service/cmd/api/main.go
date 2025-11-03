package main

import (
	"board-service/internal/cache"
	"board-service/internal/client"
	"board-service/internal/config"
	"board-service/internal/database"
	"board-service/internal/handler"
	"board-service/internal/middleware"
	"board-service/internal/repository"
	"board-service/internal/service"
	"board-service/pkg/logger"
	"os"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	// 1. Load configuration
	cfg, err := config.Load()
	if err != nil {
		panic("Failed to load config: " + err.Error())
	}

	// 2. Initialize logger
	log, err := logger.Init(cfg.Log.Level, cfg.Server.Env)
	if err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
	defer log.Sync()

	log.Info("Starting board-service",
		zap.String("env", cfg.Server.Env),
		zap.String("port", cfg.Server.Port),
	)

	// 3. Connect to database
	db, err := database.Connect(cfg.Database.URL, log)
	if err != nil {
		log.Fatal("Failed to connect to database", zap.Error(err))
	}

	// 4. Connect to Redis
	rdb, err := cache.Connect(cfg.Redis.URL, log)
	if err != nil {
		log.Fatal("Failed to connect to Redis", zap.Error(err))
	}

	// 5. Initialize User Service client
	userClient := client.NewUserClient(cfg.UserService.URL)
	log.Info("User Service client initialized", zap.String("url", cfg.UserService.URL))

	// 5.5. Initialize repositories
	roleRepo := repository.NewRoleRepository(db)
	workspaceRepo := repository.NewWorkspaceRepository(db)
	projectRepo := repository.NewProjectRepository(db)
	customFieldRepo := repository.NewCustomFieldRepository(db)

	// 5.6. Initialize services
	// Note: customFieldService is created first, then injected into projectService
	customFieldService := service.NewCustomFieldService(customFieldRepo, projectRepo, roleRepo, log, db)
	workspaceService := service.NewWorkspaceService(workspaceRepo, roleRepo, userClient, log, db)
	projectService := service.NewProjectService(projectRepo, workspaceRepo, roleRepo, customFieldService, userClient, log, db)

	// 6. Configure Gin mode
	if cfg.Server.Env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 7. Create router
	r := gin.New()

	// 8. Register middleware (order is important)
	r.Use(middleware.RequestIDMiddleware())
	r.Use(middleware.LoggerMiddleware(log))
	r.Use(middleware.RecoveryMiddleware(log))
	r.Use(middleware.CORSMiddleware(cfg.CORS.Origins))

	// 9. Register health check (no authentication required)
	healthHandler := handler.NewHealthHandler(db, rdb)
	handler.RegisterRoutes(r, healthHandler)

	// 10. API routes group (authentication required)
	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware(cfg.JWT.Secret))
	{
		// Initialize handlers
		workspaceHandler := handler.NewWorkspaceHandler(workspaceService)
		projectHandler := handler.NewProjectHandler(projectService)
		customFieldHandler := handler.NewCustomFieldHandler(customFieldService)

		// Workspace routes
		workspaces := api.Group("/workspaces")
		{
			// Workspace CRUD
			workspaces.POST("", workspaceHandler.CreateWorkspace)
			workspaces.GET("/search", workspaceHandler.SearchWorkspaces) // Must be before /:id
			workspaces.GET("/:id", workspaceHandler.GetWorkspace)
			workspaces.PUT("/:id", workspaceHandler.UpdateWorkspace)
			workspaces.DELETE("/:id", workspaceHandler.DeleteWorkspace)

			// Join Requests
			workspaces.POST("/join-requests", workspaceHandler.CreateJoinRequest)
			workspaces.GET("/:id/join-requests", workspaceHandler.GetJoinRequests)
			workspaces.PUT("/join-requests/:id", workspaceHandler.UpdateJoinRequest)

			// Members
			workspaces.GET("/:id/members", workspaceHandler.GetWorkspaceMembers)
			workspaces.PUT("/:id/members/:memberId/role", workspaceHandler.UpdateMemberRole)
			workspaces.DELETE("/:id/members/:memberId", workspaceHandler.RemoveMember)

			// Default Workspace
			workspaces.POST("/default", workspaceHandler.SetDefaultWorkspace)
		}

		// Project routes
		projects := api.Group("/projects")
		{
			// Project CRUD
			projects.POST("", projectHandler.CreateProject)
			projects.GET("/search", projectHandler.SearchProjects) // Must be before /:id
			projects.GET("/:id", projectHandler.GetProject)
			projects.PUT("/:id", projectHandler.UpdateProject)
			projects.DELETE("/:id", projectHandler.DeleteProject)

			// Join Requests
			projects.POST("/join-requests", projectHandler.CreateJoinRequest)
			projects.GET("/:id/join-requests", projectHandler.GetJoinRequests)
			projects.PUT("/join-requests/:id", projectHandler.UpdateJoinRequest)

			// Members
			projects.GET("/:id/members", projectHandler.GetProjectMembers)
			projects.PUT("/:id/members/:memberId/role", projectHandler.UpdateMemberRole)
			projects.DELETE("/:id/members/:memberId", projectHandler.RemoveMember)
		}

		// Custom Fields routes
		customFields := api.Group("/custom-fields")
		{
			// Custom Roles
			customFields.POST("/roles", customFieldHandler.CreateCustomRole)
			customFields.GET("/projects/:projectId/roles", customFieldHandler.GetCustomRoles)
			customFields.GET("/roles/:id", customFieldHandler.GetCustomRole)
			customFields.PUT("/roles/:id", customFieldHandler.UpdateCustomRole)
			customFields.DELETE("/roles/:id", customFieldHandler.DeleteCustomRole)
			customFields.PUT("/projects/:projectId/roles/order", customFieldHandler.UpdateCustomRoleOrder)

			// Custom Stages
			customFields.POST("/stages", customFieldHandler.CreateCustomStage)
			customFields.GET("/projects/:projectId/stages", customFieldHandler.GetCustomStages)
			customFields.GET("/stages/:id", customFieldHandler.GetCustomStage)
			customFields.PUT("/stages/:id", customFieldHandler.UpdateCustomStage)
			customFields.DELETE("/stages/:id", customFieldHandler.DeleteCustomStage)
			customFields.PUT("/projects/:projectId/stages/order", customFieldHandler.UpdateCustomStageOrder)

			// Custom Importance
			customFields.POST("/importance", customFieldHandler.CreateCustomImportance)
			customFields.GET("/projects/:projectId/importance", customFieldHandler.GetCustomImportances)
			customFields.GET("/importance/:id", customFieldHandler.GetCustomImportance)
			customFields.PUT("/importance/:id", customFieldHandler.UpdateCustomImportance)
			customFields.DELETE("/importance/:id", customFieldHandler.DeleteCustomImportance)
			customFields.PUT("/projects/:projectId/importance/order", customFieldHandler.UpdateCustomImportanceOrder)
		}
	}

	// 11. Start server
	addr := ":" + cfg.Server.Port
	log.Info("Server starting", zap.String("address", addr))

	if err := r.Run(addr); err != nil {
		log.Fatal("Server failed to start", zap.Error(err))
		os.Exit(1)
	}
}

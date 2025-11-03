package main

import (
	"board-service/internal/cache"
	"board-service/internal/client"
	"board-service/internal/config"
	"board-service/internal/database"
	"board-service/internal/handler"
	"board-service/internal/middleware"
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

	// Suppress unused variable warning (will be used in future phases)
	_ = userClient

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
		// TODO: Add API routes in future phases
		// Example:
		// workspaceHandler := handler.NewWorkspaceHandler(...)
		// api.POST("/workspaces", workspaceHandler.Create)
	}

	// 11. Start server
	addr := ":" + cfg.Server.Port
	log.Info("Server starting", zap.String("address", addr))

	if err := r.Run(addr); err != nil {
		log.Fatal("Server failed to start", zap.Error(err))
		os.Exit(1)
	}
}

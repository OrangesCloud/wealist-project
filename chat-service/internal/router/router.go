// internal/router/router.go
package router

import (
	"chat-service/internal/client"
	"chat-service/internal/handler"
	"chat-service/internal/middleware"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

func SetupRouter(
	logger *zap.Logger,
	userClient client.UserClient,
	chatHandler *handler.ChatHandler,
	messageHandler *handler.MessageHandler,
	wsHandler *handler.WSHandler,
	corsOrigins string,
) *gin.Engine {
	router := gin.New()

	// Global Middleware
	router.Use(middleware.Logger(logger))
	router.Use(middleware.Recovery(logger))
	router.Use(middleware.CORS(corsOrigins))

	// Health Check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy", "service": "chat-service"})
	})

	// Swagger UI
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Auth Middleware
	authMiddleware := middleware.NewAuthMiddleware(userClient, logger)

	// 🔥 Presence Handler (router 내부에서 생성)
	presenceHandler := handler.NewPresenceHandler(wsHandler)

	// API Routes
	api := router.Group("/api/chats")
	api.Use(authMiddleware.RequireAuth())
	{
		// Chat Routes
		api.POST("", chatHandler.CreateChat)
		api.GET("/my", chatHandler.GetMyChats)
		api.GET("/workspace/:workspaceId", chatHandler.GetWorkspaceChats)
		api.GET("/:chatId", chatHandler.GetChat)
		api.DELETE("/:chatId", chatHandler.DeleteChat)
		api.POST("/:chatId/participants", chatHandler.AddParticipants)
		api.DELETE("/:chatId/participants/:userId", chatHandler.RemoveParticipant)

		// Message Routes
		api.GET("/messages/:chatId", messageHandler.GetMessages)
		api.POST("/messages/:chatId", messageHandler.SendMessage)
		api.DELETE("/messages/:messageId", messageHandler.DeleteMessage)
		api.POST("/messages/read", messageHandler.MarkMessagesAsRead)
		api.GET("/messages/:chatId/unread", messageHandler.GetUnreadCount)
		api.PUT("/messages/:chatId/last-read", messageHandler.UpdateLastRead)

		// WebSocket
		api.GET("/ws/:chatId", wsHandler.HandleWebSocket)

		// 🔥 Presence Routes
		api.GET("/presence/online", presenceHandler.GetOnlineUsers)
		api.GET("/presence/status/:userId", presenceHandler.CheckUserStatus)
	}

	return router
}
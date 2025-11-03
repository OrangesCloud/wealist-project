package handler

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	db    *gorm.DB
	redis *redis.Client
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp string            `json:"timestamp"`
	Services  map[string]string `json:"services"`
}

// NewHealthHandler creates a new health check handler
func NewHealthHandler(db *gorm.DB, redis *redis.Client) *HealthHandler {
	return &HealthHandler{
		db:    db,
		redis: redis,
	}
}

// Check performs health check on all services
func (h *HealthHandler) Check(c *gin.Context) {
	services := make(map[string]string)
	allHealthy := true

	// Check database
	sqlDB, err := h.db.DB()
	if err != nil || sqlDB.Ping() != nil {
		services["database"] = "down"
		allHealthy = false
	} else {
		services["database"] = "up"
	}

	// Check Redis
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	if err := h.redis.Ping(ctx).Err(); err != nil {
		services["redis"] = "down"
		allHealthy = false
	} else {
		services["redis"] = "up"
	}

	// Prepare response
	response := HealthResponse{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Services:  services,
	}

	if allHealthy {
		response.Status = "healthy"
		c.JSON(200, response)
	} else {
		response.Status = "unhealthy"
		c.JSON(503, response)
	}
}

// RegisterRoutes registers health check routes
func RegisterRoutes(r *gin.Engine, h *HealthHandler) {
	r.GET("/health", h.Check)
}

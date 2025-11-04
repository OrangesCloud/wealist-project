package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestIDMiddleware adds a unique request ID to each request
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if request ID is already present in headers
		requestID := c.GetHeader("X-Request-ID")

		// Generate new UUID if not present
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Store in context for later use
		c.Set("request_id", requestID)

		// Add to response headers
		c.Writer.Header().Set("X-Request-ID", requestID)

		c.Next()
	}
}

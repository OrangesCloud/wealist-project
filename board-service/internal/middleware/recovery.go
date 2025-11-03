package middleware

import (
	"board-service/internal/apperrors"
	"board-service/internal/dto"
	"fmt"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RecoveryMiddleware recovers from panics and logs the error
func RecoveryMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Get request ID
				requestID, _ := c.Get("request_id")

				// Log panic with stack trace
				logger.Error("Panic recovered",
					zap.Any("error", err),
					zap.String("request_id", fmt.Sprintf("%v", requestID)),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
					zap.ByteString("stack", debug.Stack()),
				)

				// Return 500 error
				appErr := apperrors.New(
					apperrors.ErrCodeInternalServer,
					"서버 내부 오류가 발생했습니다",
					500,
				)

				dto.Error(c, appErr)
				c.Abort()
			}
		}()

		c.Next()
	}
}

package middleware

import (
	"board-service/internal/apperrors"
	"board-service/internal/dto"
	jwtpkg "board-service/pkg/jwt"
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
)

// Context keys
const (
	UserIDKey = "user_id"
	TokenKey  = "token"
)

// AuthMiddleware validates JWT token and extracts user ID
func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get Authorization header
		authHeader := c.GetHeader("Authorization")

		// Debug logging
		c.Writer.Header().Set("X-Debug-Auth-Header-Present", "false")
		if authHeader != "" {
			c.Writer.Header().Set("X-Debug-Auth-Header-Present", "true")
		}

		if authHeader == "" {
			// Log missing token error
			c.Writer.Header().Set("X-Debug-Error", "Missing Authorization header")
			dto.Error(c, apperrors.ErrMissingToken)
			c.Abort()
			return
		}

		// Check Bearer prefix
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			// Log invalid format error
			c.Writer.Header().Set("X-Debug-Error", "Invalid Authorization header format")
			dto.Error(c, apperrors.ErrInvalidToken)
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Validate token
		claims, err := jwtpkg.ValidateToken(tokenString, jwtSecret)
		if err != nil {
			var appErr *apperrors.AppError

			// Map JWT errors to app errors
			if errors.Is(err, jwtpkg.ErrExpiredToken) {
				appErr = apperrors.ErrTokenExpired
				c.Writer.Header().Set("X-Debug-Error", "Token expired")
			} else if errors.Is(err, jwtpkg.ErrInvalidToken) || errors.Is(err, jwtpkg.ErrInvalidClaims) {
				appErr = apperrors.ErrInvalidToken
				c.Writer.Header().Set("X-Debug-Error", "Invalid token or claims: "+err.Error())
			} else {
				appErr = apperrors.ErrUnauthorized
				c.Writer.Header().Set("X-Debug-Error", "Unauthorized: "+err.Error())
			}

			dto.Error(c, appErr)
			c.Abort()
			return
		}

		// Store user ID and token in context
		c.Set(UserIDKey, claims.Sub)
		c.Set(TokenKey, tokenString)

		c.Next()
	}
}

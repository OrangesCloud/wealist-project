package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"project-board-api/internal/response"
)

// handleServiceError maps service layer errors to appropriate HTTP responses
func handleServiceError(c *gin.Context, err error) {
	// Check for GORM errors
	if errors.Is(err, gorm.ErrRecordNotFound) {
		response.SendError(c, http.StatusNotFound, response.ErrCodeNotFound, "Resource not found")
		return
	}

	// Check for custom AppError
	var appErr *response.AppError
	if errors.As(err, &appErr) {
		statusCode := mapErrorCodeToHTTPStatus(appErr.Code)
		response.SendError(c, statusCode, appErr.Code, appErr.Message)
		return
	}

	// Default to internal server error
	response.SendError(c, http.StatusInternalServerError, response.ErrCodeInternal, "Internal server error")
}

// mapErrorCodeToHTTPStatus maps error codes to HTTP status codes
func mapErrorCodeToHTTPStatus(code string) int {
	switch code {
	case response.ErrCodeNotFound:
		return http.StatusNotFound
	case response.ErrCodeAlreadyExists:
		return http.StatusConflict
	case response.ErrCodeValidation:
		return http.StatusBadRequest
	case response.ErrCodeUnauthorized:
		return http.StatusUnauthorized
	case response.ErrCodeForbidden:
		return http.StatusForbidden
	default:
		return http.StatusInternalServerError
	}
}

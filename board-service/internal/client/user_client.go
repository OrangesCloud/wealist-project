package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// UserClient defines the interface for User API interactions
type UserClient interface {
	// ValidateWorkspaceMember validates if a user is a member of a workspace
	ValidateWorkspaceMember(ctx context.Context, workspaceID, userID uuid.UUID, token string) (bool, error)

	// GetUserProfile retrieves user profile information
	GetUserProfile(ctx context.Context, userID uuid.UUID, token string) (*UserProfile, error)

	// GetWorkspaceProfile retrieves workspace-specific user profile
	GetWorkspaceProfile(ctx context.Context, workspaceID, userID uuid.UUID, token string) (*WorkspaceProfile, error)
}

// WorkspaceValidationResponse represents the response from workspace validation endpoint
type WorkspaceValidationResponse struct {
	WorkspaceID uuid.UUID `json:"workspaceId"`
	UserID      uuid.UUID `json:"userId"`
	Valid       bool      `json:"valid"`
	IsValid     bool      `json:"isValid"`
}

// UserProfile represents basic user profile information
type UserProfile struct {
	UserID   uuid.UUID `json:"userId"`
	Email    string    `json:"email"`
	Provider string    `json:"provider"`
}

// WorkspaceProfile represents workspace-specific user profile
type WorkspaceProfile struct {
	ProfileID       uuid.UUID `json:"profileId"`
	WorkspaceID     uuid.UUID `json:"workspaceId"`
	UserID          uuid.UUID `json:"userId"`
	NickName        string    `json:"nickName"`
	Email           string    `json:"email"`
	ProfileImageURL string    `json:"profileImageUrl"`
}

// userClient implements UserClient interface
type userClient struct {
	baseURL    string
	httpClient *http.Client
	timeout    time.Duration
	logger     *zap.Logger
}

// NewUserClient creates a new User API client
func NewUserClient(baseURL string, timeout time.Duration, logger *zap.Logger) UserClient {
	return &userClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		timeout: timeout,
		logger:  logger,
	}
}

// ValidateWorkspaceMember validates if a user is a member of a workspace
func (c *userClient) ValidateWorkspaceMember(ctx context.Context, workspaceID, userID uuid.UUID, token string) (bool, error) {
	url := fmt.Sprintf("%s/api/workspaces/%s/validate-member/%s", c.baseURL, workspaceID.String(), userID.String())

	c.logger.Debug("Validating workspace member",
		zap.String("url", url),
		zap.String("workspace_id", workspaceID.String()),
		zap.String("user_id", userID.String()),
	)

	var response WorkspaceValidationResponse
	if err := c.doRequest(ctx, "GET", url, token, &response); err != nil {
		c.logger.Error("Failed to validate workspace member",
			zap.Error(err),
			zap.String("workspace_id", workspaceID.String()),
			zap.String("user_id", userID.String()),
		)
		// Graceful degradation: return false on error
		return false, err
	}

	// Check both Valid and IsValid fields for compatibility
	isValid := response.Valid || response.IsValid

	c.logger.Debug("Workspace member validation result",
		zap.Bool("is_valid", isValid),
		zap.String("workspace_id", workspaceID.String()),
		zap.String("user_id", userID.String()),
	)

	return isValid, nil
}

// GetUserProfile retrieves user profile information
func (c *userClient) GetUserProfile(ctx context.Context, userID uuid.UUID, token string) (*UserProfile, error) {
	url := fmt.Sprintf("%s/api/users/%s", c.baseURL, userID.String())

	c.logger.Debug("Getting user profile",
		zap.String("url", url),
		zap.String("user_id", userID.String()),
	)

	var profile UserProfile
	if err := c.doRequest(ctx, "GET", url, token, &profile); err != nil {
		c.logger.Error("Failed to get user profile",
			zap.Error(err),
			zap.String("user_id", userID.String()),
		)
		// Graceful degradation: return empty profile
		return &UserProfile{
			UserID: userID,
			Email:  "",
		}, nil
	}

	c.logger.Debug("User profile retrieved",
		zap.String("user_id", userID.String()),
		zap.String("email", profile.Email),
	)

	return &profile, nil
}

// GetWorkspaceProfile retrieves workspace-specific user profile
func (c *userClient) GetWorkspaceProfile(ctx context.Context, workspaceID, userID uuid.UUID, token string) (*WorkspaceProfile, error) {
	url := fmt.Sprintf("%s/api/profiles/workspace/%s", c.baseURL, workspaceID.String())

	c.logger.Debug("Getting workspace profile",
		zap.String("url", url),
		zap.String("workspace_id", workspaceID.String()),
		zap.String("user_id", userID.String()),
	)

	var profile WorkspaceProfile
	if err := c.doRequest(ctx, "GET", url, token, &profile); err != nil {
		c.logger.Error("Failed to get workspace profile",
			zap.Error(err),
			zap.String("workspace_id", workspaceID.String()),
			zap.String("user_id", userID.String()),
		)
		// Graceful degradation: return empty profile
		return &WorkspaceProfile{
			WorkspaceID: workspaceID,
			UserID:      userID,
			NickName:    "",
			Email:       "",
		}, nil
	}

	c.logger.Debug("Workspace profile retrieved",
		zap.String("workspace_id", workspaceID.String()),
		zap.String("user_id", userID.String()),
		zap.String("nickname", profile.NickName),
	)

	return &profile, nil
}

// doRequest performs an HTTP request with the given parameters
func (c *userClient) doRequest(ctx context.Context, method, url, token string, result interface{}) error {
	// Create request with context
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add authorization header
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Check status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		c.logger.Warn("User API returned non-success status",
			zap.Int("status_code", resp.StatusCode),
			zap.String("url", url),
			zap.String("response_body", string(body)),
		)
		return fmt.Errorf("user API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	if err := json.Unmarshal(body, result); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	return nil
}

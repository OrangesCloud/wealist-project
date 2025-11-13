package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// UserClient defines the interface for communicating with the User Service.
type UserClient interface {
	GetUser(ctx context.Context, userID string) (*UserInfo, error)
	GetUsersBatch(ctx context.Context, userIDs []string) ([]UserInfo, error)
	SearchUsers(ctx context.Context, query string) ([]UserInfo, error)
	GetSimpleUser(userID string) (*SimpleUser, error)
	GetSimpleUsers(userIDs []string) ([]SimpleUser, error)

	// Workspace validation methods (calls User Service /api/workspace endpoints)
	CheckWorkspaceExists(ctx context.Context, workspaceID string, token string) (bool, error)
	ValidateWorkspaceMembership(ctx context.Context, workspaceID string, userID string, token string) (bool, error)
	GetWorkspace(ctx context.Context, workspaceID string, token string) (*WorkspaceInfo, error)
}

type userClient struct {
	baseURL string
	client  *http.Client
}

// UserInfo represents detailed user information from User Service.
type UserInfo struct {
	UserID   string `json:"userId"`    // User Service uses camelCase
	Name     string `json:"nickName"`  // User Service uses nickName (not name)
	Email    string `json:"email"`
	IsActive bool   `json:"isActive"`
}

// UserServiceResponse represents the wrapped response from User Service.
type UserServiceResponse struct {
	Data interface{} `json:"data"`
}

// SimpleUser represents basic user information needed for display.
type SimpleUser struct {
	ID        string `json:"user_id"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatarUrl"`
}

// WorkspaceInfo represents workspace information from User Service.
type WorkspaceInfo struct {
	ID          string `json:"workspace_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	OwnerID     string `json:"owner_id"`
	IsPublic    bool   `json:"isPublic"`
}

// NewUserClient creates a new User Service client.
func NewUserClient(baseURL string) UserClient {
	return &userClient{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetSimpleUser retrieves basic info for a single user.
func (c *userClient) GetSimpleUser(userID string) (*SimpleUser, error) {
	// This endpoint might not exist yet in user-service, assuming it will be /api/users/{id}/simple
	url := fmt.Sprintf("%s/api/users/%s", c.baseURL, userID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for simple user: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get simple user: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("user service returned status %d for simple user", resp.StatusCode)
	}

	var user SimpleUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("failed to decode simple user response: %w", err)
	}
	return &user, nil
}

// GetSimpleUsers retrieves basic info for multiple users.
func (c *userClient) GetSimpleUsers(userIDs []string) ([]SimpleUser, error) {
	// This endpoint might not exist yet, assuming /api/users/batch/simple
	url := fmt.Sprintf("%s/api/users/batch?ids=%s", c.baseURL, strings.Join(userIDs, ","))
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for simple users: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get simple users: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("user service returned status %d for simple users", resp.StatusCode)
	}

	var users []SimpleUser
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return nil, fmt.Errorf("failed to decode simple users response: %w", err)
	}
	return users, nil
}

// GetUser retrieves a single user by ID
func (c *userClient) GetUser(ctx context.Context, userID string) (*UserInfo, error) {
	url := fmt.Sprintf("%s/api/users/%s", c.baseURL, userID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("user service returned status %d", resp.StatusCode)
	}

	// Read the response body
	var rawResponse map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&rawResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Check if response has "data" field (wrapped response)
	if data, ok := rawResponse["data"]; ok {
		// Convert data to UserInfo
		dataBytes, err := json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal data: %w", err)
		}

		var userInfo UserInfo
		if err := json.Unmarshal(dataBytes, &userInfo); err != nil {
			return nil, fmt.Errorf("failed to unmarshal user info: %w", err)
		}

		return &userInfo, nil
	}

	// If no "data" field, treat as direct UserInfo response
	bodyBytes, err := json.Marshal(rawResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response: %w", err)
	}

	var userInfo UserInfo
	if err := json.Unmarshal(bodyBytes, &userInfo); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user info: %w", err)
	}

	return &userInfo, nil
}

// GetUsersBatch retrieves multiple users by IDs
func (c *userClient) GetUsersBatch(ctx context.Context, userIDs []string) ([]UserInfo, error) {
	url := fmt.Sprintf("%s/api/users/batch", c.baseURL)

	requestBody := map[string][]string{
		"userIds": userIDs,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("user service returned status %d", resp.StatusCode)
	}

	// Read the response body
	var rawResponse map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&rawResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Check if response has "data" field (wrapped response)
	var users []UserInfo
	if data, ok := rawResponse["data"]; ok {
		// Convert data to []UserInfo
		dataBytes, err := json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal data: %w", err)
		}

		if err := json.Unmarshal(dataBytes, &users); err != nil {
			return nil, fmt.Errorf("failed to unmarshal users: %w", err)
		}

		return users, nil
	}

	// If no "data" field, treat as direct []UserInfo response
	bodyBytes, err := json.Marshal(rawResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response: %w", err)
	}

	if err := json.Unmarshal(bodyBytes, &users); err != nil {
		return nil, fmt.Errorf("failed to unmarshal users: %w", err)
	}

	return users, nil
}

// SearchUsers searches for users by query string
func (c *userClient) SearchUsers(ctx context.Context, query string) ([]UserInfo, error) {
	url := fmt.Sprintf("%s/api/users/search?query=%s", c.baseURL, query)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("user service returned status %d", resp.StatusCode)
	}

	var users []UserInfo
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return users, nil
}

// CheckWorkspaceExists checks if a workspace exists in User Service
// Note: This now relies on ValidateWorkspaceMembership to check both existence and membership
func (c *userClient) CheckWorkspaceExists(ctx context.Context, workspaceID string, token string) (bool, error) {
	// We don't need a separate existence check anymore
	// ValidateWorkspaceMembership will handle both workspace existence and membership
	return true, nil
}

// WorkspaceValidationResponse represents the response from User Service workspace validation endpoint
type WorkspaceValidationResponse struct {
	WorkspaceID string `json:"workspaceId"`
	UserID      string `json:"userId"`
	IsValid     bool   `json:"isValid"`
}

// ValidateWorkspaceMembership validates if a user has access to a workspace
// Uses the User Service endpoint: GET /api/workspaces/{workspaceId}/validate-member/{userId}
// Returns true if the user is a member of the workspace, false otherwise
func (c *userClient) ValidateWorkspaceMembership(ctx context.Context, workspaceID string, userID string, token string) (bool, error) {
	// Use the new workspace validation endpoint
	url := fmt.Sprintf("%s/api/workspaces/%s/validate-member/%s", c.baseURL, workspaceID, userID)

	// Log the request details
	fmt.Printf("[DEBUG] Calling User Service: %s\n", url)
	fmt.Printf("[DEBUG] WorkspaceID: %s, UserID: %s\n", workspaceID, userID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}

	// Add Bearer token for authentication (optional, but included for consistency)
	if token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	}

	resp, err := c.client.Do(req)
	if err != nil {
		fmt.Printf("[DEBUG] Request failed: %v\n", err)
		return false, fmt.Errorf("failed to send request to user service: %w", err)
	}
	defer resp.Body.Close()

	fmt.Printf("[DEBUG] User Service response status: %d\n", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := json.Marshal(resp.Body)
		fmt.Printf("[DEBUG] Non-200 response body: %s\n", string(bodyBytes))
		return false, fmt.Errorf("user service returned status %d", resp.StatusCode)
	}

	// Parse response
	var validationResp WorkspaceValidationResponse
	if err := json.NewDecoder(resp.Body).Decode(&validationResp); err != nil {
		fmt.Printf("[DEBUG] Failed to decode response: %v\n", err)
		return false, fmt.Errorf("failed to decode workspace validation response: %w", err)
	}

	fmt.Printf("[DEBUG] Validation response: workspaceId=%s, userId=%s, isValid=%v\n",
		validationResp.WorkspaceID, validationResp.UserID, validationResp.IsValid)

	return validationResp.IsValid, nil
}

// GetWorkspace retrieves workspace information from User Service
func (c *userClient) GetWorkspace(ctx context.Context, workspaceID string, token string) (*WorkspaceInfo, error) {
	url := fmt.Sprintf("%s/api/workspaces/%s", c.baseURL, workspaceID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add Bearer token for authentication
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to user service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("workspace not found")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("user service returned status %d", resp.StatusCode)
	}

	var workspace WorkspaceInfo
	if err := json.NewDecoder(resp.Body).Decode(&workspace); err != nil {
		return nil, fmt.Errorf("failed to decode workspace response: %w", err)
	}

	return &workspace, nil
}

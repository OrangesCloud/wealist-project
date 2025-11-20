package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type UserClient interface {
	ValidateToken(ctx context.Context, token string) (*TokenValidationResponse, error)
	GetUserInfo(ctx context.Context, userID, token string) (*UserInfo, error)
}

type userClient struct {
	baseURL    string
	httpClient *http.Client
}

type TokenValidationResponse struct {
	UserID  string `json:"userId"`
	Valid   bool   `json:"valid"`
	Message string `json:"message"`
}

type UserInfo struct {
	UserID          string `json:"userId"`
	Email           string `json:"email"`
	NickName        string `json:"nickName,omitempty"`
	ProfileImageURL string `json:"profileImageUrl,omitempty"`
}

func NewUserClient(baseURL string, timeout time.Duration) UserClient {
	return &userClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *userClient) ValidateToken(ctx context.Context, token string) (*TokenValidationResponse, error) {
	url := fmt.Sprintf("%s/auth/validate-access-token?token=%s", c.baseURL, token)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("validation failed: status=%d, body=%s", resp.StatusCode, string(body))
	}

	var result TokenValidationResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

func (c *userClient) GetUserInfo(ctx context.Context, userID, token string) (*UserInfo, error) {
	url := fmt.Sprintf("%s/users/%s", c.baseURL, userID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get user info failed: status=%d, body=%s", resp.StatusCode, string(body))
	}

	var result UserInfo
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}
package metrics

import (
	"errors"
	"regexp"
	"strings"
	"testing"
	"testing/quick"
)

func TestNormalizeEndpoint(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		expected string
	}{
		{
			name:     "UUID in path",
			endpoint: "/api/users/123e4567-e89b-12d3-a456-426614174000",
			expected: "/api/users/{id}",
		},
		{
			name:     "Multiple UUIDs",
			endpoint: "/api/users/123e4567-e89b-12d3-a456-426614174000/projects/987fcdeb-51a2-43f1-b456-789012345678",
			expected: "/api/users/{id}/projects/{id}",
		},
		{
			name:     "No UUID",
			endpoint: "/api/users",
			expected: "/api/users",
		},
		{
			name:     "UUID with query params",
			endpoint: "/api/users/123e4567-e89b-12d3-a456-426614174000?include=profile",
			expected: "/api/users/{id}?include=profile",
		},
		{
			name:     "Lowercase UUID",
			endpoint: "/api/users/abcdef12-3456-7890-abcd-ef1234567890",
			expected: "/api/users/{id}",
		},
		{
			name:     "Empty string",
			endpoint: "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeEndpoint(tt.endpoint)
			if result != tt.expected {
				t.Errorf("normalizeEndpoint(%q) = %q, want %q", tt.endpoint, result, tt.expected)
			}
		})
	}
}

func TestGetErrorType(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "Nil error",
			err:      nil,
			expected: "none",
		},
		{
			name:     "Generic error",
			err:      errors.New("some error"),
			expected: "unknown",
		},
		{
			name:     "Network error",
			err:      errors.New("connection refused"),
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getErrorType(tt.err)
			if result != tt.expected {
				t.Errorf("getErrorType(%v) = %q, want %q", tt.err, result, tt.expected)
			}
		})
	}
}

// Property 10: 엔드포인트 정규화
// Feature: board-service-prometheus-metrics, Property 10: Endpoint normalization
// For any endpoint containing UUIDs, normalizeEndpoint should replace all UUIDs with {id} template
// Validates: Requirements 5.5
func TestProperty_EndpointNormalization(t *testing.T) {
	// UUID pattern to verify
	uuidRegex := regexp.MustCompile(`[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`)

	// Property: For any endpoint, after normalization, it should not contain any UUIDs
	property := func(endpoint string) bool {
		// Normalize the endpoint
		normalized := normalizeEndpoint(endpoint)

		// Verify that no UUIDs remain in the normalized endpoint
		if uuidRegex.MatchString(normalized) {
			t.Logf("UUID found in normalized endpoint: %s -> %s", endpoint, normalized)
			return false
		}

		// Verify that if the original had UUIDs, they were replaced with {id}
		originalHadUUID := uuidRegex.MatchString(endpoint)
		normalizedHasTemplate := strings.Contains(normalized, "{id}")

		if originalHadUUID && !normalizedHasTemplate {
			t.Logf("Original had UUID but normalized doesn't have {id}: %s -> %s", endpoint, normalized)
			return false
		}

		return true
	}

	config := &quick.Config{
		MaxCount: 100,
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property test failed: %v", err)
	}
}

// Additional property test: Idempotence of normalization
// Normalizing an already normalized endpoint should not change it
func TestProperty_NormalizationIdempotence(t *testing.T) {
	property := func(endpoint string) bool {
		// Normalize once
		normalized1 := normalizeEndpoint(endpoint)

		// Normalize again
		normalized2 := normalizeEndpoint(normalized1)

		// They should be identical
		if normalized1 != normalized2 {
			t.Logf("Normalization is not idempotent: %s -> %s -> %s", endpoint, normalized1, normalized2)
			return false
		}

		return true
	}

	config := &quick.Config{
		MaxCount: 100,
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property test failed: %v", err)
	}
}

package metrics

import (
	"regexp"
	"strconv"
	"time"
)

var (
	// UUID pattern for endpoint normalization
	uuidPattern = regexp.MustCompile(`[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`)
)

// RecordExternalAPICall records external API call metrics
func (m *Metrics) RecordExternalAPICall(endpoint, method string, statusCode int, duration time.Duration, err error) {
	m.safeExecute("RecordExternalAPICall", func() {
		endpoint = normalizeEndpoint(endpoint)
		status := strconv.Itoa(statusCode)

		m.ExternalAPIRequestsTotal.WithLabelValues(endpoint, method, status).Inc()
		m.ExternalAPIRequestDuration.WithLabelValues(endpoint, status).Observe(duration.Seconds())

		if err != nil {
			errorType := getErrorType(err)
			m.ExternalAPIErrors.WithLabelValues(endpoint, errorType).Inc()
		}
	})
}

// normalizeEndpoint converts actual IDs to templates
// Example: /api/users/123e4567-e89b-12d3-a456-426614174000 -> /api/users/{id}
func normalizeEndpoint(endpoint string) string {
	return uuidPattern.ReplaceAllString(endpoint, "{id}")
}

// getErrorType categorizes error types
func getErrorType(err error) string {
	if err == nil {
		return "none"
	}
	// Simple categorization - can be enhanced later
	return "unknown"
}

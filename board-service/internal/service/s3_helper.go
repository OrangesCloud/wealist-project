package service

import (
	// strings만 필요
	"strings"
)

// extractS3KeyFromURL extracts the S3 key from a full S3 URL
// Example: https://bucket.s3.region.amazonaws.com/board/boards/workspace/2024/01/file.jpg -> board/boards/workspace/2024/01/file.jpg
// 이 함수는 project_service.go의 deleteAttachmentsWithS3에서 사용됩니다.
func extractS3KeyFromURL(fileURL string) string {
	// Find the position after the domain
	// Format: https://{bucket}.s3.{region}.amazonaws.com/{key}
	start := strings.Index(fileURL, ".amazonaws.com/")
	if start == -1 {
		// Try alternative format for MinIO or custom endpoints
		// Format: http://localhost:9000/{bucket}/{key}
		parts := strings.SplitN(fileURL, "/", 5)
		if len(parts) >= 5 {
			// Skip protocol, domain, and bucket name, return key only
			return strings.Join(parts[4:], "/")
		}
		return ""
	}

	// Extract key after .amazonaws.com/
	key := fileURL[start+len(".amazonaws.com/"):]
	return key
}

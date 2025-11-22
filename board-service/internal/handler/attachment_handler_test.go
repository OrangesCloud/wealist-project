package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"project-board-api/internal/client"
	"project-board-api/internal/config"
	"project-board-api/internal/domain"
)

// Mock attachment repository for testing
type mockAttachmentRepository struct {
	createFunc func(ctx context.Context, attachment *domain.Attachment) error
}

func (m *mockAttachmentRepository) Create(ctx context.Context, attachment *domain.Attachment) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, attachment)
	}
	// Default: set ID and timestamps
	attachment.ID = uuid.New()
	attachment.CreatedAt = time.Now()
	attachment.UpdatedAt = time.Now()
	return nil
}

func (m *mockAttachmentRepository) FindByEntityID(ctx context.Context, entityType domain.EntityType, entityID uuid.UUID) ([]*domain.Attachment, error) {
	return nil, nil
}

func (m *mockAttachmentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

// setupAttachmentHandler creates a test handler with a mock S3 client
func setupAttachmentHandler(t *testing.T) (*AttachmentHandler, *gin.Engine) {
	gin.SetMode(gin.TestMode)

	// Create S3 config for testing
	cfg := &config.S3Config{
		Bucket:    "test-bucket",
		Region:    "us-east-1",
		AccessKey: "test-access-key",
		SecretKey: "test-secret-key",
		Endpoint:  "", // Use real AWS SDK for testing
	}

	// Create S3 client
	s3Client, err := client.NewS3Client(cfg)
	require.NoError(t, err, "Failed to create S3 client")

	// Create mock repository
	mockRepo := &mockAttachmentRepository{}

	// Create handler
	handler := NewAttachmentHandler(s3Client, mockRepo)

	// Setup router
	router := gin.New()
	router.POST("/attachments/presigned-url", handler.GeneratePresignedURL)

	return handler, router
}

// TestGeneratePresignedURL_Success tests successful presigned URL generation
func TestGeneratePresignedURL_Success(t *testing.T) {
	_, router := setupAttachmentHandler(t)

	tests := []struct {
		name        string
		entityType  string
		fileName    string
		contentType string
	}{
		{
			name:        "Board image upload",
			entityType:  "BOARD",
			fileName:    "test-image.jpg",
			contentType: "image/jpeg",
		},
		{
			name:        "Comment PDF upload",
			entityType:  "COMMENT",
			fileName:    "document.pdf",
			contentType: "application/pdf",
		},
		{
			name:        "Project PNG upload",
			entityType:  "PROJECT",
			fileName:    "diagram.png",
			contentType: "image/png",
		},
		{
			name:        "Board DOCX upload",
			entityType:  "BOARD",
			fileName:    "report.docx",
			contentType: "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody := PresignedURLRequest{
				EntityType:  tt.entityType,
				WorkspaceID: "550e8400-e29b-41d4-a716-446655440000",
				FileName:    tt.fileName,
				FileSize:    1024000, // 1MB
				ContentType: tt.contentType,
			}

			body, err := json.Marshal(reqBody)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/attachments/presigned-url", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			// Check response structure
			data, ok := response["data"].(map[string]interface{})
			require.True(t, ok, "Response should contain data field")

			// Verify presigned URL fields
			uploadURL, ok := data["uploadUrl"].(string)
			assert.True(t, ok, "Response should contain uploadUrl")
			assert.NotEmpty(t, uploadURL, "Upload URL should not be empty")

			fileKey, ok := data["fileKey"].(string)
			assert.True(t, ok, "Response should contain fileKey")
			assert.NotEmpty(t, fileKey, "File key should not be empty")

			expiresIn, ok := data["expiresIn"].(float64)
			assert.True(t, ok, "Response should contain expiresIn")
			assert.Equal(t, float64(300), expiresIn, "ExpiresIn should be 300 seconds")
		})
	}
}

// TestGeneratePresignedURL_FileSizeExceeded tests file size validation
func TestGeneratePresignedURL_FileSizeExceeded(t *testing.T) {
	_, router := setupAttachmentHandler(t)

	reqBody := PresignedURLRequest{
		EntityType:  "BOARD",
		WorkspaceID: "550e8400-e29b-41d4-a716-446655440000",
		FileName:    "large-file.jpg",
		FileSize:    21 * 1024 * 1024, // 21MB (exceeds 20MB limit)
		ContentType: "image/jpeg",
	}

	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/attachments/presigned-url", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	errorData, ok := response["error"].(map[string]interface{})
	require.True(t, ok, "Response should contain error field")

	code, ok := errorData["code"].(string)
	assert.True(t, ok)
	assert.Equal(t, "FILE_TOO_LARGE", code)
}

// TestGeneratePresignedURL_InvalidFileType tests file type validation
func TestGeneratePresignedURL_InvalidFileType(t *testing.T) {
	_, router := setupAttachmentHandler(t)

	tests := []struct {
		name        string
		fileName    string
		contentType string
	}{
		{
			name:        "Audio file (mp3)",
			fileName:    "audio.mp3",
			contentType: "audio/mpeg",
		},
		{
			name:        "Video file (mp4)",
			fileName:    "video.mp4",
			contentType: "video/mp4",
		},
		{
			name:        "Unsupported document type",
			fileName:    "archive.zip",
			contentType: "application/zip",
		},
		{
			name:        "Mismatched extension and content type",
			fileName:    "image.jpg",
			contentType: "application/pdf",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody := PresignedURLRequest{
				EntityType:  "BOARD",
				WorkspaceID: "550e8400-e29b-41d4-a716-446655440000",
				FileName:    tt.fileName,
				FileSize:    1024000,
				ContentType: tt.contentType,
			}

			body, err := json.Marshal(reqBody)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/attachments/presigned-url", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)

			var response map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			errorData, ok := response["error"].(map[string]interface{})
			require.True(t, ok, "Response should contain error field")

			code, ok := errorData["code"].(string)
			assert.True(t, ok)
			assert.Equal(t, "INVALID_FILE_TYPE", code)
		})
	}
}

// TestGeneratePresignedURL_InvalidEntityType tests entity type validation
func TestGeneratePresignedURL_InvalidEntityType(t *testing.T) {
	_, router := setupAttachmentHandler(t)

	tests := []struct {
		name       string
		entityType string
	}{
		{
			name:       "Invalid entity type - USER",
			entityType: "USER",
		},
		{
			name:       "Invalid entity type - WORKSPACE",
			entityType: "WORKSPACE",
		},
		{
			name:       "Empty entity type",
			entityType: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody := PresignedURLRequest{
				EntityType:  tt.entityType,
				WorkspaceID: "550e8400-e29b-41d4-a716-446655440000",
				FileName:    "test.jpg",
				FileSize:    1024000,
				ContentType: "image/jpeg",
			}

			body, err := json.Marshal(reqBody)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/attachments/presigned-url", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)

			var response map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			errorData, ok := response["error"].(map[string]interface{})
			require.True(t, ok, "Response should contain error field")

			code, ok := errorData["code"].(string)
			assert.True(t, ok)
			assert.Equal(t, "VALIDATION_ERROR", code)
		})
	}
}

// TestGeneratePresignedURL_InvalidWorkspaceID tests workspace ID validation
func TestGeneratePresignedURL_InvalidWorkspaceID(t *testing.T) {
	_, router := setupAttachmentHandler(t)

	reqBody := PresignedURLRequest{
		EntityType:  "BOARD",
		WorkspaceID: "invalid-uuid",
		FileName:    "test.jpg",
		FileSize:    1024000,
		ContentType: "image/jpeg",
	}

	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/attachments/presigned-url", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	errorData, ok := response["error"].(map[string]interface{})
	require.True(t, ok, "Response should contain error field")

	code, ok := errorData["code"].(string)
	assert.True(t, ok)
	assert.Equal(t, "VALIDATION_ERROR", code)
}

// TestGeneratePresignedURL_ZeroFileSize tests zero file size validation
func TestGeneratePresignedURL_ZeroFileSize(t *testing.T) {
	_, router := setupAttachmentHandler(t)

	reqBody := PresignedURLRequest{
		EntityType:  "BOARD",
		WorkspaceID: "550e8400-e29b-41d4-a716-446655440000",
		FileName:    "test.jpg",
		FileSize:    0,
		ContentType: "image/jpeg",
	}

	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/attachments/presigned-url", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	errorData, ok := response["error"].(map[string]interface{})
	require.True(t, ok, "Response should contain error field")

	code, ok := errorData["code"].(string)
	assert.True(t, ok)
	assert.Equal(t, "VALIDATION_ERROR", code)
}

// TestGeneratePresignedURL_MissingExtension tests file without extension
func TestGeneratePresignedURL_MissingExtension(t *testing.T) {
	_, router := setupAttachmentHandler(t)

	reqBody := PresignedURLRequest{
		EntityType:  "BOARD",
		WorkspaceID: "550e8400-e29b-41d4-a716-446655440000",
		FileName:    "testfile",
		FileSize:    1024000,
		ContentType: "image/jpeg",
	}

	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/attachments/presigned-url", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	errorData, ok := response["error"].(map[string]interface{})
	require.True(t, ok, "Response should contain error field")

	code, ok := errorData["code"].(string)
	assert.True(t, ok)
	assert.Equal(t, "INVALID_FILE_TYPE", code)
}

// TestValidateEntityType tests the validateEntityType function
func TestValidateEntityType(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
	}{
		{"Valid BOARD", "BOARD", false},
		{"Valid board (lowercase)", "board", false},
		{"Valid COMMENT", "COMMENT", false},
		{"Valid PROJECT", "PROJECT", false},
		{"Invalid USER", "USER", true},
		{"Invalid empty", "", true},
		{"Invalid random", "RANDOM", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := validateEntityType(tt.input)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestValidateFileType tests the validateFileType function
func TestValidateFileType(t *testing.T) {
	tests := []struct {
		name        string
		fileName    string
		contentType string
		expectError bool
	}{
		{"Valid JPEG", "test.jpg", "image/jpeg", false},
		{"Valid PNG", "test.png", "image/png", false},
		{"Valid PDF", "doc.pdf", "application/pdf", false},
		{"Valid DOCX", "report.docx", "application/vnd.openxmlformats-officedocument.wordprocessingml.document", false},
		{"Invalid MP3", "audio.mp3", "audio/mpeg", true},
		{"Invalid MP4", "video.mp4", "video/mp4", true},
		{"Mismatched type", "image.jpg", "application/pdf", true},
		{"No extension", "file", "image/jpeg", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateFileType(tt.fileName, tt.contentType)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestGeneratePresignedURL_AllSupportedImageTypes tests all supported image types
func TestGeneratePresignedURL_AllSupportedImageTypes(t *testing.T) {
	_, router := setupAttachmentHandler(t)

	imageTypes := []struct {
		ext         string
		contentType string
	}{
		{".jpg", "image/jpeg"},
		{".jpeg", "image/jpeg"},
		{".png", "image/png"},
		{".gif", "image/gif"},
		{".webp", "image/webp"},
	}

	for _, img := range imageTypes {
		t.Run("Image type "+img.ext, func(t *testing.T) {
			reqBody := PresignedURLRequest{
				EntityType:  "BOARD",
				WorkspaceID: "550e8400-e29b-41d4-a716-446655440000",
				FileName:    "test" + img.ext,
				FileSize:    1024000,
				ContentType: img.contentType,
			}

			body, err := json.Marshal(reqBody)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/attachments/presigned-url", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
		})
	}
}

// TestGeneratePresignedURL_AllSupportedDocTypes tests all supported document types
func TestGeneratePresignedURL_AllSupportedDocTypes(t *testing.T) {
	_, router := setupAttachmentHandler(t)

	docTypes := []struct {
		ext         string
		contentType string
	}{
		{".pdf", "application/pdf"},
		{".txt", "text/plain"},
		{".doc", "application/msword"},
		{".docx", "application/vnd.openxmlformats-officedocument.wordprocessingml.document"},
		{".xls", "application/vnd.ms-excel"},
		{".xlsx", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"},
		{".ppt", "application/vnd.ms-powerpoint"},
		{".pptx", "application/vnd.openxmlformats-officedocument.presentationml.presentation"},
	}

	for _, doc := range docTypes {
		t.Run("Document type "+doc.ext, func(t *testing.T) {
			reqBody := PresignedURLRequest{
				EntityType:  "COMMENT",
				WorkspaceID: "550e8400-e29b-41d4-a716-446655440000",
				FileName:    "document" + doc.ext,
				FileSize:    2048000,
				ContentType: doc.contentType,
			}

			body, err := json.Marshal(reqBody)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/attachments/presigned-url", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
		})
	}
}

// Benchmark tests
func BenchmarkGeneratePresignedURL(b *testing.B) {
	_, router := setupAttachmentHandler(&testing.T{})

	reqBody := PresignedURLRequest{
		EntityType:  "BOARD",
		WorkspaceID: "550e8400-e29b-41d4-a716-446655440000",
		FileName:    "test.jpg",
		FileSize:    1024000,
		ContentType: "image/jpeg",
	}

	body, _ := json.Marshal(reqBody)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/attachments/presigned-url", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

// Mock S3 client for testing without AWS credentials
type mockS3Client struct{}

func (m *mockS3Client) GeneratePresignedURL(ctx context.Context, entityType, workspaceID, fileName, contentType string) (string, string, error) {
	return "https://mock-presigned-url.com", "mock-file-key", nil
}

// setupAttachmentHandlerWithRepo creates a test handler with mock S3 client and repository
func setupAttachmentHandlerWithRepo(t *testing.T, mockRepo *mockAttachmentRepository) (*AttachmentHandler, *gin.Engine) {
	gin.SetMode(gin.TestMode)

	// Create S3 config for testing
	cfg := &config.S3Config{
		Bucket:    "test-bucket",
		Region:    "us-east-1",
		AccessKey: "test-access-key",
		SecretKey: "test-secret-key",
		Endpoint:  "",
	}

	// Create S3 client
	s3Client, err := client.NewS3Client(cfg)
	require.NoError(t, err, "Failed to create S3 client")

	// Use provided mock repo or create default one
	if mockRepo == nil {
		mockRepo = &mockAttachmentRepository{}
	}

	// Create handler
	handler := NewAttachmentHandler(s3Client, mockRepo)

	// Setup router with auth middleware
	router := gin.New()
	// Add middleware to set user_id in context (simulating auth middleware)
	router.Use(func(c *gin.Context) {
		c.Set("user_id", uuid.New())
		c.Next()
	})
	router.POST("/attachments", handler.SaveAttachmentMetadata)

	return handler, router
}


// TestSaveAttachmentMetadata_Success tests successful attachment metadata save
func TestSaveAttachmentMetadata_Success(t *testing.T) {
	tests := []struct {
		name        string
		entityType  string
		fileKey     string
		fileName    string
		fileSize    int64
		contentType string
	}{
		{
			name:        "Board image attachment",
			entityType:  "BOARD",
			fileKey:     "board/boards/550e8400-e29b-41d4-a716-446655440000/2024/01/test-uuid_1704067200.jpg",
			fileName:    "test-image.jpg",
			fileSize:    1024000,
			contentType: "image/jpeg",
		},
		{
			name:        "Comment PDF attachment",
			entityType:  "COMMENT",
			fileKey:     "board/comments/550e8400-e29b-41d4-a716-446655440000/2024/01/test-uuid_1704067200.pdf",
			fileName:    "document.pdf",
			fileSize:    2048000,
			contentType: "application/pdf",
		},
		{
			name:        "Project PNG attachment",
			entityType:  "PROJECT",
			fileKey:     "board/projects/550e8400-e29b-41d4-a716-446655440000/2024/01/test-uuid_1704067200.png",
			fileName:    "diagram.png",
			fileSize:    512000,
			contentType: "image/png",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, router := setupAttachmentHandlerWithRepo(t, nil)

			reqBody := SaveAttachmentMetadataRequest{
				EntityType:  tt.entityType,
				FileKey:     tt.fileKey,
				FileName:    tt.fileName,
				FileSize:    tt.fileSize,
				ContentType: tt.contentType,
			}

			body, err := json.Marshal(reqBody)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/attachments", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusCreated, w.Code)

			var response map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			// Check response structure
			data, ok := response["data"].(map[string]interface{})
			require.True(t, ok, "Response should contain data field")

			// Verify attachment fields
			assert.NotEmpty(t, data["id"], "ID should not be empty")
			assert.Equal(t, tt.entityType, data["entityType"], "Entity type should match")
			assert.Nil(t, data["entityId"], "EntityID should be nil for temporary attachment")
			assert.Equal(t, "TEMP", data["status"], "Status should be TEMP")
			assert.Equal(t, tt.fileName, data["fileName"], "File name should match")
			assert.NotEmpty(t, data["fileUrl"], "File URL should not be empty")
			assert.Equal(t, float64(tt.fileSize), data["fileSize"], "File size should match")
			assert.Equal(t, tt.contentType, data["contentType"], "Content type should match")
			assert.NotEmpty(t, data["uploadedBy"], "UploadedBy should not be empty")
			assert.NotEmpty(t, data["uploadedAt"], "UploadedAt should not be empty")
			assert.NotNil(t, data["expiresAt"], "ExpiresAt should not be nil")
		})
	}
}

// TestSaveAttachmentMetadata_InvalidFileKey tests invalid file key validation
func TestSaveAttachmentMetadata_InvalidFileKey(t *testing.T) {
	tests := []struct {
		name    string
		fileKey string
	}{
		{
			name:    "Empty file key",
			fileKey: "",
		},
		{
			name:    "Invalid prefix",
			fileKey: "user/abc123/2024/01/test.jpg",
		},
		{
			name:    "Missing prefix",
			fileKey: "test.jpg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, router := setupAttachmentHandlerWithRepo(t, nil)

			reqBody := SaveAttachmentMetadataRequest{
				EntityType:  "BOARD",
				FileKey:     tt.fileKey,
				FileName:    "test.jpg",
				FileSize:    1024000,
				ContentType: "image/jpeg",
			}

			body, err := json.Marshal(reqBody)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/attachments", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)

			var response map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			errorData, ok := response["error"].(map[string]interface{})
			require.True(t, ok, "Response should contain error field")

			code, ok := errorData["code"].(string)
			assert.True(t, ok)
			assert.Equal(t, "VALIDATION_ERROR", code)
		})
	}
}

// TestSaveAttachmentMetadata_DBError tests database save failure
func TestSaveAttachmentMetadata_DBError(t *testing.T) {
	mockRepo := &mockAttachmentRepository{
		createFunc: func(ctx context.Context, attachment *domain.Attachment) error {
			return fmt.Errorf("database error")
		},
	}

	_, router := setupAttachmentHandlerWithRepo(t, mockRepo)

	reqBody := SaveAttachmentMetadataRequest{
		EntityType:  "BOARD",
		FileKey:     "board/boards/550e8400-e29b-41d4-a716-446655440000/2024/01/test-uuid_1704067200.jpg",
		FileName:    "test.jpg",
		FileSize:    1024000,
		ContentType: "image/jpeg",
	}

	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/attachments", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	errorData, ok := response["error"].(map[string]interface{})
	require.True(t, ok, "Response should contain error field")

	code, ok := errorData["code"].(string)
	assert.True(t, ok)
	assert.Equal(t, "INTERNAL_ERROR", code)
}

// TestSaveAttachmentMetadata_Unauthorized tests missing user authentication
func TestSaveAttachmentMetadata_Unauthorized(t *testing.T) {
	// Create router WITHOUT auth middleware
	gin.SetMode(gin.TestMode)
	cfg := &config.S3Config{
		Bucket:    "test-bucket",
		Region:    "us-east-1",
		AccessKey: "test-access-key",
		SecretKey: "test-secret-key",
		Endpoint:  "",
	}
	s3Client, err := client.NewS3Client(cfg)
	require.NoError(t, err)
	mockRepo := &mockAttachmentRepository{}
	handler := NewAttachmentHandler(s3Client, mockRepo)
	router := gin.New()
	// No auth middleware - user_id will not be set
	router.POST("/attachments", handler.SaveAttachmentMetadata)

	reqBody := SaveAttachmentMetadataRequest{
		EntityType:  "BOARD",
		FileKey:     "board/boards/550e8400-e29b-41d4-a716-446655440000/2024/01/test-uuid_1704067200.jpg",
		FileName:    "test.jpg",
		FileSize:    1024000,
		ContentType: "image/jpeg",
	}

	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/attachments", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	errorData, ok := response["error"].(map[string]interface{})
	require.True(t, ok, "Response should contain error field")

	code, ok := errorData["code"].(string)
	assert.True(t, ok)
	assert.Equal(t, "UNAUTHORIZED", code)
}

// TestSaveAttachmentMetadata_FileSizeValidation tests file size validation
func TestSaveAttachmentMetadata_FileSizeValidation(t *testing.T) {
	tests := []struct {
		name     string
		fileSize int64
	}{
		{
			name:     "Zero file size",
			fileSize: 0,
		},
		{
			name:     "Negative file size",
			fileSize: -1,
		},
		{
			name:     "File size exceeds limit",
			fileSize: 21 * 1024 * 1024, // 21MB
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, router := setupAttachmentHandlerWithRepo(t, nil)

			reqBody := SaveAttachmentMetadataRequest{
				EntityType:  "BOARD",
				FileKey:     "board/boards/550e8400-e29b-41d4-a716-446655440000/2024/01/test-uuid_1704067200.jpg",
				FileName:    "test.jpg",
				FileSize:    tt.fileSize,
				ContentType: "image/jpeg",
			}

			body, err := json.Marshal(reqBody)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/attachments", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

// TestSaveAttachmentMetadata_InvalidFileType tests file type validation
func TestSaveAttachmentMetadata_InvalidFileType(t *testing.T) {
	tests := []struct {
		name        string
		fileName    string
		contentType string
	}{
		{
			name:        "Audio file",
			fileName:    "audio.mp3",
			contentType: "audio/mpeg",
		},
		{
			name:        "Video file",
			fileName:    "video.mp4",
			contentType: "video/mp4",
		},
		{
			name:        "Unsupported type",
			fileName:    "archive.zip",
			contentType: "application/zip",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, router := setupAttachmentHandlerWithRepo(t, nil)

			reqBody := SaveAttachmentMetadataRequest{
				EntityType:  "BOARD",
				FileKey:     "board/boards/550e8400-e29b-41d4-a716-446655440000/2024/01/test-uuid_1704067200" + tt.fileName[strings.LastIndex(tt.fileName, "."):],
				FileName:    tt.fileName,
				FileSize:    1024000,
				ContentType: tt.contentType,
			}

			body, err := json.Marshal(reqBody)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/attachments", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)

			var response map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			errorData, ok := response["error"].(map[string]interface{})
			require.True(t, ok, "Response should contain error field")

			code, ok := errorData["code"].(string)
			assert.True(t, ok)
			assert.Equal(t, "INVALID_FILE_TYPE", code)
		})
	}
}

// TestSaveAttachmentMetadata_InvalidEntityType tests entity type validation
func TestSaveAttachmentMetadata_InvalidEntityType(t *testing.T) {
	tests := []struct {
		name       string
		entityType string
	}{
		{
			name:       "Invalid entity type - USER",
			entityType: "USER",
		},
		{
			name:       "Invalid entity type - WORKSPACE",
			entityType: "WORKSPACE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, router := setupAttachmentHandlerWithRepo(t, nil)

			reqBody := SaveAttachmentMetadataRequest{
				EntityType:  tt.entityType,
				FileKey:     "board/boards/550e8400-e29b-41d4-a716-446655440000/2024/01/test-uuid_1704067200.jpg",
				FileName:    "test.jpg",
				FileSize:    1024000,
				ContentType: "image/jpeg",
			}

			body, err := json.Marshal(reqBody)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/attachments", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)

			var response map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			errorData, ok := response["error"].(map[string]interface{})
			require.True(t, ok, "Response should contain error field")

			code, ok := errorData["code"].(string)
			assert.True(t, ok)
			assert.Equal(t, "VALIDATION_ERROR", code)
		})
	}
}

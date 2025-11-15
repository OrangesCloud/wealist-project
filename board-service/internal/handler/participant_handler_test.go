package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	"project-board-api/internal/dto"
	"project-board-api/internal/response"
)

// MockParticipantService is a mock implementation of ParticipantService
type MockParticipantService struct {
	AddParticipantFunc    func(ctx context.Context, req *dto.AddParticipantRequest) error
	GetParticipantsFunc   func(ctx context.Context, boardID uuid.UUID) ([]*dto.ParticipantResponse, error)
	RemoveParticipantFunc func(ctx context.Context, boardID, userID uuid.UUID) error
}

func (m *MockParticipantService) AddParticipant(ctx context.Context, req *dto.AddParticipantRequest) error {
	if m.AddParticipantFunc != nil {
		return m.AddParticipantFunc(ctx, req)
	}
	return nil
}

func (m *MockParticipantService) GetParticipants(ctx context.Context, boardID uuid.UUID) ([]*dto.ParticipantResponse, error) {
	if m.GetParticipantsFunc != nil {
		return m.GetParticipantsFunc(ctx, boardID)
	}
	return nil, nil
}

func (m *MockParticipantService) RemoveParticipant(ctx context.Context, boardID, userID uuid.UUID) error {
	if m.RemoveParticipantFunc != nil {
		return m.RemoveParticipantFunc(ctx, boardID, userID)
	}
	return nil
}

func TestParticipantHandler_AddParticipant(t *testing.T) {
	boardID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name           string
		requestBody    interface{}
		mockService    func(*MockParticipantService)
		expectedStatus int
	}{
		{
			name: "성공: 참여자 추가",
			requestBody: dto.AddParticipantRequest{
				BoardID: boardID,
				UserID:  userID,
			},
			mockService: func(m *MockParticipantService) {
				m.AddParticipantFunc = func(ctx context.Context, req *dto.AddParticipantRequest) error {
					return nil
				}
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "실패: 잘못된 요청 본문",
			requestBody:    "invalid json",
			mockService:    func(m *MockParticipantService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "실패: Board가 존재하지 않음",
			requestBody: dto.AddParticipantRequest{
				BoardID: boardID,
				UserID:  userID,
			},
			mockService: func(m *MockParticipantService) {
				m.AddParticipantFunc = func(ctx context.Context, req *dto.AddParticipantRequest) error {
					return response.NewAppError(response.ErrCodeNotFound, "Board not found", "")
				}
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "실패: 이미 참여 중인 사용자",
			requestBody: dto.AddParticipantRequest{
				BoardID: boardID,
				UserID:  userID,
			},
			mockService: func(m *MockParticipantService) {
				m.AddParticipantFunc = func(ctx context.Context, req *dto.AddParticipantRequest) error {
					return response.NewAppError(response.ErrCodeAlreadyExists, "Participant already exists", "")
				}
			},
			expectedStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			mockService := &MockParticipantService{}
			tt.mockService(mockService)
			handler := NewParticipantHandler(mockService)

			router := setupTestRouter()
			router.POST("/api/participants", handler.AddParticipant)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/participants", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// When
			router.ServeHTTP(w, req)

			// Then
			if w.Code != tt.expectedStatus {
				t.Errorf("AddParticipant() status = %v, want %v", w.Code, tt.expectedStatus)
			}
		})
	}
}

func TestParticipantHandler_GetParticipants(t *testing.T) {
	boardID := uuid.New()

	tests := []struct {
		name           string
		boardID        string
		mockService    func(*MockParticipantService)
		expectedStatus int
	}{
		{
			name:    "성공: 참여자 목록 조회",
			boardID: boardID.String(),
			mockService: func(m *MockParticipantService) {
				m.GetParticipantsFunc = func(ctx context.Context, id uuid.UUID) ([]*dto.ParticipantResponse, error) {
					return []*dto.ParticipantResponse{
						{
							ID:      uuid.New(),
							BoardID: id,
							UserID:  uuid.New(),
						},
						{
							ID:      uuid.New(),
							BoardID: id,
							UserID:  uuid.New(),
						},
					}, nil
				}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "실패: 잘못된 UUID",
			boardID:        "invalid-uuid",
			mockService:    func(m *MockParticipantService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:    "실패: Board가 존재하지 않음",
			boardID: boardID.String(),
			mockService: func(m *MockParticipantService) {
				m.GetParticipantsFunc = func(ctx context.Context, id uuid.UUID) ([]*dto.ParticipantResponse, error) {
					return nil, response.NewAppError(response.ErrCodeNotFound, "Board not found", "")
				}
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			mockService := &MockParticipantService{}
			tt.mockService(mockService)
			handler := NewParticipantHandler(mockService)

			router := setupTestRouter()
			router.GET("/api/participants/board/:boardId", handler.GetParticipants)

			req := httptest.NewRequest(http.MethodGet, "/api/participants/board/"+tt.boardID, nil)
			w := httptest.NewRecorder()

			// When
			router.ServeHTTP(w, req)

			// Then
			if w.Code != tt.expectedStatus {
				t.Errorf("GetParticipants() status = %v, want %v", w.Code, tt.expectedStatus)
			}
		})
	}
}

func TestParticipantHandler_RemoveParticipant(t *testing.T) {
	boardID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name           string
		boardID        string
		userID         string
		mockService    func(*MockParticipantService)
		expectedStatus int
	}{
		{
			name:    "성공: 참여자 제거",
			boardID: boardID.String(),
			userID:  userID.String(),
			mockService: func(m *MockParticipantService) {
				m.RemoveParticipantFunc = func(ctx context.Context, bID, uID uuid.UUID) error {
					return nil
				}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "실패: 잘못된 Board UUID",
			boardID:        "invalid-uuid",
			userID:         userID.String(),
			mockService:    func(m *MockParticipantService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "실패: 잘못된 User UUID",
			boardID:        boardID.String(),
			userID:         "invalid-uuid",
			mockService:    func(m *MockParticipantService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:    "실패: 참여자가 존재하지 않음",
			boardID: boardID.String(),
			userID:  userID.String(),
			mockService: func(m *MockParticipantService) {
				m.RemoveParticipantFunc = func(ctx context.Context, bID, uID uuid.UUID) error {
					return response.NewAppError(response.ErrCodeNotFound, "Participant not found", "")
				}
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			mockService := &MockParticipantService{}
			tt.mockService(mockService)
			handler := NewParticipantHandler(mockService)

			router := setupTestRouter()
			router.DELETE("/api/participants/board/:boardId/user/:userId", handler.RemoveParticipant)

			req := httptest.NewRequest(http.MethodDelete, "/api/participants/board/"+tt.boardID+"/user/"+tt.userID, nil)
			w := httptest.NewRecorder()

			// When
			router.ServeHTTP(w, req)

			// Then
			if w.Code != tt.expectedStatus {
				t.Errorf("RemoveParticipant() status = %v, want %v", w.Code, tt.expectedStatus)
			}
		})
	}
}

package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"project-board-api/internal/domain"
	"project-board-api/internal/dto"
	"project-board-api/internal/response"
)

// MockParticipantRepository is a mock implementation of ParticipantRepository
type MockParticipantRepository struct {
	CreateFunc             func(ctx context.Context, participant *domain.Participant) error
	FindByBoardIDFunc      func(ctx context.Context, boardID uuid.UUID) ([]*domain.Participant, error)
	FindByBoardAndUserFunc func(ctx context.Context, boardID, userID uuid.UUID) (*domain.Participant, error)
	DeleteFunc             func(ctx context.Context, boardID, userID uuid.UUID) error
}

func (m *MockParticipantRepository) Create(ctx context.Context, participant *domain.Participant) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, participant)
	}
	return nil
}

func (m *MockParticipantRepository) FindByBoardID(ctx context.Context, boardID uuid.UUID) ([]*domain.Participant, error) {
	if m.FindByBoardIDFunc != nil {
		return m.FindByBoardIDFunc(ctx, boardID)
	}
	return nil, nil
}

func (m *MockParticipantRepository) FindByBoardAndUser(ctx context.Context, boardID, userID uuid.UUID) (*domain.Participant, error) {
	if m.FindByBoardAndUserFunc != nil {
		return m.FindByBoardAndUserFunc(ctx, boardID, userID)
	}
	return nil, nil
}

func (m *MockParticipantRepository) Delete(ctx context.Context, boardID, userID uuid.UUID) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, boardID, userID)
	}
	return nil
}

func TestParticipantService_AddParticipant(t *testing.T) {
	boardID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name            string
		req             *dto.AddParticipantRequest
		mockBoard       func(*MockBoardRepository)
		mockParticipant func(*MockParticipantRepository)
		wantErr         bool
		wantErrCode     string
	}{
		{
			name: "성공: 정상적인 Participant 추가",
			req: &dto.AddParticipantRequest{
				BoardID: boardID,
				UserID:  userID,
			},
			mockBoard: func(m *MockBoardRepository) {
				m.FindByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.Board, error) {
					return &domain.Board{}, nil
				}
			},
			mockParticipant: func(m *MockParticipantRepository) {
				m.FindByBoardAndUserFunc = func(ctx context.Context, bID, uID uuid.UUID) (*domain.Participant, error) {
					return nil, gorm.ErrRecordNotFound
				}
				m.CreateFunc = func(ctx context.Context, participant *domain.Participant) error {
					participant.ID = uuid.New()
					participant.CreatedAt = time.Now()
					participant.UpdatedAt = time.Now()
					return nil
				}
			},
			wantErr: false,
		},
		{
			name: "실패: Board가 존재하지 않음",
			req: &dto.AddParticipantRequest{
				BoardID: boardID,
				UserID:  userID,
			},
			mockBoard: func(m *MockBoardRepository) {
				m.FindByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.Board, error) {
					return nil, gorm.ErrRecordNotFound
				}
			},
			mockParticipant: func(m *MockParticipantRepository) {},
			wantErr:         true,
			wantErrCode:     response.ErrCodeNotFound,
		},
		{
			name: "실패: 이미 참여 중인 User",
			req: &dto.AddParticipantRequest{
				BoardID: boardID,
				UserID:  userID,
			},
			mockBoard: func(m *MockBoardRepository) {
				m.FindByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.Board, error) {
					return &domain.Board{}, nil
				}
			},
			mockParticipant: func(m *MockParticipantRepository) {
				m.FindByBoardAndUserFunc = func(ctx context.Context, bID, uID uuid.UUID) (*domain.Participant, error) {
					return &domain.Participant{
						BaseModel: domain.BaseModel{ID: uuid.New()},
						BoardID:   boardID,
						UserID:    userID,
					}, nil
				}
			},
			wantErr:     true,
			wantErrCode: response.ErrCodeAlreadyExists,
		},
		{
			name: "실패: Unique constraint 위반",
			req: &dto.AddParticipantRequest{
				BoardID: boardID,
				UserID:  userID,
			},
			mockBoard: func(m *MockBoardRepository) {
				m.FindByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.Board, error) {
					return &domain.Board{}, nil
				}
			},
			mockParticipant: func(m *MockParticipantRepository) {
				m.FindByBoardAndUserFunc = func(ctx context.Context, bID, uID uuid.UUID) (*domain.Participant, error) {
					return nil, gorm.ErrRecordNotFound
				}
				m.CreateFunc = func(ctx context.Context, participant *domain.Participant) error {
					return errors.New("duplicate key value violates unique constraint")
				}
			},
			wantErr:     true,
			wantErrCode: response.ErrCodeAlreadyExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			mockBoardRepo := &MockBoardRepository{}
			mockParticipantRepo := &MockParticipantRepository{}
			tt.mockBoard(mockBoardRepo)
			tt.mockParticipant(mockParticipantRepo)

			service := NewParticipantService(mockParticipantRepo, mockBoardRepo)

			// When
			err := service.AddParticipant(context.Background(), tt.req)

			// Then
			if tt.wantErr {
				if err == nil {
					t.Errorf("AddParticipant() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if appErr, ok := err.(*response.AppError); ok {
					if appErr.Code != tt.wantErrCode {
						t.Errorf("AddParticipant() error code = %v, want %v", appErr.Code, tt.wantErrCode)
					}
				}
			} else {
				if err != nil {
					t.Errorf("AddParticipant() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestParticipantService_GetParticipants(t *testing.T) {
	boardID := uuid.New()

	tests := []struct {
		name            string
		boardID         uuid.UUID
		mockBoard       func(*MockBoardRepository)
		mockParticipant func(*MockParticipantRepository)
		wantErr         bool
		wantErrCode     string
		wantCount       int
	}{
		{
			name:    "성공: Participant 목록 조회",
			boardID: boardID,
			mockBoard: func(m *MockBoardRepository) {
				m.FindByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.Board, error) {
					return &domain.Board{}, nil
				}
			},
			mockParticipant: func(m *MockParticipantRepository) {
				m.FindByBoardIDFunc = func(ctx context.Context, bID uuid.UUID) ([]*domain.Participant, error) {
					return []*domain.Participant{
						{
							BaseModel: domain.BaseModel{ID: uuid.New(), CreatedAt: time.Now(), UpdatedAt: time.Now()},
							BoardID:   boardID,
							UserID:    uuid.New(),
						},
						{
							BaseModel: domain.BaseModel{ID: uuid.New(), CreatedAt: time.Now(), UpdatedAt: time.Now()},
							BoardID:   boardID,
							UserID:    uuid.New(),
						},
					}, nil
				}
			},
			wantErr:   false,
			wantCount: 2,
		},
		{
			name:    "성공: 빈 Participant 목록",
			boardID: boardID,
			mockBoard: func(m *MockBoardRepository) {
				m.FindByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.Board, error) {
					return &domain.Board{}, nil
				}
			},
			mockParticipant: func(m *MockParticipantRepository) {
				m.FindByBoardIDFunc = func(ctx context.Context, bID uuid.UUID) ([]*domain.Participant, error) {
					return []*domain.Participant{}, nil
				}
			},
			wantErr:   false,
			wantCount: 0,
		},
		{
			name:    "실패: Board가 존재하지 않음",
			boardID: boardID,
			mockBoard: func(m *MockBoardRepository) {
				m.FindByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.Board, error) {
					return nil, gorm.ErrRecordNotFound
				}
			},
			mockParticipant: func(m *MockParticipantRepository) {},
			wantErr:         true,
			wantErrCode:     response.ErrCodeNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			mockBoardRepo := &MockBoardRepository{}
			mockParticipantRepo := &MockParticipantRepository{}
			tt.mockBoard(mockBoardRepo)
			tt.mockParticipant(mockParticipantRepo)

			service := NewParticipantService(mockParticipantRepo, mockBoardRepo)

			// When
			got, err := service.GetParticipants(context.Background(), tt.boardID)

			// Then
			if tt.wantErr {
				if err == nil {
					t.Errorf("GetParticipants() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if appErr, ok := err.(*response.AppError); ok {
					if appErr.Code != tt.wantErrCode {
						t.Errorf("GetParticipants() error code = %v, want %v", appErr.Code, tt.wantErrCode)
					}
				}
			} else {
				if err != nil {
					t.Errorf("GetParticipants() unexpected error = %v", err)
					return
				}
				if got == nil {
					t.Error("GetParticipants() returned nil response")
					return
				}
				if len(got) != tt.wantCount {
					t.Errorf("GetParticipants() count = %v, want %v", len(got), tt.wantCount)
				}
			}
		})
	}
}

func TestParticipantService_RemoveParticipant(t *testing.T) {
	boardID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name            string
		boardID         uuid.UUID
		userID          uuid.UUID
		mockBoard       func(*MockBoardRepository)
		mockParticipant func(*MockParticipantRepository)
		wantErr         bool
		wantErrCode     string
	}{
		{
			name:    "성공: Participant 제거",
			boardID: boardID,
			userID:  userID,
			mockBoard: func(m *MockBoardRepository) {
				m.FindByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.Board, error) {
					return &domain.Board{}, nil
				}
			},
			mockParticipant: func(m *MockParticipantRepository) {
				m.FindByBoardAndUserFunc = func(ctx context.Context, bID, uID uuid.UUID) (*domain.Participant, error) {
					return &domain.Participant{
						BaseModel: domain.BaseModel{ID: uuid.New()},
						BoardID:   boardID,
						UserID:    userID,
					}, nil
				}
				m.DeleteFunc = func(ctx context.Context, bID, uID uuid.UUID) error {
					return nil
				}
			},
			wantErr: false,
		},
		{
			name:    "실패: Board가 존재하지 않음",
			boardID: boardID,
			userID:  userID,
			mockBoard: func(m *MockBoardRepository) {
				m.FindByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.Board, error) {
					return nil, gorm.ErrRecordNotFound
				}
			},
			mockParticipant: func(m *MockParticipantRepository) {},
			wantErr:         true,
			wantErrCode:     response.ErrCodeNotFound,
		},
		{
			name:    "실패: Participant가 존재하지 않음",
			boardID: boardID,
			userID:  userID,
			mockBoard: func(m *MockBoardRepository) {
				m.FindByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.Board, error) {
					return &domain.Board{}, nil
				}
			},
			mockParticipant: func(m *MockParticipantRepository) {
				m.FindByBoardAndUserFunc = func(ctx context.Context, bID, uID uuid.UUID) (*domain.Participant, error) {
					return nil, gorm.ErrRecordNotFound
				}
			},
			wantErr:     true,
			wantErrCode: response.ErrCodeNotFound,
		},
		{
			name:    "실패: Participant 삭제 중 DB 에러",
			boardID: boardID,
			userID:  userID,
			mockBoard: func(m *MockBoardRepository) {
				m.FindByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.Board, error) {
					return &domain.Board{}, nil
				}
			},
			mockParticipant: func(m *MockParticipantRepository) {
				m.FindByBoardAndUserFunc = func(ctx context.Context, bID, uID uuid.UUID) (*domain.Participant, error) {
					return &domain.Participant{
						BaseModel: domain.BaseModel{ID: uuid.New()},
						BoardID:   boardID,
						UserID:    userID,
					}, nil
				}
				m.DeleteFunc = func(ctx context.Context, bID, uID uuid.UUID) error {
					return errors.New("database error")
				}
			},
			wantErr:     true,
			wantErrCode: response.ErrCodeInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			mockBoardRepo := &MockBoardRepository{}
			mockParticipantRepo := &MockParticipantRepository{}
			tt.mockBoard(mockBoardRepo)
			tt.mockParticipant(mockParticipantRepo)

			service := NewParticipantService(mockParticipantRepo, mockBoardRepo)

			// When
			err := service.RemoveParticipant(context.Background(), tt.boardID, tt.userID)

			// Then
			if tt.wantErr {
				if err == nil {
					t.Errorf("RemoveParticipant() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if appErr, ok := err.(*response.AppError); ok {
					if appErr.Code != tt.wantErrCode {
						t.Errorf("RemoveParticipant() error code = %v, want %v", appErr.Code, tt.wantErrCode)
					}
				}
			} else {
				if err != nil {
					t.Errorf("RemoveParticipant() unexpected error = %v", err)
				}
			}
		})
	}
}

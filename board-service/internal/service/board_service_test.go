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

// MockBoardRepository is a mock implementation of BoardRepository
type MockBoardRepository struct {
	CreateFunc         func(ctx context.Context, board *domain.Board) error
	FindByIDFunc       func(ctx context.Context, id uuid.UUID) (*domain.Board, error)
	FindByProjectIDFunc func(ctx context.Context, projectID uuid.UUID) ([]*domain.Board, error)
	UpdateFunc         func(ctx context.Context, board *domain.Board) error
	DeleteFunc         func(ctx context.Context, id uuid.UUID) error
}

func (m *MockBoardRepository) Create(ctx context.Context, board *domain.Board) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, board)
	}
	return nil
}

func (m *MockBoardRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Board, error) {
	if m.FindByIDFunc != nil {
		return m.FindByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockBoardRepository) FindByProjectID(ctx context.Context, projectID uuid.UUID) ([]*domain.Board, error) {
	if m.FindByProjectIDFunc != nil {
		return m.FindByProjectIDFunc(ctx, projectID)
	}
	return nil, nil
}

func (m *MockBoardRepository) Update(ctx context.Context, board *domain.Board) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, board)
	}
	return nil
}

func (m *MockBoardRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

// MockProjectRepository is a mock implementation of ProjectRepository
type MockProjectRepository struct {
	CreateFunc                    func(ctx context.Context, project *domain.Project) error
	FindByIDFunc                  func(ctx context.Context, id uuid.UUID) (*domain.Project, error)
	FindByWorkspaceIDFunc         func(ctx context.Context, workspaceID uuid.UUID) ([]*domain.Project, error)
	FindDefaultByWorkspaceIDFunc  func(ctx context.Context, workspaceID uuid.UUID) (*domain.Project, error)
	UpdateFunc                    func(ctx context.Context, project *domain.Project) error
	DeleteFunc                    func(ctx context.Context, id uuid.UUID) error
}

func (m *MockProjectRepository) Create(ctx context.Context, project *domain.Project) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, project)
	}
	return nil
}

func (m *MockProjectRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Project, error) {
	if m.FindByIDFunc != nil {
		return m.FindByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockProjectRepository) FindByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) ([]*domain.Project, error) {
	if m.FindByWorkspaceIDFunc != nil {
		return m.FindByWorkspaceIDFunc(ctx, workspaceID)
	}
	return nil, nil
}

func (m *MockProjectRepository) FindDefaultByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) (*domain.Project, error) {
	if m.FindDefaultByWorkspaceIDFunc != nil {
		return m.FindDefaultByWorkspaceIDFunc(ctx, workspaceID)
	}
	return nil, nil
}

func (m *MockProjectRepository) Update(ctx context.Context, project *domain.Project) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, project)
	}
	return nil
}

func (m *MockProjectRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func (m *MockProjectRepository) Search(ctx context.Context, workspaceID uuid.UUID, query string, page, limit int) ([]*domain.Project, int64, error) {
	return nil, 0, nil
}

func (m *MockProjectRepository) AddMember(ctx context.Context, member *domain.ProjectMember) error {
	return nil
}

func (m *MockProjectRepository) FindMembersByProjectID(ctx context.Context, projectID uuid.UUID) ([]*domain.ProjectMember, error) {
	return nil, nil
}

func (m *MockProjectRepository) FindMemberByProjectAndUser(ctx context.Context, projectID, userID uuid.UUID) (*domain.ProjectMember, error) {
	return nil, nil
}

func (m *MockProjectRepository) UpdateMemberRole(ctx context.Context, memberID uuid.UUID, role domain.ProjectRole) error {
	return nil
}

func (m *MockProjectRepository) RemoveMember(ctx context.Context, memberID uuid.UUID) error {
	return nil
}

func (m *MockProjectRepository) CreateJoinRequest(ctx context.Context, request *domain.ProjectJoinRequest) error {
	return nil
}

func (m *MockProjectRepository) FindJoinRequestsByProjectID(ctx context.Context, projectID uuid.UUID, status *domain.ProjectJoinRequestStatus) ([]*domain.ProjectJoinRequest, error) {
	return nil, nil
}

func (m *MockProjectRepository) FindJoinRequestByID(ctx context.Context, requestID uuid.UUID) (*domain.ProjectJoinRequest, error) {
	return nil, nil
}

func (m *MockProjectRepository) UpdateJoinRequestStatus(ctx context.Context, requestID uuid.UUID, status domain.ProjectJoinRequestStatus) error {
	return nil
}

func TestBoardService_CreateBoard(t *testing.T) {
	projectID := uuid.New()
	
	tests := []struct {
		name          string
		req           *dto.CreateBoardRequest
		mockProject   func(*MockProjectRepository)
		mockBoard     func(*MockBoardRepository)
		wantErr       bool
		wantErrCode   string
	}{
		{
			name: "성공: 정상적인 Board 생성",
			req: &dto.CreateBoardRequest{
				ProjectID:  projectID,
				Title:      "Test Board",
				Content:    "Test Content",
				Stage:      "in_progress",
				Importance: "urgent",
				Role:       "developer",
			},
			mockProject: func(m *MockProjectRepository) {
				m.FindByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.Project, error) {
					return &domain.Project{}, nil
				}
			},
			mockBoard: func(m *MockBoardRepository) {
				m.CreateFunc = func(ctx context.Context, board *domain.Board) error {
					board.ID = uuid.New()
					board.CreatedAt = time.Now()
					board.UpdatedAt = time.Now()
					return nil
				}
			},
			wantErr: false,
		},
		{
			name: "실패: Project가 존재하지 않음",
			req: &dto.CreateBoardRequest{
				ProjectID:  projectID,
				Title:      "Test Board",
				Content:    "Test Content",
				Stage:      "in_progress",
				Importance: "urgent",
				Role:       "developer",
			},
			mockProject: func(m *MockProjectRepository) {
				m.FindByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.Project, error) {
					return nil, gorm.ErrRecordNotFound
				}
			},
			mockBoard: func(m *MockBoardRepository) {},
			wantErr:     true,
			wantErrCode: response.ErrCodeNotFound,
		},
		{
			name: "실패: Board 생성 중 DB 에러",
			req: &dto.CreateBoardRequest{
				ProjectID:  projectID,
				Title:      "Test Board",
				Content:    "Test Content",
				Stage:      "in_progress",
				Importance: "urgent",
				Role:       "developer",
			},
			mockProject: func(m *MockProjectRepository) {
				m.FindByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.Project, error) {
					return &domain.Project{}, nil
				}
			},
			mockBoard: func(m *MockBoardRepository) {
				m.CreateFunc = func(ctx context.Context, board *domain.Board) error {
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
			mockProjectRepo := &MockProjectRepository{}
			mockBoardRepo := &MockBoardRepository{}
			tt.mockProject(mockProjectRepo)
			tt.mockBoard(mockBoardRepo)
			
			service := NewBoardService(mockBoardRepo, mockProjectRepo)

			// When
			got, err := service.CreateBoard(context.Background(), tt.req)

			// Then
			if tt.wantErr {
				if err == nil {
					t.Errorf("CreateBoard() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if appErr, ok := err.(*response.AppError); ok {
					if appErr.Code != tt.wantErrCode {
						t.Errorf("CreateBoard() error code = %v, want %v", appErr.Code, tt.wantErrCode)
					}
				}
			} else {
				if err != nil {
					t.Errorf("CreateBoard() unexpected error = %v", err)
					return
				}
				if got == nil {
					t.Error("CreateBoard() returned nil response")
					return
				}
				if got.Title != tt.req.Title {
					t.Errorf("CreateBoard() Title = %v, want %v", got.Title, tt.req.Title)
				}
			}
		})
	}
}

func TestBoardService_GetBoard(t *testing.T) {
	boardID := uuid.New()
	
	tests := []struct {
		name        string
		boardID     uuid.UUID
		mockBoard   func(*MockBoardRepository)
		wantErr     bool
		wantErrCode string
	}{
		{
			name:    "성공: Board 조회",
			boardID: boardID,
			mockBoard: func(m *MockBoardRepository) {
				m.FindByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.Board, error) {
					return &domain.Board{
						BaseModel: domain.BaseModel{
							ID:        boardID,
							CreatedAt: time.Now(),
							UpdatedAt: time.Now(),
						},
						Title:        "Test Board",
						Content:      "Test Content",
						Stage:        domain.StageInProgress,
						Importance:   domain.ImportanceUrgent,
						Role:         domain.RoleDeveloper,
						Participants: []domain.Participant{},
						Comments:     []domain.Comment{},
					}, nil
				}
			},
			wantErr: false,
		},
		{
			name:    "실패: Board가 존재하지 않음",
			boardID: boardID,
			mockBoard: func(m *MockBoardRepository) {
				m.FindByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.Board, error) {
					return nil, gorm.ErrRecordNotFound
				}
			},
			wantErr:     true,
			wantErrCode: response.ErrCodeNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			mockBoardRepo := &MockBoardRepository{}
			mockProjectRepo := &MockProjectRepository{}
			tt.mockBoard(mockBoardRepo)
			
			service := NewBoardService(mockBoardRepo, mockProjectRepo)

			// When
			got, err := service.GetBoard(context.Background(), tt.boardID)

			// Then
			if tt.wantErr {
				if err == nil {
					t.Errorf("GetBoard() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if appErr, ok := err.(*response.AppError); ok {
					if appErr.Code != tt.wantErrCode {
						t.Errorf("GetBoard() error code = %v, want %v", appErr.Code, tt.wantErrCode)
					}
				}
			} else {
				if err != nil {
					t.Errorf("GetBoard() unexpected error = %v", err)
					return
				}
				if got == nil {
					t.Error("GetBoard() returned nil response")
				}
			}
		})
	}
}

func TestBoardService_UpdateBoard(t *testing.T) {
	boardID := uuid.New()
	newTitle := "Updated Title"
	newStage := "approved"
	
	tests := []struct {
		name        string
		boardID     uuid.UUID
		req         *dto.UpdateBoardRequest
		mockBoard   func(*MockBoardRepository)
		wantErr     bool
		wantErrCode string
	}{
		{
			name:    "성공: Board 업데이트",
			boardID: boardID,
			req: &dto.UpdateBoardRequest{
				Title: &newTitle,
				Stage: &newStage,
			},
			mockBoard: func(m *MockBoardRepository) {
				m.FindByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.Board, error) {
					return &domain.Board{
						BaseModel: domain.BaseModel{
							ID:        boardID,
							CreatedAt: time.Now(),
							UpdatedAt: time.Now(),
						},
						Title:      "Old Title",
						Stage:      domain.StageInProgress,
						Importance: domain.ImportanceUrgent,
						Role:       domain.RoleDeveloper,
					}, nil
				}
				m.UpdateFunc = func(ctx context.Context, board *domain.Board) error {
					return nil
				}
			},
			wantErr: false,
		},
		{
			name:    "실패: Board가 존재하지 않음",
			boardID: boardID,
			req: &dto.UpdateBoardRequest{
				Title: &newTitle,
			},
			mockBoard: func(m *MockBoardRepository) {
				m.FindByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.Board, error) {
					return nil, gorm.ErrRecordNotFound
				}
			},
			wantErr:     true,
			wantErrCode: response.ErrCodeNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			mockBoardRepo := &MockBoardRepository{}
			mockProjectRepo := &MockProjectRepository{}
			tt.mockBoard(mockBoardRepo)
			
			service := NewBoardService(mockBoardRepo, mockProjectRepo)

			// When
			got, err := service.UpdateBoard(context.Background(), tt.boardID, tt.req)

			// Then
			if tt.wantErr {
				if err == nil {
					t.Errorf("UpdateBoard() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if appErr, ok := err.(*response.AppError); ok {
					if appErr.Code != tt.wantErrCode {
						t.Errorf("UpdateBoard() error code = %v, want %v", appErr.Code, tt.wantErrCode)
					}
				}
			} else {
				if err != nil {
					t.Errorf("UpdateBoard() unexpected error = %v", err)
					return
				}
				if got == nil {
					t.Error("UpdateBoard() returned nil response")
					return
				}
				if tt.req.Title != nil && got.Title != *tt.req.Title {
					t.Errorf("UpdateBoard() Title = %v, want %v", got.Title, *tt.req.Title)
				}
			}
		})
	}
}

func TestBoardService_DeleteBoard(t *testing.T) {
	boardID := uuid.New()
	
	tests := []struct {
		name        string
		boardID     uuid.UUID
		mockBoard   func(*MockBoardRepository)
		wantErr     bool
		wantErrCode string
	}{
		{
			name:    "성공: Board 삭제",
			boardID: boardID,
			mockBoard: func(m *MockBoardRepository) {
				m.FindByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.Board, error) {
					return &domain.Board{
						BaseModel: domain.BaseModel{ID: boardID},
					}, nil
				}
				m.DeleteFunc = func(ctx context.Context, id uuid.UUID) error {
					return nil
				}
			},
			wantErr: false,
		},
		{
			name:    "실패: Board가 존재하지 않음",
			boardID: boardID,
			mockBoard: func(m *MockBoardRepository) {
				m.FindByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.Board, error) {
					return nil, gorm.ErrRecordNotFound
				}
			},
			wantErr:     true,
			wantErrCode: response.ErrCodeNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			mockBoardRepo := &MockBoardRepository{}
			mockProjectRepo := &MockProjectRepository{}
			tt.mockBoard(mockBoardRepo)
			
			service := NewBoardService(mockBoardRepo, mockProjectRepo)

			// When
			err := service.DeleteBoard(context.Background(), tt.boardID)

			// Then
			if tt.wantErr {
				if err == nil {
					t.Errorf("DeleteBoard() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if appErr, ok := err.(*response.AppError); ok {
					if appErr.Code != tt.wantErrCode {
						t.Errorf("DeleteBoard() error code = %v, want %v", appErr.Code, tt.wantErrCode)
					}
				}
			} else {
				if err != nil {
					t.Errorf("DeleteBoard() unexpected error = %v", err)
				}
			}
		})
	}
}

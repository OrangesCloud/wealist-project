package service

import (
	"context"
	"encoding/json"
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
	FindByProjectIDFunc func(ctx context.Context, projectID uuid.UUID, filters interface{}) ([]*domain.Board, error)
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

func (m *MockBoardRepository) FindByProjectID(ctx context.Context, projectID uuid.UUID, filters interface{}) ([]*domain.Board, error) {
	if m.FindByProjectIDFunc != nil {
		return m.FindByProjectIDFunc(ctx, projectID, filters)
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
	CreateFunc                       func(ctx context.Context, project *domain.Project) error
	FindByIDFunc                     func(ctx context.Context, id uuid.UUID) (*domain.Project, error)
	FindByWorkspaceIDFunc            func(ctx context.Context, workspaceID uuid.UUID) ([]*domain.Project, error)
	FindDefaultByWorkspaceIDFunc     func(ctx context.Context, workspaceID uuid.UUID) (*domain.Project, error)
	UpdateFunc                       func(ctx context.Context, project *domain.Project) error
	DeleteFunc                       func(ctx context.Context, id uuid.UUID) error
	SearchFunc                       func(ctx context.Context, workspaceID uuid.UUID, query string, page, limit int) ([]*domain.Project, int64, error)
	AddMemberFunc                    func(ctx context.Context, member *domain.ProjectMember) error
	FindMemberByProjectAndUserFunc   func(ctx context.Context, projectID, userID uuid.UUID) (*domain.ProjectMember, error)
	RemoveMemberFunc                 func(ctx context.Context, memberID uuid.UUID) error
	UpdateMemberRoleFunc             func(ctx context.Context, memberID uuid.UUID, role domain.ProjectRole) error
	IsProjectMemberFunc              func(ctx context.Context, projectID, userID uuid.UUID) (bool, error)
	FindMembersByProjectIDFunc       func(ctx context.Context, projectID uuid.UUID) ([]*domain.ProjectMember, error)
	CreateJoinRequestFunc            func(ctx context.Context, request *domain.ProjectJoinRequest) error
	FindJoinRequestByIDFunc          func(ctx context.Context, id uuid.UUID) (*domain.ProjectJoinRequest, error)
	FindJoinRequestsByProjectIDFunc  func(ctx context.Context, projectID uuid.UUID, status *domain.ProjectJoinRequestStatus) ([]*domain.ProjectJoinRequest, error)
	FindPendingByProjectAndUserFunc  func(ctx context.Context, projectID, userID uuid.UUID) (*domain.ProjectJoinRequest, error)
	UpdateJoinRequestStatusFunc      func(ctx context.Context, id uuid.UUID, status domain.ProjectJoinRequestStatus) error
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
	if m.SearchFunc != nil {
		return m.SearchFunc(ctx, workspaceID, query, page, limit)
	}
	return nil, 0, nil
}

func (m *MockProjectRepository) AddMember(ctx context.Context, member *domain.ProjectMember) error {
	if m.AddMemberFunc != nil {
		return m.AddMemberFunc(ctx, member)
	}
	return nil
}

func (m *MockProjectRepository) FindMembersByProjectID(ctx context.Context, projectID uuid.UUID) ([]*domain.ProjectMember, error) {
	if m.FindMembersByProjectIDFunc != nil {
		return m.FindMembersByProjectIDFunc(ctx, projectID)
	}
	return nil, nil
}

func (m *MockProjectRepository) FindMemberByProjectAndUser(ctx context.Context, projectID, userID uuid.UUID) (*domain.ProjectMember, error) {
	if m.FindMemberByProjectAndUserFunc != nil {
		return m.FindMemberByProjectAndUserFunc(ctx, projectID, userID)
	}
	return nil, nil
}

func (m *MockProjectRepository) RemoveMember(ctx context.Context, memberID uuid.UUID) error {
	if m.RemoveMemberFunc != nil {
		return m.RemoveMemberFunc(ctx, memberID)
	}
	return nil
}

func (m *MockProjectRepository) UpdateMemberRole(ctx context.Context, memberID uuid.UUID, role domain.ProjectRole) error {
	if m.UpdateMemberRoleFunc != nil {
		return m.UpdateMemberRoleFunc(ctx, memberID, role)
	}
	return nil
}

func (m *MockProjectRepository) IsProjectMember(ctx context.Context, projectID, userID uuid.UUID) (bool, error) {
	if m.IsProjectMemberFunc != nil {
		return m.IsProjectMemberFunc(ctx, projectID, userID)
	}
	return false, nil
}

func (m *MockProjectRepository) CreateJoinRequest(ctx context.Context, request *domain.ProjectJoinRequest) error {
	if m.CreateJoinRequestFunc != nil {
		return m.CreateJoinRequestFunc(ctx, request)
	}
	return nil
}

func (m *MockProjectRepository) FindJoinRequestByID(ctx context.Context, id uuid.UUID) (*domain.ProjectJoinRequest, error) {
	if m.FindJoinRequestByIDFunc != nil {
		return m.FindJoinRequestByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockProjectRepository) FindJoinRequestsByProjectID(ctx context.Context, projectID uuid.UUID, status *domain.ProjectJoinRequestStatus) ([]*domain.ProjectJoinRequest, error) {
	if m.FindJoinRequestsByProjectIDFunc != nil {
		return m.FindJoinRequestsByProjectIDFunc(ctx, projectID, status)
	}
	return nil, nil
}

func (m *MockProjectRepository) FindPendingByProjectAndUser(ctx context.Context, projectID, userID uuid.UUID) (*domain.ProjectJoinRequest, error) {
	if m.FindPendingByProjectAndUserFunc != nil {
		return m.FindPendingByProjectAndUserFunc(ctx, projectID, userID)
	}
	return nil, nil
}

func (m *MockProjectRepository) UpdateJoinRequestStatus(ctx context.Context, id uuid.UUID, status domain.ProjectJoinRequestStatus) error {
	if m.UpdateJoinRequestStatusFunc != nil {
		return m.UpdateJoinRequestStatusFunc(ctx, id, status)
	}
	return nil
}

func TestBoardService_CreateBoard(t *testing.T) {
	projectID := uuid.New()
	validUserID := uuid.New()
	
	tests := []struct {
		name          string
		req           *dto.CreateBoardRequest
		ctx           context.Context
		mockProject   func(*MockProjectRepository)
		mockBoard     func(*MockBoardRepository)
		wantErr       bool
		wantErrCode   string
	}{
		{
			name: "성공: 정상적인 Board 생성",
			ctx:  context.WithValue(context.Background(), "user_id", validUserID),
			req: &dto.CreateBoardRequest{
				ProjectID:  projectID,
				Title:      "Test Board",
				Content:    "Test Content",
				CustomFields: map[string]interface{}{
					"stage":      "in_progress",
					"importance": "urgent",
					"role":       "developer",
				},
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
			name: "성공: CustomFields 없이 Board 생성",
			ctx:  context.WithValue(context.Background(), "user_id", validUserID),
			req: &dto.CreateBoardRequest{
				ProjectID: projectID,
				Title:     "Test Board",
				Content:   "Test Content",
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
			name: "실패: Context에 user_id가 없음",
			ctx:  context.Background(),
			req: &dto.CreateBoardRequest{
				ProjectID: projectID,
				Title:     "Test Board",
				Content:   "Test Content",
			},
			mockProject: func(m *MockProjectRepository) {},
			mockBoard:   func(m *MockBoardRepository) {},
			wantErr:     true,
			wantErrCode: response.ErrCodeUnauthorized,
		},
		{
			name: "실패: Project가 존재하지 않음",
			ctx:  context.WithValue(context.Background(), "user_id", validUserID),
			req: &dto.CreateBoardRequest{
				ProjectID:  projectID,
				Title:      "Test Board",
				Content:    "Test Content",
				CustomFields: map[string]interface{}{
					"stage": "in_progress",
				},
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
			ctx:  context.WithValue(context.Background(), "user_id", validUserID),
			req: &dto.CreateBoardRequest{
				ProjectID:  projectID,
				Title:      "Test Board",
				Content:    "Test Content",
				CustomFields: map[string]interface{}{
					"stage": "in_progress",
				},
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
			mockFieldOptionRepo := &MockFieldOptionRepository{}
			mockConverter := &MockFieldOptionConverter{}
			tt.mockProject(mockProjectRepo)
			tt.mockBoard(mockBoardRepo)
			
			service := NewBoardService(mockBoardRepo, mockProjectRepo, mockFieldOptionRepo, mockConverter, nil)

			// When
			got, err := service.CreateBoard(tt.ctx, tt.req)

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
				// Verify CustomFields are preserved
				if tt.req.CustomFields != nil {
					if got.CustomFields == nil {
						t.Error("CreateBoard() CustomFields = nil, want non-nil")
					}
				}
			}
		})
	}
}

func TestBoardService_CreateBoard_CustomFields(t *testing.T) {
	projectID := uuid.New()
	
	tests := []struct {
		name         string
		customFields map[string]interface{}
		wantFields   map[string]interface{}
	}{
		{
			name: "CustomFields 저장: stage, role, importance",
			customFields: map[string]interface{}{
				"stage":      "in_progress",
				"role":       "developer",
				"importance": "urgent",
			},
			wantFields: map[string]interface{}{
				"stage":      "in_progress",
				"role":       "developer",
				"importance": "urgent",
			},
		},
		{
			name: "CustomFields 저장: stage만",
			customFields: map[string]interface{}{
				"stage": "pending",
			},
			wantFields: map[string]interface{}{
				"stage": "pending",
			},
		},
		{
			name:         "CustomFields 저장: 빈 맵",
			customFields: map[string]interface{}{},
			wantFields:   map[string]interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			mockProjectRepo := &MockProjectRepository{
				FindByIDFunc: func(ctx context.Context, id uuid.UUID) (*domain.Project, error) {
					return &domain.Project{}, nil
				},
			}
			
			var savedBoard *domain.Board
			mockBoardRepo := &MockBoardRepository{
				CreateFunc: func(ctx context.Context, board *domain.Board) error {
					savedBoard = board
					board.ID = uuid.New()
					board.CreatedAt = time.Now()
					board.UpdatedAt = time.Now()
					return nil
				},
			}
			
			mockFieldOptionRepo := &MockFieldOptionRepository{}
			mockConverter := &MockFieldOptionConverter{}
			service := NewBoardService(mockBoardRepo, mockProjectRepo, mockFieldOptionRepo, mockConverter, nil)
			
			req := &dto.CreateBoardRequest{
				ProjectID:    projectID,
				Title:        "Test Board",
				Content:      "Test Content",
				CustomFields: tt.customFields,
			}

			// Create context with user_id (as uuid.UUID type)
			ctx := context.WithValue(context.Background(), "user_id", uuid.New())

			// When
			got, err := service.CreateBoard(ctx, req)

			// Then
			if err != nil {
				t.Errorf("CreateBoard() unexpected error = %v", err)
				return
			}
			
			// Verify CustomFields were saved to domain model
			if savedBoard == nil {
				t.Fatal("Board was not saved")
			}
			
			if len(tt.wantFields) > 0 {
				if savedBoard.CustomFields == nil {
					t.Error("Board.CustomFields = nil, want non-nil")
					return
				}
				
				var customFields map[string]interface{}
				if err := json.Unmarshal(savedBoard.CustomFields, &customFields); err != nil {
					t.Errorf("Failed to unmarshal CustomFields: %v", err)
					return
				}
				
				for key, expectedValue := range tt.wantFields {
					if actualValue, ok := customFields[key]; !ok {
						t.Errorf("Board.CustomFields[%s] not found", key)
					} else if actualValue != expectedValue {
						t.Errorf("Board.CustomFields[%s] = %v, want %v", key, actualValue, expectedValue)
					}
				}
			}
			
			// Verify CustomFields are in response
			if got.CustomFields == nil && len(tt.wantFields) > 0 {
				t.Error("Response.CustomFields = nil, want non-nil")
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
					customFieldsJSON, _ := json.Marshal(map[string]interface{}{
						"stage":      "in_progress",
						"importance": "urgent",
						"role":       "developer",
					})
					return &domain.Board{
						BaseModel: domain.BaseModel{
							ID:        boardID,
							CreatedAt: time.Now(),
							UpdatedAt: time.Now(),
						},
						Title:        "Test Board",
						Content:      "Test Content",
						CustomFields: customFieldsJSON,
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
			mockFieldOptionRepo := &MockFieldOptionRepository{}
			mockConverter := &MockFieldOptionConverter{}
			tt.mockBoard(mockBoardRepo)
			
			service := NewBoardService(mockBoardRepo, mockProjectRepo, mockFieldOptionRepo, mockConverter, nil)

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
	newCustomFields := map[string]interface{}{
		"stage":      "approved",
		"importance": "normal",
	}
	
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
			},
			mockBoard: func(m *MockBoardRepository) {
				m.FindByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.Board, error) {
					customFieldsJSON, _ := json.Marshal(map[string]interface{}{
						"stage": "in_progress",
					})
					return &domain.Board{
						BaseModel: domain.BaseModel{
							ID:        boardID,
							CreatedAt: time.Now(),
							UpdatedAt: time.Now(),
						},
						Title:        "Old Title",
						CustomFields: customFieldsJSON,
					}, nil
				}
				m.UpdateFunc = func(ctx context.Context, board *domain.Board) error {
					return nil
				}
			},
			wantErr: false,
		},
		{
			name:    "성공: CustomFields 업데이트",
			boardID: boardID,
			req: &dto.UpdateBoardRequest{
				CustomFields: &newCustomFields,
			},
			mockBoard: func(m *MockBoardRepository) {
				m.FindByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.Board, error) {
					customFieldsJSON, _ := json.Marshal(map[string]interface{}{
						"stage": "in_progress",
					})
					return &domain.Board{
						BaseModel: domain.BaseModel{
							ID:        boardID,
							CreatedAt: time.Now(),
							UpdatedAt: time.Now(),
						},
						Title:        "Test Board",
						CustomFields: customFieldsJSON,
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
			mockFieldOptionRepo := &MockFieldOptionRepository{}
			mockConverter := &MockFieldOptionConverter{}
			tt.mockBoard(mockBoardRepo)
			
			service := NewBoardService(mockBoardRepo, mockProjectRepo, mockFieldOptionRepo, mockConverter, nil)

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

func TestBoardService_UpdateBoard_CustomFields(t *testing.T) {
	boardID := uuid.New()
	
	tests := []struct {
		name             string
		existingFields   map[string]interface{}
		updateFields     map[string]interface{}
		wantFields       map[string]interface{}
	}{
		{
			name: "CustomFields 수정: 전체 교체",
			existingFields: map[string]interface{}{
				"stage":      "in_progress",
				"importance": "urgent",
			},
			updateFields: map[string]interface{}{
				"stage":      "approved",
				"importance": "normal",
				"role":       "developer",
			},
			wantFields: map[string]interface{}{
				"stage":      "approved",
				"importance": "normal",
				"role":       "developer",
			},
		},
		{
			name: "CustomFields 수정: 일부 필드만 변경",
			existingFields: map[string]interface{}{
				"stage":      "in_progress",
				"importance": "urgent",
				"role":       "planner",
			},
			updateFields: map[string]interface{}{
				"stage": "approved",
			},
			wantFields: map[string]interface{}{
				"stage": "approved",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			var updatedBoard *domain.Board
			mockBoardRepo := &MockBoardRepository{
				FindByIDFunc: func(ctx context.Context, id uuid.UUID) (*domain.Board, error) {
					customFieldsJSON, _ := json.Marshal(tt.existingFields)
					return &domain.Board{
						BaseModel: domain.BaseModel{
							ID:        boardID,
							CreatedAt: time.Now(),
							UpdatedAt: time.Now(),
						},
						Title:        "Test Board",
						CustomFields: customFieldsJSON,
					}, nil
				},
				UpdateFunc: func(ctx context.Context, board *domain.Board) error {
					updatedBoard = board
					return nil
				},
			}
			
			mockProjectRepo := &MockProjectRepository{}
			mockFieldOptionRepo := &MockFieldOptionRepository{}
			mockConverter := &MockFieldOptionConverter{}
			service := NewBoardService(mockBoardRepo, mockProjectRepo, mockFieldOptionRepo, mockConverter, nil)
			
			req := &dto.UpdateBoardRequest{
				CustomFields: &tt.updateFields,
			}

			// When
			got, err := service.UpdateBoard(context.Background(), boardID, req)

			// Then
			if err != nil {
				t.Errorf("UpdateBoard() unexpected error = %v", err)
				return
			}
			
			// Verify CustomFields were updated in domain model
			if updatedBoard == nil {
				t.Fatal("Board was not updated")
			}
			
			if updatedBoard.CustomFields == nil {
				t.Error("Board.CustomFields = nil, want non-nil")
				return
			}
			
			var customFields map[string]interface{}
			if err := json.Unmarshal(updatedBoard.CustomFields, &customFields); err != nil {
				t.Errorf("Failed to unmarshal CustomFields: %v", err)
				return
			}
			
			for key, expectedValue := range tt.wantFields {
				if actualValue, ok := customFields[key]; !ok {
					t.Errorf("Board.CustomFields[%s] not found", key)
				} else if actualValue != expectedValue {
					t.Errorf("Board.CustomFields[%s] = %v, want %v", key, actualValue, expectedValue)
				}
			}
			
			// Verify CustomFields are in response
			if got.CustomFields == nil {
				t.Error("Response.CustomFields = nil, want non-nil")
			}
		})
	}
}

func TestBoardService_GetBoardsByProject_CustomFieldsFilter(t *testing.T) {
	projectID := uuid.New()
	
	tests := []struct {
		name        string
		filters     *dto.BoardFilters
		mockProject func(*MockProjectRepository)
		mockBoard   func(*MockBoardRepository)
		wantCount   int
		wantErr     bool
		wantErrCode string
	}{
		{
			name: "성공: CustomFields 필터링 없이 조회",
			filters: nil,
			mockProject: func(m *MockProjectRepository) {
				m.FindByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.Project, error) {
					return &domain.Project{}, nil
				}
			},
			mockBoard: func(m *MockBoardRepository) {
				m.FindByProjectIDFunc = func(ctx context.Context, pid uuid.UUID, filters interface{}) ([]*domain.Board, error) {
					customFields1JSON, _ := json.Marshal(map[string]interface{}{"stage": "in_progress"})
					customFields2JSON, _ := json.Marshal(map[string]interface{}{"stage": "approved"})
					return []*domain.Board{
						{
							BaseModel:    domain.BaseModel{ID: uuid.New()},
							Title:        "Board 1",
							CustomFields: customFields1JSON,
						},
						{
							BaseModel:    domain.BaseModel{ID: uuid.New()},
							Title:        "Board 2",
							CustomFields: customFields2JSON,
						},
					}, nil
				}
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name: "성공: stage 필터링",
			filters: &dto.BoardFilters{
				CustomFields: map[string]interface{}{
					"stage": "in_progress",
				},
			},
			mockProject: func(m *MockProjectRepository) {
				m.FindByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.Project, error) {
					return &domain.Project{}, nil
				}
			},
			mockBoard: func(m *MockBoardRepository) {
				m.FindByProjectIDFunc = func(ctx context.Context, pid uuid.UUID, filters interface{}) ([]*domain.Board, error) {
					// Simulate filtering
					if customFields, ok := filters.(map[string]interface{}); ok {
						if stage, ok := customFields["stage"]; ok && stage == "in_progress" {
							customFieldsJSON, _ := json.Marshal(map[string]interface{}{"stage": "in_progress"})
							return []*domain.Board{
								{
									BaseModel:    domain.BaseModel{ID: uuid.New()},
									Title:        "Board 1",
									CustomFields: customFieldsJSON,
								},
							}, nil
						}
					}
					return []*domain.Board{}, nil
				}
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name: "성공: 여러 필드로 필터링",
			filters: &dto.BoardFilters{
				CustomFields: map[string]interface{}{
					"stage":      "in_progress",
					"importance": "urgent",
				},
			},
			mockProject: func(m *MockProjectRepository) {
				m.FindByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.Project, error) {
					return &domain.Project{}, nil
				}
			},
			mockBoard: func(m *MockBoardRepository) {
				m.FindByProjectIDFunc = func(ctx context.Context, pid uuid.UUID, filters interface{}) ([]*domain.Board, error) {
					// Simulate AND filtering
					if customFields, ok := filters.(map[string]interface{}); ok {
						stage, hasStage := customFields["stage"]
						importance, hasImportance := customFields["importance"]
						if hasStage && hasImportance && stage == "in_progress" && importance == "urgent" {
							customFieldsJSON, _ := json.Marshal(map[string]interface{}{
								"stage":      "in_progress",
								"importance": "urgent",
							})
							return []*domain.Board{
								{
									BaseModel:    domain.BaseModel{ID: uuid.New()},
									Title:        "Urgent Board",
									CustomFields: customFieldsJSON,
								},
							}, nil
						}
					}
					return []*domain.Board{}, nil
				}
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name: "성공: 필터 조건에 맞는 보드 없음",
			filters: &dto.BoardFilters{
				CustomFields: map[string]interface{}{
					"stage": "nonexistent",
				},
			},
			mockProject: func(m *MockProjectRepository) {
				m.FindByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.Project, error) {
					return &domain.Project{}, nil
				}
			},
			mockBoard: func(m *MockBoardRepository) {
				m.FindByProjectIDFunc = func(ctx context.Context, pid uuid.UUID, filters interface{}) ([]*domain.Board, error) {
					return []*domain.Board{}, nil
				}
			},
			wantCount: 0,
			wantErr:   false,
		},
		{
			name:    "실패: Project가 존재하지 않음",
			filters: nil,
			mockProject: func(m *MockProjectRepository) {
				m.FindByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.Project, error) {
					return nil, gorm.ErrRecordNotFound
				}
			},
			mockBoard:   func(m *MockBoardRepository) {},
			wantErr:     true,
			wantErrCode: response.ErrCodeNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			mockProjectRepo := &MockProjectRepository{}
			mockBoardRepo := &MockBoardRepository{}
			mockFieldOptionRepo := &MockFieldOptionRepository{}
			mockConverter := &MockFieldOptionConverter{}
			tt.mockProject(mockProjectRepo)
			tt.mockBoard(mockBoardRepo)
			
			service := NewBoardService(mockBoardRepo, mockProjectRepo, mockFieldOptionRepo, mockConverter, nil)

			// When
			got, err := service.GetBoardsByProject(context.Background(), projectID, tt.filters)

			// Then
			if tt.wantErr {
				if err == nil {
					t.Errorf("GetBoardsByProject() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if appErr, ok := err.(*response.AppError); ok {
					if appErr.Code != tt.wantErrCode {
						t.Errorf("GetBoardsByProject() error code = %v, want %v", appErr.Code, tt.wantErrCode)
					}
				}
			} else {
				if err != nil {
					t.Errorf("GetBoardsByProject() unexpected error = %v", err)
					return
				}
				if len(got) != tt.wantCount {
					t.Errorf("GetBoardsByProject() count = %v, want %v", len(got), tt.wantCount)
				}
				// Verify CustomFields are in response
				for _, board := range got {
					if board.CustomFields == nil && tt.filters != nil && tt.filters.CustomFields != nil {
						t.Error("Board.CustomFields = nil in response")
					}
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
			mockFieldOptionRepo := &MockFieldOptionRepository{}
			mockConverter := &MockFieldOptionConverter{}
			tt.mockBoard(mockBoardRepo)
			
			service := NewBoardService(mockBoardRepo, mockProjectRepo, mockFieldOptionRepo, mockConverter, nil)

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

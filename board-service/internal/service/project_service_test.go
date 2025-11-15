package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"project-board-api/internal/client"
	"project-board-api/internal/domain"
	"project-board-api/internal/dto"
	"project-board-api/internal/response"
)

// MockUserClient is a mock implementation of UserClient
type MockUserClient struct {
	ValidateWorkspaceMemberFunc func(ctx context.Context, workspaceID, userID uuid.UUID, token string) (bool, error)
	GetUserProfileFunc          func(ctx context.Context, userID uuid.UUID, token string) (*client.UserProfile, error)
	GetWorkspaceProfileFunc     func(ctx context.Context, workspaceID, userID uuid.UUID, token string) (*client.WorkspaceProfile, error)
}

func (m *MockUserClient) ValidateWorkspaceMember(ctx context.Context, workspaceID, userID uuid.UUID, token string) (bool, error) {
	if m.ValidateWorkspaceMemberFunc != nil {
		return m.ValidateWorkspaceMemberFunc(ctx, workspaceID, userID, token)
	}
	return true, nil
}

func (m *MockUserClient) GetUserProfile(ctx context.Context, userID uuid.UUID, token string) (*client.UserProfile, error) {
	if m.GetUserProfileFunc != nil {
		return m.GetUserProfileFunc(ctx, userID, token)
	}
	return &client.UserProfile{UserID: userID, Email: "test@example.com"}, nil
}

func (m *MockUserClient) GetWorkspaceProfile(ctx context.Context, workspaceID, userID uuid.UUID, token string) (*client.WorkspaceProfile, error) {
	if m.GetWorkspaceProfileFunc != nil {
		return m.GetWorkspaceProfileFunc(ctx, workspaceID, userID, token)
	}
	return &client.WorkspaceProfile{
		WorkspaceID: workspaceID,
		UserID:      userID,
		NickName:    "Test User",
		Email:       "test@example.com",
	}, nil
}

func TestProjectService_CreateProject(t *testing.T) {
	workspaceID := uuid.New()
	userID := uuid.New()
	token := "test-jwt-token"

	tests := []struct {
		name        string
		req         *dto.CreateProjectRequest
		mockProject func(*MockProjectRepository)
		mockUser    func(*MockUserClient)
		wantErr     bool
		wantErrCode string
	}{
		{
			name: "성공: 정상적인 Project 생성",
			req: &dto.CreateProjectRequest{
				WorkspaceID: workspaceID,
				Name:        "Test Project",
				Description: "Test Description",
			},
			mockProject: func(m *MockProjectRepository) {
				m.CreateFunc = func(ctx context.Context, project *domain.Project) error {
					project.ID = uuid.New()
					project.CreatedAt = time.Now()
					project.UpdatedAt = time.Now()
					return nil
				}
			},
			mockUser: func(m *MockUserClient) {
				m.ValidateWorkspaceMemberFunc = func(ctx context.Context, wID, uID uuid.UUID, t string) (bool, error) {
					return true, nil
				}
			},
			wantErr: false,
		},
		{
			name: "성공: Default Project 생성",
			req: &dto.CreateProjectRequest{
				WorkspaceID: workspaceID,
				Name:        "Default Project",
				Description: "Default Description",
			},
			mockProject: func(m *MockProjectRepository) {
				m.CreateFunc = func(ctx context.Context, project *domain.Project) error {
					project.ID = uuid.New()
					project.CreatedAt = time.Now()
					project.UpdatedAt = time.Now()
					return nil
				}
			},
			mockUser: func(m *MockUserClient) {
				m.ValidateWorkspaceMemberFunc = func(ctx context.Context, wID, uID uuid.UUID, t string) (bool, error) {
					return true, nil
				}
			},
			wantErr: false,
		},
		{
			name: "실패: Workspace 멤버십 검증 실패",
			req: &dto.CreateProjectRequest{
				WorkspaceID: workspaceID,
				Name:        "Test Project",
				Description: "Test Description",
			},
			mockProject: func(m *MockProjectRepository) {},
			mockUser: func(m *MockUserClient) {
				m.ValidateWorkspaceMemberFunc = func(ctx context.Context, wID, uID uuid.UUID, t string) (bool, error) {
					return false, nil
				}
			},
			wantErr:     true,
			wantErrCode: response.ErrCodeForbidden,
		},
		{
			name: "실패: User API 호출 에러",
			req: &dto.CreateProjectRequest{
				WorkspaceID: workspaceID,
				Name:        "Test Project",
				Description: "Test Description",
			},
			mockProject: func(m *MockProjectRepository) {},
			mockUser: func(m *MockUserClient) {
				m.ValidateWorkspaceMemberFunc = func(ctx context.Context, wID, uID uuid.UUID, t string) (bool, error) {
					return false, errors.New("user API error")
				}
			},
			wantErr:     true,
			wantErrCode: response.ErrCodeForbidden,
		},
		{
			name: "실패: Project 생성 중 DB 에러",
			req: &dto.CreateProjectRequest{
				WorkspaceID: workspaceID,
				Name:        "Test Project",
				Description: "Test Description",
			},
			mockProject: func(m *MockProjectRepository) {
				m.CreateFunc = func(ctx context.Context, project *domain.Project) error {
					return errors.New("database error")
				}
			},
			mockUser: func(m *MockUserClient) {
				m.ValidateWorkspaceMemberFunc = func(ctx context.Context, wID, uID uuid.UUID, t string) (bool, error) {
					return true, nil
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
			mockUserClient := &MockUserClient{}
			tt.mockProject(mockProjectRepo)
			tt.mockUser(mockUserClient)

			service := NewProjectService(mockProjectRepo, mockUserClient)

			// When
			got, err := service.CreateProject(context.Background(), tt.req, userID, token)

			// Then
			if tt.wantErr {
				if err == nil {
					t.Errorf("CreateProject() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if appErr, ok := err.(*response.AppError); ok {
					if appErr.Code != tt.wantErrCode {
						t.Errorf("CreateProject() error code = %v, want %v", appErr.Code, tt.wantErrCode)
					}
				}
			} else {
				if err != nil {
					t.Errorf("CreateProject() unexpected error = %v", err)
					return
				}
				if got == nil {
					t.Error("CreateProject() returned nil response")
					return
				}
				if got.Name != tt.req.Name {
					t.Errorf("CreateProject() Name = %v, want %v", got.Name, tt.req.Name)
				}
			}
		})
	}
}

func TestProjectService_GetProjectsByWorkspace(t *testing.T) {
	workspaceID := uuid.New()
	userID := uuid.New()
	ownerID := uuid.New()
	token := "test-jwt-token"

	tests := []struct {
		name        string
		workspaceID uuid.UUID
		mockProject func(*MockProjectRepository)
		mockUser    func(*MockUserClient)
		wantErr     bool
		wantErrCode string
		wantCount   int
	}{
		{
			name:        "성공: Project 목록 조회 with 프로필 정보",
			workspaceID: workspaceID,
			mockProject: func(m *MockProjectRepository) {
				m.FindByWorkspaceIDFunc = func(ctx context.Context, wID uuid.UUID) ([]*domain.Project, error) {
					return []*domain.Project{
						{
							BaseModel:   domain.BaseModel{ID: uuid.New(), CreatedAt: time.Now(), UpdatedAt: time.Now()},
							WorkspaceID: workspaceID,
							OwnerID:     ownerID,
							Name:        "Project 1",
							IsPublic:    true,
						},
						{
							BaseModel:   domain.BaseModel{ID: uuid.New(), CreatedAt: time.Now(), UpdatedAt: time.Now()},
							WorkspaceID: workspaceID,
							OwnerID:     ownerID,
							Name:        "Project 2",
							IsPublic:    false,
						},
					}, nil
				}
			},
			mockUser: func(m *MockUserClient) {
				m.ValidateWorkspaceMemberFunc = func(ctx context.Context, wID, uID uuid.UUID, t string) (bool, error) {
					return true, nil
				}
				m.GetWorkspaceProfileFunc = func(ctx context.Context, wID, uID uuid.UUID, t string) (*client.WorkspaceProfile, error) {
					return &client.WorkspaceProfile{
						WorkspaceID: wID,
						UserID:      uID,
						NickName:    "Test Owner",
						Email:       "owner@example.com",
					}, nil
				}
			},
			wantErr:   false,
			wantCount: 2,
		},
		{
			name:        "성공: 빈 Project 목록",
			workspaceID: workspaceID,
			mockProject: func(m *MockProjectRepository) {
				m.FindByWorkspaceIDFunc = func(ctx context.Context, wID uuid.UUID) ([]*domain.Project, error) {
					return []*domain.Project{}, nil
				}
			},
			mockUser: func(m *MockUserClient) {
				m.ValidateWorkspaceMemberFunc = func(ctx context.Context, wID, uID uuid.UUID, t string) (bool, error) {
					return true, nil
				}
			},
			wantErr:   false,
			wantCount: 0,
		},
		{
			name:        "성공: 프로필 조회 실패 시 graceful degradation",
			workspaceID: workspaceID,
			mockProject: func(m *MockProjectRepository) {
				m.FindByWorkspaceIDFunc = func(ctx context.Context, wID uuid.UUID) ([]*domain.Project, error) {
					return []*domain.Project{
						{
							BaseModel:   domain.BaseModel{ID: uuid.New(), CreatedAt: time.Now(), UpdatedAt: time.Now()},
							WorkspaceID: workspaceID,
							OwnerID:     ownerID,
							Name:        "Project 1",
							IsPublic:    true,
						},
					}, nil
				}
			},
			mockUser: func(m *MockUserClient) {
				m.ValidateWorkspaceMemberFunc = func(ctx context.Context, wID, uID uuid.UUID, t string) (bool, error) {
					return true, nil
				}
				m.GetWorkspaceProfileFunc = func(ctx context.Context, wID, uID uuid.UUID, t string) (*client.WorkspaceProfile, error) {
					return nil, errors.New("profile API error")
				}
			},
			wantErr:   false,
			wantCount: 1,
		},
		{
			name:        "실패: Workspace 멤버십 검증 실패",
			workspaceID: workspaceID,
			mockProject: func(m *MockProjectRepository) {},
			mockUser: func(m *MockUserClient) {
				m.ValidateWorkspaceMemberFunc = func(ctx context.Context, wID, uID uuid.UUID, t string) (bool, error) {
					return false, nil
				}
			},
			wantErr:     true,
			wantErrCode: response.ErrCodeForbidden,
		},
		{
			name:        "실패: DB 에러",
			workspaceID: workspaceID,
			mockProject: func(m *MockProjectRepository) {
				m.FindByWorkspaceIDFunc = func(ctx context.Context, wID uuid.UUID) ([]*domain.Project, error) {
					return nil, errors.New("database error")
				}
			},
			mockUser: func(m *MockUserClient) {
				m.ValidateWorkspaceMemberFunc = func(ctx context.Context, wID, uID uuid.UUID, t string) (bool, error) {
					return true, nil
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
			mockUserClient := &MockUserClient{}
			tt.mockProject(mockProjectRepo)
			tt.mockUser(mockUserClient)

			service := NewProjectService(mockProjectRepo, mockUserClient)

			// When
			got, err := service.GetProjectsByWorkspace(context.Background(), tt.workspaceID, userID, token)

			// Then
			if tt.wantErr {
				if err == nil {
					t.Errorf("GetProjectsByWorkspace() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if appErr, ok := err.(*response.AppError); ok {
					if appErr.Code != tt.wantErrCode {
						t.Errorf("GetProjectsByWorkspace() error code = %v, want %v", appErr.Code, tt.wantErrCode)
					}
				}
			} else {
				if err != nil {
					t.Errorf("GetProjectsByWorkspace() unexpected error = %v", err)
					return
				}
				if got == nil {
					t.Error("GetProjectsByWorkspace() returned nil response")
					return
				}
				if len(got) != tt.wantCount {
					t.Errorf("GetProjectsByWorkspace() count = %v, want %v", len(got), tt.wantCount)
				}
			}
		})
	}
}

func TestProjectService_GetDefaultProject(t *testing.T) {
	workspaceID := uuid.New()
	projectID := uuid.New()
	userID := uuid.New()
	ownerID := uuid.New()
	token := "test-jwt-token"

	tests := []struct {
		name        string
		workspaceID uuid.UUID
		mockProject func(*MockProjectRepository)
		mockUser    func(*MockUserClient)
		wantErr     bool
		wantErrCode string
	}{
		{
			name:        "성공: Default Project 조회 with 프로필 정보",
			workspaceID: workspaceID,
			mockProject: func(m *MockProjectRepository) {
				m.FindDefaultByWorkspaceIDFunc = func(ctx context.Context, wID uuid.UUID) (*domain.Project, error) {
					return &domain.Project{
						BaseModel:   domain.BaseModel{ID: projectID, CreatedAt: time.Now(), UpdatedAt: time.Now()},
						WorkspaceID: workspaceID,
						OwnerID:     ownerID,
						Name:        "Default Project",
						IsPublic:    true,
					}, nil
				}
			},
			mockUser: func(m *MockUserClient) {
				m.ValidateWorkspaceMemberFunc = func(ctx context.Context, wID, uID uuid.UUID, t string) (bool, error) {
					return true, nil
				}
				m.GetWorkspaceProfileFunc = func(ctx context.Context, wID, uID uuid.UUID, t string) (*client.WorkspaceProfile, error) {
					return &client.WorkspaceProfile{
						WorkspaceID: wID,
						UserID:      uID,
						NickName:    "Default Owner",
						Email:       "default@example.com",
					}, nil
				}
			},
			wantErr: false,
		},
		{
			name:        "성공: 프로필 조회 실패 시 graceful degradation",
			workspaceID: workspaceID,
			mockProject: func(m *MockProjectRepository) {
				m.FindDefaultByWorkspaceIDFunc = func(ctx context.Context, wID uuid.UUID) (*domain.Project, error) {
					return &domain.Project{
						BaseModel:   domain.BaseModel{ID: projectID, CreatedAt: time.Now(), UpdatedAt: time.Now()},
						WorkspaceID: workspaceID,
						OwnerID:     ownerID,
						Name:        "Default Project",
						IsPublic:    true,
					}, nil
				}
			},
			mockUser: func(m *MockUserClient) {
				m.ValidateWorkspaceMemberFunc = func(ctx context.Context, wID, uID uuid.UUID, t string) (bool, error) {
					return true, nil
				}
				m.GetWorkspaceProfileFunc = func(ctx context.Context, wID, uID uuid.UUID, t string) (*client.WorkspaceProfile, error) {
					return nil, errors.New("profile API error")
				}
			},
			wantErr: false,
		},
		{
			name:        "실패: Workspace 멤버십 검증 실패",
			workspaceID: workspaceID,
			mockProject: func(m *MockProjectRepository) {},
			mockUser: func(m *MockUserClient) {
				m.ValidateWorkspaceMemberFunc = func(ctx context.Context, wID, uID uuid.UUID, t string) (bool, error) {
					return false, nil
				}
			},
			wantErr:     true,
			wantErrCode: response.ErrCodeForbidden,
		},
		{
			name:        "실패: Default Project가 존재하지 않음",
			workspaceID: workspaceID,
			mockProject: func(m *MockProjectRepository) {
				m.FindDefaultByWorkspaceIDFunc = func(ctx context.Context, wID uuid.UUID) (*domain.Project, error) {
					return nil, gorm.ErrRecordNotFound
				}
			},
			mockUser: func(m *MockUserClient) {
				m.ValidateWorkspaceMemberFunc = func(ctx context.Context, wID, uID uuid.UUID, t string) (bool, error) {
					return true, nil
				}
			},
			wantErr:     true,
			wantErrCode: response.ErrCodeNotFound,
		},
		{
			name:        "실패: DB 에러",
			workspaceID: workspaceID,
			mockProject: func(m *MockProjectRepository) {
				m.FindDefaultByWorkspaceIDFunc = func(ctx context.Context, wID uuid.UUID) (*domain.Project, error) {
					return nil, errors.New("database error")
				}
			},
			mockUser: func(m *MockUserClient) {
				m.ValidateWorkspaceMemberFunc = func(ctx context.Context, wID, uID uuid.UUID, t string) (bool, error) {
					return true, nil
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
			mockUserClient := &MockUserClient{}
			tt.mockProject(mockProjectRepo)
			tt.mockUser(mockUserClient)

			service := NewProjectService(mockProjectRepo, mockUserClient)

			// When
			got, err := service.GetDefaultProject(context.Background(), tt.workspaceID, userID, token)

			// Then
			if tt.wantErr {
				if err == nil {
					t.Errorf("GetDefaultProject() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if appErr, ok := err.(*response.AppError); ok {
					if appErr.Code != tt.wantErrCode {
						t.Errorf("GetDefaultProject() error code = %v, want %v", appErr.Code, tt.wantErrCode)
					}
				}
			} else {
				if err != nil {
					t.Errorf("GetDefaultProject() unexpected error = %v", err)
					return
				}
				if got == nil {
					t.Error("GetDefaultProject() returned nil response")
					return
				}
			}
		})
	}
}

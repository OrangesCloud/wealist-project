package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"project-board-api/internal/client"
	"project-board-api/internal/domain"
	"project-board-api/internal/dto"
	"project-board-api/internal/response"
)

// getTestLogger returns a no-op logger for testing
func getTestLogger() *zap.Logger {
	return zap.NewNop()
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
				m.AddMemberFunc = func(ctx context.Context, member *domain.ProjectMember) error {
					// Verify that OWNER member is being added
					if member.RoleName != domain.ProjectRoleOwner {
						return errors.New("expected OWNER role")
					}
					if member.UserID != userID {
						return errors.New("expected creator as member")
					}
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
				m.AddMemberFunc = func(ctx context.Context, member *domain.ProjectMember) error {
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
			name: "성공: Project 생성 시 owner_id 설정 및 OWNER 멤버 자동 추가",
			req: &dto.CreateProjectRequest{
				WorkspaceID: workspaceID,
				Name:        "Test Project with Owner",
				Description: "Test Description",
			},
			mockProject: func(m *MockProjectRepository) {
				var createdProject *domain.Project
				m.CreateFunc = func(ctx context.Context, project *domain.Project) error {
					// Verify owner_id is set to the creator's userID
					if project.OwnerID != userID {
						return errors.New("owner_id should be set to creator's userID")
					}
					project.ID = uuid.New()
					project.CreatedAt = time.Now()
					project.UpdatedAt = time.Now()
					createdProject = project
					return nil
				}
				m.AddMemberFunc = func(ctx context.Context, member *domain.ProjectMember) error {
					// Verify OWNER member is added with correct details
					if member.ProjectID != createdProject.ID {
						return errors.New("member projectID should match created project")
					}
					if member.UserID != userID {
						return errors.New("member userID should match creator")
					}
					if member.RoleName != domain.ProjectRoleOwner {
						return errors.New("member role should be OWNER")
					}
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
		{
			name: "실패: OWNER 멤버 추가 중 DB 에러",
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
				m.AddMemberFunc = func(ctx context.Context, member *domain.ProjectMember) error {
					return errors.New("failed to add member")
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
			mockFieldOptionRepo := &MockFieldOptionRepository{}
			mockUserClient := &MockUserClient{}
			tt.mockProject(mockProjectRepo)
			tt.mockUser(mockUserClient)

			service := NewProjectService(mockProjectRepo, mockFieldOptionRepo, mockUserClient, getTestLogger())

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
			mockFieldOptionRepo := &MockFieldOptionRepository{}
			mockUserClient := &MockUserClient{}
			tt.mockProject(mockProjectRepo)
			tt.mockUser(mockUserClient)

			service := NewProjectService(mockProjectRepo, mockFieldOptionRepo, mockUserClient, getTestLogger())

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
			mockFieldOptionRepo := &MockFieldOptionRepository{}
			mockUserClient := &MockUserClient{}
			tt.mockProject(mockProjectRepo)
			tt.mockUser(mockUserClient)

			service := NewProjectService(mockProjectRepo, mockFieldOptionRepo, mockUserClient, getTestLogger())

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

func TestProjectService_GetProject(t *testing.T) {
	projectID := uuid.New()
	userID := uuid.New()
	ownerID := uuid.New()
	workspaceID := uuid.New()
	token := "test-jwt-token"

	tests := []struct {
		name        string
		mockProject func(*MockProjectRepository)
		mockUser    func(*MockUserClient)
		wantErr     bool
		wantErrCode string
	}{
		{
			name: "성공: 프로젝트 조회 with 프로필 정보",
			mockProject: func(m *MockProjectRepository) {
				m.FindByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.Project, error) {
					return &domain.Project{
						BaseModel:   domain.BaseModel{ID: projectID, CreatedAt: time.Now(), UpdatedAt: time.Now()},
						WorkspaceID: workspaceID,
						OwnerID:     ownerID,
						Name:        "Test Project",
						Description: "Test Description",
						IsPublic:    true,
					}, nil
				}
				m.IsProjectMemberFunc = func(ctx context.Context, pID, uID uuid.UUID) (bool, error) {
					return true, nil
				}
			},
			mockUser: func(m *MockUserClient) {
				m.GetWorkspaceProfileFunc = func(ctx context.Context, wID, uID uuid.UUID, t string) (*client.WorkspaceProfile, error) {
					return &client.WorkspaceProfile{
						WorkspaceID: wID,
						UserID:      uID,
						NickName:    "Test Owner",
						Email:       "owner@example.com",
					}, nil
				}
			},
			wantErr: false,
		},
		{
			name: "실패: 프로젝트가 존재하지 않음",
			mockProject: func(m *MockProjectRepository) {
				m.FindByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.Project, error) {
					return nil, gorm.ErrRecordNotFound
				}
			},
			mockUser:    func(m *MockUserClient) {},
			wantErr:     true,
			wantErrCode: response.ErrCodeNotFound,
		},
		{
			name: "실패: 프로젝트 멤버가 아님",
			mockProject: func(m *MockProjectRepository) {
				m.FindByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.Project, error) {
					return &domain.Project{
						BaseModel:   domain.BaseModel{ID: projectID, CreatedAt: time.Now(), UpdatedAt: time.Now()},
						WorkspaceID: workspaceID,
						OwnerID:     ownerID,
						Name:        "Test Project",
					}, nil
				}
				m.IsProjectMemberFunc = func(ctx context.Context, pID, uID uuid.UUID) (bool, error) {
					return false, nil
				}
			},
			mockUser: func(m *MockUserClient) {
				// Mock workspace membership validation to return false
				m.ValidateWorkspaceMemberFunc = func(ctx context.Context, wID, uID uuid.UUID, t string) (bool, error) {
					return false, nil
				}
			},
			wantErr:     true,
			wantErrCode: response.ErrCodeForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			mockProjectRepo := &MockProjectRepository{}
			mockFieldOptionRepo := &MockFieldOptionRepository{}
			mockUserClient := &MockUserClient{}
			tt.mockProject(mockProjectRepo)
			tt.mockUser(mockUserClient)

			service := NewProjectService(mockProjectRepo, mockFieldOptionRepo, mockUserClient, getTestLogger())

			// When
			got, err := service.GetProject(context.Background(), projectID, userID, token)

			// Then
			if tt.wantErr {
				if err == nil {
					t.Errorf("GetProject() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if appErr, ok := err.(*response.AppError); ok {
					if appErr.Code != tt.wantErrCode {
						t.Errorf("GetProject() error code = %v, want %v", appErr.Code, tt.wantErrCode)
					}
				}
			} else {
				if err != nil {
					t.Errorf("GetProject() unexpected error = %v", err)
					return
				}
				if got == nil {
					t.Error("GetProject() returned nil response")
				}
			}
		})
	}
}

func TestProjectService_UpdateProject(t *testing.T) {
	projectID := uuid.New()
	userID := uuid.New()
	ownerID := userID
	workspaceID := uuid.New()
	newName := "Updated Project"
	newDescription := "Updated Description"

	tests := []struct {
		name        string
		req         *dto.UpdateProjectRequest
		mockProject func(*MockProjectRepository)
		wantErr     bool
		wantErrCode string
	}{
		{
			name: "성공: OWNER가 프로젝트 수정",
			req: &dto.UpdateProjectRequest{
				Name:        &newName,
				Description: &newDescription,
			},
			mockProject: func(m *MockProjectRepository) {
				m.FindByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.Project, error) {
					return &domain.Project{
						BaseModel:   domain.BaseModel{ID: projectID, CreatedAt: time.Now(), UpdatedAt: time.Now()},
						WorkspaceID: workspaceID,
						OwnerID:     ownerID,
						Name:        "Old Name",
						Description: "Old Description",
					}, nil
				}
				m.FindMemberByProjectAndUserFunc = func(ctx context.Context, pID, uID uuid.UUID) (*domain.ProjectMember, error) {
					return &domain.ProjectMember{
						ID:        uuid.New(),
						ProjectID: pID,
						UserID:    uID,
						RoleName:  domain.ProjectRoleOwner,
					}, nil
				}
				m.UpdateFunc = func(ctx context.Context, project *domain.Project) error {
					if project.Name != newName {
						return errors.New("name not updated")
					}
					if project.Description != newDescription {
						return errors.New("description not updated")
					}
					return nil
				}
			},
			wantErr: false,
		},
		{
			name: "실패: OWNER가 아닌 사용자가 수정 시도",
			req: &dto.UpdateProjectRequest{
				Name: &newName,
			},
			mockProject: func(m *MockProjectRepository) {
				m.FindByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.Project, error) {
					return &domain.Project{
						BaseModel:   domain.BaseModel{ID: projectID, CreatedAt: time.Now(), UpdatedAt: time.Now()},
						WorkspaceID: workspaceID,
						OwnerID:     ownerID,
						Name:        "Old Name",
					}, nil
				}
				m.FindMemberByProjectAndUserFunc = func(ctx context.Context, pID, uID uuid.UUID) (*domain.ProjectMember, error) {
					return &domain.ProjectMember{
						ID:        uuid.New(),
						ProjectID: pID,
						UserID:    uID,
						RoleName:  domain.ProjectRoleAdmin,
					}, nil
				}
			},
			wantErr:     true,
			wantErrCode: response.ErrCodeForbidden,
		},
		{
			name: "실패: 프로젝트가 존재하지 않음",
			req: &dto.UpdateProjectRequest{
				Name: &newName,
			},
			mockProject: func(m *MockProjectRepository) {
				m.FindByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.Project, error) {
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
			mockProjectRepo := &MockProjectRepository{}
			mockFieldOptionRepo := &MockFieldOptionRepository{}
			mockUserClient := &MockUserClient{}
			tt.mockProject(mockProjectRepo)

			service := NewProjectService(mockProjectRepo, mockFieldOptionRepo, mockUserClient, getTestLogger())

			// When
			got, err := service.UpdateProject(context.Background(), projectID, userID, tt.req)

			// Then
			if tt.wantErr {
				if err == nil {
					t.Errorf("UpdateProject() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if appErr, ok := err.(*response.AppError); ok {
					if appErr.Code != tt.wantErrCode {
						t.Errorf("UpdateProject() error code = %v, want %v", appErr.Code, tt.wantErrCode)
					}
				}
			} else {
				if err != nil {
					t.Errorf("UpdateProject() unexpected error = %v", err)
					return
				}
				if got == nil {
					t.Error("UpdateProject() returned nil response")
				}
			}
		})
	}
}

func TestProjectService_DeleteProject(t *testing.T) {
	projectID := uuid.New()
	userID := uuid.New()
	ownerID := userID
	workspaceID := uuid.New()

	tests := []struct {
		name        string
		mockProject func(*MockProjectRepository)
		wantErr     bool
		wantErrCode string
	}{
		{
			name: "성공: OWNER가 프로젝트 삭제",
			mockProject: func(m *MockProjectRepository) {
				m.FindByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.Project, error) {
					return &domain.Project{
						BaseModel:   domain.BaseModel{ID: projectID, CreatedAt: time.Now(), UpdatedAt: time.Now()},
						WorkspaceID: workspaceID,
						OwnerID:     ownerID,
						Name:        "Test Project",
					}, nil
				}
				m.FindMemberByProjectAndUserFunc = func(ctx context.Context, pID, uID uuid.UUID) (*domain.ProjectMember, error) {
					return &domain.ProjectMember{
						ID:        uuid.New(),
						ProjectID: pID,
						UserID:    uID,
						RoleName:  domain.ProjectRoleOwner,
					}, nil
				}
				m.DeleteFunc = func(ctx context.Context, id uuid.UUID) error {
					return nil
				}
			},
			wantErr: false,
		},
		{
			name: "실패: OWNER가 아닌 사용자가 삭제 시도",
			mockProject: func(m *MockProjectRepository) {
				m.FindByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.Project, error) {
					return &domain.Project{
						BaseModel:   domain.BaseModel{ID: projectID, CreatedAt: time.Now(), UpdatedAt: time.Now()},
						WorkspaceID: workspaceID,
						OwnerID:     ownerID,
						Name:        "Test Project",
					}, nil
				}
				m.FindMemberByProjectAndUserFunc = func(ctx context.Context, pID, uID uuid.UUID) (*domain.ProjectMember, error) {
					return &domain.ProjectMember{
						ID:        uuid.New(),
						ProjectID: pID,
						UserID:    uID,
						RoleName:  domain.ProjectRoleMember,
					}, nil
				}
			},
			wantErr:     true,
			wantErrCode: response.ErrCodeForbidden,
		},
		{
			name: "실패: 프로젝트가 존재하지 않음",
			mockProject: func(m *MockProjectRepository) {
				m.FindByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.Project, error) {
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
			mockProjectRepo := &MockProjectRepository{}
			mockFieldOptionRepo := &MockFieldOptionRepository{}
			mockUserClient := &MockUserClient{}
			tt.mockProject(mockProjectRepo)

			service := NewProjectService(mockProjectRepo, mockFieldOptionRepo, mockUserClient, getTestLogger())

			// When
			err := service.DeleteProject(context.Background(), projectID, userID)

			// Then
			if tt.wantErr {
				if err == nil {
					t.Errorf("DeleteProject() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if appErr, ok := err.(*response.AppError); ok {
					if appErr.Code != tt.wantErrCode {
						t.Errorf("DeleteProject() error code = %v, want %v", appErr.Code, tt.wantErrCode)
					}
				}
			} else {
				if err != nil {
					t.Errorf("DeleteProject() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestProjectService_SearchProjects(t *testing.T) {
	workspaceID := uuid.New()
	userID := uuid.New()
	ownerID := uuid.New()
	token := "test-jwt-token"

	tests := []struct {
		name        string
		query       string
		page        int
		limit       int
		mockProject func(*MockProjectRepository)
		mockUser    func(*MockUserClient)
		wantErr     bool
		wantErrCode string
	}{
		{
			name:  "성공: 프로젝트 검색",
			query: "test",
			page:  1,
			limit: 10,
			mockProject: func(m *MockProjectRepository) {
				m.SearchFunc = func(ctx context.Context, wID uuid.UUID, q string, p, l int) ([]*domain.Project, int64, error) {
					return []*domain.Project{
						{
							BaseModel:   domain.BaseModel{ID: uuid.New(), CreatedAt: time.Now(), UpdatedAt: time.Now()},
							WorkspaceID: wID,
							OwnerID:     ownerID,
							Name:        "Test Project",
						},
					}, 1, nil
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
			wantErr: false,
		},
		{
			name:  "실패: 빈 검색어",
			query: "",
			page:  1,
			limit: 10,
			mockProject: func(m *MockProjectRepository) {},
			mockUser: func(m *MockUserClient) {
				m.ValidateWorkspaceMemberFunc = func(ctx context.Context, wID, uID uuid.UUID, t string) (bool, error) {
					return true, nil
				}
			},
			wantErr:     true,
			wantErrCode: response.ErrCodeValidation,
		},
		{
			name:  "실패: Workspace 멤버가 아님",
			query: "test",
			page:  1,
			limit: 10,
			mockProject: func(m *MockProjectRepository) {},
			mockUser: func(m *MockUserClient) {
				m.ValidateWorkspaceMemberFunc = func(ctx context.Context, wID, uID uuid.UUID, t string) (bool, error) {
					return false, nil
				}
			},
			wantErr:     true,
			wantErrCode: response.ErrCodeForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			mockProjectRepo := &MockProjectRepository{}
			mockFieldOptionRepo := &MockFieldOptionRepository{}
			mockUserClient := &MockUserClient{}
			tt.mockProject(mockProjectRepo)
			tt.mockUser(mockUserClient)

			service := NewProjectService(mockProjectRepo, mockFieldOptionRepo, mockUserClient, getTestLogger())

			// When
			got, err := service.SearchProjects(context.Background(), workspaceID, userID, tt.query, tt.page, tt.limit, token)

			// Then
			if tt.wantErr {
				if err == nil {
					t.Errorf("SearchProjects() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if appErr, ok := err.(*response.AppError); ok {
					if appErr.Code != tt.wantErrCode {
						t.Errorf("SearchProjects() error code = %v, want %v", appErr.Code, tt.wantErrCode)
					}
				}
			} else {
				if err != nil {
					t.Errorf("SearchProjects() unexpected error = %v", err)
					return
				}
				if got == nil {
					t.Error("SearchProjects() returned nil response")
				}
			}
		})
	}
}

func TestProjectService_GetProjectInitSettings(t *testing.T) {
	projectID := uuid.New()
	userID := uuid.New()
	ownerID := uuid.New()
	workspaceID := uuid.New()
	token := "test-jwt-token"

	tests := []struct {
		name             string
		mockProject      func(*MockProjectRepository)
		mockFieldOption  func(*MockFieldOptionRepository)
		mockUser         func(*MockUserClient)
		wantErr          bool
		wantErrCode      string
		wantFieldsCount  int
	}{
		{
			name: "성공: 새 프로젝트의 초기 설정 조회 - 12개의 field options 반환",
			mockProject: func(m *MockProjectRepository) {
				m.FindByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.Project, error) {
					return &domain.Project{
						BaseModel:   domain.BaseModel{ID: projectID, CreatedAt: time.Now(), UpdatedAt: time.Now()},
						WorkspaceID: workspaceID,
						OwnerID:     ownerID,
						Name:        "New Project",
						Description: "New Project Description",
						IsPublic:    true,
					}, nil
				}
				m.IsProjectMemberFunc = func(ctx context.Context, pID, uID uuid.UUID) (bool, error) {
					return true, nil
				}
			},
			mockFieldOption: func(m *MockFieldOptionRepository) {
				m.FindByProjectAndFieldTypeFunc = func(ctx context.Context, pID uuid.UUID, ft domain.FieldType) ([]*domain.FieldOption, error) {
					switch ft {
					case domain.FieldTypeStage:
						return []*domain.FieldOption{
							{BaseModel: domain.BaseModel{ID: uuid.New()}, ProjectID: &pID, FieldType: domain.FieldTypeStage, Value: "pending", Label: "대기", Color: "#F59E0B", DisplayOrder: 1},
							{BaseModel: domain.BaseModel{ID: uuid.New()}, ProjectID: &pID, FieldType: domain.FieldTypeStage, Value: "in_progress", Label: "진행중", Color: "#3B82F6", DisplayOrder: 2},
							{BaseModel: domain.BaseModel{ID: uuid.New()}, ProjectID: &pID, FieldType: domain.FieldTypeStage, Value: "review", Label: "검토", Color: "#8B5CF6", DisplayOrder: 3},
							{BaseModel: domain.BaseModel{ID: uuid.New()}, ProjectID: &pID, FieldType: domain.FieldTypeStage, Value: "approved", Label: "완료", Color: "#10B981", DisplayOrder: 4},
						}, nil
					case domain.FieldTypeImportance:
						return []*domain.FieldOption{
							{BaseModel: domain.BaseModel{ID: uuid.New()}, ProjectID: &pID, FieldType: domain.FieldTypeImportance, Value: "urgent", Label: "긴급", Color: "#EF4444", DisplayOrder: 1},
							{BaseModel: domain.BaseModel{ID: uuid.New()}, ProjectID: &pID, FieldType: domain.FieldTypeImportance, Value: "high", Label: "높음", Color: "#F97316", DisplayOrder: 2},
							{BaseModel: domain.BaseModel{ID: uuid.New()}, ProjectID: &pID, FieldType: domain.FieldTypeImportance, Value: "normal", Label: "보통", Color: "#10B981", DisplayOrder: 3},
							{BaseModel: domain.BaseModel{ID: uuid.New()}, ProjectID: &pID, FieldType: domain.FieldTypeImportance, Value: "low", Label: "낮음", Color: "#6B7280", DisplayOrder: 4},
						}, nil
					case domain.FieldTypeRole:
						return []*domain.FieldOption{
							{BaseModel: domain.BaseModel{ID: uuid.New()}, ProjectID: &pID, FieldType: domain.FieldTypeRole, Value: "developer", Label: "개발자", Color: "#8B5CF6", DisplayOrder: 1},
							{BaseModel: domain.BaseModel{ID: uuid.New()}, ProjectID: &pID, FieldType: domain.FieldTypeRole, Value: "planner", Label: "기획자", Color: "#EC4899", DisplayOrder: 2},
							{BaseModel: domain.BaseModel{ID: uuid.New()}, ProjectID: &pID, FieldType: domain.FieldTypeRole, Value: "designer", Label: "디자이너", Color: "#F59E0B", DisplayOrder: 3},
							{BaseModel: domain.BaseModel{ID: uuid.New()}, ProjectID: &pID, FieldType: domain.FieldTypeRole, Value: "qa", Label: "QA", Color: "#06B6D4", DisplayOrder: 4},
						}, nil
					default:
						return []*domain.FieldOption{}, nil
					}
				}
			},
			mockUser: func(m *MockUserClient) {
				m.GetWorkspaceFunc = func(ctx context.Context, wID uuid.UUID, t string) (*client.Workspace, error) {
					return &client.Workspace{
						ID:         wID,
						Name:       "Test Workspace",
						OwnerEmail: "workspace@example.com",
					}, nil
				}
			},
			wantErr:         false,
			wantFieldsCount: 3,
		},
		{
			name: "성공: 프로젝트 초기 설정 조회 with 필드 옵션",
			mockProject: func(m *MockProjectRepository) {
				m.FindByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.Project, error) {
					return &domain.Project{
						BaseModel:   domain.BaseModel{ID: projectID, CreatedAt: time.Now(), UpdatedAt: time.Now()},
						WorkspaceID: workspaceID,
						OwnerID:     ownerID,
						Name:        "Test Project",
						Description: "Test Description",
						IsPublic:    true,
					}, nil
				}
				m.IsProjectMemberFunc = func(ctx context.Context, pID, uID uuid.UUID) (bool, error) {
					return true, nil
				}
			},
			mockFieldOption: func(m *MockFieldOptionRepository) {
				m.FindByProjectAndFieldTypeFunc = func(ctx context.Context, pID uuid.UUID, ft domain.FieldType) ([]*domain.FieldOption, error) {
					switch ft {
					case domain.FieldTypeStage:
						return []*domain.FieldOption{
							{
								BaseModel:    domain.BaseModel{ID: uuid.New()},
								ProjectID:    &pID,
								FieldType:    domain.FieldTypeStage,
								Value:        "pending",
								Label:        "대기",
								Color:        "#F59E0B",
								DisplayOrder: 1,
							},
							{
								BaseModel:    domain.BaseModel{ID: uuid.New()},
								ProjectID:    &pID,
								FieldType:    domain.FieldTypeStage,
								Value:        "in_progress",
								Label:        "진행중",
								Color:        "#3B82F6",
								DisplayOrder: 2,
							},
						}, nil
					case domain.FieldTypeRole:
						return []*domain.FieldOption{
							{
								BaseModel:    domain.BaseModel{ID: uuid.New()},
								ProjectID:    &pID,
								FieldType:    domain.FieldTypeRole,
								Value:        "developer",
								Label:        "개발자",
								Color:        "#8B5CF6",
								DisplayOrder: 1,
							},
						}, nil
					case domain.FieldTypeImportance:
						return []*domain.FieldOption{
							{
								BaseModel:    domain.BaseModel{ID: uuid.New()},
								ProjectID:    &pID,
								FieldType:    domain.FieldTypeImportance,
								Value:        "urgent",
								Label:        "긴급",
								Color:        "#EF4444",
								DisplayOrder: 1,
							},
						}, nil
					default:
						return []*domain.FieldOption{}, nil
					}
				}
			},
			mockUser: func(m *MockUserClient) {
				m.GetWorkspaceFunc = func(ctx context.Context, wID uuid.UUID, t string) (*client.Workspace, error) {
					return &client.Workspace{
						ID:         wID,
						Name:       "Test Workspace",
						OwnerEmail: "workspace@example.com",
					}, nil
				}
			},
			wantErr:         false,
			wantFieldsCount: 3,
		},
		{
			name: "성공: Workspace 조회 실패 시 graceful degradation",
			mockProject: func(m *MockProjectRepository) {
				m.FindByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.Project, error) {
					return &domain.Project{
						BaseModel:   domain.BaseModel{ID: projectID, CreatedAt: time.Now(), UpdatedAt: time.Now()},
						WorkspaceID: workspaceID,
						OwnerID:     ownerID,
						Name:        "Test Project",
						Description: "Test Description",
						IsPublic:    true,
					}, nil
				}
				m.IsProjectMemberFunc = func(ctx context.Context, pID, uID uuid.UUID) (bool, error) {
					return true, nil
				}
			},
			mockFieldOption: func(m *MockFieldOptionRepository) {
				m.FindByFieldTypeFunc = func(ctx context.Context, ft domain.FieldType) ([]*domain.FieldOption, error) {
					return []*domain.FieldOption{}, nil
				}
			},
			mockUser: func(m *MockUserClient) {
				m.GetWorkspaceFunc = func(ctx context.Context, wID uuid.UUID, t string) (*client.Workspace, error) {
					return nil, errors.New("workspace API error")
				}
			},
			wantErr:         false,
			wantFieldsCount: 3,
		},
		{
			name: "실패: 프로젝트가 존재하지 않음",
			mockProject: func(m *MockProjectRepository) {
				m.FindByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.Project, error) {
					return nil, gorm.ErrRecordNotFound
				}
			},
			mockFieldOption: func(m *MockFieldOptionRepository) {},
			mockUser:        func(m *MockUserClient) {},
			wantErr:         true,
			wantErrCode:     response.ErrCodeNotFound,
		},
		{
			name: "실패: 프로젝트 멤버가 아님",
			mockProject: func(m *MockProjectRepository) {
				m.FindByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.Project, error) {
					return &domain.Project{
						BaseModel:   domain.BaseModel{ID: projectID, CreatedAt: time.Now(), UpdatedAt: time.Now()},
						WorkspaceID: workspaceID,
						OwnerID:     ownerID,
						Name:        "Test Project",
					}, nil
				}
				m.IsProjectMemberFunc = func(ctx context.Context, pID, uID uuid.UUID) (bool, error) {
					return false, nil
				}
			},
			mockFieldOption: func(m *MockFieldOptionRepository) {},
			mockUser: func(m *MockUserClient) {
				m.ValidateWorkspaceMemberFunc = func(ctx context.Context, wID, uID uuid.UUID, t string) (bool, error) {
					return false, nil
				}
			},
			wantErr:         true,
			wantErrCode:     response.ErrCodeForbidden,
		},
		{
			name: "실패: Stage 옵션 조회 실패",
			mockProject: func(m *MockProjectRepository) {
				m.FindByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.Project, error) {
					return &domain.Project{
						BaseModel:   domain.BaseModel{ID: projectID, CreatedAt: time.Now(), UpdatedAt: time.Now()},
						WorkspaceID: workspaceID,
						OwnerID:     ownerID,
						Name:        "Test Project",
					}, nil
				}
				m.IsProjectMemberFunc = func(ctx context.Context, pID, uID uuid.UUID) (bool, error) {
					return true, nil
				}
			},
			mockFieldOption: func(m *MockFieldOptionRepository) {
				m.FindByProjectAndFieldTypeFunc = func(ctx context.Context, pID uuid.UUID, ft domain.FieldType) ([]*domain.FieldOption, error) {
					if ft == domain.FieldTypeStage {
						return nil, errors.New("database error")
					}
					return []*domain.FieldOption{}, nil
				}
			},
			mockUser: func(m *MockUserClient) {
				m.GetWorkspaceFunc = func(ctx context.Context, wID uuid.UUID, t string) (*client.Workspace, error) {
					return &client.Workspace{
						ID:   wID,
						Name: "Test Workspace",
					}, nil
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
			mockFieldOptionRepo := &MockFieldOptionRepository{}
			mockUserClient := &MockUserClient{}
			tt.mockProject(mockProjectRepo)
			tt.mockFieldOption(mockFieldOptionRepo)
			tt.mockUser(mockUserClient)

			service := NewProjectService(mockProjectRepo, mockFieldOptionRepo, mockUserClient, getTestLogger())

			// When
			got, err := service.GetProjectInitSettings(context.Background(), projectID, userID, token)

			// Then
			if tt.wantErr {
				if err == nil {
					t.Errorf("GetProjectInitSettings() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if appErr, ok := err.(*response.AppError); ok {
					if appErr.Code != tt.wantErrCode {
						t.Errorf("GetProjectInitSettings() error code = %v, want %v", appErr.Code, tt.wantErrCode)
					}
				}
			} else {
				if err != nil {
					t.Errorf("GetProjectInitSettings() unexpected error = %v", err)
					return
				}
				if got == nil {
					t.Error("GetProjectInitSettings() returned nil response")
					return
				}
				if len(got.Fields) != tt.wantFieldsCount {
					t.Errorf("GetProjectInitSettings() fields count = %v, want %v", len(got.Fields), tt.wantFieldsCount)
				}
				if got.Project.ProjectID != projectID {
					t.Errorf("GetProjectInitSettings() project ID = %v, want %v", got.Project.ProjectID, projectID)
				}
			}
		})
	}
}

func TestProjectService_GetProjectInitSettings_FieldOptionsCount(t *testing.T) {
	projectID := uuid.New()
	userID := uuid.New()
	ownerID := uuid.New()
	workspaceID := uuid.New()
	token := "test-jwt-token"

	t.Run("성공: 각 필드별로 정확히 4개의 옵션 반환", func(t *testing.T) {
		// Given
		mockProjectRepo := &MockProjectRepository{}
		mockFieldOptionRepo := &MockFieldOptionRepository{}
		mockUserClient := &MockUserClient{}

		mockProjectRepo.FindByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.Project, error) {
			return &domain.Project{
				BaseModel:   domain.BaseModel{ID: projectID, CreatedAt: time.Now(), UpdatedAt: time.Now()},
				WorkspaceID: workspaceID,
				OwnerID:     ownerID,
				Name:        "New Project",
				Description: "New Project Description",
				IsPublic:    true,
			}, nil
		}
		mockProjectRepo.IsProjectMemberFunc = func(ctx context.Context, pID, uID uuid.UUID) (bool, error) {
			return true, nil
		}

		mockFieldOptionRepo.FindByProjectAndFieldTypeFunc = func(ctx context.Context, pID uuid.UUID, ft domain.FieldType) ([]*domain.FieldOption, error) {
			switch ft {
			case domain.FieldTypeStage:
				return []*domain.FieldOption{
					{BaseModel: domain.BaseModel{ID: uuid.New()}, ProjectID: &pID, FieldType: domain.FieldTypeStage, Value: "pending", Label: "대기", Color: "#F59E0B", DisplayOrder: 1},
					{BaseModel: domain.BaseModel{ID: uuid.New()}, ProjectID: &pID, FieldType: domain.FieldTypeStage, Value: "in_progress", Label: "진행중", Color: "#3B82F6", DisplayOrder: 2},
					{BaseModel: domain.BaseModel{ID: uuid.New()}, ProjectID: &pID, FieldType: domain.FieldTypeStage, Value: "review", Label: "검토", Color: "#8B5CF6", DisplayOrder: 3},
					{BaseModel: domain.BaseModel{ID: uuid.New()}, ProjectID: &pID, FieldType: domain.FieldTypeStage, Value: "approved", Label: "완료", Color: "#10B981", DisplayOrder: 4},
				}, nil
			case domain.FieldTypeImportance:
				return []*domain.FieldOption{
					{BaseModel: domain.BaseModel{ID: uuid.New()}, ProjectID: &pID, FieldType: domain.FieldTypeImportance, Value: "urgent", Label: "긴급", Color: "#EF4444", DisplayOrder: 1},
					{BaseModel: domain.BaseModel{ID: uuid.New()}, ProjectID: &pID, FieldType: domain.FieldTypeImportance, Value: "high", Label: "높음", Color: "#F97316", DisplayOrder: 2},
					{BaseModel: domain.BaseModel{ID: uuid.New()}, ProjectID: &pID, FieldType: domain.FieldTypeImportance, Value: "normal", Label: "보통", Color: "#10B981", DisplayOrder: 3},
					{BaseModel: domain.BaseModel{ID: uuid.New()}, ProjectID: &pID, FieldType: domain.FieldTypeImportance, Value: "low", Label: "낮음", Color: "#6B7280", DisplayOrder: 4},
				}, nil
			case domain.FieldTypeRole:
				return []*domain.FieldOption{
					{BaseModel: domain.BaseModel{ID: uuid.New()}, ProjectID: &pID, FieldType: domain.FieldTypeRole, Value: "developer", Label: "개발자", Color: "#8B5CF6", DisplayOrder: 1},
					{BaseModel: domain.BaseModel{ID: uuid.New()}, ProjectID: &pID, FieldType: domain.FieldTypeRole, Value: "planner", Label: "기획자", Color: "#EC4899", DisplayOrder: 2},
					{BaseModel: domain.BaseModel{ID: uuid.New()}, ProjectID: &pID, FieldType: domain.FieldTypeRole, Value: "designer", Label: "디자이너", Color: "#F59E0B", DisplayOrder: 3},
					{BaseModel: domain.BaseModel{ID: uuid.New()}, ProjectID: &pID, FieldType: domain.FieldTypeRole, Value: "qa", Label: "QA", Color: "#06B6D4", DisplayOrder: 4},
				}, nil
			default:
				return []*domain.FieldOption{}, nil
			}
		}

		mockUserClient.GetWorkspaceFunc = func(ctx context.Context, wID uuid.UUID, t string) (*client.Workspace, error) {
			return &client.Workspace{
				ID:         wID,
				Name:       "Test Workspace",
				OwnerEmail: "workspace@example.com",
			}, nil
		}

		service := NewProjectService(mockProjectRepo, mockFieldOptionRepo, mockUserClient, getTestLogger())

		// When
		got, err := service.GetProjectInitSettings(context.Background(), projectID, userID, token)

		// Then
		if err != nil {
			t.Fatalf("GetProjectInitSettings() unexpected error = %v", err)
		}
		if got == nil {
			t.Fatal("GetProjectInitSettings() returned nil response")
		}

		// Verify 3 fields are returned
		if len(got.Fields) != 3 {
			t.Errorf("Expected 3 fields, got %d", len(got.Fields))
		}

		// Count total options across all fields
		totalOptions := 0
		for _, field := range got.Fields {
			totalOptions += len(field.Options)
		}

		// Verify 12 total field options are returned
		if totalOptions != 12 {
			t.Errorf("Expected 12 total field options, got %d", totalOptions)
		}

		// Verify stage field has 4 options
		var stageField *dto.FieldWithOptionsResponse
		for i := range got.Fields {
			if got.Fields[i].FieldID == "stage" {
				stageField = &got.Fields[i]
				break
			}
		}
		if stageField == nil {
			t.Error("Stage field not found")
		} else if len(stageField.Options) != 4 {
			t.Errorf("Expected stage field to have 4 options, got %d", len(stageField.Options))
		}

		// Verify importance field has 4 options
		var importanceField *dto.FieldWithOptionsResponse
		for i := range got.Fields {
			if got.Fields[i].FieldID == "importance" {
				importanceField = &got.Fields[i]
				break
			}
		}
		if importanceField == nil {
			t.Error("Importance field not found")
		} else if len(importanceField.Options) != 4 {
			t.Errorf("Expected importance field to have 4 options, got %d", len(importanceField.Options))
		}

		// Verify role field has 4 options
		var roleField *dto.FieldWithOptionsResponse
		for i := range got.Fields {
			if got.Fields[i].FieldID == "role" {
				roleField = &got.Fields[i]
				break
			}
		}
		if roleField == nil {
			t.Error("Role field not found")
		} else if len(roleField.Options) != 4 {
			t.Errorf("Expected role field to have 4 options, got %d", len(roleField.Options))
		}
	})
}

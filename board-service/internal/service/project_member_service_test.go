package service

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"project-board-api/internal/domain"
	"project-board-api/internal/response"
)

func TestProjectMemberService_GetMembers(t *testing.T) {
	projectID := uuid.New()
	userID := uuid.New()
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
			name: "성공: 멤버 목록 조회",
			mockProject: func(m *MockProjectRepository) {
				m.IsProjectMemberFunc = func(ctx context.Context, pID, uID uuid.UUID) (bool, error) {
					return true, nil
				}
				m.FindByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.Project, error) {
					return &domain.Project{
						BaseModel:   domain.BaseModel{ID: projectID},
						WorkspaceID: workspaceID,
						Name:        "Test Project",
					}, nil
				}
				m.FindMembersByProjectIDFunc = func(ctx context.Context, pID uuid.UUID) ([]*domain.ProjectMember, error) {
					return []*domain.ProjectMember{
						{
							ID:        uuid.New(),
							ProjectID: pID,
							UserID:    uuid.New(),
							RoleName:  domain.ProjectRoleOwner,
						},
					}, nil
				}
			},
			mockUser: func(m *MockUserClient) {
				m.GetWorkspaceProfileFunc = func(ctx context.Context, wID, uID uuid.UUID, t string) (*client.WorkspaceProfile, error) {
					return &client.WorkspaceProfile{
						WorkspaceID: wID,
						UserID:      uID,
						NickName:    "Test User",
						Email:       "test@example.com",
					}, nil
				}
			},
			wantErr: false,
		},
		{
			name: "실패: 프로젝트 멤버가 아님",
			mockProject: func(m *MockProjectRepository) {
				m.IsProjectMemberFunc = func(ctx context.Context, pID, uID uuid.UUID) (bool, error) {
					return false, nil
				}
			},
			mockUser:    func(m *MockUserClient) {},
			wantErr:     true,
			wantErrCode: response.ErrCodeForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			mockProjectRepo := &MockProjectRepository{}
			mockUserClient := &MockUserClient{}
			tt.mockProject(mockProjectRepo)
			tt.mockUser(mockUserClient)

			service := NewProjectMemberService(mockProjectRepo, mockUserClient)

			// When
			got, err := service.GetMembers(context.Background(), projectID, userID, token)

			// Then
			if tt.wantErr {
				if err == nil {
					t.Errorf("GetMembers() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if appErr, ok := err.(*response.AppError); ok {
					if appErr.Code != tt.wantErrCode {
						t.Errorf("GetMembers() error code = %v, want %v", appErr.Code, tt.wantErrCode)
					}
				}
			} else {
				if err != nil {
					t.Errorf("GetMembers() unexpected error = %v", err)
					return
				}
				if got == nil {
					t.Error("GetMembers() returned nil response")
				}
			}
		})
	}
}

func TestProjectMemberService_RemoveMember(t *testing.T) {
	projectID := uuid.New()
	requesterID := uuid.New()
	memberID := uuid.New()

	tests := []struct {
		name        string
		mockProject func(*MockProjectRepository)
		wantErr     bool
		wantErrCode string
	}{
		{
			name: "성공: ADMIN이 멤버 제거",
			mockProject: func(m *MockProjectRepository) {
				m.FindMemberByProjectAndUserFunc = func(ctx context.Context, pID, uID uuid.UUID) (*domain.ProjectMember, error) {
					if uID == requesterID {
						return &domain.ProjectMember{
							ID:        uuid.New(),
							ProjectID: pID,
							UserID:    uID,
							RoleName:  domain.ProjectRoleAdmin,
						}, nil
					}
					return &domain.ProjectMember{
						ID:        uuid.New(),
						ProjectID: pID,
						UserID:    uID,
						RoleName:  domain.ProjectRoleMember,
					}, nil
				}
				m.RemoveMemberFunc = func(ctx context.Context, memberID uuid.UUID) error {
					return nil
				}
			},
			wantErr: false,
		},
		{
			name: "성공: OWNER가 멤버 제거",
			mockProject: func(m *MockProjectRepository) {
				m.FindMemberByProjectAndUserFunc = func(ctx context.Context, pID, uID uuid.UUID) (*domain.ProjectMember, error) {
					if uID == requesterID {
						return &domain.ProjectMember{
							ID:        uuid.New(),
							ProjectID: pID,
							UserID:    uID,
							RoleName:  domain.ProjectRoleOwner,
						}, nil
					}
					return &domain.ProjectMember{
						ID:        uuid.New(),
						ProjectID: pID,
						UserID:    uID,
						RoleName:  domain.ProjectRoleMember,
					}, nil
				}
				m.RemoveMemberFunc = func(ctx context.Context, memberID uuid.UUID) error {
					return nil
				}
			},
			wantErr: false,
		},
		{
			name: "실패: OWNER 제거 시도",
			mockProject: func(m *MockProjectRepository) {
				m.FindMemberByProjectAndUserFunc = func(ctx context.Context, pID, uID uuid.UUID) (*domain.ProjectMember, error) {
					if uID == requesterID {
						return &domain.ProjectMember{
							ID:        uuid.New(),
							ProjectID: pID,
							UserID:    uID,
							RoleName:  domain.ProjectRoleAdmin,
						}, nil
					}
					return &domain.ProjectMember{
						ID:        uuid.New(),
						ProjectID: pID,
						UserID:    uID,
						RoleName:  domain.ProjectRoleOwner,
					}, nil
				}
			},
			wantErr:     true,
			wantErrCode: response.ErrCodeValidation,
		},
		{
			name: "실패: 자기 자신 제거 시도",
			mockProject: func(m *MockProjectRepository) {
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
			wantErrCode: response.ErrCodeValidation,
		},
		{
			name: "실패: MEMBER가 제거 시도",
			mockProject: func(m *MockProjectRepository) {
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			mockProjectRepo := &MockProjectRepository{}
			mockUserClient := &MockUserClient{}
			tt.mockProject(mockProjectRepo)

			service := NewProjectMemberService(mockProjectRepo, mockUserClient)

			// When
			err := service.RemoveMember(context.Background(), projectID, requesterID, memberID)

			// Then
			if tt.wantErr {
				if err == nil {
					t.Errorf("RemoveMember() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if appErr, ok := err.(*response.AppError); ok {
					if appErr.Code != tt.wantErrCode {
						t.Errorf("RemoveMember() error code = %v, want %v", appErr.Code, tt.wantErrCode)
					}
				}
			} else {
				if err != nil {
					t.Errorf("RemoveMember() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestProjectMemberService_UpdateMemberRole(t *testing.T) {
	projectID := uuid.New()
	requesterID := uuid.New()
	memberID := uuid.New()

	tests := []struct {
		name        string
		role        string
		mockProject func(*MockProjectRepository)
		wantErr     bool
		wantErrCode string
	}{
		{
			name: "성공: MEMBER를 ADMIN으로 변경",
			role: "ADMIN",
			mockProject: func(m *MockProjectRepository) {
				m.FindMemberByProjectAndUserFunc = func(ctx context.Context, pID, uID uuid.UUID) (*domain.ProjectMember, error) {
					if uID == requesterID {
						return &domain.ProjectMember{
							ID:        uuid.New(),
							ProjectID: pID,
							UserID:    uID,
							RoleName:  domain.ProjectRoleOwner,
						}, nil
					}
					return &domain.ProjectMember{
						ID:        uuid.New(),
						ProjectID: pID,
						UserID:    uID,
						RoleName:  domain.ProjectRoleMember,
					}, nil
				}
				m.UpdateMemberRoleFunc = func(ctx context.Context, memberID uuid.UUID, role domain.ProjectRole) error {
					return nil
				}
			},
			wantErr: false,
		},
		{
			name: "성공: ADMIN을 MEMBER로 변경",
			role: "MEMBER",
			mockProject: func(m *MockProjectRepository) {
				m.FindMemberByProjectAndUserFunc = func(ctx context.Context, pID, uID uuid.UUID) (*domain.ProjectMember, error) {
					if uID == requesterID {
						return &domain.ProjectMember{
							ID:        uuid.New(),
							ProjectID: pID,
							UserID:    uID,
							RoleName:  domain.ProjectRoleOwner,
						}, nil
					}
					return &domain.ProjectMember{
						ID:        uuid.New(),
						ProjectID: pID,
						UserID:    uID,
						RoleName:  domain.ProjectRoleAdmin,
					}, nil
				}
				m.UpdateMemberRoleFunc = func(ctx context.Context, memberID uuid.UUID, role domain.ProjectRole) error {
					return nil
				}
			},
			wantErr: false,
		},
		{
			name: "실패: OWNER 역할 변경 시도",
			role: "ADMIN",
			mockProject: func(m *MockProjectRepository) {
				m.FindMemberByProjectAndUserFunc = func(ctx context.Context, pID, uID uuid.UUID) (*domain.ProjectMember, error) {
					if uID == requesterID {
						return &domain.ProjectMember{
							ID:        uuid.New(),
							ProjectID: pID,
							UserID:    uID,
							RoleName:  domain.ProjectRoleOwner,
						}, nil
					}
					return &domain.ProjectMember{
						ID:        uuid.New(),
						ProjectID: pID,
						UserID:    uID,
						RoleName:  domain.ProjectRoleOwner,
					}, nil
				}
			},
			wantErr:     true,
			wantErrCode: response.ErrCodeValidation,
		},
		{
			name: "실패: ADMIN이 역할 변경 시도",
			role: "MEMBER",
			mockProject: func(m *MockProjectRepository) {
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
			name: "실패: 유효하지 않은 역할",
			role: "INVALID",
			mockProject: func(m *MockProjectRepository) {
				m.FindMemberByProjectAndUserFunc = func(ctx context.Context, pID, uID uuid.UUID) (*domain.ProjectMember, error) {
					if uID == requesterID {
						return &domain.ProjectMember{
							ID:        uuid.New(),
							ProjectID: pID,
							UserID:    uID,
							RoleName:  domain.ProjectRoleOwner,
						}, nil
					}
					return &domain.ProjectMember{
						ID:        uuid.New(),
						ProjectID: pID,
						UserID:    uID,
						RoleName:  domain.ProjectRoleMember,
					}, nil
				}
			},
			wantErr:     true,
			wantErrCode: response.ErrCodeValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			mockProjectRepo := &MockProjectRepository{}
			mockUserClient := &MockUserClient{}
			tt.mockProject(mockProjectRepo)

			service := NewProjectMemberService(mockProjectRepo, mockUserClient)

			// When
			_, err := service.UpdateMemberRole(context.Background(), projectID, requesterID, memberID, tt.role)

			// Then
			if tt.wantErr {
				if err == nil {
					t.Errorf("UpdateMemberRole() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if appErr, ok := err.(*response.AppError); ok {
					if appErr.Code != tt.wantErrCode {
						t.Errorf("UpdateMemberRole() error code = %v, want %v", appErr.Code, tt.wantErrCode)
					}
				}
			} else {
				if err != nil {
					t.Errorf("UpdateMemberRole() unexpected error = %v", err)
				}
			}
		})
	}
}

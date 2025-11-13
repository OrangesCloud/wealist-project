package service

import (
	"board-service/internal/domain"
	"board-service/internal/testutil"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zaptest"
	"gorm.io/gorm"
)

// ==================== Additional Mocks ====================

type MockProjectFieldRepository struct {
	mock.Mock
}

func (m *MockProjectFieldRepository) Create(field *domain.ProjectField) error {
	args := m.Called(field)
	return args.Error(0)
}

func (m *MockProjectFieldRepository) FindByID(id uuid.UUID) (*domain.ProjectField, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ProjectField), args.Error(1)
}

func (m *MockProjectFieldRepository) Update(field *domain.ProjectField) error {
	args := m.Called(field)
	return args.Error(0)
}

func (m *MockProjectFieldRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockProjectFieldRepository) FindByProject(projectID uuid.UUID) ([]domain.ProjectField, error) {
	args := m.Called(projectID)
	return args.Get(0).([]domain.ProjectField), args.Error(1)
}

func (m *MockProjectFieldRepository) FindByIDs(ids []uuid.UUID) ([]domain.ProjectField, error) {
	args := m.Called(ids)
	return args.Get(0).([]domain.ProjectField), args.Error(1)
}

func (m *MockProjectFieldRepository) UpdateOrder(fieldID uuid.UUID, newOrder int) error {
	args := m.Called(fieldID, newOrder)
	return args.Error(0)
}

func (m *MockProjectFieldRepository) BatchUpdateOrders(orders map[uuid.UUID]int) error {
	args := m.Called(orders)
	return args.Error(0)
}

type MockFieldOptionRepository struct {
	mock.Mock
}

func (m *MockFieldOptionRepository) Create(option *domain.FieldOption) error {
	args := m.Called(option)
	return args.Error(0)
}

func (m *MockFieldOptionRepository) FindByID(id uuid.UUID) (*domain.FieldOption, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.FieldOption), args.Error(1)
}

func (m *MockFieldOptionRepository) Update(option *domain.FieldOption) error {
	args := m.Called(option)
	return args.Error(0)
}

func (m *MockFieldOptionRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockFieldOptionRepository) FindByField(fieldID uuid.UUID) ([]domain.FieldOption, error) {
	args := m.Called(fieldID)
	return args.Get(0).([]domain.FieldOption), args.Error(1)
}

func (m *MockFieldOptionRepository) FindByIDs(ids []uuid.UUID) ([]domain.FieldOption, error) {
	args := m.Called(ids)
	return args.Get(0).([]domain.FieldOption), args.Error(1)
}

func (m *MockFieldOptionRepository) UpdateOrder(optionID uuid.UUID, newOrder int) error {
	args := m.Called(optionID, newOrder)
	return args.Error(0)
}

func (m *MockFieldOptionRepository) BatchUpdateOrders(orders map[uuid.UUID]int) error {
	args := m.Called(orders)
	return args.Error(0)
}

type MockBoardOrderRepository struct {
	mock.Mock
}

func (m *MockBoardOrderRepository) Set(order *domain.UserBoardOrder) error {
	args := m.Called(order)
	return args.Error(0)
}

func (m *MockBoardOrderRepository) FindByView(viewID, userID uuid.UUID) ([]domain.UserBoardOrder, error) {
	args := m.Called(viewID, userID)
	return args.Get(0).([]domain.UserBoardOrder), args.Error(1)
}

func (m *MockBoardOrderRepository) BatchUpdate(orders []domain.UserBoardOrder) error {
	args := m.Called(orders)
	return args.Error(0)
}

func (m *MockBoardOrderRepository) Delete(viewID, userID, boardID uuid.UUID) error {
	args := m.Called(viewID, userID, boardID)
	return args.Error(0)
}

type MockViewRepository struct {
	mock.Mock
}

func (m *MockViewRepository) Create(view *domain.SavedView) error {
	args := m.Called(view)
	return args.Error(0)
}

func (m *MockViewRepository) FindByID(id uuid.UUID) (*domain.SavedView, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.SavedView), args.Error(1)
}

func (m *MockViewRepository) Update(view *domain.SavedView) error {
	args := m.Called(view)
	return args.Error(0)
}

func (m *MockViewRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockViewRepository) FindByProject(projectID uuid.UUID) ([]domain.SavedView, error) {
	args := m.Called(projectID)
	return args.Get(0).([]domain.SavedView), args.Error(1)
}

func (m *MockViewRepository) FindDefault(projectID uuid.UUID) (*domain.SavedView, error) {
	args := m.Called(projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.SavedView), args.Error(1)
}

// ==================== GetProjectInitSettings Tests ====================

func TestProjectService_GetProjectInitSettings_Success(t *testing.T) {
	// Setup
	projectRepo := new(testutil.MockProjectRepository)
	roleRepo := new(testutil.MockRoleRepository)
	fieldRepo := new(testutil.MockFieldRepository)
	boardRepo := new(testutil.MockBoardRepository)
	projectFieldRepo := new(MockProjectFieldRepository)
	fieldOptionRepo := new(MockFieldOptionRepository)
	boardOrderRepo := new(MockBoardOrderRepository)
	viewRepo := new(MockViewRepository)
	userClient := new(MockUserClient)
	workspaceCache := new(MockWorkspaceCache)
	userInfoCache := new(MockUserInfoCache)
	logger := zaptest.NewLogger(t)

	service := NewProjectService(
		projectRepo,
		roleRepo,
		fieldRepo,
		boardRepo,
		projectFieldRepo,
		fieldOptionRepo,
		boardOrderRepo,
		viewRepo,
		userClient,
		workspaceCache,
		userInfoCache,
		logger,
		nil,
	)

	// Test data
	projectID := uuid.New()
	userID := uuid.New()
	workspaceID := uuid.New()
	ownerID := uuid.New()
	fieldID := uuid.New()
	optionID := uuid.New()
	viewID := uuid.New()

	project := &domain.Project{
		BaseModel:   domain.BaseModel{ID: projectID, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		WorkspaceID: workspaceID,
		Name:        "Test Project",
		Description: "Test Description",
		OwnerID:     ownerID,
		IsPublic:    false,
	}

	member := &domain.ProjectMember{
		ProjectID: projectID,
		UserID:    userID,
		RoleID:    uuid.New(),
		Role:      &domain.Role{Name: "MEMBER", Level: 10},
		JoinedAt:  time.Now(),
	}

	defaultView := &domain.SavedView{
		BaseModel: domain.BaseModel{ID: viewID},
		ProjectID: projectID,
		IsDefault: true,
	}

	fields := []domain.ProjectField{
		{
			BaseModel:   domain.BaseModel{ID: fieldID, CreatedAt: time.Now(), UpdatedAt: time.Now()},
			ProjectID:   projectID,
			Name:        "Status",
			FieldType:   domain.FieldTypeSingleSelect,
			Description: "Board status",
			Config:      "{}",
		},
	}

	options := []domain.FieldOption{
		{
			BaseModel:    domain.BaseModel{ID: optionID, CreatedAt: time.Now(), UpdatedAt: time.Now()},
			FieldID:      fieldID,
			Label:        "To Do",
			Color:        "#94A3B8",
			DisplayOrder: 0,
		},
	}

	// Mock expectations
	projectRepo.On("FindByID", projectID).Return(project, nil)
	projectRepo.On("FindMemberByUserAndProject", userID, projectID).Return(member, nil)
	viewRepo.On("FindDefault", projectID).Return(defaultView, nil)
	projectFieldRepo.On("FindByProject", projectID).Return(fields, nil)
	fieldOptionRepo.On("FindByField", fieldID).Return(options, nil)

	// When
	result, err := service.GetProjectInitSettings(projectID.String(), userID.String())

	// Then
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Verify project info
	assert.Equal(t, projectID.String(), result.Project.ProjectID)
	assert.Equal(t, "Test Project", result.Project.Name)
	assert.Equal(t, "Test Description", result.Project.Description)
	assert.False(t, result.Project.IsPublic)

	// Verify fields with options
	assert.Len(t, result.Fields, 1)
	assert.Equal(t, fieldID.String(), result.Fields[0].FieldID)
	assert.Equal(t, "Status", result.Fields[0].Name)
	assert.Equal(t, "single_select", result.Fields[0].FieldType)
	assert.Len(t, result.Fields[0].Options, 1)
	assert.Equal(t, optionID.String(), result.Fields[0].Options[0].OptionID)
	assert.Equal(t, "To Do", result.Fields[0].Options[0].Label)
	assert.Equal(t, "#94A3B8", result.Fields[0].Options[0].Color)

	// Verify field types
	assert.Len(t, result.FieldTypes, 10)
	assert.Equal(t, "text", result.FieldTypes[0].Type)
	assert.Equal(t, "텍스트", result.FieldTypes[0].DisplayName)

	// Verify default view ID
	assert.Equal(t, viewID.String(), result.DefaultViewID)

	// Verify all mocks were called
	projectRepo.AssertExpectations(t)
	projectFieldRepo.AssertExpectations(t)
	fieldOptionRepo.AssertExpectations(t)
	viewRepo.AssertExpectations(t)
}

func TestProjectService_GetProjectInitSettings_ProjectNotFound(t *testing.T) {
	// Setup
	projectRepo := new(testutil.MockProjectRepository)
	logger := zaptest.NewLogger(t)

	service := NewProjectService(
		projectRepo,
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		logger,
		nil,
	)

	projectID := uuid.New()
	userID := uuid.New()

	projectRepo.On("FindByID", projectID).Return(nil, gorm.ErrRecordNotFound)

	// When
	result, err := service.GetProjectInitSettings(projectID.String(), userID.String())

	// Then
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "프로젝트를 찾을 수 없습니다")
}

func TestProjectService_GetProjectInitSettings_NotProjectMember(t *testing.T) {
	// Setup
	projectRepo := new(testutil.MockProjectRepository)
	logger := zaptest.NewLogger(t)

	service := NewProjectService(
		projectRepo,
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		logger,
		nil,
	)

	projectID := uuid.New()
	userID := uuid.New()

	project := &domain.Project{
		BaseModel: domain.BaseModel{ID: projectID},
	}

	projectRepo.On("FindByID", projectID).Return(project, nil)
	projectRepo.On("FindMemberByUserAndProject", userID, projectID).Return(nil, gorm.ErrRecordNotFound)

	// When
	result, err := service.GetProjectInitSettings(projectID.String(), userID.String())

	// Then
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "프로젝트 멤버가 아닙니다")
}

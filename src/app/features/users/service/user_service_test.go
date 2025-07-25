package service

import (
	"errors"
	"strings"
	"testing"

	entities "github.com/Testzyler/banking-api/app/.entities"
	models "github.com/Testzyler/banking-api/app/.models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserRepository is a mock implementation of UserRepository interface
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) GetByID(userID string) (*models.User, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetAll(perPage, page int, search string) ([]*models.User, error) {
	args := m.Called(perPage, page, search)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

func TestUserService_GetUserByID_MultipleScenarios(t *testing.T) {
	tests := []struct {
		name          string
		input         entities.GetUserByIdParams
		mockSetup     func(*MockUserRepository)
		expectError   bool
		expectUser    *models.User
		errorContains string
	}{
		{
			name: "valid user ID - success",
			input: entities.GetUserByIdParams{
				UserID: "user123",
			},
			mockSetup: func(mockRepo *MockUserRepository) {
				expectedUser := &models.User{
					UserID:   "user123",
					Name:     "John Doe",
					DummyCol: "test_data",
				}
				mockRepo.On("GetByID", "user123").Return(expectedUser, nil)
			},
			expectError: false,
			expectUser: &models.User{
				UserID:   "user123",
				Name:     "John Doe",
				DummyCol: "test_data",
			},
		},
		{
			name: "empty user ID",
			input: entities.GetUserByIdParams{
				UserID: "",
			},
			mockSetup: func(mockRepo *MockUserRepository) {
				mockRepo.On("GetByID", "").Return(nil, errors.New("user not found"))
			},
			expectError:   true,
			expectUser:    nil,
			errorContains: "user not found",
		},
		{
			name: "whitespace user ID",
			input: entities.GetUserByIdParams{
				UserID: "   ",
			},
			mockSetup: func(mockRepo *MockUserRepository) {
				mockRepo.On("GetByID", "   ").Return(nil, errors.New("user not found"))
			},
			expectError:   true,
			expectUser:    nil,
			errorContains: "user not found",
		},
		{
			name: "nonexistent user ID",
			input: entities.GetUserByIdParams{
				UserID: "nonexistent",
			},
			mockSetup: func(mockRepo *MockUserRepository) {
				mockRepo.On("GetByID", "nonexistent").Return(nil, errors.New("user not found"))
			},
			expectError:   true,
			expectUser:    nil,
			errorContains: "user not found",
		},
		{
			name: "repository database error",
			input: entities.GetUserByIdParams{
				UserID: "user123",
			},
			mockSetup: func(mockRepo *MockUserRepository) {
				mockRepo.On("GetByID", "user123").Return(nil, errors.New("database connection error"))
			},
			expectError:   true,
			expectUser:    nil,
			errorContains: "database connection error",
		},
		{
			name: "special characters in user ID",
			input: entities.GetUserByIdParams{
				UserID: "user@#$%",
			},
			mockSetup: func(mockRepo *MockUserRepository) {
				mockRepo.On("GetByID", "user@#$%").Return(nil, errors.New("user not found"))
			},
			expectError:   true,
			expectUser:    nil,
			errorContains: "user not found",
		},
		{
			name: "very long user ID",
			input: entities.GetUserByIdParams{
				UserID: strings.Repeat("a", 1000),
			},
			mockSetup: func(mockRepo *MockUserRepository) {
				longID := strings.Repeat("a", 1000)
				mockRepo.On("GetByID", longID).Return(nil, errors.New("user not found"))
			},
			expectError:   true,
			expectUser:    nil,
			errorContains: "user not found",
		},
		{
			name: "unicode user ID",
			input: entities.GetUserByIdParams{
				UserID: "ผู้ใช้123",
			},
			mockSetup: func(mockRepo *MockUserRepository) {
				expectedUser := &models.User{
					UserID:   "ผู้ใช้123",
					Name:     "Thai User",
					DummyCol: "test_data",
				}
				mockRepo.On("GetByID", "ผู้ใช้123").Return(expectedUser, nil)
			},
			expectError: false,
			expectUser: &models.User{
				UserID:   "ผู้ใช้123",
				Name:     "Thai User",
				DummyCol: "test_data",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockRepo := new(MockUserRepository)
			service := NewUserService(mockRepo)

			// Setup mock expectations
			tt.mockSetup(mockRepo)

			// Act
			user, err := service.GetUserByID(tt.input)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, user)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				if tt.expectUser != nil {
					assert.Equal(t, tt.expectUser.UserID, user.UserID)
					assert.Equal(t, tt.expectUser.Name, user.Name)
					assert.Equal(t, tt.expectUser.DummyCol, user.DummyCol)
				}
			}

			// Verify mock expectations
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_GetAllUsers_MultipleScenarios(t *testing.T) {
	tests := []struct {
		name          string
		input         entities.PaginationParams
		mockSetup     func(*MockUserRepository)
		expectError   bool
		expectUsers   []*models.User
		expectCount   int
		errorContains string
	}{
		{
			name: "valid pagination - first page",
			input: entities.PaginationParams{
				Page:    1,
				PerPage: 10,
				Search:  "",
			},
			mockSetup: func(mockRepo *MockUserRepository) {
				expectedUsers := []*models.User{
					{UserID: "user1", Name: "Alice", DummyCol: "data1"},
					{UserID: "user2", Name: "Bob", DummyCol: "data2"},
				}
				mockRepo.On("GetAll", 10, 1, "").Return(expectedUsers, nil)
			},
			expectError: false,
			expectCount: 2,
		},
		{
			name: "valid pagination with search",
			input: entities.PaginationParams{
				Page:    1,
				PerPage: 5,
				Search:  "alice",
			},
			mockSetup: func(mockRepo *MockUserRepository) {
				expectedUsers := []*models.User{
					{UserID: "user1", Name: "Alice", DummyCol: "data1"},
				}
				mockRepo.On("GetAll", 5, 1, "alice").Return(expectedUsers, nil)
			},
			expectError: false,
			expectCount: 1,
		},
		{
			name: "empty results",
			input: entities.PaginationParams{
				Page:    1,
				PerPage: 10,
				Search:  "nonexistent",
			},
			mockSetup: func(mockRepo *MockUserRepository) {
				mockRepo.On("GetAll", 10, 1, "nonexistent").Return([]*models.User{}, nil)
			},
			expectError: false,
			expectCount: 0,
		},
		{
			name: "second page pagination",
			input: entities.PaginationParams{
				Page:    2,
				PerPage: 5,
				Search:  "",
			},
			mockSetup: func(mockRepo *MockUserRepository) {
				expectedUsers := []*models.User{
					{UserID: "user6", Name: "Frank", DummyCol: "data6"},
					{UserID: "user7", Name: "Grace", DummyCol: "data7"},
				}
				mockRepo.On("GetAll", 5, 2, "").Return(expectedUsers, nil)
			},
			expectError: false,
			expectCount: 2,
		},
		{
			name: "repository error",
			input: entities.PaginationParams{
				Page:    1,
				PerPage: 10,
				Search:  "",
			},
			mockSetup: func(mockRepo *MockUserRepository) {
				mockRepo.On("GetAll", 10, 1, "").Return(nil, errors.New("database error"))
			},
			expectError:   true,
			errorContains: "database error",
		},
		{
			name: "special characters in search",
			input: entities.PaginationParams{
				Page:    1,
				PerPage: 10,
				Search:  "@#$%",
			},
			mockSetup: func(mockRepo *MockUserRepository) {
				mockRepo.On("GetAll", 10, 1, "@#$%").Return([]*models.User{}, nil)
			},
			expectError: false,
			expectCount: 0,
		},
		{
			name: "unicode search term",
			input: entities.PaginationParams{
				Page:    1,
				PerPage: 10,
				Search:  "ผู้ใช้",
			},
			mockSetup: func(mockRepo *MockUserRepository) {
				expectedUsers := []*models.User{
					{UserID: "thai_user", Name: "Thai User", DummyCol: "thai_data"},
				}
				mockRepo.On("GetAll", 10, 1, "ผู้ใช้").Return(expectedUsers, nil)
			},
			expectError: false,
			expectCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockRepo := new(MockUserRepository)
			service := NewUserService(mockRepo)

			// Setup mock expectations
			tt.mockSetup(mockRepo)

			// Act
			users, err := service.GetAllUsers(tt.input)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, users)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, users)
				assert.Len(t, users, tt.expectCount)
			}

			// Verify mock expectations
			mockRepo.AssertExpectations(t)
		})
	}
}

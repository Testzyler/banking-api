package service

import (
	"errors"
	"strings"
	"testing"

	entities "github.com/Testzyler/banking-api/app/entities"
	models "github.com/Testzyler/banking-api/app/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
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

func (m *MockUserRepository) GetAllWithCount(perPage, page int, search string) ([]*models.User, int64, error) {
	args := m.Called(perPage, page, search)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*models.User), args.Get(1).(int64), args.Error(2)
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
					UserID: "user123",
					Name:   "John Doe",
				}
				mockRepo.On("GetByID", "user123").Return(expectedUser, nil)
			},
			expectError: false,
			expectUser: &models.User{
				UserID: "user123",
				Name:   "John Doe",
			},
		},
		{
			name: "empty user ID",
			input: entities.GetUserByIdParams{
				UserID: "",
			},
			mockSetup: func(mockRepo *MockUserRepository) {
				mockRepo.On("GetByID", "").Return(nil, gorm.ErrRecordNotFound)
			},
			expectError:   true,
			expectUser:    nil,
			errorContains: "User with ID '' does not exist",
		},
		{
			name: "whitespace user ID",
			input: entities.GetUserByIdParams{
				UserID: "   ",
			},
			mockSetup: func(mockRepo *MockUserRepository) {
				mockRepo.On("GetByID", "   ").Return(nil, gorm.ErrRecordNotFound)
			},
			expectError:   true,
			expectUser:    nil,
			errorContains: "User with ID '   ' does not exist",
		},
		{
			name: "nonexistent user ID",
			input: entities.GetUserByIdParams{
				UserID: "nonexistent",
			},
			mockSetup: func(mockRepo *MockUserRepository) {
				mockRepo.On("GetByID", "nonexistent").Return(nil, gorm.ErrRecordNotFound)
			},
			expectError:   true,
			expectUser:    nil,
			errorContains: "User with ID 'nonexistent' does not exist",
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
				mockRepo.On("GetByID", "user@#$%").Return(nil, gorm.ErrRecordNotFound)
			},
			expectError:   true,
			expectUser:    nil,
			errorContains: "User with ID 'user@#$%' does not exist",
		},
		{
			name: "very long user ID",
			input: entities.GetUserByIdParams{
				UserID: strings.Repeat("a", 1000),
			},
			mockSetup: func(mockRepo *MockUserRepository) {
				longID := strings.Repeat("a", 1000)
				mockRepo.On("GetByID", longID).Return(nil, gorm.ErrRecordNotFound)
			},
			expectError:   true,
			expectUser:    nil,
			errorContains: "does not exist",
		},
		{
			name: "unicode user ID",
			input: entities.GetUserByIdParams{
				UserID: "ผู้ใช้123",
			},
			mockSetup: func(mockRepo *MockUserRepository) {
				expectedUser := &models.User{
					UserID: "ผู้ใช้123",
					Name:   "Thai User",
				}
				mockRepo.On("GetByID", "ผู้ใช้123").Return(expectedUser, nil)
			},
			expectError: false,
			expectUser: &models.User{
				UserID: "ผู้ใช้123",
				Name:   "Thai User",
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
					{UserID: "user1", Name: "Alice"},
					{UserID: "user2", Name: "Bob"},
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
					{UserID: "user1", Name: "Alice"},
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
					{UserID: "user6", Name: "Frank"},
					{UserID: "user7", Name: "Grace"},
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
					{UserID: "thai_user", Name: "Thai User"},
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

func TestUserService_GetAllUsersWithMeta_MultipleScenarios(t *testing.T) {
	tests := []struct {
		name          string
		input         entities.PaginationParams
		mockSetup     func(*MockUserRepository)
		expectError   bool
		expectUsers   []*models.User
		expectCount   int
		expectMeta    entities.PaginationMeta
		errorContains string
	}{
		{
			name: "valid pagination - first page with total count",
			input: entities.PaginationParams{
				Page:    1,
				PerPage: 10,
				Search:  "",
			},
			mockSetup: func(mockRepo *MockUserRepository) {
				expectedUsers := []*models.User{
					{UserID: "user1", Name: "Alice"},
					{UserID: "user2", Name: "Bob"},
				}
				mockRepo.On("GetAllWithCount", 10, 1, "").Return(expectedUsers, int64(25), nil)
			},
			expectError: false,
			expectCount: 2,
			expectMeta: entities.PaginationMeta{
				Page:        1,
				PerPage:     10,
				Total:       25,
				TotalPages:  3,
				HasNext:     true,
				HasPrevious: false,
			},
		},
		{
			name: "last page pagination",
			input: entities.PaginationParams{
				Page:    3,
				PerPage: 10,
				Search:  "",
			},
			mockSetup: func(mockRepo *MockUserRepository) {
				expectedUsers := []*models.User{
					{UserID: "user21", Name: "User21"},
				}
				mockRepo.On("GetAllWithCount", 10, 3, "").Return(expectedUsers, int64(21), nil)
			},
			expectError: false,
			expectCount: 1,
			expectMeta: entities.PaginationMeta{
				Page:        3,
				PerPage:     10,
				Total:       21,
				TotalPages:  3,
				HasNext:     false,
				HasPrevious: true,
			},
		},
		{
			name: "middle page pagination",
			input: entities.PaginationParams{
				Page:    2,
				PerPage: 5,
				Search:  "test",
			},
			mockSetup: func(mockRepo *MockUserRepository) {
				expectedUsers := []*models.User{
					{UserID: "user6", Name: "TestUser6"},
					{UserID: "user7", Name: "TestUser7"},
				}
				mockRepo.On("GetAllWithCount", 5, 2, "test").Return(expectedUsers, int64(12), nil)
			},
			expectError: false,
			expectCount: 2,
			expectMeta: entities.PaginationMeta{
				Page:        2,
				PerPage:     5,
				Total:       12,
				TotalPages:  3,
				HasNext:     true,
				HasPrevious: true,
			},
		},
		{
			name: "empty results with search",
			input: entities.PaginationParams{
				Page:    1,
				PerPage: 10,
				Search:  "nonexistent",
			},
			mockSetup: func(mockRepo *MockUserRepository) {
				mockRepo.On("GetAllWithCount", 10, 1, "nonexistent").Return([]*models.User{}, int64(0), nil)
			},
			expectError: false,
			expectCount: 0,
			expectMeta: entities.PaginationMeta{
				Page:        1,
				PerPage:     10,
				Total:       0,
				TotalPages:  0,
				HasNext:     false,
				HasPrevious: false,
			},
		},
		{
			name: "single page result",
			input: entities.PaginationParams{
				Page:    1,
				PerPage: 20,
				Search:  "",
			},
			mockSetup: func(mockRepo *MockUserRepository) {
				expectedUsers := []*models.User{
					{UserID: "user1", Name: "User1"},
					{UserID: "user2", Name: "User2"},
				}
				mockRepo.On("GetAllWithCount", 20, 1, "").Return(expectedUsers, int64(5), nil)
			},
			expectError: false,
			expectCount: 2,
			expectMeta: entities.PaginationMeta{
				Page:        1,
				PerPage:     20,
				Total:       5,
				TotalPages:  1,
				HasNext:     false,
				HasPrevious: false,
			},
		},
		{
			name: "repository error",
			input: entities.PaginationParams{
				Page:    1,
				PerPage: 10,
				Search:  "",
			},
			mockSetup: func(mockRepo *MockUserRepository) {
				mockRepo.On("GetAllWithCount", 10, 1, "").Return(nil, int64(0), errors.New("database connection error"))
			},
			expectError:   true,
			errorContains: "database connection error",
		},
		{
			name: "unicode search with results",
			input: entities.PaginationParams{
				Page:    1,
				PerPage: 10,
				Search:  "ผู้ใช้",
			},
			mockSetup: func(mockRepo *MockUserRepository) {
				expectedUsers := []*models.User{
					{UserID: "thai_user", Name: "Thai User"},
				}
				mockRepo.On("GetAllWithCount", 10, 1, "ผู้ใช้").Return(expectedUsers, int64(1), nil)
			},
			expectError: false,
			expectCount: 1,
			expectMeta: entities.PaginationMeta{
				Page:        1,
				PerPage:     10,
				Total:       1,
				TotalPages:  1,
				HasNext:     false,
				HasPrevious: false,
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
			users, meta, err := service.GetAllUsersWithMeta(tt.input)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, users)
				assert.Equal(t, entities.PaginationMeta{}, meta)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, users)
				assert.Len(t, users, tt.expectCount)

				// Verify pagination metadata
				assert.Equal(t, tt.expectMeta.Page, meta.Page)
				assert.Equal(t, tt.expectMeta.PerPage, meta.PerPage)
				assert.Equal(t, tt.expectMeta.Total, meta.Total)
				assert.Equal(t, tt.expectMeta.TotalPages, meta.TotalPages)
				assert.Equal(t, tt.expectMeta.HasNext, meta.HasNext)
				assert.Equal(t, tt.expectMeta.HasPrevious, meta.HasPrevious)
			}

			// Verify mock expectations
			mockRepo.AssertExpectations(t)
		})
	}
}

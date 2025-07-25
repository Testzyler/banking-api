package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"testing"

	entities "github.com/Testzyler/banking-api/app/.entities"
	models "github.com/Testzyler/banking-api/app/.models"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserService is a mock implementation of UserService interface
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) GetUserByID(params entities.GetUserByIdParams) (*models.User, error) {
	args := m.Called(params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) GetAllUsers(params entities.PaginationParams) ([]*models.User, error) {
	args := m.Called(params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

func setupTestApp() *fiber.App {
	app := fiber.New()
	return app
}

func TestUserHandler_GetUser_TableDriven(t *testing.T) {
	tests := []struct {
		name               string
		userID             string
		mockSetup          func(*MockUserService, string)
		expectStatusCode   int
		expectErrorMessage string
		expectUserData     bool
		expectUserID       string
	}{
		{
			name:   "valid user ID - success",
			userID: "user123",
			mockSetup: func(mockService *MockUserService, userID string) {
				expectedUser := &models.User{
					UserID:   userID,
					Name:     "John Doe",
					DummyCol: "test_data",
				}
				params := entities.GetUserByIdParams{UserID: userID}
				mockService.On("GetUserByID", params).Return(expectedUser, nil)
			},
			expectStatusCode: fiber.StatusOK,
			expectUserData:   true,
			expectUserID:     "user123",
		},
		{
			name:   "empty user ID - routes to list",
			userID: "",
			mockSetup: func(mockService *MockUserService, userID string) {
				// Empty user ID routes to GetUsers (list endpoint)
				params := entities.PaginationParams{Page: 1, PerPage: 10, Search: ""}
				mockService.On("GetAllUsers", params).Return([]*models.User{}, nil)
			},
			expectStatusCode: fiber.StatusOK,
			expectUserData:   true,
			expectUserID:     "", // Empty means expect array response
		},
		{
			name:   "nonexistent user ID",
			userID: "nonexistent",
			mockSetup: func(mockService *MockUserService, userID string) {
				params := entities.GetUserByIdParams{UserID: userID}
				mockService.On("GetUserByID", params).Return(nil, errors.New("user not found"))
			},
			expectStatusCode:   fiber.StatusNotFound,
			expectErrorMessage: "User not found",
			expectUserData:     false,
		},
		{
			name:   "service error",
			userID: "user123",
			mockSetup: func(mockService *MockUserService, userID string) {
				params := entities.GetUserByIdParams{UserID: userID}
				mockService.On("GetUserByID", params).Return(nil, errors.New("database error"))
			},
			expectStatusCode:   fiber.StatusNotFound,
			expectErrorMessage: "User not found",
			expectUserData:     false,
		},
		{
			name:   "special characters in user ID",
			userID: "user@#$%",
			mockSetup: func(mockService *MockUserService, userID string) {
				// The URL encoded version is what actually gets passed to the handler
				encodedUserID := url.PathEscape(userID)
				params := entities.GetUserByIdParams{UserID: encodedUserID}
				mockService.On("GetUserByID", params).Return(nil, errors.New("user not found"))
			},
			expectStatusCode:   fiber.StatusNotFound,
			expectErrorMessage: "User not found",
			expectUserData:     false,
		},
		{
			name:   "unicode user ID - success",
			userID: "ผู้ใช้123",
			mockSetup: func(mockService *MockUserService, userID string) {
				// The URL encoded version is what actually gets passed to the handler
				encodedUserID := url.PathEscape(userID)
				expectedUser := &models.User{
					UserID:   encodedUserID,
					Name:     "Thai User",
					DummyCol: "test_data",
				}
				params := entities.GetUserByIdParams{UserID: encodedUserID}
				mockService.On("GetUserByID", params).Return(expectedUser, nil)
			},
			expectStatusCode: fiber.StatusOK,
			expectUserData:   true,
			expectUserID:     url.PathEscape("ผู้ใช้123"), // Expect the encoded version
		},
		{
			name:   "whitespace user ID",
			userID: "   ",
			mockSetup: func(mockService *MockUserService, userID string) {
				// The URL encoded version is what actually gets passed to the handler
				encodedUserID := url.PathEscape(userID)
				params := entities.GetUserByIdParams{UserID: encodedUserID}
				mockService.On("GetUserByID", params).Return(nil, errors.New("user not found"))
			},
			expectStatusCode:   fiber.StatusNotFound,
			expectErrorMessage: "User not found",
			expectUserData:     false,
		},
		{
			name:   "very long user ID",
			userID: "user" + string(make([]byte, 100)), // Reduced for efficiency
			mockSetup: func(mockService *MockUserService, userID string) {
				encodedUserID := url.PathEscape(userID)
				params := entities.GetUserByIdParams{UserID: encodedUserID}
				mockService.On("GetUserByID", params).Return(nil, errors.New("user not found"))
			},
			expectStatusCode:   fiber.StatusNotFound,
			expectErrorMessage: "User not found",
			expectUserData:     false,
		},
		{
			name:   "SQL injection attempt in user ID",
			userID: "'; DROP TABLE users; --",
			mockSetup: func(mockService *MockUserService, userID string) {
				encodedUserID := url.PathEscape(userID)
				params := entities.GetUserByIdParams{UserID: encodedUserID}
				mockService.On("GetUserByID", params).Return(nil, errors.New("user not found"))
			},
			expectStatusCode:   fiber.StatusNotFound,
			expectErrorMessage: "User not found",
			expectUserData:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			app := setupTestApp()
			mockService := new(MockUserService)

			// Setup routes
			api := app.Group("/api/v1")
			NewUserHandler(api, mockService)

			// Setup mock expectations
			tt.mockSetup(mockService, tt.userID)

			// Act
			var req *http.Request
			if tt.userID == "" {
				req, _ = http.NewRequest("GET", "/api/v1/users", nil)
			} else {
				req, _ = http.NewRequest("GET", "/api/v1/users/"+url.PathEscape(tt.userID), nil)
			}
			resp, err := app.Test(req)

			// Assert
			assert.NoError(t, err)
			assert.Equal(t, tt.expectStatusCode, resp.StatusCode)

			body, err := io.ReadAll(resp.Body)
			assert.NoError(t, err)

			if tt.expectUserData {
				if tt.expectUserID != "" {
					// Expect single user object
					var user models.User
					err = json.Unmarshal(body, &user)
					assert.NoError(t, err)
					assert.Equal(t, tt.expectUserID, user.UserID)
					assert.NotEmpty(t, user.Name)
				} else {
					// Expect array of users (empty user ID routes to list)
					var users []*models.User
					err = json.Unmarshal(body, &users)
					assert.NoError(t, err)
				}
			} else if tt.expectErrorMessage != "" {
				// Expect error message in response
				var response map[string]string
				err = json.Unmarshal(body, &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectErrorMessage, response["error"])
			}

			// Verify mock expectations
			mockService.AssertExpectations(t)
		})
	}
}

func TestUserHandler_GetUsers_TableDriven(t *testing.T) {
	tests := []struct {
		name               string
		queryParams        string
		mockSetup          func(*MockUserService)
		expectStatusCode   int
		expectErrorMessage string
		expectUserData     bool
		expectCount        int
	}{
		{
			name:        "valid request - default pagination",
			queryParams: "",
			mockSetup: func(mockService *MockUserService) {
				expectedUsers := []*models.User{
					{UserID: "user1", Name: "Alice", DummyCol: "data1"},
					{UserID: "user2", Name: "Bob", DummyCol: "data2"},
				}
				params := entities.PaginationParams{Page: 1, PerPage: 10, Search: ""}
				mockService.On("GetAllUsers", params).Return(expectedUsers, nil)
			},
			expectStatusCode: fiber.StatusOK,
			expectUserData:   true,
			expectCount:      2,
		},
		{
			name:        "valid request with pagination",
			queryParams: "?page=2&per_page=5",
			mockSetup: func(mockService *MockUserService) {
				expectedUsers := []*models.User{
					{UserID: "user6", Name: "Frank", DummyCol: "data6"},
				}
				params := entities.PaginationParams{Page: 2, PerPage: 5, Search: ""}
				mockService.On("GetAllUsers", params).Return(expectedUsers, nil)
			},
			expectStatusCode: fiber.StatusOK,
			expectUserData:   true,
			expectCount:      1,
		},
		{
			name:        "valid request with search",
			queryParams: "?search=alice",
			mockSetup: func(mockService *MockUserService) {
				expectedUsers := []*models.User{
					{UserID: "user1", Name: "Alice", DummyCol: "data1"},
				}
				params := entities.PaginationParams{Page: 1, PerPage: 10, Search: "alice"}
				mockService.On("GetAllUsers", params).Return(expectedUsers, nil)
			},
			expectStatusCode: fiber.StatusOK,
			expectUserData:   true,
			expectCount:      1,
		},
		{
			name:        "empty results",
			queryParams: "?search=nonexistent",
			mockSetup: func(mockService *MockUserService) {
				params := entities.PaginationParams{Page: 1, PerPage: 10, Search: "nonexistent"}
				mockService.On("GetAllUsers", params).Return([]*models.User{}, nil)
			},
			expectStatusCode: fiber.StatusOK,
			expectUserData:   true,
			expectCount:      0,
		},
		{
			name:               "invalid per_page parameter",
			queryParams:        "?per_page=invalid",
			mockSetup:          func(mockService *MockUserService) {},
			expectStatusCode:   fiber.StatusBadRequest,
			expectErrorMessage: "Invalid per_page parameter",
			expectUserData:     false,
		},
		{
			name:               "invalid page parameter",
			queryParams:        "?page=0",
			mockSetup:          func(mockService *MockUserService) {},
			expectStatusCode:   fiber.StatusBadRequest,
			expectErrorMessage: "Invalid page parameter",
			expectUserData:     false,
		},
		{
			name:               "negative per_page parameter",
			queryParams:        "?per_page=-5",
			mockSetup:          func(mockService *MockUserService) {},
			expectStatusCode:   fiber.StatusBadRequest,
			expectErrorMessage: "Invalid per_page parameter",
			expectUserData:     false,
		},
		{
			name:        "service error",
			queryParams: "",
			mockSetup: func(mockService *MockUserService) {
				params := entities.PaginationParams{Page: 1, PerPage: 10, Search: ""}
				mockService.On("GetAllUsers", params).Return(nil, errors.New("database error"))
			},
			expectStatusCode:   fiber.StatusInternalServerError,
			expectErrorMessage: "Failed to fetch users",
			expectUserData:     false,
		},
		{
			name:        "special characters in search",
			queryParams: "?search=" + url.QueryEscape("@#$%^&*()"),
			mockSetup: func(mockService *MockUserService) {
				params := entities.PaginationParams{Page: 1, PerPage: 10, Search: "@#$%^&*()"}
				mockService.On("GetAllUsers", params).Return([]*models.User{}, nil)
			},
			expectStatusCode: fiber.StatusOK,
			expectUserData:   true,
			expectCount:      0,
		},
		{
			name:        "unicode search term",
			queryParams: "?search=" + url.QueryEscape("ผู้ใช้"),
			mockSetup: func(mockService *MockUserService) {
				expectedUsers := []*models.User{
					{UserID: "thai_user", Name: "Thai User", DummyCol: "thai_data"},
				}
				params := entities.PaginationParams{Page: 1, PerPage: 10, Search: "ผู้ใช้"}
				mockService.On("GetAllUsers", params).Return(expectedUsers, nil)
			},
			expectStatusCode: fiber.StatusOK,
			expectUserData:   true,
			expectCount:      1,
		},
		{
			name:        "large per_page value",
			queryParams: "?per_page=1000",
			mockSetup: func(mockService *MockUserService) {
				// Simulate large dataset
				users := make([]*models.User, 1000)
				for i := 0; i < 1000; i++ {
					users[i] = &models.User{
						UserID:   fmt.Sprintf("user%d", i),
						Name:     fmt.Sprintf("User %d", i),
						DummyCol: fmt.Sprintf("data%d", i),
					}
				}
				params := entities.PaginationParams{Page: 1, PerPage: 1000, Search: ""}
				mockService.On("GetAllUsers", params).Return(users, nil)
			},
			expectStatusCode: fiber.StatusOK,
			expectUserData:   true,
			expectCount:      1000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			app := setupTestApp()
			mockService := new(MockUserService)

			// Setup routes
			api := app.Group("/api/v1")
			NewUserHandler(api, mockService)

			// Setup mock expectations
			tt.mockSetup(mockService)

			// Act
			req, _ := http.NewRequest("GET", "/api/v1/users"+tt.queryParams, nil)
			resp, err := app.Test(req)

			// Assert
			assert.NoError(t, err)
			assert.Equal(t, tt.expectStatusCode, resp.StatusCode)

			body, err := io.ReadAll(resp.Body)
			assert.NoError(t, err)

			if tt.expectUserData {
				// Expect user data in response
				var users []*models.User
				err = json.Unmarshal(body, &users)
				assert.NoError(t, err)
				assert.Len(t, users, tt.expectCount)
			} else if tt.expectErrorMessage != "" {
				// Expect error message in response
				var response map[string]string
				err = json.Unmarshal(body, &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectErrorMessage, response["error"])
			}

			// Verify mock expectations
			mockService.AssertExpectations(t)
		})
	}
}

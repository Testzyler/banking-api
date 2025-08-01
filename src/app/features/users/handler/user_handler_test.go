package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"testing"

	entities "github.com/Testzyler/banking-api/app/entities"
	models "github.com/Testzyler/banking-api/app/models"
	"github.com/Testzyler/banking-api/logger"
	"github.com/Testzyler/banking-api/server/middlewares"
	"github.com/Testzyler/banking-api/server/response"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
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

func (m *MockUserService) GetAllUsersWithMeta(params entities.PaginationParams) ([]*models.User, entities.PaginationMeta, error) {
	args := m.Called(params)
	if args.Get(0) == nil {
		return nil, entities.PaginationMeta{}, args.Error(2)
	}
	return args.Get(0).([]*models.User), args.Get(1).(entities.PaginationMeta), args.Error(2)
}

func setupTestApp() *fiber.App {
	// Initialize logger for tests to prevent nil pointer panics
	Logger := zap.NewNop().Sugar()
	logger.Logger = Logger
	app := fiber.New(fiber.Config{
		ErrorHandler: middlewares.ErrorHandler(),
	})
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
				meta := entities.PaginationMeta{
					Page:        1,
					PerPage:     10,
					Total:       0,
					TotalPages:  0,
					HasNext:     false,
					HasPrevious: false,
				}
				mockService.On("GetAllUsersWithMeta", params).Return([]*models.User{}, meta, nil)
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
				// Service returns the specific UserNotFoundError which results in 404
				mockService.On("GetUserByID", params).Return(nil, &response.ErrorResponse{
					HttpStatusCode: fiber.StatusNotFound,
					Code:           response.ErrCodeNotFound,
					Message:        "User not found",
					Details:        "User with ID 'nonexistent' does not exist",
				})
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
				// Service returns internal error which results in 500
				mockService.On("GetUserByID", params).Return(nil, &response.ErrorResponse{
					HttpStatusCode: fiber.StatusInternalServerError,
					Code:           response.ErrCodeInternalServer,
					Message:        "Internal server error",
					Details:        "An unexpected error occurred while processing your request",
				})
			},
			expectStatusCode:   fiber.StatusInternalServerError,
			expectErrorMessage: "Internal server error",
			expectUserData:     false,
		},
		{
			name:   "special characters in user ID",
			userID: "user@#$%",
			mockSetup: func(mockService *MockUserService, userID string) {
				// The URL encoded version is what actually gets passed to the handler
				encodedUserID := url.PathEscape(userID)
				params := entities.GetUserByIdParams{UserID: encodedUserID}
				mockService.On("GetUserByID", params).Return(nil, &response.ErrorResponse{
					HttpStatusCode: fiber.StatusNotFound,
					Code:           response.ErrCodeNotFound,
					Message:        "User not found",
					Details:        "User with ID '" + encodedUserID + "' does not exist",
				})
			},
			expectStatusCode:   fiber.StatusNotFound,
			expectErrorMessage: "User not found",
			expectUserData:     false,
		},
		{
			name:   "unicode user ID - validation error",
			userID: "ผู้ใช้123",
			mockSetup: func(mockService *MockUserService, userID string) {
				// URL-encoded Unicode string is 57 chars, exceeds validation limit of 50
				// This will fail validation before reaching the service, so no mock setup needed
			},
			expectStatusCode:   fiber.StatusUnprocessableEntity, // 422
			expectErrorMessage: "Validation failed",
			expectUserData:     false,
		},
		{
			name:   "whitespace user ID",
			userID: "   ",
			mockSetup: func(mockService *MockUserService, userID string) {
				// URL-encoded whitespace becomes %20%20%20 (9 chars), passes length validation
				// but should return not found from service
				encodedUserID := url.PathEscape(userID)
				params := entities.GetUserByIdParams{UserID: encodedUserID}
				mockService.On("GetUserByID", params).Return(nil, &response.ErrorResponse{
					HttpStatusCode: fiber.StatusNotFound,
					Code:           response.ErrCodeNotFound,
					Message:        "User not found",
					Details:        "User with ID '" + encodedUserID + "' does not exist",
				})
			},
			expectStatusCode:   fiber.StatusNotFound,
			expectErrorMessage: "User not found",
			expectUserData:     false,
		},
		{
			name:   "very long user ID",
			userID: "user" + string(make([]byte, 100)), // Over 50 characters
			mockSetup: func(mockService *MockUserService, userID string) {
				// This will fail validation before reaching the service, so no mock setup needed
			},
			expectStatusCode:   fiber.StatusUnprocessableEntity, // 422
			expectErrorMessage: "Validation failed",
			expectUserData:     false,
		},
		{
			name:   "SQL injection attempt in user ID",
			userID: "'; DROP TABLE users; --",
			mockSetup: func(mockService *MockUserService, userID string) {
				encodedUserID := url.PathEscape(userID)
				params := entities.GetUserByIdParams{UserID: encodedUserID}
				mockService.On("GetUserByID", params).Return(nil, &response.ErrorResponse{
					HttpStatusCode: fiber.StatusNotFound,
					Code:           response.ErrCodeNotFound,
					Message:        "User not found",
					Details:        "User with ID '" + encodedUserID + "' does not exist",
				})
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
					// Expect single user object (GetUser endpoint)
					var response struct {
						Message string      `json:"message"`
						Data    models.User `json:"data"`
					}
					err = json.Unmarshal(body, &response)
					assert.NoError(t, err)
					assert.Equal(t, tt.expectUserID, response.Data.UserID)
					assert.NotEmpty(t, response.Data.Name)
				} else {
					// Expect paginated response (GetUsers endpoint)
					var response entities.PaginatedResponse
					err = json.Unmarshal(body, &response)
					assert.NoError(t, err)

					// Parse the data as array of users
					dataBytes, err := json.Marshal(response.Data)
					assert.NoError(t, err)
					var users []*models.User
					err = json.Unmarshal(dataBytes, &users)
					assert.NoError(t, err)

					// Verify metadata structure exists
					assert.NotNil(t, response.Meta)
				}
			} else if tt.expectErrorMessage != "" {
				// Expect error message in response
				var errorResponse struct {
					Code    uint64      `json:"code"`
					Message string      `json:"message"`
					Details interface{} `json:"details,omitempty"`
				}
				err = json.Unmarshal(body, &errorResponse)
				assert.NoError(t, err)
				assert.Contains(t, errorResponse.Message, tt.expectErrorMessage)
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
					{UserID: "user1", Name: "Alice"},
					{UserID: "user2", Name: "Bob"},
				}
				params := entities.PaginationParams{Page: 1, PerPage: 10, Search: ""}
				meta := entities.PaginationMeta{
					Page:        1,
					PerPage:     10,
					Total:       2,
					TotalPages:  1,
					HasNext:     false,
					HasPrevious: false,
				}
				mockService.On("GetAllUsersWithMeta", params).Return(expectedUsers, meta, nil)
			},
			expectStatusCode: fiber.StatusOK,
			expectUserData:   true,
			expectCount:      2,
		},
		{
			name:        "valid request with pagination",
			queryParams: "?page=2&perPage=5",
			mockSetup: func(mockService *MockUserService) {
				expectedUsers := []*models.User{
					{UserID: "user6", Name: "Frank"},
				}
				params := entities.PaginationParams{Page: 2, PerPage: 5, Search: ""}
				meta := entities.PaginationMeta{
					Page:        2,
					PerPage:     5,
					Total:       6,
					TotalPages:  2,
					HasNext:     false,
					HasPrevious: true,
				}
				mockService.On("GetAllUsersWithMeta", params).Return(expectedUsers, meta, nil)
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
					{UserID: "user1", Name: "Alice"},
				}
				params := entities.PaginationParams{Page: 1, PerPage: 10, Search: "alice"}
				meta := entities.PaginationMeta{
					Page:        1,
					PerPage:     10,
					Total:       1,
					TotalPages:  1,
					HasNext:     false,
					HasPrevious: false,
				}
				mockService.On("GetAllUsersWithMeta", params).Return(expectedUsers, meta, nil)
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
				meta := entities.PaginationMeta{
					Page:        1,
					PerPage:     10,
					Total:       0,
					TotalPages:  0,
					HasNext:     false,
					HasPrevious: false,
				}
				mockService.On("GetAllUsersWithMeta", params).Return([]*models.User{}, meta, nil)
			},
			expectStatusCode: fiber.StatusOK,
			expectUserData:   true,
			expectCount:      0,
		},
		{
			name:               "invalid perPage parameter",
			queryParams:        "?perPage=invalid",
			mockSetup:          func(mockService *MockUserService) {},
			expectStatusCode:   fiber.StatusUnprocessableEntity,
			expectErrorMessage: "Validation failed",
			expectUserData:     false,
		},
		{
			name:               "invalid page parameter",
			queryParams:        "?page=0",
			mockSetup:          func(mockService *MockUserService) {},
			expectStatusCode:   fiber.StatusUnprocessableEntity,
			expectErrorMessage: "Validation failed",
			expectUserData:     false,
		},
		{
			name:               "negative perPage parameter",
			queryParams:        "?perPage=-5",
			mockSetup:          func(mockService *MockUserService) {},
			expectStatusCode:   fiber.StatusUnprocessableEntity,
			expectErrorMessage: "Validation failed",
			expectUserData:     false,
		},
		{
			name:        "service error",
			queryParams: "",
			mockSetup: func(mockService *MockUserService) {
				params := entities.PaginationParams{Page: 1, PerPage: 10, Search: ""}
				mockService.On("GetAllUsersWithMeta", params).Return(nil, entities.PaginationMeta{}, &response.ErrorResponse{
					HttpStatusCode: fiber.StatusInternalServerError,
					Code:           response.ErrCodeInternalServer,
					Message:        "Internal server error",
					Details:        "An unexpected error occurred while processing your request",
				})
			},
			expectStatusCode:   fiber.StatusInternalServerError,
			expectErrorMessage: "Internal server error",
			expectUserData:     false,
		},
		{
			name:        "special characters in search",
			queryParams: "?search=" + url.QueryEscape("@#$%^&*()"),
			mockSetup: func(mockService *MockUserService) {
				params := entities.PaginationParams{Page: 1, PerPage: 10, Search: "@#$%^&*()"}
				meta := entities.PaginationMeta{
					Page:        1,
					PerPage:     10,
					Total:       0,
					TotalPages:  0,
					HasNext:     false,
					HasPrevious: false,
				}
				mockService.On("GetAllUsersWithMeta", params).Return([]*models.User{}, meta, nil)
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
					{UserID: "thai_user", Name: "Thai User"},
				}
				params := entities.PaginationParams{Page: 1, PerPage: 10, Search: "ผู้ใช้"}
				meta := entities.PaginationMeta{
					Page:        1,
					PerPage:     10,
					Total:       1,
					TotalPages:  1,
					HasNext:     false,
					HasPrevious: false,
				}
				mockService.On("GetAllUsersWithMeta", params).Return(expectedUsers, meta, nil)
			},
			expectStatusCode: fiber.StatusOK,
			expectUserData:   true,
			expectCount:      1,
		},
		{
			name:               "large perPage value exceeds maximum",
			queryParams:        "?perPage=1000",
			mockSetup:          func(mockService *MockUserService) {},
			expectStatusCode:   fiber.StatusUnprocessableEntity,
			expectErrorMessage: "Validation failed",
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
				// Expect paginated response with user data
				var response entities.PaginatedResponse
				err = json.Unmarshal(body, &response)
				assert.NoError(t, err)

				// Parse the data as array of users
				dataBytes, err := json.Marshal(response.Data)
				assert.NoError(t, err)
				var users []*models.User
				err = json.Unmarshal(dataBytes, &users)
				assert.NoError(t, err)
				assert.Len(t, users, tt.expectCount)

				// Verify metadata structure
				assert.NotNil(t, response.Meta)
				assert.Equal(t, "Users retrieved successfully", response.Message)
			} else if tt.expectErrorMessage != "" {
				// Expect error message in response
				var errorResponse struct {
					Code    uint64      `json:"code"`
					Message string      `json:"message"`
					Details interface{} `json:"details,omitempty"`
				}
				err = json.Unmarshal(body, &errorResponse)
				assert.NoError(t, err)
				assert.Contains(t, errorResponse.Message, tt.expectErrorMessage)
			}

			// Verify mock expectations
			mockService.AssertExpectations(t)
		})
	}
}

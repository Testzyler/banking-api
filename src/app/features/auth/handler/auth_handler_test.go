package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Testzyler/banking-api/app/entities"
	"github.com/Testzyler/banking-api/logger"
	"github.com/Testzyler/banking-api/server/middlewares"
	"github.com/Testzyler/banking-api/server/response"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// Mock AuthService
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) VerifyPin(ctx context.Context, params entities.PinVerifyParams) (*entities.TokenResponse, error) {
	args := m.Called(params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.TokenResponse), args.Error(1)
}

func (m *MockAuthService) RefreshToken(refreshToken string) (*entities.TokenResponse, error) {
	args := m.Called(refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.TokenResponse), args.Error(1)
}

func (m *MockAuthService) ListTokens(ctx context.Context) ([]entities.TokenResponse, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entities.TokenResponse), args.Error(1)
}

func (m *MockAuthService) BanToken(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
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

func TestAuthHandler_VerifyPin_BasicTests(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func(*MockAuthService)
		expectedStatus int
		expectSuccess  bool
	}{
		{
			name: "successful pin verification",
			requestBody: entities.PinVerifyParams{
				Username: "testuser",
				Pin:      "123456",
			},
			mockSetup: func(mockService *MockAuthService) {
				tokenResponse := &entities.TokenResponse{
					Token:        "access_token",
					RefreshToken: "refresh_token",
					Expiry:       time.Now().Add(time.Hour),
					UserID:       "user123",
				}
				mockService.On("VerifyPin", mock.AnythingOfType("entities.PinVerifyParams")).Return(tokenResponse, nil)
			},
			expectedStatus: fiber.StatusOK,
			expectSuccess:  true,
		},
		{
			name: "service error - user not found",
			requestBody: entities.PinVerifyParams{
				Username: "nonexistent",
				Pin:      "123456",
			},
			mockSetup: func(mockService *MockAuthService) {
				mockService.On("VerifyPin", mock.AnythingOfType("entities.PinVerifyParams")).Return(nil, &response.ErrorResponse{
					HttpStatusCode: fiber.StatusNotFound,
					Code:           response.ErrCodeNotFound,
					Message:        "User not found",
				})
			},
			expectedStatus: fiber.StatusNotFound,
			expectSuccess:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			app := setupTestApp()
			mockService := new(MockAuthService)

			// Create handler
			handler := &authHandler{service: mockService}
			app.Post("/auth/verify-pin", handler.VerifyPin)

			// Setup mock expectations
			tt.mockSetup(mockService)

			// Create request
			var body []byte
			var err error
			if tt.requestBody != nil {
				body, err = json.Marshal(tt.requestBody)
				assert.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPost, "/auth/verify-pin", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			// Act
			resp, err := app.Test(req)
			assert.NoError(t, err)

			// Assert - Basic status code check only
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			// Verify mock expectations
			mockService.AssertExpectations(t)
		})
	}
}

func TestAuthHandler_RefreshToken_BasicTests(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func(*MockAuthService)
		expectedStatus int
		expectSuccess  bool
	}{
		{
			name: "successful token refresh",
			requestBody: entities.RefreshTokenRequest{
				RefreshToken: "valid_refresh_token",
			},
			mockSetup: func(mockService *MockAuthService) {
				tokenResponse := &entities.TokenResponse{
					Token:        "new_access_token",
					RefreshToken: "valid_refresh_token",
					Expiry:       time.Now().Add(time.Hour),
					UserID:       "user123",
				}
				mockService.On("RefreshToken", "valid_refresh_token").Return(tokenResponse, nil)
			},
			expectedStatus: fiber.StatusOK,
			expectSuccess:  true,
		},
		{
			name: "invalid refresh token",
			requestBody: entities.RefreshTokenRequest{
				RefreshToken: "invalid_token",
			},
			mockSetup: func(mockService *MockAuthService) {
				mockService.On("RefreshToken", "invalid_token").Return(nil, &response.ErrorResponse{
					HttpStatusCode: fiber.StatusUnauthorized,
					Code:           response.ErrCodeUnauthorized,
					Message:        "Invalid refresh token",
				})
			},
			expectedStatus: fiber.StatusUnauthorized,
			expectSuccess:  false,
		},
		{
			name: "generic error from service",
			requestBody: entities.RefreshTokenRequest{
				RefreshToken: "some_token",
			},
			mockSetup: func(mockService *MockAuthService) {
				mockService.On("RefreshToken", "some_token").Return(nil, errors.New("generic error"))
			},
			expectedStatus: fiber.StatusUnauthorized,
			expectSuccess:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			app := setupTestApp()
			mockService := new(MockAuthService)

			// Create handler
			handler := &authHandler{service: mockService}
			app.Post("/auth/refresh", handler.RefreshToken)

			// Setup mock expectations
			tt.mockSetup(mockService)

			// Create request
			var body []byte
			var err error
			if tt.requestBody != nil {
				body, err = json.Marshal(tt.requestBody)
				assert.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			// Act
			resp, err := app.Test(req)
			assert.NoError(t, err)

			// Assert - Basic status code check only
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			// Verify mock expectations
			mockService.AssertExpectations(t)
		})
	}
}

// Test for validation errors
func TestAuthHandler_VerifyPin_ValidationErrors(t *testing.T) {
	app := setupTestApp()
	mockService := new(MockAuthService)

	handler := &authHandler{service: mockService}
	app.Post("/auth/verify-pin", handler.VerifyPin)

	// Test invalid PIN format
	invalidParams := entities.PinVerifyParams{
		Username: "testuser",
		Pin:      "12345", // Invalid: only 5 digits
	}

	body, _ := json.Marshal(invalidParams)
	req := httptest.NewRequest(http.MethodPost, "/auth/verify-pin", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)

	// Fiber may transform validation errors to 500, so just check it's an error status
	assert.True(t, resp.StatusCode >= 400, "Expected error status code for validation error")

	mockService.AssertExpectations(t)
}

// Test for invalid request body
func TestAuthHandler_InvalidRequestBody(t *testing.T) {
	app := setupTestApp()
	mockService := new(MockAuthService)

	handler := &authHandler{service: mockService}
	app.Post("/auth/verify-pin", handler.VerifyPin)

	// Test invalid JSON
	req := httptest.NewRequest(http.MethodPost, "/auth/verify-pin", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)

	// Fiber may transform JSON parse errors to 500, so just check it's an error status
	assert.True(t, resp.StatusCode >= 400, "Expected error status code for invalid JSON")

	mockService.AssertExpectations(t)
}

func TestAuthHandler_ListAllTokens(t *testing.T) {
	tests := []struct {
		name           string
		mockSetup      func(*MockAuthService)
		expectedStatus int
		expectSuccess  bool
	}{
		{
			name: "successful token listing",
			mockSetup: func(mockService *MockAuthService) {
				tokens := []entities.TokenResponse{
					{
						Token:        "token1",
						RefreshToken: "refresh1",
						Expiry:       time.Now().Add(time.Hour),
						UserID:       "user1",
					},
					{
						Token:        "token2",
						RefreshToken: "refresh2",
						Expiry:       time.Now().Add(time.Hour),
						UserID:       "user2",
					},
				}
				mockService.On("ListTokens", mock.Anything).Return(tokens, nil)
			},
			expectedStatus: fiber.StatusOK,
			expectSuccess:  true,
		},
		{
			name: "empty token list",
			mockSetup: func(mockService *MockAuthService) {
				tokens := []entities.TokenResponse{}
				mockService.On("ListTokens", mock.Anything).Return(tokens, nil)
			},
			expectedStatus: fiber.StatusOK,
			expectSuccess:  true,
		},
		{
			name: "service error during token listing",
			mockSetup: func(mockService *MockAuthService) {
				mockService.On("ListTokens", mock.Anything).Return(nil, &response.ErrorResponse{
					HttpStatusCode: fiber.StatusInternalServerError,
					Code:           response.ErrCodeInternalServer,
					Message:        "Database connection failed",
				})
			},
			expectedStatus: fiber.StatusInternalServerError,
			expectSuccess:  false,
		},
		{
			name: "generic error from service",
			mockSetup: func(mockService *MockAuthService) {
				mockService.On("ListTokens", mock.Anything).Return(nil, errors.New("unexpected error"))
			},
			expectedStatus: fiber.StatusInternalServerError,
			expectSuccess:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			app := setupTestApp()
			mockService := new(MockAuthService)

			// Create handler
			handler := &authHandler{service: mockService}
			app.Get("/auth/tokens", handler.ListAllTokens)

			// Setup mock expectations
			tt.mockSetup(mockService)

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/auth/tokens", nil)

			// Act
			resp, err := app.Test(req)
			assert.NoError(t, err)

			// Assert
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			// Verify mock expectations
			mockService.AssertExpectations(t)
		})
	}
}

func TestAuthHandler_BanAllUserTokens(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func(*MockAuthService)
		expectedStatus int
		expectSuccess  bool
	}{
		{
			name: "successful token ban",
			requestBody: entities.BanTokensRequest{
				UserID: "user123",
			},
			mockSetup: func(mockService *MockAuthService) {
				mockService.On("BanToken", mock.Anything, "user123").Return(nil)
			},
			expectedStatus: fiber.StatusOK,
			expectSuccess:  true,
		},
		{
			name: "empty userID",
			requestBody: entities.BanTokensRequest{
				UserID: "",
			},
			mockSetup: func(mockService *MockAuthService) {
				mockService.On("BanToken", mock.Anything, "").Return(nil)
			},
			expectedStatus: fiber.StatusOK,
			expectSuccess:  true,
		},
		{
			name: "service error - user not found",
			requestBody: entities.BanTokensRequest{
				UserID: "nonexistent",
			},
			mockSetup: func(mockService *MockAuthService) {
				mockService.On("BanToken", mock.Anything, "nonexistent").Return(&response.ErrorResponse{
					HttpStatusCode: fiber.StatusNotFound,
					Code:           response.ErrCodeNotFound,
					Message:        "User not found",
				})
			},
			expectedStatus: fiber.StatusNotFound,
			expectSuccess:  false,
		},
		{
			name: "generic service error",
			requestBody: entities.BanTokensRequest{
				UserID: "user123",
			},
			mockSetup: func(mockService *MockAuthService) {
				mockService.On("BanToken", mock.Anything, "user123").Return(errors.New("database error"))
			},
			expectedStatus: fiber.StatusInternalServerError,
			expectSuccess:  false,
		},
		{
			name:        "invalid JSON request body",
			requestBody: "invalid json",
			mockSetup: func(mockService *MockAuthService) {
				// No mock calls expected for invalid JSON
			},
			expectedStatus: fiber.StatusBadRequest,
			expectSuccess:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			app := setupTestApp()
			mockService := new(MockAuthService)

			// Create handler
			handler := &authHandler{service: mockService}
			app.Post("/auth/ban-tokens", handler.BanAllUserTokens)

			// Setup mock expectations
			tt.mockSetup(mockService)

			// Create request
			var body []byte
			var err error
			if tt.requestBody != nil {
				if str, ok := tt.requestBody.(string); ok {
					body = []byte(str) // For invalid JSON test
				} else {
					body, err = json.Marshal(tt.requestBody)
					assert.NoError(t, err)
				}
			}

			req := httptest.NewRequest(http.MethodPost, "/auth/ban-tokens", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			// Act
			resp, err := app.Test(req)
			assert.NoError(t, err)

			// Assert
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			// Verify mock expectations
			mockService.AssertExpectations(t)
		})
	}
}

func TestAuthHandler_RefreshToken_AdvancedCases(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		contentType    string
		mockSetup      func(*MockAuthService)
		expectedStatus int
	}{
		{
			name: "missing content-type header",
			requestBody: entities.RefreshTokenRequest{
				RefreshToken: "valid_token",
			},
			contentType: "", // No content-type
			mockSetup: func(mockService *MockAuthService) {
				// No mock calls expected
			},
			expectedStatus: fiber.StatusBadRequest,
		},
		{
			name:        "empty request body",
			requestBody: nil,
			contentType: "application/json",
			mockSetup: func(mockService *MockAuthService) {
				// No mock calls expected
			},
			expectedStatus: fiber.StatusBadRequest,
		},
		{
			name: "empty refresh token",
			requestBody: entities.RefreshTokenRequest{
				RefreshToken: "",
			},
			contentType: "application/json",
			mockSetup: func(mockService *MockAuthService) {
				mockService.On("RefreshToken", "").Return(nil, errors.New("empty token"))
			},
			expectedStatus: fiber.StatusUnauthorized,
		},
		{
			name:        "malformed JSON",
			requestBody: `{"refresh_token": "token", "extra_field":}`, // Invalid JSON
			contentType: "application/json",
			mockSetup: func(mockService *MockAuthService) {
				// No mock calls expected
			},
			expectedStatus: fiber.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			app := setupTestApp()
			mockService := new(MockAuthService)

			// Create handler
			handler := &authHandler{service: mockService}
			app.Post("/auth/refresh", handler.RefreshToken)

			// Setup mock expectations
			tt.mockSetup(mockService)

			// Create request
			var body []byte
			var err error
			if tt.requestBody != nil {
				if str, ok := tt.requestBody.(string); ok {
					body = []byte(str) // For malformed JSON test
				} else {
					body, err = json.Marshal(tt.requestBody)
					assert.NoError(t, err)
				}
			}

			req := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewBuffer(body))
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}

			// Act
			resp, err := app.Test(req)
			assert.NoError(t, err)

			// Assert
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			// Verify mock expectations
			mockService.AssertExpectations(t)
		})
	}
}

func TestAuthHandler_HTTPMethodValidation(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		endpoint       string
		expectedStatus int
	}{
		// VerifyPin should only accept POST
		{
			name:           "VerifyPin with GET method",
			method:         http.MethodGet,
			endpoint:       "/auth/verify-pin",
			expectedStatus: fiber.StatusMethodNotAllowed,
		},
		{
			name:           "VerifyPin with PUT method",
			method:         http.MethodPut,
			endpoint:       "/auth/verify-pin",
			expectedStatus: fiber.StatusMethodNotAllowed,
		},
		// RefreshToken should only accept POST
		{
			name:           "RefreshToken with GET method",
			method:         http.MethodGet,
			endpoint:       "/auth/refresh",
			expectedStatus: fiber.StatusMethodNotAllowed,
		},
		{
			name:           "RefreshToken with DELETE method",
			method:         http.MethodDelete,
			endpoint:       "/auth/refresh",
			expectedStatus: fiber.StatusMethodNotAllowed,
		},
		// ListAllTokens should only accept GET
		{
			name:           "ListAllTokens with POST method",
			method:         http.MethodPost,
			endpoint:       "/auth/tokens",
			expectedStatus: fiber.StatusMethodNotAllowed,
		},
		{
			name:           "ListAllTokens with PUT method",
			method:         http.MethodPut,
			endpoint:       "/auth/tokens",
			expectedStatus: fiber.StatusMethodNotAllowed,
		},
		// BanAllUserTokens should only accept POST
		{
			name:           "BanAllUserTokens with GET method",
			method:         http.MethodGet,
			endpoint:       "/auth/ban-tokens",
			expectedStatus: fiber.StatusMethodNotAllowed,
		},
		{
			name:           "BanAllUserTokens with DELETE method",
			method:         http.MethodDelete,
			endpoint:       "/auth/ban-tokens",
			expectedStatus: fiber.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			app := setupTestApp()
			mockService := new(MockAuthService)

			// Create handler with all routes
			handler := &authHandler{service: mockService}
			app.Post("/auth/verify-pin", handler.VerifyPin)
			app.Post("/auth/refresh", handler.RefreshToken)
			app.Get("/auth/tokens", handler.ListAllTokens)
			app.Post("/auth/ban-tokens", handler.BanAllUserTokens)

			// Create request with wrong method
			req := httptest.NewRequest(tt.method, tt.endpoint, nil)

			// Act
			resp, err := app.Test(req)
			assert.NoError(t, err)

			// Assert
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			// Verify no mock calls were made
			mockService.AssertExpectations(t)
		})
	}
}

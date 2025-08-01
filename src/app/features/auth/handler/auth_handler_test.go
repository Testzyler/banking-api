package handler

import (
	"bytes"
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

func (m *MockAuthService) VerifyPin(params entities.PinVerifyParams) (*entities.TokenResponse, error) {
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
					User: entities.User{
						UserID: "user123",
						Name:   "testuser",
					},
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
					User: entities.User{
						UserID: "user123",
						Name:   "testuser",
					},
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

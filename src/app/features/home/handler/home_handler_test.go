package dashboard

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Testzyler/banking-api/app/entities"
	"github.com/Testzyler/banking-api/logger"
	"github.com/Testzyler/banking-api/server/middlewares"
	"github.com/Testzyler/banking-api/server/response"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)


// Mock HomeService
type MockHomeService struct {
	mock.Mock
}

func (m *MockHomeService) GetHomeData(userID string) (entities.HomeResponse, error) {
	args := m.Called(userID)
	return args.Get(0).(entities.HomeResponse), args.Error(1)
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

func TestHomeHandler_GetHomeData_BasicTests(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		mockSetup      func(*MockHomeService)
		expectedStatus int
		expectSuccess  bool
	}{
		{
			name:   "successful get home data",
			userID: "user123",
			mockSetup: func(mockService *MockHomeService) {
				homeData := entities.HomeResponse{
					User: entities.User{
						UserID:   "user123",
						Name:     "testuser",
						Greeting: "Hello, testuser!",
					},
					DebitCards: []entities.DebitCards{
						{
							CardID:   "card1",
							CardName: "Main Card",
							Status:   "active",
						},
					},
					Banners: []entities.Banner{
						{
							BannerID:    "banner1",
							Title:       "Welcome Banner",
							Description: "Welcome to our app",
						},
					},
					Transactions: []entities.Transaction{
						{
							TransactionID: "txn1",
							Name:          "Coffee Shop",
							IsBank:        false,
						},
					},
					Accounts: []entities.Account{
						{
							AccountID: "acc1",
							Type:      "savings",
							Currency:  "THB",
							Amount:    5000.0,
						},
					},
					TotalBalance: 5000.0,
				}
				mockService.On("GetHomeData", "user123").Return(homeData, nil)
			},
			expectedStatus: fiber.StatusOK,
			expectSuccess:  true,
		},
		{
			name:   "service returns error",
			userID: "user123",
			mockSetup: func(mockService *MockHomeService) {
				mockService.On("GetHomeData", "user123").Return(entities.HomeResponse{}, errors.New("service error"))
			},
			expectedStatus: fiber.StatusInternalServerError,
			expectSuccess:  false,
		},
		{
			name:   "user not found error",
			userID: "nonexistent",
			mockSetup: func(mockService *MockHomeService) {
				mockService.On("GetHomeData", "nonexistent").Return(entities.HomeResponse{}, &response.ErrorResponse{
					HttpStatusCode: fiber.StatusNotFound,
					Code:           response.ErrCodeNotFound,
					Message:        "User not found",
				})
			},
			expectedStatus: fiber.StatusNotFound,
			expectSuccess:  false,
		},
		{
			name:   "internal server error",
			userID: "user123",
			mockSetup: func(mockService *MockHomeService) {
				mockService.On("GetHomeData", "user123").Return(entities.HomeResponse{}, &response.ErrorResponse{
					HttpStatusCode: fiber.StatusInternalServerError,
					Code:           response.ErrCodeInternalServer,
					Message:        "Internal server error",
				})
			},
			expectedStatus: fiber.StatusInternalServerError,
			expectSuccess:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			app := setupTestApp()
			mockService := new(MockHomeService)

			// Create handler
			handler := &homeHandler{service: mockService}

			// Mock auth middleware by setting user in context
			app.Get("/home", func(c *fiber.Ctx) error {
				// Mock the auth middleware behavior
				claims := &entities.Claims{
					UserID: tt.userID,
				}
				c.Locals("user", claims)
				return handler.GetHomeData(c)
			})

			// Setup mock expectations
			tt.mockSetup(mockService)

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/home", nil)

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

func TestHomeHandler_GetHomeData_Integration(t *testing.T) {
	// This test focuses on the happy path only
	app := setupTestApp()
	mockService := new(MockHomeService)

	// Create handler
	handler := &homeHandler{service: mockService}

	// Setup route with mock auth middleware
	app.Get("/home", func(c *fiber.Ctx) error {
		claims := &entities.Claims{
			UserID: "user123",
		}
		c.Locals("user", claims)
		return handler.GetHomeData(c)
	})

	// Setup mock expectations for successful case
	homeData := entities.HomeResponse{
		User: entities.User{
			UserID:   "user123",
			Name:     "testuser",
			Greeting: "Hello, testuser!",
		},
		DebitCards: []entities.DebitCards{
			{
				CardID:   "card1",
				CardName: "Main Card",
				Status:   "active",
			},
		},
		Banners: []entities.Banner{
			{
				BannerID:    "banner1",
				Title:       "Welcome Banner",
				Description: "Welcome to our app",
			},
		},
		Transactions: []entities.Transaction{
			{
				TransactionID: "txn1",
				Name:          "Coffee Shop",
				IsBank:        false,
			},
		},
		Accounts: []entities.Account{
			{
				AccountID: "acc1",
				Type:      "savings",
				Currency:  "THB",
				Amount:    5000.0,
			},
		},
		TotalBalance: 5000.0,
	}
	mockService.On("GetHomeData", "user123").Return(homeData, nil)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/home", nil)

	// Act
	resp, err := app.Test(req)
	assert.NoError(t, err)

	// Assert
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	// Check Content-Type header for successful response
	assert.Contains(t, resp.Header.Get("Content-Type"), "application/json")

	// Verify mock expectations
	mockService.AssertExpectations(t)
}

func TestHomeHandler_GetHomeData_ErrorHandling(t *testing.T) {
	// Test various error scenarios without complex JSON parsing
	tests := []struct {
		name        string
		serviceErr  error
		expectError bool
	}{
		{
			name: "generic error",
			serviceErr: &response.ErrorResponse{
				HttpStatusCode: fiber.StatusInternalServerError,
				Code:           response.ErrCodeInternalServer,
				Message:        "Internal server error",
			},
			expectError: true,
		},
		{
			name:        "database connection error",
			serviceErr:  errors.New("database connection failed"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := setupTestApp()
			mockService := new(MockHomeService)

			handler := &homeHandler{service: mockService}

			app.Get("/home", func(c *fiber.Ctx) error {
				claims := &entities.Claims{UserID: "user123"}
				c.Locals("user", claims)
				return handler.GetHomeData(c)
			})

			mockService.On("GetHomeData", "user123").Return(entities.HomeResponse{}, tt.serviceErr)

			req := httptest.NewRequest(http.MethodGet, "/home", nil)
			resp, err := app.Test(req)
			assert.NoError(t, err)

			if tt.expectError {
				// Just check it's not 200 OK for error cases
				assert.NotEqual(t, fiber.StatusOK, resp.StatusCode)
			}

			mockService.AssertExpectations(t)
		})
	}
}

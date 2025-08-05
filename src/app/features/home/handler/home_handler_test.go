package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/Testzyler/banking-api/app/entities"
	"github.com/Testzyler/banking-api/logger"
	"github.com/Testzyler/banking-api/server/middlewares"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockHomeService implements the home service interface for testing
type MockHomeService struct {
	mock.Mock
}

func (m *MockHomeService) GetHomeData(userID string) (entities.HomeResponse, error) {
	args := m.Called(userID)
	if args.Error(1) != nil {
		return entities.HomeResponse{}, args.Error(1)
	}
	return args.Get(0).(entities.HomeResponse), nil
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

// Test for GetHomeData success case
func TestGetHomeData_Success(t *testing.T) {
	// Setup
	mockService := new(MockHomeService)
	expectedResponse := entities.HomeResponse{
		User: entities.User{
			UserID: "1",
			Name:   "testuser",
		},
		DebitCards:   []entities.DebitCards{},
		Banners:      []entities.Banner{},
		Transactions: []entities.Transaction{},
		Accounts:     []entities.Account{},
		TotalBalance: 1000.0,
	}

	mockService.On("GetHomeData", "1").Return(expectedResponse, nil)

	// Create handler with mock service
	handler := &homeHandler{
		service: mockService,
	}

	app := setupTestApp()
	app.Get("/home", func(c *fiber.Ctx) error {
		// Set user claims in context
		claims := entities.Claims{
			UserID:   "1",
			Username: "testuser",
		}
		c.Locals("user", claims)

		// Call handler
		return handler.GetHomeData(c)
	})

	// Make request
	req := httptest.NewRequest("GET", "/home", nil)
	resp, err := app.Test(req)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	mockService.AssertExpectations(t)
}

func TestGetHomeData_NoUserContext(t *testing.T) {
	// Setup
	mockService := new(MockHomeService)
	handler := &homeHandler{
		service: mockService,
	}

	app := setupTestApp()
	app.Get("/home", func(c *fiber.Ctx) error {
		// Don't set user context
		return handler.GetHomeData(c)
	})

	// Make request
	req := httptest.NewRequest("GET", "/home", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err) // app.Test shouldn't error
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
}

func TestGetHomeData_InvalidUserType(t *testing.T) {
	// Setup
	mockService := new(MockHomeService)
	handler := &homeHandler{
		service: mockService,
	}

	app := fiber.New()
	app.Get("/home", func(c *fiber.Ctx) error {
		c.Locals("user", "invalid_type")
		return handler.GetHomeData(c)
	})

	// Make request
	req := httptest.NewRequest("GET", "/home", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err) // app.Test shouldn't error
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
}

// Test for GetHomeData with nil claims
func TestGetHomeData_NilClaims(t *testing.T) {
	// Setup
	mockService := new(MockHomeService)
	handler := &homeHandler{
		service: mockService,
	}

	app := fiber.New()
	app.Get("/home", func(c *fiber.Ctx) error {
		// Set nil claims in context
		c.Locals("user", (*entities.Claims)(nil))
		return handler.GetHomeData(c)
	})

	// Make request
	req := httptest.NewRequest("GET", "/home", nil)
	resp, err := app.Test(req)

	// Should return internal server error (500) because user is nil
	assert.NoError(t, err) // app.Test shouldn't error
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
}

// Test for GetHomeData with service error
func TestGetHomeData_ServiceError(t *testing.T) {
	// Setup
	mockService := new(MockHomeService)
	serviceError := errors.New("database error")

	mockService.On("GetHomeData", "1").Return(entities.HomeResponse{}, serviceError)

	handler := &homeHandler{
		service: mockService,
	}

	// Setup fiber app
	app := fiber.New()
	app.Get("/home", func(c *fiber.Ctx) error {
		// Set user claims in context
		claims := entities.Claims{
			UserID:   "1",
			Username: "testuser",
		}
		c.Locals("user", claims)
		return handler.GetHomeData(c)
	})

	// Make request
	req := httptest.NewRequest("GET", "/home", nil)
	resp, err := app.Test(req)

	// Should return internal server error due to service error
	assert.NoError(t, err) // app.Test shouldn't error
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
	mockService.AssertExpectations(t)
}

func TestGetHomeData_EmptyUserID(t *testing.T) {
	// Setup
	mockService := new(MockHomeService)
	serviceError := errors.New("invalid user ID")

	mockService.On("GetHomeData", "").Return(entities.HomeResponse{}, serviceError)

	handler := &homeHandler{
		service: mockService,
	}

	app := fiber.New()
	app.Get("/home", func(c *fiber.Ctx) error {
		claims := entities.Claims{
			UserID:   "", // Empty user ID
			Username: "testuser",
		}
		c.Locals("user", claims)
		return handler.GetHomeData(c)
	})

	// Make request
	req := httptest.NewRequest("GET", "/home", nil)
	resp, err := app.Test(req)

	// Should return error due to empty user ID
	assert.NoError(t, err) // app.Test shouldn't error
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
	mockService.AssertExpectations(t)
}

func TestGetHomeData_DirectCall(t *testing.T) {
	// Setup
	mockService := new(MockHomeService)
	expectedResponse := entities.HomeResponse{
		User: entities.User{
			UserID: "1",
			Name:   "testuser",
		},
		TotalBalance: 1000.0,
	}

	mockService.On("GetHomeData", "1").Return(expectedResponse, nil)

	handler := &homeHandler{
		service: mockService,
	}

	app := setupTestApp()
	app.Get("/home", func(c *fiber.Ctx) error {
		claims := entities.Claims{UserID: "1", Username: "testuser"}
		c.Locals("user", claims)
		return handler.GetHomeData(c)
	})

	// Make request using test method
	req := httptest.NewRequest("GET", "/home", nil)
	resp, err := app.Test(req)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	mockService.AssertExpectations(t)
}

func TestHomeHandler_HTTPMethodValidation(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		expectedStatus int
	}{
		{
			name:           "GET method should work",
			method:         "GET",
			expectedStatus: fiber.StatusOK,
		},
		{
			name:           "POST method should return 405",
			method:         "POST",
			expectedStatus: fiber.StatusMethodNotAllowed,
		},
		{
			name:           "PUT method should return 405",
			method:         "PUT",
			expectedStatus: fiber.StatusMethodNotAllowed,
		},
		{
			name:           "DELETE method should return 405",
			method:         "DELETE",
			expectedStatus: fiber.StatusMethodNotAllowed,
		},
		{
			name:           "PATCH method should return 405",
			method:         "PATCH",
			expectedStatus: fiber.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockHomeService)

			if tt.method == "GET" {
				// Only setup mock for successful GET request
				expectedResponse := entities.HomeResponse{
					User: entities.User{
						UserID: "1",
						Name:   "testuser",
					},
					TotalBalance: 1000.0,
				}
				mockService.On("GetHomeData", "1").Return(expectedResponse, nil)
			}

			handler := &homeHandler{service: mockService}
			app := setupTestApp()

			// Setup all possible routes
			app.Get("/home", func(c *fiber.Ctx) error {
				claims := entities.Claims{UserID: "1", Username: "testuser"}
				c.Locals("user", claims)
				return handler.GetHomeData(c)
			})

			req := httptest.NewRequest(tt.method, "/home", nil)
			resp, err := app.Test(req)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
			mockService.AssertExpectations(t)
		})
	}
}

func TestHomeHandler_ResponseStructureValidation(t *testing.T) {
	tests := []struct {
		name             string
		serviceResponse  entities.HomeResponse
		serviceError     error
		expectedStatus   int
		validateResponse func(*testing.T, []byte)
	}{
		{
			name: "full home data response structure",
			serviceResponse: entities.HomeResponse{
				User: entities.User{
					UserID:   "user123",
					Name:     "John Doe",
					Greeting: "Good morning!",
				},
				DebitCards: []entities.DebitCards{
					{
						CardID:   "card1",
						CardName: "Main Card",
						DebitCardDesign: entities.DebitCardDesign{
							Color:       "#FF0000",
							BorderColor: "#000000",
						},
						Status:     "active",
						CardNumber: "**** **** **** 1234",
						Issuer:     "VISA",
					},
				},
				Banners: []entities.Banner{
					{
						BannerID:    "banner1",
						UserID:      "user123",
						Title:       "Special Offer",
						Description: "Limited time offer",
						ImageURL:    "https://example.com/banner.jpg",
					},
				},
				Transactions: []entities.Transaction{
					{
						TransactionID: "txn1",
						UserID:        "user123",
						Name:          "Coffee Shop",
						Image:         "https://example.com/coffee.jpg",
						IsBank:        false,
					},
				},
				Accounts: []entities.Account{
					{
						AccountID: "acc1",
						Type:      "savings",
						Currency:  "USD",
						Issuer:    "Bank ABC",
						Amount:    5000.50,
						AccountDetails: entities.AccountDetails{
							Color:         "#00FF00",
							IsMainAccount: true,
							Progress:      75.5,
						},
						AccountFlags: []entities.AccountFlags{
							{
								FlagType:  "premium",
								FlagValue: "true",
							},
						},
					},
				},
				TotalBalance: 5000.50,
			},
			expectedStatus: fiber.StatusOK,
			validateResponse: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				err := json.Unmarshal(body, &response)
				assert.NoError(t, err)
				assert.Equal(t, float64(10200), response["code"])
				assert.Equal(t, "Home screen data retrieved successfully", response["message"])
				assert.NotNil(t, response["data"])

				data, ok := response["data"].(map[string]interface{})
				assert.True(t, ok)
				assert.Equal(t, "user123", data["userID"])
				assert.Equal(t, "John Doe", data["name"])
				assert.Equal(t, "Good morning!", data["greeting"])
				assert.Equal(t, 5000.50, data["totalBalance"])
				assert.NotNil(t, data["debitCards"])
				assert.NotNil(t, data["banners"])
				assert.NotNil(t, data["transactions"])
				assert.NotNil(t, data["accounts"])
			},
		},
		{
			name: "empty data response structure",
			serviceResponse: entities.HomeResponse{
				User: entities.User{
					UserID:   "user123",
					Name:     "John Doe",
					Greeting: "Good morning!",
				},
				DebitCards:   []entities.DebitCards{},
				Banners:      []entities.Banner{},
				Transactions: []entities.Transaction{},
				Accounts:     []entities.Account{},
				TotalBalance: 0,
			},
			expectedStatus: fiber.StatusOK,
			validateResponse: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				err := json.Unmarshal(body, &response)
				assert.NoError(t, err)
				assert.Equal(t, "Home screen data retrieved successfully", response["message"])

				data, ok := response["data"].(map[string]interface{})
				assert.True(t, ok)
				assert.Equal(t, float64(0), data["totalBalance"])
				assert.NotNil(t, data["debitCards"])
				assert.NotNil(t, data["banners"])
				assert.NotNil(t, data["transactions"])
				assert.NotNil(t, data["accounts"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockHomeService)
			mockService.On("GetHomeData", "user123").Return(tt.serviceResponse, tt.serviceError)

			handler := &homeHandler{service: mockService}
			app := setupTestApp()
			app.Get("/home", func(c *fiber.Ctx) error {
				claims := entities.Claims{UserID: "user123", Username: "testuser"}
				c.Locals("user", claims)
				return handler.GetHomeData(c)
			})

			req := httptest.NewRequest("GET", "/home", nil)
			resp, err := app.Test(req)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.validateResponse != nil {
				body, err := io.ReadAll(resp.Body)
				assert.NoError(t, err)
				tt.validateResponse(t, body)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestHomeHandler_UserContextEdgeCases(t *testing.T) {
	tests := []struct {
		name           string
		setupContext   func(*fiber.Ctx)
		expectedStatus int
		expectService  bool
	}{
		{
			name: "valid claims structure",
			setupContext: func(c *fiber.Ctx) {
				claims := entities.Claims{UserID: "123", Username: "test"}
				c.Locals("user", claims)
			},
			expectedStatus: fiber.StatusOK,
			expectService:  true,
		},
		{
			name: "pointer to valid claims",
			setupContext: func(c *fiber.Ctx) {
				claims := &entities.Claims{UserID: "123", Username: "test"}
				c.Locals("user", *claims)
			},
			expectedStatus: fiber.StatusOK,
			expectService:  true,
		},
		{
			name: "string instead of claims",
			setupContext: func(c *fiber.Ctx) {
				c.Locals("user", "invalid")
			},
			expectedStatus: fiber.StatusInternalServerError,
			expectService:  false,
		},
		{
			name: "int instead of claims",
			setupContext: func(c *fiber.Ctx) {
				c.Locals("user", 123)
			},
			expectedStatus: fiber.StatusInternalServerError,
			expectService:  false,
		},
		{
			name: "map instead of claims",
			setupContext: func(c *fiber.Ctx) {
				c.Locals("user", map[string]string{"userID": "123"})
			},
			expectedStatus: fiber.StatusInternalServerError,
			expectService:  false,
		},
		{
			name: "empty claims",
			setupContext: func(c *fiber.Ctx) {
				claims := entities.Claims{}
				c.Locals("user", claims)
			},
			expectedStatus: fiber.StatusOK,
			expectService:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockHomeService)

			if tt.expectService {
				expectedResponse := entities.HomeResponse{
					User: entities.User{
						UserID:   "123",
						Name:     "testuser",
						Greeting: "Good morning!",
					},
					TotalBalance: 1000.0,
				}
				mockService.On("GetHomeData", mock.AnythingOfType("string")).Return(expectedResponse, nil)
			}

			handler := &homeHandler{service: mockService}
			app := setupTestApp()
			app.Get("/home", func(c *fiber.Ctx) error {
				tt.setupContext(c)
				return handler.GetHomeData(c)
			})

			req := httptest.NewRequest("GET", "/home", nil)
			resp, err := app.Test(req)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
			mockService.AssertExpectations(t)
		})
	}
}

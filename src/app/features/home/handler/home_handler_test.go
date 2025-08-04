package handler

import (
	"errors"
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
		claims := &entities.Claims{
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
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
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
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
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

	// Should return unauthorized error
	assert.NoError(t, err) // app.Test shouldn't error
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
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
		claims := &entities.Claims{
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
		claims := &entities.Claims{
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

	app := fiber.New()
	c := app.AcquireCtx(nil)
	defer app.ReleaseCtx(c)

	// Initialize request
	c.Request().SetRequestURI("/home")
	c.Request().Header.SetMethod("GET")

	// Set user claims
	claims := &entities.Claims{
		UserID:   "1",
		Username: "testuser",
	}
	c.Locals("user", claims)

	// Call handler directly
	err := handler.GetHomeData(c)

	// Assertions
	assert.NoError(t, err)
	mockService.AssertExpectations(t)
}

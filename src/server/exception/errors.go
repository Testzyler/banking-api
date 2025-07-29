package exception

import (
	"github.com/Testzyler/banking-api/server/response"
	"github.com/gofiber/fiber/v2"
)

// Predefined error responses that implement the error interface (error_handler)
var (
	// 4xx Client Errors
	ErrUserNotFound = &response.ErrorResponse{
		HttpStatusCode: fiber.StatusNotFound,
		Code:           response.ErrCodeNotFound,
		Message:        "User not found",
		Details:        "The user with the specified ID does not exist",
	}

	ErrInvalidUserID = &response.ErrorResponse{
		HttpStatusCode: fiber.StatusBadRequest,
		Code:           response.ErrCodeBadRequest,
		Message:        "Invalid user ID",
		Details:        "The provided user ID is not valid or empty",
	}

	ErrInvalidPagination = &response.ErrorResponse{
		HttpStatusCode: fiber.StatusBadRequest,
		Code:           response.ErrCodeBadRequest,
		Message:        "Invalid pagination parameters",
		Details:        "Page number must be positive and perPage must be between 1 and 100",
	}

	ErrValidationFailed = &response.ErrorResponse{
		HttpStatusCode: fiber.StatusUnprocessableEntity,
		Code:           response.ErrCodeValidationFailed,
		Message:        "Validation failed",
		Details:        "The provided data did not pass validation checks",
	}

	ErrUnauthorized = &response.ErrorResponse{
		HttpStatusCode: fiber.StatusUnauthorized,
		Code:           response.ErrCodeUnauthorized,
		Message:        "Unauthorized",
		Details:        "Authentication credentials are missing or invalid",
	}

	ErrForbidden = &response.ErrorResponse{
		HttpStatusCode: fiber.StatusForbidden,
		Code:           response.ErrCodeForbidden,
		Message:        "Forbidden",
		Details:        "You do not have permission to access this resource",
	}

	// 5xx Server Errors
	ErrInternalServer = &response.ErrorResponse{
		HttpStatusCode: fiber.StatusInternalServerError,
		Code:           response.ErrCodeInternalServer,
		Message:        "Internal server error",
		Details:        "An unexpected error occurred while processing your request",
	}

	ErrDatabaseConnection = &response.ErrorResponse{
		HttpStatusCode: fiber.StatusInternalServerError,
		Code:           response.ErrCodeDatabaseError,
		Message:        "Database connection error",
		Details:        "Unable to connect to the database",
	}

	ErrServiceUnavailable = &response.ErrorResponse{
		HttpStatusCode: fiber.StatusServiceUnavailable,
		Code:           response.ErrCodeServiceUnavailable,
		Message:        "Service unavailable",
		Details:        "The service is temporarily unavailable",
	}
)

// Helper functions to create custom errors with dynamic details
func NewUserNotFoundError(userID string) *response.ErrorResponse {
	return &response.ErrorResponse{
		HttpStatusCode: fiber.StatusNotFound,
		Code:           response.ErrCodeNotFound,
		Message:        "User not found",
		Details:        "User with ID '" + userID + "' does not exist",
	}
}

func NewValidationError(details interface{}) *response.ErrorResponse {
	return &response.ErrorResponse{
		HttpStatusCode: fiber.StatusUnprocessableEntity,
		Code:           response.ErrCodeValidationFailed,
		Message:        "Validation failed",
		Details:        details,
	}
}

func NewInternalError(err error) *response.ErrorResponse {
	return &response.ErrorResponse{
		HttpStatusCode: fiber.StatusInternalServerError,
		Code:           response.ErrCodeInternalServer,
		Message:        "Internal server error",
		Details:        err.Error(),
	}
}

func NewDatabaseError(err error) *response.ErrorResponse {
	return &response.ErrorResponse{
		HttpStatusCode: fiber.StatusInternalServerError,
		Code:           response.ErrCodeDatabaseError,
		Message:        "Database error",
		Details:        err.Error(),
	}
}

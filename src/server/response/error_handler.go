package response

import (
	"github.com/Testzyler/banking-api/server/middlewares"
	"github.com/gofiber/fiber/v2"
)

func ErrorHandler(c *fiber.Ctx, err error) error {
	// Use request-scoped logger if available, otherwise fallback to global logger
	requestLogger := middlewares.GetRequestLogger(c)
	requestLogger.Errorw("Request error", "error", err, "path", c.Path())

	// If the error is a fiber error, return it directly
	if cErr, ok := err.(*fiber.Error); ok {
		return c.Status(cErr.Code).JSON(&ErrorResponse{
			Message: cErr.Message,
			Code:    ErrCodeInternalServer,
			Details: "An error occurred while processing your request.",
		})
	}

	// If the error is a custom ErrorResponse, return it directly
	if cErr, ok := err.(*ErrorResponse); ok {
		exception := NewException(cErr.Code, cErr.Message, cErr.Details)
		return c.Status(cErr.HttpStatusCode).JSON(&ErrorResponse{
			Message: exception.Message,
			Code:    exception.Code,
			Details: exception.Details,
		})
	}

	return c.Status(fiber.StatusInternalServerError).JSON(&ErrorResponse{
		Message: err.Error(),
		Code:    ErrCodeInternalServer,
		Details: "An error occurred while processing your request.",
	})
}

func NotFoundHandler(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotFound).JSON(&ErrorResponse{
		Code:    ErrCodeNotFound,
		Message: "Resource not found",
		Details: "The requested resource could not be found on the server.",
	})
}

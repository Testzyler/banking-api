package middlewares

import (
	"github.com/Testzyler/banking-api/server/exception"
	"github.com/Testzyler/banking-api/server/response"
	"github.com/gofiber/fiber/v2"
)

// ErrorHandler
func ErrorHandler() fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		// Get request logger if available
		requestLogger := GetRequestLogger(c)

		// Check if it's a custom error response type
		if errResp, ok := err.(*response.ErrorResponse); ok {
			if errResp.HttpStatusCode >= 500 {
				requestLogger.Errorf("Server error: %s (Source: %s)", errResp.Message, errResp.Source)
			} else {
				requestLogger.Warnf("Client error: %s", errResp.Message)
			}
			return c.Status(errResp.HttpStatusCode).JSON(errResp)
		}

		// Check if it's a fiber error
		if fiberErr, ok := err.(*fiber.Error); ok {
			requestLogger.Warnf("Fiber error: %s", fiberErr.Message)
			return c.Status(fiberErr.Code).JSON(&response.ErrorResponse{
				HttpStatusCode: fiberErr.Code,
				Code:           response.ErrCodeBadRequest,
				Message:        fiberErr.Message,
				Details:        "Request processing failed",
			})
		}

		// Unexpected error - log and return generic error
		requestLogger.Errorf("Unexpected error: %v", err)
		return c.Status(exception.ErrInternalServer.HttpStatusCode).JSON(exception.ErrInternalServer)
	}
}

func NotFoundHandler(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotFound).JSON(&response.ErrorResponse{
		HttpStatusCode: fiber.StatusNotFound,
		Code:           response.ErrCodeNotFound,
		Message:        "Resource not found",
		Details:        "The requested resource could not be found on the server.",
	})
}

package middlewares

import (
	"github.com/Testzyler/banking-api/server/exception"
	"github.com/Testzyler/banking-api/server/response"
	"github.com/gofiber/fiber/v2"
)

// ErrorHandler
func ErrorHandler() fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		// Check if it's a custom error response type
		if errResp, ok := err.(*response.ErrorResponse); ok {
			return c.Status(errResp.HttpStatusCode).JSON(errResp)
		}

		// Check if it's a fiber error
		if fiberErr, ok := err.(*fiber.Error); ok {
			return c.Status(fiberErr.Code).JSON(&response.ErrorResponse{
				HttpStatusCode: fiber.StatusInternalServerError,
				Code:           response.ErrCodeBadRequest,
				Message:        fiberErr.Message,
				Details:        "Request processing failed",
			})
		}

		responseError := exception.NewInternalError(err)
		return c.Status(fiber.StatusInternalServerError).JSON(responseError)
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

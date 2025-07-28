package middlewares

import (
	"github.com/Testzyler/banking-api/server/response"
	"github.com/gofiber/fiber/v2"
)

func GlobalErrorMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Handle errors globally
		if err := c.Next(); err != nil {
			// Log the error (optional)
			// log.Error(err)

			// Return a generic error response
			return c.Status(fiber.StatusInternalServerError).JSON(&response.ErrorResponse{
				Message: "Internal Server Error",
				Code:    response.ErrCodeInternalServer,
				Details: "An unexpected error occurred",
			})
		}
		return nil
	}
}

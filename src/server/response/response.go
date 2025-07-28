package response

import (
	"github.com/Testzyler/banking-api/server/exceptions"
	"github.com/gofiber/fiber/v2"
)

// BaseResponse represents the base structure for all API responses
type BaseResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// SuccessResponse represents a successful API response
type SuccessResponse struct {
	BaseResponse
}

type PaginatedResponse struct {
	BaseResponse
	Meta PaginationMeta `json:"meta"`
}

type PaginationMeta struct {
	Page        int  `json:"page"`
	PerPage     int  `json:"perPage"`
	Total       int  `json:"total"`
	TotalPages  int  `json:"totalPages"`
	HasNext     bool `json:"hasNext"`
	HasPrevious bool `json:"hasPrevious"`
}

// ErrorResponse represents detailed error response
type ErrorResponse struct {
	BaseResponse
	Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	ErrorCode exceptions.ErrorCode `json:"errorCode"`
	Details   interface{}          `json:"details,omitempty"`
}

func NotFoundHandler(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotFound).JSON(&ErrorResponse{
		BaseResponse: BaseResponse{
			Message: "Resource not found",
		},
		Error: ErrorDetail{
			ErrorCode: exceptions.ErrCodeNotFound,
			Details:   "The requested resource could not be found",
		},
	})
}

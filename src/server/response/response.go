package response

import (
	"fmt"
)

// SuccessResponse represents a successful API response
type SuccessResponse struct {
	Code    ResponseCode `json:"code"`
	Message string       `json:"message"`
	Data    interface{}  `json:"data,omitempty"`
}

// ErrorResponse represents detailed error response
type ErrorResponse struct {
	HttpStatusCode int          `json:"-"`
	Code           ResponseCode `json:"code"`
	Message        string       `json:"message"`
	Details        interface{}  `json:"details,omitempty"`
	Source         string       `json:"-"` // For logging purposes only, not returned to client
}

func (e *ErrorResponse) Error() string {
	if e.Details != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Details)
	}
	return e.Message
}

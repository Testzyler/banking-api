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
}

func (e *ErrorResponse) Error() string {
	if e.Details != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Details)
	}
	return e.Message
}

func NewException(code ResponseCode, message string, details ...interface{}) *ErrorResponse {
	err := &ErrorResponse{
		Code:    code,
		Message: message,
		Details: details,
	}
	if message == "" {
		err.Message = CustomResponseMessages[code]
	} else {
		err.Message = fmt.Sprintf("%s: %s", CustomResponseMessages[code], message)
	}

	if details != nil {
		err.Details = details
	}

	return err
}

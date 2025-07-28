package entities

import "github.com/Testzyler/banking-api/server/response"

type GetUserByIdParams struct {
	UserID string `json:"user_id" validate:"required"`
}

func (p *GetUserByIdParams) Validate() error {
	if p.UserID == "" {
		return response.NewException(response.ErrCodeValidationFailed, "User ID is required", "user_id cannot be empty")
	}
	if len(p.UserID) < 3 || len(p.UserID) > 50 {
		return response.NewException(response.ErrCodeValidationFailed, "Invalid User ID length", "user_id must be between 3 and 50 characters")
	}
	return nil
}

package entities

import "github.com/Testzyler/banking-api/server/exception"

type GetUserByIdParams struct {
	UserID string `json:"user_id" validate:"required"`
}

func (p *GetUserByIdParams) Validate() error {
	if p.UserID == "" {
		return exception.ErrInvalidUserID
	}
	if len(p.UserID) < 3 || len(p.UserID) > 50 {
		return exception.ErrInvalidUserID
	}
	return nil
}

package entities

import "github.com/Testzyler/banking-api/app/validators"

type GetUserByIdParams struct {
	UserID string `json:"user_id" validate:"required,min=3,max=50"`
}

func (p *GetUserByIdParams) Validate() error {
	return validators.ValidateStruct(p)
}

type User struct {
	UserID   string `json:"userID"`
	Name     string `json:"name"`
	Greeting string `json:"greeting"`
}

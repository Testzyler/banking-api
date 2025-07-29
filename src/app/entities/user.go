package entities

import "github.com/Testzyler/banking-api/utils"

type GetUserByIdParams struct {
	UserID string `json:"user_id" validate:"required,min=3,max=50"`
}

func (p *GetUserByIdParams) Validate() error {
	return utils.ValidateStruct(p)
}

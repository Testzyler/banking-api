package entities

type GetUserByIdParams struct {
	UserID string `json:"user_id" validate:"required"`
}

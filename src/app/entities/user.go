package entities

import (
	"time"

	"github.com/Testzyler/banking-api/app/validators"
)

type GetUserByIdParams struct {
	UserID string `json:"user_id" validate:"required,min=3,max=50"`
}

func (p *GetUserByIdParams) Validate() error {
	return validators.ValidateStruct(p)
}

type User struct {
	UserID   string  `json:"userID"`
	Name     string  `json:"name"`
	Greeting string  `json:"greeting,omitempty"`
	UserPin  UserPin `json:"user_pin"`
}

type UserPin struct {
	HashedPin         string    `json:"hashed_pin"`
	FailedPinAttempts int       `json:"failed_pin_attempts"`
	PinLockedUntil    time.Time `json:"pin_locked_until"`
	LastPinAttemptAt  time.Time `json:"last_pin_attempt_at"`
}

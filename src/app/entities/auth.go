package entities

import (
	"time"

	"github.com/Testzyler/banking-api/app/validators"
	"github.com/golang-jwt/jwt/v5"
)

type PinVerifyParams struct {
	Username string `json:"username"`
	Pin      string `json:"pin" validate:"required,min=6,max=6,numeric"`
}

func (p *PinVerifyParams) Validate() error {
	return validators.ValidateStruct(p)
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" validate:"required"`
}

type TokenResponse struct {
	Token        string    `json:"token"`
	Expiry       time.Time `json:"expiry"`
	RefreshToken string    `json:"refreshToken"`
	UserID       string    `json:"userID"`
	User
}

type Claims struct {
	UserID   string `json:"userID"`
	Username string `json:"username"`
	Type     string `json:"type" validate:"required,oneof=access refresh"`
	jwt.RegisteredClaims
}

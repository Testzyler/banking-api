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
	TokenVersion int64     `json:"tokenVersion"` // Track when token was issued (timestamp)
	TokenID      string    `json:"tokenID"`      // Unique token identifier
	IsBanned     *bool     `json:"isBanned,omitempty"`
}

type Claims struct {
	UserID       string `json:"userID"`
	Username     string `json:"username"`
	Type         string `json:"type" validate:"required,oneof=access refresh"`
	TokenVersion int64  `json:"tokenVersion"`
	TokenID      string `json:"tokenID"`
	jwt.RegisteredClaims
}

type BannedToken struct {
	TokenID      string    `json:"tokenID"`
	UserID       string    `json:"userID"`
	Token        string    `json:"token"`
	BannedAt     time.Time `json:"bannedAt"`
	Reason       string    `json:"reason,omitempty"`
	TokenVersion int64     `json:"tokenVersion"`
}

type BlacklistBan struct {
	UserID       string    `json:"userID"`
	BannedAt     time.Time `json:"bannedAt"`
	Reason       string    `json:"reason,omitempty"`
	BanTimestamp int64     `json:"banTimestamp"` // Timestamp when user was banned - tokens issued before this are invalid
}

type TokenValidationResult struct {
	Valid        bool   `json:"valid"`
	Reason       string `json:"reason,omitempty"`
	TokenVersion int64  `json:"tokenVersion"`
	Claims       Claims `json:"claims,omitempty"`
}

type PinAttemptData struct {
	UserID         string     `json:"userID"`
	FailedAttempts int        `json:"failedAttempts"`
	PinLockedUntil *time.Time `json:"pinLockedUntil,omitempty"`
	LastAttemptAt  *time.Time `json:"lastAttemptAt,omitempty"`
}

type BanTokensRequest struct {
	UserID string `json:"userID" validate:"required"`
}

func (b *BanTokensRequest) Validate() error {
	return validators.ValidateStruct(b)
}

type GenerateTokenParams struct {
	UserID       string
	Username     string
	TokenVersion int64
	TokenID      string
	TokenType    string
}

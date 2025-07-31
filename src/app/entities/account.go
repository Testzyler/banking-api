package entities

import "time"

type Account struct {
	AccountID      string         `json:"accountID"`
	Type           string         `json:"type"`
	Currency       string         `json:"currency"`
	AccountNumber  string         `json:"accountNumber"`
	Issuer         string         `json:"issuer"`
	AccountDetails AccountDetails `json:"accountDetails"`
	AccountFlags   AccountFlags   `json:"accountFlags"`
	AccountBalance
}

type AccountDetails struct {
	Color         string `json:"color"`
	IsMainAccount bool   `json:"isMainAccount"`
	Progress      float64
}

type AccountBalance struct {
	Amount float64 `json:"amount"`
}

type AccountFlags struct {
	AccountID string    `json:"accountID"`
	FlagType  string    `json:"flagType"`
	FlagValue string    `json:"flagValue"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

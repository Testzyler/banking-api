package entities

import "time"

type Account struct {
	AccountID      string         `json:"accountID"`
	Amount         float64        `json:"amount"`
	Type           string         `json:"type"`
	Currency       string         `json:"currency"`
	AccountNumber  string         `json:"accountNumber"`
	Issuer         string         `json:"issuer"`
	AccountDetails AccountDetails `json:"accountDetails"`
	AccountFlags   []AccountFlags `json:"accountFlags"`
}

type AccountDetails struct {
	Color         string `json:"color"`
	IsMainAccount bool   `json:"isMainAccount"`
	Progress      float64
}

type AccountFlags struct {
	FlagType  string    `json:"flagType"`
	FlagValue string    `json:"flagValue"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

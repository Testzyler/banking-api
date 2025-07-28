package models

import "time"

type Account struct {
	AccountID     string `json:"account_id" gorm:"column:account_id;primaryKey"`
	UserID        string `json:"user_id" gorm:"column:user_id"`
	Type          string `json:"type" gorm:"column:type"`
	Currency      string `json:"currency" gorm:"column:currency"`
	AccountNumber string `json:"account_number" gorm:"column:account_number"`
	Issuer        string `json:"issuer" gorm:"column:issuer"`
	DummyCol      string `json:"dummy_col_3,omitempty" gorm:"column:dummy_col_3"`
}

func (Account) TableName() string {
	return "accounts"
}

type AccountBalance struct {
	AccountID string  `json:"account_id" gorm:"column:account_id;primaryKey"`
	UserID    string  `json:"user_id" gorm:"column:user_id"`
	Amount    float64 `json:"amount" gorm:"column:amount"`
	DummyCol  string  `json:"dummy_col_4,omitempty" gorm:"column:dummy_col_4"`
}

func (AccountBalance) TableName() string {
	return "account_balances"
}

type AccountDetail struct {
	AccountID     string `json:"account_id" gorm:"column:account_id;primaryKey"`
	UserID        string `json:"user_id" gorm:"column:user_id"`
	Color         string `json:"color" gorm:"column:color"`
	IsMainAccount bool   `json:"is_main_account" gorm:"column:is_main_account"`
	Progress      int    `json:"progress" gorm:"column:progress"`
	DummyCol      string `json:"dummy_col_5,omitempty" gorm:"column:dummy_col_5"`
}

func (AccountDetail) TableName() string {
	return "account_details"
}

type AccountFlag struct {
	FlagID    int       `json:"flag_id" gorm:"column:flag_id;primaryKey;autoIncrement"`
	AccountID string    `json:"account_id" gorm:"column:account_id"`
	UserID    string    `json:"user_id" gorm:"column:user_id"`
	FlagType  string    `json:"flag_type" gorm:"column:flag_type"`
	FlagValue string    `json:"flag_value" gorm:"column:flag_value"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at"`
}

func (AccountFlag) TableName() string {
	return "account_flags"
}

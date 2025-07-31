package models

import "time"

type Account struct {
	AccountID     string `gorm:"column:account_id;primaryKey"`
	UserID        string `gorm:"column:user_id"`
	Type          string `gorm:"column:type"`
	Currency      string `gorm:"column:currency"`
	AccountNumber string `gorm:"column:account_number"`
	Issuer        string `gorm:"column:issuer"`
	DummyCol      string `gorm:"column:dummy_col_3"`
}

func (Account) TableName() string {
	return "accounts"
}

type AccountBalance struct {
	AccountID string  `gorm:"column:account_id;primaryKey"`
	UserID    string  `gorm:"column:user_id"`
	Amount    float64 `gorm:"column:amount"`
	DummyCol  string  `gorm:"column:dummy_col_4"`
}

func (AccountBalance) TableName() string {
	return "account_balances"
}

type AccountDetail struct {
	AccountID     string  `gorm:"column:account_id;primaryKey"`
	UserID        string  `gorm:"column:user_id"`
	Color         string  `gorm:"column:color"`
	IsMainAccount bool    `gorm:"column:is_main_account"`
	Progress      float64 `gorm:"column:progress"`
	DummyCol      string  `gorm:"column:dummy_col_5"`
}

func (AccountDetail) TableName() string {
	return "account_details"
}

type AccountFlag struct {
	FlagID    int       `gorm:"column:flag_id;primaryKey;autoIncrement"`
	AccountID string    `gorm:"column:account_id"`
	UserID    string    `gorm:"column:user_id"`
	FlagType  string    `gorm:"column:flag_type"`
	FlagValue string    `gorm:"column:flag_value"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (AccountFlag) TableName() string {
	return "account_flags"
}

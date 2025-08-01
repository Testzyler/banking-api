package models

import "time"

type Account struct {
	AccountID     string `gorm:"column:account_id;primaryKey"`
	UserID        string `gorm:"column:user_id"`
	Type          string `gorm:"column:type"`
	Currency      string `gorm:"column:currency"`
	AccountNumber string `gorm:"column:account_number"`
	Issuer        string `gorm:"column:issuer"`

	AccountDetails AccountDetail  `gorm:"foreignKey:AccountID"`
	AccountBalance AccountBalance `gorm:"foreignKey:AccountID"`
	AccountFlags   []AccountFlag  `gorm:"foreignKey:AccountID"`
}

func (Account) TableName() string {
	return "accounts"
}

type AccountBalance struct {
	AccountID string  `gorm:"column:account_id;primaryKey"`
	UserID    string  `gorm:"column:user_id"`
	Amount    float64 `gorm:"column:amount"`
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

	User *User `gorm:"foreignKey:UserID;references:UserID" json:"user,omitempty"`
}

func (AccountFlag) TableName() string {
	return "account_flags"
}

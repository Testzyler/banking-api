package models

import "time"

type User struct {
	UserID    string     `gorm:"column:user_id;primaryKey"`
	Name      string     `gorm:"column:name"`
	Password  string     `gorm:"column:password"`
	UpdatedAt *time.Time `gorm:"column:updated_at;autoUpdateTime"`

	// Relationships - Proper GORM associations
	Accounts      []Account       `gorm:"foreignKey:UserID;references:UserID" json:"accounts,omitempty"`
	AccountDetail []AccountDetail `gorm:"foreignKey:UserID;references:UserID" json:"accountDetail,omitempty"`
	AccountFlag   []AccountFlag   `gorm:"foreignKey:UserID;references:UserID" json:"accountFlag,omitempty"`
	DebitCards    []DebitCard     `gorm:"foreignKey:UserID;references:UserID" json:"debitCards,omitempty"`
	Transactions  []Transaction   `gorm:"foreignKey:UserID;references:UserID" json:"transactions,omitempty"`
	Banners       []Banner        `gorm:"foreignKey:UserID;references:UserID" json:"banners,omitempty"`

	UserPin      *UserPin      `gorm:"foreignKey:UserID;references:UserID" json:"userPin,omitempty"`
	UserGreeting *UserGreeting `gorm:"foreignKey:UserID;references:UserID" json:"userGreeting,omitempty"`
}

func (User) TableName() string {
	return "users"
}

type UserGreeting struct {
	UserID   string `gorm:"column:user_id;primaryKey"`
	Greeting string `gorm:"column:greeting"`
}

func (UserGreeting) TableName() string {
	return "user_greetings"
}

type UserPin struct {
	UserID            string     `gorm:"column:user_id;primaryKey;not null;index"`
	HashedPin         string     `gorm:"column:hashed_pin"`
	FailedPinAttempts int        `gorm:"column:failed_pin_attempts"`
	LastPinAttemptAt  *time.Time `gorm:"column:last_pin_attempt_at"`
	PinLockedUntil    *time.Time `gorm:"column:pin_locked_until"`

	// Relationship back to User
	User *User `gorm:"foreignKey:UserID;references:UserID" json:"user,omitempty"`
}

func (UserPin) TableName() string {
	return "user_pins"
}

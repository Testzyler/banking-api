package models

import "time"

type User struct {
	UserID   string `gorm:"column:user_id;primaryKey"`
	Name     string `gorm:"column:name"`
	Password string `gorm:"column:password"`
	DummyCol string `gorm:"column:dummy_col_1"`

	// Relationships - Proper GORM associations
	UserPin      *UserPin      `gorm:"foreignKey:UserID;references:UserID" json:"userPin,omitempty"`
	UserGreeting *UserGreeting `gorm:"foreignKey:UserID;references:UserID" json:"userGreeting,omitempty"`
}

func (User) TableName() string {
	return "users"
}

type UserGreeting struct {
	UserID   string `gorm:"column:user_id;primaryKey"`
	Greeting string `gorm:"column:greeting"`
	DummyCol string `gorm:"column:dummy_col_2"`
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

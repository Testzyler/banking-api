package models

type User struct {
	UserID   string `gorm:"column:user_id;primaryKey"`
	Name     string `gorm:"column:name"`
	DummyCol string `gorm:"column:dummy_col_1"`
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

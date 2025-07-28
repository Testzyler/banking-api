package models

type User struct {
	UserID   string `json:"user_id" gorm:"column:user_id;primaryKey"`
	Name     string `json:"name" gorm:"column:name"`
	DummyCol string `json:"dummy_col_1,omitempty" gorm:"column:dummy_col_1"`
}

func (User) TableName() string {
	return "users"
}

type UserGreeting struct {
	UserID   string `json:"user_id" gorm:"column:user_id;primaryKey"`
	Greeting string `json:"greeting" gorm:"column:greeting"`
	DummyCol string `json:"dummy_col_2,omitempty" gorm:"column:dummy_col_2"`
}

func (UserGreeting) TableName() string {
	return "user_greetings"
}

package models

type DebitCard struct {
	CardID   string `json:"card_id" gorm:"column:card_id;primaryKey"`
	UserID   string `json:"user_id" gorm:"column:user_id"`
	Name     string `json:"name" gorm:"column:name"`
	DummyCol string `json:"dummy_col_7,omitempty" gorm:"column:dummy_col_7"`
}

func (DebitCard) TableName() string {
	return "debit_cards"
}

type DebitCardStatus struct {
	CardID   string `json:"card_id" gorm:"column:card_id;primaryKey"`
	UserID   string `json:"user_id" gorm:"column:user_id"`
	Status   string `json:"status" gorm:"column:status"`
	DummyCol string `json:"dummy_col_8,omitempty" gorm:"column:dummy_col_8"`
}

func (DebitCardStatus) TableName() string {
	return "debit_card_status"
}

type DebitCardDesign struct {
	CardID      string `json:"card_id" gorm:"column:card_id;primaryKey"`
	UserID      string `json:"user_id" gorm:"column:user_id"`
	Color       string `json:"color" gorm:"column:color"`
	BorderColor string `json:"border_color" gorm:"column:border_color"`
	DummyCol    string `json:"dummy_col_9,omitempty" gorm:"column:dummy_col_9"`
}

func (DebitCardDesign) TableName() string {
	return "debit_card_design"
}

type DebitCardDetail struct {
	CardID   string `json:"card_id" gorm:"column:card_id;primaryKey"`
	UserID   string `json:"user_id" gorm:"column:user_id"`
	Issuer   string `json:"issuer" gorm:"column:issuer"`
	Number   string `json:"number" gorm:"column:number"`
	DummyCol string `json:"dummy_col_10,omitempty" gorm:"column:dummy_col_10"`
}

func (DebitCardDetail) TableName() string {
	return "debit_card_details"
}

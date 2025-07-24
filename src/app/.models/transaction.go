package models

type Transaction struct {
	TransactionID string `json:"transaction_id" gorm:"column:transaction_id;primaryKey"`
	UserID        string `json:"user_id" gorm:"column:user_id"`
	Name          string `json:"name" gorm:"column:name"`
	Image         string `json:"image" gorm:"column:image"`
	IsBank        bool   `json:"isBank" gorm:"column:isBank"`
	DummyCol      string `json:"dummy_col_6,omitempty" gorm:"column:dummy_col_6"`
}

func (Transaction) TableName() string {
	return "transactions"
}

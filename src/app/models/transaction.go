package models

type Transaction struct {
	TransactionID string `gorm:"column:transaction_id;primaryKey"`
	UserID        string `gorm:"column:user_id"`
	Name          string `gorm:"column:name"`
	Image         string `gorm:"column:image"`
	IsBank        bool   `gorm:"column:isBank"`
	DummyCol      string `gorm:"column:dummy_col_6"`
}

func (Transaction) TableName() string {
	return "transactions"
}

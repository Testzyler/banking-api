package entities

type Transaction struct {
	TransactionID string `json:"transactionID"`
	UserID        string `json:"userID"`
	Name          string `json:"name"`
	Image         string `json:"image"`
	IsBank        bool   `json:"isBank"`
}

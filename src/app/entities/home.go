package entities

type HomeParams struct {
	UserID string `json:"userID" validate:"required,min=3,max=50"`
}

type HomeResponse struct {
	User
	DebitCards   []DebitCards  `json:"debitCards"`
	Banners      []Banner      `json:"banners"`
	Transactions []Transaction `json:"transactions"`
	Accounts     []Account     `json:"accounts"`
	TotalBalance float64       `json:"totalBalance"`
}

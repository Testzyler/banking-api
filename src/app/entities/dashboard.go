package entities

import (
	"github.com/Testzyler/banking-api/app/validators"
)

type DashboardParams struct {
	UserID string `json:"userID" validate:"required,min=3,max=50"`
}

func (p *DashboardParams) Validate() error {
	return validators.ValidateStruct(p)
}

type DashboardResponse struct {
	User
	DebitCards   []DebitCards  `json:"debitCards"`
	Banners      []Banner      `json:"banners"`
	Transactions []Transaction `json:"transactions"`
	Accounts     []Account     `json:"accounts"`
	TotalBalance float64       `json:"totalBalance"`
}

type Banner struct {
	BannerID    string `json:"bannerID"`
	UserID      string `json:"userID"`
	Title       string `json:"title"`
	Description string `json:"description"`
	ImageURL    string `json:"imageURL"`
}

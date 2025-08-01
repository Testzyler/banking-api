package entities

type DebitCards struct {
	CardID          string          `json:"cardID"`
	CardNumber      string          `json:"cardNumber"`
	CardName        string          `json:"cardName"`
	Issuer          string          `json:"issuer"`
	Status          string          `json:"status"`
	DebitCardDesign DebitCardDesign `json:"cardDesign"`
}

type DebitCardDesign struct {
	Color       string `json:"color"`
	BorderColor string `json:"borderColor"`
}

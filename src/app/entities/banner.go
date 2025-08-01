package entities

type Banner struct {
	BannerID    string `json:"bannerID"`
	UserID      string `json:"userID"`
	Title       string `json:"title"`
	Description string `json:"description"`
	ImageURL    string `json:"imageURL"`
}

package models

type Banner struct {
	BannerID    string `json:"banner_id" gorm:"column:banner_id;primaryKey"`
	UserID      string `json:"user_id" gorm:"column:user_id"`
	Title       string `json:"title" gorm:"column:title"`
	Description string `json:"description" gorm:"column:description"`
	Image       string `json:"image" gorm:"column:image"`
	DummyCol    string `json:"dummy_col_11,omitempty" gorm:"column:dummy_col_11"`
}

func (Banner) TableName() string {
	return "banners"
}

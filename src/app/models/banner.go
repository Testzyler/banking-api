package models

type Banner struct {
	BannerID    string `gorm:"column:banner_id;primaryKey"`
	UserID      string `gorm:"column:user_id"`
	Title       string `gorm:"column:title"`
	Description string `gorm:"column:description"`
	Image       string `gorm:"column:image"`
	DummyCol    string `gorm:"column:dummy_col_11"`

	User *User `gorm:"foreignKey:UserID;references:UserID" json:"user,omitempty"`
}

func (Banner) TableName() string {
	return "banners"
}

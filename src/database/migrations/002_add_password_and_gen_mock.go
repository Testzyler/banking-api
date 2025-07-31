package migrations

import (
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var addPasswordAndSetDefault = &Migration{
	Number: 2,
	Name:   "add password and set default",

	Forwards: func(db *gorm.DB) error {
		return Migrate_AddPasswordAndSetDefault(db)
	},
}

// "P@ss!w0rd" (encrypted with bcrypt)
func Migrate_AddPasswordAndSetDefault(db *gorm.DB) error {
	type User struct {
		Password string `gorm:"column:password"`
	}
	if !db.Migrator().HasColumn(&User{}, "password") {
		if err := db.Migrator().AddColumn(&User{}, "password"); err != nil {
			return err
		}
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("P@ss!w0rd"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	if err := db.Table("users").Where("password IS NULL OR password = ''").Update("password", string(hashedPassword)).Error; err != nil {
		return err
	}

	return nil
}

func init() {
	Migrations = append(Migrations, addPasswordAndSetDefault)
}

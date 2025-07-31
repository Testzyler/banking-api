package migrations

import (
	"fmt"

	"gorm.io/gorm"
)

var createIndexAndFK = &Migration{
	Number: 4,
	Name:   "create index and fk",

	Forwards: func(db *gorm.DB) error {
		return Migrate_CreateIndexAndFKForUserTable(db)
	},
}

func Migrate_CreateIndexAndFKForUserTable(db *gorm.DB) error {
	tx := db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("could not start transaction: %w", tx.Error)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	tx.Exec("ALTER TABLE accounts ADD CONSTRAINT fk_accounts_user FOREIGN KEY (user_id) REFERENCES users(user_id)")

	return tx.Commit().Error

}

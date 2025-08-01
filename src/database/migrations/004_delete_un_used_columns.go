package migrations

import (
	"github.com/Testzyler/banking-api/logger"
	"gorm.io/gorm"
)

var deleteUnUsedColumns = &Migration{
	Number: 4,
	Name:   "delete un used columns",

	Forwards: func(db *gorm.DB) error {
		return Migrate_DeleteUnUsedColumns(db)
	},
}

func init() {
	Migrations = append(Migrations, deleteUnUsedColumns)
}

func Migrate_DeleteUnUsedColumns(db *gorm.DB) error {

	statements := []string{
		// account_balances
		`ALTER TABLE users DROP COLUMN dummy_col_1;`,
		`ALTER TABLE user_greetings DROP COLUMN dummy_col_2;`,
		`ALTER TABLE accounts DROP COLUMN dummy_col_3;`,
		`ALTER TABLE account_balances DROP COLUMN dummy_col_4;`,
		`ALTER TABLE account_details DROP COLUMN dummy_col_5;`,
		`ALTER TABLE transactions DROP COLUMN dummy_col_6;`,
		`ALTER TABLE debit_cards DROP COLUMN dummy_col_7;`,
		`ALTER TABLE debit_card_status DROP COLUMN dummy_col_8;`,
		`ALTER TABLE debit_card_design DROP COLUMN dummy_col_9;`,
		`ALTER TABLE debit_card_details DROP COLUMN dummy_col_10;`,
		`ALTER TABLE banners DROP COLUMN dummy_col_11;`,
	}

	for _, stmt := range statements {
		if err := db.Exec(stmt).Error; err != nil {
			logger.Errorf("failed to execute statement: %s, error: %w", stmt, err)
		}
		logger.Infof("executed statement: %s", stmt)
	}

	logger.Infof("Index and FK created successfully")
	return nil
}

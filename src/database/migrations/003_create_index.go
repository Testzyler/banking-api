package migrations

import (
	"github.com/Testzyler/banking-api/logger"
	"gorm.io/gorm"
)

var createIndex = &Migration{
	Number: 3,
	Name:   "create index table",

	Forwards: func(db *gorm.DB) error {
		return Migrate_CreateIndex(db)
	},
}

func init() {
	Migrations = append(Migrations, createIndex)
}

func Migrate_CreateIndex(db *gorm.DB) error {

	statements := []string{
		// users
		`CREATE INDEX idx_users_user_id ON users(user_id);`,
		`CREATE INDEX idx_users_name ON users(name);`,

		// accounts
		`CREATE INDEX idx_accounts_user_id ON accounts(user_id);`,

		// account_balances
		`CREATE INDEX idx_account_balances_user_id ON account_balances(user_id);`,
		`CREATE INDEX idx_account_balances_account_id ON account_balances(account_id);`,

		// account_details
		`CREATE INDEX idx_account_details_user_id ON account_details(user_id);`,
		`CREATE INDEX idx_account_details_account_id ON account_details(account_id);`,

		// account_flags
		`CREATE INDEX idx_account_flags_account_id ON account_flags(account_id);`,
		`CREATE INDEX idx_account_flags_user_id ON account_flags(user_id);`,

		// debit_cards
		`CREATE INDEX idx_debit_cards_user_id ON debit_cards(user_id);`,

		// debit_card_details
		`CREATE INDEX idx_debit_card_details_user_id ON debit_card_details(user_id);`,
		`CREATE INDEX idx_debit_card_details_card_id ON debit_card_details(card_id);`,

		// debit_card_design
		`CREATE INDEX idx_debit_card_design_card_id ON debit_card_design(card_id);`,
		`CREATE INDEX idx_debit_card_design_user_id ON debit_card_design(user_id);`,

		// debit_card_status
		`CREATE INDEX idx_debit_card_status_card_id ON debit_card_status(card_id);`,
		`CREATE INDEX idx_debit_card_status_user_id ON debit_card_status(user_id);`,

		// user_greetings
		`CREATE INDEX idx_user_greetings_user_id ON user_greetings(user_id);`,

		// banners
		`CREATE INDEX idx_banners_user_id ON banners(user_id);`,

		// transactions
		`CREATE INDEX idx_transactions_user_id ON transactions(user_id);`,
	}

	for _, stmt := range statements {
		if err := db.Exec(stmt).Error; err != nil {
			logger.Errorf("failed to execute statement: %s, error: %w", stmt, err)
		}
	}

	logger.Infof("Index created successfully")
	return nil
}

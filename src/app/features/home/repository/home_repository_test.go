package repository

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestDashboardRepository_GetTotalBalance(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		mockSetup     func(sqlmock.Sqlmock)
		expectedTotal float64
	}{
		{
			name:   "successful get total balance",
			userID: "user123",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"SUM(amount)"}).
					AddRow(15000.50)

				mock.ExpectQuery("SELECT SUM\\(amount\\) FROM `account_balances` JOIN accounts ON account_balances\\.account_id = accounts\\.account_id WHERE accounts\\.user_id = \\?").
					WithArgs("user123").
					WillReturnRows(rows)
			},
			expectedTotal: 15000.50,
		},
		{
			name:   "user with no accounts returns zero",
			userID: "user123",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT SUM\\(amount\\) FROM `account_balances` JOIN accounts ON account_balances\\.account_id = accounts\\.account_id WHERE accounts\\.user_id = \\?").
					WithArgs("user123").
					WillReturnError(gorm.ErrRecordNotFound)
			},
			expectedTotal: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock database
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()

			gormDB, err := gorm.Open(mysql.New(mysql.Config{
				Conn:                      db,
				SkipInitializeWithVersion: true,
			}), &gorm.Config{})
			assert.NoError(t, err)

			repo := NewHomeRepository(gormDB)

			// Setup mock expectations
			tt.mockSetup(mock)

			// Act
			total := repo.GetTotalBalance(tt.userID)

			// Assert
			assert.Equal(t, tt.expectedTotal, total)

			// Verify all expectations were met
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

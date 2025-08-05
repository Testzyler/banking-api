package repository

import (
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Testzyler/banking-api/app/entities"
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

func TestHomeRepository_GetTotalBalance_AdvancedCases(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		mockSetup     func(sqlmock.Sqlmock)
		expectedTotal float64
	}{
		{
			name:   "user with negative balance",
			userID: "user123",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"SUM(amount)"}).
					AddRow(-500.25)

				mock.ExpectQuery("SELECT SUM\\(amount\\) FROM `account_balances` JOIN accounts ON account_balances\\.account_id = accounts\\.account_id WHERE accounts\\.user_id = \\?").
					WithArgs("user123").
					WillReturnRows(rows)
			},
			expectedTotal: -500.25,
		},
		{
			name:   "user with zero balance",
			userID: "user123",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"SUM(amount)"}).
					AddRow(0.0)

				mock.ExpectQuery("SELECT SUM\\(amount\\) FROM `account_balances` JOIN accounts ON account_balances\\.account_id = accounts\\.account_id WHERE accounts\\.user_id = \\?").
					WithArgs("user123").
					WillReturnRows(rows)
			},
			expectedTotal: 0.0,
		},
		{
			name:   "empty userID",
			userID: "",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"SUM(amount)"}).
					AddRow(0.0)

				mock.ExpectQuery("SELECT SUM\\(amount\\) FROM `account_balances` JOIN accounts ON account_balances\\.account_id = accounts\\.account_id WHERE accounts\\.user_id = \\?").
					WithArgs("").
					WillReturnRows(rows)
			},
			expectedTotal: 0.0,
		},
		{
			name:   "very large balance",
			userID: "user123",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"SUM(amount)"}).
					AddRow(999999999.99)

				mock.ExpectQuery("SELECT SUM\\(amount\\) FROM `account_balances` JOIN accounts ON account_balances\\.account_id = accounts\\.account_id WHERE accounts\\.user_id = \\?").
					WithArgs("user123").
					WillReturnRows(rows)
			},
			expectedTotal: 999999999.99,
		},
		{
			name:   "database connection error",
			userID: "user123",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT SUM\\(amount\\) FROM `account_balances` JOIN accounts ON account_balances\\.account_id = accounts\\.account_id WHERE accounts\\.user_id = \\?").
					WithArgs("user123").
					WillReturnError(errors.New("connection lost"))
			},
			expectedTotal: 0,
		},
		{
			name:   "query timeout error",
			userID: "user123",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT SUM\\(amount\\) FROM `account_balances` JOIN accounts ON account_balances\\.account_id = accounts\\.account_id WHERE accounts\\.user_id = \\?").
					WithArgs("user123").
					WillReturnError(errors.New("query timeout"))
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

func TestHomeRepository_GetHomeData_Success(t *testing.T) {
	// Setup
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	assert.NoError(t, err)

	repo := &homeRepository{db: gormDB}

	// Mock successful flow with minimal data
	mock.ExpectBegin()

	// Mock user query
	userRows := sqlmock.NewRows([]string{"user_id", "name"}).
		AddRow("test123", "Test User")
	mock.ExpectQuery("SELECT \\* FROM `users` WHERE user_id = \\? ORDER BY `users`.`user_id` LIMIT \\?").
		WithArgs("test123", 1).
		WillReturnRows(userRows)

	// Mock UserGreeting - return empty result
	greetingRows := sqlmock.NewRows([]string{"user_id", "greeting"})
	mock.ExpectQuery("SELECT \\* FROM `user_greetings` WHERE `user_greetings`.`user_id` = \\?").
		WithArgs("test123").
		WillReturnRows(greetingRows)

	// Mock empty results for other entities (no data found)
	cardRows := sqlmock.NewRows([]string{"card_id", "user_id", "name"})
	mock.ExpectQuery("SELECT \\* FROM `debit_cards` WHERE user_id = \\? ORDER BY name ASC").
		WithArgs("test123").
		WillReturnRows(cardRows)

	bannerRows := sqlmock.NewRows([]string{"banner_id", "user_id", "title"})
	mock.ExpectQuery("SELECT \\* FROM `banners` WHERE user_id = \\? ORDER BY banner_id ASC").
		WithArgs("test123").
		WillReturnRows(bannerRows)

	txnRows := sqlmock.NewRows([]string{"transaction_id", "user_id", "name"})
	mock.ExpectQuery("SELECT \\* FROM `transactions` WHERE user_id = \\? ORDER BY transaction_id ASC").
		WithArgs("test123").
		WillReturnRows(txnRows)

	accountRows := sqlmock.NewRows([]string{"account_id", "user_id", "type"})
	mock.ExpectQuery("SELECT `accounts`\\.`account_id`,`accounts`\\.`user_id`,`accounts`\\.`type`,`accounts`\\.`currency`,`accounts`\\.`account_number`,`accounts`\\.`issuer` FROM `accounts` JOIN account_details ON accounts\\.account_id = account_details\\.account_id WHERE accounts\\.user_id = \\? ORDER BY account_details\\.is_main_account DESC, accounts\\.type ASC").
		WithArgs("test123").
		WillReturnRows(accountRows)

	mock.ExpectCommit()

	// Act
	response, err := repo.GetHomeData("test123")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "test123", response.UserID)
	assert.Equal(t, "Test User", response.Name)
	assert.Equal(t, "", response.Greeting)      // No greeting found
	assert.Equal(t, 0.0, response.TotalBalance) // No accounts
	assert.Empty(t, response.DebitCards)
	assert.Empty(t, response.Banners)
	assert.Empty(t, response.Transactions)
	assert.Empty(t, response.Accounts)

	// Verify mock expectations
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHomeRepository_GetHomeData_ErrorCases(t *testing.T) {
	tests := []struct {
		name        string
		userID      string
		mockSetup   func(sqlmock.Sqlmock)
		expectError bool
		expectData  func(*testing.T, entities.HomeResponse)
	}{
		{
			name:   "user not found error",
			userID: "nonexistent",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectQuery("SELECT \\* FROM `users` WHERE user_id = \\? ORDER BY `users`.`user_id` LIMIT \\?").
					WithArgs("nonexistent", 1).
					WillReturnError(gorm.ErrRecordNotFound)
				mock.ExpectRollback()
			},
			expectError: true,
		},
		{
			name:   "database connection error during user fetch",
			userID: "user123",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectQuery("SELECT \\* FROM `users` WHERE user_id = \\? ORDER BY `users`.`user_id` LIMIT \\?").
					WithArgs("user123", 1).
					WillReturnError(errors.New("connection lost"))
				mock.ExpectRollback()
			},
			expectError: true,
		},
		{
			name:   "empty userID causes error",
			userID: "",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectQuery("SELECT \\* FROM `users` WHERE user_id = \\? ORDER BY `users`.`user_id` LIMIT \\?").
					WithArgs("", 1).
					WillReturnError(gorm.ErrRecordNotFound)
				mock.ExpectRollback()
			},
			expectError: true,
		},
		{
			name:   "database timeout error",
			userID: "user123",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectQuery("SELECT \\* FROM `users` WHERE user_id = \\? ORDER BY `users`.`user_id` LIMIT \\?").
					WithArgs("user123", 1).
					WillReturnError(errors.New("context deadline exceeded"))
				mock.ExpectRollback()
			},
			expectError: true,
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
			_, err = repo.GetHomeData(tt.userID)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Verify all expectations were met
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

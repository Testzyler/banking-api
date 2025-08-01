package repository

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Testzyler/banking-api/app/entities"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestDashboardRepository_GetUserByID(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		mockSetup     func(sqlmock.Sqlmock)
		expectError   bool
		expectUser    *entities.User
		errorContains string
	}{
		{
			name:   "successful get user",
			userID: "user123",
			mockSetup: func(mock sqlmock.Sqlmock) {
				userRows := sqlmock.NewRows([]string{"user_id", "name", "password"}).
					AddRow("user123", "testuser", "hashedpassword")

				greetingRows := sqlmock.NewRows([]string{"user_id", "greeting"}).
					AddRow("user123", "Hello, testuser!")

				mock.ExpectQuery("SELECT \\* FROM `users` WHERE user_id = \\? ORDER BY `users`.`user_id` LIMIT \\?").
					WithArgs("user123", 1).
					WillReturnRows(userRows)

				mock.ExpectQuery("SELECT \\* FROM `user_greetings` WHERE `user_greetings`.`user_id` = \\?").
					WithArgs("user123").
					WillReturnRows(greetingRows)
			},
			expectError: false,
			expectUser: &entities.User{
				UserID:   "user123",
				Name:     "testuser",
				Greeting: "Hello, testuser!",
			},
		},
		{
			name:   "user not found",
			userID: "nonexistent",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM `users` WHERE user_id = \\? ORDER BY `users`.`user_id` LIMIT \\?").
					WithArgs("nonexistent", 1).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			expectError:   true,
			errorContains: "record not found",
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

			repo := NewDashboardRepository(gormDB)

			// Setup mock expectations
			tt.mockSetup(mock)

			// Act
			user, err := repo.GetUserByID(tt.userID)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectUser.UserID, user.UserID)
				assert.Equal(t, tt.expectUser.Name, user.Name)
				assert.Equal(t, tt.expectUser.Greeting, user.Greeting)
			}

			// Verify all expectations were met
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

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

			repo := NewDashboardRepository(gormDB)

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

func TestDashboardRepository_GetBannersByUserID(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		mockSetup     func(sqlmock.Sqlmock)
		expectError   bool
		expectCount   int
		errorContains string
	}{
		{
			name:   "successful get banners",
			userID: "user123",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"banner_id", "user_id", "title", "description", "image"}).
					AddRow("banner1", "user123", "Welcome Banner", "Welcome to our app", "banner1.jpg").
					AddRow("banner2", "user123", "Promo Banner", "Special promotion", "banner2.jpg")

				mock.ExpectQuery("SELECT \\* FROM `banners` WHERE user_id = \\?").
					WithArgs("user123").
					WillReturnRows(rows)
			},
			expectError: false,
			expectCount: 2,
		},
		{
			name:   "no banners found",
			userID: "user123",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"banner_id", "user_id", "title", "description", "image"})

				mock.ExpectQuery("SELECT \\* FROM `banners` WHERE user_id = \\?").
					WithArgs("user123").
					WillReturnRows(rows)
			},
			expectError: false,
			expectCount: 0,
		},
		{
			name:   "database error",
			userID: "user123",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM `banners` WHERE user_id = \\?").
					WithArgs("user123").
					WillReturnError(gorm.ErrInvalidDB)
			},
			expectError:   true,
			expectCount:   0,
			errorContains: "invalid db",
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

			repo := NewDashboardRepository(gormDB)

			// Setup mock expectations
			tt.mockSetup(mock)

			// Act
			banners, err := repo.GetBannersByUserID(tt.userID)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, banners)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.Len(t, banners, tt.expectCount)
				if tt.expectCount > 0 {
					assert.Equal(t, "user123", banners[0].UserID)
					assert.NotEmpty(t, banners[0].Title)
				}
			}

			// Verify all expectations were met
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestDashboardRepository_GetTransactionsByUserID(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		mockSetup     func(sqlmock.Sqlmock)
		expectError   bool
		expectCount   int
		errorContains string
	}{
		{
			name:   "successful get transactions",
			userID: "user123",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"transaction_id", "user_id", "name", "image", "is_bank"}).
					AddRow("txn1", "user123", "Coffee Shop", "coffee.jpg", false).
					AddRow("txn2", "user123", "Bank Transfer", "bank.jpg", true)

				mock.ExpectQuery("SELECT \\* FROM `transactions` WHERE user_id = \\?").
					WithArgs("user123").
					WillReturnRows(rows)
			},
			expectError: false,
			expectCount: 2,
		},
		{
			name:   "no transactions found",
			userID: "user123",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"transaction_id", "user_id", "name", "image", "is_bank"})

				mock.ExpectQuery("SELECT \\* FROM `transactions` WHERE user_id = \\?").
					WithArgs("user123").
					WillReturnRows(rows)
			},
			expectError: false,
			expectCount: 0,
		},
		{
			name:   "database error",
			userID: "user123",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM `transactions` WHERE user_id = \\?").
					WithArgs("user123").
					WillReturnError(gorm.ErrInvalidDB)
			},
			expectError:   true,
			expectCount:   0,
			errorContains: "invalid db",
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

			repo := NewDashboardRepository(gormDB)

			// Setup mock expectations
			tt.mockSetup(mock)

			// Act
			transactions, err := repo.GetTransactionsByUserID(tt.userID)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, transactions)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.Len(t, transactions, tt.expectCount)
				if tt.expectCount > 0 {
					assert.Equal(t, "user123", transactions[0].UserID)
					assert.NotEmpty(t, transactions[0].Name)
				}
			}

			// Verify all expectations were met
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

package repository

import (
	"database/sql"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	models "github.com/Testzyler/banking-api/app/.models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	require.NoError(t, err)

	return gormDB, mock, func() {
		db.Close()
	}
}

func TestUserRepository_GetByID_TableDriven(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		mockSetup     func(sqlmock.Sqlmock, string)
		expectError   bool
		expectUser    *models.User
		errorContains string
	}{
		{
			name:   "valid user ID - success",
			userID: "user123",
			mockSetup: func(mock sqlmock.Sqlmock, userID string) {
				rows := sqlmock.NewRows([]string{"user_id", "name", "dummy_col_1"}).
					AddRow("user123", "John Doe", "test_data")
				mock.ExpectQuery("SELECT \\* FROM `users` WHERE user_id = \\? ORDER BY `users`.`user_id` LIMIT \\?").
					WithArgs(userID, 1).
					WillReturnRows(rows)
			},
			expectError: false,
			expectUser: &models.User{
				UserID:   "user123",
				Name:     "John Doe",
				DummyCol: "test_data",
			},
		},
		{
			name:   "empty user ID",
			userID: "",
			mockSetup: func(mock sqlmock.Sqlmock, userID string) {
				mock.ExpectQuery("SELECT \\* FROM `users` WHERE user_id = \\? ORDER BY `users`.`user_id` LIMIT \\?").
					WithArgs(userID, 1).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			expectError:   true,
			expectUser:    nil,
			errorContains: "user not found",
		},
		{
			name:   "whitespace user ID",
			userID: "   ",
			mockSetup: func(mock sqlmock.Sqlmock, userID string) {
				mock.ExpectQuery("SELECT \\* FROM `users` WHERE user_id = \\? ORDER BY `users`.`user_id` LIMIT \\?").
					WithArgs(userID, 1).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			expectError:   true,
			expectUser:    nil,
			errorContains: "user not found",
		},
		{
			name:   "nonexistent user ID",
			userID: "nonexistent",
			mockSetup: func(mock sqlmock.Sqlmock, userID string) {
				mock.ExpectQuery("SELECT \\* FROM `users` WHERE user_id = \\? ORDER BY `users`.`user_id` LIMIT \\?").
					WithArgs(userID, 1).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			expectError:   true,
			expectUser:    nil,
			errorContains: "user not found",
		},
		{
			name:   "database connection error",
			userID: "user123",
			mockSetup: func(mock sqlmock.Sqlmock, userID string) {
				mock.ExpectQuery("SELECT \\* FROM `users` WHERE user_id = \\? ORDER BY `users`.`user_id` LIMIT \\?").
					WithArgs(userID, 1).
					WillReturnError(sql.ErrConnDone)
			},
			expectError:   true,
			expectUser:    nil,
			errorContains: "error fetching user",
		},
		{
			name:   "SQL injection attempt",
			userID: "'; DROP TABLE users; --",
			mockSetup: func(mock sqlmock.Sqlmock, userID string) {
				// GORM should escape this properly, so it will just search for the literal string
				mock.ExpectQuery("SELECT \\* FROM `users` WHERE user_id = \\? ORDER BY `users`.`user_id` LIMIT \\?").
					WithArgs(userID, 1).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			expectError:   true,
			expectUser:    nil,
			errorContains: "user not found",
		},
		{
			name:   "special characters in user ID",
			userID: "user@#$%^&*()",
			mockSetup: func(mock sqlmock.Sqlmock, userID string) {
				mock.ExpectQuery("SELECT \\* FROM `users` WHERE user_id = \\? ORDER BY `users`.`user_id` LIMIT \\?").
					WithArgs(userID, 1).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			expectError:   true,
			expectUser:    nil,
			errorContains: "user not found",
		},
		{
			name:   "unicode user ID - success",
			userID: "ผู้ใช้123",
			mockSetup: func(mock sqlmock.Sqlmock, userID string) {
				rows := sqlmock.NewRows([]string{"user_id", "name", "dummy_col_1"}).
					AddRow("ผู้ใช้123", "Thai User", "thai_data")
				mock.ExpectQuery("SELECT \\* FROM `users` WHERE user_id = \\? ORDER BY `users`.`user_id` LIMIT \\?").
					WithArgs(userID, 1).
					WillReturnRows(rows)
			},
			expectError: false,
			expectUser: &models.User{
				UserID:   "ผู้ใช้123",
				Name:     "Thai User",
				DummyCol: "thai_data",
			},
		},
		{
			name:   "very long user ID",
			userID: strings.Repeat("a", 1000),
			mockSetup: func(mock sqlmock.Sqlmock, userID string) {
				mock.ExpectQuery("SELECT \\* FROM `users` WHERE user_id = \\? ORDER BY `users`.`user_id` LIMIT \\?").
					WithArgs(userID, 1).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			expectError:   true,
			expectUser:    nil,
			errorContains: "user not found",
		},
		{
			name:   "user ID with newlines and tabs",
			userID: "user\n\t123",
			mockSetup: func(mock sqlmock.Sqlmock, userID string) {
				mock.ExpectQuery("SELECT \\* FROM `users` WHERE user_id = \\? ORDER BY `users`.`user_id` LIMIT \\?").
					WithArgs(userID, 1).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			expectError:   true,
			expectUser:    nil,
			errorContains: "user not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			gormDB, mock, cleanup := setupTestDB(t)
			defer cleanup()

			repo := NewUserRepository(gormDB)

			// Setup mock expectations
			tt.mockSetup(mock, tt.userID)

			// Act
			user, err := repo.GetByID(tt.userID)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, user)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				if tt.expectUser != nil {
					assert.Equal(t, tt.expectUser.UserID, user.UserID)
					assert.Equal(t, tt.expectUser.Name, user.Name)
					assert.Equal(t, tt.expectUser.DummyCol, user.DummyCol)
				}
			}

			// Verify mock expectations
			err = mock.ExpectationsWereMet()
			assert.NoError(t, err)
		})
	}
}

func TestUserRepository_GetAll_TableDriven(t *testing.T) {
	tests := []struct {
		name          string
		perPage       int
		page          int
		search        string
		mockSetup     func(sqlmock.Sqlmock, int, int, string)
		expectError   bool
		expectUsers   []*models.User
		expectCount   int
		errorContains string
	}{
		{
			name:    "valid pagination - first page",
			perPage: 10,
			page:    1,
			search:  "",
			mockSetup: func(mock sqlmock.Sqlmock, perPage, page int, search string) {
				rows := sqlmock.NewRows([]string{"user_id", "name", "dummy_col_1"}).
					AddRow("user1", "Alice", "data1").
					AddRow("user2", "Bob", "data2")
				mock.ExpectQuery("SELECT \\* FROM `users` ORDER BY name ASC LIMIT \\?").
					WithArgs(perPage).
					WillReturnRows(rows)
			},
			expectError: false,
			expectCount: 2,
		},
		{
			name:    "empty results",
			perPage: 10,
			page:    1,
			search:  "",
			mockSetup: func(mock sqlmock.Sqlmock, perPage, page int, search string) {
				rows := sqlmock.NewRows([]string{"user_id", "name", "dummy_col_1"})
				mock.ExpectQuery("SELECT \\* FROM `users` ORDER BY name ASC LIMIT \\?").
					WithArgs(perPage).
					WillReturnRows(rows)
			},
			expectError: false,
			expectCount: 0,
		},
		{
			name:    "database connection error",
			perPage: 10,
			page:    1,
			search:  "",
			mockSetup: func(mock sqlmock.Sqlmock, perPage, page int, search string) {
				mock.ExpectQuery("SELECT \\* FROM `users` ORDER BY name ASC LIMIT \\?").
					WithArgs(perPage).
					WillReturnError(sql.ErrConnDone)
			},
			expectError:   true,
			expectCount:   0,
			errorContains: "error fetching users",
		},
		{
			name:    "large page size",
			perPage: 1000,
			page:    1,
			search:  "",
			mockSetup: func(mock sqlmock.Sqlmock, perPage, page int, search string) {
				rows := sqlmock.NewRows([]string{"user_id", "name", "dummy_col_1"})
				// Simulate 1000 users
				for i := 1; i <= 1000; i++ {
					rows.AddRow("user"+string(rune(i)), "User "+string(rune(i)), "data"+string(rune(i)))
				}
				mock.ExpectQuery("SELECT \\* FROM `users` ORDER BY name ASC LIMIT \\?").
					WithArgs(perPage).
					WillReturnRows(rows)
			},
			expectError: false,
			expectCount: 1000,
		},
		{
			name:    "zero per page",
			perPage: 0,
			page:    1,
			search:  "",
			mockSetup: func(mock sqlmock.Sqlmock, perPage, page int, search string) {
				rows := sqlmock.NewRows([]string{"user_id", "name", "dummy_col_1"})
				mock.ExpectQuery("SELECT \\* FROM `users` ORDER BY name ASC LIMIT \\?").
					WithArgs(perPage).
					WillReturnRows(rows)
			},
			expectError: false,
			expectCount: 0,
		},
		{
			name:    "negative page number",
			perPage: 10,
			page:    -1,
			search:  "",
			mockSetup: func(mock sqlmock.Sqlmock, perPage, page int, search string) {
				rows := sqlmock.NewRows([]string{"user_id", "name", "dummy_col_1"}).
					AddRow("user1", "Alice", "data1").
					AddRow("user2", "Bob", "data2")
				mock.ExpectQuery("SELECT \\* FROM `users` ORDER BY name ASC LIMIT \\?").
					WithArgs(perPage).
					WillReturnRows(rows)
			},
			expectError: false,
			expectCount: 2,
		},
		{
			name:    "second page pagination",
			perPage: 5,
			page:    2,
			search:  "",
			mockSetup: func(mock sqlmock.Sqlmock, perPage, page int, search string) {
				rows := sqlmock.NewRows([]string{"user_id", "name", "dummy_col_1"}).
					AddRow("user6", "Frank", "data6").
					AddRow("user7", "Grace", "data7")
				mock.ExpectQuery("SELECT \\* FROM `users` ORDER BY name ASC LIMIT \\? OFFSET \\?").
					WithArgs(perPage, (page-1)*perPage).
					WillReturnRows(rows)
			},
			expectError: false,
			expectCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			gormDB, mock, cleanup := setupTestDB(t)
			defer cleanup()

			repo := NewUserRepository(gormDB)

			// Setup mock expectations
			tt.mockSetup(mock, tt.perPage, tt.page, tt.search)

			// Act
			users, err := repo.GetAll(tt.perPage, tt.page, tt.search)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, users)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, users)
				assert.Len(t, users, tt.expectCount)
			}

			// Verify mock expectations
			err = mock.ExpectationsWereMet()
			assert.NoError(t, err)
		})
	}
}

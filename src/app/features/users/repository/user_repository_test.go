package repository

import (
	"database/sql"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	models "github.com/Testzyler/banking-api/app/models"
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
				rows := sqlmock.NewRows([]string{"user_id", "name"}).
					AddRow("user123", "John Doe")
				mock.ExpectQuery("SELECT \\* FROM `users` WHERE user_id = \\? ORDER BY `users`.`user_id` LIMIT \\?").
					WithArgs(userID, 1).
					WillReturnRows(rows)
			},
			expectError: false,
			expectUser: &models.User{
				UserID:   "user123",
				Name:     "John Doe",
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
			errorContains: "record not found",
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
			errorContains: "record not found",
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
			errorContains: "record not found",
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
			errorContains: "sql: connection is already closed",
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
			errorContains: "record not found",
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
			errorContains: "record not found",
		},
		{
			name:   "unicode user ID - success",
			userID: "ผู้ใช้123",
			mockSetup: func(mock sqlmock.Sqlmock, userID string) {
				rows := sqlmock.NewRows([]string{"user_id", "name"}).
					AddRow("ผู้ใช้123", "Thai User")
				mock.ExpectQuery("SELECT \\* FROM `users` WHERE user_id = \\? ORDER BY `users`.`user_id` LIMIT \\?").
					WithArgs(userID, 1).
					WillReturnRows(rows)
			},
			expectError: false,
			expectUser: &models.User{
				UserID:   "ผู้ใช้123",
				Name:     "Thai User",
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
			errorContains: "record not found",
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
			errorContains: "record not found",
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
				rows := sqlmock.NewRows([]string{"user_id", "name"}).
					AddRow("user1", "Alice").
					AddRow("user2", "Bob")
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
				rows := sqlmock.NewRows([]string{"user_id", "name"})
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
			errorContains: "sql: connection is already closed",
		},
		{
			name:    "large page size",
			perPage: 1000,
			page:    1,
			search:  "",
			mockSetup: func(mock sqlmock.Sqlmock, perPage, page int, search string) {
				rows := sqlmock.NewRows([]string{"user_id", "name"})
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
				rows := sqlmock.NewRows([]string{"user_id", "name"})
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
				rows := sqlmock.NewRows([]string{"user_id", "name"}).
					AddRow("user1", "Alice").
					AddRow("user2", "Bob")
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
				rows := sqlmock.NewRows([]string{"user_id", "name"}).
					AddRow("user6", "Frank").
					AddRow("user7", "Grace")
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

func TestUserRepository_GetAllWithCount_TableDriven(t *testing.T) {
	tests := []struct {
		name          string
		perPage       int
		page          int
		search        string
		mockSetup     func(sqlmock.Sqlmock, int, int, string)
		expectError   bool
		expectUsers   []*models.User
		expectCount   int64
		errorContains string
	}{
		{
			name:    "valid pagination - first page with no search",
			perPage: 10,
			page:    1,
			search:  "",
			mockSetup: func(mock sqlmock.Sqlmock, perPage, page int, search string) {
				// Mock data query
				rows := sqlmock.NewRows([]string{"user_id", "name"}).
					AddRow("user1", "Alice").
					AddRow("user2", "Bob").
					AddRow("user3", "Charlie")
				mock.ExpectQuery("SELECT \\* FROM `users` ORDER BY name ASC LIMIT \\?").
					WithArgs(perPage).
					WillReturnRows(rows)

				// Mock count query
				countRows := sqlmock.NewRows([]string{"count"}).AddRow(15)
				mock.ExpectQuery("SELECT count\\(\\*\\) FROM `users`").
					WillReturnRows(countRows)
			},
			expectError: false,
			expectUsers: []*models.User{
				{UserID: "user1", Name: "Alice"},
				{UserID: "user2", Name: "Bob"},
				{UserID: "user3", Name: "Charlie"},
			},
			expectCount: 15,
		},
		{
			name:    "valid pagination - second page with no search",
			perPage: 5,
			page:    2,
			search:  "",
			mockSetup: func(mock sqlmock.Sqlmock, perPage, page int, search string) {
				// Mock data query with offset
				rows := sqlmock.NewRows([]string{"user_id", "name"}).
					AddRow("user6", "Frank").
					AddRow("user7", "Grace")
				mock.ExpectQuery("SELECT \\* FROM `users` ORDER BY name ASC LIMIT \\? OFFSET \\?").
					WithArgs(perPage, 5).
					WillReturnRows(rows)

				// Mock count query
				countRows := sqlmock.NewRows([]string{"count"}).AddRow(12)
				mock.ExpectQuery("SELECT count\\(\\*\\) FROM `users`").
					WillReturnRows(countRows)
			},
			expectError: false,
			expectUsers: []*models.User{
				{UserID: "user6", Name: "Frank"},
				{UserID: "user7", Name: "Grace"},
			},
			expectCount: 12,
		},
		{
			name:    "search functionality - matching results",
			perPage: 10,
			page:    1,
			search:  "Admin",
			mockSetup: func(mock sqlmock.Sqlmock, perPage, page int, search string) {
				// Mock data query with search
				rows := sqlmock.NewRows([]string{"user_id", "name"}).
					AddRow("admin1", "Admin User").
					AddRow("admin2", "Super Admin")
				mock.ExpectQuery("SELECT \\* FROM `users` WHERE name LIKE \\? ORDER BY name ASC LIMIT \\?").
					WithArgs("%Admin%", perPage).
					WillReturnRows(rows)

				// Mock count query with search
				countRows := sqlmock.NewRows([]string{"count"}).AddRow(2)
				mock.ExpectQuery("SELECT count\\(\\*\\) FROM `users` WHERE name LIKE \\?").
					WithArgs("%Admin%").
					WillReturnRows(countRows)
			},
			expectError: false,
			expectUsers: []*models.User{
				{UserID: "admin1", Name: "Admin User"},
				{UserID: "admin2", Name: "Super Admin"},
			},
			expectCount: 2,
		},
		{
			name:    "search functionality - no matching results",
			perPage: 10,
			page:    1,
			search:  "nonexistent",
			mockSetup: func(mock sqlmock.Sqlmock, perPage, page int, search string) {
				// Mock empty data query
				rows := sqlmock.NewRows([]string{"user_id", "name"})
				mock.ExpectQuery("SELECT \\* FROM `users` WHERE name LIKE \\? ORDER BY name ASC LIMIT \\?").
					WithArgs("%nonexistent%", perPage).
					WillReturnRows(rows)

				// Mock count query with no results
				countRows := sqlmock.NewRows([]string{"count"}).AddRow(0)
				mock.ExpectQuery("SELECT count\\(\\*\\) FROM `users` WHERE name LIKE \\?").
					WithArgs("%nonexistent%").
					WillReturnRows(countRows)
			},
			expectError: false,
			expectUsers: []*models.User{},
			expectCount: 0,
		},
		{
			name:    "empty search string behavior",
			perPage: 5,
			page:    1,
			search:  "",
			mockSetup: func(mock sqlmock.Sqlmock, perPage, page int, search string) {
				// Mock data query without search
				rows := sqlmock.NewRows([]string{"user_id", "name"}).
					AddRow("user1", "Alice")
				mock.ExpectQuery("SELECT \\* FROM `users` ORDER BY name ASC LIMIT \\?").
					WithArgs(perPage).
					WillReturnRows(rows)

				// Mock count query
				countRows := sqlmock.NewRows([]string{"count"}).AddRow(10)
				mock.ExpectQuery("SELECT count\\(\\*\\) FROM `users`").
					WillReturnRows(countRows)
			},
			expectError: false,
			expectUsers: []*models.User{
				{UserID: "user1", Name: "Alice"},
			},
			expectCount: 10,
		},
		{
			name:    "zero per page",
			perPage: 0,
			page:    1,
			search:  "",
			mockSetup: func(mock sqlmock.Sqlmock, perPage, page int, search string) {
				// Mock data query with 0 limit
				rows := sqlmock.NewRows([]string{"user_id", "name"})
				mock.ExpectQuery("SELECT \\* FROM `users` ORDER BY name ASC LIMIT \\?").
					WithArgs(0).
					WillReturnRows(rows)

				// Mock count query
				countRows := sqlmock.NewRows([]string{"count"}).AddRow(10)
				mock.ExpectQuery("SELECT count\\(\\*\\) FROM `users`").
					WillReturnRows(countRows)
			},
			expectError: false,
			expectUsers: []*models.User{},
			expectCount: 10,
		},
		{
			name:    "negative page number handled",
			perPage: 10,
			page:    -1,
			search:  "",
			mockSetup: func(mock sqlmock.Sqlmock, perPage, page int, search string) {
				// Mock data query with offset 0 (negative page handled)
				rows := sqlmock.NewRows([]string{"user_id", "name"}).
					AddRow("user1", "Alice")
				mock.ExpectQuery("SELECT \\* FROM `users` ORDER BY name ASC LIMIT \\?").
					WithArgs(perPage).
					WillReturnRows(rows)

				// Mock count query
				countRows := sqlmock.NewRows([]string{"count"}).AddRow(5)
				mock.ExpectQuery("SELECT count\\(\\*\\) FROM `users`").
					WillReturnRows(countRows)
			},
			expectError: false,
			expectUsers: []*models.User{
				{UserID: "user1", Name: "Alice"},
			},
			expectCount: 5,
		},
		{
			name:    "database error on data query",
			perPage: 10,
			page:    1,
			search:  "",
			mockSetup: func(mock sqlmock.Sqlmock, perPage, page int, search string) {
				// Mock data query with error
				mock.ExpectQuery("SELECT \\* FROM `users` ORDER BY name ASC LIMIT \\?").
					WithArgs(perPage).
					WillReturnError(sql.ErrConnDone)
			},
			expectError:   true,
			expectUsers:   nil,
			expectCount:   0,
			errorContains: "sql: connection is already closed",
		},
		{
			name:    "database error on count query",
			perPage: 10,
			page:    1,
			search:  "",
			mockSetup: func(mock sqlmock.Sqlmock, perPage, page int, search string) {
				// Mock successful data query
				rows := sqlmock.NewRows([]string{"user_id", "name"}).
					AddRow("user1", "Alice")
				mock.ExpectQuery("SELECT \\* FROM `users` ORDER BY name ASC LIMIT \\?").
					WithArgs(perPage).
					WillReturnRows(rows)

				// Mock count query with error
				mock.ExpectQuery("SELECT count\\(\\*\\) FROM `users`").
					WillReturnError(sql.ErrConnDone)
			},
			expectError:   true,
			expectUsers:   nil,
			expectCount:   0,
			errorContains: "sql: connection is already closed",
		},
		{
			name:    "search with special characters",
			perPage: 10,
			page:    1,
			search:  "user@#$%",
			mockSetup: func(mock sqlmock.Sqlmock, perPage, page int, search string) {
				// Mock data query with special character search
				rows := sqlmock.NewRows([]string{"user_id", "name"})
				mock.ExpectQuery("SELECT \\* FROM `users` WHERE name LIKE \\? ORDER BY name ASC LIMIT \\?").
					WithArgs("%user@#$%%", perPage).
					WillReturnRows(rows)

				// Mock count query
				countRows := sqlmock.NewRows([]string{"count"}).AddRow(0)
				mock.ExpectQuery("SELECT count\\(\\*\\) FROM `users` WHERE name LIKE \\?").
					WithArgs("%user@#$%%").
					WillReturnRows(countRows)
			},
			expectError: false,
			expectUsers: []*models.User{},
			expectCount: 0,
		},
		{
			name:    "unicode search",
			perPage: 10,
			page:    1,
			search:  "ผู้ใช้",
			mockSetup: func(mock sqlmock.Sqlmock, perPage, page int, search string) {
				// Mock data query with unicode search
				rows := sqlmock.NewRows([]string{"user_id", "name"}).
					AddRow("ผู้ใช้123", "Thai User")
				mock.ExpectQuery("SELECT \\* FROM `users` WHERE name LIKE \\? ORDER BY name ASC LIMIT \\?").
					WithArgs("%ผู้ใช้%", perPage).
					WillReturnRows(rows)

				// Mock count query
				countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)
				mock.ExpectQuery("SELECT count\\(\\*\\) FROM `users` WHERE name LIKE \\?").
					WithArgs("%ผู้ใช้%").
					WillReturnRows(countRows)
			},
			expectError: false,
			expectUsers: []*models.User{
				{UserID: "ผู้ใช้123", Name: "Thai User"},
			},
			expectCount: 1,
		},
		{
			name:    "large page size",
			perPage: 1000,
			page:    1,
			search:  "",
			mockSetup: func(mock sqlmock.Sqlmock, perPage, page int, search string) {
				// Mock data query with large limit - return just 3 users for testing
				rows := sqlmock.NewRows([]string{"user_id", "name"}).
					AddRow("user1", "User 1").
					AddRow("user2", "User 2").
					AddRow("user3", "User 3")
				mock.ExpectQuery("SELECT \\* FROM `users` ORDER BY name ASC LIMIT \\?").
					WithArgs(1000).
					WillReturnRows(rows)

				// Mock count query
				countRows := sqlmock.NewRows([]string{"count"}).AddRow(500)
				mock.ExpectQuery("SELECT count\\(\\*\\) FROM `users`").
					WillReturnRows(countRows)
			},
			expectError: false,
			expectUsers: []*models.User{
				{UserID: "user1", Name: "User 1"},
				{UserID: "user2", Name: "User 2"},
				{UserID: "user3", Name: "User 3"},
			},
			expectCount: 500,
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
			users, count, err := repo.GetAllWithCount(tt.perPage, tt.page, tt.search)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, users)
				assert.Equal(t, int64(0), count)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, users)
				assert.Len(t, users, len(tt.expectUsers))
				for i := range users {
					assert.Equal(t, tt.expectUsers[i].UserID, users[i].UserID)
					assert.Equal(t, tt.expectUsers[i].Name, users[i].Name)
				}
				assert.Equal(t, tt.expectCount, count)
			}

			// Verify mock expectations
			err = mock.ExpectationsWereMet()
			assert.NoError(t, err)
		})
	}
}

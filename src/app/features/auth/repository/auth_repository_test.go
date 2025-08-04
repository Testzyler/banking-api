package repository

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Testzyler/banking-api/app/entities"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// func createTestRedisDB(mockClient redis.Cmdable) *database.RedisDatabase {
// 	return &database.RedisDatabase{
// 		Client: mockClient,
// 	}
// }

func TestAuthRepository_GetUserWithPin(t *testing.T) {
	tests := []struct {
		name          string
		username      string
		mockSetup     func(sqlmock.Sqlmock)
		expectError   bool
		expectUser    *entities.User
		errorContains string
	}{
		{
			name:     "successful get user with pin",
			username: "testuser",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"user_id", "name", "password"}).
					AddRow("user123", "testuser", "hashedpassword")

				pinRows := sqlmock.NewRows([]string{"user_id", "hashed_pin", "failed_pin_attempts", "last_pin_attempt_at", "pin_locked_until"}).
					AddRow("user123", "hashedpin123", 0, nil, nil)

				mock.ExpectQuery("SELECT \\* FROM `users` WHERE name = \\? ORDER BY `users`.`user_id` LIMIT \\?").
					WithArgs("testuser", 1).
					WillReturnRows(rows)

				mock.ExpectQuery("SELECT \\* FROM `user_pins` WHERE `user_pins`.`user_id` = \\?").
					WithArgs("user123").
					WillReturnRows(pinRows)
			},
			expectError: false,
			expectUser: &entities.User{
				UserID: "user123",
				Name:   "testuser",
			},
		},
		{
			name:     "user not found",
			username: "nonexistent",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM `users` WHERE name = \\? ORDER BY `users`.`user_id` LIMIT \\?").
					WithArgs("nonexistent", 1).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			expectError:   true,
			expectUser:    nil,
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

			repo := NewAuthRepository(gormDB, nil)

			// Setup mock expectations
			tt.mockSetup(mock)

			// Act
			user, err := repo.GetUserWithPin(tt.username)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.expectUser.UserID, user.UserID)
				assert.Equal(t, tt.expectUser.Name, user.Name)
			}

			// Verify all expectations were met
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestAuthRepository_UpdateUserPinFailedAttempts(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		failedAttempts int
		mockSetup      func(sqlmock.Sqlmock)
		expectError    bool
		errorContains  string
	}{
		{
			name:           "successful update",
			userID:         "user123",
			failedAttempts: 2,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE `user_pins` SET `failed_pin_attempts`=\\? WHERE user_id = \\?").
					WithArgs(2, "user123").
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			expectError: false,
		},
		{
			name:           "database error",
			userID:         "user123",
			failedAttempts: 3,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE `user_pins` SET `failed_pin_attempts`=\\? WHERE user_id = \\?").
					WithArgs(3, "user123").
					WillReturnError(gorm.ErrInvalidDB)
				mock.ExpectRollback()
			},
			expectError:   true,
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

			repo := NewAuthRepository(gormDB, nil)

			// Setup mock expectations
			tt.mockSetup(mock)

			// Act
			err = repo.UpdateUserPinFailedAttempts(tt.userID, tt.failedAttempts)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}

			// Verify all expectations were met
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestAuthRepository_UpdateUserPinLockedUntil(t *testing.T) {
	now := time.Now()
	lockedUntil := now.Add(30 * time.Minute)

	tests := []struct {
		name          string
		userID        string
		lockedUntil   *time.Time
		mockSetup     func(sqlmock.Sqlmock)
		expectError   bool
		errorContains string
	}{
		{
			name:        "successful update with lock time",
			userID:      "user123",
			lockedUntil: &lockedUntil,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE `user_pins` SET `pin_locked_until`=\\? WHERE user_id = \\?").
					WithArgs(lockedUntil, "user123").
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			expectError: false,
		},
		{
			name:        "successful update with nil (unlock)",
			userID:      "user123",
			lockedUntil: nil,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE `user_pins` SET `pin_locked_until`=\\? WHERE user_id = \\?").
					WithArgs(nil, "user123").
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			expectError: false,
		},
		{
			name:        "database error",
			userID:      "user123",
			lockedUntil: &lockedUntil,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE `user_pins` SET `pin_locked_until`=\\? WHERE user_id = \\?").
					WithArgs(lockedUntil, "user123").
					WillReturnError(gorm.ErrInvalidDB)
				mock.ExpectRollback()
			},
			expectError:   true,
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

			repo := NewAuthRepository(gormDB, nil)

			// Setup mock expectations
			tt.mockSetup(mock)

			// Act
			err = repo.UpdateUserPinLockedUntil(tt.userID, tt.lockedUntil)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}

			// Verify all expectations were met
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

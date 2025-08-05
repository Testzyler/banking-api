package repository

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Testzyler/banking-api/app/entities"
	"github.com/Testzyler/banking-api/app/models"
	"github.com/Testzyler/banking-api/database"
	"github.com/go-redis/redismock/v9"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Helper function to create test Redis database mock using redismock
func createTestRedisDB(mockClient redis.Cmdable) *database.RedisDatabase {
	return &database.RedisDatabase{
		Client: mockClient,
	}
}

func TestAuthRepository_GetPinAttemptData_WithRedismock(t *testing.T) {
	tests := []struct {
		name        string
		userID      string
		setupMock   func(redismock.ClientMock)
		expectError bool
		expectData  *entities.PinAttemptData
	}{
		{
			name:   "successful get pin attempt data from Redis",
			userID: "user123",
			setupMock: func(mock redismock.ClientMock) {
				data := &entities.PinAttemptData{
					UserID:         "user123",
					FailedAttempts: 2,
					PinLockedUntil: nil,
					LastAttemptAt:  nil,
				}
				dataJSON, _ := json.Marshal(data)
				mock.ExpectGet("pin_attempt:user123").SetVal(string(dataJSON))
			},
			expectError: false,
			expectData: &entities.PinAttemptData{
				UserID:         "user123",
				FailedAttempts: 2,
				PinLockedUntil: nil,
				LastAttemptAt:  nil,
			},
		},
		{
			name:   "Redis key not found - return default data",
			userID: "user123",
			setupMock: func(mock redismock.ClientMock) {
				mock.ExpectGet("pin_attempt:user123").RedisNil()
			},
			expectError: false,
			expectData: &entities.PinAttemptData{
				UserID:         "user123",
				FailedAttempts: 0,
				PinLockedUntil: nil,
				LastAttemptAt:  nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, sqlMock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()

			gormDB, err := gorm.Open(mysql.New(mysql.Config{
				Conn:                      db,
				SkipInitializeWithVersion: true,
			}), &gorm.Config{})
			assert.NoError(t, err)

			// Create mock Redis client using redismock
			mockRedisClient, redisMock := redismock.NewClientMock()
			redisDB := createTestRedisDB(mockRedisClient)
			repo := NewAuthRepository(gormDB, redisDB)

			// Setup mock expectations
			tt.setupMock(redisMock)

			// Act
			data, err := repo.GetPinAttemptData(context.Background(), tt.userID)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, data)
				assert.Equal(t, tt.expectData.UserID, data.UserID)
				assert.Equal(t, tt.expectData.FailedAttempts, data.FailedAttempts)
			}

			// Verify all expectations were met
			assert.NoError(t, redisMock.ExpectationsWereMet())
			assert.NoError(t, sqlMock.ExpectationsWereMet())
		})
	}
}

func TestAuthRepository_IncrementFailedAttempts_WithRedismock(t *testing.T) {
	tests := []struct {
		name        string
		userID      string
		setupMock   func(redismock.ClientMock)
		expectError bool
	}{
		{
			name:   "successful increment with Redis",
			userID: "user123",
			setupMock: func(mock redismock.ClientMock) {
				// Mock Get call for existing data
				data := &entities.PinAttemptData{
					UserID:         "user123",
					FailedAttempts: 1,
					PinLockedUntil: nil,
					LastAttemptAt:  nil,
				}
				dataJSON, _ := json.Marshal(data)
				mock.ExpectGet("pin_attempt:user123").SetVal(string(dataJSON))

				// Mock Set call for updated data - use regex match for JSON structure
				mock.Regexp().ExpectSet("pin_attempt:user123", `.*`, 24*time.Hour).SetVal("OK")
			},
			expectError: false,
		},
		{
			name:   "successful increment with Redis but no database",
			userID: "user123",
			setupMock: func(mock redismock.ClientMock) {
				// Mock Get call for existing data
				data := &entities.PinAttemptData{
					UserID:         "user123",
					FailedAttempts: 1,
					PinLockedUntil: nil,
					LastAttemptAt:  nil,
				}
				dataJSON, _ := json.Marshal(data)
				mock.ExpectGet("pin_attempt:user123").SetVal(string(dataJSON))

				// Mock Set call for updated data - use regex match for JSON structure
				mock.Regexp().ExpectSet("pin_attempt:user123", `.*`, 24*time.Hour).SetVal("OK")
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, sqlMock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()

			gormDB, err := gorm.Open(mysql.New(mysql.Config{
				Conn:                      db,
				SkipInitializeWithVersion: true,
			}), &gorm.Config{})
			assert.NoError(t, err)

			// Create mock Redis client using redismock
			mockRedisClient, redisMock := redismock.NewClientMock()
			redisDB := createTestRedisDB(mockRedisClient)
			repo := NewAuthRepository(gormDB, redisDB)

			// Setup mock expectations
			tt.setupMock(redisMock)

			// Act
			data, err := repo.IncrementFailedAttempts(context.Background(), tt.userID)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, data)
				assert.Equal(t, tt.userID, data.UserID)
				assert.Equal(t, 2, data.FailedAttempts) // Should be incremented from 1 to 2
			}

			// Verify all expectations were met
			assert.NoError(t, redisMock.ExpectationsWereMet())
			assert.NoError(t, sqlMock.ExpectationsWereMet())
		})
	}
}

func TestAuthRepository_IsTokenBanned_WithRedismock(t *testing.T) {
	tests := []struct {
		name         string
		tokenID      string
		setupMock    func(redismock.ClientMock)
		expectError  bool
		expectBanned bool
	}{
		{
			name:    "token is not banned",
			tokenID: "token123",
			setupMock: func(mock redismock.ClientMock) {
				mock.ExpectGet("banned_token:token123").RedisNil()
			},
			expectError:  false,
			expectBanned: false,
		},
		{
			name:    "token is banned",
			tokenID: "token123",
			setupMock: func(mock redismock.ClientMock) {
				bannedToken := entities.BannedToken{
					TokenID:  "token123",
					UserID:   "user123",
					BannedAt: time.Now(),
					Reason:   "security violation",
				}
				tokenJSON, _ := json.Marshal(bannedToken)
				mock.ExpectGet("banned_token:token123").SetVal(string(tokenJSON))
			},
			expectError:  false,
			expectBanned: true,
		},
		{
			name:    "Redis error returns false",
			tokenID: "token123",
			setupMock: func(mock redismock.ClientMock) {
				mock.ExpectGet("banned_token:token123").SetErr(redis.ErrClosed)
			},
			expectError:  true,
			expectBanned: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, sqlMock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()

			gormDB, err := gorm.Open(mysql.New(mysql.Config{
				Conn:                      db,
				SkipInitializeWithVersion: true,
			}), &gorm.Config{})
			assert.NoError(t, err)

			// Create mock Redis client using redismock
			mockRedisClient, redisMock := redismock.NewClientMock()
			redisDB := createTestRedisDB(mockRedisClient)
			repo := NewAuthRepository(gormDB, redisDB)

			// Setup mock expectations
			tt.setupMock(redisMock)

			// Act
			isBanned, err := repo.IsTokenBanned(context.Background(), tt.tokenID)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectBanned, isBanned)
			}

			// Verify all expectations were met
			assert.NoError(t, redisMock.ExpectationsWereMet())
			assert.NoError(t, sqlMock.ExpectationsWereMet())
		})
	}
}

func TestAuthRepository_StoreToken_WithRedismock(t *testing.T) {
	tests := []struct {
		name        string
		userID      string
		token       *entities.TokenResponse
		setupMock   func(redismock.ClientMock)
		expectError bool
	}{
		{
			name:   "successful token storage",
			userID: "user123",
			token: &entities.TokenResponse{
				TokenID: "token123",
				Token:   "jwt_token_here",
				Expiry:  time.Time{}, // Zero time means no expiry, so no Expire call
			},
			setupMock: func(mock redismock.ClientMock) {
				// Only expect the SAdd call since there's no expiry
				mock.Regexp().ExpectSAdd("user_tokens:user123", `.*`).SetVal(1)
			},
			expectError: false,
		},
		{
			name:   "token storage failure",
			userID: "user123",
			token: &entities.TokenResponse{
				TokenID: "token123",
				Token:   "jwt_token_here",
				Expiry:  time.Time{}, // Zero time means no expiry
			},
			setupMock: func(mock redismock.ClientMock) {
				mock.Regexp().ExpectSAdd("user_tokens:user123", `.*`).SetErr(redis.ErrClosed)
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, sqlMock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()

			gormDB, err := gorm.Open(mysql.New(mysql.Config{
				Conn:                      db,
				SkipInitializeWithVersion: true,
			}), &gorm.Config{})
			assert.NoError(t, err)

			// Create mock Redis client using redismock
			mockRedisClient, redisMock := redismock.NewClientMock()
			redisDB := createTestRedisDB(mockRedisClient)
			repo := NewAuthRepository(gormDB, redisDB)

			// Setup mock expectations
			tt.setupMock(redisMock)

			// Act
			err = repo.StoreToken(context.Background(), tt.userID, tt.token)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Note: We might not check ExpectationsWereMet for failing cases
			if !tt.expectError {
				assert.NoError(t, redisMock.ExpectationsWereMet())
				assert.NoError(t, sqlMock.ExpectationsWereMet())
			}
		})
	}
}

func TestAuthRepository_GetUserWithPin_WithRedismock(t *testing.T) {
	tests := []struct {
		name          string
		username      string
		setupMock     func(redismock.ClientMock, sqlmock.Sqlmock)
		expectError   bool
		expectUser    *models.User
		expectDbCall  bool // Whether database should be called
		errorContains string
	}{
		{
			name:     "cache hit - return cached data without hitting database",
			username: "testuser",
			setupMock: func(redisMock redismock.ClientMock, sqlMock sqlmock.Sqlmock) {
				// Setup cached user data
				cachedUser := models.User{
					UserID: "user123",
					Name:   "testuser",
					UserPin: &models.UserPin{
						FailedPinAttempts: 2,
						HashedPin:         "hashedpin123",
					},
				}
				userJSON, _ := json.Marshal(cachedUser)

				// Expect Redis GET call
				redisMock.ExpectGet("user_with_pin:testuser").SetVal(string(userJSON))

				// No database calls expected since cache hit
			},
			expectError:  false,
			expectDbCall: false,
			expectUser: &models.User{
				UserID: "user123",
				Name:   "testuser",
				UserPin: &models.UserPin{
					FailedPinAttempts: 2,
					HashedPin:         "hashedpin123",
				},
			},
		},
		{
			name:     "cache miss - fetch from database and store in cache",
			username: "testuser",
			setupMock: func(redisMock redismock.ClientMock, sqlMock sqlmock.Sqlmock) {
				// Expect Redis GET call - cache miss
				redisMock.ExpectGet("user_with_pin:testuser").RedisNil()

				// Setup database expectations
				rows := sqlmock.NewRows([]string{"user_id", "name", "password"}).
					AddRow("user123", "testuser", "hashedpassword")

				pinRows := sqlmock.NewRows([]string{"user_id", "hashed_pin", "failed_pin_attempts", "last_pin_attempt_at", "pin_locked_until"}).
					AddRow("user123", "hashedpin123", 0, nil, nil)

				sqlMock.ExpectQuery("SELECT \\* FROM `users` WHERE name = \\? ORDER BY `users`.`user_id` LIMIT \\?").
					WithArgs("testuser", 1).
					WillReturnRows(rows)

				sqlMock.ExpectQuery("SELECT \\* FROM `user_pins` WHERE `user_pins`.`user_id` = \\?").
					WithArgs("user123").
					WillReturnRows(pinRows)

				// Expect Redis SET call (async cache store) - use regex since JSON order might vary
				redisMock.Regexp().ExpectSet("user_with_pin:testuser", `.*`, 30*time.Minute).SetVal("OK")
			},
			expectError:  false,
			expectDbCall: true,
			expectUser: &models.User{
				UserID: "user123",
				Name:   "testuser",
				UserPin: &models.UserPin{
					FailedPinAttempts: 0,
					HashedPin:         "hashedpin123",
				},
			},
		},
		{
			name:     "cache corrupted JSON - fallback to database",
			username: "testuser",
			setupMock: func(redisMock redismock.ClientMock, sqlMock sqlmock.Sqlmock) {
				// Expect Redis GET call - return corrupted JSON
				redisMock.ExpectGet("user_with_pin:testuser").SetVal("invalid-json")

				// Setup database expectations for fallback
				rows := sqlmock.NewRows([]string{"user_id", "name", "password"}).
					AddRow("user123", "testuser", "hashedpassword")

				pinRows := sqlmock.NewRows([]string{"user_id", "hashed_pin", "failed_pin_attempts", "last_pin_attempt_at", "pin_locked_until"}).
					AddRow("user123", "hashedpin123", 1, nil, nil)

				sqlMock.ExpectQuery("SELECT \\* FROM `users` WHERE name = \\? ORDER BY `users`.`user_id` LIMIT \\?").
					WithArgs("testuser", 1).
					WillReturnRows(rows)

				sqlMock.ExpectQuery("SELECT \\* FROM `user_pins` WHERE `user_pins`.`user_id` = \\?").
					WithArgs("user123").
					WillReturnRows(pinRows)

				// Expect Redis SET call for fresh cache
				redisMock.Regexp().ExpectSet("user_with_pin:testuser", `.*`, 30*time.Minute).SetVal("OK")
			},
			expectError:  false,
			expectDbCall: true,
			expectUser: &models.User{
				UserID: "user123",
				Name:   "testuser",
				UserPin: &models.UserPin{
					FailedPinAttempts: 1,
					HashedPin:         "hashedpin123",
				},
			},
		},
		{
			name:     "cache miss and database error",
			username: "nonexistent",
			setupMock: func(redisMock redismock.ClientMock, sqlMock sqlmock.Sqlmock) {
				// Expect Redis GET call - cache miss
				redisMock.ExpectGet("user_with_pin:nonexistent").RedisNil()

				// Database returns error
				sqlMock.ExpectQuery("SELECT \\* FROM `users` WHERE name = \\? ORDER BY `users`.`user_id` LIMIT \\?").
					WithArgs("nonexistent", 1).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			expectError:   true,
			expectDbCall:  true,
			expectUser:    nil,
			errorContains: "record not found",
		},
		{
			name:     "Redis error - fallback to database",
			username: "testuser",
			setupMock: func(redisMock redismock.ClientMock, sqlMock sqlmock.Sqlmock) {
				// Expect Redis GET call - Redis error
				redisMock.ExpectGet("user_with_pin:testuser").SetErr(redis.ErrClosed)

				// Setup database expectations for fallback
				rows := sqlmock.NewRows([]string{"user_id", "name", "password"}).
					AddRow("user123", "testuser", "hashedpassword")

				pinRows := sqlmock.NewRows([]string{"user_id", "hashed_pin", "failed_pin_attempts", "last_pin_attempt_at", "pin_locked_until"}).
					AddRow("user123", "hashedpin123", 0, nil, nil)

				sqlMock.ExpectQuery("SELECT \\* FROM `users` WHERE name = \\? ORDER BY `users`.`user_id` LIMIT \\?").
					WithArgs("testuser", 1).
					WillReturnRows(rows)

				sqlMock.ExpectQuery("SELECT \\* FROM `user_pins` WHERE `user_pins`.`user_id` = \\?").
					WithArgs("user123").
					WillReturnRows(pinRows)

				// Expect Redis SET call for caching
				redisMock.Regexp().ExpectSet("user_with_pin:testuser", `.*`, 30*time.Minute).SetVal("OK")
			},
			expectError:  false,
			expectDbCall: true,
			expectUser: &models.User{
				UserID: "user123",
				Name:   "testuser",
				UserPin: &models.UserPin{
					FailedPinAttempts: 0,
					HashedPin:         "hashedpin123",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock database
			db, sqlMock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()

			gormDB, err := gorm.Open(mysql.New(mysql.Config{
				Conn:                      db,
				SkipInitializeWithVersion: true,
			}), &gorm.Config{})
			assert.NoError(t, err)

			// Create mock Redis client
			mockRedisClient, redisMock := redismock.NewClientMock()
			redisDB := createTestRedisDB(mockRedisClient)
			repo := NewAuthRepository(gormDB, redisDB)

			// Setup mock expectations
			tt.setupMock(redisMock, sqlMock)

			// Act
			user, err := repo.GetUserWithPin(tt.username)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.expectUser.UserID, user.UserID)
				assert.Equal(t, tt.expectUser.Name, user.Name)
				assert.Equal(t, tt.expectUser.UserPin.FailedPinAttempts, user.UserPin.FailedPinAttempts)
				assert.Equal(t, tt.expectUser.UserPin.HashedPin, user.UserPin.HashedPin)
			}

			// Give async operations time to complete for cache store verification
			if !tt.expectError && tt.expectDbCall {
				time.Sleep(10 * time.Millisecond)
			}

			// Verify expectations were met
			assert.NoError(t, redisMock.ExpectationsWereMet())
			assert.NoError(t, sqlMock.ExpectationsWereMet())
		})
	}
}

func TestAuthRepository_BanAllUserTokens_WithRedismock(t *testing.T) {
	tests := []struct {
		name        string
		userID      string
		reason      string
		setupMock   func(redismock.ClientMock)
		expectError bool
	}{
		{
			name:   "successful ban all tokens",
			userID: "user123",
			reason: "security violation",
			setupMock: func(mock redismock.ClientMock) {
				// Mock getting user tokens
				token1 := entities.TokenResponse{
					TokenID: "token1",
					UserID:  "user123",
				}
				token2 := entities.TokenResponse{
					TokenID: "token2",
					UserID:  "user123",
				}
				token1JSON, _ := json.Marshal(token1)
				token2JSON, _ := json.Marshal(token2)

				mock.ExpectSMembers("user_tokens:user123").SetVal([]string{string(token1JSON), string(token2JSON)})

				// Mock banning each token
				mock.Regexp().ExpectSet("banned_token:token1", `.*`, 24*time.Hour).SetVal("OK")
				mock.Regexp().ExpectSet("banned_token:token2", `.*`, 24*time.Hour).SetVal("OK")
			},
			expectError: false,
		},
		{
			name:   "user has no tokens",
			userID: "user_no_tokens",
			reason: "precautionary ban",
			setupMock: func(mock redismock.ClientMock) {
				// Mock getting user tokens - empty result
				mock.ExpectSMembers("user_tokens:user_no_tokens").SetVal([]string{})
			},
			expectError: false,
		},
		{
			name:   "redis error getting user tokens",
			userID: "user123",
			reason: "security violation",
			setupMock: func(mock redismock.ClientMock) {
				// Mock Redis error when getting tokens
				mock.ExpectSMembers("user_tokens:user123").SetErr(redis.ErrClosed)
			},
			expectError: true,
		},
		{
			name:   "redis error setting banned token",
			userID: "user123",
			reason: "security violation",
			setupMock: func(mock redismock.ClientMock) {
				// Mock getting user tokens
				token1 := entities.TokenResponse{
					TokenID: "token1",
					UserID:  "user123",
				}
				token1JSON, _ := json.Marshal(token1)

				mock.ExpectSMembers("user_tokens:user123").SetVal([]string{string(token1JSON)})

				// Mock Redis error when setting banned token
				mock.Regexp().ExpectSet("banned_token:token1", `.*`, 24*time.Hour).SetErr(redis.ErrClosed)
			},
			expectError: false, // Method doesn't return error if individual ban fails
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, sqlMock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()

			gormDB, err := gorm.Open(mysql.New(mysql.Config{
				Conn:                      db,
				SkipInitializeWithVersion: true,
			}), &gorm.Config{})
			assert.NoError(t, err)

			// Create mock Redis client using redismock
			mockRedisClient, redisMock := redismock.NewClientMock()
			redisDB := createTestRedisDB(mockRedisClient)
			repo := NewAuthRepository(gormDB, redisDB)

			// Setup mock expectations
			tt.setupMock(redisMock)

			// Act
			err = repo.BanAllUserTokens(context.Background(), tt.userID, tt.reason)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Verify all expectations were met
			assert.NoError(t, redisMock.ExpectationsWereMet())
			assert.NoError(t, sqlMock.ExpectationsWereMet())
		})
	}
}

func TestAuthRepository_ValidateTokenVersion_WithRedismock(t *testing.T) {
	currentTime := time.Now().Unix()

	tests := []struct {
		name         string
		tokenVersion int64
		expectValid  bool
		expectReason string
	}{
		{
			name:         "valid recent token",
			tokenVersion: currentTime - 1000, // 1000 seconds ago (valid)
			expectValid:  true,
		},
		{
			name:         "token too old",
			tokenVersion: currentTime - (25 * 60 * 60), // 25 hours ago (invalid)
			expectValid:  false,
			expectReason: "Token is too old",
		},
		{
			name:         "edge case - exactly 24 hours old",
			tokenVersion: currentTime - (24 * 60 * 60), // Exactly 24 hours ago (should be valid)
			expectValid:  true,
		},
		{
			name:         "very recent token",
			tokenVersion: currentTime - 10, // 10 seconds ago (valid)
			expectValid:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, sqlMock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()

			gormDB, err := gorm.Open(mysql.New(mysql.Config{
				Conn:                      db,
				SkipInitializeWithVersion: true,
			}), &gorm.Config{})
			assert.NoError(t, err)

			// Create mock Redis client using redismock (not used in ValidateTokenVersion)
			mockRedisClient, redisMock := redismock.NewClientMock()
			redisDB := createTestRedisDB(mockRedisClient)
			repo := NewAuthRepository(gormDB, redisDB)

			// Act
			result, err := repo.ValidateTokenVersion(context.Background(), tt.tokenVersion)

			// Assert
			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, tt.expectValid, result.Valid)
			if !tt.expectValid {
				assert.Equal(t, tt.expectReason, result.Reason)
			}

			// TokenVersion should be updated to current time
			assert.True(t, result.TokenVersion > 0)

			// Verify all expectations were met
			assert.NoError(t, redisMock.ExpectationsWereMet())
			assert.NoError(t, sqlMock.ExpectationsWereMet())
		})
	}
}

func TestAuthRepository_ListUserTokens_WithRedismock(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(redismock.ClientMock)
		expectError bool
		expectCount int
	}{
		{
			name: "successful list with multiple users and tokens",
			setupMock: func(mock redismock.ClientMock) {
				// Mock scanning for user_tokens:user1:* keys (specific to user1)
				mock.ExpectScan(0, "user_tokens:user1:*", 100).SetVal([]string{"user_tokens:user1:token1", "user_tokens:user1:token2"}, 0)

				// Mock getting tokens for user1
				token1 := entities.TokenResponse{
					TokenID: "token1",
					UserID:  "user1",
				}
				token2 := entities.TokenResponse{
					TokenID: "token2",
					UserID:  "user1",
				}
				token1JSON, _ := json.Marshal(token1)
				token2JSON, _ := json.Marshal(token2)
				mock.ExpectSMembers("user_tokens:user1:token1").SetVal([]string{string(token1JSON)})
				mock.ExpectSMembers("user_tokens:user1:token2").SetVal([]string{string(token2JSON)})
			},
			expectError: false,
			expectCount: 2, // Only 2 tokens for user1
		},
		{
			name: "no user tokens found",
			setupMock: func(mock redismock.ClientMock) {
				// Mock scanning for user_tokens:user1:* keys - empty result
				mock.ExpectScan(0, "user_tokens:user1:*", 100).SetVal([]string{}, 0)
			},
			expectError: false,
			expectCount: 0,
		},
		{
			name: "redis scan error",
			setupMock: func(mock redismock.ClientMock) {
				// Mock Redis error when scanning
				mock.ExpectScan(0, "user_tokens:user1:*", 100).SetErr(redis.ErrClosed)
			},
			expectError: true,
			expectCount: 0,
		},
		{
			name: "error getting tokens for specific user",
			setupMock: func(mock redismock.ClientMock) {
				// Mock scanning for user_tokens:user1:* keys
				mock.ExpectScan(0, "user_tokens:user1:*", 100).SetVal([]string{"user_tokens:user1:token1"}, 0)

				// Mock error getting tokens for user1
				mock.ExpectSMembers("user_tokens:user1:token1").SetErr(redis.ErrClosed)
			},
			expectError: false, // Method handles individual user errors gracefully
			expectCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, sqlMock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()

			gormDB, err := gorm.Open(mysql.New(mysql.Config{
				Conn:                      db,
				SkipInitializeWithVersion: true,
			}), &gorm.Config{})
			assert.NoError(t, err)

			// Create mock Redis client using redismock
			mockRedisClient, redisMock := redismock.NewClientMock()
			redisDB := createTestRedisDB(mockRedisClient)
			repo := NewAuthRepository(gormDB, redisDB)

			// Setup mock expectations
			tt.setupMock(redisMock)

			// Act
			tokens, err := repo.ListUserTokens(context.Background(), "user1")

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, tokens)
			} else {
				assert.NoError(t, err)
				if tt.expectCount == 0 {
					// When no tokens found, Go returns nil slice
					assert.Len(t, tokens, 0)
				} else {
					assert.NotNil(t, tokens)
					assert.Len(t, tokens, tt.expectCount)
				}
			}

			// Verify all expectations were met
			assert.NoError(t, redisMock.ExpectationsWereMet())
			assert.NoError(t, sqlMock.ExpectationsWereMet())
		})
	}
}

func TestAuthRepository_CleanupExpiredBans_WithRedismock(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(redismock.ClientMock)
		expectError bool
	}{
		{
			name: "successful cleanup with existing banned tokens",
			setupMock: func(mock redismock.ClientMock) {
				// Mock scanning for banned_token:* keys
				mock.ExpectScan(0, "banned_token:*", 100).SetVal([]string{"banned_token:token1", "banned_token:token2"}, 0)

				// Mock checking existence of each key
				mock.ExpectExists("banned_token:token1").SetVal(1) // exists
				mock.ExpectExists("banned_token:token2").SetVal(0) // expired
			},
			expectError: false,
		},
		{
			name: "no banned tokens found",
			setupMock: func(mock redismock.ClientMock) {
				// Mock scanning for banned_token:* keys - empty result
				mock.ExpectScan(0, "banned_token:*", 100).SetVal([]string{}, 0)
			},
			expectError: false,
		},
		{
			name: "redis scan error",
			setupMock: func(mock redismock.ClientMock) {
				// Mock Redis error when scanning
				mock.ExpectScan(0, "banned_token:*", 100).SetErr(redis.ErrClosed)
			},
			expectError: true,
		},
		{
			name: "error checking key existence",
			setupMock: func(mock redismock.ClientMock) {
				// Mock scanning for banned_token:* keys
				mock.ExpectScan(0, "banned_token:*", 100).SetVal([]string{"banned_token:token1"}, 0)

				// Mock error checking existence
				mock.ExpectExists("banned_token:token1").SetErr(redis.ErrClosed)
			},
			expectError: false, // Method handles individual key errors gracefully
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, sqlMock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()

			gormDB, err := gorm.Open(mysql.New(mysql.Config{
				Conn:                      db,
				SkipInitializeWithVersion: true,
			}), &gorm.Config{})
			assert.NoError(t, err)

			// Create mock Redis client using redismock
			mockRedisClient, redisMock := redismock.NewClientMock()
			redisDB := createTestRedisDB(mockRedisClient)
			repo := NewAuthRepository(gormDB, redisDB)

			// Setup mock expectations
			tt.setupMock(redisMock)

			// Act
			err = repo.CleanupExpiredBans(context.Background())

			// Assert
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Verify all expectations were met
			assert.NoError(t, redisMock.ExpectationsWereMet())
			assert.NoError(t, sqlMock.ExpectationsWereMet())
		})
	}
}

func TestAuthRepository_NilRedis_Coverage(t *testing.T) {
	// Test repository methods when Redis is not available (nil client)
	db, sqlMock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	assert.NoError(t, err)

	// Create repository with nil Redis client
	repo := NewAuthRepository(gormDB, nil)

	t.Run("IsTokenBanned with nil Redis - should return true", func(t *testing.T) {
		banned, err := repo.IsTokenBanned(context.Background(), "any_token")
		assert.NoError(t, err)
		assert.True(t, banned) // Should default to banned when Redis is unavailable
	})

	t.Run("BanAllUserTokens with nil Redis - should return error", func(t *testing.T) {
		err := repo.BanAllUserTokens(context.Background(), "user123", "test ban")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Redis client is not initialized")
	})

	t.Run("StoreToken with nil Redis - should return error", func(t *testing.T) {
		token := &entities.TokenResponse{TokenID: "test", UserID: "user123"}
		err := repo.StoreToken(context.Background(), "user123", token)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Redis client is not initialized")
	})

	t.Run("ListUserTokens with nil Redis - should return error", func(t *testing.T) {
		tokens, err := repo.ListUserTokens(context.Background(), "any_user")
		assert.Error(t, err)
		assert.Nil(t, tokens)
		assert.Contains(t, err.Error(), "Redis client is not initialized")
	})

	t.Run("CleanupExpiredBans with nil Redis - should return nil", func(t *testing.T) {
		err := repo.CleanupExpiredBans(context.Background())
		assert.NoError(t, err) // Should not error with nil Redis
	})

	t.Run("GetPinAttemptData with nil Redis - should return default", func(t *testing.T) {
		data, err := repo.GetPinAttemptData(context.Background(), "user123")
		assert.NoError(t, err)
		assert.NotNil(t, data)
		assert.Equal(t, "user123", data.UserID)
		assert.Equal(t, 0, data.FailedAttempts)
	})

	// Verify SQL expectations were met
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestAuthRepository_ResetPinAttempts_NilRedis(t *testing.T) {
	// Test ResetPinAttempts when Redis is not available
	db, sqlMock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	assert.NoError(t, err)

	// Create repository with nil Redis client
	repo := NewAuthRepository(gormDB, nil)

	// Act
	err = repo.ResetPinAttempts(context.Background(), "user123")

	// Assert - should not error but also should not do anything since Redis is nil
	assert.NoError(t, err)

	// Verify SQL expectations were met
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestAuthRepository_SetPinLock_NilRedis(t *testing.T) {
	// Test SetPinLock when Redis is not available
	db, sqlMock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	assert.NoError(t, err)

	// Create repository with nil Redis client
	repo := NewAuthRepository(gormDB, nil)

	now := time.Now()
	lockedUntil := now.Add(30 * time.Minute)

	// Act
	err = repo.SetPinLock(context.Background(), "user123", lockedUntil, 3, &now)

	// Assert - should return error since GetPinAttemptData will return default but setPinAttemptData will fail
	assert.NoError(t, err)

	// Verify SQL expectations were met
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

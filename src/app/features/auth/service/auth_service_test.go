package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Testzyler/banking-api/app/entities"
	"github.com/Testzyler/banking-api/app/models"
	"github.com/Testzyler/banking-api/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Mock AuthRepository
type MockAuthRepository struct {
	mock.Mock
}

func (m *MockAuthRepository) GetUserWithPin(username string) (*models.User, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockAuthRepository) UpdateUserPinFailedAttempts(userID string, failedAttempts int) error {
	args := m.Called(userID, failedAttempts)
	return args.Error(0)
}

func (m *MockAuthRepository) UpdateUserPinLockedUntil(userID string, lockedUntil *time.Time) error {
	args := m.Called(userID, lockedUntil)
	return args.Error(0)
}

func (m *MockAuthRepository) UpdateUserPinLastAttemptAt(userID string, lastAttemptAt *time.Time) error {
	args := m.Called(userID, lastAttemptAt)
	return args.Error(0)
}

// Redis methods
func (m *MockAuthRepository) GetPinAttemptData(ctx context.Context, userID string) (*entities.PinAttemptData, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.PinAttemptData), args.Error(1)
}

func (m *MockAuthRepository) IncrementFailedAttempts(ctx context.Context, userID string) (*entities.PinAttemptData, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.PinAttemptData), args.Error(1)
}

func (m *MockAuthRepository) SetPinLock(ctx context.Context, userID string, lockedUntil time.Time, failedAttempts int, lastAttemptAt *time.Time) error {
	args := m.Called(ctx, userID, lockedUntil, failedAttempts, lastAttemptAt)
	return args.Error(0)
}

func (m *MockAuthRepository) ResetPinAttempts(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockAuthRepository) ListUserTokens(ctx context.Context, userID string) ([]entities.TokenResponse, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entities.TokenResponse), args.Error(1)
}

func (m *MockAuthRepository) StoreToken(ctx context.Context, userID string, tokenResponse *entities.TokenResponse) error {
	args := m.Called(ctx, userID, tokenResponse)
	return args.Error(0)
}

func (m *MockAuthRepository) BanAllUserTokens(ctx context.Context, userID, reason string) error {
	args := m.Called(ctx, userID, reason)
	return args.Error(0)
}

func (m *MockAuthRepository) IsTokenBanned(ctx context.Context, tokenID string) (bool, error) {
	args := m.Called(ctx, tokenID)
	return args.Bool(0), args.Error(1)
}

func (m *MockAuthRepository) IsInBlacklist(ctx context.Context, userID string, tokenVersion int64) (bool, error) {
	args := m.Called(ctx, userID, tokenVersion)
	return args.Bool(0), args.Error(1)
}

func (m *MockAuthRepository) ValidateTokenVersion(ctx context.Context, tokenVersion int64) (*entities.TokenValidationResult, error) {
	args := m.Called(ctx, tokenVersion)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.TokenValidationResult), args.Error(1)
}

func (m *MockAuthRepository) CleanupExpiredBans(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// Helper function to create test models.User
func createTestUser(userID, username, hashedPin string, failedAttempts int, lockedUntil, lastAttempt *time.Time) *models.User {
	return &models.User{
		UserID: userID,
		Name:   username,
		UserPin: &models.UserPin{
			HashedPin:         hashedPin,
			FailedPinAttempts: failedAttempts,
			PinLockedUntil:    lockedUntil,
			LastPinAttemptAt:  lastAttempt,
		},
	}
}

// Mock JwtService
type MockJwtService struct {
	mock.Mock
}

func (m *MockJwtService) GenerateTokens(userID, username string) (*entities.TokenResponse, error) {
	args := m.Called(userID, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.TokenResponse), args.Error(1)
}

func (m *MockJwtService) ValidateAccessToken(tokenString string) (*entities.Claims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Claims), args.Error(1)
}

func (m *MockJwtService) ValidateRefreshToken(tokenString string) (*entities.Claims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Claims), args.Error(1)
}

func (m *MockJwtService) RefreshAccessToken(refreshTokenString string) (*entities.TokenResponse, error) {
	args := m.Called(refreshTokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.TokenResponse), args.Error(1)
}

func (m *MockJwtService) ValidateTokenWithBanCheck(tokenString string) (*entities.TokenValidationResult, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.TokenValidationResult), args.Error(1)
}

func TestAuthService_VerifyPin(t *testing.T) {
	// Create test pin hash
	hashedPin, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)

	tests := []struct {
		name          string
		params        entities.PinVerifyParams
		mockSetup     func(*MockAuthRepository, *MockJwtService)
		expectError   bool
		expectToken   bool
		errorContains string
	}{
		{
			name: "successful pin verification",
			params: entities.PinVerifyParams{
				Username: "testuser",
				Pin:      "123456",
			},
			mockSetup: func(mockRepo *MockAuthRepository, mockJwt *MockJwtService) {
				user := createTestUser("user123", "testuser", string(hashedPin), 0, nil, nil)
				mockRepo.On("GetUserWithPin", "testuser").Return(user, nil)

				// Mock Redis cache data
				cacheData := &entities.PinAttemptData{
					UserID:         "user123",
					FailedAttempts: 0,
					PinLockedUntil: nil,
				}
				mockRepo.On("GetPinAttemptData", mock.Anything, "user123").Return(cacheData, nil)
				mockRepo.On("ResetPinAttempts", mock.Anything, "user123").Return(nil)

				tokenResponse := &entities.TokenResponse{
					Token:        "access_token",
					RefreshToken: "refresh_token",
					Expiry:       time.Now().Add(time.Hour),
				}
				mockJwt.On("GenerateTokens", "user123", "testuser").Return(tokenResponse, nil)
				mockRepo.On("StoreToken", mock.Anything, "user123", mock.AnythingOfType("*entities.TokenResponse")).Return(nil)
			},
			expectError: false,
			expectToken: true,
		},
		{
			name: "user not found",
			params: entities.PinVerifyParams{
				Username: "nonexistent",
				Pin:      "123456",
			},
			mockSetup: func(mockRepo *MockAuthRepository, mockJwt *MockJwtService) {
				mockRepo.On("GetUserWithPin", "nonexistent").Return(nil, gorm.ErrRecordNotFound)
			},
			expectError:   true,
			expectToken:   false,
			errorContains: "User not found",
		},
		{
			name: "incorrect pin",
			params: entities.PinVerifyParams{
				Username: "testuser",
				Pin:      "wrong123",
			},
			mockSetup: func(mockRepo *MockAuthRepository, mockJwt *MockJwtService) {
				user := createTestUser("user123", "testuser", string(hashedPin), 0, nil, nil)
				mockRepo.On("GetUserWithPin", "testuser").Return(user, nil)

				// Mock cache data for pin attempts
				cacheData := &entities.PinAttemptData{
					UserID:         "user123",
					FailedAttempts: 0,
					PinLockedUntil: nil,
				}
				mockRepo.On("GetPinAttemptData", mock.Anything, "user123").Return(cacheData, nil)

				// Mock IncrementFailedAttempts (called by handleFailedAttempt)
				incrementedData := &entities.PinAttemptData{
					UserID:         "user123",
					FailedAttempts: 1,
					PinLockedUntil: nil,
				}
				mockRepo.On("IncrementFailedAttempts", mock.Anything, "user123").Return(incrementedData, nil)
			},
			expectError:   true,
			expectToken:   false,
			errorContains: "PIN is incorrect",
		},
		{
			name: "database error",
			params: entities.PinVerifyParams{
				Username: "testuser",
				Pin:      "123456",
			},
			mockSetup: func(mockRepo *MockAuthRepository, mockJwt *MockJwtService) {
				mockRepo.On("GetUserWithPin", "testuser").Return(nil, errors.New("database connection error"))
			},
			expectError:   true,
			expectToken:   false,
			errorContains: "database connection error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockAuthRepository)
			mockJwt := new(MockJwtService)

			config := &config.Config{
				Auth: &config.AuthConfig{
					Pin: &config.PinConfig{
						BaseDuration:    10 * time.Second,
						LockThreshold:   3,
						MaxLockDuration: 300 * time.Second,
					},
				},
			}

			service := NewAuthService(mockRepo, mockJwt, config)

			// Setup mock expectations
			tt.mockSetup(mockRepo, mockJwt)

			// Act
			ctx := context.Background()
			tokenResponse, err := service.VerifyPin(ctx, tt.params)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, tokenResponse)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				if tt.expectToken {
					assert.NotNil(t, tokenResponse)
					assert.NotEmpty(t, tokenResponse.Token)
					assert.Equal(t, "user123", tokenResponse.UserID)
				}
			}

			// Verify all expectations were met
			mockRepo.AssertExpectations(t)
			mockJwt.AssertExpectations(t)
		})
	}
}

func TestAuthService_RefreshToken(t *testing.T) {
	tests := []struct {
		name          string
		refreshToken  string
		mockSetup     func(*MockAuthRepository, *MockJwtService)
		expectError   bool
		expectToken   bool
		errorContains string
	}{
		{
			name:         "successful token refresh",
			refreshToken: "valid_refresh_token",
			mockSetup: func(mockRepo *MockAuthRepository, mockJwt *MockJwtService) {
				tokenResponse := &entities.TokenResponse{
					Token:        "new_access_token",
					RefreshToken: "valid_refresh_token",
					Expiry:       time.Now().Add(time.Hour),
					UserID:       "user123",
				}
				mockJwt.On("RefreshAccessToken", "valid_refresh_token").Return(tokenResponse, nil)
				mockRepo.On("StoreToken", mock.Anything, "user123", mock.AnythingOfType("*entities.TokenResponse")).Return(nil)
			},
			expectError: false,
			expectToken: true,
		},
		{
			name:         "invalid refresh token",
			refreshToken: "invalid_token",
			mockSetup: func(mockRepo *MockAuthRepository, mockJwt *MockJwtService) {
				mockJwt.On("RefreshAccessToken", "invalid_token").Return(nil, errors.New("invalid token"))
			},
			expectError:   true,
			expectToken:   false,
			errorContains: "invalid token",
		},
		{
			name:         "empty refresh token",
			refreshToken: "",
			mockSetup: func(mockRepo *MockAuthRepository, mockJwt *MockJwtService) {
				mockJwt.On("RefreshAccessToken", "").Return(nil, errors.New("token is empty"))
			},
			expectError:   true,
			expectToken:   false,
			errorContains: "token is empty",
		},
		{
			name:         "banned refresh token",
			refreshToken: "banned_token",
			mockSetup: func(mockRepo *MockAuthRepository, mockJwt *MockJwtService) {
				mockJwt.On("RefreshAccessToken", "banned_token").Return(nil, errors.New("token is banned"))
			},
			expectError:   true,
			expectToken:   false,
			errorContains: "token is banned",
		},
		{
			name:         "successful token refresh with store",
			refreshToken: "valid_refresh_token_store",
			mockSetup: func(mockRepo *MockAuthRepository, mockJwt *MockJwtService) {
				tokenResponse := &entities.TokenResponse{
					Token:        "new_access_token",
					RefreshToken: "valid_refresh_token_store",
					Expiry:       time.Now().Add(time.Hour),
					UserID:       "user123",
				}
				mockJwt.On("RefreshAccessToken", "valid_refresh_token_store").Return(tokenResponse, nil)
				mockRepo.On("StoreToken", mock.Anything, "user123", mock.AnythingOfType("*entities.TokenResponse")).Return(nil)
			},
			expectError: false,
			expectToken: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockAuthRepository)
			mockJwt := new(MockJwtService)

			config := &config.Config{}
			service := NewAuthService(mockRepo, mockJwt, config)

			// Setup mock expectations
			tt.mockSetup(mockRepo, mockJwt)

			// Act
			tokenResponse, err := service.RefreshToken(tt.refreshToken)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, tokenResponse)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				if tt.expectToken {
					assert.NotNil(t, tokenResponse)
					assert.NotEmpty(t, tokenResponse.Token)
					assert.Equal(t, "user123", tokenResponse.UserID)
				}
			}

			// Verify all expectations were met
			mockJwt.AssertExpectations(t)
		})
	}
}

func TestAuthService_VerifyPin_FailedAttempts_And_Locking(t *testing.T) {
	// Create test pin hash
	hashedPin, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)

	tests := []struct {
		name                      string
		params                    entities.PinVerifyParams
		currentFailedAttempts     int
		mockSetup                 func(*MockAuthRepository, *MockJwtService)
		expectError               bool
		expectLocked              bool
		expectedRemainingAttempts int
		errorContains             string
	}{
		{
			name: "first failed attempt - not locked yet",
			params: entities.PinVerifyParams{
				Username: "testuser",
				Pin:      "wrong123",
			},
			currentFailedAttempts: 0,
			mockSetup: func(mockRepo *MockAuthRepository, mockJwt *MockJwtService) {
				user := createTestUser("user123", "testuser", string(hashedPin), 0, nil, nil)
				mockRepo.On("GetUserWithPin", "testuser").Return(user, nil)

				// Mock cache data for pin attempts
				cacheData := &entities.PinAttemptData{
					UserID:         "user123",
					FailedAttempts: 0,
					PinLockedUntil: nil,
				}
				mockRepo.On("GetPinAttemptData", mock.Anything, "user123").Return(cacheData, nil)

				// Mock IncrementFailedAttempts (called by handleFailedAttempt)
				incrementedData := &entities.PinAttemptData{
					UserID:         "user123",
					FailedAttempts: 1,
					PinLockedUntil: nil,
				}
				mockRepo.On("IncrementFailedAttempts", mock.Anything, "user123").Return(incrementedData, nil)
			},
			expectError:               true,
			expectLocked:              false,
			expectedRemainingAttempts: 2, // threshold=3, newAttempts=1, remaining=2
			errorContains:             "2 attempts remaining",
		},
		{
			name: "second failed attempt - still not locked",
			params: entities.PinVerifyParams{
				Username: "testuser",
				Pin:      "wrong123",
			},
			currentFailedAttempts: 1,
			mockSetup: func(mockRepo *MockAuthRepository, mockJwt *MockJwtService) {
				user := createTestUser("user123", "testuser", string(hashedPin), 1, nil, nil)
				mockRepo.On("GetUserWithPin", "testuser").Return(user, nil)

				// Mock cache data for pin attempts
				cacheData := &entities.PinAttemptData{
					UserID:         "user123",
					FailedAttempts: 1,
					PinLockedUntil: nil,
				}
				mockRepo.On("GetPinAttemptData", mock.Anything, "user123").Return(cacheData, nil)

				// Mock IncrementFailedAttempts (called by handleFailedAttempt)
				incrementedData := &entities.PinAttemptData{
					UserID:         "user123",
					FailedAttempts: 2,
					PinLockedUntil: nil,
				}
				mockRepo.On("IncrementFailedAttempts", mock.Anything, "user123").Return(incrementedData, nil)
			},
			expectError:               true,
			expectLocked:              false,
			expectedRemainingAttempts: 1, // threshold=3, newAttempts=2, remaining=1
			errorContains:             "1 attempts remaining",
		},
		{
			name: "third failed attempt - PIN gets locked",
			params: entities.PinVerifyParams{
				Username: "testuser",
				Pin:      "wrong123",
			},
			currentFailedAttempts: 2,
			mockSetup: func(mockRepo *MockAuthRepository, mockJwt *MockJwtService) {
				user := createTestUser("user123", "testuser", string(hashedPin), 2, nil, nil)
				mockRepo.On("GetUserWithPin", "testuser").Return(user, nil)

				// Mock cache data for pin attempts
				cacheData := &entities.PinAttemptData{
					UserID:         "user123",
					FailedAttempts: 2,
					PinLockedUntil: nil,
				}
				mockRepo.On("GetPinAttemptData", mock.Anything, "user123").Return(cacheData, nil)

				// Mock IncrementFailedAttempts (called by handleFailedAttempt) - will reach threshold
				incrementedData := &entities.PinAttemptData{
					UserID:         "user123",
					FailedAttempts: 3,
					PinLockedUntil: nil,
				}
				mockRepo.On("IncrementFailedAttempts", mock.Anything, "user123").Return(incrementedData, nil)

				// Mock SetPinLock (called when threshold reached)
				mockRepo.On("SetPinLock", mock.Anything, "user123", mock.AnythingOfType("time.Time"), 3, mock.Anything).Return(nil)
			},
			expectError:   true,
			expectLocked:  true,
			errorContains: "PIN locked",
		},
		{
			name: "fourth failed attempt - longer lock duration",
			params: entities.PinVerifyParams{
				Username: "testuser",
				Pin:      "wrong123",
			},
			currentFailedAttempts: 3,
			mockSetup: func(mockRepo *MockAuthRepository, mockJwt *MockJwtService) {
				user := createTestUser("user123", "testuser", string(hashedPin), 3, nil, nil)
				mockRepo.On("GetUserWithPin", "testuser").Return(user, nil)

				// Mock cache data for pin attempts
				cacheData := &entities.PinAttemptData{
					UserID:         "user123",
					FailedAttempts: 3,
					PinLockedUntil: nil,
				}
				mockRepo.On("GetPinAttemptData", mock.Anything, "user123").Return(cacheData, nil)

				// Mock IncrementFailedAttempts (called by handleFailedAttempt) - longer lock
				incrementedData := &entities.PinAttemptData{
					UserID:         "user123",
					FailedAttempts: 4,
					PinLockedUntil: nil,
				}
				mockRepo.On("IncrementFailedAttempts", mock.Anything, "user123").Return(incrementedData, nil)

				// Mock SetPinLock (called for longer duration)
				mockRepo.On("SetPinLock", mock.Anything, "user123", mock.AnythingOfType("time.Time"), 4, mock.Anything).Return(nil)
			},
			expectError:   true,
			expectLocked:  true,
			errorContains: "PIN locked",
		},
		{
			name: "many failed attempts - should cap at max duration",
			params: entities.PinVerifyParams{
				Username: "testuser",
				Pin:      "wrong123",
			},
			currentFailedAttempts: 10, // Very high number to test max duration cap
			mockSetup: func(mockRepo *MockAuthRepository, mockJwt *MockJwtService) {
				user := createTestUser("user123", "testuser", string(hashedPin), 10, nil, nil)
				mockRepo.On("GetUserWithPin", "testuser").Return(user, nil)

				// Mock cache data for pin attempts (very high failed attempts)
				cacheData := &entities.PinAttemptData{
					UserID:         "user123",
					FailedAttempts: 10,
					PinLockedUntil: nil,
				}
				mockRepo.On("GetPinAttemptData", mock.Anything, "user123").Return(cacheData, nil)

				// Mock IncrementFailedAttempts (called by handleFailedAttempt) - max duration
				incrementedData := &entities.PinAttemptData{
					UserID:         "user123",
					FailedAttempts: 11,
					PinLockedUntil: nil,
				}
				mockRepo.On("IncrementFailedAttempts", mock.Anything, "user123").Return(incrementedData, nil)

				// Mock SetPinLock (called for max duration cap)
				mockRepo.On("SetPinLock", mock.Anything, "user123", mock.AnythingOfType("time.Time"), 11, mock.Anything).Return(nil)
			},
			expectError:   true,
			expectLocked:  true,
			errorContains: "PIN locked",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockAuthRepository)
			mockJwt := new(MockJwtService)

			config := &config.Config{
				Auth: &config.AuthConfig{
					Pin: &config.PinConfig{
						BaseDuration:    10 * time.Second,
						LockThreshold:   3,
						MaxLockDuration: 300 * time.Second,
					},
				},
			}

			service := NewAuthService(mockRepo, mockJwt, config)

			// Setup mock expectations
			tt.mockSetup(mockRepo, mockJwt)

			// Act
			ctx := context.Background()
			tokenResponse, err := service.VerifyPin(ctx, tt.params)

			// Assert
			assert.Error(t, err)
			assert.Nil(t, tokenResponse)

			if tt.errorContains != "" {
				assert.Contains(t, err.Error(), tt.errorContains)
			}

			// Verify all expectations were met
			mockRepo.AssertExpectations(t)
			mockJwt.AssertExpectations(t)
		})
	}
}

func TestAuthService_VerifyPin_AlreadyLocked(t *testing.T) {
	// Create test pin hash
	hashedPin, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)

	tests := []struct {
		name          string
		params        entities.PinVerifyParams
		lockedUntil   time.Time
		mockSetup     func(*MockAuthRepository, *MockJwtService)
		expectError   bool
		errorContains string
	}{
		{
			name: "PIN is currently locked",
			params: entities.PinVerifyParams{
				Username: "testuser",
				Pin:      "123456", // Even correct PIN should fail when locked
			},
			lockedUntil: time.Now().Add(30 * time.Minute), // Locked for 30 minutes
			mockSetup: func(mockRepo *MockAuthRepository, mockJwt *MockJwtService) {
				lockedUntil := time.Now().Add(30 * time.Minute)
				user := createTestUser("user123", "testuser", string(hashedPin), 5, &lockedUntil, nil)
				mockRepo.On("GetUserWithPin", "testuser").Return(user, nil)

				// Mock cache data for pin attempts (locked)
				cacheData := &entities.PinAttemptData{
					UserID:         "user123",
					FailedAttempts: 5,
					PinLockedUntil: &lockedUntil,
				}
				mockRepo.On("GetPinAttemptData", mock.Anything, "user123").Return(cacheData, nil)

				// No other repository calls expected since PIN is locked
			},
			expectError:   true,
			errorContains: "PIN locked",
		},
		{
			name: "PIN lock has expired - should work normally",
			params: entities.PinVerifyParams{
				Username: "testuser",
				Pin:      "123456", // Correct PIN
			},
			lockedUntil: time.Now().Add(-1 * time.Minute), // Expired 1 minute ago
			mockSetup: func(mockRepo *MockAuthRepository, mockJwt *MockJwtService) {
				expiredLockTime := time.Now().Add(-1 * time.Minute)
				user := createTestUser("user123", "testuser", string(hashedPin), 3, &expiredLockTime, nil)
				mockRepo.On("GetUserWithPin", "testuser").Return(user, nil)

				// Mock cache data for pin attempts (expired lock)
				cacheData := &entities.PinAttemptData{
					UserID:         "user123",
					FailedAttempts: 3,
					PinLockedUntil: &expiredLockTime,
				}
				mockRepo.On("GetPinAttemptData", mock.Anything, "user123").Return(cacheData, nil)

				// Mock ResetPinAttempts since PIN is correct
				mockRepo.On("ResetPinAttempts", mock.Anything, "user123").Return(nil)

				// Mock token generation
				tokenResponse := &entities.TokenResponse{
					Token:        "access_token",
					RefreshToken: "refresh_token",
					Expiry:       time.Now().Add(time.Hour),
				}
				mockJwt.On("GenerateTokens", "user123", "testuser").Return(tokenResponse, nil)

				// Mock token storage
				mockRepo.On("StoreToken", mock.Anything, "user123", mock.AnythingOfType("*entities.TokenResponse")).Return(nil)
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockAuthRepository)
			mockJwt := new(MockJwtService)

			config := &config.Config{
				Auth: &config.AuthConfig{
					Pin: &config.PinConfig{
						BaseDuration:    10 * time.Second,
						LockThreshold:   3,
						MaxLockDuration: 300 * time.Second,
					},
				},
			}

			service := NewAuthService(mockRepo, mockJwt, config)

			// Setup mock expectations
			tt.mockSetup(mockRepo, mockJwt)

			// Act

			ctx := context.Background()
			tokenResponse, err := service.VerifyPin(ctx, tt.params)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, tokenResponse)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, tokenResponse)
			}

			// Verify all expectations were met
			mockRepo.AssertExpectations(t)
			mockJwt.AssertExpectations(t)
		})
	}
}

func TestAuthService_VerifyPin_RepositoryErrors_During_FailedAttempts(t *testing.T) {
	// Create test pin hash
	hashedPin, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)

	tests := []struct {
		name          string
		params        entities.PinVerifyParams
		mockSetup     func(*MockAuthRepository, *MockJwtService)
		expectError   bool
		errorContains string
	}{
		{
			name: "error getting pin attempt data from Redis",
			params: entities.PinVerifyParams{
				Username: "testuser",
				Pin:      "wrong123",
			},
			mockSetup: func(mockRepo *MockAuthRepository, mockJwt *MockJwtService) {
				user := createTestUser("user123", "testuser", string(hashedPin), 0, nil, nil)
				mockRepo.On("GetUserWithPin", "testuser").Return(user, nil)
				mockRepo.On("GetPinAttemptData", mock.Anything, "user123").Return(nil, errors.New("redis error"))
				// Mock IncrementFailedAttempts since it's called even when GetPinAttemptData fails
				mockRepo.On("IncrementFailedAttempts", mock.Anything, "user123").Return(nil, errors.New("redis error"))
			},
			expectError:   true,
			errorContains: "redis error",
		},
		{
			name: "error incrementing failed attempts in Redis",
			params: entities.PinVerifyParams{
				Username: "testuser",
				Pin:      "wrong123",
			},
			mockSetup: func(mockRepo *MockAuthRepository, mockJwt *MockJwtService) {
				user := createTestUser("user123", "testuser", string(hashedPin), 0, nil, nil)
				mockRepo.On("GetUserWithPin", "testuser").Return(user, nil)

				// Mock initial pin attempt data
				cacheData := &entities.PinAttemptData{
					UserID:         "user123",
					FailedAttempts: 0,
					PinLockedUntil: nil,
				}
				mockRepo.On("GetPinAttemptData", mock.Anything, "user123").Return(cacheData, nil)

				// Mock increment error
				mockRepo.On("IncrementFailedAttempts", mock.Anything, "user123").Return(nil, errors.New("redis increment error"))
			},
			expectError:   true,
			errorContains: "redis increment error",
		},
		{
			name: "error updating lock time when locking PIN",
			params: entities.PinVerifyParams{
				Username: "testuser",
				Pin:      "wrong123",
			},
			mockSetup: func(mockRepo *MockAuthRepository, mockJwt *MockJwtService) {
				user := createTestUser("user123", "testuser", string(hashedPin), 2, nil, nil)
				mockRepo.On("GetUserWithPin", "testuser").Return(user, nil)

				// Mock initial pin attempt data (2 failed attempts)
				cacheData := &entities.PinAttemptData{
					UserID:         "user123",
					FailedAttempts: 2,
					PinLockedUntil: nil,
				}
				mockRepo.On("GetPinAttemptData", mock.Anything, "user123").Return(cacheData, nil)

				// Mock increment to reach threshold
				incrementedData := &entities.PinAttemptData{
					UserID:         "user123",
					FailedAttempts: 3,
					PinLockedUntil: nil,
				}
				mockRepo.On("IncrementFailedAttempts", mock.Anything, "user123").Return(incrementedData, nil)

				// Mock error when setting lock
				mockRepo.On("SetPinLock", mock.Anything, "user123", mock.AnythingOfType("time.Time"), 3, mock.AnythingOfType("*time.Time")).Return(errors.New("lock update error"))
			},
			expectError:   true,
			errorContains: "PIN locked", // Should still return lock error even if update fails
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockAuthRepository)
			mockJwt := new(MockJwtService)

			config := &config.Config{
				Auth: &config.AuthConfig{
					Pin: &config.PinConfig{
						BaseDuration:    10 * time.Second,
						LockThreshold:   3,
						MaxLockDuration: 300 * time.Second,
					},
				},
			}

			service := NewAuthService(mockRepo, mockJwt, config)

			// Setup mock expectations
			tt.mockSetup(mockRepo, mockJwt)

			// Act
			ctx := context.Background()
			tokenResponse, err := service.VerifyPin(ctx, tt.params)

			// Assert
			assert.Error(t, err)
			assert.Nil(t, tokenResponse)
			if tt.errorContains != "" {
				assert.Contains(t, err.Error(), tt.errorContains)
			}

			// Verify all expectations were met
			mockRepo.AssertExpectations(t)
			mockJwt.AssertExpectations(t)
		})
	}
}

func TestAuthService_BanToken(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		mockSetup     func(*MockAuthRepository)
		expectError   bool
		errorContains string
	}{
		{
			name:   "successful token ban",
			userID: "user123",
			mockSetup: func(mockRepo *MockAuthRepository) {
				mockRepo.On("BanAllUserTokens", mock.Anything, "user123", "Manually banned by user request").Return(nil)
			},
			expectError: false,
		},
		{
			name:   "empty userID",
			userID: "",
			mockSetup: func(mockRepo *MockAuthRepository) {
				mockRepo.On("BanAllUserTokens", mock.Anything, "", "Manually banned by user request").Return(nil)
			},
			expectError: false, // Should still work with empty userID
		},
		{
			name:   "repository error during ban",
			userID: "user123",
			mockSetup: func(mockRepo *MockAuthRepository) {
				mockRepo.On("BanAllUserTokens", mock.Anything, "user123", "Manually banned by user request").Return(errors.New("redis connection failed"))
			},
			expectError:   true,
			errorContains: "redis connection failed",
		},
		{
			name:   "ban for user with no tokens",
			userID: "user_no_tokens",
			mockSetup: func(mockRepo *MockAuthRepository) {
				// Repository should handle case where user has no tokens gracefully
				mockRepo.On("BanAllUserTokens", mock.Anything, "user_no_tokens", "Manually banned by user request").Return(nil)
			},
			expectError: false,
		},
		{
			name:   "ban for non-existent user",
			userID: "nonexistent_user",
			mockSetup: func(mockRepo *MockAuthRepository) {
				// Repository should handle non-existent user case
				mockRepo.On("BanAllUserTokens", mock.Anything, "nonexistent_user", "Manually banned by user request").Return(nil)
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockAuthRepository)
			mockJwt := new(MockJwtService)

			config := &config.Config{
				Auth: &config.AuthConfig{
					Pin: &config.PinConfig{
						BaseDuration:    10 * time.Second,
						LockThreshold:   3,
						MaxLockDuration: 300 * time.Second,
					},
				},
			}

			service := NewAuthService(mockRepo, mockJwt, config)

			// Setup mock expectations
			tt.mockSetup(mockRepo)

			// Act
			ctx := context.Background()
			err := service.BanToken(ctx, tt.userID)

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
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestAuthService_ListTokens(t *testing.T) {
	tests := []struct {
		name          string
		mockSetup     func(*MockAuthRepository)
		expectError   bool
		expectTokens  int
		errorContains string
	}{
		{
			name: "successful token list",
			mockSetup: func(mockRepo *MockAuthRepository) {
				tokens := []entities.TokenResponse{
					{
						Token:        "token1",
						UserID:       "user123",
						TokenID:      "tokenid1",
						TokenVersion: time.Now().Unix(),
					},
					{
						Token:        "token2",
						UserID:       "user456",
						TokenID:      "tokenid2",
						TokenVersion: time.Now().Unix(),
					},
				}
				mockRepo.On("ListUserTokens", mock.Anything, mock.AnythingOfType("string")).Return(tokens, nil)
			},
			expectError:  false,
			expectTokens: 2,
		},
		{
			name: "empty token list",
			mockSetup: func(mockRepo *MockAuthRepository) {
				tokens := []entities.TokenResponse{}
				mockRepo.On("ListUserTokens", mock.Anything, mock.AnythingOfType("string")).Return(tokens, nil)
			},
			expectError:  false,
			expectTokens: 0,
		},
		{
			name: "repository error",
			mockSetup: func(mockRepo *MockAuthRepository) {
				mockRepo.On("ListUserTokens", mock.Anything, mock.AnythingOfType("string")).Return(nil, errors.New("redis scan failed"))
			},
			expectError:   true,
			expectTokens:  0,
			errorContains: "redis scan failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockAuthRepository)
			mockJwt := new(MockJwtService)

			config := &config.Config{
				Auth: &config.AuthConfig{
					Pin: &config.PinConfig{
						BaseDuration:    10 * time.Second,
						LockThreshold:   3,
						MaxLockDuration: 300 * time.Second,
					},
				},
			}

			service := NewAuthService(mockRepo, mockJwt, config)

			// Setup mock expectations
			tt.mockSetup(mockRepo)

			// Act
			ctx := context.Background()
			tokens, err := service.ListUserTokens(ctx, "test-user-123")

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, tokens)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, tokens)
				assert.Len(t, tokens, tt.expectTokens)
			}

			// Verify all expectations were met
			mockRepo.AssertExpectations(t)
		})
	}
}

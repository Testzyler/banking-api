package service

import (
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
				user := &models.User{
					UserID:   "user123",
					Name:     "testuser",
					Password: "hashedpassword",
					UserPin: &models.UserPin{
						UserID:            "user123",
						HashedPin:         string(hashedPin),
						FailedPinAttempts: 0,
						LastPinAttemptAt:  nil,
						PinLockedUntil:    nil,
					},
				}
				mockRepo.On("GetUserWithPin", "testuser").Return(user, nil)
				mockRepo.On("UpdateUserPinFailedAttempts", "user123", 0).Return(nil)
				mockRepo.On("UpdateUserPinLockedUntil", "user123", (*time.Time)(nil)).Return(nil)

				tokenResponse := &entities.TokenResponse{
					Token:        "access_token",
					RefreshToken: "refresh_token",
					Expiry:       time.Now().Add(time.Hour),
				}
				mockJwt.On("GenerateTokens", "user123", "testuser").Return(tokenResponse, nil)
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
				user := &models.User{
					UserID:   "user123",
					Name:     "testuser",
					Password: "hashedpassword",
					UserPin: &models.UserPin{
						UserID:            "user123",
						HashedPin:         string(hashedPin),
						FailedPinAttempts: 0,
						LastPinAttemptAt:  nil,
						PinLockedUntil:    nil,
					},
				}
				mockRepo.On("GetUserWithPin", "testuser").Return(user, nil)
				mockRepo.On("UpdateUserPinFailedAttempts", "user123", 1).Return(nil)
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
			tokenResponse, err := service.VerifyPin(tt.params)

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
		mockSetup     func(*MockJwtService)
		expectError   bool
		expectToken   bool
		errorContains string
	}{
		{
			name:         "successful token refresh",
			refreshToken: "valid_refresh_token",
			mockSetup: func(mockJwt *MockJwtService) {
				tokenResponse := &entities.TokenResponse{
					Token:        "new_access_token",
					RefreshToken: "valid_refresh_token",
					Expiry:       time.Now().Add(time.Hour),
					UserID:       "user123",
					User: entities.User{
						UserID: "user123",
						Name:   "testuser",
					},
				}
				mockJwt.On("RefreshAccessToken", "valid_refresh_token").Return(tokenResponse, nil)
			},
			expectError: false,
			expectToken: true,
		},
		{
			name:         "invalid refresh token",
			refreshToken: "invalid_token",
			mockSetup: func(mockJwt *MockJwtService) {
				mockJwt.On("RefreshAccessToken", "invalid_token").Return(nil, errors.New("invalid token"))
			},
			expectError:   true,
			expectToken:   false,
			errorContains: "invalid token",
		},
		{
			name:         "empty refresh token",
			refreshToken: "",
			mockSetup: func(mockJwt *MockJwtService) {
				mockJwt.On("RefreshAccessToken", "").Return(nil, errors.New("token is empty"))
			},
			expectError:   true,
			expectToken:   false,
			errorContains: "token is empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockAuthRepository)
			mockJwt := new(MockJwtService)

			config := &config.Config{}
			service := NewAuthService(mockRepo, mockJwt, config)

			// Setup mock expectations
			tt.mockSetup(mockJwt)

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
				user := &models.User{
					UserID:   "user123",
					Name:     "testuser",
					Password: "hashedpassword",
					UserPin: &models.UserPin{
						UserID:            "user123",
						HashedPin:         string(hashedPin),
						FailedPinAttempts: 0,
						LastPinAttemptAt:  nil,
						PinLockedUntil:    nil,
					},
				}
				mockRepo.On("GetUserWithPin", "testuser").Return(user, nil)
				mockRepo.On("UpdateUserPinFailedAttempts", "user123", 1).Return(nil)
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
				user := &models.User{
					UserID:   "user123",
					Name:     "testuser",
					Password: "hashedpassword",
					UserPin: &models.UserPin{
						UserID:            "user123",
						HashedPin:         string(hashedPin),
						FailedPinAttempts: 1,
						LastPinAttemptAt:  nil,
						PinLockedUntil:    nil,
					},
				}
				mockRepo.On("GetUserWithPin", "testuser").Return(user, nil)
				mockRepo.On("UpdateUserPinFailedAttempts", "user123", 2).Return(nil)
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
				user := &models.User{
					UserID:   "user123",
					Name:     "testuser",
					Password: "hashedpassword",
					UserPin: &models.UserPin{
						UserID:            "user123",
						HashedPin:         string(hashedPin),
						FailedPinAttempts: 2,
						LastPinAttemptAt:  nil,
						PinLockedUntil:    nil,
					},
				}
				mockRepo.On("GetUserWithPin", "testuser").Return(user, nil)
				mockRepo.On("UpdateUserPinFailedAttempts", "user123", 3).Return(nil)
				mockRepo.On("UpdateUserPinLockedUntil", "user123", mock.AnythingOfType("*time.Time")).Return(nil)
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
				user := &models.User{
					UserID:   "user123",
					Name:     "testuser",
					Password: "hashedpassword",
					UserPin: &models.UserPin{
						UserID:            "user123",
						HashedPin:         string(hashedPin),
						FailedPinAttempts: 3,
						LastPinAttemptAt:  nil,
						PinLockedUntil:    nil,
					},
				}
				mockRepo.On("GetUserWithPin", "testuser").Return(user, nil)
				mockRepo.On("UpdateUserPinFailedAttempts", "user123", 4).Return(nil)
				mockRepo.On("UpdateUserPinLockedUntil", "user123", mock.AnythingOfType("*time.Time")).Return(nil)
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
				user := &models.User{
					UserID:   "user123",
					Name:     "testuser",
					Password: "hashedpassword",
					UserPin: &models.UserPin{
						UserID:            "user123",
						HashedPin:         string(hashedPin),
						FailedPinAttempts: 10,
						LastPinAttemptAt:  nil,
						PinLockedUntil:    nil,
					},
				}
				mockRepo.On("GetUserWithPin", "testuser").Return(user, nil)
				mockRepo.On("UpdateUserPinFailedAttempts", "user123", 11).Return(nil)
				mockRepo.On("UpdateUserPinLockedUntil", "user123", mock.AnythingOfType("*time.Time")).Return(nil)
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
			tokenResponse, err := service.VerifyPin(tt.params)

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
				user := &models.User{
					UserID:   "user123",
					Name:     "testuser",
					Password: "hashedpassword",
					UserPin: &models.UserPin{
						UserID:            "user123",
						HashedPin:         string(hashedPin),
						FailedPinAttempts: 5,
						LastPinAttemptAt:  nil,
						PinLockedUntil:    &lockedUntil,
					},
				}
				mockRepo.On("GetUserWithPin", "testuser").Return(user, nil)
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
				user := &models.User{
					UserID:   "user123",
					Name:     "testuser",
					Password: "hashedpassword",
					UserPin: &models.UserPin{
						UserID:            "user123",
						HashedPin:         string(hashedPin),
						FailedPinAttempts: 3,
						LastPinAttemptAt:  nil,
						PinLockedUntil:    &expiredLockTime, // Expired
					},
				}
				mockRepo.On("GetUserWithPin", "testuser").Return(user, nil)
				mockRepo.On("UpdateUserPinFailedAttempts", "user123", 0).Return(nil)
				mockRepo.On("UpdateUserPinLockedUntil", "user123", (*time.Time)(nil)).Return(nil)

				tokenResponse := &entities.TokenResponse{
					Token:        "access_token",
					RefreshToken: "refresh_token",
					Expiry:       time.Now().Add(time.Hour),
				}
				mockJwt.On("GenerateTokens", "user123", "testuser").Return(tokenResponse, nil)
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
			tokenResponse, err := service.VerifyPin(tt.params)

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
			name: "error updating failed attempts count",
			params: entities.PinVerifyParams{
				Username: "testuser",
				Pin:      "wrong123",
			},
			mockSetup: func(mockRepo *MockAuthRepository, mockJwt *MockJwtService) {
				user := &models.User{
					UserID:   "user123",
					Name:     "testuser",
					Password: "hashedpassword",
					UserPin: &models.UserPin{
						UserID:            "user123",
						HashedPin:         string(hashedPin),
						FailedPinAttempts: 0,
						LastPinAttemptAt:  nil,
						PinLockedUntil:    nil,
					},
				}
				mockRepo.On("GetUserWithPin", "testuser").Return(user, nil)
				mockRepo.On("UpdateUserPinFailedAttempts", "user123", 1).Return(errors.New("database error"))
			},
			expectError:   true,
			errorContains: "database error",
		},
		{
			name: "error updating lock time when locking PIN",
			params: entities.PinVerifyParams{
				Username: "testuser",
				Pin:      "wrong123",
			},
			mockSetup: func(mockRepo *MockAuthRepository, mockJwt *MockJwtService) {
				user := &models.User{
					UserID:   "user123",
					Name:     "testuser",
					Password: "hashedpassword",
					UserPin: &models.UserPin{
						UserID:            "user123",
						HashedPin:         string(hashedPin),
						FailedPinAttempts: 2, // Will reach threshold
						LastPinAttemptAt:  nil,
						PinLockedUntil:    nil,
					},
				}
				mockRepo.On("GetUserWithPin", "testuser").Return(user, nil)
				mockRepo.On("UpdateUserPinFailedAttempts", "user123", 3).Return(nil)
				mockRepo.On("UpdateUserPinLockedUntil", "user123", mock.AnythingOfType("*time.Time")).Return(errors.New("lock update error"))
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
			tokenResponse, err := service.VerifyPin(tt.params)

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

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

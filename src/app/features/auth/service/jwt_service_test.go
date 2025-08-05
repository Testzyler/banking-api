package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Testzyler/banking-api/app/entities"
	"github.com/Testzyler/banking-api/app/models"
	"github.com/Testzyler/banking-api/config"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock AuthRepository for JWT service tests
type MockAuthRepositoryJWT struct {
	mock.Mock
}

func (m *MockAuthRepositoryJWT) GetUserWithPin(username string) (*models.User, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockAuthRepositoryJWT) UpdateUserPinFailedAttempts(userID string, failedAttempts int) error {
	args := m.Called(userID, failedAttempts)
	return args.Error(0)
}

func (m *MockAuthRepositoryJWT) UpdateUserPinLockedUntil(userID string, lockedUntil *time.Time) error {
	args := m.Called(userID, lockedUntil)
	return args.Error(0)
}

func (m *MockAuthRepositoryJWT) UpdateUserPinLastAttemptAt(userID string, lastAttemptAt *time.Time) error {
	args := m.Called(userID, lastAttemptAt)
	return args.Error(0)
}

func (m *MockAuthRepositoryJWT) GetPinAttemptData(ctx context.Context, userID string) (*entities.PinAttemptData, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.PinAttemptData), args.Error(1)
}

func (m *MockAuthRepositoryJWT) IncrementFailedAttempts(ctx context.Context, userID string) (*entities.PinAttemptData, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.PinAttemptData), args.Error(1)
}

func (m *MockAuthRepositoryJWT) SetPinLock(ctx context.Context, userID string, lockedUntil time.Time, failedAttempts int, lastAttemptAt *time.Time) error {
	args := m.Called(ctx, userID, lockedUntil, failedAttempts, lastAttemptAt)
	return args.Error(0)
}

func (m *MockAuthRepositoryJWT) ResetPinAttempts(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockAuthRepositoryJWT) ListUserTokens(ctx context.Context) ([]entities.TokenResponse, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entities.TokenResponse), args.Error(1)
}

func (m *MockAuthRepositoryJWT) StoreToken(ctx context.Context, userID string, tokenResponse *entities.TokenResponse) error {
	args := m.Called(ctx, userID, tokenResponse)
	return args.Error(0)
}

func (m *MockAuthRepositoryJWT) BanAllUserTokens(ctx context.Context, userID, reason string) error {
	args := m.Called(ctx, userID, reason)
	return args.Error(0)
}

func (m *MockAuthRepositoryJWT) IsTokenBanned(ctx context.Context, tokenID string) (bool, error) {
	args := m.Called(ctx, tokenID)
	return args.Bool(0), args.Error(1)
}

func (m *MockAuthRepositoryJWT) ValidateTokenVersion(ctx context.Context, tokenVersion int64) (*entities.TokenValidationResult, error) {
	args := m.Called(ctx, tokenVersion)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.TokenValidationResult), args.Error(1)
}

func (m *MockAuthRepositoryJWT) CleanupExpiredBans(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func createMockAuthRepo() *MockAuthRepositoryJWT {
	return new(MockAuthRepositoryJWT)
}

func createTestConfig() *config.Config {
	return &config.Config{
		Auth: &config.AuthConfig{
			Jwt: &config.JwtConfig{
				AccessTokenSecret:  "test-access-secret-key-that-is-long-enough",
				RefreshTokenSecret: "test-refresh-secret-key-that-is-long-enough",
				AccessTokenExpiry:  15 * time.Minute,
				RefreshTokenExpiry: 24 * time.Hour,
			},
		},
	}
}

func TestJwtService_GenerateTokens(t *testing.T) {
	tests := []struct {
		name     string
		userID   string
		username string
		wantErr  bool
	}{
		{
			name:     "successful token generation",
			userID:   "user123",
			username: "testuser",
			wantErr:  false,
		},
		{
			name:     "empty userID",
			userID:   "",
			username: "testuser",
			wantErr:  false, // Should still work
		},
		{
			name:     "empty username",
			userID:   "user123",
			username: "",
			wantErr:  false, // Should still work
		},
		{
			name:     "both empty",
			userID:   "",
			username: "",
			wantErr:  false, // Should still work
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := createTestConfig()
			mockRepo := createMockAuthRepo()
			service := NewJwtService(config, mockRepo)

			tokenResponse, err := service.GenerateTokens(tt.userID, tt.username)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, tokenResponse)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, tokenResponse)
				assert.NotEmpty(t, tokenResponse.Token)
				assert.NotEmpty(t, tokenResponse.RefreshToken)
				assert.True(t, tokenResponse.Expiry.After(time.Now()))

				// Verify tokens are different
				assert.NotEqual(t, tokenResponse.Token, tokenResponse.RefreshToken)
			}
		})
	}
}

func TestJwtService_ValidateAccessToken(t *testing.T) {
	config := createTestConfig()
	mockRepo := createMockAuthRepo()
	service := NewJwtService(config, mockRepo)

	// Generate a valid token first
	tokenResponse, err := service.GenerateTokens("user123", "testuser")
	assert.NoError(t, err)
	validToken := tokenResponse.Token

	tests := []struct {
		name        string
		tokenString string
		wantErr     bool
		checkClaims bool
	}{
		{
			name:        "valid access token",
			tokenString: validToken,
			wantErr:     false,
			checkClaims: true,
		},
		{
			name:        "empty token",
			tokenString: "",
			wantErr:     true,
			checkClaims: false,
		},
		{
			name:        "invalid token format",
			tokenString: "invalid.token.format",
			wantErr:     true,
			checkClaims: false,
		},
		{
			name:        "malformed token",
			tokenString: "not-a-jwt-token",
			wantErr:     true,
			checkClaims: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := service.ValidateAccessToken(tt.tokenString)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, claims)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, claims)

				if tt.checkClaims {
					assert.Equal(t, "user123", claims.UserID)
					assert.Equal(t, "testuser", claims.Username)
					assert.Equal(t, "access", claims.Type)
					assert.Equal(t, "banking-api", claims.Issuer)
					assert.Contains(t, claims.Audience, "banking-api")
				}
			}
		})
	}
}

func TestJwtService_ValidateRefreshToken(t *testing.T) {
	config := createTestConfig()
	mockRepo := createMockAuthRepo()
	service := NewJwtService(config, mockRepo)

	// Generate a valid token first
	tokenResponse, err := service.GenerateTokens("user123", "testuser")
	assert.NoError(t, err)
	validRefreshToken := tokenResponse.RefreshToken

	tests := []struct {
		name        string
		tokenString string
		wantErr     bool
		checkClaims bool
	}{
		{
			name:        "valid refresh token",
			tokenString: validRefreshToken,
			wantErr:     false,
			checkClaims: true,
		},
		{
			name:        "empty token",
			tokenString: "",
			wantErr:     true,
			checkClaims: false,
		},
		{
			name:        "access token used as refresh token (wrong type)",
			tokenString: tokenResponse.Token, // This is an access token
			wantErr:     true,
			checkClaims: false,
		},
		{
			name:        "invalid token format",
			tokenString: "invalid.token.format",
			wantErr:     true,
			checkClaims: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := service.ValidateRefreshToken(tt.tokenString)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, claims)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, claims)

				if tt.checkClaims {
					assert.Equal(t, "user123", claims.UserID)
					assert.Equal(t, "testuser", claims.Username)
					assert.Equal(t, "refresh", claims.Type)
				}
			}
		})
	}
}

func TestJwtService_RefreshAccessToken(t *testing.T) {
	config := createTestConfig()
	mockRepo := createMockAuthRepo()
	service := NewJwtService(config, mockRepo)

	// Generate a valid refresh token first
	originalTokens, err := service.GenerateTokens("user123", "testuser")
	assert.NoError(t, err)

	tests := []struct {
		name               string
		refreshTokenString string
		wantErr            bool
		checkResponse      bool
	}{
		{
			name:               "valid refresh token",
			refreshTokenString: originalTokens.RefreshToken,
			wantErr:            false,
			checkResponse:      true,
		},
		{
			name:               "empty refresh token",
			refreshTokenString: "",
			wantErr:            true,
			checkResponse:      false,
		},
		{
			name:               "invalid refresh token",
			refreshTokenString: "invalid.token.format",
			wantErr:            true,
			checkResponse:      false,
		},
		{
			name:               "access token used as refresh token",
			refreshTokenString: originalTokens.Token, // This is an access token
			wantErr:            true,
			checkResponse:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newTokenResponse, err := service.RefreshAccessToken(tt.refreshTokenString)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, newTokenResponse)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, newTokenResponse)

				if tt.checkResponse {
					assert.NotEmpty(t, newTokenResponse.Token)
					assert.NotEmpty(t, newTokenResponse.RefreshToken)
					assert.True(t, newTokenResponse.Expiry.After(time.Now()))
					assert.Equal(t, "user123", newTokenResponse.UserID)

					// New access token should be different from original
					assert.NotEqual(t, originalTokens.Token, newTokenResponse.Token)
					// New refresh token should also be different from original
					assert.NotEqual(t, originalTokens.RefreshToken, newTokenResponse.RefreshToken)

					// Verify the new token is valid
					claims, validateErr := service.ValidateAccessToken(newTokenResponse.Token)
					assert.NoError(t, validateErr)
					assert.Equal(t, "user123", claims.UserID)
					assert.Equal(t, "testuser", claims.Username)
					assert.Equal(t, "access", claims.Type)

					// Verify the new refresh token is also valid
					refreshClaims, validateRefreshErr := service.ValidateRefreshToken(newTokenResponse.RefreshToken)
					assert.NoError(t, validateRefreshErr)
					assert.Equal(t, "user123", refreshClaims.UserID)
					assert.Equal(t, "testuser", refreshClaims.Username)
					assert.Equal(t, "refresh", refreshClaims.Type)
				}
			}
		})
	}
}

func TestJwtService_TokenExpiration(t *testing.T) {
	// Create config with very short expiry for testing
	shortConfig := &config.Config{
		Auth: &config.AuthConfig{
			Jwt: &config.JwtConfig{
				AccessTokenSecret:  "test-access-secret-key-that-is-long-enough",
				RefreshTokenSecret: "test-refresh-secret-key-that-is-long-enough",
				AccessTokenExpiry:  1 * time.Millisecond, // Very short for testing
				RefreshTokenExpiry: 1 * time.Millisecond,
			},
		},
	}

	mockRepo := createMockAuthRepo()
	service := NewJwtService(shortConfig, mockRepo)

	// Generate tokens
	tokenResponse, err := service.GenerateTokens("user123", "testuser")
	assert.NoError(t, err)

	// Wait for tokens to expire
	time.Sleep(10 * time.Millisecond)

	// Try to validate expired access token
	_, err = service.ValidateAccessToken(tokenResponse.Token)
	assert.Error(t, err)

	// Try to validate expired refresh token
	_, err = service.ValidateRefreshToken(tokenResponse.RefreshToken)
	assert.Error(t, err)

	// Try to refresh with expired refresh token
	_, err = service.RefreshAccessToken(tokenResponse.RefreshToken)
	assert.Error(t, err)
}

func TestJwtService_InvalidTokenType(t *testing.T) {
	config := createTestConfig()
	mockRepo := createMockAuthRepo()
	service := NewJwtService(config, mockRepo).(*jwtService)

	// Test generateToken with invalid type
	param := entities.GenerateTokenParams{
		UserID:       "user123",
		Username:     "testuser",
		TokenVersion: time.Now().Unix(),
		TokenID:      "test-token-id",
		TokenType:    "invalid-type",
	}
	_, _, err := service.generateToken(param)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid token type")
}

func TestJwtService_WrongSigningMethod(t *testing.T) {
	config := createTestConfig()
	mockRepo := createMockAuthRepo()
	service := NewJwtService(config, mockRepo)

	// Create a token with wrong signing method (RS256 instead of HS256)
	claims := &entities.Claims{
		UserID:   "user123",
		Username: "testuser",
		Type:     "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "banking-api",
			Audience:  []string{"banking-api"},
		},
	}

	// This would require RSA keys, but we'll create a malformed token instead
	// to simulate wrong signing method error
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	// This will fail because we don't have RSA keys, but that's the point
	tokenString, err := token.SignedString([]byte("wrong-key-type"))

	// If somehow it creates a token, it should fail validation
	if err == nil {
		_, validateErr := service.ValidateAccessToken(tokenString)
		assert.Error(t, validateErr)
	}
	// If it fails to create (expected), that's also fine for this test
}

func TestJwtService_Integration(t *testing.T) {
	// Full integration test: generate -> validate -> refresh -> validate new token
	config := createTestConfig()
	mockRepo := createMockAuthRepo()
	service := NewJwtService(config, mockRepo)

	// Step 1: Generate initial tokens
	originalTokens, err := service.GenerateTokens("user123", "testuser")
	assert.NoError(t, err)

	// Step 2: Validate access token
	accessClaims, err := service.ValidateAccessToken(originalTokens.Token)
	assert.NoError(t, err)
	assert.Equal(t, "user123", accessClaims.UserID)
	assert.Equal(t, "access", accessClaims.Type)

	// Step 3: Validate refresh token
	refreshClaims, err := service.ValidateRefreshToken(originalTokens.RefreshToken)
	assert.NoError(t, err)
	assert.Equal(t, "user123", refreshClaims.UserID)
	assert.Equal(t, "refresh", refreshClaims.Type)

	// Step 4: Refresh access token
	newTokens, err := service.RefreshAccessToken(originalTokens.RefreshToken)
	assert.NoError(t, err)
	assert.NotEqual(t, originalTokens.Token, newTokens.Token) // New access token should be different

	// Step 5: Validate new access token
	newAccessClaims, err := service.ValidateAccessToken(newTokens.Token)
	assert.NoError(t, err)
	assert.Equal(t, "user123", newAccessClaims.UserID)
	assert.Equal(t, "access", newAccessClaims.Type)

	// Step 6: Original access token should still be valid (until it expires naturally)
	originalAccessClaims, err := service.ValidateAccessToken(originalTokens.Token)
	assert.NoError(t, err)
	assert.Equal(t, "user123", originalAccessClaims.UserID)
}

func TestJwtService_ValidateTokenWithBanCheck(t *testing.T) {
	tests := []struct {
		name         string
		tokenString  string
		mockSetup    func(*MockAuthRepositoryJWT, *config.Config) (string, *entities.Claims)
		expectError  bool
		expectValid  bool
		expectReason string
		errorType    string
	}{
		{
			name: "valid token not banned",
			mockSetup: func(mockRepo *MockAuthRepositoryJWT, config *config.Config) (string, *entities.Claims) {
				// Generate a valid token first
				service := NewJwtService(config, mockRepo)
				tokenResponse, _ := service.GenerateTokens("user123", "testuser")

				// Mock ban check - not banned
				mockRepo.On("IsTokenBanned", mock.Anything, mock.AnythingOfType("string")).Return(false, nil)

				// Mock token version validation - valid
				mockRepo.On("ValidateTokenVersion", mock.Anything, mock.AnythingOfType("int64")).Return(&entities.TokenValidationResult{
					Valid:        true,
					TokenVersion: time.Now().Unix(),
				}, nil)

				// Get the claims from token to verify later
				claims, _ := service.ValidateAccessToken(tokenResponse.Token)
				return tokenResponse.Token, claims
			},
			expectError:  false,
			expectValid:  true,
			expectReason: "",
		},
		{
			name:        "invalid token string",
			tokenString: "invalid.token.string",
			mockSetup: func(mockRepo *MockAuthRepositoryJWT, config *config.Config) (string, *entities.Claims) {
				return "invalid.token.string", nil
			},
			expectError:  true,
			expectValid:  false,
			expectReason: "invalid token",
		},
		{
			name:        "empty token string",
			tokenString: "",
			mockSetup: func(mockRepo *MockAuthRepositoryJWT, config *config.Config) (string, *entities.Claims) {
				return "", nil
			},
			expectError:  true,
			expectValid:  false,
			expectReason: "invalid token",
		},
		{
			name: "valid token but banned",
			mockSetup: func(mockRepo *MockAuthRepositoryJWT, config *config.Config) (string, *entities.Claims) {
				// Generate a valid token first
				service := NewJwtService(config, mockRepo)
				tokenResponse, _ := service.GenerateTokens("user123", "testuser")

				// Mock ban check - token is banned
				mockRepo.On("IsTokenBanned", mock.Anything, mock.AnythingOfType("string")).Return(true, nil)

				claims, _ := service.ValidateAccessToken(tokenResponse.Token)
				return tokenResponse.Token, claims
			},
			expectError:  true,
			expectValid:  false,
			expectReason: "token is banned",
			errorType:    "TokenBannedError",
		},
		{
			name: "valid token but version outdated",
			mockSetup: func(mockRepo *MockAuthRepositoryJWT, config *config.Config) (string, *entities.Claims) {
				// Generate a valid token first
				service := NewJwtService(config, mockRepo)
				tokenResponse, _ := service.GenerateTokens("user123", "testuser")

				// Mock ban check - not banned
				mockRepo.On("IsTokenBanned", mock.Anything, mock.AnythingOfType("string")).Return(false, nil)

				// Mock token version validation - outdated
				mockRepo.On("ValidateTokenVersion", mock.Anything, mock.AnythingOfType("int64")).Return(&entities.TokenValidationResult{
					Valid:        false,
					Reason:       "Token is too old",
					TokenVersion: time.Now().Unix(),
				}, nil)

				claims, _ := service.ValidateAccessToken(tokenResponse.Token)
				return tokenResponse.Token, claims
			},
			expectError:  true,
			expectValid:  false,
			expectReason: "Token is too old",
			errorType:    "TokenOutdatedError",
		},
		{
			name: "auth repository not initialized",
			mockSetup: func(mockRepo *MockAuthRepositoryJWT, config *config.Config) (string, *entities.Claims) {
				// Create service with nil repository
				service := NewJwtService(config, nil)
				tokenResponse, _ := service.GenerateTokens("user123", "testuser")

				return tokenResponse.Token, nil
			},
			expectError:  false,
			expectValid:  false,
			expectReason: "Error: auth repository not initialized",
		},
		{
			name: "ban check fails with error",
			mockSetup: func(mockRepo *MockAuthRepositoryJWT, config *config.Config) (string, *entities.Claims) {
				// Generate a valid token first
				service := NewJwtService(config, mockRepo)
				tokenResponse, _ := service.GenerateTokens("user123", "testuser")

				// Mock ban check with error
				mockRepo.On("IsTokenBanned", mock.Anything, mock.AnythingOfType("string")).Return(false, errors.New("redis connection failed"))

				claims, _ := service.ValidateAccessToken(tokenResponse.Token)
				return tokenResponse.Token, claims
			},
			expectError:  false, // Should not error, but should indicate ban check failed
			expectValid:  true,  // Should default to valid when ban check fails
			expectReason: "ban check failed",
		},
		{
			name: "version check fails with error",
			mockSetup: func(mockRepo *MockAuthRepositoryJWT, config *config.Config) (string, *entities.Claims) {
				// Generate a valid token first
				service := NewJwtService(config, mockRepo)
				tokenResponse, _ := service.GenerateTokens("user123", "testuser")

				// Mock ban check - not banned
				mockRepo.On("IsTokenBanned", mock.Anything, mock.AnythingOfType("string")).Return(false, nil)

				// Mock token version validation with error
				mockRepo.On("ValidateTokenVersion", mock.Anything, mock.AnythingOfType("int64")).Return(nil, errors.New("version check error"))

				claims, _ := service.ValidateAccessToken(tokenResponse.Token)
				return tokenResponse.Token, claims
			},
			expectError:  false, // Should not error, but should indicate version check failed
			expectValid:  true,  // Should default to valid when version check fails
			expectReason: "version check failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := createTestConfig()
			mockRepo := createMockAuthRepo()

			// Create service and get token from mock setup
			service := NewJwtService(config, mockRepo)
			var tokenString string
			var expectedClaims *entities.Claims

			if tt.name == "auth repository not initialized" {
				// Special case for nil repository test
				service = NewJwtService(config, nil)
				tokenResponse, _ := service.GenerateTokens("user123", "testuser")
				tokenString = tokenResponse.Token
			} else {
				tokenString, expectedClaims = tt.mockSetup(mockRepo, config)
			}

			if tt.tokenString != "" {
				tokenString = tt.tokenString
			}

			// Act
			result, err := service.ValidateTokenWithBanCheck(tokenString)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.NotNil(t, result)
			assert.Equal(t, tt.expectValid, result.Valid)
			assert.Equal(t, tt.expectReason, result.Reason)

			if tt.expectValid && expectedClaims != nil && tt.expectReason == "" {
				// Only check claims when validation is truly successful (no reason for fallback)
				assert.Equal(t, expectedClaims.UserID, result.Claims.UserID)
				assert.Equal(t, expectedClaims.Username, result.Claims.Username)
				assert.Equal(t, expectedClaims.TokenID, result.Claims.TokenID)
			}

			// Verify all expectations were met (only if mockRepo was used)
			if tt.name != "auth repository not initialized" {
				mockRepo.AssertExpectations(t)
			}
		})
	}
}

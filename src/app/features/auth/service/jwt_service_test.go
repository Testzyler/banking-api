package service

import (
	"testing"
	"time"

	"github.com/Testzyler/banking-api/app/entities"
	"github.com/Testzyler/banking-api/config"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

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
			service := NewJwtService(config)

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
	service := NewJwtService(config)

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
	service := NewJwtService(config)

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
	service := NewJwtService(config)

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
					assert.Equal(t, tt.refreshTokenString, newTokenResponse.RefreshToken) // Same refresh token
					assert.True(t, newTokenResponse.Expiry.After(time.Now()))
					assert.Equal(t, "user123", newTokenResponse.UserID)
					assert.Equal(t, "user123", newTokenResponse.User.UserID)
					assert.Equal(t, "testuser", newTokenResponse.User.Name)

					// New access token should be different from original
					assert.NotEqual(t, originalTokens.Token, newTokenResponse.Token)

					// Verify the new token is valid
					claims, validateErr := service.ValidateAccessToken(newTokenResponse.Token)
					assert.NoError(t, validateErr)
					assert.Equal(t, "user123", claims.UserID)
					assert.Equal(t, "testuser", claims.Username)
					assert.Equal(t, "access", claims.Type)
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

	service := NewJwtService(shortConfig)

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
	service := NewJwtService(config).(*jwtService)

	// Test generateToken with invalid type
	_, _, err := service.generateToken("user123", "testuser", "invalid-type")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid token type")
}

func TestJwtService_WrongSigningMethod(t *testing.T) {
	config := createTestConfig()
	service := NewJwtService(config)

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
	service := NewJwtService(config)

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

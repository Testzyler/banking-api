package service

import (
	"context"
	"errors"
	"time"

	"github.com/Testzyler/banking-api/app/entities"
	"github.com/Testzyler/banking-api/app/features/auth/repository"
	"github.com/Testzyler/banking-api/config"
	"github.com/Testzyler/banking-api/logger"
	"github.com/Testzyler/banking-api/server/exception"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type jwtService struct {
	config   *config.Config
	authRepo repository.AuthRepository
}

type JwtService interface {
	GenerateTokens(userID, username string) (*entities.TokenResponse, error)
	ValidateAccessToken(tokenString string) (*entities.Claims, error)
	ValidateRefreshToken(tokenString string) (*entities.Claims, error)
	RefreshAccessToken(refreshTokenString string) (*entities.TokenResponse, error)
	ValidateTokenWithBanCheck(tokenString string) (*entities.TokenValidationResult, error)
}

func NewJwtService(config *config.Config, authRepo repository.AuthRepository) JwtService {
	return &jwtService{config: config, authRepo: authRepo}
}

func (s *jwtService) GenerateTokens(userID, username string) (*entities.TokenResponse, error) {
	// Generate unique token ID and use current timestamp as token version
	tokenID := uuid.New().String()
	tokenVersion := time.Now().Unix()
	accessTokenParam := entities.GenerateTokenParams{
		UserID:       userID,
		Username:     username,
		TokenVersion: tokenVersion,
		TokenID:      tokenID,
		TokenType:    "access",
	}
	accessToken, accessExpiry, err := s.generateToken(accessTokenParam)
	if err != nil {
		return nil, err
	}

	refreshTokenID := uuid.New().String()
	refreshTokenParam := entities.GenerateTokenParams{
		UserID:       userID,
		Username:     username,
		TokenVersion: tokenVersion,
		TokenID:      refreshTokenID,
		TokenType:    "refresh",
	}
	refreshToken, _, err := s.generateToken(refreshTokenParam)
	if err != nil {
		return nil, err
	}

	return &entities.TokenResponse{
		Token:        accessToken,
		Expiry:       accessExpiry,
		RefreshToken: refreshToken,
		TokenVersion: tokenVersion,
		TokenID:      tokenID,
	}, nil
}

func (s *jwtService) generateToken(param entities.GenerateTokenParams) (string, time.Time, error) {
	var secret string
	var expiry time.Duration

	switch param.TokenType {
	case "access":
		secret = s.config.Auth.Jwt.AccessTokenSecret
		expiry = s.config.Auth.Jwt.AccessTokenExpiry
	case "refresh":
		secret = s.config.Auth.Jwt.RefreshTokenSecret
		expiry = s.config.Auth.Jwt.RefreshTokenExpiry
	default:
		return "", time.Time{}, errors.New("invalid token type")
	}

	now := time.Now()
	expiryTime := now.Add(expiry)

	claims := &entities.Claims{
		UserID:       param.UserID,
		Username:     param.Username,
		Type:         param.TokenType,
		TokenVersion: param.TokenVersion,
		TokenID:      param.TokenID,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        param.TokenID,
			Subject:   param.UserID,
			Audience:  []string{"banking-api"},
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiryTime),
			Issuer:    "banking-api",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiryTime, nil
}

func (s *jwtService) ValidateAccessToken(tokenString string) (*entities.Claims, error) {
	return s.validateToken(tokenString, s.config.Auth.Jwt.AccessTokenSecret, "access")
}

func (s *jwtService) ValidateRefreshToken(tokenString string) (*entities.Claims, error) {
	return s.validateToken(tokenString, s.config.Auth.Jwt.RefreshTokenSecret, "refresh")
}

func (s *jwtService) validateToken(tokenString, secret, expectedType string) (*entities.Claims, error) {
	if tokenString == "" {
		return nil, errors.New("token is empty")
	}

	token, err := jwt.ParseWithClaims(tokenString, &entities.Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(secret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, exception.ErrTokenExpired
		}
		return nil, err
	}

	if claims, ok := token.Claims.(*entities.Claims); ok && token.Valid {
		if claims.Type != expectedType {
			return nil, jwt.ErrInvalidType
		}
		return claims, nil
	}

	return nil, jwt.ErrTokenInvalidClaims
}

func (s *jwtService) RefreshAccessToken(refreshTokenString string) (*entities.TokenResponse, error) {
	claims, err := s.ValidateRefreshToken(refreshTokenString)
	if err != nil {
		return nil, err
	}

	if s.authRepo == nil {
		return nil, exception.ErrInternalServer
	}

	ctx := context.Background()

	// Check if the specific refresh token is banned
	isBanned, err := s.authRepo.IsTokenBanned(ctx, claims.TokenID)
	if err != nil {
		logger.Errorf("Failed to check if token is banned: %v", err)
		return nil, exception.NewTokenBannedError()
	}

	if isBanned {
		return nil, exception.NewTokenBannedError()
	}

	// Check if the user is banned (this will catch tokens issued before user ban)
	inBlacklist, err := s.authRepo.IsInBlacklist(ctx, claims.UserID, claims.TokenVersion)
	if err != nil {
		logger.Errorf("Failed to check token blacklist status: %v", err)
		return nil, exception.NewTokenBannedError()
	}

	if inBlacklist {
		logger.Infof("Rejecting refresh token for blacklisted user %s (token version: %d)", claims.UserID, claims.TokenVersion)
		return nil, exception.NewTokenBannedError()
	}

	// Check token version validity
	validationResult, err := s.authRepo.ValidateTokenVersion(ctx, claims.TokenVersion)
	if err != nil {
		logger.Errorf("Failed to validate token version: %v", err)
		return nil, exception.ErrInternalServer
	}

	if !validationResult.Valid {
		return nil, exception.NewTokenOutdatedError(validationResult.Reason)
	}

	newTokenID := uuid.New().String()
	newTokenVersion := time.Now().Unix()
	// Generate new access token
	accessTokenParam := entities.GenerateTokenParams{
		UserID:       claims.UserID,
		Username:     claims.Username,
		TokenVersion: newTokenVersion,
		TokenID:      newTokenID,
		TokenType:    "access",
	}
	accessToken, accessExpiry, err := s.generateToken(accessTokenParam)
	if err != nil {
		return nil, err
	}

	refreshTokenID := uuid.New().String()
	refreshTokenParam := entities.GenerateTokenParams{
		UserID:       claims.UserID,
		Username:     claims.Username,
		TokenVersion: newTokenVersion,
		TokenID:      refreshTokenID,
		TokenType:    "refresh",
	}
	refreshToken, _, err := s.generateToken(refreshTokenParam)
	if err != nil {
		return nil, err
	}

	return &entities.TokenResponse{
		Token:        accessToken,
		Expiry:       accessExpiry,
		RefreshToken: refreshToken,
		UserID:       claims.UserID,
		TokenVersion: newTokenVersion,
		TokenID:      newTokenID,
	}, nil
}

func (s *jwtService) ValidateTokenWithBanCheck(tokenString string) (*entities.TokenValidationResult, error) {
	claims, err := s.ValidateAccessToken(tokenString)
	if err != nil {
		return &entities.TokenValidationResult{
			Valid:        false,
			Reason:       "invalid token",
			TokenVersion: 0,
		}, err
	}

	if s.authRepo == nil {
		return &entities.TokenValidationResult{
			Valid:        false,
			Reason:       "Error: auth repository not initialized",
			TokenVersion: claims.TokenVersion,
		}, nil
	}

	ctx := context.Background()

	// Check if the specific token is banned
	isBanned, err := s.authRepo.IsTokenBanned(ctx, claims.TokenID)
	if err != nil {
		logger.Errorf("Failed to check if token is banned: %v", err)
		return &entities.TokenValidationResult{
			Valid:        true,
			Reason:       "ban check failed",
			TokenVersion: claims.TokenVersion,
		}, nil
	}

	if isBanned {
		return &entities.TokenValidationResult{
			Valid:        false,
			Reason:       "token is banned",
			TokenVersion: claims.TokenVersion,
		}, exception.NewTokenBannedError()
	}

	// Check if the token is blacklisted (this will catch tokens issued before user ban)
	inBlacklist, err := s.authRepo.IsInBlacklist(ctx, claims.UserID, claims.TokenVersion)
	if err != nil {
		logger.Errorf("Failed to check user ban status: %v", err)
		return &entities.TokenValidationResult{
			Valid:        true,
			Reason:       "token ban check failed",
			TokenVersion: claims.TokenVersion,
		}, nil
	}

	if inBlacklist {
		logger.Infof("Rejecting access token for blacklisted user %s (token version: %d)", claims.UserID, claims.TokenVersion)
		return &entities.TokenValidationResult{
			Valid:        false,
			Reason:       "token is blacklisted",
			TokenVersion: claims.TokenVersion,
		}, exception.NewTokenBannedError()
	}

	validationResult, err := s.authRepo.ValidateTokenVersion(ctx, claims.TokenVersion)
	if err != nil {
		logger.Errorf("Failed to validate token version: %v", err)
		return &entities.TokenValidationResult{
			Valid:        true,
			Reason:       "version check failed",
			TokenVersion: claims.TokenVersion,
		}, nil
	}

	if !validationResult.Valid {
		return &entities.TokenValidationResult{
			Valid:        false,
			Reason:       validationResult.Reason,
			TokenVersion: claims.TokenVersion,
		}, exception.NewTokenOutdatedError(validationResult.Reason)
	}

	return &entities.TokenValidationResult{
		Valid:        true,
		Reason:       "",
		TokenVersion: claims.TokenVersion,
		Claims:       *claims,
	}, nil
}

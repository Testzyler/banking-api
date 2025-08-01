package service

import (
	"errors"
	"time"

	"github.com/Testzyler/banking-api/app/entities"
	"github.com/Testzyler/banking-api/config"
	"github.com/Testzyler/banking-api/server/exception"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type jwtService struct {
	config *config.Config
}

type JwtService interface {
	GenerateTokens(userID, username string) (*entities.TokenResponse, error)
	ValidateAccessToken(tokenString string) (*entities.Claims, error)
	ValidateRefreshToken(tokenString string) (*entities.Claims, error)
	RefreshAccessToken(refreshTokenString string) (*entities.TokenResponse, error)
}

func NewJwtService(config *config.Config) JwtService {
	return &jwtService{config: config}
}

func (s *jwtService) GenerateTokens(userID, username string) (*entities.TokenResponse, error) {
	accessToken, accessExpiry, err := s.generateToken(userID, username, "access")
	if err != nil {
		return nil, err
	}

	refreshToken, _, err := s.generateToken(userID, username, "refresh")
	if err != nil {
		return nil, err
	}

	return &entities.TokenResponse{
		Token:        accessToken,
		Expiry:       accessExpiry,
		RefreshToken: refreshToken,
	}, nil
}

func (s *jwtService) generateToken(userID, username, tokenType string) (string, time.Time, error) {
	var secret string
	var expiry time.Duration

	switch tokenType {
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
		UserID:   userID,
		Username: username,
		Type:     tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(),
			Subject:   userID,
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

	accessToken, accessExpiry, err := s.generateToken(claims.UserID, claims.Username, "access")
	if err != nil {
		return nil, err
	}

	return &entities.TokenResponse{
		Token:        accessToken,
		Expiry:       accessExpiry,
		RefreshToken: refreshTokenString,
		UserID:       claims.UserID,
		User: entities.User{
			UserID: claims.UserID,
			Name:   claims.Username,
		},
	}, nil
}

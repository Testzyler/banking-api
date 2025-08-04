package service

import (
	"context"
	"errors"
	"time"

	"github.com/Testzyler/banking-api/app/entities"
	"github.com/Testzyler/banking-api/app/features/auth/repository"
	"github.com/Testzyler/banking-api/app/models"
	"github.com/Testzyler/banking-api/config"
	"github.com/Testzyler/banking-api/logger"
	"github.com/Testzyler/banking-api/server/exception"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type authService struct {
	config     *config.Config
	jwtService JwtService
	repository repository.AuthRepository
}

type AuthService interface {
	VerifyPin(ctx context.Context, params entities.PinVerifyParams) (*entities.TokenResponse, error)
	RefreshToken(refreshToken string) (*entities.TokenResponse, error)
	ListTokens(ctx context.Context) ([]entities.TokenResponse, error)
	BanToken(ctx context.Context, userID string) error
}

func NewAuthService(repository repository.AuthRepository, jwtService JwtService, config *config.Config) AuthService {
	return &authService{repository: repository, jwtService: jwtService, config: config}
}

func (s *authService) ListTokens(ctx context.Context) ([]entities.TokenResponse, error) {
	tokens, err := s.repository.ListUserTokens(ctx)
	if err != nil {
		return nil, err
	}
	return tokens, nil
}

func (s *authService) VerifyPin(ctx context.Context, params entities.PinVerifyParams) (*entities.TokenResponse, error) {
	user, err := s.repository.GetUserWithPin(params.Username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, exception.ErrUserNotFound
		}
		return nil, err
	}

	now := time.Now()
	// Check Redis cache
	cacheData, err := s.repository.GetPinAttemptData(ctx, user.UserID)
	if err != nil {
		logger.Errorf("Failed to get cache data for user %s: %v", user.UserID, err)
		cacheData = &entities.PinAttemptData{
			UserID:         user.UserID,
			FailedAttempts: user.UserPin.FailedPinAttempts,
			PinLockedUntil: user.UserPin.PinLockedUntil,
			LastAttemptAt:  user.UserPin.LastPinAttemptAt,
		}
	}

	if isLocked, remainingTime := isPinLocked(cacheData, now); isLocked {
		return nil, exception.NewPinLockedError(remainingTime.String())
	}

	if !isPinCorrect(user.UserPin.HashedPin, params.Pin) {
		return nil, s.handleFailedAttempt(ctx, user, now)
	}

	if err := s.repository.ResetPinAttempts(ctx, user.UserID); err != nil {
		logger.Errorf("Failed to reset cache attempts for user %s: %v", user.UserID, err)
	}

	// Generate JWT tokens with token version (timestamp)
	tokenResponse, err := s.jwtService.GenerateTokens(user.UserID, params.Username)
	if err != nil {
		return nil, exception.NewInternalError(err)
	}

	// Store token in Redis for tracking
	tokenResponse.UserID = user.UserID
	if err := s.repository.StoreToken(ctx, user.UserID, tokenResponse); err != nil {
		logger.Errorf("Failed to store token in Redis for user %s: %v", user.UserID, err)
	}

	return tokenResponse, nil
}

func isPinLocked(cacheData *entities.PinAttemptData, now time.Time) (bool, time.Duration) {
	if cacheData.PinLockedUntil != nil && now.Before(*cacheData.PinLockedUntil) {
		remainingTime := cacheData.PinLockedUntil.Sub(now).Round(time.Second)
		return true, remainingTime
	}
	return false, 0
}

func isPinCorrect(hashedPin, inputPin string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashedPin), []byte(inputPin)) == nil
}

func (s *authService) handleFailedAttempt(ctx context.Context, user *models.User, now time.Time) error {
	cacheData, err := s.repository.IncrementFailedAttempts(ctx, user.UserID)
	if err != nil {
		return err
	}

	baseDuration := s.config.Auth.Pin.BaseDuration
	lockThreshold := s.config.Auth.Pin.LockThreshold
	maxLockDuration := s.config.Auth.Pin.MaxLockDuration

	if cacheData.FailedAttempts >= lockThreshold {
		power := cacheData.FailedAttempts - lockThreshold
		lockDuration := baseDuration * time.Duration(1<<power) // 2^power

		// Cap max lock duration
		if lockDuration > maxLockDuration {
			lockDuration = maxLockDuration
		}

		lockedUntil := now.Add(lockDuration)

		// Set lock in Redis immediately, then async DB sync
		if err := s.repository.SetPinLock(ctx, user.UserID, lockedUntil, cacheData.FailedAttempts, cacheData.LastAttemptAt); err != nil {
			logger.Errorf("Failed to set pin lock in cache for user %s: %v", user.UserID, err)
		}

		return exception.NewPinLockedError(lockDuration.String())
	}

	remainingAttempts := lockThreshold - cacheData.FailedAttempts
	return exception.NewInvalidPinError(remainingAttempts)
}

func (s *authService) RefreshToken(refreshToken string) (*entities.TokenResponse, error) {
	tokenResponse, err := s.jwtService.RefreshAccessToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// Store the new token in Redis for tracking
	if tokenResponse.UserID != "" {
		ctx := context.Background()
		if err := s.repository.StoreToken(ctx, tokenResponse.UserID, tokenResponse); err != nil {
			logger.Errorf("Failed to store refreshed token in Redis for user %s: %v", tokenResponse.UserID, err)
		}
	}

	return tokenResponse, nil
}

func (s *authService) BanToken(ctx context.Context, userID string) error {
	reason := "Manually banned by admin"
	if err := s.repository.BanAllUserTokens(ctx, userID, reason); err != nil {
		return err
	}

	logger.Infof("User %s has been banned all tokens successfully", userID)
	return nil
}

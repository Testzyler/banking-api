package service

import (
	"errors"
	"time"

	"github.com/Testzyler/banking-api/app/entities"
	"github.com/Testzyler/banking-api/app/features/auth/repository"
	"github.com/Testzyler/banking-api/app/models"
	"github.com/Testzyler/banking-api/config"
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
	VerifyPin(params entities.PinVerifyParams) (*entities.TokenResponse, error)
	RefreshToken(refreshToken string) (*entities.TokenResponse, error)
}

func NewAuthService(repository repository.AuthRepository, jwtService JwtService, config *config.Config) AuthService {
	return &authService{repository: repository, jwtService: jwtService, config: config}
}

func (s *authService) VerifyPin(params entities.PinVerifyParams) (*entities.TokenResponse, error) {
	user, err := s.repository.GetUserWithPin(params.Username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, exception.ErrUserNotFound
		}
		return nil, err
	}

	now := time.Now()

	if isPinLocked(user.UserPin, now) {
		remainingTime := user.UserPin.PinLockedUntil.Sub(now).Round(time.Second)
		return nil, exception.NewPinLockedError(remainingTime.String())
	}

	if isPinExpired(user.UserPin, now) {
		return nil, exception.ErrPinExpired
	}

	if !isPinCorrect(user.UserPin.HashedPin, params.Pin) {
		return nil, s.handleFailedAttempt(user, now)
	}

	// Success: reset failed attempts and lock
	_ = s.repository.UpdateUserPinFailedAttempts(user.UserID, 0)
	_ = s.repository.UpdateUserPinLockedUntil(user.UserID, nil)

	// Generate JWT tokens
	tokenResponse, err := s.jwtService.GenerateTokens(user.UserID, params.Username)
	if err != nil {
		return nil, exception.NewInternalError(err)
	}

	return tokenResponse, nil
}

func isPinLocked(pin *models.UserPin, now time.Time) bool {
	return pin.PinLockedUntil != nil && now.Before(*pin.PinLockedUntil)
}

func isPinExpired(pin *models.UserPin, now time.Time) bool {
	return pin.LastPinAttemptAt != nil && now.After(*pin.LastPinAttemptAt)
}

func isPinCorrect(hashedPin, inputPin string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashedPin), []byte(inputPin)) == nil
}

func (s *authService) handleFailedAttempt(user *models.User, now time.Time) error {
	newAttempts := user.UserPin.FailedPinAttempts + 1
	if err := s.repository.UpdateUserPinFailedAttempts(user.UserID, newAttempts); err != nil {
		return exception.NewInternalError(err)
	}

	// Configurable
	// const (
	// 	baseDuration   = 10 * time.Second
	// 	lockThreshold  = 3
	// 	maxLockSeconds = 300
	// )

	baseDuration := s.config.Auth.Pin.BaseDuration
	lockThreshold := s.config.Auth.Pin.LockThreshold
	maxLockDuration := s.config.Auth.Pin.MaxLockDuration

	if newAttempts >= lockThreshold {
		power := newAttempts - lockThreshold
		lockDuration := baseDuration * time.Duration(1<<power) // 2^power

		// Cap max lock duration
		if lockDuration > maxLockDuration*time.Second {
			lockDuration = maxLockDuration * time.Second
		}

		lockedUntil := now.Add(lockDuration)
		_ = s.repository.UpdateUserPinLockedUntil(user.UserID, &lockedUntil)

		// PIN is now locked
		return exception.NewPinLockedError(lockDuration.String())
	}

	// Return error with remaining attempts
	remainingAttempts := lockThreshold - newAttempts
	return exception.NewInvalidPinError(remainingAttempts)
}

func (s *authService) RefreshToken(refreshToken string) (*entities.TokenResponse, error) {
	tokenResponse, err := s.jwtService.RefreshAccessToken(refreshToken)
	if err != nil {
		return nil, err
	}
	return tokenResponse, nil
}

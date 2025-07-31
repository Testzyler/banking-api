package repository

import (
	"time"

	"github.com/Testzyler/banking-api/app/models"
	"gorm.io/gorm"
)

type authRepository struct {
	db *gorm.DB
}

type AuthRepository interface {
	GetUserWithPin(username string) (*models.User, error)
	UpdateUserPinFailedAttempts(userID string, failedAttempts int) error
	UpdateUserPinLockedUntil(userID string, lockedUntil *time.Time) error
}

func NewAuthRepository(db *gorm.DB) AuthRepository {
	return &authRepository{db: db}
}

func (r *authRepository) GetUserWithPin(username string) (*models.User, error) {
	var user models.User

	err := r.db.
		Preload("UserPin").
		Where("name = ?", username).
		First(&user).Error

	return &user, err
}

func (r *authRepository) UpdateUserPinFailedAttempts(userID string, failedAttempts int) error {
	return r.db.Model(&models.UserPin{}).Where("user_id = ?", userID).Update("failed_pin_attempts", failedAttempts).Error
}

func (r *authRepository) UpdateUserPinLockedUntil(userID string, lockedUntil *time.Time) error {
	return r.db.Model(&models.UserPin{}).Where("user_id = ?", userID).Update("pin_locked_until", lockedUntil).Error
}

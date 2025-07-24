package repository

import (
	"fmt"

	models "github.com/Testzyler/banking-api/app/.models"
	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}
type UserRepository interface {
	GetByID(userID string) (*models.User, error)
	GetAll(perPage, page int, search string) ([]*models.User, error)
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) GetByID(userID string) (*models.User, error) {
	var user models.User
	err := r.db.Where("user_id = ?", userID).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("error fetching user: %w", err)
	}

	return &user, nil
}

func (r *userRepository) GetAll(perPage, page int, search string) ([]*models.User, error) {
	var users []*models.User
	err := r.db.Order("name ASC").Limit(perPage).Offset((page - 1) * perPage).Find(&users).Error
	if err != nil {
		return nil, fmt.Errorf("error fetching users: %w", err)
	}

	return users, nil
}

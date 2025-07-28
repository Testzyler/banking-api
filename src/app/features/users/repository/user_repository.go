package repository

import (
	models "github.com/Testzyler/banking-api/app/models"
	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}
type UserRepository interface {
	GetByID(userID string) (*models.User, error)
	GetAll(perPage, page int, search string) ([]*models.User, error)
	GetAllWithCount(perPage, page int, search string) ([]*models.User, int64, error)
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) GetByID(userID string) (*models.User, error) {
	var user models.User
	err := r.db.Where("user_id = ?", userID).First(&user).Error
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) GetAll(perPage, page int, search string) ([]*models.User, error) {
	var users []*models.User
	err := r.db.Order("name ASC").Limit(perPage).Offset((page - 1) * perPage).Find(&users).Error
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (r *userRepository) GetAllWithCount(perPage, page int, search string) ([]*models.User, int64, error) {
	var users []*models.User
	var total int64

	query := r.db.Model(&models.User{})

	// Apply search filter if provided
	if search != "" {
		query = query.Where("name LIKE ? OR email LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := query.Order("name ASC").Limit(perPage).Offset((page - 1) * perPage).Find(&users).Error
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

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

	// Build the query for data retrieval
	dataQuery := r.db.Model(&models.User{})

	// Apply search filter if provided
	if search != "" {
		dataQuery = dataQuery.Where("name LIKE ?", "%"+search+"%")
	}

	// Calculate offset - ensure it's never negative and always specified
	offset := (page - 1) * perPage
	if offset < 0 {
		offset = 0
	}

	// Get paginated results first (to match test expectations) - Always specify both Limit and Offset
	err := dataQuery.Order("name ASC").Limit(perPage).Offset(offset).Find(&users).Error
	if err != nil {
		return nil, 0, err
	}

	// Build a separate query for count
	countQuery := r.db.Model(&models.User{})

	// Apply the same search filter for count
	if search != "" {
		countQuery = countQuery.Where("name LIKE ?", "%"+search+"%")
	}

	// Get total count
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

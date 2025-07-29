package service

import (
	"math"

	entities "github.com/Testzyler/banking-api/app/entities"
	"github.com/Testzyler/banking-api/app/features/users/repository"
	models "github.com/Testzyler/banking-api/app/models"
	"github.com/Testzyler/banking-api/server/exception"
	"gorm.io/gorm"
)

type userService struct {
	userRepo repository.UserRepository
}

type UserService interface {
	GetUserByID(params entities.GetUserByIdParams) (*models.User, error)
	GetAllUsers(params entities.PaginationParams) ([]*models.User, error)
	GetAllUsersWithMeta(params entities.PaginationParams) ([]*models.User, entities.PaginationMeta, error)
}

func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{
		userRepo: userRepo,
	}
}

func (s *userService) GetUserByID(params entities.GetUserByIdParams) (*models.User, error) {
	user, err := s.userRepo.GetByID(params.UserID)
	if err != nil {
		// Convert to business exception
		if err == gorm.ErrRecordNotFound {
			return nil, exception.NewUserNotFoundError(params.UserID)
		}
		return nil, exception.NewInternalError(err)
	}

	return user, nil
}

func (s *userService) GetAllUsers(paginationParams entities.PaginationParams) ([]*models.User, error) {
	users, err := s.userRepo.GetAll(paginationParams.PerPage, paginationParams.Page, paginationParams.Search)
	if err != nil {
		return nil, exception.NewInternalError(err)
	}

	return users, nil
}

func (s *userService) GetAllUsersWithMeta(paginationParams entities.PaginationParams) ([]*models.User, entities.PaginationMeta, error) {
	users, total, err := s.userRepo.GetAllWithCount(paginationParams.PerPage, paginationParams.Page, paginationParams.Search)
	if err != nil {
		return nil, entities.PaginationMeta{}, exception.NewInternalError(err)
	}

	// Calculate pagination metadata
	totalPages := int(math.Ceil(float64(total) / float64(paginationParams.PerPage)))
	hasNext := paginationParams.Page < totalPages
	hasPrevious := paginationParams.Page > 1

	meta := entities.PaginationMeta{
		Page:        paginationParams.Page,
		PerPage:     paginationParams.PerPage,
		Total:       int(total),
		TotalPages:  totalPages,
		HasNext:     hasNext,
		HasPrevious: hasPrevious,
	}

	return users, meta, nil
}

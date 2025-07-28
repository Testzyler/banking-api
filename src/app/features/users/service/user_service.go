package service

import (
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
		return nil, exception.NewInternalError("user_service.GetUserByID")
	}

	return user, nil
}

func (s *userService) GetAllUsers(paginationParams entities.PaginationParams) ([]*models.User, error) {
	users, err := s.userRepo.GetAll(paginationParams.PerPage, paginationParams.Page, paginationParams.Search)
	if err != nil {
		return nil, exception.NewInternalError("user_service.GetAllUsers")
	}

	return users, nil
}

package service

import (
	entities "github.com/Testzyler/banking-api/app/.entities"
	models "github.com/Testzyler/banking-api/app/.models"
	"github.com/Testzyler/banking-api/app/features/users/repository"
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
	return s.userRepo.GetByID(params.UserID)
}

func (s *userService) GetAllUsers(paginationParams entities.PaginationParams) ([]*models.User, error) {
	return s.userRepo.GetAll(paginationParams.PerPage, paginationParams.Page, paginationParams.Search)
}

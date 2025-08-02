package service

import (
	"github.com/Testzyler/banking-api/app/entities"
	"github.com/Testzyler/banking-api/app/features/home/repository"
)

type homeService struct {
	repo repository.HomeRepository
}

type HomeService interface {
	GetHomeData(userID string) (entities.HomeResponse, error)
}

func NewHomeService(repo repository.HomeRepository) *homeService {
	return &homeService{
		repo: repo,
	}
}

func (s *homeService) GetHomeData(userID string) (entities.HomeResponse, error) {
	homeData, err := s.repo.GetHomeData(userID)
	if err != nil {
		return entities.HomeResponse{}, err
	}

	return homeData, nil
}

package service

import (
	"github.com/Testzyler/banking-api/app/entities"
	"github.com/Testzyler/banking-api/app/features/dashboard/repository"
)

type dashboardService struct {
	repo repository.DashboardRepository
}

type DashboardService interface {
	GetDashboardData(userID string) (entities.DashboardResponse, error)
}

func NewDashboardService(repo repository.DashboardRepository) *dashboardService {
	return &dashboardService{
		repo: repo,
	}
}

func (s *dashboardService) GetDashboardData(userID string) (entities.DashboardResponse, error) {
	dashboard, err := s.repo.GetDashboardData(userID)
	if err != nil {
		return entities.DashboardResponse{}, err
	}

	return dashboard, nil
}

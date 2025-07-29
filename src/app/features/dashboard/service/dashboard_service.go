package service

import (
	"github.com/Testzyler/banking-api/app/entities"
	"github.com/Testzyler/banking-api/app/features/dashboard/repository"
	"github.com/shopspring/decimal"
)

type dashboardService struct {
	repo repository.DashboardRepository
}

type DashboardService interface {
	GetDashboardData(entities.DashboardParams) (entities.DashboardResponse, error)
}

func NewDashboardService(repo repository.DashboardRepository) *dashboardService {
	return &dashboardService{
		repo: repo,
	}
}

func (s *dashboardService) GetDashboardData(params entities.DashboardParams) (entities.DashboardResponse, error) {
	user, err := s.repo.GetUserByID(params.UserID)
	if err != nil {
		return entities.DashboardResponse{}, err
	}

	cards, err := s.repo.GetCardsByUserID(params.UserID)
	if err != nil {
		return entities.DashboardResponse{}, err
	}

	banners, err := s.repo.GetBannersByUserID(params.UserID)
	if err != nil {
		return entities.DashboardResponse{}, err
	}

	transactions, err := s.repo.GetTransactionsByUserID(params.UserID)
	if err != nil {
		return entities.DashboardResponse{}, err
	}

	accounts, err := s.repo.GetAccountsByUserID(params.UserID)
	if err != nil {
		return entities.DashboardResponse{}, err
	}

	data := entities.DashboardResponse{
		User:         user,
		DebitCards:   cards,
		Banners:      banners,
		Transactions: transactions,
		Accounts:     accounts,
		TotalBalance: getTotalBalance(accounts),
	}
	return data, nil
}

func getTotalBalance(accounts []entities.Account) float64 {
	totalBalance := decimal.Zero
	for _, account := range accounts {
		totalBalance = totalBalance.Add(decimal.NewFromFloat(account.Amount))
	}
	return totalBalance.InexactFloat64()
}

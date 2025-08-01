package service

import (
	"errors"
	"testing"

	"github.com/Testzyler/banking-api/app/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// Mock DashboardRepository
type MockDashboardRepository struct {
	mock.Mock
}

func (m *MockDashboardRepository) GetAccountsByUserID(userID string) ([]entities.Account, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entities.Account), args.Error(1)
}

func (m *MockDashboardRepository) GetTransactionsByUserID(userID string) ([]entities.Transaction, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entities.Transaction), args.Error(1)
}

func (m *MockDashboardRepository) GetBannersByUserID(userID string) ([]entities.Banner, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entities.Banner), args.Error(1)
}

func (m *MockDashboardRepository) GetUserByID(userID string) (entities.User, error) {
	args := m.Called(userID)
	return args.Get(0).(entities.User), args.Error(1)
}

func (m *MockDashboardRepository) GetCardsByUserID(userID string) ([]entities.DebitCards, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entities.DebitCards), args.Error(1)
}

func (m *MockDashboardRepository) GetTotalBalance(userID string) float64 {
	args := m.Called(userID)
	return args.Get(0).(float64)
}

func (m *MockDashboardRepository) GetDashboardDataWithTrx(userID string) (entities.DashboardResponse, error) {
	args := m.Called(userID)
	return args.Get(0).(entities.DashboardResponse), args.Error(1)
}

func TestDashboardService_GetDashboardData(t *testing.T) {
	tests := []struct {
		name          string
		params        entities.DashboardParams
		mockSetup     func(*MockDashboardRepository)
		expectError   bool
		expectData    bool
		errorContains string
	}{
		{
			name: "successful get dashboard data",
			params: entities.DashboardParams{
				UserID: "user123",
			},
			mockSetup: func(mockRepo *MockDashboardRepository) {
				user := entities.User{
					UserID:   "user123",
					Name:     "testuser",
					Greeting: "Hello, testuser!",
				}
				accounts := []entities.Account{
					{
						AccountID: "acc1",
						Type:      "savings",
						Amount:    5000.0,
					},
				}
				cards := []entities.DebitCards{
					{
						CardID:   "card1",
						CardName: "Main Card",
					},
				}
				banners := []entities.Banner{
					{
						BannerID: "banner1",
						Title:    "Welcome",
					},
				}
				transactions := []entities.Transaction{
					{
						TransactionID: "txn1",
						Name:          "Coffee Shop",
					},
				}
				totalBalance := 5000.0

				mockRepo.On("GetUserByID", "user123").Return(user, nil)
				mockRepo.On("GetCardsByUserID", "user123").Return(cards, nil)
				mockRepo.On("GetBannersByUserID", "user123").Return(banners, nil)
				mockRepo.On("GetTransactionsByUserID", "user123").Return(transactions, nil)
				mockRepo.On("GetAccountsByUserID", "user123").Return(accounts, nil)
				mockRepo.On("GetTotalBalance", "user123").Return(totalBalance)
			},
			expectError: false,
			expectData:  true,
		},
		{
			name: "user not found",
			params: entities.DashboardParams{
				UserID: "nonexistent",
			},
			mockSetup: func(mockRepo *MockDashboardRepository) {
				mockRepo.On("GetUserByID", "nonexistent").Return(entities.User{}, gorm.ErrRecordNotFound)
			},
			expectError:   true,
			expectData:    false,
			errorContains: "record not found",
		},
		{
			name: "error getting cards",
			params: entities.DashboardParams{
				UserID: "user123",
			},
			mockSetup: func(mockRepo *MockDashboardRepository) {
				user := entities.User{
					UserID: "user123",
					Name:   "testuser",
				}
				mockRepo.On("GetUserByID", "user123").Return(user, nil)
				mockRepo.On("GetCardsByUserID", "user123").Return(nil, errors.New("database error"))
			},
			expectError:   true,
			expectData:    false,
			errorContains: "database error",
		},
		{
			name: "error getting banners",
			params: entities.DashboardParams{
				UserID: "user123",
			},
			mockSetup: func(mockRepo *MockDashboardRepository) {
				user := entities.User{UserID: "user123", Name: "testuser"}
				cards := []entities.DebitCards{{CardID: "card1"}}

				mockRepo.On("GetUserByID", "user123").Return(user, nil)
				mockRepo.On("GetCardsByUserID", "user123").Return(cards, nil)
				mockRepo.On("GetBannersByUserID", "user123").Return(nil, errors.New("database error"))
			},
			expectError:   true,
			expectData:    false,
			errorContains: "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockDashboardRepository)
			service := NewDashboardService(mockRepo)

			// Setup mock expectations
			tt.mockSetup(mockRepo)

			// Act
			data, err := service.GetDashboardData(tt.params)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				if tt.expectData {
					assert.NotEmpty(t, data.UserID)
					assert.NotEmpty(t, data.Name)
					assert.NotNil(t, data.DebitCards)
					assert.NotNil(t, data.Banners)
					assert.NotNil(t, data.Transactions)
					assert.NotNil(t, data.Accounts)
					assert.GreaterOrEqual(t, data.TotalBalance, 0.0)
				}
			}

			// Verify all expectations were met
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestDashboardService_GetDashboardDataWithTrx(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		mockSetup     func(*MockDashboardRepository)
		expectError   bool
		expectData    bool
		errorContains string
	}{
		{
			name:   "successful get dashboard data with transaction",
			userID: "user123",
			mockSetup: func(mockRepo *MockDashboardRepository) {
				dashboardData := entities.DashboardResponse{
					User: entities.User{
						UserID:   "user123",
						Name:     "testuser",
						Greeting: "Hello!",
					},
					DebitCards: []entities.DebitCards{
						{
							CardID:   "card1",
							CardName: "Main Card",
						},
					},
					Banners: []entities.Banner{
						{
							BannerID: "banner1",
							Title:    "Welcome",
						},
					},
					Transactions: []entities.Transaction{
						{
							TransactionID: "txn1",
							Name:          "Coffee Shop",
						},
					},
					Accounts: []entities.Account{
						{
							AccountID: "acc1",
							Type:      "savings",
							Amount:    5000.0,
						},
					},
					TotalBalance: 5000.0,
				}
				mockRepo.On("GetDashboardDataWithTrx", "user123").Return(dashboardData, nil)
			},
			expectError: false,
			expectData:  true,
		},
		{
			name:   "database transaction error",
			userID: "user123",
			mockSetup: func(mockRepo *MockDashboardRepository) {
				mockRepo.On("GetDashboardDataWithTrx", "user123").Return(entities.DashboardResponse{}, errors.New("transaction failed"))
			},
			expectError:   true,
			expectData:    false,
			errorContains: "transaction failed",
		},
		{
			name:   "user not found in transaction",
			userID: "nonexistent",
			mockSetup: func(mockRepo *MockDashboardRepository) {
				mockRepo.On("GetDashboardDataWithTrx", "nonexistent").Return(entities.DashboardResponse{}, gorm.ErrRecordNotFound)
			},
			expectError:   true,
			expectData:    false,
			errorContains: "record not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockDashboardRepository)
			service := NewDashboardService(mockRepo)

			// Setup mock expectations
			tt.mockSetup(mockRepo)

			// Act
			data, err := service.GetDashboardDataWithTrx(tt.userID)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				if tt.expectData {
					assert.NotEmpty(t, data.UserID)
					assert.NotEmpty(t, data.Name)
					assert.NotNil(t, data.DebitCards)
					assert.NotNil(t, data.Banners)
					assert.NotNil(t, data.Transactions)
					assert.NotNil(t, data.Accounts)
					assert.GreaterOrEqual(t, data.TotalBalance, 0.0)
				}
			}

			// Verify all expectations were met
			mockRepo.AssertExpectations(t)
		})
	}
}

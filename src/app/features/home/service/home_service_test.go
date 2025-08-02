package service

import (
	"errors"
	"testing"

	"github.com/Testzyler/banking-api/app/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// Mock HomeRepository
type MockHomeRepository struct {
	mock.Mock
}

// GetTotalBalance implements repository.HomeRepository.
func (m *MockHomeRepository) GetTotalBalance(userID string) float64 {
	panic("unimplemented")
}

func (m *MockHomeRepository) GetHomeData(userID string) (entities.HomeResponse, error) {
	args := m.Called(userID)
	return args.Get(0).(entities.HomeResponse), args.Error(1)
}

func TestHomeService_GetHomeData(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		mockSetup     func(*MockHomeRepository)
		expectError   bool
		expectData    bool
		errorContains string
	}{
		{
			name:   "successful get home data",
			userID: "user123",
			mockSetup: func(mockRepo *MockHomeRepository) {
				homeData := entities.HomeResponse{
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
				mockRepo.On("GetHomeData", "user123").Return(homeData, nil)
			},
			expectError: false,
			expectData:  true,
		},
		{
			name:   "database transaction error",
			userID: "user123",
			mockSetup: func(mockRepo *MockHomeRepository) {
				mockRepo.On("GetHomeData", "user123").Return(entities.HomeResponse{}, errors.New("transaction failed"))
			},
			expectError:   true,
			expectData:    false,
			errorContains: "transaction failed",
		},
		{
			name:   "user not found in transaction",
			userID: "nonexistent",
			mockSetup: func(mockRepo *MockHomeRepository) {
				mockRepo.On("GetHomeData", "nonexistent").Return(entities.HomeResponse{}, gorm.ErrRecordNotFound)
			},
			expectError:   true,
			expectData:    false,
			errorContains: "record not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockHomeRepository)
			service := NewHomeService(mockRepo)

			// Setup mock expectations
			tt.mockSetup(mockRepo)

			// Act
			data, err := service.GetHomeData(tt.userID)

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

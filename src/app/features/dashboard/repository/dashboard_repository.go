package repository

import (
	"github.com/Testzyler/banking-api/app/entities"
	"github.com/Testzyler/banking-api/app/models"
	"gorm.io/gorm"
)

type dashboardRepository struct {
	db *gorm.DB
}

type DashboardRepository interface {
	GetTotalBalance(userID string) float64
	GetDashboardData(userID string) (entities.DashboardResponse, error)
}

func NewDashboardRepository(repo *gorm.DB) DashboardRepository {
	return &dashboardRepository{
		db: repo,
	}
}

func (r *dashboardRepository) GetTotalBalance(userID string) float64 {
	var total float64
	if err := r.db.Model(&models.AccountBalance{}).
		Select("SUM(amount)").
		Joins("JOIN accounts ON account_balances.account_id = accounts.account_id").
		Where("accounts.user_id = ?", userID).
		Scan(&total).Error; err != nil {
		return 0
	}
	return total
}

func (r *dashboardRepository) GetDashboardData(userID string) (entities.DashboardResponse, error) {
	var response entities.DashboardResponse

	err := r.db.Transaction(func(tx *gorm.DB) error {
		// User + Greeting
		var user models.User
		if err := tx.Preload("UserGreeting").First(&user, "user_id = ?", userID).Error; err != nil {
			return err
		}
		response.UserID = user.UserID
		response.Name = user.Name
		if user.UserGreeting != nil {
			response.Greeting = user.UserGreeting.Greeting
		}

		// Debit Cards
		var cards []models.DebitCard
		if err := tx.Preload("DebitCardDetail").
			Preload("DebitCardDesign").
			Preload("DebitCardStatus").
			Where("user_id = ?", userID).
			Find(&cards).Error; err != nil {
			return err
		}
		for _, c := range cards {
			response.DebitCards = append(response.DebitCards, entities.DebitCards{
				CardID:   c.CardID,
				CardName: c.Name,
				DebitCardDesign: entities.DebitCardDesign{
					Color:       c.DebitCardDesign.Color,
					BorderColor: c.DebitCardDesign.BorderColor,
				},
				Status:     c.DebitCardStatus.Status,
				CardNumber: c.DebitCardDetail.Number,
				Issuer:     c.DebitCardDetail.Issuer,
			})
		}

		// Banners
		var banners []models.Banner
		if err := tx.Find(&banners, "user_id = ?", userID).Error; err != nil {
			return err
		}
		for _, b := range banners {
			response.Banners = append(response.Banners, entities.Banner{
				BannerID:    b.BannerID,
				UserID:      b.UserID,
				Title:       b.Title,
				Description: b.Description,
				ImageURL:    b.Image,
			})
		}

		// Transactions
		var transactions []models.Transaction
		if err := tx.Find(&transactions, "user_id = ?", userID).Error; err != nil {
			return err
		}
		for _, t := range transactions {
			response.Transactions = append(response.Transactions, entities.Transaction{
				TransactionID: t.TransactionID,
				UserID:        t.UserID,
				Name:          t.Name,
				Image:         t.Image,
				IsBank:        t.IsBank,
			})
		}

		// Accounts + preload related
		var accounts []models.Account
		if err := tx.Preload("AccountDetails").
			Preload("AccountBalance").
			Preload("AccountFlags").
			Where("user_id = ?", userID).
			Find(&accounts).Error; err != nil {
			return err
		}

		total := 0.0
		for _, acc := range accounts {
			var flags []entities.AccountFlags
			for _, f := range acc.AccountFlags {
				flags = append(flags, entities.AccountFlags{
					FlagType:  f.FlagType,
					FlagValue: f.FlagValue,
					CreatedAt: f.CreatedAt,
					UpdatedAt: f.UpdatedAt,
				})
			}
			response.Accounts = append(response.Accounts, entities.Account{
				AccountID: acc.AccountID,
				Type:      acc.Type,
				Currency:  acc.Currency,
				Issuer:    acc.Issuer,
				Amount:    acc.AccountBalance.Amount,
				AccountDetails: entities.AccountDetails{
					Color:         acc.AccountDetails.Color,
					IsMainAccount: acc.AccountDetails.IsMainAccount,
					Progress:      acc.AccountDetails.Progress,
				},
				AccountFlags: flags,
			})
			total += acc.AccountBalance.Amount
		}
		response.TotalBalance = total

		return nil
	})

	if err != nil {
		return entities.DashboardResponse{}, err
	}
	return response, nil
}

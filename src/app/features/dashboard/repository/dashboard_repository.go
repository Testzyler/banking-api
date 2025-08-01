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
	GetAccountsByUserID(userID string) ([]entities.Account, error)
	GetTransactionsByUserID(userID string) ([]entities.Transaction, error)
	GetBannersByUserID(userID string) ([]entities.Banner, error)
	GetUserByID(userID string) (entities.User, error)
	GetCardsByUserID(userID string) ([]entities.DebitCards, error)
	GetTotalBalance(userID string) float64

	GetDashboardDataWithTrx(userID string) (entities.DashboardResponse, error)
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

func (r *dashboardRepository) GetAccountsByUserID(userID string) ([]entities.Account, error) {
	var result []entities.Account
	var accounts []models.Account
	if err := r.db.
		Preload("AccountDetails").
		Preload("AccountBalance").
		Preload("AccountFlags").
		Find(&accounts, "user_id = ?", userID).Error; err != nil {
		return nil, err
	}

	for _, account := range accounts {
		result = append(result, convertAccountToEntity(account, account.AccountDetails, account.AccountBalance, account.AccountFlags))
	}

	return result, nil
}

func convertAccountToEntity(
	account models.Account,
	accountDetail models.AccountDetail,
	accountBalance models.AccountBalance,
	accountFlags []models.AccountFlag,
) entities.Account {
	var flags []entities.AccountFlags
	for _, flag := range accountFlags {
		flags = append(flags, entities.AccountFlags{
			FlagType:  flag.FlagType,
			FlagValue: flag.FlagValue,
			CreatedAt: flag.CreatedAt,
			UpdatedAt: flag.UpdatedAt,
		})
	}

	return entities.Account{
		AccountID: account.AccountID,
		Type:      account.Type,
		Currency:  account.Currency,
		Issuer:    account.Issuer,
		Amount:    accountBalance.Amount,
		AccountDetails: entities.AccountDetails{
			Color:         accountDetail.Color,
			IsMainAccount: accountDetail.IsMainAccount,
			Progress:      accountDetail.Progress,
		},
		AccountFlags: flags,
	}
}

func (r *dashboardRepository) GetUserByID(userID string) (entities.User, error) {
	var user models.User

	if err := r.db.Preload("UserGreeting").
		First(&user, "user_id = ?", userID).Error; err != nil {
		return entities.User{}, err
	}

	result := entities.User{
		UserID:   user.UserID,
		Name:     user.Name,
		Greeting: user.UserGreeting.Greeting,
	}

	return result, nil
}

func (r *dashboardRepository) GetCardsByUserID(userID string) ([]entities.DebitCards, error) {
	var cards []models.DebitCard
	if err := r.db.Preload("DebitCardDetail").
		Preload("DebitCardDesign").
		Preload("DebitCardStatus").
		Where("user_id = ?", userID).
		Find(&cards).Error; err != nil {
		return nil, err
	}

	var result []entities.DebitCards
	for _, c := range cards {
		result = append(result, entities.DebitCards{
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
	return result, nil
}

func (r *dashboardRepository) GetBannersByUserID(userID string) ([]entities.Banner, error) {
	var banners []models.Banner
	if err := r.db.Find(&banners, "user_id = ?", userID).Error; err != nil {
		return nil, err
	}

	var result []entities.Banner
	for _, banner := range banners {
		result = append(result, entities.Banner{
			BannerID:    banner.BannerID,
			UserID:      banner.UserID,
			Title:       banner.Title,
			Description: banner.Description,
			ImageURL:    banner.Image,
		})
	}
	return result, nil
}

func (r *dashboardRepository) GetTransactionsByUserID(userID string) ([]entities.Transaction, error) {
	var transactions []models.Transaction
	if err := r.db.Find(&transactions, "user_id = ?", userID).Error; err != nil {
		return nil, err
	}

	var result []entities.Transaction
	for _, transaction := range transactions {
		result = append(result, entities.Transaction{
			TransactionID: transaction.TransactionID,
			UserID:        transaction.UserID,
			Name:          transaction.Name,
			Image:         transaction.Image,
			IsBank:        transaction.IsBank,
		})
	}
	return result, nil
}

func (r *dashboardRepository) GetDashboardDataWithTrx(userID string) (entities.DashboardResponse, error) {
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

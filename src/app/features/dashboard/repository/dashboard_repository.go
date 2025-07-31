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
}

func NewDashboardRepository(repo *gorm.DB) DashboardRepository {
	return &dashboardRepository{
		db: repo,
	}
}

func (r *dashboardRepository) GetAccountsByUserID(userID string) ([]entities.Account, error) {
	var accounts []models.Account
	if err := r.db.Find(&accounts, "user_id = ?", userID).Error; err != nil {
		return nil, err
	}

	var result []entities.Account
	for _, account := range accounts {
		var accountDetail models.AccountDetail
		if err := r.db.Find(&accountDetail, "account_id = ?", account.AccountID).Error; err != nil {
			return nil, err
		}

		var accountBalance models.AccountBalance
		if err := r.db.Find(&accountBalance, "account_id = ?", account.AccountID).Error; err != nil {
			return nil, err
		}

		var accountFlags models.AccountFlag
		if err := r.db.Find(&accountFlags, "account_id = ?", account.AccountID).Error; err != nil {
			return nil, err
		}

		result = append(result, convertAccountToEntity(account, accountDetail, accountBalance, accountFlags))
	}
	return result, nil
}

func convertAccountToEntity(account models.Account, accountDetail models.AccountDetail, accountBalance models.AccountBalance, accountFlags models.AccountFlag) entities.Account {
	return entities.Account{
		AccountID: account.AccountID,
		Type:      account.Type,
		Currency:  account.Currency,
		Issuer:    account.Issuer,
		AccountDetails: entities.AccountDetails{
			Color:         accountDetail.Color,
			IsMainAccount: accountDetail.IsMainAccount,
			Progress:      accountDetail.Progress,
		},
		AccountBalance: entities.AccountBalance{
			Amount: accountBalance.Amount,
		},
		AccountFlags: entities.AccountFlags{
			AccountID: accountFlags.AccountID,
			FlagType:  accountFlags.FlagType,
			FlagValue: accountFlags.FlagValue,
			CreatedAt: accountFlags.CreatedAt,
			UpdatedAt: accountFlags.UpdatedAt,
		},
	}
}

func (r *dashboardRepository) GetUserByID(userID string) (entities.User, error) {
	result := entities.User{}
	var user models.User
	if err := r.db.Find(&user, "user_id = ?", userID).Error; err != nil {
		return entities.User{}, err
	}

	var greeting models.UserGreeting
	if err := r.db.Find(&greeting, "user_id = ?", userID).Error; err != nil {
		return entities.User{}, err
	}

	result.UserID = user.UserID
	result.Name = user.Name
	result.Greeting = greeting.Greeting

	return result, nil
}

func (r *dashboardRepository) GetCardsByUserID(userID string) ([]entities.DebitCards, error) {
	var result []entities.DebitCards

	var cards []models.DebitCard
	if err := r.db.Find(&cards, "user_id = ?", userID).Error; err != nil {
		return nil, err
	}

	for _, card := range cards {
		var cardDetail models.DebitCardDetail
		if err := r.db.Find(&cardDetail, "card_id = ?", card.CardID).Error; err != nil {
			return nil, err
		}

		var cardDesign models.DebitCardDesign
		if err := r.db.Find(&cardDesign, "card_id = ?", card.CardID).Error; err != nil {
			return nil, err
		}

		var cardStatus models.DebitCardStatus
		if err := r.db.Find(&cardStatus, "card_id = ?", card.CardID).Error; err != nil {
			return nil, err
		}

		result = append(result, convertCardToEntity(card, cardDetail, cardStatus, cardDesign))
	}
	return result, nil
}

func convertCardToEntity(card models.DebitCard, cardDetail models.DebitCardDetail, cardStatus models.DebitCardStatus, cardDesign models.DebitCardDesign) entities.DebitCards {
	return entities.DebitCards{
		CardID:     card.CardID,
		CardName:   card.Name,
		CardNumber: cardDetail.Number,
		Issuer:     cardDetail.Issuer,
		Status:     cardStatus.Status,
		DebitCardDesign: entities.DebitCardDesign{
			CardID:      cardDesign.CardID,
			Color:       cardDesign.Color,
			BorderColor: cardDesign.BorderColor,
		},
	}
}

func (r *dashboardRepository) GetBannersByUserID(userID string) ([]entities.Banner, error) {
	var banners []models.Banner
	if err := r.db.Find(&banners, "user_id = ?", userID).Error; err != nil {
		return nil, err
	}

	var result []entities.Banner
	for _, banner := range banners {
		result = append(result, convertBannerToEntity(banner))
	}
	return result, nil
}

func convertBannerToEntity(banner models.Banner) entities.Banner {
	return entities.Banner{
		BannerID:    banner.BannerID,
		UserID:      banner.UserID,
		Title:       banner.Title,
		Description: banner.Description,
		ImageURL:    banner.Image,
	}
}

func (r *dashboardRepository) GetTransactionsByUserID(userID string) ([]entities.Transaction, error) {
	var transactions []models.Transaction
	if err := r.db.Find(&transactions, "user_id = ?", userID).Error; err != nil {
		return nil, err
	}

	var result []entities.Transaction
	for _, transaction := range transactions {
		result = append(result, convertTransactionToEntity(transaction))
	}
	return result, nil
}

func convertTransactionToEntity(transaction models.Transaction) entities.Transaction {
	return entities.Transaction{
		TransactionID: transaction.TransactionID,
		UserID:        transaction.UserID,
		Name:          transaction.Name,
		Image:         transaction.Image,
		IsBank:        transaction.IsBank,
	}
}

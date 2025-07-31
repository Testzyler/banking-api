package migrations

import (
	"fmt"

	"github.com/Testzyler/banking-api/app/models"
	"github.com/Testzyler/banking-api/logger"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var createUserPin = &Migration{
	Number: 3,
	Name:   "create user pin",

	Forwards: func(db *gorm.DB) error {
		return Migrate_CreateUserPin(db)
	},
}

func Migrate_CreateUserPin(db *gorm.DB) error {
	if err := db.Migrator().AutoMigrate(&models.UserPin{}); err != nil {
		return err
	}

	const defaultPIN = "123456"
	hashedDefaultPIN, err := bcrypt.GenerateFromPassword([]byte(defaultPIN), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	var existingUsers []models.User
	if err := db.Table("users").Select("user_id").Find(&existingUsers).Error; err != nil {
		logger.Error("Error query existing users: %v\n", err)
		return fmt.Errorf("failed to query existing users: %w", err)
	}
	logger.Info("Found %d existing users to migrate.\n", len(existingUsers))

	if len(existingUsers) == 0 {
		logger.Error("No existing users found. Skipping PIN population.")
		return nil
	}
	var userPinsToCreate []models.UserPin
	for _, user := range existingUsers {
		userPinsToCreate = append(userPinsToCreate, models.UserPin{
			UserID:            user.UserID,
			HashedPin:         string(hashedDefaultPIN),
			FailedPinAttempts: 0,
		})
	}

	if err := db.CreateInBatches(userPinsToCreate, 1000).Error; err != nil {
		logger.Error("Error UserPin records in batches: %v\n", err)
		return fmt.Errorf("failed to insert user PINs: %w", err)
	}
	logger.Info("Created %d user pins.\n", len(userPinsToCreate))

	return nil
}

func init() {
	Migrations = append(Migrations, createUserPin)
}

package database

import (
	"github.com/Testzyler/banking-api/config"
	"gorm.io/gorm"
)

type DatabaseInterface interface {
	GetDB() *gorm.DB
	RunMigrations() error
	Close() error
}

type Database struct {
	DB *gorm.DB
}

func NewDatabase(config *config.Config) (DatabaseInterface, error) {
	return NewMySQLDatabase(config)
}

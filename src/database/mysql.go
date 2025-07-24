package database

import (
	"fmt"
	"log"
	"time"

	"github.com/Testzyler/banking-api/config"
	"github.com/Testzyler/banking-api/database/migrations"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewMySQLDatabase(config *config.Config) (DatabaseInterface, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.Database.Username,
		config.Database.Password,
		config.Database.Host,
		config.Database.Port,
		config.Database.Name,
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying sql.DB to configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxOpenConns(int(config.Database.MaxOpenConns))
	sqlDB.SetMaxIdleConns(int(config.Database.MaxOpenConns) / 2)
	sqlDB.SetConnMaxIdleTime(time.Duration(config.Database.MaxIdleTimeInSecond) * time.Second)

	database := &Database{DB: db}

	return database, nil
}

func (d *Database) GetDB() *gorm.DB {
	return d.DB
}

func (d *Database) RunMigrations() error {
	// First, create migrations table if it doesn't exist
	if err := d.DB.AutoMigrate(&migrations.Migration{}); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Run each migration that hasn't been run yet
	for _, migration := range migrations.Migrations {
		var existingMigration migrations.Migration
		result := d.DB.Where("number = ?", migration.Number).First(&existingMigration)

		if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
			return fmt.Errorf("failed to check migration %d: %w", migration.Number, result.Error)
		}

		// If migration doesn't exist, run it
		if result.Error == gorm.ErrRecordNotFound {
			log.Printf("Running migration %d: %s", migration.Number, migration.Name)

			if err := migration.Forwards(d.DB); err != nil {
				return fmt.Errorf("migration %d failed: %w", migration.Number, err)
			}

			// Record that migration was run
			migrationRecord := &migrations.Migration{
				Number: migration.Number,
				Name:   migration.Name,
			}
			if err := d.DB.Create(migrationRecord).Error; err != nil {
				return fmt.Errorf("failed to record migration %d: %w", migration.Number, err)
			}
		}
	}

	return nil
}

func (d *Database) Close() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

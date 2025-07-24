package migrations

import (
	"fmt"
	"os"
	"strings"

	"gorm.io/gorm"
)

var path = ""

var initialSchemaSetup = &Migration{
	Number: 1,
	Name:   "add seed mock data",

	Forwards: func(db *gorm.DB) error {
		tx := db.Begin()
		if tx.Error != nil {
			return fmt.Errorf("could not start transaction: %w", tx.Error)
		}

		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
			}
		}()

		mockDir := "database/seeds"
		files, err := os.ReadDir(mockDir)
		if err != nil {
			return fmt.Errorf("could not read mock directory: %w", err)
		}

		for _, file := range files {
			if !file.IsDir() && strings.HasSuffix(file.Name(), ".sql") {
				fmt.Printf("Executing SQL file: %s\n", file.Name())
				filePath := fmt.Sprintf("%s/%s", mockDir, file.Name())
				err := executeSQLFile(tx, filePath)
				if err != nil {
					tx.Rollback()
					return fmt.Errorf("could not execute SQL file %s: %w", file.Name(), err)
				}
			}
		}

		return tx.Commit().Error
	},
}

func executeSQLFile(db *gorm.DB, filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("could not read SQL file %s: %w", filePath, err)
	}

	sql := string(content)
	if strings.TrimSpace(sql) == "" {
		return nil // skip empty files
	}

	// multiple statements
	statements := strings.Split(sql, ";")

	for i, statement := range statements {
		statement = strings.TrimSpace(statement)
		if statement == "" {
			continue
		}

		err = db.Exec(statement).Error
		if err != nil {
			return fmt.Errorf("could not execute statement %d from file %s: %w", i+1, filePath, err)
		}
	}

	return nil
}

func init() {
	Migrations = append(Migrations, initialSchemaSetup)
}

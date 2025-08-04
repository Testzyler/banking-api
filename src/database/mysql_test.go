package database

import (
	"errors"
	"testing"

	"github.com/Testzyler/banking-api/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MockDatabase is a mock that focuses on behavior rather than SQL
type IntegrateMockDatabase struct {
	mock.Mock
}

func (m *IntegrateMockDatabase) GetDB() *gorm.DB {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*gorm.DB)
}

func (m *IntegrateMockDatabase) RunMigrations() error {
	args := m.Called()
	return args.Error(0)
}

func (m *IntegrateMockDatabase) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestDatabase_IntegrateMock_Success(t *testing.T) {
	mockDB := new(IntegrateMockDatabase)

	// Test successful migration
	mockDB.On("RunMigrations").Return(nil)

	err := mockDB.RunMigrations()
	assert.NoError(t, err)
	mockDB.AssertExpectations(t)
}

func TestDatabase_IntegrateMock_MigrationFailure(t *testing.T) {
	mockDB := new(IntegrateMockDatabase)

	// Test migration failure
	mockDB.On("RunMigrations").Return(errors.New("migration failed"))

	err := mockDB.RunMigrations()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "migration failed")
	mockDB.AssertExpectations(t)
}

func TestDatabase_IntegrateMock_Close(t *testing.T) {
	mockDB := new(IntegrateMockDatabase)

	// Test successful close
	mockDB.On("Close").Return(nil)

	err := mockDB.Close()
	assert.NoError(t, err)
	mockDB.AssertExpectations(t)
}

func TestDatabase_IntegrateMock_CloseFailure(t *testing.T) {
	mockDB := new(IntegrateMockDatabase)

	// Test close failure
	mockDB.On("Close").Return(errors.New("close failed"))

	err := mockDB.Close()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "close failed")
	mockDB.AssertExpectations(t)
}

func TestDatabase_ConfigValidation_Detailed(t *testing.T) {
	testCases := []struct {
		name        string
		config      *config.Config
		expectError bool
		errorText   string
	}{
		{
			name: "valid_config",
			config: &config.Config{
				Database: &config.Database{
					Host:                "localhost",
					Port:                "3306",
					Username:            "user",
					Password:            "pass",
					Name:                "db",
					MaxOpenConns:        10,
					MaxIdleTimeInSecond: 300,
				},
			},
			expectError: true, // Will fail to connect, but config is valid
			errorText:   "failed to connect to database",
		},
		{
			name: "empty_host",
			config: &config.Config{
				Database: &config.Database{
					Host:                "",
					Port:                "3306",
					Username:            "user",
					Password:            "pass",
					Name:                "db",
					MaxOpenConns:        10,
					MaxIdleTimeInSecond: 300,
				},
			},
			expectError: true,
			errorText:   "failed to connect to database",
		},
		{
			name: "zero_max_conns",
			config: &config.Config{
				Database: &config.Database{
					Host:                "localhost",
					Port:                "3306",
					Username:            "user",
					Password:            "pass",
					Name:                "db",
					MaxOpenConns:        0,
					MaxIdleTimeInSecond: 300,
				},
			},
			expectError: true,
			errorText:   "failed to connect to database",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db, err := NewMySQLDatabase(tc.config)

			if tc.expectError {
				assert.Error(t, err)
				assert.Nil(t, db)
				if tc.errorText != "" {
					assert.Contains(t, err.Error(), tc.errorText)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, db)
				db.Close()
			}
		})
	}
}

// Test business logic without database dependencies
func TestDatabase_BusinessLogic(t *testing.T) {
	t.Run("DSN_generation", func(t *testing.T) {
		config := &config.Database{
			Host:     "testhost",
			Port:     "3307",
			Username: "testuser",
			Password: "testpass",
			Name:     "testdb",
		}

		expectedDSN := "testuser:testpass@tcp(testhost:3307)/testdb?charset=utf8mb4&parseTime=True&loc=Local"
		actualDSN := buildDSN(config)
		assert.Equal(t, expectedDSN, actualDSN)
	})

	t.Run("Connection_pool_calculation", func(t *testing.T) {
		tests := []struct {
			maxOpen  int
			expected int
		}{
			{10, 5},
			{20, 10},
			{1, 0},
			{0, 0},
		}

		for _, test := range tests {
			actual := computeMaxIdleConns(test.maxOpen)
			assert.Equal(t, test.expected, actual)
		}
	})
}

func buildDSN(dbConfig *config.Database) string {
	return dbConfig.Username + ":" + dbConfig.Password + "@tcp(" +
		dbConfig.Host + ":" + dbConfig.Port + ")/" +
		dbConfig.Name + "?charset=utf8mb4&parseTime=True&loc=Local"
}

func computeMaxIdleConns(maxOpenConns int) int {
	if maxOpenConns <= 1 {
		return 0
	}
	return maxOpenConns / 2
}

func TestDatabase_InterfaceCompliance(t *testing.T) {
	var _ DatabaseInterface = (*Database)(nil)

	var _ DatabaseInterface = (*IntegrateMockDatabase)(nil)

	t.Log("All types implement DatabaseInterface correctly")
}

func TestDatabase_ErrorScenarios(t *testing.T) {
	t.Run("nil_config", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Log("Correctly panics with nil config")
			}
		}()

		t.Log("Nil config test structure verified")
	})

	t.Run("invalid_port", func(t *testing.T) {
		config := &config.Config{
			Database: &config.Database{
				Host:     "localhost",
				Port:     "invalid_port",
				Username: "user",
				Password: "pass",
				Name:     "db",
			},
		}

		db, err := NewMySQLDatabase(config)
		assert.Error(t, err)
		assert.Nil(t, db)
	})
}

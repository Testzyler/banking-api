package database

import (
	"sync"

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

// Global variables สำหรับ Database และ Cache instances
var (
	dbOnce        sync.Once
	cacheOnce     sync.Once
	dbInstance    DatabaseInterface
	cacheInstance *RedisDatabase
)

// GetDatabase returns global database instance (singleton)
func GetDatabase() DatabaseInterface {
	if dbInstance == nil {
		dbOnce.Do(func() {
			var err error
			dbInstance, err = NewMySQLDatabase(config.GetConfig())
			if err != nil {
				panic("Failed to connect to database: " + err.Error())
			}
		})
	}
	return dbInstance
}

// GetCache returns global cache instance (singleton)
func GetCache() *RedisDatabase {
	if cacheInstance == nil {
		cacheOnce.Do(func() {
			var err error
			cacheInstance, err = NewCacheClient(config.GetConfig().Cache)
			if err != nil {
				panic("Failed to connect to cache: " + err.Error())
			}
		})
	}
	return cacheInstance
}

// InitDatabase initializes global database instance
func InitDatabase(cfg *config.Config) error {
	dbOnce.Do(func() {
		var err error
		dbInstance, err = NewMySQLDatabase(cfg)
		if err != nil {
			panic("Failed to connect to database: " + err.Error())
		}
	})
	return nil
}

// InitCache initializes global cache instance
func InitCache(cfg *config.CacheConfig) error {
	cacheOnce.Do(func() {
		var err error
		cacheInstance, err = NewCacheClient(cfg)
		if err != nil {
			panic("Failed to connect to cache: " + err.Error())
		}
	})
	return nil
}

func NewDatabase(config *config.Config) (DatabaseInterface, error) {
	return NewMySQLDatabase(config)
}

func NewCache(config *config.CacheConfig) (*RedisDatabase, error) {
	return NewCacheClient(config)
}

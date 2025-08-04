package config

import (
	"log"
	"strings"
	"sync"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server   *Server
	Database *Database
	Cache    *CacheConfig
	Logger   *Logger
	Auth     *AuthConfig
}

type Server struct {
	Port            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	Environment     string
	MaxConnections  int
	ShutdownTimeout int
}

type Database struct {
	Host                string
	Port                string
	Username            string
	Password            string
	Name                string
	MaxOpenConns        int
	MaxIdleTimeInSecond int
}

type CacheConfig struct {
	// Single instance configuration
	Host     string
	Port     int
	Password string
	DB       int

	// Connection pool settings
	MaxIdle     int
	IdleTimeout time.Duration

	// Sentinel configuration
	UseSentinel      bool
	SentinelAddrs    []string
	SentinelPassword string
	MasterName       string
}

type Logger struct {
	Level    string
	LogColor bool
	LogJson  bool
}

type AuthConfig struct {
	Jwt *JwtConfig
	Pin *PinConfig
}

type JwtConfig struct {
	AccessTokenSecret  string
	RefreshTokenSecret string
	AccessTokenExpiry  time.Duration
	RefreshTokenExpiry time.Duration
}

type PinConfig struct {
	BaseDuration    time.Duration
	LockThreshold   int
	MaxLockDuration time.Duration
}

var (
	once   sync.Once
	config *Config
)

func NewConfig(envFile string) *Config {
	once.Do(func() {
		config = loadConfig(envFile)
	})
	return config
}

var envFile string

func loadConfig(envFile string) *Config {
	if envFile != "" {
		viper.SetConfigFile(envFile)
	} else {
		viper.AddConfigPath(".")
		viper.AddConfigPath("..")
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("unable to read config: %v\n", err)
	}
	return &Config{
		Server: &Server{
			Port:            viper.GetString("Server.Port"),
			ReadTimeout:     viper.GetDuration("Server.ReadTimeout"),
			WriteTimeout:    viper.GetDuration("Server.WriteTimeout"),
			IdleTimeout:     viper.GetDuration("Server.IdleTimeout"),
			MaxConnections:  viper.GetInt("Server.MaxConnections"),
			ShutdownTimeout: viper.GetInt("Server.ShutdownTimeout"),
			Environment:     viper.GetString("Server.Environment"),
		},
		Database: &Database{
			Host:                viper.GetString("Database.Host"),
			Port:                viper.GetString("Database.Port"),
			Username:            viper.GetString("Database.Username"),
			Password:            viper.GetString("Database.Password"),
			Name:                viper.GetString("Database.Name"),
			MaxOpenConns:        viper.GetInt("Database.MaxOpenConns"),
			MaxIdleTimeInSecond: viper.GetInt("Database.MaxIdleTimeInSecond"),
		},
		Cache: &CacheConfig{
			Host:             viper.GetString("Redis.Host"),
			Port:             viper.GetInt("Redis.Port"),
			Password:         viper.GetString("Redis.Password"),
			DB:               viper.GetInt("Redis.DB"),
			MaxIdle:          viper.GetInt("Redis.MaxIdle"),
			IdleTimeout:      viper.GetDuration("Redis.IdleTimeout"),
			UseSentinel:      viper.GetBool("Redis.UseSentinel"),
			SentinelAddrs:    viper.GetStringSlice("Redis.SentinelAddrs"),
			SentinelPassword: viper.GetString("Redis.SentinelPassword"),
			MasterName:       viper.GetString("Redis.MasterName"),
		},
		Logger: &Logger{
			Level:    viper.GetString("Logger.Level"),
			LogColor: viper.GetBool("Logger.LogColor"),
			LogJson:  viper.GetBool("Logger.LogJson"),
		},
		Auth: &AuthConfig{
			Jwt: &JwtConfig{
				AccessTokenSecret:  viper.GetString("Auth.Jwt.AccessTokenSecret"),
				RefreshTokenSecret: viper.GetString("Auth.Jwt.RefreshTokenSecret"),
				AccessTokenExpiry:  time.Duration(viper.GetInt("Auth.Jwt.AccessTokenExpiry")) * time.Minute,
				RefreshTokenExpiry: time.Duration(viper.GetInt("Auth.Jwt.RefreshTokenExpiry")) * 24 * time.Hour,
			},
			Pin: &PinConfig{
				BaseDuration:    viper.GetDuration("Auth.Pin.BaseDuration"),
				LockThreshold:   viper.GetInt("Auth.Pin.LockThreshold"),
				MaxLockDuration: viper.GetDuration("Auth.Pin.MaxLockDuration"),
			},
		},
	}
}

func GetConfig() *Config {
	return config
}

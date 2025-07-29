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
	Logger   *Logger
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

type Logger struct {
	Level    string
	LogColor bool
	LogJson  bool
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
		Logger: &Logger{
			Level:    viper.GetString("Logger.Level"),
			LogColor: viper.GetBool("Logger.LogColor"),
			LogJson:  viper.GetBool("Logger.LogJson"),
		},
	}
}

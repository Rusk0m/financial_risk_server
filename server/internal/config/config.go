package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Config представляет конфигурацию приложения
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Uploads  UploadsConfig
	MarketData MarketDataConfig
}

// ServerConfig конфигурация HTTP сервера
type ServerConfig struct {
	Port         string
	Host         string
	ReadTimeout  int
	WriteTimeout int
}

// DatabaseConfig конфигурация базы данных
type DatabaseConfig struct {
	Host         string
	Port         int
	User         string
	Password     string
	DBName       string
	SSLMode      string
	MaxOpenConns int
	MaxIdleConns int
}

// UploadsConfig конфигурация загрузки файлов
type UploadsConfig struct {
	Directory string
	MaxSize   int64 // в байтах
}
type MarketDataConfig struct {
	Enabled             bool   `env:"MARKET_DATA_SYNC_ENABLED" default:"true"`
	IntervalHours       int    `env:"MARKET_DATA_SYNC_INTERVAL_HOURS" default:"24"`
	InitialDelaySeconds int    `env:"MARKET_DATA_SYNC_INITIAL_DELAY" default:"60"`
	HistoryDays         int    `env:"MARKET_DATA_HISTORY_DAYS" default:"30"`
}

// Load загружает конфигурацию из .env файла и переменных окружения
func Load() (*Config, error) {
	// Загружаем .env файл если существует
	_ = godotenv.Load()

	config := &Config{
		Server: ServerConfig{
			Port:         getEnv("SERVER_PORT", "8080"),
			Host:         getEnv("SERVER_HOST", "0.0.0.0"),
			ReadTimeout:  getIntEnv("SERVER_READ_TIMEOUT", 30),
			WriteTimeout: getIntEnv("SERVER_WRITE_TIMEOUT", 30),
		},
		Database: DatabaseConfig{
			Host:         getEnv("DB_HOST", "localhost"),
			Port:         getIntEnv("DB_PORT", 5432),
			User:         getEnv("DB_USER", "risk_user"),
			Password:     getEnv("DB_PASSWORD", "risk_password"),
			DBName:       getEnv("DB_NAME", "financial_risk_db2"),
			SSLMode:      getEnv("DB_SSL_MODE", "disable"),
			MaxOpenConns: getIntEnv("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns: getIntEnv("DB_MAX_IDLE_CONNS", 5),
		},
		Uploads: UploadsConfig{
			Directory: getEnv("UPLOADS_DIRECTORY", "./uploads"),
			MaxSize:   getInt64Env("UPLOADS_MAX_SIZE", 10*1024*1024), // 10 МБ по умолчанию
		},
		MarketData: MarketDataConfig{
			Enabled:             getBoolEnv("MARKET_DATA_SYNC_ENABLED", true),
			IntervalHours:       getIntEnv("MARKET_DATA_SYNC_INTERVAL_HOURS", 24),
			InitialDelaySeconds: getIntEnv("MARKET_DATA_SYNC_INITIAL_DELAY", 60),
			HistoryDays:         getIntEnv("MARKET_DATA_HISTORY_DAYS", 30),
		},
	}

	// Валидация критически важных параметров
	if config.Database.User == "" || config.Database.Password == "" || config.Database.DBName == "" {
		return nil, fmt.Errorf("database configuration is incomplete")
	}

	return config, nil
}

// getEnv возвращает значение переменной окружения или значение по умолчанию
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getIntEnv возвращает целочисленное значение переменной окружения или значение по умолчанию
func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var result int
		_, err := fmt.Sscanf(value, "%d", &result)
		if err == nil {
			return result
		}
	}
	return defaultValue
}

// getInt64Env возвращает 64-битное целочисленное значение переменной окружения или значение по умолчанию
func getInt64Env(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		var result int64
		_, err := fmt.Sscanf(value, "%d", &result)
		if err == nil {
			return result
		}
	}
	return defaultValue
}
// ✅ ДОБАВЬТЕ ЭТУ ФУНКЦИЮ:
func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		switch value {
		case "true", "1", "yes", "on":
			return true
		case "false", "0", "no", "off":
			return false
		}
	}
	return defaultValue
}

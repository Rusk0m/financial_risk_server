package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

// Load загружает конфигурацию из YAML файла и переменных окружения
func Load(configPath string) (*Config, error) {
	// Загружаем .env файл, если он существует
	if err := godotenv.Load(); err != nil {
		// Не критично, если .env не найден
		fmt.Println("ℹ️  Файл .env не найден, используем значения по умолчанию")
	}

	// Читаем YAML файл конфигурации
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения файла конфигурации %s: %w", configPath, err)
	}

	// Парсим YAML в структуру
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("ошибка парсинга YAML: %w", err)
	}

	// Переопределяем значения из переменных окружения
	cfg.Server.Port = getEnv("SERVER_PORT", cfg.Server.Port)
	cfg.Server.Environment = getEnv("SERVER_ENV", cfg.Server.Environment)

	cfg.Database.Host = getEnv("DB_HOST", cfg.Database.Host)
	cfg.Database.Port = int(getEnvInt("DB_PORT", int64(cfg.Database.Port)))
	cfg.Database.User = getEnv("DB_USER", cfg.Database.User)
	cfg.Database.Password = getEnv("DB_PASSWORD", cfg.Database.Password)
	cfg.Database.Name = getEnv("DB_NAME", cfg.Database.Name)
	cfg.Database.SSLMode = getEnv("DB_SSLMODE", cfg.Database.SSLMode)

	cfg.Logger.Level = getEnv("LOG_LEVEL", cfg.Logger.Level)
	cfg.Logger.Format = getEnv("LOG_FORMAT", cfg.Logger.Format)
	cfg.Logger.Output = getEnv("LOG_OUTPUT", cfg.Logger.Output)

	return &cfg, nil
}

// getEnv возвращает значение переменной окружения или значение по умолчанию
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt возвращает значение переменной окружения как int64 или значение по умолчанию
func getEnvInt(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		var result int64
		_, err := fmt.Sscan(value, &result)
		if err == nil {
			return result
		}
	}
	return defaultValue
}

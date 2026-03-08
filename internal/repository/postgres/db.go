package postgres

import (
	"database/sql"
	"financial-risk-server/internal/config"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

// DB представляет обёртку над *sql.DB
type DB struct {
	*sql.DB
}

// Connect подключается к базе данных PostgreSQL
func Connect(cfg *config.DatabaseConfig) (*DB, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.Name,
		cfg.SSLMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания подключения: %w", err)
	}

	// Проверка подключения
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ошибка проверки подключения: %w", err)
	}

	// Настройка пула соединений
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)

	fmt.Printf("✅ Подключение к PostgreSQL установлено: %s@%s:%d/%s\n",
		cfg.User, cfg.Host, cfg.Port, cfg.Name)

	return &DB{db}, nil
}

// Close закрывает подключение к базе данных
func (db *DB) Close() error {
	return db.DB.Close()
}

#!/bin/bash
set -e  # Останавливает выполнение при ошибке

# Параметры подключения
DB_USER="risk_user"
DB_PASSWORD="2005"
DB_NAME="financial_risk_db"
DB_HOST="localhost"
DB_PORT="5432"

# Формирует строку подключения: postgres://user:pass@host:port/db?sslmode=disable
DB_URL="postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable"

# Проверка/установка goose
if ! command -v goose &> /dev/null; then
    echo "❌ goose не установлен. Установка..."
    go install github.com/pressly/goose/v3/cmd/goose@latest
fi

# Запуск миграций из папки migrations
goose -dir migrations postgres "$DB_URL" up
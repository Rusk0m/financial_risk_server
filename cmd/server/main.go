package main

import (
	"finantial-risk-server/internal/config"
	"fmt"
	"log"
	"os"
)

func main() {
	// 1. Загружаем конфигурацию
	cfg, err := config.Load("../../internal/config/config.yaml")
	if err != nil {
		log.Fatalf("❌ Ошибка загрузки конфигурации: %v", err)
	}

	// 2. Выводим информацию о конфигурации
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║  FINANCIAL RISK ANALYSIS SYSTEM                           ║")
	fmt.Println("╠════════════════════════════════════════════════════════════╣")
	fmt.Printf("║  Версия: 1.0.0                                             ║\n")
	fmt.Printf("║  Окружение: %s", cfg.Server.Environment)
	fmt.Printf("%s║\n", spaces(51-len(cfg.Server.Environment)))
	fmt.Printf("║  Порт: %s", cfg.Server.Port)
	fmt.Printf("%s║\n", spaces(55-len(cfg.Server.Port)))
	fmt.Printf("║  БД: %s@%s:%d/%s", cfg.Database.User, cfg.Database.Host, cfg.Database.Port, cfg.Database.Name)
	fmt.Printf("%s║\n", spaces(56-len(cfg.Database.User)-len(cfg.Database.Host)-len(cfg.Database.Name)-3))
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// 3. Проверяем подключение к базе данных (пока заглушка)
	fmt.Println("⏳ Проверка подключения к базе данных...")

	// Здесь позже будет:
	// db, err := postgres.Connect(&cfg.Database)
	// if err != nil { ... }

	fmt.Println("✅ Подключение к базе данных успешно")
	fmt.Println()

	// 4. Инициализация репозиториев, сервисов, обработчиков
	// (будет добавлено позже)

	// 5. Запуск сервера
	// (будет добавлено позже)

	// Пока просто выводим сообщение
	fmt.Println("🚀 Сервер готов к запуску!")
	fmt.Println("   Нажмите Ctrl+C для остановки")
	fmt.Println()

	// Ждём сигнала остановки
	sigChan := make(chan os.Signal, 1)
	<-sigChan
}

// spaces возвращает строку из n пробелов
func spaces(n int) string {
	if n < 0 {
		return ""
	}
	return fmt.Sprintf("%*s", n, "")
}

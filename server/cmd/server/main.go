// main.go
// Точка входа в систему анализа финансовых рисков ОАО «Беларуськалий»

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"financial-risk-server/internal/config"
	"financial-risk-server/internal/delivery/server"
	"financial-risk-server/internal/delivery/server/middleware"
	"financial-risk-server/internal/repository/postgres"
	"financial-risk-server/internal/service"
)

func main() {
	// 1. Загружаем конфигурацию из переменных окружения / .env
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌ Ошибка загрузки конфигурации: %v", err)
	}

	// 2. Выводим информацию о запуске
	printStartupInfo(cfg)

	// 3. Подключаемся к базе данных
	log.Println("⏳ Подключение к базе данных...")
	
	db, err := postgres.Connect(&postgres.Config{
		Host:            cfg.Database.Host,
		Port:            cfg.Database.Port,
		User:            cfg.Database.User,
		Password:        cfg.Database.Password,
		DBName:          cfg.Database.DBName,
		SSLMode:         cfg.Database.SSLMode,
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		MaxIdleConns:    cfg.Database.MaxIdleConns,
		ConnMaxLifetime: 0,
	})
	if err != nil {
		log.Fatalf("❌ Ошибка подключения к БД: %v", err)
	}
	defer db.Close()
	
	log.Println("✅ Подключение к базе данных установлено")

	// 4. Создаём репозитории
	enterpriseRepo := postgres.NewEnterpriseRepository(db)
	reportRepo := postgres.NewReportRepository(db)
	contractRepo := postgres.NewExportContractRepository(db)
	balanceRepo := postgres.NewBalanceSheetRepository(db)
	finResultRepo := postgres.NewFinancialResultRepository(db)
	creditRepo := postgres.NewCreditAgreementRepository(db)
	marketDataRepo := postgres.NewMarketDataRepository(db)
	riskRepo := postgres.NewRiskCalculationRepository(db)
	comprehensiveReportRepo := postgres.NewComprehensiveRiskReportRepository(db)

	// 5. Создаём сервисы (соблюдаем порядок зависимостей)
	
	// EnterpriseService не имеет зависимостей от других сервисов
	enterpriseService := service.NewEnterpriseService(enterpriseRepo)
	
	// ReportService зависит от enterpriseService
	reportService := service.NewReportService(
		reportRepo,
		contractRepo,
		balanceRepo,
		finResultRepo,
		creditRepo,
		enterpriseService,
		&service.ReportServiceConfig{
			UploadDirectory: cfg.Uploads.Directory,
			//MaxFileSize:     cfg.Uploads.MaxSize,
		},
	)
	
	// RiskCalculationService
	riskService := service.NewRiskCalculationService(
		riskRepo,
		contractRepo,
		balanceRepo,
		creditRepo,
		marketDataRepo,
		enterpriseService,
	)
	
	// ComprehensiveReportService
	comprehensiveReportService := service.NewComprehensiveReportService(
		comprehensiveReportRepo,
		riskRepo,
		enterpriseService,
	)

	// 6. Настраиваем маршрутизатор (передаём сервисы, не хендлеры)
	router := server.SetupRouter(
		enterpriseService,
		reportService,
		riskService,
		comprehensiveReportService,
	)

	// 7. Создаём HTTP-сервер с настройками таймаутов
	serverAddr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	
	loggedRouter := middleware.LoggerMiddleware(router)

	
	httpServer := &http.Server{
		Addr:         serverAddr,
		Handler:      loggedRouter,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 8. Запускаем сервер в отдельной горутине
	go func() {
		log.Printf("🚀 Сервер запущен на %s", serverAddr)
		log.Printf("📊 Environment: %s", getEnv("SERVER_ENV", "development"))
		
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Ошибка запуска сервера: %v", err)
		}
	}()

	// 9. Ожидаем сигнал завершения для graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	
	log.Println("🛑 Получен сигнал завершения, останавливаем сервер...")

	// Даём 10 секунд на завершение активных запросов
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("⚠️  Ошибка при остановке сервера: %v", err)
	}
	
	log.Println("✅ Сервер остановлен корректно")
}

// printStartupInfo выводит информацию о конфигурации при запуске
func printStartupInfo(cfg *config.Config) {
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║  FINANCIAL RISK ANALYSIS SYSTEM — ОАО «Беларуськалий»    ║")
	fmt.Println("╠════════════════════════════════════════════════════════════╣")
	fmt.Printf("║  Порт сервера:    %-39s║\n", cfg.Server.Host+":"+cfg.Server.Port)
	fmt.Printf("║  БД:              %-39s║\n", 
		fmt.Sprintf("%s@%s:%d/%s", cfg.Database.User, cfg.Database.Host, cfg.Database.Port, cfg.Database.DBName))
	fmt.Printf("║  Загрузка файлов: %-39s║\n", cfg.Uploads.Directory)
	fmt.Printf("║  Max файл:        %-39s║\n", formatBytes(cfg.Uploads.MaxSize))
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println()
}

// formatBytes форматирует размер в байтах в человекочитаемый вид
func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

// getEnv возвращает значение переменной окружения или значение по умолчанию
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
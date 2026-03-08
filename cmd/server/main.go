package main

import (
	"financial-risk-server/internal/config"
	"financial-risk-server/internal/delivery/server"
	"financial-risk-server/internal/delivery/server/handlers"
	"financial-risk-server/internal/repository/postgres"
	"financial-risk-server/internal/service"
	"fmt"
	"log"
)

func main() {
	// 1. Загружаем конфигурацию
	cfg, err := config.Load("internal/config/config.yaml")
	if err != nil {
		log.Fatalf("❌ Ошибка загрузки конфигурации: %v", err)
	}

	// 2. Выводим баннер
	printBanner(cfg)

	// 3. Подключаемся к базе данных
	fmt.Println("⏳ Подключение к базе данных...")
	db, err := postgres.Connect(&cfg.Database)
	if err != nil {
		log.Fatalf("❌ Ошибка подключения к БД: %v", err)
	}
	defer db.Close()
	fmt.Println("✅ Подключение к базе данных установлено")
	fmt.Println()

	// 4. Создаём репозитории
	enterpriseRepo := postgres.NewEnterpriseRepository(db)
	exportRepo := postgres.NewExportContractRepository(db)
	marketRepo := postgres.NewMarketDataRepository(db)
	balanceRepo := postgres.NewBalanceSheetRepository(db)
	costRepo := postgres.NewCostStructureRepository(db)
	riskCalcRepo := postgres.NewRiskCalculationRepository(db)
	compReportRepo := postgres.NewComprehensiveReportRepository(db)

	// 5. Создаём фабрику сервисов и сервис расчёта рисков
	factory := service.NewServiceFactory(nil)
	riskService := factory.CreateRiskCalculationService(
		exportRepo,
		marketRepo,
		balanceRepo,
		costRepo,
		enterpriseRepo,
		riskCalcRepo,
		compReportRepo,
	)

	// 6. Создаём обработчики
	enterpriseHandler := handlers.NewEnterpriseHandler(enterpriseRepo)
	riskHandler := handlers.NewRiskHandler(riskService)
	balanceHandler := handlers.NewBalanceSheetHandler(balanceRepo)
	exportContractHandler := handlers.NewExportContractHandler(exportRepo)
	costStructureHandler := handlers.NewCostStructureHandler(costRepo)
	marketHandler := handlers.NewMarketDataHandler(marketRepo)
	healthHandler := handlers.NewHealthHandler()

	// 7. Настраиваем маршрутизатор
	router := server.SetupRouter(enterpriseHandler, exportContractHandler, costStructureHandler, balanceHandler, marketHandler, riskHandler, healthHandler)

	// 8. Создаём и настраиваем сервер
	serverConfig := &server.ServerConfig{
		Port:         cfg.Server.Port,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}
	httpServer := server.NewServer(serverConfig, router)

	// 9. Запускаем сервер
	// Используем блокирующий вызов Run()
	if err := httpServer.Run(); err != nil {
		log.Fatalf("❌ Ошибка запуска сервера: %v", err)
	}
}

// printBanner выводит красивый баннер при запуске
func printBanner(cfg *config.Config) {
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║  FINANCIAL RISK ANALYSIS SYSTEM                           ║")
	fmt.Println("╠════════════════════════════════════════════════════════════╣")
	fmt.Printf("║  Версия: 1.0.0                                             ║\n")
	fmt.Printf("║  Окружение: %-47s║\n", cfg.Server.Environment)
	fmt.Printf("║  Порт: %-51s║\n", cfg.Server.Port)
	fmt.Printf("║  БД: %s@%s:%d/%s%*s║\n",
		cfg.Database.User, cfg.Database.Host, cfg.Database.Port, cfg.Database.Name,
		56-len(cfg.Database.User)-len(cfg.Database.Host)-len(fmt.Sprint(cfg.Database.Port))-len(cfg.Database.Name)-3, "")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println()
}

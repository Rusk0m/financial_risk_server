package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"financial-risk-server/internal/config"
	"financial-risk-server/internal/delivery/server"
	"financial-risk-server/internal/repository/postgres"
	"financial-risk-server/internal/service"
)

func main() {
	// Загружаем конфигурацию
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Подключаемся к базе данных
	db, err := postgres.Connect(&postgres.Config{
		Host:             cfg.Database.Host,
		Port:             cfg.Database.Port,
		User:             cfg.Database.User,
		Password:         cfg.Database.Password,
		DBName:           cfg.Database.DBName,
		SSLMode:          cfg.Database.SSLMode,
		MaxOpenConns:     cfg.Database.MaxOpenConns,
		MaxIdleConns:     cfg.Database.MaxIdleConns,
		ConnMaxLifetime:  0,
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close() // Стандартный метод *sql.DB

	log.Println("✅ Database connection established")

	// Создаём репозитории
	enterpriseRepo := postgres.NewEnterpriseRepository(db)
	reportRepo := postgres.NewReportRepository(db)
	contractRepo := postgres.NewExportContractRepository(db)
	balanceRepo := postgres.NewBalanceSheetRepository(db)
	finResultRepo := postgres.NewFinancialResultRepository(db)
	creditRepo := postgres.NewCreditAgreementRepository(db)
	marketDataRepo := postgres.NewMarketDataRepository(db)
	riskRepo := postgres.NewRiskCalculationRepository(db)
	comprehensiveReportRepo := postgres.NewComprehensiveRiskReportRepository(db)

	// Создаём сервисы
	enterpriseService := service.NewEnterpriseService(enterpriseRepo)
	reportService := service.NewReportService(
		reportRepo,
		contractRepo,
		balanceRepo,
		finResultRepo,
		creditRepo,
		enterpriseService,
		&service.ReportServiceConfig{
			UploadDirectory: cfg.Uploads.Directory,
		},
	)
	riskService := service.NewRiskCalculationService(
		riskRepo,
		contractRepo,
		balanceRepo,
		creditRepo,
		marketDataRepo,
		enterpriseService,
	)
	comprehensiveReportService := service.NewComprehensiveReportService(
		comprehensiveReportRepo,
		riskRepo,
		enterpriseService,
	)

	// Настраиваем маршрутизатор
	router := server.SetupRouter(
		enterpriseService,
		reportService,
		riskService,
		comprehensiveReportService,
	)

	// Запускаем сервер
	serverAddr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	log.Printf("🚀 Starting Financial Risk Server on %s", serverAddr)

	// Используем стандартный http.Server
	server := &http.Server{
		Addr:         serverAddr,
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
	}

	// Запускаем сервер
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}

	log.Println("✅ Server stopped gracefully")
}
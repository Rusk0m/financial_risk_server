package main

import (
	"fmt"
	"log"

	"financial-risk-server/internal/config"
	"financial-risk-server/internal/repository/postgres"
	"financial-risk-server/internal/service"
)

func main() {
	fmt.Println("🧪 ТЕСТИРОВАНИЕ СЕРВИСОВ РАСЧЁТА РИСКОВ")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println()

	// 1. Загружаем конфигурацию и подключаемся к БД
	cfg, _ := config.Load("../internal/config/config.yaml")
	db, _ := postgres.Connect(&cfg.Database)
	defer db.Close()

	// 2. Создаём репозитории
	enterpriseRepo := postgres.NewEnterpriseRepository(db)
	exportRepo := postgres.NewExportContractRepository(db)
	marketRepo := postgres.NewMarketDataRepository(db)
	balanceRepo := postgres.NewBalanceSheetRepository(db)
	costRepo := postgres.NewCostStructureRepository(db)
	riskRepo := postgres.NewRiskCalculationRepository(db)
	reportRepo := postgres.NewComprehensiveReportRepository(db)

	// 3. Создаём фабрику и сервис расчёта рисков
	factory := service.NewServiceFactory(nil)
	riskService := factory.CreateRiskCalculationService(
		exportRepo,
		marketRepo,
		balanceRepo,
		costRepo,
		enterpriseRepo,
		riskRepo,
		reportRepo,
	)

	// 4. Получаем тестовое предприятие (Беларуськалий)
	enterprises, _ := enterpriseRepo.GetAll()
	if len(enterprises) == 0 {
		log.Fatal("❌ Нет предприятий в базе данных. Сначала выполните скрипт заполнения данных.")
	}
	enterpriseID := enterprises[0].ID

	fmt.Printf("🏢 Тестируемое предприятие: %s (ID: %d)\n", enterprises[0].Name, enterpriseID)
	fmt.Println()

	// 5. Тестируем расчёт валютного риска
	fmt.Println("1️⃣  Расчёт валютного риска (30 дней, 95% уровень доверия)...")
	currencyRisk, err := riskService.CalculateCurrencyRisk(enterpriseID, 30, 0.95)
	if err != nil {
		log.Fatalf("❌ Ошибка расчёта валютного риска: %v", err)
	}
	fmt.Printf("   ✅ Уровень риска: %s\n", currencyRisk.RiskLevel)
	fmt.Printf("   ✅ VaR (95%%): $%.2f млн\n", currencyRisk.VaRValue/1000000)
	fmt.Printf("   ✅ Стресс-тест (+25%% к курсу): $%.2f млн\n", currencyRisk.StressTestLoss/1000000)
	fmt.Println()

	// 6. Тестируем расчёт кредитного риска
	fmt.Println("2️⃣  Расчёт кредитного риска...")
	creditRisk, err := riskService.CalculateCreditRisk(enterpriseID)
	if err != nil {
		log.Fatalf("❌ Ошибка расчёта кредитного риска: %v", err)
	}
	fmt.Printf("   ✅ Уровень риска: %s\n", creditRisk.RiskLevel)
	fmt.Printf("   ✅ Ожидаемые потери: $%.2f млн\n", creditRisk.VaRValue/1000000)
	fmt.Printf("   ✅ Стресс-тест (дефолт крупнейшего контрагента): $%.2f млн\n", creditRisk.StressTestLoss/1000000)
	fmt.Println()

	// 7. Тестируем расчёт риска ликвидности
	fmt.Println("3️⃣  Расчёт риска ликвидности...")
	liquidityRisk, err := riskService.CalculateLiquidityRisk(enterpriseID)
	if err != nil {
		log.Fatalf("❌ Ошибка расчёта риска ликвидности: %v", err)
	}
	fmt.Printf("   ✅ Уровень риска: %s\n", liquidityRisk.RiskLevel)
	fmt.Printf("   ✅ Условный риск: $%.2f млн\n", liquidityRisk.VaRValue/1000000)
	fmt.Printf("   ✅ Стресс-тест (отток 30%% ликвидности): $%.2f млн\n", liquidityRisk.StressTestLoss/1000000)
	fmt.Println()

	// 8. Тестируем расчёт фондового риска
	fmt.Println("4️⃣  Расчёт фондового риска (падение цен на 15%)...")
	marketRisk, err := riskService.CalculateMarketRisk(enterpriseID, -15.0)
	if err != nil {
		log.Fatalf("❌ Ошибка расчёта фондового риска: %v", err)
	}
	fmt.Printf("   ✅ Уровень риска: %s\n", marketRisk.RiskLevel)
	fmt.Printf("   ✅ Изменение выручки: $%.2f млн (%.1f%%)\n", marketRisk.VaRValue/1000000, (marketRisk.VaRValue/marketRisk.ExposureAmount)*100)
	fmt.Printf("   ✅ Стресс-тест (-30%% к цене): $%.2f млн\n", marketRisk.StressTestLoss/1000000)
	fmt.Println()

	// 9. Тестируем комплексный расчёт
	fmt.Println("5️⃣  Комплексный расчёт всех рисков...")
	compReport, err := riskService.CalculateAllRisks(
		enterpriseID,
		30,
		0.95,
		-15.0,
		1.0,
	)
	if err != nil {
		log.Fatalf("❌ Ошибка комплексного расчёта: %v", err)
	}
	fmt.Printf("   ✅ Общий уровень риска: %s\n", compReport.OverallRiskLevel)
	fmt.Printf("   ✅ Суммарный риск: $%.2f млн\n", compReport.TotalRiskValue/1000000)
	fmt.Printf("   ✅ Критический риск: %s\n", compReport.MaxRiskType)
	fmt.Printf("   ✅ Комплексный отчёт сохранён (ID: %d)\n", compReport.ID)
	fmt.Println()

	fmt.Println("✅ ВСЕ ТЕСТЫ ПРОЙДЕНЫ УСПЕШНО!")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println()
	fmt.Println("💡 Результаты расчётов сохранены в базе данных.")
	fmt.Println("   Для просмотра выполните:")
	fmt.Println("   psql -U risk_user -d financial_risk_db -c \"SELECT risk_type, risk_level, ROUND(var_value/1000000, 2) AS var_mln FROM risk_calculations ORDER BY calculation_date DESC LIMIT 10;\"")
}

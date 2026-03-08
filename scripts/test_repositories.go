package main

// package main

// import (
// 	"fmt"
// 	"log"
// 	"time"

// 	"financial-risk-server/internal/config"
// 	"financial-risk-server/internal/domain/models"
// 	"financial-risk-server/internal/repository/postgres"
// )

// func main() {
// 	fmt.Println("🧪 ТЕСТИРОВАНИЕ РЕПОЗИТОРИЕВ")
// 	fmt.Println("═══════════════════════════════════════════════════════════════")
// 	fmt.Println()

// 	// 1. Загружаем конфигурацию
// 	cfg, err := config.Load("../internal/config/config.yaml")
// 	if err != nil {
// 		log.Fatalf("❌ Ошибка загрузки конфигурации: %v", err)
// 	}

// 	// 2. Подключаемся к БД
// 	fmt.Println("1️⃣  Подключение к базе данных...")
// 	db, err := postgres.Connect(&cfg.Database)
// 	if err != nil {
// 		log.Fatalf("❌ Ошибка подключения к БД: %v", err)
// 	}
// 	defer db.Close()
// 	fmt.Println("✅ Подключение успешно")
// 	fmt.Println()

// 	// 3. Создаём репозитории
// 	enterpriseRepo := postgres.NewEnterpriseRepository(db)
// 	exportRepo := postgres.NewExportContractRepository(db)
// 	costRepo := postgres.NewCostStructureRepository(db)
// 	balanceRepo := postgres.NewBalanceSheetRepository(db)
// 	marketRepo := postgres.NewMarketDataRepository(db)

// 	// 4. Создаём тестовое предприятие
// 	fmt.Println("2️⃣  Создание тестового предприятия...")
// 	enterprise := &models.Enterprise{
// 		Name:               "Тестовое предприятие «Беларуськалий»",
// 		Industry:           "Горнодобывающая (калийные удобрения)",
// 		AnnualProductionT:  8500000,
// 		ExportSharePercent: 95,
// 		MainCurrency:       "USD",
// 	}
// 	if err := enterpriseRepo.Create(enterprise); err != nil {
// 		log.Fatalf("❌ Ошибка создания предприятия: %v", err)
// 	}
// 	fmt.Printf("✅ Предприятие создано (ID: %d)\n", enterprise.ID)
// 	fmt.Println()

// 	// 5. Создаём экспортные контракты
// 	fmt.Println("3️⃣  Создание экспортных контрактов...")
// 	contracts := []*models.ExportContract{
// 		{
// 			EnterpriseID:    enterprise.ID,
// 			ContractDate:    time.Now().AddDate(0, 0, -10),
// 			Country:         "Китай",
// 			VolumeT:         350000,
// 			PriceUSDPerT:    285,
// 			Currency:        "USD",
// 			PaymentTermDays: 60,
// 			ShipmentDate:    time.Now().AddDate(0, 0, -5),
// 			PaymentStatus:   "pending",
// 			ExchangeRate:    3.25,
// 		},
// 		{
// 			EnterpriseID:    enterprise.ID,
// 			ContractDate:    time.Now().AddDate(0, 0, -8),
// 			Country:         "Бразилия",
// 			VolumeT:         220000,
// 			PriceUSDPerT:    282,
// 			Currency:        "USD",
// 			PaymentTermDays: 45,
// 			ShipmentDate:    time.Now().AddDate(0, 0, -3),
// 			PaymentStatus:   "pending",
// 			ExchangeRate:    3.26,
// 		},
// 		{
// 			EnterpriseID:    enterprise.ID,
// 			ContractDate:    time.Now().AddDate(0, 0, -5),
// 			Country:         "Индия",
// 			VolumeT:         130000,
// 			PriceUSDPerT:    280,
// 			Currency:        "USD",
// 			PaymentTermDays: 60,
// 			ShipmentDate:    time.Now(),
// 			PaymentStatus:   "pending",
// 			ExchangeRate:    3.27,
// 		},
// 	}

// 	for i, contract := range contracts {
// 		if err := exportRepo.Create(contract); err != nil {
// 			log.Fatalf("❌ Ошибка создания контракта #%d: %v", i+1, err)
// 		}
// 		fmt.Printf("   ✅ Контракт #%d создан (Страна: %s, Объём: %.0f т)\n", i+1, contract.Country, contract.VolumeT)
// 	}
// 	fmt.Println()

// 	// 6. Создаём структуру затрат
// 	fmt.Println("4️⃣  Создание структуры затрат...")
// 	costs := []*models.CostStructure{
// 		{
// 			EnterpriseID: enterprise.ID,
// 			CostItem:     "Заработная плата",
// 			Currency:     "BYN",
// 			CostPerT:     45.00,
// 		},
// 		{
// 			EnterpriseID: enterprise.ID,
// 			PeriodStart:  time.Now().AddDate(0, -1, 0),
// 			PeriodEnd:    time.Now().AddDate(1, 0, 0),
// 			Comment:      "Персонал горных работ, переработки, логистики",
// 		},
// 		{
// 			EnterpriseID: enterprise.ID,
// 			CostItem:     "Электроэнергия",
// 			Currency:     "BYN",
// 			CostPerT:     28.00,
// 			PeriodStart:  time.Now().AddDate(0, -1, 0),
// 			PeriodEnd:    time.Now().AddDate(1, 0, 0),
// 			Comment:      "Тарифы РУП «Белэнерго»",
// 		},
// 		{
// 			EnterpriseID: enterprise.ID,
// 		},
// 		{
// 			EnterpriseID: enterprise.ID,
// 			CostItem:     "Газ (импорт)",
// 			Currency:     "USD",
// 			CostPerT:     8.00,
// 			PeriodStart:  time.Now().AddDate(0, -1, 0),
// 			PeriodEnd:    time.Now().AddDate(1, 0, 0),
// 			Comment:      "Импорт из России",
// 		},
// 	}

// 	for i, cost := range costs {
// 		if err := costRepo.Create(cost); err != nil {
// 			log.Fatalf("❌ Ошибка создания статьи затрат #%d: %v", i+1, err)
// 		}
// 		fmt.Printf("   ✅ Статья затрат #%d создана (%s: %.2f %s/т)\n", i+1, cost.CostItem, cost.CostPerT, cost.Currency)
// 	}
// 	fmt.Println()

// 	// 7. Создаём финансовый баланс
// 	fmt.Println("5️⃣  Создание финансового баланса...")
// 	balance := &models.BalanceSheet{
// 		EnterpriseID:       enterprise.ID,
// 		ReportDate:         time.Now(),
// 		CashBYN:            1200000000,
// 		CashUSD:            150000000,
// 		AccountsReceivable: 2800000000,
// 		Inventories:        950000000,
// 		AccountsPayable:    1900000000,
// 		ShortTermDebt:      850000000,
// 		LongTermDebt:       4200000000,
// 	}
// 	if err := balanceRepo.Create(balance); err != nil {
// 		log.Fatalf("❌ Ошибка создания баланса: %v", err)
// 	}
// 	fmt.Printf("✅ Финансовый баланс создан (ID: %d)\n", balance.ID)
// 	fmt.Printf("   • Оборотные активы: %.2f млн BYN\n", (balance.CashBYN+balance.CashUSD*3.25+balance.AccountsReceivable+balance.Inventories)/1000000)
// 	fmt.Printf("   • Краткосрочные обязательства: %.2f млн BYN\n", (balance.AccountsPayable+balance.ShortTermDebt)/1000000)
// 	fmt.Println()

// 	// 8. Создаём рыночные данные
// 	fmt.Println("6️⃣  Создание рыночных данных...")
// 	marketData := []*models.MarketData{
// 		{
// 			DataDate:          time.Now(),
// 			CurrencyPair:      "BYN/USD",
// 			ExchangeRate:      3.25,
// 			Volatility30d:     0.09,
// 			PotassiumPriceUSD: 0,
// 			Source:            "НБ РБ",
// 		},
// 		{
// 			DataDate:          time.Now(),
// 			CurrencyPair:      "POTASSIUM",
// 			ExchangeRate:      3.25,
// 			Volatility30d:     0,
// 			PotassiumPriceUSD: 285,
// 			Source:            "World Bank",
// 		},
// 	}

// 	for i, data := range marketData {
// 		if err := marketRepo.Create(data); err != nil {
// 			log.Fatalf("❌ Ошибка создания рыночных данных #%d: %v", i+1, err)
// 		}
// 		if data.CurrencyPair == "BYN/USD" {
// 			fmt.Printf("   ✅ Курс BYN/USD: %.4f\n", data.ExchangeRate)
// 		} else {
// 			fmt.Printf("   ✅ Цена калия: $%.2f/т\n", data.PotassiumPriceUSD)
// 		}
// 	}
// 	fmt.Println()

// 	// 9. Тестируем получение данных из репозиториев
// 	fmt.Println("7️⃣  Тестирование методов получения данных...")

// 	// Тест 1: Получение предприятия по ID
// 	fmt.Println("   🔍 Тест 1: GetByID для предприятия...")
// 	enterpriseFromDB, err := enterpriseRepo.GetByID(enterprise.ID)
// 	if err != nil {
// 		log.Fatalf("   ❌ Ошибка GetByID: %v", err)
// 	}
// 	if enterpriseFromDB.Name != enterprise.Name {
// 		log.Fatalf("   ❌ Несоответствие данных: ожидалось '%s', получено '%s'", enterprise.Name, enterpriseFromDB.Name)
// 	}
// 	fmt.Printf("   ✅ Предприятие найдено: %s\n", enterpriseFromDB.Name)

// 	// Тест 2: Получение всех предприятий
// 	fmt.Println("   🔍 Тест 2: GetAll для предприятий...")
// 	enterprises, err := enterpriseRepo.GetAll()
// 	if err != nil {
// 		log.Fatalf("   ❌ Ошибка GetAll: %v", err)
// 	}
// 	fmt.Printf("   ✅ Получено предприятий: %d\n", len(enterprises))

// 	// Тест 3: Получение контрактов по предприятию
// 	fmt.Println("   🔍 Тест 3: GetByEnterpriseID для экспортных контрактов...")
// 	contractsFromDB, err := exportRepo.GetByEnterpriseID(enterprise.ID)
// 	if err != nil {
// 		log.Fatalf("   ❌ Ошибка GetByEnterpriseID: %v", err)
// 	}
// 	fmt.Printf("   ✅ Получено контрактов: %d\n", len(contractsFromDB))

// 	// Тест 4: Получение незакрытых контрактов
// 	fmt.Println("   🔍 Тест 4: GetPendingContracts для экспортных контрактов...")
// 	pendingContracts, err := exportRepo.GetPendingContracts(enterprise.ID)
// 	if err != nil {
// 		log.Fatalf("   ❌ Ошибка GetPendingContracts: %v", err)
// 	}
// 	fmt.Printf("   ✅ Получено незакрытых контрактов: %d\n", len(pendingContracts))

// 	// Тест 5: Получение затрат по предприятию
// 	fmt.Println("   🔍 Тест 5: GetByEnterpriseID для структуры затрат...")
// 	costsFromDB, err := costRepo.GetByEnterpriseID(enterprise.ID)
// 	if err != nil {
// 		log.Fatalf("   ❌ Ошибка GetByEnterpriseID для затрат: %v", err)
// 	}
// 	fmt.Printf("   ✅ Получено статей затрат: %d\n", len(costsFromDB))

// 	// Тест 6: Получение последнего баланса
// 	fmt.Println("   🔍 Тест 6: GetLatest для финансового баланса...")
// 	latestBalance, err := balanceRepo.GetLatest(enterprise.ID)
// 	if err != nil {
// 		log.Fatalf("   ❌ Ошибка GetLatest для баланса: %v", err)
// 	}
// 	fmt.Printf("   ✅ Последний баланс от: %s\n", latestBalance.ReportDate.Format("2006-01-02"))

// 	// Тест 7: Получение последних рыночных данных
// 	fmt.Println("   🔍 Тест 7: GetLatestByCurrencyPair для рыночных данных...")
// 	latestUSD, err := marketRepo.GetLatestByCurrencyPair("BYN/USD")
// 	if err != nil {
// 		log.Fatalf("   ❌ Ошибка GetLatestByCurrencyPair для BYN/USD: %v", err)
// 	}
// 	latestPotassium, err := marketRepo.GetLatestByCurrencyPair("POTASSIUM")
// 	if err != nil {
// 		log.Fatalf("   ❌ Ошибка GetLatestByCurrencyPair для POTASSIUM: %v", err)
// 	}
// 	fmt.Printf("   ✅ Последний курс BYN/USD: %.4f\n", latestUSD.ExchangeRate)
// 	fmt.Printf("   ✅ Последняя цена калия: $%.2f/т\n", latestPotassium.PotassiumPriceUSD)

// 	// Тест 8: Расчёт волатильности
// 	fmt.Println("   🔍 Тест 8: GetVolatility для рыночных данных...")
// 	// Для теста создадим дополнительные рыночные данные с разными курсами
// 	for i := 1; i <= 30; i++ {
// 		md := &models.MarketData{
// 			DataDate:     time.Now().AddDate(0, 0, -i),
// 			CurrencyPair: "BYN/USD",
// 			ExchangeRate: 3.25 + float64(i%5-2)*0.01, // Небольшие колебания
// 			Source:       "Тест",
// 		}
// 		_ = marketRepo.Create(md) // Игнорируем ошибки для теста
// 	}

// 	volatility, err := marketRepo.GetVolatility("BYN/USD", 30)
// 	if err != nil {
// 		log.Fatalf("   ❌ Ошибка GetVolatility: %v", err)
// 	}
// 	fmt.Printf("   ✅ Волатильность за 30 дней: %.2f%%\n", volatility*100)

// 	// Тест 9: Обновление предприятия
// 	fmt.Println("   🔍 Тест 9: Update для предприятия...")
// 	enterpriseFromDB.Industry = "Горнодобывающая (модифицировано для теста)"
// 	if err := enterpriseRepo.Update(enterpriseFromDB); err != nil {
// 		log.Fatalf("   ❌ Ошибка Update: %v", err)
// 	}
// 	// Проверяем обновление
// 	updatedEnterprise, err := enterpriseRepo.GetByID(enterprise.ID)
// 	if err != nil {
// 		log.Fatalf("   ❌ Ошибка получения после обновления: %v", err)
// 	}
// 	if updatedEnterprise.Industry != enterpriseFromDB.Industry {
// 		log.Fatalf("   ❌ Обновление не применилось")
// 	}
// 	fmt.Printf("   ✅ Предприятие успешно обновлено\n")

// 	// Тест 10: Удаление предприятия (и каскадное удаление зависимостей)
// 	// Пропускаем удаление, чтобы сохранить данные для дальнейших тестов
// 	// В реальном тесте можно раскомментировать:
// 	// fmt.Println("   🔍 Тест 10: Delete для предприятия...")
// 	// if err := enterpriseRepo.Delete(enterprise.ID); err != nil {
// 	// 	log.Fatalf("   ❌ Ошибка удаления: %v", err)
// 	// }
// 	// fmt.Printf("   ✅ Предприятие успешно удалено\n")

// 	fmt.Println()
// 	fmt.Println("✅ ВСЕ ТЕСТЫ ПРОЙДЕНЫ УСПЕШНО!")
// 	fmt.Println("═══════════════════════════════════════════════════════════════")
// 	fmt.Println()
// 	fmt.Printf("📊 Сводка:\n")
// 	fmt.Printf("   • Предприятий создано: 1 (ID: %d)\n", enterprise.ID)
// 	fmt.Printf("   • Экспортных контрактов: %d\n", len(contracts))
// 	fmt.Printf("   • Статей затрат: %d\n", len(costs))
// 	fmt.Printf("   • Финансовых балансов: 1 (ID: %d)\n", balance.ID)
// 	fmt.Printf("   • Рыночных данных: %d записей\n", len(marketData)+30) // +30 для волатильности
// 	fmt.Println()
// 	fmt.Println("💡 Данные сохранены в базе для дальнейшего использования.")
// 	fmt.Println("   Для очистки тестовых данных выполните:")
// 	fmt.Println("   psql -U risk_user -d financial_risk_db -c \"DELETE FROM enterprises WHERE name LIKE 'Тестовое предприятие%';\"")
// }

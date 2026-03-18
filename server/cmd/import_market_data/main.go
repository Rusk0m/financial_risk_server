package main

import (
	"context"
	"flag"
	"log"
	"time"

	"financial-risk-server/internal/config"
	"financial-risk-server/internal/repository/postgres"
	importer "financial-risk-server/scripts"
)

func main() {
	// Аргументы командной строки
	//configPath := flag.String("config", "", "Путь к config.yaml (опционально)")
	filePath := flag.String("file", "/home/ruskom/Документы/Отчетность белкалия/мировые_цены_калий.xlsx", "Путь к Excel-файлу")
	dryRun := flag.Bool("dry-run", false, "Режим проверки без записи в БД")
	
	// Настройки данных
	currencyPair := flag.String("pair", "POTASSIUM", "Значение currency_pair в БД")
	source := flag.String("source", "Excel import", "Значение поля source в БД")
	dateFormat := flag.String("date-format", "2006-01-02", "Формат даты в Excel")
	
	flag.Parse()

	log.Printf("🚀 Starting potassium price import")
	log.Printf("   File: %s", *filePath)
	log.Printf("   Currency pair: %s", *currencyPair)
	log.Printf("   Source: %s", *source)
	log.Printf("   Date format: %s", *dateFormat)
	log.Printf("   Dry run: %v", *dryRun)

	// Загружаем конфигурацию
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌ Failed to load config: %v", err)
	}

	// Подключаемся к БД
	log.Println("⏳ Connecting to database...")
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
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}
	defer db.Close()
	log.Println("✅ Database connected")

	// Создаём репозиторий
	marketRepo := postgres.NewMarketDataRepository(db)

	// Запускаем импорт
	start := time.Now()
	stats, err := importer.ImportPotassiumPrices(
		context.Background(),
		marketRepo,
		*filePath,
		importer.PotassiumImportOptions{
			CurrencyPair: *currencyPair,
			Source:       *source,
			DateFormat:   *dateFormat,
			DryRun:       *dryRun,
			SkipDuplicates: true,
		},
	)
	
	if err != nil {
		log.Fatalf("❌ Import failed: %v", err)
	}

	// Выводим статистику
	elapsed := time.Since(start)
	log.Println("✅ Import completed successfully")
	log.Printf("📊 Statistics:")
	log.Printf("   Total rows processed: %d", stats.TotalRows)
	log.Printf("   Successfully saved: %d", stats.SavedCount)
	log.Printf("   Skipped (duplicates): %d", stats.SkippedCount)
	log.Printf("   Errors: %d", stats.ErrorCount)
	log.Printf("   Date range: %s to %s", stats.MinDate, stats.MaxDate)
	log.Printf("   Time elapsed: %v", elapsed.Round(time.Second))
}
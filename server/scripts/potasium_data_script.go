package importer

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"

	"financial-risk-server/internal/domain/models"
	"financial-risk-server/internal/repository/interfaces"
)

// PotassiumImportOptions настройки импорта цен на калий
type PotassiumImportOptions struct {
	CurrencyPair   string // Значение для currency_pair (обычно "POTASSIUM")
	Source         string // Значение для поля source
	DateFormat     string // Формат даты в Excel (по умолчанию: "2006-01-02")
	DryRun         bool   // Не записывать в БД, только проверка
	SkipDuplicates bool   // Пропускать уже существующие записи
}

// ImportStats статистика импорта
type ImportStats struct {
	TotalRows    int
	SavedCount   int
	SkippedCount int
	ErrorCount   int
	MinDate      string
	MaxDate      string
	Errors       []string
}

// ImportPotassiumPrices импортирует цены на калий из простого Excel-файла
// Формат файла:
//   Колонка A: дата (ГГГГ-ММ-ДД)
//   Колонка B: цена в USD
//   Без заголовков
func ImportPotassiumPrices(
	ctx context.Context,
	repo interfaces.MarketDataRepository,
	filePath string,
	opts PotassiumImportOptions,
) (*ImportStats, error) {
	
	log.Printf("📂 Opening Excel file: %s", filePath)
	
	file, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open Excel file: %w", err)
	}
	defer file.Close()

	stats := &ImportStats{}
	sheets := file.GetSheetList()

	if len(sheets) == 0 {
		return nil, fmt.Errorf("Excel file has no sheets")
	}

	// Берём первый лист
	sheetName := sheets[0]
	log.Printf("📋 Processing sheet: %s", sheetName)
	
	rows, err := file.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to read sheet %s: %w", sheetName, err)
	}

	if len(rows) == 0 {
		return nil, fmt.Errorf("sheet %s is empty", sheetName)
	}

	log.Printf("📊 Found %d rows in sheet", len(rows))

	var minDate, maxDate time.Time
	firstDate := true

	for i, row := range rows {
		rowNum := i + 1 // 1-based номер строки

		// Пропускаем пустые строки
		if isEmptyRow(row) {
			continue
		}

		// Ожидаем минимум 2 колонки: дата и цена
		if len(row) < 2 {
			stats.ErrorCount++
			stats.Errors = append(stats.Errors, fmt.Sprintf("row %d: not enough columns (got %d, need 2)", rowNum, len(row)))
			log.Printf("⚠️  Row %d skipped: not enough columns", rowNum)
			continue
		}

		// === Парсим дату (колонка A) ===
		dateStr := trim(row[0])
		if dateStr == "" {
			stats.ErrorCount++
			stats.Errors = append(stats.Errors, fmt.Sprintf("row %d: empty date", rowNum))
			continue
		}

		dataDate, err := parseExcelDate(row[0])
		if err != nil || dataDate.IsZero() {
			log.Printf("⚠️  [Parser] Row %d skipped: invalid agreement date '%s': %v", i+1, row[1], err)
			stats.ErrorCount++
			continue
		}

		// === Парсим цену (колонка B) ===
		priceStr := trim(row[1])
		if priceStr == "" {
			stats.ErrorCount++
			stats.Errors = append(stats.Errors, fmt.Sprintf("row %d: empty price", rowNum))
			continue
		}

		price, err := parseFloat(priceStr)
		if err != nil {
			stats.ErrorCount++
			stats.Errors = append(stats.Errors, fmt.Sprintf("row %d: invalid price '%s': %v", rowNum, priceStr, err))
			log.Printf("⚠️  Row %d skipped: invalid price '%s'", rowNum, priceStr)
			continue
		}

		if price <= 0 {
			stats.ErrorCount++
			stats.Errors = append(stats.Errors, fmt.Sprintf("row %d: price must be positive (got %.2f)", rowNum, price))
			continue
		}

		stats.TotalRows++

		// Обновляем диапазон дат
		if firstDate || dataDate.Before(minDate) {
			minDate = dataDate
		}
		if firstDate || dataDate.After(maxDate) {
			maxDate = dataDate
		}
		firstDate = false

		// Проверяем дубликаты если нужно
		if opts.SkipDuplicates {
			existing, err := repo.GetByDateAndPair(ctx, dataDate, opts.CurrencyPair)
			if err != nil {
				log.Printf("⚠️  Failed to check duplicate for %s: %v", dataDate.Format("2006-01-02"), err)
			} else if existing != nil {
				stats.SkippedCount++
				log.Printf("⏭️  Row %d skipped (duplicate): %s / $%.2f", rowNum, dataDate.Format("2006-01-02"), price)
				continue
			}
		}

		// Создаём запись
		marketData := &models.MarketData{
			DataDate:          dataDate,
			CurrencyPair:      opts.CurrencyPair,
			ExchangeRate:      nil, // Не заполняем для калия
			PotassiumPriceUSD: &price,
			Source:            opts.Source,
			CreatedAt:         time.Now(),
		}

		// Dry run: только логирование
		if opts.DryRun {
			stats.SavedCount++
			log.Printf("✅ [DRY RUN] Would save: %s / %s / $%.2f",
				dataDate.Format("2006-01-02"),
				opts.CurrencyPair,
				price,
			)
			continue
		}

		// Сохраняем в БД
		if err := repo.Create(ctx, marketData); err != nil {
			stats.ErrorCount++
			stats.Errors = append(stats.Errors, fmt.Sprintf("row %d DB error: %v", rowNum, err))
			log.Printf("❌ Row %d failed to save: %v", rowNum, err)
			continue
		}

		stats.SavedCount++
		if stats.TotalRows%10 == 0 {
			log.Printf("✅ Saved %d/%d rows...", stats.SavedCount, stats.TotalRows)
		}
	}

	// Сохраняем диапазон дат в статистику
	if !firstDate {
		stats.MinDate = minDate.Format("2006-01-02")
		stats.MaxDate = maxDate.Format("2006-01-02")
	}

	return stats, nil
}

// === Вспомогательные функции ===

// trim удаляет пробелы и невидимые символы
func trim(s string) string {
	if s == "" {
		return ""
	}
	result := ""
	for _, r := range s {
		if r != ' ' && r != '\t' && r != '\n' && r != '\r' && r != '\u00A0' {
			result += string(r)
		}
	}
	return result
}

// parseFloat парсит строку в float64 (поддерживает запятую как разделитель)
func parseFloat(s string) (float64, error) {
	s = trim(s)
	if s == "" {
		return 0, fmt.Errorf("empty value")
	}
	
	// Заменяем запятую на точку для совместимости с русским Excel
	s = replaceCommaWithDot(s)
	
	var result float64
	_, err := fmt.Sscanf(s, "%f", &result)
	return result, err
}

// replaceCommaWithDot заменяет запятую на точку в числе
func replaceCommaWithDot(s string) string {
	// Простая замена: если есть запятая и нет точки — заменяем
	if len(s) > 0 && s[len(s)-1] == ',' {
		return s[:len(s)-1] + "."
	}
	// Если запятая в середине (как десятичный разделитель)
	for i, c := range s {
		if c == ',' && i > 0 && i < len(s)-1 {
			// Проверяем, что это не разделитель тысяч
			after := s[i+1:]
			if len(after) == 2 || len(after) == 3 {
				// Скорее всего десятичный разделитель
				return s[:i] + "." + after
			}
		}
	}
	return s
}

// isEmptyRow проверяет, пустая ли строка
func isEmptyRow(row []string) bool {
	for _, cell := range row {
		if trim(cell) != "" {
			return false
		}
	}
	return true
}

func excelDateToTime(excelDate float64) (time.Time, error) {
	if excelDate == 0 {
		return time.Time{}, nil
	}
	
	// Excel epoch: 30 декабря 1899
	// Но есть баг: Excel считает 1900 високосным (хотя это не так)
	// Поэтому для дат после 28.02.1900 нужно вычесть 1 день
	epoch := time.Date(1899, 12, 30, 0, 0, 0, 0, time.UTC)
	
	result := epoch.AddDate(0, 0, int(excelDate))
	
	// Коррекция бага 1900 года
	if excelDate > 59 { // 59 = 28.02.1900
		result = result.AddDate(0, 0, -1)
	}
	
	return result, nil
}

// parseExcelDate пытается распарсить дату из Excel (строка или серийный номер)
func parseExcelDate(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, nil
	}
	
	// Сначала пробуем как обычную строку ДД.ММ.ГГГГ
	if t, err := time.Parse("02.01.2006", value); err == nil {
		return t, nil
	}
	
	// Если не вышло — пробуем как серийный номер Excel
	if serial, err := strconv.ParseFloat(value, 64); err == nil && serial > 0 {
		return excelDateToTime(serial)
	}
	
	return time.Time{}, fmt.Errorf("unable to parse date: %s", value)
}
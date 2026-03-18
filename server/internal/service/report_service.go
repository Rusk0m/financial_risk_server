package service

import (
	"context"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"financial-risk-server/internal/domain/models"
	"financial-risk-server/internal/repository/interfaces"

	"github.com/xuri/excelize/v2"
)

// ReportService предоставляет бизнес-логику для работы с отчётами
type ReportService struct {
	reportRepo        interfaces.ReportRepository
	contractRepo      interfaces.ExportContractRepository
	balanceRepo       interfaces.BalanceSheetRepository
	finResultRepo     interfaces.FinancialResultRepository
	creditRepo        interfaces.CreditAgreementRepository
	enterpriseService *EnterpriseService
	uploadDir         string
}

// ReportServiceConfig конфигурация сервиса отчётов
type ReportServiceConfig struct {
	UploadDirectory string // Папка для загрузки файлов (например: "./uploads")
}

// NewReportService создаёт новый сервис отчётов
func NewReportService(
	reportRepo interfaces.ReportRepository,
	contractRepo interfaces.ExportContractRepository,
	balanceRepo interfaces.BalanceSheetRepository,
	finResultRepo interfaces.FinancialResultRepository,
	creditRepo interfaces.CreditAgreementRepository,
	enterpriseService *EnterpriseService,
	config *ReportServiceConfig,
) *ReportService {
	// Создаём папку для загрузок, если она не существует
	uploadDir := "./uploads"
	if config != nil && config.UploadDirectory != "" {
		uploadDir = config.UploadDirectory
	}

	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		panic(fmt.Sprintf("failed to create upload directory %s: %v", uploadDir, err))
	}

	return &ReportService{
		reportRepo:        reportRepo,
		contractRepo:      contractRepo,
		balanceRepo:       balanceRepo,
		finResultRepo:     finResultRepo,
		creditRepo:        creditRepo,
		enterpriseService: enterpriseService,
		uploadDir:         uploadDir,
	}
}

// UploadReport загружает и обрабатывает отчёт
func (s *ReportService) UploadReport(
	ctx context.Context,
	file *multipart.FileHeader,
	req *models.ReportUploadRequest,
) (*models.Report, error) {
	// 1. Валидация
	if file == nil {
		return nil, fmt.Errorf("file is required")
	}
	if file.Size == 0 {
		return nil, fmt.Errorf("file is empty")
	}
	if file.Size > 10*1024*1024 {
		return nil, fmt.Errorf("file size exceeds 10 MB limit")
	}

	// 2. Генерация имени и сохранение
	uniqueFileName := s.generateUniqueFileName(req.ReportType, file.Filename)
	filePath, err := s.saveFile(file, uniqueFileName)
	if err != nil {
		return nil, fmt.Errorf("failed to save file: %w", err)
	}

	// 3. Создание записи в БД
	report := &models.Report{
		EnterpriseID:     req.EnterpriseID,
		ReportType:       req.ReportType,
		FileName:         uniqueFileName,
		OriginalName:     file.Filename,
		FilePath:         filePath,
		FileSizeBytes:    file.Size,
		UploadDate:       time.Now(),
		ProcessingStatus: "pending",
		UploadedBy:       req.UploadedBy,
		Description:      req.Description,
	}

	if req.PeriodStart != "" {
		if t, err := time.Parse("2006-01-02", req.PeriodStart); err == nil {
			report.PeriodStart = &t
		}
	}
	if req.PeriodEnd != "" {
		if t, err := time.Parse("2006-01-02", req.PeriodEnd); err == nil {
			report.PeriodEnd = &t
		}
	}

	if err := s.reportRepo.Create(ctx, report); err != nil {
		_ = os.Remove(filePath)
		return nil, fmt.Errorf("failed to create report record: %w", err)
	}

	// 4. ✅ Асинхронная обработка с логированием
	go func(reportID int64, filePath string, reportType string) {
		ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		log.Printf("🔄 [Parser] Starting async processing: report_id=%d, type=%s", reportID, reportType)

		if err := s.processReportFile(ctxWithTimeout, reportID, filePath, reportType); err != nil {
			log.Printf("❌ [Parser] Failed to process report_id=%d: %v", reportID, err)
			errorMsg := err.Error()
			_ = s.reportRepo.UpdateProcessingStatus(ctxWithTimeout, reportID, "error", &errorMsg)
		} else {
			log.Printf("✅ [Parser] Successfully processed report_id=%d", reportID)
			_ = s.reportRepo.UpdateProcessingStatus(ctxWithTimeout, reportID, "processed", nil)
		}
	}(report.ID, filePath, req.ReportType)

	return report, nil
}
// processReportFile парсит файл и сохраняет данные в соответствующую таблицу
func (s *ReportService) processReportFile(ctx context.Context, reportID int64, filePath, reportType string) error {
	log.Printf("📂 [Parser] Opening file: %s", filePath)

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open report file: %w", err)
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(filePath))
	if ext != ".xlsx" && ext != ".xls" {
		return fmt.Errorf("unsupported file format: %s", ext)
	}

	xlFile, err := excelize.OpenFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to open Excel file: %w", err)
	}
	defer xlFile.Close()

	log.Printf("📊 [Parser] Parsing report type: %s", reportType)

	switch reportType {
	case "export_contracts":
		return s.parseExportContracts(ctx, xlFile, reportID)
	case "balance_sheet":
		return s.parseBalanceSheet(ctx, xlFile, reportID)
	case "financial_results":
		return s.parseFinancialResults(ctx, xlFile, reportID)
	case "credit_agreements":
		return s.parseCreditAgreements(ctx, xlFile, reportID)
	default:
		return fmt.Errorf("unsupported report type: %s", reportType)
	}
}

// parseExportContracts парсит экспортные контракты из Excel
func (s *ReportService) parseExportContracts(ctx context.Context, xlFile *excelize.File, reportID int64) error {
	enterpriseID := int64(1)

	sheetName := xlFile.GetSheetName(0)
	rows, err := xlFile.GetRows(sheetName)
	if err != nil {
		return fmt.Errorf("failed to get rows from sheet: %w", err)
	}

	log.Printf("📋 [Parser] Found %d total rows in sheet '%s' (export_contracts)", len(rows), sheetName)

	if len(rows) < 2 {
		return fmt.Errorf("export contracts file is empty or has no data rows")
	}

	// 🔍 Лог заголовков
	if len(rows) > 0 {
		log.Printf("📋 [Parser] Header row: %v", rows[0])
	}

	contracts := []*models.ExportContract{}
	skippedCount := 0

	for i := 1; i < len(rows); i++ {
		row := rows[i]
		log.Printf("🔍 [Parser] Row %d: %v (len=%d)", i+1, row, len(row))

		// Минимум колонок: номер, дата, страна, объём, цена, валюта, срок, курс(опц)
		if len(row) < 7 {
			log.Printf("⚠️  [Parser] Row %d skipped: not enough columns (got %d, need 7)", i+1, len(row))
			skippedCount++
			continue
		}

		// ✅ Парсим дату контракта (поддерживает строки и Excel serial)
		contractDate, err := parseExcelDate(row[1])
		if err != nil || contractDate.IsZero() {
			log.Printf("⚠️  [Parser] Row %d skipped: invalid contract date '%s': %v", i+1, row[1], err)
			skippedCount++
			continue
		}

		// Парсим объём
		volumeT, err := parseFloat(strings.TrimSpace(row[3]))
		if err != nil || volumeT <= 0 {
			log.Printf("⚠️  [Parser] Row %d skipped: invalid volume '%s': %v", i+1, row[3], err)
			skippedCount++
			continue
		}

		// Парсим цену
		priceContract, err := parseFloat(strings.TrimSpace(row[4]))
		if err != nil || priceContract <= 0 {
			log.Printf("⚠️  [Parser] Row %d skipped: invalid price '%s': %v", i+1, row[4], err)
			skippedCount++
			continue
		}

		// Парсим курс (опционально, есть дефолт)
		exchangeRate := 3.25 // Курс по умолчанию
		if len(row) > 7 && row[7] != "" {
			if rate, err := parseFloat(strings.TrimSpace(row[7])); err == nil && rate > 0 {
				exchangeRate = rate
			}
		}

		// Парсим срок оплаты (опционально)
		paymentTermDays := 60
		if len(row) > 6 && row[6] != "" {
			if term, err := strconv.Atoi(strings.TrimSpace(row[6])); err == nil && term > 0 {
				paymentTermDays = term
			}
		}

		// Рассчитываем дату отгрузки
		shipmentDate := contractDate.AddDate(0, 0, 30)
		if len(row) > 8 && row[8] != "" {
			if sd, err := parseExcelDate(row[8]); err == nil && !sd.IsZero() {
				shipmentDate = sd
			}
		}

		contract := &models.ExportContract{
			EnterpriseID:    enterpriseID,
			ReportID:        &reportID,
			ContractNumber:  strings.TrimSpace(row[0]),
			ContractDate:    contractDate,
			Country:         strings.TrimSpace(row[2]),
			VolumeT:         volumeT,
			PriceContract:   priceContract,
			Currency:        strings.TrimSpace(row[5]),
			PaymentTermDays: paymentTermDays,
			ShipmentDate:    shipmentDate,
			PaymentStatus:   "pending",
			ExchangeRate:    exchangeRate,
		}

		contracts = append(contracts, contract)
		log.Printf("✅ [Parser] Row %d parsed: contract=%s, date=%s, volume=%.0f t",
			i+1, contract.ContractNumber, contractDate.Format("02.01.2006"), volumeT)
	}

	log.Printf("📊 [Parser] Export contracts summary: parsed %d, skipped %d", len(contracts), skippedCount)

	if len(contracts) == 0 {
		return fmt.Errorf("no valid export contracts found in file")
	}

	// Сохраняем в БД
	for i, contract := range contracts {
		if err := s.contractRepo.Create(ctx, contract); err != nil {
			log.Printf("❌ [Parser] Failed to save contract %d: %v", i+1, err)
			return fmt.Errorf("failed to save export contract #%d: %w", i+1, err)
		}
		log.Printf("💾 [Parser] Saved contract %d: %s", i+1, contract.ContractNumber)
	}

	return nil
}
// parseBalanceSheet парсит финансовый баланс из Excel
func (s *ReportService) parseBalanceSheet(ctx context.Context, xlFile *excelize.File, reportID int64) error {
	enterpriseID := int64(1)

	sheetName := xlFile.GetSheetName(0)
	
	// 🔴 Получаем количество строк
	rows, err := xlFile.GetRows(sheetName)
	if err != nil {
		return fmt.Errorf("failed to get rows from sheet: %w", err)
	}

	log.Printf("📋 [Parser] Found %d total rows in sheet '%s' (balance_sheet)", len(rows), sheetName)

	if len(rows) < 2 {
		return fmt.Errorf("balance sheet file is empty or has no data rows")
	}

	// 🔍 Лог заголовков
	if len(rows) > 0 {
		log.Printf("📋 [Parser] Header row: %v", rows[0])
	}

	balances := []*models.BalanceSheet{}
	skippedCount := 0

	// 🔴 Парсим ВСЕ строки данных (начиная с индекса 1)
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		
		// 🔍 Лог для отладки
		log.Printf("🔍 [Parser] Row %d: %v (len=%d)", i+1, row, len(row))

		// Пропускаем пустые строки
		if len(row) < 2 || isEmptyRow(row) {
			log.Printf("⚠️  [Parser] Row %d skipped: empty row", i+1)
			skippedCount++
			continue
		}

		// ✅ Парсим дату отчёта через GetCellValue (более надёжно)
		cellDate, _ := xlFile.GetCellValue(sheetName, fmt.Sprintf("A%d", i+1))
		reportDate, err := parseExcelDate(cellDate)
		if err != nil || reportDate.IsZero() {
			// Пробуем из rows
			reportDate, err = parseExcelDate(row[0])
			if err != nil || reportDate.IsZero() {
				log.Printf("⚠️  [Parser] Row %d skipped: invalid report date '%s': %v", i+1, row[0], err)
				skippedCount++
				continue
			}
		}
		log.Printf("📅 [Parser] Row %d - Report date: %s", i+1, reportDate.Format("02.01.2006"))

		// 🔴 Парсим каждое поле через GetCellValue для надёжности
		cashBYN := getFloatFromCell(xlFile, sheetName, i+1, 2)  // Колонка B
		cashUSD := getFloatFromCell(xlFile, sheetName, i+1, 3)  // Колонка C
		shortTermInvestments := getFloatFromCell(xlFile, sheetName, i+1, 4)
		accountsReceivable := getFloatFromCell(xlFile, sheetName, i+1, 5)
		vatReceivable := getFloatFromCell(xlFile, sheetName, i+1, 6)
		inventories := getFloatFromCell(xlFile, sheetName, i+1, 7)
		rawMaterials := getFloatFromCell(xlFile, sheetName, i+1, 8)
		workInProgress := getFloatFromCell(xlFile, sheetName, i+1, 9)
		finishedGoods := getFloatFromCell(xlFile, sheetName, i+1, 10)
		propertyPlantEquipment := getFloatFromCell(xlFile, sheetName, i+1, 11)
		intangibleAssets := getFloatFromCell(xlFile, sheetName, i+1, 12)
		accountsPayable := getFloatFromCell(xlFile, sheetName, i+1, 13)
		vatPayable := getFloatFromCell(xlFile, sheetName, i+1, 14)
		payrollPayable := getFloatFromCell(xlFile, sheetName, i+1, 15)
		shortTermDebt := getFloatFromCell(xlFile, sheetName, i+1, 16)
		longTermDebt := getFloatFromCell(xlFile, sheetName, i+1, 17)
		authorizedCapital := getFloatFromCell(xlFile, sheetName, i+1, 18)
		retainedEarnings := getFloatFromCell(xlFile, sheetName, i+1, 19)

		// 🔍 Лог ключевых показателей
		log.Printf("💰 [Parser] Row %d - Cash BYN=%.0f, Cash USD=%.0f, Receivables=%.0f, Payables=%.0f",
			i+1, cashBYN, cashUSD, accountsReceivable, accountsPayable)

		balance := &models.BalanceSheet{
			EnterpriseID:           enterpriseID,
			ReportID:               &reportID,
			ReportDate:             reportDate,
			CashBYN:                cashBYN,
			CashUSD:                cashUSD,
			ShortTermInvestments:   shortTermInvestments,
			AccountsReceivable:     accountsReceivable,
			VATReceivable:          vatReceivable,
			Inventories:            inventories,
			RawMaterials:           rawMaterials,
			WorkInProgress:         workInProgress,
			FinishedGoods:          finishedGoods,
			PropertyPlantEquipment: propertyPlantEquipment,
			IntangibleAssets:       intangibleAssets,
			AccountsPayable:        accountsPayable,
			VATPayable:             vatPayable,
			PayrollPayable:         payrollPayable,
			ShortTermDebt:          shortTermDebt,
			LongTermDebt:           longTermDebt,
			AuthorizedCapital:      authorizedCapital,
			RetainedEarnings:       retainedEarnings,
		}

		balances = append(balances, balance)
		log.Printf("✅ [Parser] Row %d parsed successfully", i+1)
	}

	log.Printf("📊 [Parser] Balance sheet summary: parsed %d, skipped %d", len(balances), skippedCount)

	if len(balances) == 0 {
		return fmt.Errorf("no valid balance sheet records found in file")
	}

	// 🔴 Сохраняем ВСЕ записи в БД
	for i, balance := range balances {
		if err := s.balanceRepo.Create(ctx, balance); err != nil {
			log.Printf("❌ [Parser] Failed to save balance %d: %v", i+1, err)
			return fmt.Errorf("failed to save balance sheet #%d: %w", i+1, err)
		}
		log.Printf("💾 [Parser] Saved balance sheet %d: date=%s", i+1, balance.ReportDate.Format("02.01.2006"))
	}

	return nil
}


// parseFinancialResults парсит отчёт о финансовых результатах из Excel
func (s *ReportService) parseFinancialResults(ctx context.Context, xlFile *excelize.File, reportID int64) error {
	enterpriseID := int64(1)

	sheetName := xlFile.GetSheetName(0)
	rows, err := xlFile.GetRows(sheetName)
	if err != nil {
		return fmt.Errorf("failed to get rows from sheet: %w", err)
	}

	log.Printf("📋 [Parser] Found %d total rows in sheet '%s' (financial_results)", len(rows), sheetName)

	if len(rows) < 2 {
		return fmt.Errorf("financial results file is empty or has no data rows")
	}

	if len(rows) > 0 {
		log.Printf("📋 [Parser] Header row: %v", rows[0])
	}

	results := []*models.FinancialResult{}
	skippedCount := 0

	// 🔴 Парсим ВСЕ строки
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		
		if len(row) < 2 || isEmptyRow(row) {
			log.Printf("⚠️  [Parser] Row %d skipped: empty row", i+1)
			skippedCount++
			continue
		}

		// Парсим дату
		cellDate, _ := xlFile.GetCellValue(sheetName, fmt.Sprintf("A%d", i+1))
		reportDate, err := parseExcelDate(cellDate)
		if err != nil || reportDate.IsZero() {
			reportDate, err = parseExcelDate(row[0])
			if err != nil || reportDate.IsZero() {
				log.Printf("⚠️  [Parser] Row %d skipped: invalid report date", i+1)
				skippedCount++
				continue
			}
		}

		// Парсим период
		cellPeriodStart, _ := xlFile.GetCellValue(sheetName, fmt.Sprintf("B%d", i+1))
		periodStart, _ := parseExcelDate(cellPeriodStart)
		if periodStart.IsZero() {
			periodStart = reportDate.AddDate(0, -3, 0)
		}

		cellPeriodEnd, _ := xlFile.GetCellValue(sheetName, fmt.Sprintf("C%d", i+1))
		periodEnd, _ := parseExcelDate(cellPeriodEnd)
		if periodEnd.IsZero() {
			periodEnd = reportDate
		}

		// 🔴 Парсим все поля через GetCellValue
		revenueSales := getFloatFromCell(xlFile, sheetName, i+1, 4)
		revenueExport := getFloatFromCell(xlFile, sheetName, i+1, 5)
		revenueDomestic := getFloatFromCell(xlFile, sheetName, i+1, 6)
		revenueOther := getFloatFromCell(xlFile, sheetName, i+1, 7)
		revenueTotal := getFloatFromCell(xlFile, sheetName, i+1, 8)

		costOfSales := getFloatFromCell(xlFile, sheetName, i+1, 9)
		costRawMaterials := getFloatFromCell(xlFile, sheetName, i+1, 10)
		costEnergy := getFloatFromCell(xlFile, sheetName, i+1, 11)
		costLabor := getFloatFromCell(xlFile, sheetName, i+1, 12)
		costDepreciation := getFloatFromCell(xlFile, sheetName, i+1, 13)
		costOther := getFloatFromCell(xlFile, sheetName, i+1, 14)
		costTotal := getFloatFromCell(xlFile, sheetName, i+1, 15)

		commercialExpenses := getFloatFromCell(xlFile, sheetName, i+1, 16)
		administrativeExpenses := getFloatFromCell(xlFile, sheetName, i+1, 17)
		otherExpenses := getFloatFromCell(xlFile, sheetName, i+1, 18)

		grossProfit := getFloatFromCell(xlFile, sheetName, i+1, 19)
		operatingProfit := getFloatFromCell(xlFile, sheetName, i+1, 20)
		profitBeforeTax := getFloatFromCell(xlFile, sheetName, i+1, 21)
		taxExpense := getFloatFromCell(xlFile, sheetName, i+1, 22)
		netProfit := getFloatFromCell(xlFile, sheetName, i+1, 23)

		ebitda := getFloatFromCell(xlFile, sheetName, i+1, 24)
		operatingMargin := getFloatFromCell(xlFile, sheetName, i+1, 25)
		netMargin := getFloatFromCell(xlFile, sheetName, i+1, 26)

		log.Printf("💰 [Parser] Row %d - Revenue: total=%.0f, Net profit: %.0f", i+1, revenueTotal, netProfit)

		result := &models.FinancialResult{
			EnterpriseID:           enterpriseID,
			ReportID:               &reportID,
			ReportDate:             reportDate,
			PeriodStart:            periodStart,
			PeriodEnd:              periodEnd,
			RevenueSales:           revenueSales,
			RevenueExport:          revenueExport,
			RevenueDomestic:        revenueDomestic,
			RevenueOther:           revenueOther,
			RevenueTotal:           revenueTotal,
			CostOfSales:            costOfSales,
			CostRawMaterials:       costRawMaterials,
			CostEnergy:             costEnergy,
			CostLabor:              costLabor,
			CostDepreciation:       costDepreciation,
			CostOther:              costOther,
			CostTotal:              costTotal,
			CommercialExpenses:     commercialExpenses,
			AdministrativeExpenses: administrativeExpenses,
			OtherExpenses:          otherExpenses,
			GrossProfit:            grossProfit,
			OperatingProfit:        operatingProfit,
			ProfitBeforeTax:        profitBeforeTax,
			TaxExpense:             taxExpense,
			NetProfit:              netProfit,
			EBITDA:                 &ebitda,
			OperatingMargin:        &operatingMargin,
			NetMargin:              &netMargin,
		}

		results = append(results, result)
	}

	log.Printf("📊 [Parser] Financial results summary: parsed %d, skipped %d", len(results), skippedCount)

	if len(results) == 0 {
		return fmt.Errorf("no valid financial results found in file")
	}

	// 🔴 Сохраняем ВСЕ записи
	for i, result := range results {
		if err := s.finResultRepo.Create(ctx, result); err != nil {
			log.Printf("❌ [Parser] Failed to save result %d: %v", i+1, err)
			return fmt.Errorf("failed to save financial result #%d: %w", i+1, err)
		}
		log.Printf("💾 [Parser] Saved financial result %d: date=%s", i+1, result.ReportDate.Format("02.01.2006"))
	}

	return nil
}
// parseCreditAgreements парсит кредитные договоры из Excel
func (s *ReportService) parseCreditAgreements(ctx context.Context, xlFile *excelize.File, reportID int64) error {
	enterpriseID := int64(1)

	sheetName := xlFile.GetSheetName(0)
	rows, err := xlFile.GetRows(sheetName)
	if err != nil {
		return fmt.Errorf("failed to get rows from sheet: %w", err)
	}

	log.Printf("📋 [Parser] Found %d total rows in sheet '%s'", len(rows), sheetName)

	if len(rows) < 2 {
		return fmt.Errorf("credit agreements file is empty or has no data rows")
	}

	// 🔍 Лог заголовков
	if len(rows) > 0 {
		log.Printf("📋 [Parser] Header row: %v", rows[0])
	}

	agreements := []*models.CreditAgreement{}
	skippedCount := 0

	for i := 1; i < len(rows); i++ {
		row := rows[i]
		log.Printf("🔍 [Parser] Row %d: %v (len=%d)", i+1, row, len(row))
		
		// Пропускаем пустые строки
		if len(row) < 11 || isEmptyRow(row) {
			log.Printf("⚠️  [Parser] Row %d skipped: not enough columns (got %d, need 11)", i+1, len(row))
			skippedCount++
			continue
		}

		// 🔴 Используем GetCellValue для надёжного чтения
		// Парсим даты
		cellDate, _ := xlFile.GetCellValue(sheetName, fmt.Sprintf("B%d", i+1))
		agreementDate, err := parseExcelDate(cellDate)
		if err != nil || agreementDate.IsZero() {
			log.Printf("⚠️  [Parser] Row %d skipped: invalid agreement date '%s': %v", i+1, cellDate, err)
			skippedCount++
			continue
		}

		cellMaturity, _ := xlFile.GetCellValue(sheetName, fmt.Sprintf("J%d", i+1))
		maturityDate, err := parseExcelDate(cellMaturity)
		if err != nil || maturityDate.IsZero() {
			log.Printf("⚠️  [Parser] Row %d skipped: invalid maturity date '%s': %v", i+1, cellMaturity, err)
			skippedCount++
			continue
		}

		// 🔴 Парсим сумму через GetCellValue + parseFloat
		cellAmount, _ := xlFile.GetCellValue(sheetName, fmt.Sprintf("E%d", i+1))
		principalAmount, err := parseFloat(cellAmount)
		if err != nil || principalAmount <= 0 {
			log.Printf("⚠️  [Parser] Row %d skipped: invalid amount '%s': %v", i+1, cellAmount, err)
			skippedCount++
			continue
		}

		// 🔴 Парсим ставку
		cellRate, _ := xlFile.GetCellValue(sheetName, fmt.Sprintf("G%d", i+1))
		interestRate, err := parseFloat(cellRate)
		if err != nil {
			log.Printf("⚠️  [Parser] Row %d skipped: invalid rate '%s': %v", i+1, cellRate, err)
			skippedCount++
			continue
		}

		// ✅ ЧИТАЕМ СРОК ИЗ ФАЙЛА (колонка "Срок (мес.)" — индекс 10, колонка K)
		termMonths := 0
		cellTerm, _ := xlFile.GetCellValue(sheetName, fmt.Sprintf("K%d", i+1))
		if cellTerm != "" {
			if parsed, err := strconv.Atoi(strings.TrimSpace(cellTerm)); err == nil && parsed > 0 {
				termMonths = parsed
				log.Printf("📅 [Parser] Using term from file: %d months", termMonths)
			}
		}
		
		// Fallback: рассчитываем, если в файле нет или неверно
		if termMonths <= 0 {
			diffDays := maturityDate.Sub(agreementDate).Hours() / 24
			termMonths = int(diffDays / 30.44)
			log.Printf("📅 [Parser] Calculated term: %d months (from %d days)", termMonths, int(diffDays))
		}
		
		// ✅ Валидация: проверяем, что срок в допустимых пределах
		if termMonths < 1 || termMonths > 360 {
			log.Printf("⚠️  [Parser] Row %d skipped: term_months %d out of range [1, 360]", i+1, termMonths)
			skippedCount++
			continue
		}

		// 🔴 Читаем остальные поля
		cellNumber, _ := xlFile.GetCellValue(sheetName, fmt.Sprintf("A%d", i+1))
		cellCreditor, _ := xlFile.GetCellValue(sheetName, fmt.Sprintf("C%d", i+1))
		cellCountry, _ := xlFile.GetCellValue(sheetName, fmt.Sprintf("D%d", i+1))
		cellCurrency, _ := xlFile.GetCellValue(sheetName, fmt.Sprintf("F%d", i+1))
		cellRateType, _ := xlFile.GetCellValue(sheetName, fmt.Sprintf("H%d", i+1))
		cellCollateral, _ := xlFile.GetCellValue(sheetName, fmt.Sprintf("L%d", i+1))
		cellStatus, _ := xlFile.GetCellValue(sheetName, fmt.Sprintf("M%d", i+1))

		agreement := &models.CreditAgreement{
			EnterpriseID:    enterpriseID,
			ReportID:        &reportID,
			AgreementNumber: strings.TrimSpace(cellNumber),
			AgreementDate:   agreementDate,
			CreditorName:    strings.TrimSpace(cellCreditor),
			CreditorCountry: stringPtr(strings.TrimSpace(cellCountry)),
			PrincipalAmount: principalAmount,
			Currency:        strings.TrimSpace(cellCurrency),
			InterestRate:    interestRate,
			RateType:        strings.TrimSpace(cellRateType),
			StartDate:       agreementDate,
			MaturityDate:    maturityDate,
			TermMonths:      termMonths,
			CollateralType:  stringPtr(strings.TrimSpace(cellCollateral)),
			Status:          strings.TrimSpace(cellStatus),
		}

		agreements = append(agreements, agreement)
		log.Printf("✅ [Parser] Row %d parsed: contract=%s, amount=%.0f, term=%d months", 
			i+1, agreement.AgreementNumber, principalAmount, termMonths)
	}

	log.Printf("📊 [Parser] Summary: parsed %d agreements, skipped %d rows", len(agreements), skippedCount)

	if len(agreements) == 0 {
		return fmt.Errorf("no valid credit agreements found in file")
	}

	// Сохраняем в БД
	for i, agreement := range agreements {
		if err := s.creditRepo.Create(ctx, agreement); err != nil {
			log.Printf("❌ [Parser] Failed to save agreement %d: %v", i+1, err)
			return fmt.Errorf("failed to save credit agreement #%d: %w", i+1, err)
		}
		log.Printf("💾 [Parser] Saved agreement %d: %s", i+1, agreement.AgreementNumber)
	}

	return nil
}
// generateUniqueFileName генерирует уникальное имя файла
func (s *ReportService) generateUniqueFileName(reportType, originalName string) string {
	timestamp := time.Now().Format("20060102_150405")
	ext := filepath.Ext(originalName)
	
	// Очищаем имя файла от недопустимых символов
	baseName := strings.TrimSuffix(originalName, ext)
	baseName = strings.ReplaceAll(baseName, " ", "_")
	baseName = strings.ReplaceAll(baseName, ".", "_")
	
	return fmt.Sprintf("%s_%s_%s%s", reportType, timestamp, baseName, ext)
}

// saveFile сохраняет загруженный файл на диск
func (s *ReportService) saveFile(fileHeader *multipart.FileHeader, fileName string) (string, error) {
	// Открываем загруженный файл
	file, err := fileHeader.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer file.Close()

	// Формируем полный путь к файлу
	filePath := filepath.Join(s.uploadDir, fileName)

	// Создаём файл на диске
	out, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file %s: %w", filePath, err)
	}
	defer out.Close()

	// Копируем содержимое
	_, err = io.Copy(out, file)
	if err != nil {
		// Удаляем частично записанный файл
		_ = os.Remove(filePath)
		return "", fmt.Errorf("failed to write file %s: %w", filePath, err)
	}

	return filePath, nil
}

// GetReport получает отчёт по ID
func (s *ReportService) GetReport(ctx context.Context, id int64) (*models.Report, error) {
	return s.reportRepo.GetByID(ctx, id)
}

// GetAllReports получает список отчётов с фильтрацией
func (s *ReportService) GetAllReports(ctx context.Context, filter interfaces.ReportFilter) ([]*models.Report, error) {
	return s.reportRepo.GetAll(ctx, filter)
}

// DeleteReport удаляет отчёт по ID
func (s *ReportService) DeleteReport(ctx context.Context, id int64) error {
	// Получаем отчёт для получения пути к файлу
	report, err := s.reportRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Удаляем запись из БД
	if err := s.reportRepo.Delete(ctx, id); err != nil {
		return err
	}

	// Удаляем файл с диска
	if report.FilePath != "" {
		fullPath := report.FilePath
		if !filepath.IsAbs(fullPath) {
			fullPath = filepath.Join(s.uploadDir, fullPath)
		}
		if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
			// Логируем ошибку, но не прерываем удаление из БД
		}
	}

	return nil
}

// Вспомогательные функции
func parseFloat(s string) (float64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, nil
	}
	return strconv.ParseFloat(s,64)
}


// excelDateToTime конвертирует серийный номер даты Excel в time.Time
// Excel хранит даты как количество дней с 30.12.1899\
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

// parseExcelDate пытается распарсить дату из Excel (множество форматов)
func parseExcelDate(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, nil
	}

	// Список форматов для попытки парсинга (в порядке приоритета)
	dateFormats := []string{
		"02.01.2006",           // ДД.ММ.ГГГГ (русский)
		"02/01/2006",           // ДД/ММ/ГГГГ
		"01/02/2006",           // ММ/ДД/ГГГГ (американский) ⭐ НУЖЕН
		"01/02/2006 15:04:05",  // ММ/ДД/ГГГГ ЧЧ:ММ:СС ⭐ НУЖЕН
		"02.01.2006 15:04:05",  // ДД.ММ.ГГГГ ЧЧ:ММ:СС
		"2006-01-02",           // ГГГГ-ММ-ДД (ISO)
		"2006-01-02 15:04:05",  // ГГГГ-ММ-ДД ЧЧ:ММ:СС
		"02-Jan-2006",          // ДД-Ммм-ГГГГ
		"Jan 2, 2006",          // Ммм Д, ГГГГ
	}

	// Пробуем каждый формат
	for _, format := range dateFormats {
		if t, err := time.Parse(format, value); err == nil {
			return t, nil
		}
	}

	// Если не вышло — пробуем как серийный номер Excel
	if serial, err := strconv.ParseFloat(value, 64); err == nil && serial > 0 {
		return excelDateToTime(serial)
	}

	// Последняя попытка: убрать время и попробовать снова
	// Например: "01/12/2024 0:00:00" → "01/12/2024"
	if idx := strings.Index(value, " "); idx > 0 {
		datePart := strings.TrimSpace(value[:idx])
		for _, format := range []string{"02.01.2006", "02/01/2006", "01/02/2006", "2006-01-02"} {
			if t, err := time.Parse(format, datePart); err == nil {
				return t, nil
			}
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", value)
}
// 🔴 Вспомогательная функция для получения float из ячейки
func getFloatFromCell(xlFile *excelize.File, sheet string, row, col int) float64 {
	cellName := fmt.Sprintf("%c%d", 'A'+col-1, row)
	value, err := xlFile.GetCellValue(sheet, cellName)
	if err != nil || value == "" {
		return 0
	}
	
	// Очищаем значение (пробелы, пробелы-неразрывные, etc.)
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, " ", "")
	value = strings.ReplaceAll(value, "\u00A0", "") // Неразрывный пробел
	
	// Пробуем распарсить
	result, err := parseFloat(value)
	if err != nil {
		log.Printf("⚠️  [Parser] Failed to parse cell %s='%s': %v", cellName, value, err)
		return 0
	}
	
	return result
}

// 🔴 Проверка на пустую строку
func isEmptyRow(row []string) bool {
	for _, cell := range row {
		if strings.TrimSpace(cell) != "" {
			return false
		}
	}
	return true
}
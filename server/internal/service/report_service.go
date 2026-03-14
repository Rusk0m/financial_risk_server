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
		priceUSDPerT, err := parseFloat(strings.TrimSpace(row[4]))
		if err != nil || priceUSDPerT <= 0 {
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
			PriceUSDPerT:    priceUSDPerT,
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

	// Парсим данные из второй строки (индекс 1)
	// Ожидаемая структура:
	// 0:Дата | 1:ДС BYN | 2:ДС USD | 3:КФВ | 4:Дебиторка | 5:НДС получ | 6:Запасы | 7:Сырьё | 8:НЗП | 9:ГП | 10:ОС | 11:НМА | 12:Кредиторка | 13:НДС упл | 14:ЗП | 15:Кратк.кредиты | 16:Долг.кредиты | 17:УК | 18:НП

	// ✅ Парсим дату отчёта
	reportDate, err := parseExcelDate(rows[1][0])
	if err != nil || reportDate.IsZero() {
		log.Printf("⚠️  [Parser] Invalid report date '%s', using current date", rows[1][0])
		reportDate = time.Now()
	}
	log.Printf("📅 [Parser] Report date: %s", reportDate.Format("02.01.2006"))

	// Парсим числовые поля с логированием ошибок
	cashBYN, _ := parseFloat(strings.TrimSpace(rows[1][1]))
	cashUSD, _ := parseFloat(strings.TrimSpace(rows[1][2]))
	shortTermInvestments, _ := parseFloat(strings.TrimSpace(rows[1][3]))
	accountsReceivable, _ := parseFloat(strings.TrimSpace(rows[1][4]))
	vatReceivable, _ := parseFloat(strings.TrimSpace(rows[1][5]))
	inventories, _ := parseFloat(strings.TrimSpace(rows[1][6]))
	rawMaterials, _ := parseFloat(strings.TrimSpace(rows[1][7]))
	workInProgress, _ := parseFloat(strings.TrimSpace(rows[1][8]))
	finishedGoods, _ := parseFloat(strings.TrimSpace(rows[1][9]))
	propertyPlantEquipment, _ := parseFloat(strings.TrimSpace(rows[1][10]))
	intangibleAssets, _ := parseFloat(strings.TrimSpace(rows[1][11]))
	accountsPayable, _ := parseFloat(strings.TrimSpace(rows[1][12]))
	vatPayable, _ := parseFloat(strings.TrimSpace(rows[1][13]))
	payrollPayable, _ := parseFloat(strings.TrimSpace(rows[1][14]))
	shortTermDebt, _ := parseFloat(strings.TrimSpace(rows[1][15]))
	longTermDebt, _ := parseFloat(strings.TrimSpace(rows[1][16]))
	authorizedCapital, _ := parseFloat(strings.TrimSpace(rows[1][17]))
	retainedEarnings, _ := parseFloat(strings.TrimSpace(rows[1][18]))

	// 🔍 Лог ключевых показателей для отладки
	log.Printf("💰 [Parser] Key figures: Cash BYN=%.0f, Cash USD=%.0f, Receivables=%.0f, Payables=%.0f",
		cashBYN, cashUSD, accountsReceivable, accountsPayable)

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

	// Сохраняем в БД
	if err := s.balanceRepo.Create(ctx, balance); err != nil {
		log.Printf("❌ [Parser] Failed to save balance sheet: %v", err)
		return fmt.Errorf("failed to save balance sheet: %w", err)
	}

	log.Printf("✅ [Parser] Balance sheet saved successfully for report_id=%d", reportID)
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

	// 🔍 Лог заголовков
	if len(rows) > 0 {
		log.Printf("📋 [Parser] Header row: %v", rows[0])
	}

	// ✅ Парсим дату отчёта
	reportDate, err := parseExcelDate(rows[1][0])
	if err != nil || reportDate.IsZero() {
		log.Printf("⚠️  [Parser] Invalid report date '%s', using current date", rows[1][0])
		reportDate = time.Now()
	}

	// ✅ Парсим период
	periodStart, err := parseExcelDate(rows[1][1])
	if err != nil || periodStart.IsZero() {
		log.Printf("⚠️  [Parser] Invalid period_start '%s', using report_date - 3 months", rows[1][1])
		periodStart = reportDate.AddDate(0, -3, 0)
	}

	periodEnd, err := parseExcelDate(rows[1][2])
	if err != nil || periodEnd.IsZero() {
		log.Printf("⚠️  [Parser] Invalid period_end '%s', using report_date", rows[1][2])
		periodEnd = reportDate
	}

	log.Printf("📅 [Parser] Period: %s — %s", periodStart.Format("02.01.2006"), periodEnd.Format("02.01.2006"))

	// Парсим финансовые показатели с логированием
	revenueSales, _ := parseFloat(strings.TrimSpace(rows[1][3]))
	revenueExport, _ := parseFloat(strings.TrimSpace(rows[1][4]))
	revenueDomestic, _ := parseFloat(strings.TrimSpace(rows[1][5]))
	revenueOther, _ := parseFloat(strings.TrimSpace(rows[1][6]))
	revenueTotal, _ := parseFloat(strings.TrimSpace(rows[1][7]))

	costOfSales, _ := parseFloat(strings.TrimSpace(rows[1][8]))
	costRawMaterials, _ := parseFloat(strings.TrimSpace(rows[1][9]))
	costEnergy, _ := parseFloat(strings.TrimSpace(rows[1][10]))
	costLabor, _ := parseFloat(strings.TrimSpace(rows[1][11]))
	costDepreciation, _ := parseFloat(strings.TrimSpace(rows[1][12]))
	costOther, _ := parseFloat(strings.TrimSpace(rows[1][13]))
	costTotal, _ := parseFloat(strings.TrimSpace(rows[1][14]))

	commercialExpenses, _ := parseFloat(strings.TrimSpace(rows[1][15]))
	administrativeExpenses, _ := parseFloat(strings.TrimSpace(rows[1][16]))
	otherExpenses, _ := parseFloat(strings.TrimSpace(rows[1][17]))

	grossProfit, _ := parseFloat(strings.TrimSpace(rows[1][18]))
	operatingProfit, _ := parseFloat(strings.TrimSpace(rows[1][19]))
	profitBeforeTax, _ := parseFloat(strings.TrimSpace(rows[1][20]))
	taxExpense, _ := parseFloat(strings.TrimSpace(rows[1][21]))
	netProfit, _ := parseFloat(strings.TrimSpace(rows[1][22]))

	ebitda, _ := parseFloat(strings.TrimSpace(rows[1][23]))
	operatingMargin, _ := parseFloat(strings.TrimSpace(rows[1][24]))
	netMargin, _ := parseFloat(strings.TrimSpace(rows[1][25]))

	// 🔍 Лог ключевых показателей
	log.Printf("💰 [Parser] Revenue: total=%.0f, export=%.0f; Costs: total=%.0f; Net profit: %.0f",
		revenueTotal, revenueExport, costTotal, netProfit)

	// Валидация: прибыль не должна быть отрицательной при положительной выручке (опционально)
	// if revenueTotal > 0 && netProfit < -revenueTotal*0.5 {
	// 	log.Printf("⚠️  [Parser] Suspicious net profit: %.0f (revenue: %.0f)", netProfit, revenueTotal)
	// }

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

	// Сохраняем в БД
	if err := s.finResultRepo.Create(ctx, result); err != nil {
		log.Printf("❌ [Parser] Failed to save financial results: %v", err)
		return fmt.Errorf("failed to save financial results: %w", err)
	}

	log.Printf("✅ [Parser] Financial results saved successfully for report_id=%d", reportID)
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

	agreements := []*models.CreditAgreement{}
	skippedCount := 0

	for i := 1; i < len(rows); i++ {
		row := rows[i]
		log.Printf("🔍 [Parser] Row %d: %v (len=%d)", i+1, row, len(row))
		
		if len(row) < 11 { // Теперь нужно 11 колонок (индекс 10 = срок в месяцах)
			log.Printf("⚠️  [Parser] Row %d skipped: not enough columns (got %d, need 11)", i+1, len(row))
			skippedCount++
			continue
		}

		// Парсим даты
		agreementDate, err := parseExcelDate(row[1])
		if err != nil || agreementDate.IsZero() {
			log.Printf("⚠️  [Parser] Row %d skipped: invalid agreement date '%s': %v", i+1, row[1], err)
			skippedCount++
			continue
		}

		maturityDate, err := parseExcelDate(row[9]) // Индекс 9 = Дата погашения
		if err != nil || maturityDate.IsZero() {
			log.Printf("⚠️  [Parser] Row %d skipped: invalid maturity date '%s': %v", i+1, row[9], err)
			skippedCount++
			continue
		}

		// ✅ Парсим сумму и ставку
		principalAmount, err := parseFloat(strings.TrimSpace(row[4]))
		if err != nil {
			log.Printf("⚠️  [Parser] Row %d skipped: invalid amount '%s': %v", i+1, row[4], err)
			skippedCount++
			continue
		}

		interestRate, err := parseFloat(strings.TrimSpace(row[6]))
		if err != nil {
			log.Printf("⚠️  [Parser] Row %d skipped: invalid rate '%s': %v", i+1, row[6], err)
			skippedCount++
			continue
		}

		// ✅ ЧИТАЕМ СРОК ИЗ ФАЙЛА (колонка "Срок (мес.)" — индекс 10)
		termMonths := 0
		if row[10] != "" {
			if parsed, err := strconv.Atoi(strings.TrimSpace(row[10])); err == nil && parsed > 0 {
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
		// (предполагаем, что ограничение в БД: 1 <= term_months <= 360)
		if termMonths < 1 || termMonths > 360 {
			log.Printf("⚠️  [Parser] Row %d skipped: term_months %d out of range [1, 360]", i+1, termMonths)
			skippedCount++
			continue
		}

		agreement := &models.CreditAgreement{
			EnterpriseID:    enterpriseID,
			ReportID:        &reportID,
			AgreementNumber: strings.TrimSpace(row[0]),
			AgreementDate:   agreementDate,
			CreditorName:    strings.TrimSpace(row[2]),
			CreditorCountry: stringPtr(strings.TrimSpace(row[3])),
			PrincipalAmount: principalAmount,
			Currency:        strings.TrimSpace(row[5]),
			InterestRate:    interestRate,
			RateType:        strings.TrimSpace(row[7]),
			StartDate:       agreementDate, // или row[8] если нужна отдельная дата начала
			MaturityDate:    maturityDate,
			TermMonths:      termMonths, // ✅ Используем прочитанное/рассчитанное значение
			CollateralType:  stringPtr(strings.TrimSpace(row[11])), // Индекс 11 = Тип обеспечения
			Status:          strings.TrimSpace(row[12]),            // Индекс 12 = Статус
		}

		agreements = append(agreements, agreement)
		log.Printf("✅ [Parser] Row %d parsed: contract=%s, term=%d months", 
			i+1, agreement.AgreementNumber, termMonths)
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

func float64Ptr(v float64) *float64 {
	return &v
}

// internal/service/report_service.go — добавьте в конец файла

// excelDateToTime конвертирует серийный номер даты Excel в time.Time
// Excel хранит даты как количество дней с 30.12.1899
// См: https://docs.microsoft.com/en-us/office/troubleshoot/excel/1900-and-1904-date-system
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
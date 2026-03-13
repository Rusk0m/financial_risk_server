package service

import (
	"context"
	"fmt"
	"io"
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
	// 1. Валидация файла
	if file == nil {
		return nil, fmt.Errorf("file is required")
	}

	if file.Size == 0 {
		return nil, fmt.Errorf("file is empty")
	}

	if file.Size > 10*1024*1024 { // 10 МБ максимум
		return nil, fmt.Errorf("file size exceeds 10 MB limit")
	}

	// 2. Генерация уникального имени файла
	uniqueFileName := s.generateUniqueFileName(req.ReportType, file.Filename)

	// 3. Сохранение файла на диск
	filePath, err := s.saveFile(file, uniqueFileName)
	if err != nil {
		return nil, fmt.Errorf("failed to save file: %w", err)
	}

	// 4. Создание записи в БД
	report := &models.Report{
		EnterpriseID:    req.EnterpriseID,
		ReportType:      req.ReportType,
		FileName:        uniqueFileName,
		OriginalName:    file.Filename,
		FilePath:        filePath,
		FileSizeBytes:   file.Size,
		UploadDate:      time.Now(),
		ProcessingStatus: "pending",
		UploadedBy:      req.UploadedBy,
		Description:     req.Description,
	}

	if req.PeriodStart != "" {
		t, err := time.Parse("2006-01-02", req.PeriodStart)
		if err == nil {
			report.PeriodStart = &t
		}
	}
	if req.PeriodEnd != "" {
		t, err := time.Parse("2006-01-02", req.PeriodEnd)
		if err == nil {
			report.PeriodEnd = &t
		}
	}

	// Сохраняем отчёт в БД
	if err := s.reportRepo.Create(ctx, report); err != nil {
		// Если не удалось сохранить в БД, удаляем файл
		_ = os.Remove(filePath)
		return nil, fmt.Errorf("failed to create report record: %w", err)
	}

	// 5. Асинхронная обработка файла (парсинг)
	// В реальном приложении здесь был бы запуск горутины или задачи в очереди
	// Для диплома делаем синхронную обработку
	go func() {
		ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := s.processReportFile(ctxWithTimeout, report); err != nil {
			// Обновляем статус на "ошибка"
			errorMsg := err.Error()
			_ = s.reportRepo.UpdateProcessingStatus(ctxWithTimeout, report.ID, "error", &errorMsg)
		} else {
			// Обновляем статус на "обработан"
			_ = s.reportRepo.UpdateProcessingStatus(ctxWithTimeout, report.ID, "processed", nil)
		}
	}()

	return report, nil
}

// processReportFile парсит файл отчёта и сохраняет данные в соответствующие таблицы
func (s *ReportService) processReportFile(ctx context.Context, report *models.Report) error {
	// Открываем файл
	file, err := os.Open(report.FilePath)
	if err != nil {
		return fmt.Errorf("failed to open report file: %w", err)
	}
	defer file.Close()

	// Определяем тип файла по расширению
	ext := strings.ToLower(filepath.Ext(report.FileName))
	if ext != ".xlsx" && ext != ".xls" {
		return fmt.Errorf("unsupported file format: %s", ext)
	}

	// Открываем Excel файл
	xlFile, err := excelize.OpenFile(report.FilePath)
	if err != nil {
		return fmt.Errorf("failed to open Excel file: %w", err)
	}
	defer xlFile.Close()

	// Парсим в зависимости от типа отчёта
	switch report.ReportType {
	case "export_contracts":
		return s.parseExportContracts(ctx, xlFile, report)
	case "balance_sheet":
		return s.parseBalanceSheet(ctx, xlFile, report)
	case "financial_results":
		return s.parseFinancialResults(ctx, xlFile, report)
	case "credit_agreements":
		return s.parseCreditAgreements(ctx, xlFile, report)
	default:
		return fmt.Errorf("unsupported report type: %s", report.ReportType)
	}
}

// parseExportContracts парсит экспортные контракты из Excel
func (s *ReportService) parseExportContracts(ctx context.Context, xlFile *excelize.File, report *models.Report) error {
	// Получаем первую таблицу (лист)
	sheetName := xlFile.GetSheetName(0)
	rows, err := xlFile.GetRows(sheetName)
	if err != nil {
		return fmt.Errorf("failed to get rows from sheet: %w", err)
	}

	// Пропускаем заголовок (первая строка)
	if len(rows) < 2 {
		return fmt.Errorf("export contracts file is empty or has no data rows")
	}

	// Парсим каждую строку (начиная со второй)
	contracts := []*models.ExportContract{}
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		if len(row) < 7 { // Минимум 7 колонок для контракта
			continue
		}

		// Парсим дату контракта
		contractDate, err := time.Parse("02.01.2006", row[1])
		if err != nil {
			continue // Пропускаем строку с некорректной датой
		}

		// Парсим объём
		volumeT, err := parseFloat(row[3])
		if err != nil {
			continue
		}

		// Парсим цену
		priceUSDPerT, err := parseFloat(row[4])
		if err != nil {
			continue
		}

		// Парсим курс
		exchangeRate, err := parseFloat(row[7])
		if err != nil {
			exchangeRate = 3.25 // Курс по умолчанию
		}

		contract := &models.ExportContract{
			EnterpriseID:    report.EnterpriseID,
			ReportID:        &report.ID,
			ContractNumber:  row[0],
			ContractDate:    contractDate,
			Country:         row[2],
			VolumeT:         volumeT,
			PriceUSDPerT:    priceUSDPerT,
			Currency:        "USD",
			PaymentTermDays: 60,
			ShipmentDate:    contractDate.AddDate(0, 0, 30), // Отгрузка через 30 дней
			PaymentStatus:   "pending",
			ExchangeRate:    exchangeRate,
		}

		contracts = append(contracts, contract)
	}

	// Сохраняем контракты в БД
	for _, contract := range contracts {
		if err := s.contractRepo.Create(ctx, contract); err != nil {
			return fmt.Errorf("failed to save contract: %w", err)
		}
	}

	return nil
}

// parseBalanceSheet парсит финансовый баланс из Excel и сохраняет в БД
func (s *ReportService) parseBalanceSheet(ctx context.Context, xlFile *excelize.File, report *models.Report) error {
	// Получаем первый лист
	sheetName := xlFile.GetSheetName(0)
	rows, err := xlFile.GetRows(sheetName)
	if err != nil {
		return fmt.Errorf("failed to get rows from sheet: %w", err)
	}

	// Пропускаем заголовок (первая строка)
	if len(rows) < 2 {
		return fmt.Errorf("balance sheet file is empty or has no data rows")
	}

	// Парсим данные из второй строки (индекс 1)
	// Формат ожидаемого файла:
	// Строка 1: Заголовки
	// Строка 2: Данные баланса
	
	// Пример структуры файла:
	// Дата отчёта | ДС в BYN | ДС в USD | КФВ | Дебиторка | НДС к получению | Запасы | Сырьё | НЗП | ГП | ОС | НМА | Кредиторка | НДС к уплате | ЗП | Краткосрочные кредиты | Долгосрочные кредиты | УК | НП
	
	// Парсим дату отчёта
	reportDateStr := rows[1][0]
	reportDate, err := time.Parse("02.01.2006", reportDateStr)
	if err != nil {
		reportDate = time.Now() // Используем текущую дату при ошибке
	}

	// Парсим числовые поля с обработкой ошибок
	cashBYN, _ := parseFloat(rows[1][1])
	cashUSD, _ := parseFloat(rows[1][2])
	shortTermInvestments, _ := parseFloat(rows[1][3])
	accountsReceivable, _ := parseFloat(rows[1][4])
	vatReceivable, _ := parseFloat(rows[1][5])
	inventories, _ := parseFloat(rows[1][6])
	rawMaterials, _ := parseFloat(rows[1][7])
	workInProgress, _ := parseFloat(rows[1][8])
	finishedGoods, _ := parseFloat(rows[1][9])
	propertyPlantEquipment, _ := parseFloat(rows[1][10])
	intangibleAssets, _ := parseFloat(rows[1][11])
	accountsPayable, _ := parseFloat(rows[1][12])
	vatPayable, _ := parseFloat(rows[1][13])
	payrollPayable, _ := parseFloat(rows[1][14])
	shortTermDebt, _ := parseFloat(rows[1][15])
	longTermDebt, _ := parseFloat(rows[1][16])
	authorizedCapital, _ := parseFloat(rows[1][17])
	retainedEarnings, _ := parseFloat(rows[1][18])

	// Создаём баланс
	balance := &models.BalanceSheet{
		EnterpriseID:        report.EnterpriseID,
		ReportID:            &report.ID,
		ReportDate:          reportDate,
		CashBYN:             cashBYN,
		CashUSD:             cashUSD,
		ShortTermInvestments: shortTermInvestments,
		AccountsReceivable:  accountsReceivable,
		VATReceivable:       vatReceivable,
		Inventories:         inventories,
		RawMaterials:        rawMaterials,
		WorkInProgress:      workInProgress,
		FinishedGoods:       finishedGoods,
		PropertyPlantEquipment: propertyPlantEquipment,
		IntangibleAssets:    intangibleAssets,
		AccountsPayable:     accountsPayable,
		VATPayable:          vatPayable,
		PayrollPayable:      payrollPayable,
		ShortTermDebt:       shortTermDebt,
		LongTermDebt:        longTermDebt,
		AuthorizedCapital:   authorizedCapital,
		RetainedEarnings:    retainedEarnings,
	}

	// Сохраняем в БД
	return s.balanceRepo.Create(ctx, balance)
}

// parseFinancialResults парсит отчёт о финансовых результатах из Excel
func (s *ReportService) parseFinancialResults(ctx context.Context, xlFile *excelize.File, report *models.Report) error {
	sheetName := xlFile.GetSheetName(0)
	rows, err := xlFile.GetRows(sheetName)
	if err != nil {
		return fmt.Errorf("failed to get rows from sheet: %w", err)
	}

	if len(rows) < 2 {
		return fmt.Errorf("financial results file is empty or has no data rows")
	}

	// Парсим дату отчёта
	reportDateStr := rows[1][0]
	reportDate, err := time.Parse("02.01.2006", reportDateStr)
	if err != nil {
		reportDate = time.Now()
	}

	// Парсим период
	periodStartStr := rows[1][1]
	periodEndStr := rows[1][2]
	
	periodStart, _ := time.Parse("02.01.2006", periodStartStr)
	periodEnd, _ := time.Parse("02.01.2006", periodEndStr)
	if periodStart.IsZero() {
		periodStart = reportDate.AddDate(0, -3, 0) // Квартал назад по умолчанию
	}
	if periodEnd.IsZero() {
		periodEnd = reportDate
	}

	// Парсим финансовые показатели
	revenueSales, _ := parseFloat(rows[1][3])
	revenueExport, _ := parseFloat(rows[1][4])
	revenueDomestic, _ := parseFloat(rows[1][5])
	revenueOther, _ := parseFloat(rows[1][6])
	revenueTotal, _ := parseFloat(rows[1][7])
	
	costOfSales, _ := parseFloat(rows[1][8])
	costRawMaterials, _ := parseFloat(rows[1][9])
	costEnergy, _ := parseFloat(rows[1][10])
	costLabor, _ := parseFloat(rows[1][11])
	costDepreciation, _ := parseFloat(rows[1][12])
	costOther, _ := parseFloat(rows[1][13])
	costTotal, _ := parseFloat(rows[1][14])
	
	commercialExpenses, _ := parseFloat(rows[1][15])
	administrativeExpenses, _ := parseFloat(rows[1][16])
	otherExpenses, _ := parseFloat(rows[1][17])
	
	grossProfit, _ := parseFloat(rows[1][18])
	operatingProfit, _ := parseFloat(rows[1][19])
	profitBeforeTax, _ := parseFloat(rows[1][20])
	taxExpense, _ := parseFloat(rows[1][21])
	netProfit, _ := parseFloat(rows[1][22])
	
	ebitda, _ := parseFloat(rows[1][23])
	operatingMargin, _ := parseFloat(rows[1][24])
	netMargin, _ := parseFloat(rows[1][25])

	// Создаём отчёт
	result := &models.FinancialResult{
		EnterpriseID:         report.EnterpriseID,
		ReportID:             &report.ID,
		ReportDate:           reportDate,
		PeriodStart:          periodStart,
		PeriodEnd:            periodEnd,
		RevenueSales:         revenueSales,
		RevenueExport:        revenueExport,
		RevenueDomestic:      revenueDomestic,
		RevenueOther:         revenueOther,
		RevenueTotal:         revenueTotal,
		CostOfSales:          costOfSales,
		CostRawMaterials:     costRawMaterials,
		CostEnergy:           costEnergy,
		CostLabor:            costLabor,
		CostDepreciation:     costDepreciation,
		CostOther:            costOther,
		CostTotal:            costTotal,
		CommercialExpenses:   commercialExpenses,
		AdministrativeExpenses: administrativeExpenses,
		OtherExpenses:        otherExpenses,
		GrossProfit:          grossProfit,
		OperatingProfit:      operatingProfit,
		ProfitBeforeTax:      profitBeforeTax,
		TaxExpense:           taxExpense,
		NetProfit:            netProfit,
		EBITDA:               &ebitda,
		OperatingMargin:      &operatingMargin,
		NetMargin:            &netMargin,
	}

	// Сохраняем в БД
	return s.finResultRepo.Create(ctx, result)
}

// parseCreditAgreements парсит кредитные договоры из Excel
func (s *ReportService) parseCreditAgreements(ctx context.Context, xlFile *excelize.File, report *models.Report) error {
	sheetName := xlFile.GetSheetName(0)
	rows, err := xlFile.GetRows(sheetName)
	if err != nil {
		return fmt.Errorf("failed to get rows from sheet: %w", err)
	}

	// Пропускаем заголовок (первая строка)
	if len(rows) < 2 {
		return fmt.Errorf("credit agreements file is empty or has no data rows")
	}

	// Парсим каждый договор (начиная со второй строки)
	agreements := []*models.CreditAgreement{}
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		if len(row) < 10 { // Минимум 10 колонок для договора
			continue
		}

		// Парсим дату заключения
		agreementDate, err := time.Parse("02.01.2006", row[1])
		if err != nil {
			continue // Пропускаем строку с некорректной датой
		}

		// Парсим сумму кредита
		principalAmount, err := parseFloat(row[4])
		if err != nil {
			continue
		}

		// Парсим процентную ставку
		interestRate, err := parseFloat(row[6])
		if err != nil {
			continue
		}

		// Парсим дату погашения
		maturityDate, err := time.Parse("02.01.2006", row[8])
		if err != nil {
			continue
		}

		// Рассчитываем срок в месяцах
		termMonths := int(maturityDate.Sub(agreementDate).Hours() / 24 / 30.44)

		agreement := &models.CreditAgreement{
			EnterpriseID:     report.EnterpriseID,
			ReportID:         &report.ID,
			AgreementNumber:  row[0],
			AgreementDate:    agreementDate,
			CreditorName:     row[2],
			CreditorCountry:  stringPtr(row[3]),
			PrincipalAmount:  principalAmount,
			Currency:         row[5],
			InterestRate:     interestRate,
			RateType:         row[7],
			StartDate:        agreementDate,
			MaturityDate:     maturityDate,
			TermMonths:       termMonths,
			CollateralType:   stringPtr(row[9]),
			Status:           "active",
		}

		agreements = append(agreements, agreement)
	}

	// Сохраняем все договоры в БД
	for _, agreement := range agreements {
		if err := s.creditRepo.Create(ctx, agreement); err != nil {
			return fmt.Errorf("failed to save credit agreement: %w", err)
		}
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
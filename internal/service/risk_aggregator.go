package service

import (
	"errors"
	"financial-risk-server/internal/domain/models"
	"financial-risk-server/internal/repository/interfaces"
	"time"
)

type riskAggregatorService struct {
	currencyRiskService  *currencyRiskService
	creditRiskService    *creditRiskService
	liquidityRiskService *liquidityRiskService
	marketRiskService    *marketRiskService
	interestRiskService  *interestRiskService
	riskCalcRepo         interfaces.RiskCalculationRepository
	config               *RiskCalculationServiceConfig
}

// NewRiskAggregatorService создаёт агрегирующий сервис расчёта рисков
func NewRiskAggregatorService(
	currencyRiskService *currencyRiskService,
	creditRiskService *creditRiskService,
	liquidityRiskService *liquidityRiskService,
	marketRiskService *marketRiskService,
	interestRiskService *interestRiskService,
	riskCalcRepo interfaces.RiskCalculationRepository,
	config *RiskCalculationServiceConfig,
) *riskAggregatorService {
	return &riskAggregatorService{
		currencyRiskService:  currencyRiskService,
		creditRiskService:    creditRiskService,
		liquidityRiskService: liquidityRiskService,
		marketRiskService:    marketRiskService,
		interestRiskService:  interestRiskService,
		riskCalcRepo:         riskCalcRepo,
		config:               config,
	}
}

// CalculateAllRisks рассчитывает все 5 типов рисков и формирует комплексный отчёт
func (s *riskAggregatorService) CalculateAllRisks(
	enterpriseID int64,
	horizonDays int,
	confidenceLevel float64,
	priceChangePercent float64,
	rateChangePercent float64,
) (*models.ComprehensiveRiskReport, error) {
	// Валидация
	if enterpriseID <= 0 {
		return nil, errors.New("enterprise_id должен быть больше 0")
	}

	// Используем значения по умолчанию, если не заданы
	if horizonDays <= 0 {
		horizonDays = s.config.DefaultHorizonDays
	}
	if confidenceLevel <= 0 || confidenceLevel > 1 {
		confidenceLevel = s.config.DefaultConfidenceLevel
	}
	if priceChangePercent == 0 {
		priceChangePercent = s.config.DefaultPriceChangePct
	}
	if rateChangePercent == 0 {
		rateChangePercent = s.config.DefaultRateChangePct
	}

	// 1. Рассчитываем все 5 типов рисков параллельно
	var currencyRisk, creditRisk, liquidityRisk, marketRisk, interestRisk *models.RiskResult
	var currencyErr, creditErr, liquidityErr, marketErr, interestErr error

	// Используем горутины для параллельного расчёта
	errChan := make(chan error, 5)

	go func() {
		currencyRisk, currencyErr = s.currencyRiskService.CalculateCurrencyRisk(
			enterpriseID,
			horizonDays,
			confidenceLevel,
		)
		errChan <- currencyErr
	}()

	go func() {
		creditRisk, creditErr = s.creditRiskService.CalculateCreditRisk(enterpriseID)
		errChan <- creditErr
	}()

	go func() {
		liquidityRisk, liquidityErr = s.liquidityRiskService.CalculateLiquidityRisk(enterpriseID)
		errChan <- liquidityErr
	}()

	go func() {
		marketRisk, marketErr = s.marketRiskService.CalculateMarketRisk(
			enterpriseID,
			priceChangePercent,
		)
		errChan <- marketErr
	}()

	go func() {
		interestRisk, interestErr = s.interestRiskService.CalculateInterestRisk(
			enterpriseID,
			rateChangePercent,
		)
		errChan <- interestErr
	}()

	// Ждём завершения всех расчётов
	for i := 0; i < 5; i++ {
		if err := <-errChan; err != nil {
			// Продолжаем ждать остальные, но запоминаем ошибку
			// В реальной системе можно использовать более сложную логику
		}
	}

	// Проверяем критические ошибки
	if currencyErr != nil && creditErr != nil && liquidityErr != nil &&
		marketErr != nil && interestErr != nil {
		return nil, errors.New("не удалось рассчитать ни один тип риска")
	}

	// 2. Рассчитываем сводные показатели
	totalRiskValue := s.calculateTotalRiskValue(
		currencyRisk,
		creditRisk,
		liquidityRisk,
		marketRisk,
		interestRisk,
	)

	// 3. Определяем тип риска с максимальным значением
	maxRiskType := s.findMaxRiskType(
		currencyRisk,
		creditRisk,
		liquidityRisk,
		marketRisk,
		interestRisk,
	)

	// 4. Определяем общий уровень риска
	overallRiskLevel := s.determineOverallRiskLevel(
		currencyRisk,
		creditRisk,
		liquidityRisk,
		marketRisk,
		interestRisk,
	)

	// 5. Формируем текстовое резюме
	summary := s.generateSummary(
		currencyRisk,
		creditRisk,
		liquidityRisk,
		marketRisk,
		interestRisk,
		overallRiskLevel,
	)

	// 6. Формируем приоритетные действия
	priorityActions := s.generatePriorityActions(
		currencyRisk,
		creditRisk,
		liquidityRisk,
		marketRisk,
		interestRisk,
	)

	// 7. Формируем предупреждения
	warnings := s.generateWarnings(
		currencyRisk,
		creditRisk,
		liquidityRisk,
		marketRisk,
		interestRisk,
	)

	// 8. Создаём комплексный отчёт
	report := &models.ComprehensiveRiskReport{
		EnterpriseID:     enterpriseID,
		ReportDate:       time.Now(),
		PeriodStart:      time.Now().AddDate(0, 0, -horizonDays),
		PeriodEnd:        time.Now(),
		CurrencyRisk:     currencyRisk,
		CreditRisk:       creditRisk,
		LiquidityRisk:    liquidityRisk,
		MarketRisk:       marketRisk,
		InterestRisk:     interestRisk,
		TotalRiskValue:   totalRiskValue,
		MaxRiskType:      maxRiskType,
		OverallRiskLevel: overallRiskLevel,
		Summary:          summary,
		PriorityActions:  priorityActions,
		Warnings:         warnings,
		CreatedAt:        time.Now(),
	}

	// 9. Сохраняем результаты расчётов в БД
	s.saveRiskResults(currencyRisk, creditRisk, liquidityRisk, marketRisk, interestRisk)

	return report, nil
}

// calculateTotalRiskValue рассчитывает суммарное значение риска
func (s *riskAggregatorService) calculateTotalRiskValue(
	currencyRisk, creditRisk, liquidityRisk, marketRisk, interestRisk *models.RiskResult,
) float64 {
	total := 0.0

	if currencyRisk != nil {
		total += currencyRisk.VaRValue
	}
	if creditRisk != nil {
		total += creditRisk.VaRValue
	}
	if liquidityRisk != nil {
		total += liquidityRisk.VaRValue
	}
	if marketRisk != nil {
		total += marketRisk.VaRValue
	}
	if interestRisk != nil {
		total += interestRisk.VaRValue
	}

	return total
}

// findMaxRiskType находит тип риска с максимальным значением VaR
func (s *riskAggregatorService) findMaxRiskType(
	currencyRisk, creditRisk, liquidityRisk, marketRisk, interestRisk *models.RiskResult,
) models.RiskType {
	maxRiskValue := 0.0
	var maxRiskType models.RiskType

	risks := []struct {
		riskType models.RiskType
		risk     *models.RiskResult
	}{
		{models.RiskTypeCurrency, currencyRisk},
		{models.RiskTypeCredit, creditRisk},
		{models.RiskTypeLiquidity, liquidityRisk},
		{models.RiskTypeMarket, marketRisk},
		{models.RiskTypeInterest, interestRisk},
	}

	for _, r := range risks {
		if r.risk != nil && r.risk.VaRValue > maxRiskValue {
			maxRiskValue = r.risk.VaRValue
			maxRiskType = r.riskType
		}
	}

	return maxRiskType
}

// determineOverallRiskLevel определяет общий уровень риска на основе всех типов
func (s *riskAggregatorService) determineOverallRiskLevel(
	currencyRisk, creditRisk, liquidityRisk, marketRisk, interestRisk *models.RiskResult,
) models.RiskLevel {
	// Считаем количество высоких и средних рисков
	highRiskCount := 0
	mediumRiskCount := 0

	risks := []*models.RiskResult{
		currencyRisk,
		creditRisk,
		liquidityRisk,
		marketRisk,
		interestRisk,
	}

	for _, risk := range risks {
		if risk != nil {
			switch risk.RiskLevel {
			case models.RiskLevelHigh:
				highRiskCount++
			case models.RiskLevelMedium:
				mediumRiskCount++
			}
		}
	}

	// Определяем общий уровень
	if highRiskCount >= 2 {
		return models.RiskLevelHigh
	}
	if highRiskCount == 1 || mediumRiskCount >= 3 {
		return models.RiskLevelMedium
	}
	return models.RiskLevelLow
}

// generateSummary формирует текстовое резюме отчёта
func (s *riskAggregatorService) generateSummary(
	currencyRisk, creditRisk, liquidityRisk, marketRisk, interestRisk *models.RiskResult,
	overallRiskLevel models.RiskLevel,
) string {
	// Считаем статистику
	highRiskCount := 0
	mediumRiskCount := 0
	lowRiskCount := 0

	risks := []*models.RiskResult{
		currencyRisk,
		creditRisk,
		liquidityRisk,
		marketRisk,
		interestRisk,
	}

	for _, risk := range risks {
		if risk != nil {
			switch risk.RiskLevel {
			case models.RiskLevelHigh:
				highRiskCount++
			case models.RiskLevelMedium:
				mediumRiskCount++
			case models.RiskLevelLow:
				lowRiskCount++
			}
		}
	}

	// Формируем резюме
	summary := ""

	switch overallRiskLevel {
	case models.RiskLevelHigh:
		summary = "КРИТИЧЕСКИЙ УРОВЕНЬ РИСКА: "
	case models.RiskLevelMedium:
		summary = "СРЕДНИЙ УРОВЕНЬ РИСКА: "
	case models.RiskLevelLow:
		summary = "НИЗКИЙ УРОВЕНЬ РИСКА: "
	}

	summary += "Из 5 типов финансовых рисков "

	if highRiskCount > 0 {
		summary += "критических: " + string(rune('0'+highRiskCount)) + ", "
	}
	if mediumRiskCount > 0 {
		summary += "средних: " + string(rune('0'+mediumRiskCount)) + ", "
	}
	summary += "низких: " + string(rune('0'+lowRiskCount)) + "."

	return summary
}

// generatePriorityActions формирует список приоритетных действий
func (s *riskAggregatorService) generatePriorityActions(
	currencyRisk, creditRisk, liquidityRisk, marketRisk, interestRisk *models.RiskResult,
) []string {
	actions := []string{}

	// Добавляем действия для высоких рисков
	risks := []*models.RiskResult{
		currencyRisk,
		creditRisk,
		liquidityRisk,
		marketRisk,
		interestRisk,
	}

	for _, risk := range risks {
		if risk != nil && risk.RiskLevel == models.RiskLevelHigh {
			if len(risk.Recommendations) > 0 {
				// Добавляем первые 2 рекомендации как приоритетные
				for i := 0; i < len(risk.Recommendations) && i < 2; i++ {
					actions = append(actions, "⚠️ "+risk.Recommendations[i])
				}
			}
		}
	}

	// Если нет высоких рисков, добавляем действия для средних
	if len(actions) == 0 {
		for _, risk := range risks {
			if risk != nil && risk.RiskLevel == models.RiskLevelMedium {
				if len(risk.Recommendations) > 0 {
					actions = append(actions, "⚠️ "+risk.Recommendations[0])
				}
			}
		}
	}

	// Если нет никаких действий, добавляем общие рекомендации
	if len(actions) == 0 {
		actions = append(actions, "✅ Все риски находятся на допустимом уровне")
		actions = append(actions, "Продолжайте регулярный мониторинг финансовых рисков")
	}

	return actions
}

// generateWarnings формирует список предупреждений
func (s *riskAggregatorService) generateWarnings(
	currencyRisk, creditRisk, liquidityRisk, marketRisk, interestRisk *models.RiskResult,
) []string {
	warnings := []string{}

	risks := []*models.RiskResult{
		currencyRisk,
		creditRisk,
		liquidityRisk,
		marketRisk,
		interestRisk,
	}

	for _, risk := range risks {
		if risk != nil {
			riskPercentage := risk.GetRiskPercentage()
			if riskPercentage > 15 {
				warnings = append(warnings,
					"⚠️ "+string(risk.RiskType)+" риск превышает 15% от экспозиции: "+
						string(rune('0'+int(riskPercentage)))+"%")
			}
			if risk.StressTestLoss > risk.ExposureAmount*0.3 {
				warnings = append(warnings,
					"⚠️ Стресс-сценарий для "+string(risk.RiskType)+
						" риска показывает потери >30% от экспозиции")
			}
		}
	}

	return warnings
}

// saveRiskResults сохраняет результаты расчётов в базу данных
func (s *riskAggregatorService) saveRiskResults(
	currencyRisk, creditRisk, liquidityRisk, marketRisk, interestRisk *models.RiskResult,
) {
	if s.riskCalcRepo == nil {
		return
	}

	risks := []*models.RiskResult{
		currencyRisk,
		creditRisk,
		liquidityRisk,
		marketRisk,
		interestRisk,
	}

	for _, risk := range risks {
		if risk != nil {
			// Игнорируем ошибку, так как это не критично для основного расчёта
			_ = s.riskCalcRepo.Create(risk)
		}
	}
}

// CalculateCurrencyRisk делегирует расчёт валютного риска соответствующему сервису
func (s *riskAggregatorService) CalculateCurrencyRisk(
	enterpriseID int64,
	horizonDays int,
	confidenceLevel float64,
) (*models.RiskResult, error) {
	return s.currencyRiskService.CalculateCurrencyRisk(enterpriseID, horizonDays, confidenceLevel)
}

// CalculateCreditRisk делегирует расчёт кредитного риска соответствующему сервису
func (s *riskAggregatorService) CalculateCreditRisk(
	enterpriseID int64,
) (*models.RiskResult, error) {
	return s.creditRiskService.CalculateCreditRisk(enterpriseID)
}

// CalculateLiquidityRisk делегирует расчёт риска ликвидности соответствующему сервису
func (s *riskAggregatorService) CalculateLiquidityRisk(
	enterpriseID int64,
) (*models.RiskResult, error) {
	return s.liquidityRiskService.CalculateLiquidityRisk(enterpriseID)
}

// CalculateMarketRisk делегирует расчёт фондового риска соответствующему сервису
func (s *riskAggregatorService) CalculateMarketRisk(
	enterpriseID int64,
	priceChangePercent float64,
) (*models.RiskResult, error) {
	return s.marketRiskService.CalculateMarketRisk(enterpriseID, priceChangePercent)
}

// CalculateInterestRisk делегирует расчёт процентного риска соответствующему сервису
func (s *riskAggregatorService) CalculateInterestRisk(
	enterpriseID int64,
	rateChangePercent float64,
) (*models.RiskResult, error) {
	return s.interestRiskService.CalculateInterestRisk(enterpriseID, rateChangePercent)
}

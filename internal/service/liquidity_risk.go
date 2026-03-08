package service

import (
	"errors"
	"financial-risk-server/internal/domain/models"
	"financial-risk-server/internal/repository/interfaces"
	"time"
)

type liquidityRiskService struct {
	balanceRepo interfaces.BalanceSheetRepository
	costRepo    interfaces.CostStructureRepository
	config      *RiskCalculationServiceConfig
}

// NewLiquidityRiskService создаёт сервис расчёта риска ликвидности
func NewLiquidityRiskService(
	balanceRepo interfaces.BalanceSheetRepository,
	costRepo interfaces.CostStructureRepository,
	config *RiskCalculationServiceConfig,
) *liquidityRiskService {
	return &liquidityRiskService{
		balanceRepo: balanceRepo,
		costRepo:    costRepo,
		config:      config,
	}
}

// CalculateLiquidityRisk рассчитывает риск ликвидности на основе финансовых коэффициентов
func (s *liquidityRiskService) CalculateLiquidityRisk(enterpriseID int64) (*models.RiskResult, error) {
	// Валидация
	if enterpriseID <= 0 {
		return nil, errors.New("enterprise_id должен быть больше 0")
	}

	// 1. Получаем последний баланс предприятия
	balance, err := s.balanceRepo.GetLatest(enterpriseID)
	if err != nil {
		return nil, errors.New("не удалось получить данные баланса: " + err.Error())
	}

	// 2. Рассчитываем ключевые коэффициенты ликвидности
	currentRatio := balance.GetCurrentRatio()
	quickRatio := balance.GetQuickRatio()

	// 3. Рассчитываем ликвидный буфер (в месяцах)
	// Для этого нам нужно знать среднемесячные затраты
	monthlyCosts, err := s.calculateMonthlyCosts(enterpriseID)
	if err != nil {
		// Используем оценочное значение, если нет данных
		monthlyCosts = balance.GetTotalCurrentAssets() / 6 // ~2 месяца
	}

	// Ликвидный буфер = Денежные средства / Среднемесячные затраты
	liquidBufferMonths := calculateLiquidBuffer(balance, monthlyCosts)

	// 4. Рассчитываем экспозицию (общие текущие обязательства)
	exposure := balance.GetTotalCurrentLiabilities()

	// 5. Рассчитываем потенциальные потери при стресс-сценарии
	// (например, отток 30% ликвидных активов)
	stressTestLoss := balance.GetTotalCurrentAssets() * 0.3

	// 6. Определяем уровень риска на основе коэффициентов
	riskLevel := determineLiquidityRiskLevel(currentRatio, quickRatio, liquidBufferMonths)

	// 7. Рассчитываем "условный VaR" для ликвидности
	// (это не настоящий VaR, а показатель потенциальных проблем)
	varValue := calculateLiquidityRiskValue(currentRatio, quickRatio, liquidBufferMonths, exposure)

	// 8. Генерируем рекомендации
	recommendations := generateLiquidityRiskRecommendations(
		currentRatio,
		quickRatio,
		liquidBufferMonths,
		riskLevel,
	)

	// 9. Формируем результат
	result := &models.RiskResult{
		EnterpriseID:    enterpriseID,
		RiskType:        models.RiskTypeLiquidity,
		CalculationDate: time.Now(),
		HorizonDays:     30, // Ликвидность оценивается на 30 дней
		ConfidenceLevel: 0.95,
		ExposureAmount:  exposure,
		VaRValue:        varValue,
		StressTestLoss:  stressTestLoss,
		RiskLevel:       riskLevel,
		Recommendations: recommendations,
		ScenarioType:    "base",
		CreatedAt:       time.Now(),
	}

	return result, nil
}

// calculateMonthlyCosts рассчитывает среднемесячные затраты предприятия
func (s *liquidityRiskService) calculateMonthlyCosts(enterpriseID int64) (float64, error) {
	// Получаем структуру затрат за последний период
	costs, err := s.costRepo.GetByEnterpriseID(enterpriseID)
	if err != nil {
		return 0, err
	}

	if len(costs) == 0 {
		return 0, errors.New("нет данных о затратах")
	}

	// Рассчитываем общие годовые затраты
	totalAnnualCosts := 0.0
	for _, cost := range costs {
		// Предполагаем, что в CostStructure хранится годовой объём производства
		totalAnnualCosts += cost.CostPerT * 8500000 // 8.5 млн т для Беларуськалий
	}

	// Среднемесячные затраты
	return totalAnnualCosts / 12, nil
}

// calculateLiquidBuffer рассчитывает ликвидный буфер в месяцах
func calculateLiquidBuffer(balance *models.BalanceSheet, monthlyCosts float64) float64 {
	if monthlyCosts <= 0 {
		return 0
	}

	// Ликвидные активы = ДС в BYN + ДС в USD (переведённые в BYN)
	// Для упрощения считаем, что все ДС уже в одной валюте
	liquidAssets := balance.CashBYN + balance.CashUSD

	return liquidAssets / monthlyCosts
}

// determineLiquidityRiskLevel определяет уровень риска ликвидности
func determineLiquidityRiskLevel(
	currentRatio, quickRatio, liquidBufferMonths float64,
) models.RiskLevel {
	// Критерии для оценки ликвидности:
	// - Текущая ликвидность (current ratio): норма >= 1.5
	// - Быстрая ликвидность (quick ratio): норма >= 1.0
	// - Ликвидный буфер: норма >= 3 месяца

	currentRatioOK := currentRatio >= 1.5
	quickRatioOK := quickRatio >= 1.0
	bufferOK := liquidBufferMonths >= 3.0

	// Считаем количество нарушенных критериев
	violations := 0
	if !currentRatioOK {
		violations++
	}
	if !quickRatioOK {
		violations++
	}
	if !bufferOK {
		violations++
	}

	if violations >= 2 || liquidBufferMonths < 1.0 {
		return models.RiskLevelHigh
	}
	if violations == 1 || liquidBufferMonths < 2.0 {
		return models.RiskLevelMedium
	}
	return models.RiskLevelLow
}

// calculateLiquidityRiskValue рассчитывает условный "VaR" для ликвидности
func calculateLiquidityRiskValue(
	currentRatio, quickRatio, liquidBufferMonths, exposure float64,
) float64 {
	if exposure <= 0 {
		return 0
	}

	// Определяем коэффициент риска на основе отклонения от нормы
	riskFactor := 0.0

	if currentRatio < 1.0 {
		riskFactor += 0.3
	} else if currentRatio < 1.5 {
		riskFactor += 0.15
	}

	if quickRatio < 0.8 {
		riskFactor += 0.25
	} else if quickRatio < 1.0 {
		riskFactor += 0.1
	}

	if liquidBufferMonths < 1.0 {
		riskFactor += 0.3
	} else if liquidBufferMonths < 2.0 {
		riskFactor += 0.15
	} else if liquidBufferMonths < 3.0 {
		riskFactor += 0.05
	}

	return exposure * riskFactor
}

// generateLiquidityRiskRecommendations генерирует рекомендации по управлению ликвидностью
func generateLiquidityRiskRecommendations(
	currentRatio, quickRatio, liquidBufferMonths float64,
	riskLevel models.RiskLevel,
) []string {
	recommendations := []string{}

	switch riskLevel {
	case models.RiskLevelHigh:
		recommendations = append(recommendations,
			"🔴 КРИТИЧЕСКИЙ РИСК ЛИКВИДНОСТИ: Срочные меры по увеличению денежных средств",
			"Оптимизировать дебиторскую задолженность (ускорить сбор платежей)",
			"Пересмотреть условия оплаты с поставщиками (увеличить сроки)",
			"Рассмотреть краткосрочное кредитование для покрытия обязательств",
			"Сократить капитальные расходы на 30-50%",
		)

	case models.RiskLevelMedium:
		recommendations = append(recommendations,
			"🟡 СРЕДНИЙ РИСК ЛИКВИДНОСТИ: Оптимизация денежных потоков",
			"Усилить контроль за дебиторской задолженностью",
			"Переговорить с поставщиками о более выгодных условиях оплаты",
			"Создать резервный фонд на 2-3 месяца операционных расходов",
			"Мониторинг ликвидности ежедневно",
		)

	case models.RiskLevelLow:
		recommendations = append(recommendations,
			"🟢 НИЗКИЙ РИСК ЛИКВИДНОСТИ: Текущая политика достаточна",
			"Поддерживать ликвидный буфер на уровне 3+ месяцев",
			"Мониторинг ликвидности 1 раз в неделю",
			"Оптимизировать структуру оборотных активов",
		)
	}

	// Дополнительные рекомендации на основе конкретных показателей
	if currentRatio < 1.5 {
		recommendations = append(recommendations,
			"⚠️ Текущая ликвидность ниже нормы (< 1.5)",
			"Увеличить оборотные активы или сократить краткосрочные обязательства",
		)
	}

	if quickRatio < 1.0 {
		recommendations = append(recommendations,
			"⚠️ Быстрая ликвидность ниже нормы (< 1.0)",
			"Увеличить высоколиквидные активы (ДС, КФВ)",
		)
	}

	if liquidBufferMonths < 3.0 {
		recommendations = append(recommendations,
			"⚠️ Ликвидный буфер менее 3 месяцев",
			"Наращивать денежные средства до уровня 3+ месячных расходов",
		)
	}

	return recommendations
}

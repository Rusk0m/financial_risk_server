package service

import (
	"errors"
	"financial-risk-server/internal/domain/models"
	"financial-risk-server/internal/repository/interfaces"
	"time"
)

type currencyRiskService struct {
	exportRepo interfaces.ExportContractRepository
	marketRepo interfaces.MarketDataRepository
	config     *RiskCalculationServiceConfig
}

// NewCurrencyRiskService создаёт сервис расчёта валютного риска
func NewCurrencyRiskService(
	exportRepo interfaces.ExportContractRepository,
	marketRepo interfaces.MarketDataRepository,
	config *RiskCalculationServiceConfig,
) *currencyRiskService {
	return &currencyRiskService{
		exportRepo: exportRepo,
		marketRepo: marketRepo,
		config:     config,
	}
}

// CalculateCurrencyRisk рассчитывает валютный риск по методу VaR
func (s *currencyRiskService) CalculateCurrencyRisk(
	enterpriseID int64,
	horizonDays int,
	confidenceLevel float64,
) (*models.RiskResult, error) {

	// Валидация входных параметров
	if enterpriseID <= 0 {
		return nil, errors.New("enterprise_id должен быть больше 0")
	}

	if horizonDays <= 0 {
		horizonDays = s.config.DefaultHorizonDays
	}

	if confidenceLevel <= 0 || confidenceLevel > 1 {
		confidenceLevel = s.config.DefaultConfidenceLevel
	}

	// 1. Получаем незакрытые экспортные контракты
	pendingContracts, err := s.exportRepo.GetPendingContracts(enterpriseID)
	if err != nil {
		return nil, err
	}

	if len(pendingContracts) == 0 {
		return nil, errors.New("нет незакрытых экспортных контрактов для расчёта риска")
	}

	// 2. Рассчитываем общую экспозицию в валюте
	totalExposureUSD := calculateTotalExposure(pendingContracts)
	if totalExposureUSD <= 0 {
		return nil, errors.New("общая экспозиция равна 0")
	}

	// 3. Получаем волатильность курса BYN/USD
	volatility, err := s.marketRepo.GetVolatility("BYN/USD", 30)
	if err != nil {
		// Используем значение по умолчанию, если нет данных
		volatility = 0.09 // 9% волатильность (среднее значение для BYN/USD)
	}

	// 4. Рассчитываем VaR (Value at Risk)
	varValue := calculateVaR(
		totalExposureUSD,
		volatility,
		confidenceLevel,
		horizonDays,
	)

	// 5. Рассчитываем стресс-тест (+25% к курсу)
	stressTestLoss := calculateStressTestLoss(totalExposureUSD, 0.25)

	// 6. Определяем уровень риска
	riskLevel := determineRiskLevel(varValue, totalExposureUSD)

	// 7. Генерируем рекомендации
	recommendations := generateCurrencyRiskRecommendations(varValue, totalExposureUSD, riskLevel)

	// 8. Формируем результат
	result := &models.RiskResult{
		EnterpriseID:    enterpriseID,
		RiskType:        models.RiskTypeCurrency,
		CalculationDate: time.Now(),
		HorizonDays:     horizonDays,
		ConfidenceLevel: confidenceLevel,
		ExposureAmount:  totalExposureUSD,
		VaRValue:        varValue,
		StressTestLoss:  stressTestLoss,
		RiskLevel:       riskLevel,
		Recommendations: recommendations,
		ScenarioType:    "base",
		CreatedAt:       time.Now(),
	}

	return result, nil
}

// calculateTotalExposure рассчитывает общую экспозицию из списка контрактов
func calculateTotalExposure(contracts []*models.ExportContract) float64 {
	total := 0.0
	for _, contract := range contracts {
		total += contract.GetContractValue()
	}
	return total
}

// generateCurrencyRiskRecommendations генерирует рекомендации по управлению валютным риском
func generateCurrencyRiskRecommendations(
	varValue, exposure float64,
	riskLevel models.RiskLevel,
) []string {
	ratio := (varValue / exposure) * 100
	recommendations := []string{}

	switch riskLevel {
	case models.RiskLevelHigh:
		recommendations = append(recommendations,
			"🔴 КРИТИЧЕСКИЙ РИСК: Немедленно хеджировать 50-70% валютной экспозиции",
			"Сократить срок оплаты экспортных контрактов до 30-45 дней",
			"Диверсифицировать валюту расчётов (включить CNY, EUR до 30%)",
			"Рассмотреть возможность предоплаты от покупателей",
		)

	case models.RiskLevelMedium:
		recommendations = append(recommendations,
			"🟡 СРЕДНИЙ РИСК: Хеджировать 30-40% валютной экспозиции через форвардные контракты",
			"Сократить срок оплаты экспортных контрактов до 45-60 дней",
			"Диверсифицировать валюту расчётов (включить CNY, EUR до 20%)",
		)

	case models.RiskLevelLow:
		recommendations = append(recommendations,
			"🟢 НИЗКИЙ РИСК: Мониторинг валютных позиций 1 раз в неделю",
			"Хеджировать 10-20% экспозиции для стабилизации денежных потоков",
			"Поддерживать текущую структуру валютных расчётов",
		)
	}

	// Дополнительные рекомендации на основе конкретных значений
	if ratio > 10 {
		recommendations = append(recommendations,
			"Усилить контроль за курсовыми разницами в бухгалтерском учёте",
		)
	}

	if ratio > 20 {
		recommendations = append(recommendations,
			"Разработать план действий на случай резкой девальвации белорусского рубля",
			"Увеличить ликвидные резервы в иностранной валюте",
		)
	}

	return recommendations
}

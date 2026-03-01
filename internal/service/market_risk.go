package service

import (
	"errors"
	"finantial-risk-server/internal/domain/models"
	"finantial-risk-server/internal/repository/interfaces"
	"math"
	"time"
)

type marketRiskService struct {
	enterpriseRepo interfaces.EnterpriseRepository
	marketRepo     interfaces.MarketDataRepository
	exportRepo     interfaces.ExportContractRepository
	config         *RiskCalculationServiceConfig
}

// NewMarketRiskService создаёт сервис расчёта фондового риска
func NewMarketRiskService(
	enterpriseRepo interfaces.EnterpriseRepository,
	marketRepo interfaces.MarketDataRepository,
	exportRepo interfaces.ExportContractRepository,
	config *RiskCalculationServiceConfig,
) *marketRiskService {
	return &marketRiskService{
		enterpriseRepo: enterpriseRepo,
		marketRepo:     marketRepo,
		exportRepo:     exportRepo,
		config:         config,
	}
}

// CalculateMarketRisk рассчитывает фондовый риск на основе чувствительности к ценам на калий
func (s *marketRiskService) CalculateMarketRisk(
	enterpriseID int64,
	priceChangePercent float64,
) (*models.RiskResult, error) {
	// Валидация
	if enterpriseID <= 0 {
		return nil, errors.New("enterprise_id должен быть больше 0")
	}

	// Используем значение по умолчанию, если не задано
	if priceChangePercent == 0 {
		priceChangePercent = s.config.DefaultPriceChangePct
	}

	// 1. Получаем информацию о предприятии
	enterprise, err := s.enterpriseRepo.GetByID(enterpriseID)
	if err != nil {
		return nil, errors.New("не удалось получить данные предприятия: " + err.Error())
	}

	// 2. Получаем текущую цену калия на рынке
	marketData, err := s.marketRepo.GetLatestByCurrencyPair("POTASSIUM")
	if err != nil {
		// Используем значение по умолчанию
		marketData = &models.MarketData{
			PotassiumPriceUSD: 285.0, // $285/т - средняя цена 2025-2026
		}
	}

	currentPrice := marketData.PotassiumPriceUSD

	// 3. Рассчитываем годовую выручку при текущей цене
	// Выручка = Объём производства × Цена
	annualRevenueCurrent := enterprise.AnnualProductionT * currentPrice

	// 4. Рассчитываем выручку при изменении цены
	newPrice := currentPrice * (1 + priceChangePercent/100)
	annualRevenueNew := enterprise.AnnualProductionT * newPrice

	// 5. Рассчитываем изменение выручки
	revenueChange := annualRevenueNew - annualRevenueCurrent
	revenueChangeAbs := math.Abs(revenueChange)

	// 6. Рассчитываем экспозицию (текущая годовая выручка)
	exposure := annualRevenueCurrent

	// 7. Определяем уровень риска на основе изменения выручки
	riskLevel := determineMarketRiskLevel(priceChangePercent, revenueChangeAbs, exposure)

	// 8. Рассчитываем стресс-тест (падение цены на 30%)
	stressPrice := currentPrice * 0.7
	stressRevenue := enterprise.AnnualProductionT * stressPrice
	stressTestLoss := annualRevenueCurrent - stressRevenue

	// 9. Генерируем рекомендации
	recommendations := generateMarketRiskRecommendations(
		priceChangePercent,
		currentPrice,
		newPrice,
		revenueChange,
		riskLevel,
	)

	// 10. Формируем результат
	result := &models.RiskResult{
		EnterpriseID:    enterpriseID,
		RiskType:        models.RiskTypeMarket,
		CalculationDate: time.Now(),
		HorizonDays:     90, // Фондовый риск оценивается на 90 дней
		ConfidenceLevel: 0.95,
		ExposureAmount:  exposure,
		VaRValue:        revenueChangeAbs, // Потенциальное изменение выручки
		StressTestLoss:  stressTestLoss,
		RiskLevel:       riskLevel,
		Recommendations: recommendations,
		ScenarioType:    getScenarioType(priceChangePercent),
		CreatedAt:       time.Now(),
	}

	return result, nil
}

// determineMarketRiskLevel определяет уровень фондового риска
func determineMarketRiskLevel(
	priceChangePercent, revenueChange, exposure float64,
) models.RiskLevel {
	if exposure <= 0 {
		return models.RiskLevelLow
	}

	// Относительное изменение выручки
	revenueChangePercent := (revenueChange / exposure) * 100

	// Абсолютное падение цены
	absPriceChange := math.Abs(priceChangePercent)

	// Определяем уровень риска
	if revenueChangePercent > 25 || absPriceChange > 30 {
		return models.RiskLevelHigh
	}
	if revenueChangePercent > 10 || absPriceChange > 15 {
		return models.RiskLevelMedium
	}
	return models.RiskLevelLow
}

// getScenarioType определяет тип сценария на основе изменения цены
func getScenarioType(priceChangePercent float64) string {
	if priceChangePercent <= -25 {
		return "stress" // Стрессовый сценарий
	}
	if priceChangePercent <= -10 {
		return "pessimistic" // Пессимистичный сценарий
	}
	if priceChangePercent >= 10 {
		return "optimistic" // Оптимистичный сценарий
	}
	return "base" // Базовый сценарий
}

// generateMarketRiskRecommendations генерирует рекомендации по управлению фондово
// generateMarketRiskRecommendations генерирует рекомендации по управлению фондового риском
func generateMarketRiskRecommendations(
	priceChangePercent, currentPrice, newPrice, revenueChange float64,
	riskLevel models.RiskLevel,
) []string {
	recommendations := []string{}

	// Определяем направление изменения
	isPriceDecrease := priceChangePercent < 0
	absPriceChange := math.Abs(priceChangePercent)

	switch riskLevel {
	case models.RiskLevelHigh:
		if isPriceDecrease {
			recommendations = append(recommendations,
				"🔴 КРИТИЧЕСКИЙ РИСК: Резкое падение цен на калий",
				"Срочно оптимизировать себестоимость производства (цель: снижение на 15-20%)",
				"Диверсифицировать продуктовую линейку (другие виды удобрений)",
				"Рассмотреть возможность сокращения объёмов производства",
				"Усилить маркетинговые усилия для поиска новых рынков сбыта",
				"Хеджировать цены через фьючерсные контракты (если доступно)",
			)
		} else {
			recommendations = append(recommendations,
				"🟢 ЗНАЧИТЕЛЬНЫЙ РОСТ ЦЕН: Высокая прибыльность",
				"Увеличить объёмы производства до максимума",
				"Инвестировать в модернизацию оборудования",
				"Наращивать запасы готовой продукции для продажи по высоким ценам",
				"Рассмотреть возможность экспансии на новые рынки",
			)
		}

	case models.RiskLevelMedium:
		if isPriceDecrease {
			recommendations = append(recommendations,
				"🟡 СРЕДНИЙ РИСК: Умеренное падение цен на калий",
				"Оптимизировать себестоимость производства (цель: снижение на 5-10%)",
				"Усилить переговоры с покупателями о долгосрочных контрактах",
				"Диверсифицировать продуктовую линейку",
				"Мониторинг мировых цен ежедневно",
				"Рассмотреть возможность хеджирования цен",
			)
		} else {
			recommendations = append(recommendations,
				"📈 УМЕРЕННЫЙ РОСТ ЦЕН: Улучшение финансовых показателей",
				"Поддерживать текущие объёмы производства",
				"Оптимизировать логистику для увеличения маржинальности",
				"Инвестировать в НИОКР для улучшения качества продукции",
			)
		}

	case models.RiskLevelLow:
		recommendations = append(recommendations,
			"🟢 СТАБИЛЬНЫЙ РЫНОК: Текущая стратегия достаточна",
			"Поддерживать текущие объёмы производства и ценовую политику",
			"Мониторинг мировых цен 1 раз в неделю",
			"Планирование производства на основе прогнозов спроса",
		)
	}

	// Дополнительные рекомендации на основе конкретных значений
	if absPriceChange > 20 {
		recommendations = append(recommendations,
			"⚠️ Значительное изменение цены: >20%",
			"Провести детальный анализ причин изменения цен",
			"Скорректировать бизнес-план на следующий квартал",
		)
	}

	// Рекомендации по конкретным ценам
	if newPrice < 200 {
		recommendations = append(recommendations,
			"⚠️ КРИТИЧЕСКАЯ ЦЕНА: <$200/т",
			"Рассмотреть возможность временной остановки части мощностей",
			"Максимально сократить переменные затраты",
		)
	} else if newPrice < 250 {
		recommendations = append(recommendations,
			"⚠️ НИЗКАЯ ЦЕНА: <$250/т",
			"Оптимизировать затраты на производство",
			"Усилить переговоры с покупателями",
		)
	} else if newPrice > 350 {
		recommendations = append(recommendations,
			"📈 ВЫСОКАЯ ЦЕНА: >$350/т",
			"Максимально использовать производственные мощности",
			"Рассмотреть возможность расширения производства",
		)
	}

	// Рекомендации по изменению выручки
	revenueChangePercent := 0.0
	if currentPrice > 0 {
		revenueChangePercent = (math.Abs(revenueChange) / (currentPrice * 8500000)) * 100
	}

	if revenueChangePercent > 15 {
		recommendations = append(recommendations,
			"⚠️ Значительное изменение выручки: >15%",
			"Скорректировать финансовый план",
			"Пересмотреть бюджеты на следующий период",
		)
	}

	return recommendations
}

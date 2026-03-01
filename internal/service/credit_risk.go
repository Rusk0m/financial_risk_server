package service

import (
	"errors"
	"finantial-risk-server/internal/domain/models"
	"finantial-risk-server/internal/repository/interfaces"
	"time"
)

type creditRiskService struct {
	exportRepo interfaces.ExportContractRepository
	config     *RiskCalculationServiceConfig
}

// NewCreditRiskService создаёт сервис расчёта кредитного риска
func NewCreditRiskService(
	exportRepo interfaces.ExportContractRepository,
	config *RiskCalculationServiceConfig,
) *creditRiskService {
	return &creditRiskService{
		exportRepo: exportRepo,
		config:     config,
	}
}

// CalculateCreditRisk рассчитывает кредитный риск на основе анализа контрагентов
func (s *creditRiskService) CalculateCreditRisk(enterpriseID int64) (*models.RiskResult, error) {
	// Валидация
	if enterpriseID <= 0 {
		return nil, errors.New("enterprise_id должен быть больше 0")
	}

	// 1. Получаем все незакрытые контракты
	pendingContracts, err := s.exportRepo.GetPendingContracts(enterpriseID)
	if err != nil {
		return nil, err
	}

	if len(pendingContracts) == 0 {
		return nil, errors.New("нет незакрытых контрактов для расчёта кредитного риска")
	}

	// 2. Анализируем риски по странам
	countryRiskAnalysis := analyzeCountryRisk(pendingContracts)

	// 3. Рассчитываем общую экспозицию
	totalExposure := 0.0
	for _, contract := range pendingContracts {
		totalExposure += contract.GetContractValue()
	}

	// 4. Рассчитываем ожидаемые потери (Expected Loss)
	expectedLoss := calculateExpectedLoss(
		countryRiskAnalysis.AveragePD,  // Средняя вероятность дефолта
		countryRiskAnalysis.AverageLGD, // Средние потери при дефолте
		totalExposure,                  // Экспозиция
	)

	// 5. Рассчитываем потенциальные потери при стресс-сценарии
	// (дефолт крупнейшего контрагента)
	stressTestLoss := countryRiskAnalysis.MaxCountryExposure * countryRiskAnalysis.MaxCountryLGD

	// 6. Определяем уровень риска
	riskLevel := determineCreditRiskLevel(expectedLoss, totalExposure, countryRiskAnalysis)

	// 7. Генерируем рекомендации
	recommendations := generateCreditRiskRecommendations(
		expectedLoss,
		totalExposure,
		countryRiskAnalysis,
		riskLevel,
	)

	// 8. Формируем результат
	result := &models.RiskResult{
		EnterpriseID:    enterpriseID,
		RiskType:        models.RiskTypeCredit,
		CalculationDate: time.Now(),
		HorizonDays:     90, // Кредитный риск обычно оценивается на 90 дней
		ConfidenceLevel: 0.95,
		ExposureAmount:  totalExposure,
		VaRValue:        expectedLoss,
		StressTestLoss:  stressTestLoss,
		RiskLevel:       riskLevel,
		Recommendations: recommendations,
		ScenarioType:    "base",
		CreatedAt:       time.Now(),
	}

	return result, nil
}

// CountryRiskAnalysis результат анализа рисков по странам
type CountryRiskAnalysis struct {
	CountryExposures   map[string]float64 // Экспозиция по странам
	AveragePD          float64            // Средняя вероятность дефолта
	AverageLGD         float64            // Средние потери при дефолте
	MaxCountryExposure float64            // Максимальная экспозиция по одной стране
	MaxCountry         string             // Страна с максимальной экспозицией
	MaxCountryLGD      float64            // LGD для страны с макс. экспозицией
	ConcentrationRatio float64            // Коэффициент концентрации
}

// analyzeCountryRisk анализирует кредитные риски по странам контрагентов
func analyzeCountryRisk(contracts []*models.ExportContract) *CountryRiskAnalysis {
	countryExposures := make(map[string]float64)
	totalExposure := 0.0

	// Агрегируем экспозицию по странам
	for _, contract := range contracts {
		countryExposures[contract.Country] += contract.GetContractValue()
		totalExposure += contract.GetContractValue()
	}

	// Рассчитываем средние значения
	totalPD := 0.0
	totalLGD := 0.0
	maxExposure := 0.0
	maxCountry := ""

	for country, exposure := range countryExposures {
		pd := getCountryPD(country)   // Вероятность дефолта по стране
		lgd := getCountryLGD(country) // Потери при дефолте по стране

		// Взвешиваем по экспозиции
		weight := exposure / totalExposure
		totalPD += pd * weight
		totalLGD += lgd * weight

		if exposure > maxExposure {
			maxExposure = exposure
			maxCountry = country
		}
	}

	// Рассчитываем коэффициент концентрации (экспозиция крупнейшего контрагента / общая экспозиция)
	concentrationRatio := 0.0
	if totalExposure > 0 {
		concentrationRatio = maxExposure / totalExposure
	}

	return &CountryRiskAnalysis{
		CountryExposures:   countryExposures,
		AveragePD:          totalPD,
		AverageLGD:         totalLGD,
		MaxCountryExposure: maxExposure,
		MaxCountry:         maxCountry,
		MaxCountryLGD:      getCountryLGD(maxCountry),
		ConcentrationRatio: concentrationRatio,
	}
}

// getCountryPD возвращает вероятность дефолта по стране (на основе рейтингов)
func getCountryPD(country string) float64 {
	// Примерные значения на основе рейтингов (упрощённо)
	countryPD := map[string]float64{
		"Китай":     0.008, // Рейтинг A
		"Бразилия":  0.012, // Рейтинг BBB+
		"Индия":     0.015, // Рейтинг BBB
		"Индонезия": 0.018, // Рейтинг BBB-
		"Малайзия":  0.010, // Рейтинг A-
		"Вьетнам":   0.025, // Рейтинг BB+
		"Таиланд":   0.014, // Рейтинг A-
		"Бангладеш": 0.030, // Рейтинг BB
		"Пакистан":  0.040, // Рейтинг B+
		"Шри-Ланка": 0.050, // Рейтинг B
		"default":   0.020, // Значение по умолчанию
	}

	if pd, ok := countryPD[country]; ok {
		return pd
	}
	return countryPD["default"]
}

// getCountryLGD возвращает потери при дефолте по стране
func getCountryLGD(country string) float64 {
	// LGD зависит от качества правовой системы и возможности взыскания
	countryLGD := map[string]float64{
		"Китай":     0.45, // 45% потерь
		"Бразилия":  0.50, // 50% потерь
		"Индия":     0.55, // 55% потерь
		"Индонезия": 0.60, // 60% потерь
		"Малайзия":  0.48, // 48% потерь
		"Вьетнам":   0.65, // 65% потерь
		"Таиланд":   0.52, // 52% потерь
		"Бангладеш": 0.70, // 70% потерь
		"Пакистан":  0.75, // 75% потерь
		"Шри-Ланка": 0.80, // 80% потерь
		"default":   0.60, // Значение по умолчанию
	}

	if lgd, ok := countryLGD[country]; ok {
		return lgd
	}
	return countryLGD["default"]
}

// determineCreditRiskLevel определяет уровень кредитного риска
func determineCreditRiskLevel(
	expectedLoss, totalExposure float64,
	analysis *CountryRiskAnalysis,
) models.RiskLevel {
	// Основной критерий - отношение ожидаемых потерь к экспозиции
	lossRatio := 0.0
	if totalExposure > 0 {
		lossRatio = (expectedLoss / totalExposure) * 100
	}

	// Дополнительный критерий - концентрация риска
	concentrationRisk := analysis.ConcentrationRatio > 0.3 // >30% в одной стране

	if lossRatio > 8 || concentrationRisk {
		return models.RiskLevelHigh
	}
	if lossRatio > 3 {
		return models.RiskLevelMedium
	}
	return models.RiskLevelLow
}

// generateCreditRiskRecommendations генерирует рекомендации по управлению кредитным риском
func generateCreditRiskRecommendations(
	expectedLoss, totalExposure float64,
	analysis *CountryRiskAnalysis,
	riskLevel models.RiskLevel,
) []string {

	recommendations := []string{}

	switch riskLevel {
	case models.RiskLevelHigh:
		recommendations = append(recommendations,
			"🔴 КРИТИЧЕСКИЙ РИСК: Срочно диверсифицировать экспортные рынки",
			"Снизить долю крупнейшего контрагента до <30% от общей экспозиции",
			"Требовать предоплату или аккредитивы от контрагентов из стран с рейтингом < BBB",
			"Страховать экспортные поставки на сумму 70-100% от контракта",
		)

	case models.RiskLevelMedium:
		recommendations = append(recommendations,
			"🟡 СРЕДНИЙ РИСК: Диверсифицировать экспортные рынки в течение 6 месяцев",
			"Снизить долю крупнейшего контрагента до <40%",
			"Требовать предоплату 30-50% от контрагентов из стран с рейтингом < BBB",
			"Страховать экспортные поставки на сумму 50-70% от контракта",
		)

	case models.RiskLevelLow:
		recommendations = append(recommendations,
			"🟢 НИЗКИЙ РИСК: Поддерживать текущую структуру экспортных рынков",
			"Мониторинг кредитных рейтингов стран-импортёров 1 раз в квартал",
			"Страховать экспортные поставки на сумму 30-50% от контракта",
		)
	}

	// Дополнительные рекомендации на основе концентрации
	if analysis.ConcentrationRatio > 0.4 {
		recommendations = append(recommendations,
			"⚠️ Высокая концентрация риска: доля крупнейшего рынка превышает 40%",
			"Разработать план выхода на новые рынки (Африка, Ближний Восток, Латинская Америка)",
		)
	}

	// Рекомендации по конкретным странам
	if analysis.MaxCountryLGD > 0.6 {
		recommendations = append(recommendations,
			"⚠️ Высокие потери при дефолте в стране "+analysis.MaxCountry,
			"Рассмотреть возможность сокращения поставок в эту страну",
		)
	}

	return recommendations
}

package service

import (
	"errors"
	"financial-risk-server/internal/domain/models"
	"financial-risk-server/internal/repository/interfaces"
	"time"
)

type interestRiskService struct {
	balanceRepo interfaces.BalanceSheetRepository
	config      *RiskCalculationServiceConfig
}

// NewInterestRiskService создаёт сервис расчёта процентного риска
func NewInterestRiskService(
	balanceRepo interfaces.BalanceSheetRepository,
	config *RiskCalculationServiceConfig,
) *interestRiskService {
	return &interestRiskService{
		balanceRepo: balanceRepo,
		config:      config,
	}
}

// CalculateInterestRisk рассчитывает процентный риск на основе долговой нагрузки
func (s *interestRiskService) CalculateInterestRisk(
	enterpriseID int64,
	rateChangePercent float64,
) (*models.RiskResult, error) {
	// Валидация
	if enterpriseID <= 0 {
		return nil, errors.New("enterprise_id должен быть больше 0")
	}

	// Используем значение по умолчанию, если не задано
	if rateChangePercent == 0 {
		rateChangePercent = s.config.DefaultRateChangePct
	}

	// 1. Получаем последний баланс предприятия
	balance, err := s.balanceRepo.GetLatest(enterpriseID)
	if err != nil {
		return nil, errors.New("не удалось получить данные баланса: " + err.Error())
	}

	// 2. Рассчитываем общую долговую нагрузку
	totalDebt := balance.ShortTermDebt + balance.LongTermDebt

	// 3. Определяем долю долга с плавающей ставкой
	// Для упрощения предполагаем, что 30% долга имеет плавающую ставку
	// В реальной системе это должно быть получено из данных о долговых обязательствах
	floatingRateDebtRatio := 0.3
	floatingRateDebt := totalDebt * floatingRateDebtRatio

	// 4. Рассчитываем текущие процентные расходы
	// Предполагаем среднюю ставку 10% годовых
	currentInterestRate := 0.10
	currentInterestExpense := totalDebt * currentInterestRate

	// 5. Рассчитываем процентные расходы после изменения ставки
	newInterestRate := currentInterestRate + (rateChangePercent / 100)
	newInterestExpense := (totalDebt-floatingRateDebt)*currentInterestRate +
		floatingRateDebt*newInterestRate

	// 6. Рассчитываем изменение процентных расходов
	interestExpenseChange := newInterestExpense - currentInterestExpense

	// 7. Рассчитываем экспозицию (общий долг)
	exposure := totalDebt

	// 8. Определяем уровень риска
	riskLevel := determineInterestRiskLevel(interestExpenseChange, totalDebt, rateChangePercent)

	// 9. Рассчитываем стресс-тест (+3% к ставке)
	stressRateChange := 3.0
	stressNewRate := currentInterestRate + (stressRateChange / 100)
	stressNewExpense := (totalDebt-floatingRateDebt)*currentInterestRate +
		floatingRateDebt*stressNewRate
	stressTestLoss := stressNewExpense - currentInterestExpense

	// 10. Генерируем рекомендации
	recommendations := generateInterestRiskRecommendations(
		rateChangePercent,
		currentInterestRate,
		newInterestRate,
		interestExpenseChange,
		totalDebt,
		riskLevel,
	)

	// 11. Формируем результат
	result := &models.RiskResult{
		EnterpriseID:    enterpriseID,
		RiskType:        models.RiskTypeInterest,
		CalculationDate: time.Now(),
		HorizonDays:     365, // Процентный риск оценивается на год
		ConfidenceLevel: 0.95,
		ExposureAmount:  exposure,
		VaRValue:        interestExpenseChange,
		StressTestLoss:  stressTestLoss,
		RiskLevel:       riskLevel,
		Recommendations: recommendations,
		ScenarioType:    getInterestScenarioType(rateChangePercent),
		CreatedAt:       time.Now(),
	}

	return result, nil
}

// determineInterestRiskLevel определяет уровень процентного риска
func determineInterestRiskLevel(
	interestExpenseChange, totalDebt, rateChangePercent float64,
) models.RiskLevel {
	if totalDebt <= 0 {
		return models.RiskLevelLow
	}

	// Относительное изменение процентных расходов
	interestCoverageRatioImpact := 0.0
	if totalDebt > 0 {
		interestCoverageRatioImpact = (interestExpenseChange / totalDebt) * 100
	}

	absRateChange := rateChangePercent
	if absRateChange < 0 {
		absRateChange = -absRateChange
	}

	// Определяем уровень риска
	if interestCoverageRatioImpact > 15 || absRateChange > 2.0 {
		return models.RiskLevelHigh
	}
	if interestCoverageRatioImpact > 5 || absRateChange > 1.0 {
		return models.RiskLevelMedium
	}
	return models.RiskLevelLow
}

// getInterestScenarioType определяет тип сценария для процентного риска
func getInterestScenarioType(rateChangePercent float64) string {
	if rateChangePercent >= 2.0 {
		return "stress"
	}
	if rateChangePercent >= 1.0 {
		return "pessimistic"
	}
	if rateChangePercent <= -1.0 {
		return "optimistic"
	}
	return "base"
}

// generateInterestRiskRecommendations генерирует рекомендации по управлению процентным риском
func generateInterestRiskRecommendations(
	rateChangePercent, currentRate, newRate, expenseChange, totalDebt float64,
	riskLevel models.RiskLevel,
) []string {
	recommendations := []string{}

	// Определяем направление изменения
	isRateIncrease := rateChangePercent > 0
	absRateChange := rateChangePercent
	if absRateChange < 0 {
		absRateChange = -absRateChange
	}

	switch riskLevel {
	case models.RiskLevelHigh:
		if isRateIncrease {
			recommendations = append(recommendations,
				"🔴 КРИТИЧЕСКИЙ РИСК: Значительный рост процентных ставок",
				"Срочно рефинансировать долги с плавающей ставкой на фиксированную",
				"Рассмотреть возможность досрочного погашения долгов",
				"Оптимизировать структуру капитала (увеличить долю собственного капитала)",
				"Хеджировать процентный риск через свопы (если доступно)",
				"Пересмотреть инвестиционную программу (отложить несрочные проекты)",
			)
		} else {
			recommendations = append(recommendations,
				"🟢 ЗНАЧИТЕЛЬНОЕ СНИЖЕНИЕ СТАВОК: Благоприятные условия",
				"Рассмотреть возможность рефинансирования существующих долгов",
				"Привлечь дополнительное финансирование для развития",
				"Оптимизировать структуру капитала",
			)
		}

	case models.RiskLevelMedium:
		if isRateIncrease {
			recommendations = append(recommendations,
				"🟡 СРЕДНИЙ РИСК: Умеренный рост процентных ставок",
				"Рефинансировать долги с плавающей ставкой на фиксированную",
				"Оптимизировать структуру капитала",
				"Мониторинг процентных ставок еженедельно",
				"Рассмотреть возможность хеджирования процентного риска",
			)
		} else {
			recommendations = append(recommendations,
				"📈 УМЕРЕННОЕ СНИЖЕНИЕ СТАВОК: Улучшение условий",
				"Рассмотреть возможность рефинансирования долгов",
				"Оптимизировать структуру капитала",
			)
		}

	case models.RiskLevelLow:
		recommendations = append(recommendations,
			"🟢 СТАБИЛЬНЫЕ СТАВКИ: Текущая политика достаточна",
			"Поддерживать текущую структуру капитала",
			"Мониторинг процентных ставок 1 раз в месяц",
			"Планирование заёмной деятельности на основе прогнозов ставок",
		)
	}

	// Дополнительные рекомендации на основе конкретных значений
	if absRateChange > 1.5 {
		recommendations = append(recommendations,
			"⚠️ Значительное изменение ставки: >1.5%",
			"Провести детальный анализ влияния на финансовую устойчивость",
			"Скорректировать финансовый план",
		)
	}

	// Рекомендации по уровню долговой нагрузки
	debtToAssetsRatio := 0.0
	// В реальной системе нужно получить данные об активах
	// Для примера используем упрощённый расчёт
	if totalDebt > 0 {
		debtToAssetsRatio = totalDebt / (totalDebt + 1000000000) // 1 млрд - пример активов
	}

	if debtToAssetsRatio > 0.6 {
		recommendations = append(recommendations,
			"⚠️ ВЫСОКАЯ ДОЛГОВАЯ НАГРУЗКА: >60%",
			"Приоритетное снижение долговой нагрузки",
			"Увеличить долю собственного капитала",
		)
	} else if debtToAssetsRatio > 0.4 {
		recommendations = append(recommendations,
			"⚠️ УМЕРЕННАЯ ДОЛГОВАЯ НАГРУЗКА: >40%",
			"Контролировать рост долговой нагрузки",
			"Оптимизировать структуру капитала",
		)
	}

	// Рекомендации по изменению расходов
	if expenseChange > 0 {
		expenseChangePercent := (expenseChange / (totalDebt * 0.10)) * 100 // 10% - текущая ставка
		if expenseChangePercent > 20 {
			recommendations = append(recommendations,
				"⚠️ Значительный рост процентных расходов: >20%",
				"Скорректировать бюджет процентных расходов",
				"Пересмотреть план погашения долгов",
			)
		}
	}

	return recommendations
}

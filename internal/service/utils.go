package service

import (
	"financial-risk-server/internal/domain/models"
	"math"
)

// getZScore возвращает Z-оценку для заданного уровня доверия
func getZScore(confidenceLevel float64) float64 {
	switch {
	case confidenceLevel >= 0.99:
		return 2.33 // 99% уровень доверия
	case confidenceLevel >= 0.95:
		return 1.65 // 95% уровень доверия
	case confidenceLevel >= 0.90:
		return 1.28 // 90% уровень доверия
	default:
		return 1.65 // значение по умолчанию
	}
}

// determineRiskLevel определяет уровень риска на основе отношения VaR к экспозиции
func determineRiskLevel(varValue, exposure float64) models.RiskLevel {
	if exposure == 0 {
		return models.RiskLevelLow
	}

	ratio := (varValue / exposure) * 100 // в процентах

	switch {
	case ratio < 5:
		return models.RiskLevelLow
	case ratio < 15:
		return models.RiskLevelMedium
	default:
		return models.RiskLevelHigh
	}
}

// calculateVaR рассчитывает Value at Risk по формуле:
// VaR = Экспозиция × Волатильность × Z-оценка × √(горизонт/252)
func calculateVaR(
	exposure float64,
	volatility float64,
	confidenceLevel float64,
	horizonDays int,
) float64 {
	if exposure <= 0 || volatility <= 0 || horizonDays <= 0 {
		return 0
	}

	zScore := getZScore(confidenceLevel)
	timeFactor := math.Sqrt(float64(horizonDays) / 252.0) // 252 торговых дня в году

	return exposure * volatility * zScore * timeFactor
}

// calculateStressTestLoss рассчитывает потери при стресс-сценарии
func calculateStressTestLoss(exposure float64, stressFactor float64) float64 {
	if exposure <= 0 || stressFactor <= 0 {
		return 0
	}
	return exposure * stressFactor
}

// calculateExpectedLoss рассчитывает ожидаемые потери (для кредитного риска)
// EL = PD × LGD × EAD
func calculateExpectedLoss(
	probabilityOfDefault float64, // PD - вероятность дефолта
	lossGivenDefault float64, // LGD - потери при дефолте
	exposureAtDefault float64, // EAD - экспозиция при дефолте
) float64 {
	return probabilityOfDefault * lossGivenDefault * exposureAtDefault
}

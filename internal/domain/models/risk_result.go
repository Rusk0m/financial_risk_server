package models

import (
	"time"
)

// RiskType типы финансовых рисков
type RiskType string

const (
	RiskTypeCurrency  RiskType = "currency"  // Валютный риск
	RiskTypeCredit    RiskType = "credit"    // Кредитный риск
	RiskTypeLiquidity RiskType = "liquidity" // Риск ликвидности
	RiskTypeMarket    RiskType = "market"    // Фондовый риск
	RiskTypeInterest  RiskType = "interest"  // Процентный риск
)

// RiskLevel уровни риска
type RiskLevel string

const (
	RiskLevelLow    RiskLevel = "low"    // Низкий
	RiskLevelMedium RiskLevel = "medium" // Средний
	RiskLevelHigh   RiskLevel = "high"   // Высокий
)

// RiskResult представляет результат расчёта риска
type RiskResult struct {
	ID              int64
	EnterpriseID    int64
	RiskType        RiskType
	CalculationDate time.Time
	HorizonDays     int     // горизонт расчёта в днях
	ConfidenceLevel float64 // уровень доверия (0.95, 0.99)
	ExposureAmount  float64 // сумма экспозиции
	VaRValue        float64 // Value at Risk
	StressTestLoss  float64 // потери при стресс-сценарии
	RiskLevel       RiskLevel
	Recommendations []string // рекомендации по управлению риском
	ScenarioType    string   // "base", "pessimistic", "stress"
	CreatedAt       time.Time
}

// GetRiskPercentage возвращает риск в процентах от экспозиции
func (r *RiskResult) GetRiskPercentage() float64 {
	if r.ExposureAmount == 0 {
		return 0
	}
	return (r.VaRValue / r.ExposureAmount) * 100
}

// IsHighRisk проверяет, является ли риск высоким
func (r *RiskResult) IsHighRisk() bool {
	return r.RiskLevel == RiskLevelHigh || r.GetRiskPercentage() > 15
}

// IsMediumRisk проверяет, является ли риск средним
func (r *RiskResult) IsMediumRisk() bool {
	return r.RiskLevel == RiskLevelMedium || (r.GetRiskPercentage() > 5 && r.GetRiskPercentage() <= 15)
}

// IsLowRisk проверяет, является ли риск низким
func (r *RiskResult) IsLowRisk() bool {
	return r.RiskLevel == RiskLevelLow || r.GetRiskPercentage() <= 5
}

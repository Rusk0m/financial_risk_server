package models

import "time"

// ComprehensiveRiskReport представляет комплексный отчёт по рискам
type ComprehensiveRiskReport struct {
	ID           int64
	EnterpriseID int64
	ReportDate   time.Time
	PeriodStart  time.Time
	PeriodEnd    time.Time

	// Результаты по каждому типу риска
	CurrencyRisk  *RiskResult
	CreditRisk    *RiskResult
	LiquidityRisk *RiskResult
	MarketRisk    *RiskResult
	InterestRisk  *RiskResult

	// Сводные показатели
	TotalRiskValue   float64   // суммарный риск
	MaxRiskType      RiskType  // тип риска с максимальным значением
	OverallRiskLevel RiskLevel // общий уровень риска
	Summary          string    // текстовое резюме

	// Рекомендации
	PriorityActions []string // приоритетные действия
	Warnings        []string // предупреждения

	CreatedAt time.Time
}

// GetTotalRiskValue рассчитывает суммарное значение риска
func (r *ComprehensiveRiskReport) GetTotalRiskValue() float64 {
	total := 0.0
	if r.CurrencyRisk != nil {
		total += r.CurrencyRisk.VaRValue
	}
	if r.CreditRisk != nil {
		total += r.CreditRisk.VaRValue
	}
	if r.LiquidityRisk != nil {
		total += r.LiquidityRisk.VaRValue
	}
	if r.MarketRisk != nil {
		total += r.MarketRisk.VaRValue
	}
	if r.InterestRisk != nil {
		total += r.InterestRisk.VaRValue
	}
	return total
}

// GetRiskCountByLevel считает количество рисков по уровням
func (r *ComprehensiveRiskReport) GetRiskCountByLevel(level RiskLevel) int {
	count := 0
	risks := []*RiskResult{
		r.CurrencyRisk,
		r.CreditRisk,
		r.LiquidityRisk,
		r.MarketRisk,
		r.InterestRisk,
	}

	for _, risk := range risks {
		if risk != nil && risk.RiskLevel == level {
			count++
		}
	}
	return count
}

// HasHighRisks проверяет наличие высоких рисков
func (r *ComprehensiveRiskReport) HasHighRisks() bool {
	return r.GetRiskCountByLevel(RiskLevelHigh) > 0
}

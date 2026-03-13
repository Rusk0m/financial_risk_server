package models

import "time"

// ComprehensiveRiskReport представляет комплексный отчёт по рискам
type ComprehensiveRiskReport struct {
	ID                   int64      `json:"id"`
	EnterpriseID         int64      `json:"enterprise_id"`                 // ID предприятия (всегда 1)
	
	// Дата отчёта
	ReportDate           time.Time  `json:"report_date"`                   // Дата формирования отчёта
	
	// Сводные показатели по всем рискам
	OverallRiskLevel     string     `json:"overall_risk_level"`            // Общий уровень риска: low, medium, high
	TotalRiskValue       float64    `json:"total_risk_value"`              // Суммарный риск по всем типам
	MaxRiskType          string     `json:"max_risk_type"`                 // Тип риска с максимальным значением
	MaxRiskValue         float64    `json:"max_risk_value"`                // Значение максимального риска
	
	// Детализация по рискам (ссылки на расчёты)
	CurrencyRiskID       *int64     `json:"currency_risk_id,omitempty"`   // ID расчёта валютного риска
	InterestRiskID       *int64     `json:"interest_risk_id,omitempty"`   // ID расчёта процентного риска
	LiquidityRiskID      *int64     `json:"liquidity_risk_id,omitempty"`  // ID расчёта риска ликвидности
	
	// Текстовые поля для отчёта
	Summary              string     `json:"summary"`                       // Краткое резюме
	PriorityActions      []string   `json:"priority_actions"`              // Массив приоритетных действий
	Warnings             []string   `json:"warnings"`                      // Массив предупреждений
	
	// Метаданные
	CreatedAt            time.Time  `json:"created_at"`
	CreatedBy            *string    `json:"created_by,omitempty"`         // Кто сформировал отчёт
}

// GetOverallRiskColor возвращает цвет для общего уровня риска
func (crr *ComprehensiveRiskReport) GetOverallRiskColor() string {
	switch crr.OverallRiskLevel {
	case "high":
		return "#EF4444"
	case "medium":
		return "#F59E0B"
	case "low":
		return "#10B981"
	default:
		return "#64748B"
	}
}

// GetOverallRiskLabel возвращает текстовую метку общего уровня риска
func (crr *ComprehensiveRiskReport) GetOverallRiskLabel() string {
	switch crr.OverallRiskLevel {
	case "high":
		return "КРИТИЧЕСКИЙ УРОВЕНЬ РИСКА"
	case "medium":
		return "СРЕДНИЙ УРОВЕНЬ РИСКА"
	case "low":
		return "НИЗКИЙ УРОВЕНЬ РИСКА"
	default:
		return "НЕИЗВЕСТНЫЙ УРОВЕНЬ"
	}
}

// GetRiskDistribution возвращает распределение рисков по типам в процентах
func (crr *ComprehensiveRiskReport) GetRiskDistribution(currencyRisk, interestRisk, liquidityRisk *RiskCalculation) map[string]float64 {
	total := crr.TotalRiskValue
	if total == 0 {
		return map[string]float64{}
	}
	
	distribution := make(map[string]float64)
	
	if currencyRisk != nil {
		distribution["Валютный"] = (currencyRisk.VaRValue / total) * 100
	}
	if interestRisk != nil {
		distribution["Процентный"] = (interestRisk.VaRValue / total) * 100
	}
	if liquidityRisk != nil {
		distribution["Ликвидности"] = (liquidityRisk.VaRValue / total) * 100
	}
	
	return distribution
}

// HasCriticalWarnings проверяет наличие критических предупреждений
func (crr *ComprehensiveRiskReport) HasCriticalWarnings() bool {
	return crr.OverallRiskLevel == "high" || len(crr.Warnings) > 0
}

// GetPriorityActionsCount возвращает количество приоритетных действий
func (crr *ComprehensiveRiskReport) GetPriorityActionsCount() int {
	return len(crr.PriorityActions)
}
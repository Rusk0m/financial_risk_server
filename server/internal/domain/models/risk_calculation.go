package models

import "time"

// RiskCalculation представляет расчёт финансового риска
type RiskCalculation struct {
	ID                int64      `json:"id"`
	EnterpriseID      int64      `json:"enterprise_id"`                 // ID предприятия (всегда 1)
	ReportID          *int64     `json:"report_id,omitempty"`          // ID отчёта, на основе которого сделан расчёт
	
	// Параметры расчёта
	RiskType          string     `json:"risk_type"`                     // Тип риска: currency, interest, liquidity
	CalculationDate   time.Time  `json:"calculation_date"`              // Дата расчёта
	HorizonDays       int        `json:"horizon_days"`                  // Горизонт расчёта в днях
	ConfidenceLevel   float64    `json:"confidence_level"`              // Уровень доверия (0.95, 0.99)
	
	// Результаты расчёта
	ExposureAmount    float64    `json:"exposure_amount"`               // Экспозиция (базовая сумма риска)
	VaRValue          float64    `json:"var_value"`                     // Value at Risk
	StressTestLoss    float64    `json:"stress_test_loss"`              // Потери в стресс-сценарии
	RiskLevel         string     `json:"risk_level"`                    // Уровень риска: low, medium, high
	
	// Детали расчёта (опционально для отладки)
	CalculationMethod *string    `json:"calculation_method,omitempty"` // Метод расчёта
	ScenarioType      *string    `json:"scenario_type,omitempty"`      // Тип сценария
	Assumptions       *string    `json:"assumptions,omitempty"`        // Допущения расчёта в JSON
	
	// Рекомендации
	Recommendations   []string   `json:"recommendations"`               // Массив текстовых рекомендаций
	
	// Метаданные
	CreatedAt         time.Time  `json:"created_at"`
}

// GetRiskPercentage возвращает риск в процентах от экспозиции
func (rc *RiskCalculation) GetRiskPercentage() float64 {
	if rc.ExposureAmount == 0 {
		return 0
	}
	return (rc.VaRValue / rc.ExposureAmount) * 100
}

// GetStressTestPercentage возвращает стресс-тест в процентах от экспозиции
func (rc *RiskCalculation) GetStressTestPercentage() float64 {
	if rc.ExposureAmount == 0 {
		return 0
	}
	return (rc.StressTestLoss / rc.ExposureAmount) * 100
}

// IsHighRisk проверяет, является ли риск высоким
func (rc *RiskCalculation) IsHighRisk() bool {
	return rc.RiskLevel == "high"
}

// IsMediumRisk проверяет, является ли риск средним
func (rc *RiskCalculation) IsMediumRisk() bool {
	return rc.RiskLevel == "medium"
}

// IsLowRisk проверяет, является ли риск низким
func (rc *RiskCalculation) IsLowRisk() bool {
	return rc.RiskLevel == "low"
}

// GetRiskColor возвращает цвет для визуализации риска
func (rc *RiskCalculation) GetRiskColor() string {
	switch rc.RiskLevel {
	case "high":
		return "#EF4444" // Красный
	case "medium":
		return "#F59E0B" // Жёлтый
	case "low":
		return "#10B981" // Зелёный
	default:
		return "#64748B" // Серый
	}
}

// GetRiskLabel возвращает текстовую метку риска
func (rc *RiskCalculation) GetRiskLabel() string {
	switch rc.RiskLevel {
	case "high":
		return "ВЫСОКИЙ РИСК"
	case "medium":
		return "СРЕДНИЙ РИСК"
	case "low":
		return "НИЗКИЙ РИСК"
	default:
		return "НЕИЗВЕСТНЫЙ"
	}
}
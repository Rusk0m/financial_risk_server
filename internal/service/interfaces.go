package service

import (
	"finantial-risk-server/internal/domain/models"
)

// RiskCalculationService определяет методы для расчёта финансовых рисков
type RiskCalculationService interface {
	// Валютный риск
	CalculateCurrencyRisk(
		enterpriseID int64,
		horizonDays int,
		confidenceLevel float64,
	) (*models.RiskResult, error)

	// Кредитный риск
	CalculateCreditRisk(
		enterpriseID int64,
	) (*models.RiskResult, error)

	// Риск ликвидности
	CalculateLiquidityRisk(
		enterpriseID int64,
	) (*models.RiskResult, error)

	// Фондовый (рыночный) риск
	CalculateMarketRisk(
		enterpriseID int64,
		priceChangePercent float64,
	) (*models.RiskResult, error)

	// Процентный риск
	CalculateInterestRisk(
		enterpriseID int64,
		rateChangePercent float64,
	) (*models.RiskResult, error)

	// Комплексный расчёт всех рисков
	CalculateAllRisks(
		enterpriseID int64,
		horizonDays int,
		confidenceLevel float64,
		priceChangePercent float64,
		rateChangePercent float64,
	) (*models.ComprehensiveRiskReport, error)
}

// RiskCalculationServiceConfig конфигурация сервиса расчёта рисков
type RiskCalculationServiceConfig struct {
	DefaultHorizonDays     int     // 30 дней по умолчанию
	DefaultConfidenceLevel float64 // 0.95 по умолчанию
	DefaultPriceChangePct  float64 // -15% по умолчанию (пессимистичный сценарий)
	DefaultRateChangePct   float64 // 1.0% по умолчанию
}

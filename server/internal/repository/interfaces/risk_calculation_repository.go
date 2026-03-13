package interfaces

import (
	"context"
	"financial-risk-server/internal/domain/models"
)

// RiskCalculationFilter параметры фильтрации расчётов рисков
type RiskCalculationFilter struct {
	EnterpriseID *int64
	RiskType     *string
	StartDate    *string // YYYY-MM-DD
	EndDate      *string // YYYY-MM-DD
	Limit        int
	Offset       int
}

// RiskCalculationRepository определяет методы для работы с расчётами рисков
type RiskCalculationRepository interface {
	// Create создаёт новый расчёт риска
	Create(ctx context.Context, calculation *models.RiskCalculation) error
	
	// GetByID получает расчёт по ID
	GetByID(ctx context.Context, id int64) (*models.RiskCalculation, error)
	
	// GetAll получает расчёты с фильтрацией
	GetAll(ctx context.Context, filter RiskCalculationFilter) ([]*models.RiskCalculation, error)
	
	// GetLatestByType получает последний расчёт по типу риска для предприятия
	GetLatestByType(ctx context.Context, enterpriseID int64, riskType string) (*models.RiskCalculation, error)
	
	// GetAllByEnterpriseID получает все расчёты для предприятия
	GetAllByEnterpriseID(ctx context.Context, enterpriseID int64) ([]*models.RiskCalculation, error)
	
	// Update обновляет расчёт
	Update(ctx context.Context, calculation *models.RiskCalculation) error
	
	// Delete удаляет расчёт по ID
	Delete(ctx context.Context, id int64) error
}
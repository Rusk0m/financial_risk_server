package interfaces

import (
	"context"
	"financial-risk-server/internal/domain/models"
)

// ComprehensiveRiskReportRepository определяет методы для работы с комплексными отчётами
type ComprehensiveRiskReportRepository interface {
	// Create создаёт новый комплексный отчёт
	Create(ctx context.Context, report *models.ComprehensiveRiskReport) error
	
	// GetByID получает отчёт по ID
	GetByID(ctx context.Context, id int64) (*models.ComprehensiveRiskReport, error)
	
	// GetAllByEnterpriseID получает все отчёты для предприятия
	GetAllByEnterpriseID(ctx context.Context, enterpriseID int64) ([]*models.ComprehensiveRiskReport, error)
	
	// GetLatest получает последний отчёт для предприятия
	GetLatest(ctx context.Context, enterpriseID int64) (*models.ComprehensiveRiskReport, error)
	
	// GetAll получает все отчёты
	GetAll(ctx context.Context) ([]*models.ComprehensiveRiskReport, error)
	
	// Update обновляет отчёт
	Update(ctx context.Context, report *models.ComprehensiveRiskReport) error
	
	// Delete удаляет отчёт по ID
	Delete(ctx context.Context, id int64) error
}
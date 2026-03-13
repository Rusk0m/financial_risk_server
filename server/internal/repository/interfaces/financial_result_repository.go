package interfaces

import (
	"context"
	"financial-risk-server/internal/domain/models"
)

// FinancialResultRepository определяет методы для работы с отчётами о финансовых результатах
type FinancialResultRepository interface {
	// Create создаёт новый отчёт о ФР
	Create(ctx context.Context, result *models.FinancialResult) error
	
	// GetByID получает отчёт по ID
	GetByID(ctx context.Context, id int64) (*models.FinancialResult, error)
	
	// GetByEnterpriseID получает отчёты предприятия
	GetByEnterpriseID(ctx context.Context, enterpriseID int64) ([]*models.FinancialResult, error)
	
	// GetByReportID получает отчёт из отчёта
	GetByReportID(ctx context.Context, reportID int64) (*models.FinancialResult, error)
	
	// GetAll получает все отчёты
	GetAll(ctx context.Context) ([]*models.FinancialResult, error)
	
	// Update обновляет отчёт
	Update(ctx context.Context, result *models.FinancialResult) error
	
	// Delete удаляет отчёт по ID
	Delete(ctx context.Context, id int64) error
}
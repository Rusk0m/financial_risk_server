package interfaces

import (
	"context"
	"financial-risk-server/internal/domain/models"
)

// ExportContractRepository определяет методы для работы с экспортными контрактами
type ExportContractRepository interface {
	// Create создаёт новый контракт
	Create(ctx context.Context, contract *models.ExportContract) error
	
	// GetByID получает контракт по ID
	GetByID(ctx context.Context, id int64) (*models.ExportContract, error)
	
	// GetByEnterpriseID получает контракты предприятия
	GetByEnterpriseID(ctx context.Context, enterpriseID int64) ([]*models.ExportContract, error)
	
	// GetByReportID получает контракты из отчёта
	GetByReportID(ctx context.Context, reportID int64) ([]*models.ExportContract, error)
	
	// GetAll получает все контракты
	GetAll(ctx context.Context) ([]*models.ExportContract, error)
	
	// Update обновляет контракт
	Update(ctx context.Context, contract *models.ExportContract) error
	
	// Delete удаляет контракт по ID
	Delete(ctx context.Context, id int64) error
}
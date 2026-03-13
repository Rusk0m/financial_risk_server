package interfaces

import (
	"context"
	"financial-risk-server/internal/domain/models"
)

// BalanceSheetRepository определяет методы для работы с финансовыми балансами
type BalanceSheetRepository interface {
	// Create создаёт новый баланс
	Create(ctx context.Context, balance *models.BalanceSheet) error
	
	// GetByID получает баланс по ID
	GetByID(ctx context.Context, id int64) (*models.BalanceSheet, error)
	
	// GetByEnterpriseID получает балансы предприятия
	GetByEnterpriseID(ctx context.Context, enterpriseID int64) ([]*models.BalanceSheet, error)
	
	// GetByReportID получает баланс из отчёта
	GetByReportID(ctx context.Context, reportID int64) (*models.BalanceSheet, error)
	
	// GetLatest получает последний баланс предприятия
	GetLatest(ctx context.Context, enterpriseID int64) (*models.BalanceSheet, error)
	
	// GetAll получает все балансы
	GetAll(ctx context.Context) ([]*models.BalanceSheet, error)
	
	// Update обновляет баланс
	Update(ctx context.Context, balance *models.BalanceSheet) error
	
	// Delete удаляет баланс по ID
	Delete(ctx context.Context, id int64) error
}
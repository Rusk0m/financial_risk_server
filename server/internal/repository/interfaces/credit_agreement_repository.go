package interfaces

import (
	"context"
	"financial-risk-server/internal/domain/models"
)

// CreditAgreementRepository определяет методы для работы с кредитными договорами
type CreditAgreementRepository interface {
	// Create создаёт новый кредитный договор
	Create(ctx context.Context, agreement *models.CreditAgreement) error
	
	// GetByID получает договор по ID
	GetByID(ctx context.Context, id int64) (*models.CreditAgreement, error)
	
	// GetByEnterpriseID получает договоры предприятия
	GetByEnterpriseID(ctx context.Context, enterpriseID int64) ([]*models.CreditAgreement, error)
	
	// GetByReportID получает договоры из отчёта
	GetByReportID(ctx context.Context, reportID int64) ([]*models.CreditAgreement, error)
	
	// GetAll получает все договоры
	GetAll(ctx context.Context) ([]*models.CreditAgreement, error)
	
	// Update обновляет договор
	Update(ctx context.Context, agreement *models.CreditAgreement) error
	
	// Delete удаляет договор по ID
	Delete(ctx context.Context, id int64) error
}
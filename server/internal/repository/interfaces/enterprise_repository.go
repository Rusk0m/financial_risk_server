package interfaces

import (
	"context"
	"financial-risk-server/internal/domain/models"
)

// EnterpriseRepository определяет методы для работы с предприятиями
type EnterpriseRepository interface {
	// Create создаёт новое предприятие
	Create(ctx context.Context, enterprise *models.Enterprise) error
	
	// GetByID получает предприятие по ID
	GetByID(ctx context.Context, id int64) (*models.Enterprise, error)
	
	// Update обновляет предприятие
	Update(ctx context.Context, enterprise *models.Enterprise) error
	
	// Delete удаляет предприятие по ID
	Delete(ctx context.Context, id int64) error
}
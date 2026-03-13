package interfaces

import (
	"context"
	"financial-risk-server/internal/domain/models"
)

// ReportFilter параметры фильтрации отчётов
type ReportFilter struct {
	EnterpriseID *int64
	ReportType   *string
	Status       *string
	PeriodStart  *string // YYYY-MM-DD
	PeriodEnd    *string // YYYY-MM-DD
	Limit        int
	Offset       int
}

// ReportRepository определяет методы для работы с отчётами
type ReportRepository interface {
	// Create создаёт новый отчёт
	Create(ctx context.Context, report *models.Report) error
	
	// GetByID получает отчёт по ID
	GetByID(ctx context.Context, id int64) (*models.Report, error)
	
	// GetAll получает список отчётов с фильтрацией
	GetAll(ctx context.Context, filter ReportFilter) ([]*models.Report, error)
	
	// UpdateProcessingStatus обновляет статус обработки отчёта
	UpdateProcessingStatus(ctx context.Context, id int64, status string, errorMessage *string) error
	
	// Delete удаляет отчёт по ID
	Delete(ctx context.Context, id int64) error
}
package service

import (
	"context"
	"fmt"
	"financial-risk-server/internal/domain/models"
	"financial-risk-server/internal/repository/interfaces"
)

// EnterpriseService предоставляет бизнес-логику для работы с предприятиями
type EnterpriseService struct {
	enterpriseRepo interfaces.EnterpriseRepository
}

// NewEnterpriseService создаёт новый сервис предприятий
func NewEnterpriseService(enterpriseRepo interfaces.EnterpriseRepository) *EnterpriseService {
	return &EnterpriseService{
		enterpriseRepo: enterpriseRepo,
	}
}

// GetPrimaryEnterprise возвращает основное предприятие (ОАО «Беларуськалий»)
func (s *EnterpriseService) GetPrimaryEnterprise(ctx context.Context) (*models.Enterprise, error) {
	// Для приложения с одним предприятием всегда возвращаем ID = 1
	enterprise, err := s.enterpriseRepo.GetByID(ctx, 1)
	if err != nil {
		return nil, fmt.Errorf("failed to get primary enterprise: %w", err)
	}
	return enterprise, nil
}

// GetEnterpriseByID возвращает предприятие по ID (для совместимости)
func (s *EnterpriseService) GetEnterpriseByID(ctx context.Context, id int64) (*models.Enterprise, error) {
	enterprise, err := s.enterpriseRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get enterprise by id %d: %w", id, err)
	}
	return enterprise, nil
}

// IsSingleEnterpriseMode всегда возвращает true для нашего приложения
func (s *EnterpriseService) IsSingleEnterpriseMode() bool {
	return true
}

// GetPrimaryEnterpriseID всегда возвращает 1
func (s *EnterpriseService) GetPrimaryEnterpriseID() int64 {
	return 1
}
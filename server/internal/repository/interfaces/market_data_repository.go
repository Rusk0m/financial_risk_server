package interfaces

import (
	"context"
	"financial-risk-server/internal/domain/models"
)

// MarketDataFilter параметры фильтрации рыночных данных
type MarketDataFilter struct {
	DataDate     *string // YYYY-MM-DD
	CurrencyPair *string
	Limit        int
	Offset       int
}

// MarketDataRepository определяет методы для работы с рыночными данными
type MarketDataRepository interface {
	// Create создаёт новую запись рыночных данных
	Create(ctx context.Context, data *models.MarketData) error
	
	// GetByID получает запись по ID
	GetByID(ctx context.Context, id int64) (*models.MarketData, error)
	
	// GetAll получает записи с фильтрацией
	GetAll(ctx context.Context, filter MarketDataFilter) ([]*models.MarketData, error)
	
	// GetLatestByCurrencyPair получает последнюю запись по валютной паре
	GetLatestByCurrencyPair(ctx context.Context, currencyPair string) (*models.MarketData, error)
	
	// Update обновляет запись
	Update(ctx context.Context, data *models.MarketData) error
	
	// Delete удаляет запись по ID
	Delete(ctx context.Context, id int64) error
}
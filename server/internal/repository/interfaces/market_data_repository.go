package interfaces

import (
	"context"
	"financial-risk-server/internal/domain/models"
	"time"
)

type MarketDataRepository interface {
	// Create создаёт новую запись
	Create(ctx context.Context, data *models.MarketData) error
	
	// GetByDateAndPair получает запись по дате и типу
	GetByDateAndPair(ctx context.Context, date time.Time, pair string) (*models.MarketData, error)
	
	// GetHistory получает историю за период
	GetHistory(ctx context.Context, pair string, days int) ([]*models.MarketData, error)
	
	// GetLatest получает последнюю запись
	GetLatest(ctx context.Context, pair string) (*models.MarketData, error)
	
	// GetLatestRates получает последние курсы всех валют
	GetLatestRates(ctx context.Context) (map[string]float64, error)
	
	// Update обновляет запись
	Update(ctx context.Context, data *models.MarketData) error
}
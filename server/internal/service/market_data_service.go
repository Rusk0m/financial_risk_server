package service

import (
	"context"
	"fmt"
	"time"

	"financial-risk-server/internal/domain/models"
	"financial-risk-server/internal/repository/interfaces"
)

// MarketDataService предоставляет бизнес-логику для работы с рыночными данными
type MarketDataService struct {
	marketDataRepo interfaces.MarketDataRepository
}

// NewMarketDataService создаёт новый сервис рыночных данных
func NewMarketDataService(marketDataRepo interfaces.MarketDataRepository) *MarketDataService {
	return &MarketDataService{
		marketDataRepo: marketDataRepo,
	}
}

// GetLatestExchangeRate получает последний курс валюты
func (s *MarketDataService) GetLatestExchangeRate(ctx context.Context, currencyPair string) (float64, error) {
	data, err := s.marketDataRepo.GetLatestByCurrencyPair(ctx, currencyPair)
	if err != nil {
		return 0, fmt.Errorf("failed to get latest exchange rate for %s: %w", currencyPair, err)
	}
	return data.ExchangeRate, nil
}

// GetLatestPotassiumPrice получает последнюю цену калия
func (s *MarketDataService) GetLatestPotassiumPrice(ctx context.Context) (float64, error) {
	data, err := s.marketDataRepo.GetLatestByCurrencyPair(ctx, "POTASSIUM")
	if err != nil {
		return 0, fmt.Errorf("failed to get latest potassium price: %w", err)
	}
	if data.PotassiumPriceUSD == nil {
		return 0, fmt.Errorf("potassium price not available")
	}
	return *data.PotassiumPriceUSD, nil
}

// UpdateExchangeRate обновляет курс валюты
func (s *MarketDataService) UpdateExchangeRate(
	ctx context.Context,
	currencyPair string,
	exchangeRate float64,
	volatility30d *float64,
	source string,
) error {
	data := &models.MarketData{
		DataDate:       time.Now(),
		CurrencyPair:   currencyPair,
		ExchangeRate:   exchangeRate,
		Volatility30d:  volatility30d,
		Source:         source,
	}

	return s.marketDataRepo.Create(ctx, data)
}

// UpdatePotassiumPrice обновляет цену калия
func (s *MarketDataService) UpdatePotassiumPrice(
	ctx context.Context,
	priceUSD float64,
	source string,
) error {
	data := &models.MarketData{
		DataDate:          time.Now(),
		CurrencyPair:      "POTASSIUM",
		PotassiumPriceUSD: &priceUSD,
		Source:            source,
	}

	return s.marketDataRepo.Create(ctx, data)
}

// GetHistoricalData получает исторические данные за период
func (s *MarketDataService) GetHistoricalData(
	ctx context.Context,
	currencyPair string,
	startDate, endDate string,
) ([]*models.MarketData, error) {
	filter := interfaces.MarketDataFilter{
		CurrencyPair: &currencyPair,
		DataDate:     &startDate,
		Limit:        100,
	}
	
	// TODO: Добавить фильтрацию по endDate
	return s.marketDataRepo.GetAll(ctx, filter)
}
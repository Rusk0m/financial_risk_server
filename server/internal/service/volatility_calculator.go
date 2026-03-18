package service

import (
	"context"
	"math"

	"financial-risk-server/internal/domain/models"
	"financial-risk-server/internal/repository/interfaces"
)

// VolatilityCalculator рассчитывает волатильность on-demand
type VolatilityCalculator struct {
	marketRepo interfaces.MarketDataRepository
}

func NewVolatilityCalculator(marketRepo interfaces.MarketDataRepository) *VolatilityCalculator {
	return &VolatilityCalculator{marketRepo: marketRepo}
}

// CalculateVolatility рассчитывает волатильность за период (возвращает дневную волатильность)
func (c *VolatilityCalculator) CalculateVolatility(ctx context.Context, pair string, days int) (float64, error) {
	data, err := c.marketRepo.GetHistory(ctx, pair, days+1)
	if err != nil {
		return 0.09 / math.Sqrt(252), nil
	}

	if len(data) < 2 {
		return 0.09 / math.Sqrt(252), nil
	}

	returns := c.calculateLogReturns(data)
	if len(returns) == 0 {
		return 0.09/math.Sqrt(252), nil
	}

	return stdDev(returns), nil
}

// calculateLogReturns считает логарифмические доходности
func (c *VolatilityCalculator) calculateLogReturns(data []*models.MarketData) []float64 {
	returns := make([]float64, 0, len(data)-1)

	for i := 0; i < len(data)-1; i++ {
		var currentPrice, nextPrice float64
		
		if data[i].ExchangeRate != nil && data[i+1].ExchangeRate != nil { // Для валют
			currentPrice = *data[i].ExchangeRate
			nextPrice = *data[i+1].ExchangeRate
		} else if data[i].PotassiumPriceUSD != nil && data[i+1].PotassiumPriceUSD != nil { // Для калия
			currentPrice = *data[i].PotassiumPriceUSD
			nextPrice = *data[i+1].PotassiumPriceUSD
		} else {
			continue
		}

		if currentPrice > 0 && nextPrice > 0 {
			logReturn := math.Log(nextPrice / currentPrice)
			returns = append(returns, logReturn)
		}
	}

	return returns
}

// stdDev считает стандартное отклонение
func stdDev(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	mean := 0.0
	for _, v := range values {
		mean += v
	}
	mean /= float64(len(values))

	variance := 0.0
	for _, v := range values {
		variance += (v - mean) * (v - mean)
	}
	variance /= float64(len(values))

	return math.Sqrt(variance)
}

// CalculateVaR рассчитывает Value at Risk
func (c *VolatilityCalculator) CalculateVaR(exposure float64, volatility float64, confidenceLevel float64, horizonDays int) float64 {
	zScore := getZScore(confidenceLevel)
	return exposure * volatility * zScore * math.Sqrt(float64(horizonDays))
}

// // getZScore возвращает Z-оценку для заданного уровня доверия
// func getZScore(confidenceLevel float64) float64 {
// 	switch {
// 	case confidenceLevel >= 0.99:
// 		return 2.326
// 	case confidenceLevel >= 0.95:
// 		return 1.645
// 	case confidenceLevel >= 0.90:
// 		return 1.282
// 	default:
// 		return 1.645 // Default 95%
// 	}
// }
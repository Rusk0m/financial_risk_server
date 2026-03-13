package models

import "time"

// MarketData представляет рыночные данные (курсы валют, цены на калий)
type MarketData struct {
	ID                int64     `json:"id"`
	DataDate          time.Time `json:"data_date"`                     // Дата данных
	CurrencyPair      string    `json:"currency_pair"`                 // Валютная пара (BYN/USD, POTASSIUM)
	ExchangeRate      float64   `json:"exchange_rate"`                 // Курс обмена
	Volatility30d     *float64  `json:"volatility_30d,omitempty"`      // Волатильность за 30 дней
	PotassiumPriceUSD *float64  `json:"potassium_price_usd,omitempty"` // Цена калия в USD за тонну
	Source            string    `json:"source"`                        // Источник данных
	
	// Метаданные
	CreatedAt         time.Time `json:"created_at"`
}

// IsCurrencyData проверяет, являются ли данные валютными
func (md *MarketData) IsCurrencyData() bool {
	return md.CurrencyPair != "POTASSIUM"
}

// IsPotassiumData проверяет, являются ли данные о цене калия
func (md *MarketData) IsPotassiumData() bool {
	return md.CurrencyPair == "POTASSIUM" || md.PotassiumPriceUSD != nil
}

// GetVolatilityPercent возвращает волатильность в процентах
func (md *MarketData) GetVolatilityPercent() float64 {
	if md.Volatility30d != nil {
		return *md.Volatility30d * 100
	}
	return 0
}
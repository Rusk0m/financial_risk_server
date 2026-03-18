package models

import "time"

// MarketData представляет рыночные данные (курсы валют, цены на калий)
type MarketData struct {
	ID                int64     `json:"id"`
	DataDate          time.Time `json:"data_date"`                     // Дата данных
	CurrencyPair      string    `json:"currency_pair"`                 // Валютная пара (BYN/USD, ..., POTASSIUM)
	ExchangeRate      *float64   `json:"exchange_rate"`                 // Курс обмена
	PotassiumPriceUSD *float64  `json:"potassium_price_usd,omitempty"` // Цена калия в USD за тонну
	Source            string    `json:"source"`                        // Источник данных
	
	// Метаданные
	CreatedAt         time.Time `json:"created_at"`
}

// SupportedCurrencyPairs возвращает список поддерживаемых валютных пар
func SupportedCurrencyPairs() []string {
	return []string{"BYN/USD", "BYN/EUR", "BYN/CNY"}
}

// IsCurrencyPair проверяет, является ли пара валютной
func (m *MarketData) IsCurrencyPair() bool {
	for _, pair := range SupportedCurrencyPairs() {
		if m.CurrencyPair == pair {
			return true
		}
	}
	return false
}
// TableName возвращает имя таблицы в БД
func (MarketData) TableName() string {
	return "market_data"
}
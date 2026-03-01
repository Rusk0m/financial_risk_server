package models

import (
	"errors"
	"time"
)

// MarketData представляет рыночные данные
type MarketData struct {
	ID                int64
	DataDate          time.Time
	CurrencyPair      string  // например, "BYN/USD"
	ExchangeRate      float64 // курс валюты
	Volatility30d     float64 // волатильность за 30 дней
	PotassiumPriceUSD float64 // цена калия в $/т
	Source            string  // источник данных
	CreatedAt         time.Time
}

// IsValid проверяет валидность рыночных данных
func (m *MarketData) IsValid() error {
	if m.DataDate.IsZero() {
		return errors.New("дата данных должна быть указана")
	}
	if m.CurrencyPair == "" {
		return errors.New("валютная пара не может быть пустой")
	}
	if m.ExchangeRate <= 0 {
		return errors.New("курс валюты должен быть больше 0")
	}
	if m.Volatility30d < 0 {
		return errors.New("волатильность не может быть отрицательной")
	}
	if m.PotassiumPriceUSD < 0 {
		return errors.New("цена калия не может быть отрицательной")
	}
	if m.Source == "" {
		return errors.New("источник данных не может быть пустым")
	}
	return nil
}

// IsRecent проверяет, являются ли данные свежими (не старше 7 дней)
func (m *MarketData) IsRecent() bool {
	return time.Since(m.DataDate) < 7*24*time.Hour
}

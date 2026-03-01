package models

import (
	"errors"
	"time"
)

// ExportContract представляет экспортный контракт
type ExportContract struct {
	ID              int64
	EnterpriseID    int64
	ContractDate    time.Time
	Country         string  // страна назначения
	VolumeT         float64 // объём отгрузки в тоннах
	PriceUSDPerT    float64 // цена контракта в $/т
	Currency        string  // валюта контракта
	PaymentTermDays int     // срок оплаты в днях
	ShipmentDate    time.Time
	PaymentStatus   string  // "pending", "paid", "overdue"
	ExchangeRate    float64 // курс на дату контракта (BYN/USD)
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// IsValid проверяет валидность данных контракта
func (c *ExportContract) IsValid() error {
	if c.EnterpriseID <= 0 {
		return errors.New("enterprise_id должен быть больше 0")
	}
	if c.Country == "" {
		return errors.New("страна назначения не может быть пустой")
	}
	if c.VolumeT <= 0 {
		return errors.New("объём отгрузки должен быть больше 0")
	}
	if c.PriceUSDPerT <= 0 {
		return errors.New("цена контракта должна быть больше 0")
	}
	if c.Currency == "" {
		return errors.New("валюта контракта не может быть пустой")
	}
	if c.PaymentTermDays <= 0 || c.PaymentTermDays > 365 {
		return errors.New("срок оплаты должен быть от 1 до 365 дней")
	}
	if c.PaymentStatus == "" {
		c.PaymentStatus = "pending"
	}
	if c.ExchangeRate <= 0 {
		return errors.New("курс валюты должен быть больше 0")
	}
	return nil
}

// GetContractValue возвращает стоимость контракта в валюте контракта
func (c *ExportContract) GetContractValue() float64 {
	return c.VolumeT * c.PriceUSDPerT
}

// GetExpectedPaymentDate возвращает ожидаемую дату оплаты
func (c *ExportContract) GetExpectedPaymentDate() time.Time {
	return c.ShipmentDate.AddDate(0, 0, c.PaymentTermDays)
}

// IsPending проверяет, ожидается ли оплата
func (c *ExportContract) IsPending() bool {
	return c.PaymentStatus == "pending"
}

// IsOverdue проверяет, просрочен ли контракт
func (c *ExportContract) IsOverdue() bool {
	return time.Now().After(c.GetExpectedPaymentDate()) && c.PaymentStatus == "pending"
}

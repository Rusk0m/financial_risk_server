package models

import (
	"errors"
	"time"
)

// CostStructure представляет статью затрат
type CostStructure struct {
	ID           int64
	EnterpriseID int64
	CostItem     string  // статья затрат (ЗП, энергия, логистика и т.д.)
	Currency     string  // валюта затрат
	CostPerT     float64 // стоимость на 1 тонну
	PeriodStart  time.Time
	PeriodEnd    time.Time
	Comment      string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// IsValid проверяет валидность данных затрат
func (c *CostStructure) IsValid() error {
	if c.EnterpriseID <= 0 {
		return errors.New("enterprise_id должен быть больше 0")
	}
	if c.CostItem == "" {
		return errors.New("статья затрат не может быть пустой")
	}
	if c.Currency == "" {
		return errors.New("валюта затрат не может быть пустой")
	}
	if c.CostPerT < 0 {
		return errors.New("стоимость на 1 тонну не может быть отрицательной")
	}
	if c.PeriodStart.IsZero() || c.PeriodEnd.IsZero() {
		return errors.New("период затрат должен быть указан")
	}
	if c.PeriodEnd.Before(c.PeriodStart) {
		return errors.New("дата окончания периода не может быть раньше даты начала")
	}
	return nil
}

// GetTotalCost возвращает общую стоимость затрат за период
func (c *CostStructure) GetTotalCost(annualProductionT float64) float64 {
	return c.CostPerT * annualProductionT
}

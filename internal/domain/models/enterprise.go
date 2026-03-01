package models

import (
	"errors"
	"time"
)

type Enterprise struct {
	ID                 int64
	Name               string
	Industry           string
	AnnualProductionT  float64 //тонн в год
	ExportSharePercent float64 // доля экспорта в %
	MainCurrency       string  // оснавная ваюта расчета
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// IsValid проверяет валидность данных предприятия
func (e *Enterprise) IsValid() error {
	if e.Name == "" {
		return errors.New("имя предприятия не может быть пустым")
	}
	if len(e.Name) > 255 {
		return errors.New("имя предприятия не может быть диннее чем 255 символов")
	}
	if e.AnnualProductionT <= 0 {
		return errors.New("годовой объём производства должен быть больше 0")
	}
	if e.ExportSharePercent < 0 || e.ExportSharePercent > 100 {
		return errors.New("доля экспорта должна быть от 0 до 100%")
	}
	if e.MainCurrency == "" {
		return errors.New("основная валюта не может быть пустой")
	}
	if len(e.MainCurrency) != 3 {
		return errors.New("код валюты должен состоять из 3 символов (например, USD)")
	}
	return nil
}

// IsExportOriented проверяет, является ли предприятие экспортно-ориентированным
func (e *Enterprise) IsExportOriented() bool {
	return e.ExportSharePercent >= 50
}

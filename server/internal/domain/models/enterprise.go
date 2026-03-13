package models

import "time"

// Enterprise представляет предприятие (ОАО «Беларуськалий»)
type Enterprise struct {
	ID                 int64     `json:"id"`
	Name               string    `json:"name"`                          // Название предприятия
	Industry           string    `json:"industry"`                      // Отрасль
	AnnualProductionT  float64   `json:"annual_production_t"`           // Годовой объём производства (тонн)
	ExportSharePercent float64   `json:"export_share_percent"`          // Доля экспорта (%)
	MainCurrency       string    `json:"main_currency"`                 // Основная валюта (USD, BYN, EUR)
	IsExportOriented   bool      `json:"is_export_oriented"`            // Экспортно-ориентированное предприятие
	
	// Метаданные
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// IsSingleEnterpriseMode возвращает true, так как приложение для одного предприятия
func (e *Enterprise) IsSingleEnterpriseMode() bool {
	return true
}

// GetPrimaryEnterpriseID возвращает ID основного предприятия (ОАО «Беларуськалий»)
func GetPrimaryEnterpriseID() int64 {
	return 1 // ID предприятия "ОАО «Беларуськалий»" в БД
}
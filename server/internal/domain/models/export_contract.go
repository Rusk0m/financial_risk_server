package models

import "time"

// ExportContract представляет экспортный контракт
type ExportContract struct {
	ID              int64      `json:"id"`
	EnterpriseID    int64      `json:"enterprise_id"`                 // ID предприятия (всегда 1)
	ReportID        *int64     `json:"report_id,omitempty"`          // ID отчёта, из которого был распарсен контракт
	
	// Данные контракта
	ContractNumber   string  `json:"contract_number"`               // Номер контракта (например: KC-2026-001)
	ContractDate     time.Time `json:"contract_date"`                 // Дата заключения контракта
	Country          string  `json:"country"`                       // Страна импортёра
	VolumeT          float64 `json:"volume_t"`                      // Объём поставки в тоннах
	PriceUSDPerT     float64 `json:"price_usd_per_t"`               // Цена за тонну в USD
	Currency         string  `json:"currency"`                      // Валюта контракта (обычно USD)
	PaymentTermDays  int     `json:"payment_term_days"`             // Срок оплаты в днях
	ShipmentDate     time.Time `json:"shipment_date"`                 // Дата отгрузки
	PaymentStatus    string  `json:"payment_status"`                // Статус оплаты: pending, partial, paid, overdue
	ExchangeRate     float64 `json:"exchange_rate"`                 // Курс валюты на дату контракта
	
	// Дополнительные поля для анализа рисков
	TNVEDCode        *string `json:"tnved_code,omitempty"`          // Код ТН ВЭД
	Incoterms        *string `json:"incoterms,omitempty"`           // Условия поставки (FOB, CIF и т.д.)
	InsuranceCostUSD *float64 `json:"insurance_cost_usd,omitempty"` // Стоимость страховки в USD
	FreightCostUSD   *float64 `json:"freight_cost_usd,omitempty"`   // Стоимость фрахта в USD
	
	// Метаданные
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// GetContractValueUSD возвращает общую стоимость контракта в USD
func (ec *ExportContract) GetContractValueUSD() float64 {
	return ec.VolumeT * ec.PriceUSDPerT
}

// GetExpectedBYN возвращает ожидаемую выручку в BYN по текущему курсу
func (ec *ExportContract) GetExpectedBYN() float64 {
	return ec.GetContractValueUSD() * ec.ExchangeRate
}

// IsOverdue проверяет, просрочен ли контракт
func (ec *ExportContract) IsOverdue() bool {
	return ec.PaymentStatus == "overdue"
}

// DaysUntilPayment возвращает количество дней до оплаты
func (ec *ExportContract) DaysUntilPayment() int {
	paymentDate := ec.ContractDate.AddDate(0, 0, ec.PaymentTermDays)
	days := int(paymentDate.Sub(time.Now()).Hours() / 24)
	return days
}
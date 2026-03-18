package models

import (
	"strings"
	"time"
)

// ExportContract представляет экспортный контракт
type ExportContract struct {
	ID           int64     `json:"id"`
	EnterpriseID int64     `json:"enterprise_id"`
	ReportID     *int64    `json:"report_id,omitempty"`
	
	// === Ключевые поля для расчёта рисков ===
	
	// Для валютного и ценового риска
	ContractDate    time.Time `json:"contract_date"`      // Дата контракта
	PriceContract   float64   `json:"price_contract"`     // Цена контракта
	VolumeT         float64   `json:"volume_t"`           // Объём в тоннах
	Currency        string    `json:"currency"`           // Валюта (USD/EUR/CNY)
	ExchangeRate    float64   `json:"exchange_rate"`      // Курс на дату контракта
	
	// Для риска ликвидности
	PaymentTermDays int       `json:"payment_term_days"`  // Срок оплаты (дни)
	ShipmentDate    time.Time `json:"shipment_date"`      // Дата отгрузки
	PaymentStatus   string    `json:"payment_status"`     // pending/paid/overdue
	
	// Для странового и ценового риска
	Country         string    `json:"country"`            // Страна импортёра
	
	// === Метаданные ===
	ContractNumber  string    `json:"contract_number"`    // Только для отображения
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// GetContractValueUSD возвращает общую стоимость контракта в USD
func (ec *ExportContract) GetContractValueUSD() float64 {
	baseValue := ec.PriceContract // Это общая сумма


	
	// Если контракт в другой валюте, конвертируем
	switch strings.ToUpper(ec.Currency) {
	case "EUR":
		return baseValue * 1.08 // EUR → USD (курс обновляйте из market_data)
	case "CNY":
		return baseValue * 0.14 // CNY → USD
	case "BYN":
		return baseValue / 3.25 // BYN → USD
	default:
		return baseValue // Уже в USD
	}
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
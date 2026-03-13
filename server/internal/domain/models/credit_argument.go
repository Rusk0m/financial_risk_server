package models

import "time"

// CreditAgreement представляет кредитный договор
type CreditAgreement struct {
	ID                 int64      `json:"id"`
	EnterpriseID       int64      `json:"enterprise_id"`                 // ID предприятия (всегда 1)
	ReportID           *int64     `json:"report_id,omitempty"`          // ID отчёта, из которого был распарсен договор
	
	// Основные данные договора
	AgreementNumber    string     `json:"agreement_number"`              // Номер кредитного договора
	AgreementDate      time.Time  `json:"agreement_date"`                // Дата заключения договора
	CreditorName       string     `json:"creditor_name"`                 // Наименование кредитора (банк)
	CreditorCountry    *string    `json:"creditor_country,omitempty"`   // Страна кредитора
	
	// Условия кредита
	PrincipalAmount    float64    `json:"principal_amount"`              // Сумма кредита
	Currency           string     `json:"currency"`                      // Валюта кредита (USD, BYN, EUR)
	InterestRate       float64    `json:"interest_rate"`                 // Процентная ставка (% годовых)
	RateType           string     `json:"rate_type"`                     // Тип ставки: fixed, floating
	RateBase           *string    `json:"rate_base,omitempty"`          // Базовая ставка для плавающей ставки
	RateSpread         *float64   `json:"rate_spread,omitempty"`        // Спред над базовой ставкой
	
	// Сроки и график
	StartDate          time.Time  `json:"start_date"`                    // Дата начала действия кредита
	MaturityDate       time.Time  `json:"maturity_date"`                 // Дата погашения кредита
	TermMonths         int        `json:"term_months"`                   // Срок кредита в месяцах
	
	// Обеспечение
	CollateralType     *string    `json:"collateral_type,omitempty"`    // Тип обеспечения
	CollateralDescription *string `json:"collateral_description,omitempty"` // Описание обеспечения
	
	// Статус и метаданные
	Status             string     `json:"status"`                        // Статус: active, repaid, restructured, default
	OutstandingBalance *float64   `json:"outstanding_balance,omitempty"` // Остаток задолженности
	NextPaymentDate    *time.Time `json:"next_payment_date,omitempty"`  // Дата следующего платежа
	NextPaymentAmount  *float64   `json:"next_payment_amount,omitempty"` // Сумма следующего платежа
	
	// Метаданные
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

// GetAnnualInterestPayment возвращает годовой платёж по процентам
func (ca *CreditAgreement) GetAnnualInterestPayment() float64 {
	if ca.OutstandingBalance != nil {
		return *ca.OutstandingBalance * (ca.InterestRate / 100)
	}
	return ca.PrincipalAmount * (ca.InterestRate / 100)
}

// GetMonthlyPayment возвращает ежемесячный платёж (упрощённый расчёт)
func (ca *CreditAgreement) GetMonthlyPayment() float64 {
	// Упрощённый расчёт: аннуитетный платёж
	// В реальном приложении нужно использовать точную формулу или график платежей
	monthlyRate := ca.InterestRate / 100 / 12
	months := float64(ca.TermMonths)
	
	if monthlyRate == 0 {
		return ca.PrincipalAmount / months
	}
	
	payment := ca.PrincipalAmount * monthlyRate * (1 + monthlyRate) / ((1 + monthlyRate) - 1)
	return payment
}

// IsFloatingRate проверяет, является ли ставка плавающей
func (ca *CreditAgreement) IsFloatingRate() bool {
	return ca.RateType == "floating"
}

// GetDaysToMaturity возвращает количество дней до погашения
func (ca *CreditAgreement) GetDaysToMaturity() int {
	days := int(ca.MaturityDate.Sub(time.Now()).Hours() / 24)
	if days < 0 {
		return 0
	}
	return days
}

// IsDefaulted проверяет, находится ли договор в дефолте
func (ca *CreditAgreement) IsDefaulted() bool {
	return ca.Status == "default"
}
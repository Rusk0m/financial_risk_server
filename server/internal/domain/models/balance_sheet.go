package models

import "time"

// BalanceSheet представляет финансовый баланс (Форма №1)
type BalanceSheet struct {
	ID                     int64     `json:"id"`
	EnterpriseID           int64     `json:"enterprise_id"`                 // ID предприятия (всегда 1)
	ReportID               *int64    `json:"report_id,omitempty"`          // ID отчёта, из которого был распарсен баланс
	
	// Дата отчёта
	ReportDate             time.Time `json:"report_date"`                   // Дата составления отчёта
	
	// Активы (тыс. руб.)
	// Денежные средства и краткосрочные финансовые вложения
	CashBYN                float64   `json:"cash_byn"`                      // Денежные средства в BYN
	CashUSD                float64   `json:"cash_usd"`                      // Денежные средства в USD
	ShortTermInvestments   float64   `json:"short_term_investments"`        // Краткосрочные финансовые вложения
	
	// Дебиторская задолженность
	AccountsReceivable     float64   `json:"accounts_receivable"`           // Дебиторская задолженность
	VATReceivable          float64   `json:"vat_receivable"`                // НДС к получению
	
	// Запасы
	Inventories            float64   `json:"inventories"`                   // Запасы
	RawMaterials           float64   `json:"raw_materials"`                 // Сырьё и материалы
	WorkInProgress         float64   `json:"work_in_progress"`              // Незавершённое производство
	FinishedGoods          float64   `json:"finished_goods"`                // Готовая продукция
	
	// Долгосрочные активы
	PropertyPlantEquipment float64   `json:"property_plant_equipment"`      // Основные средства
	IntangibleAssets       float64   `json:"intangible_assets"`             // Нематериальные активы
	
	// Пассивы (тыс. руб.)
	// Краткосрочные обязательства
	AccountsPayable        float64   `json:"accounts_payable"`              // Кредиторская задолженность
	VATPayable             float64   `json:"vat_payable"`                   // НДС к уплате
	PayrollPayable         float64   `json:"payroll_payable"`               // Задолженность по зарплате
	ShortTermDebt          float64   `json:"short_term_debt"`               // Краткосрочные кредиты и займы
	
	// Долгосрочные обязательства
	LongTermDebt           float64   `json:"long_term_debt"`                // Долгосрочные кредиты и займы
	
	// Капитал
	AuthorizedCapital      float64   `json:"authorized_capital"`            // Уставный капитал
	RetainedEarnings       float64   `json:"retained_earnings"`             // Нераспределённая прибыль
	
	// Расчётные поля (хранятся в БД как GENERATED ALWAYS AS)
	TotalAssets            float64   `json:"total_assets"`                  // Итого активы
	TotalLiabilities       float64   `json:"total_liabilities"`             // Итого обязательства
	TotalEquity            float64   `json:"total_equity"`                  // Итого капитал
	
	// Метаданные
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
}

// GetCurrentRatio возвращает коэффициент текущей ликвидности
func (bs *BalanceSheet) GetCurrentRatio() float64 {
	currentAssets := bs.CashBYN + bs.CashUSD + bs.ShortTermInvestments +
		bs.AccountsReceivable + bs.VATReceivable + bs.Inventories
	currentLiabilities := bs.AccountsPayable + bs.VATPayable + bs.PayrollPayable + bs.ShortTermDebt
	
	if currentLiabilities == 0 {
		return 0
	}
	return currentAssets / currentLiabilities
}

// GetQuickRatio возвращает коэффициент быстрой ликвидности
func (bs *BalanceSheet) GetQuickRatio() float64 {
	quickAssets := bs.CashBYN + bs.CashUSD + bs.ShortTermInvestments + bs.AccountsReceivable
	currentLiabilities := bs.AccountsPayable + bs.VATPayable + bs.PayrollPayable + bs.ShortTermDebt
	
	if currentLiabilities == 0 {
		return 0
	}
	return quickAssets / currentLiabilities
}

// GetDebtToEquityRatio возвращает коэффициент финансового левериджа
func (bs *BalanceSheet) GetDebtToEquityRatio() float64 {
	totalDebt := bs.ShortTermDebt + bs.LongTermDebt
	totalEquity := bs.TotalEquity
	
	if totalEquity == 0 {
		return 0
	}
	return totalDebt / totalEquity
}

// GetWorkingCapital возвращает собственный оборотный капитал
func (bs *BalanceSheet) GetWorkingCapital() float64 {
	currentAssets := bs.CashBYN + bs.CashUSD + bs.ShortTermInvestments +
		bs.AccountsReceivable + bs.VATReceivable + bs.Inventories
	currentLiabilities := bs.AccountsPayable + bs.VATPayable + bs.PayrollPayable + bs.ShortTermDebt
	
	return currentAssets - currentLiabilities
}
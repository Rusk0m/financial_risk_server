package models

import (
	"time"
)

// BalanceSheet представляет финансовый баланс предприятия
type BalanceSheet struct {
	ID                 int64
	EnterpriseID       int64
	ReportDate         time.Time
	CashBYN            float64 // денежные средства в BYN
	CashUSD            float64 // денежные средства в USD
	AccountsReceivable float64 // дебиторская задолженность
	Inventories        float64 // запасы
	AccountsPayable    float64 // кредиторская задолженность
	ShortTermDebt      float64 // краткосрочные займы
	LongTermDebt       float64 // долгосрочные займы
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// GetTotalCurrentAssets возвращает сумму оборотных активов
func (b *BalanceSheet) GetTotalCurrentAssets() float64 {
	return b.CashBYN + b.CashUSD + b.AccountsReceivable + b.Inventories
}

// GetTotalCurrentLiabilities возвращает сумму краткосрочных обязательств
func (b *BalanceSheet) GetTotalCurrentLiabilities() float64 {
	return b.AccountsPayable + b.ShortTermDebt
}

// GetCurrentRatio рассчитывает коэффициент текущей ликвидности
func (b *BalanceSheet) GetCurrentRatio() float64 {
	liabilities := b.GetTotalCurrentLiabilities()
	if liabilities == 0 {
		return 0
	}
	return b.GetTotalCurrentAssets() / liabilities
}

// GetQuickRatio рассчитывает коэффициент быстрой ликвидности
func (b *BalanceSheet) GetQuickRatio() float64 {
	liabilities := b.GetTotalCurrentLiabilities()
	if liabilities == 0 {
		return 0
	}
	return (b.CashBYN + b.CashUSD + b.AccountsReceivable) / liabilities
}

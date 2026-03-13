package models

import "time"

// FinancialResult представляет отчёт о финансовых результатах (Форма №2)
type FinancialResult struct {
	ID                     int64     `json:"id"`
	EnterpriseID           int64     `json:"enterprise_id"`                 // ID предприятия (всегда 1)
	ReportID               *int64    `json:"report_id,omitempty"`          // ID отчёта, из которого был распарсен отчёт
	
	// Период отчёта
	ReportDate             time.Time `json:"report_date"`                   // Дата составления отчёта
	PeriodStart            time.Time `json:"period_start"`                  // Начало отчётного периода
	PeriodEnd              time.Time `json:"period_end"`                    // Конец отчётного периода
	
	// Выручка и доходы (тыс. руб.)
	RevenueSales           float64   `json:"revenue_sales"`                 // Выручка от продажи продукции
	RevenueExport          float64   `json:"revenue_export"`                // Экспортная выручка
	RevenueDomestic        float64   `json:"revenue_domestic"`              // Внутренняя выручка
	RevenueOther           float64   `json:"revenue_other"`                 // Прочие доходы
	RevenueTotal           float64   `json:"revenue_total"`                 // Итого выручка
	
	// Себестоимость (тыс. руб.)
	CostOfSales            float64   `json:"cost_of_sales"`                 // Себестоимость проданных товаров
	CostRawMaterials       float64   `json:"cost_raw_materials"`            // Сырьё и материалы
	CostEnergy             float64   `json:"cost_energy"`                   // Энергия
	CostLabor              float64   `json:"cost_labor"`                    // Заработная плата
	CostDepreciation       float64   `json:"cost_depreciation"`             // Амортизация
	CostOther              float64   `json:"cost_other"`                    // Прочие затраты
	CostTotal              float64   `json:"cost_total"`                    // Итого себестоимость
	
	// Расходы (тыс. руб.)
	CommercialExpenses     float64   `json:"commercial_expenses"`           // Коммерческие расходы
	AdministrativeExpenses float64   `json:"administrative_expenses"`       // Управленческие расходы
	OtherExpenses          float64   `json:"other_expenses"`                // Прочие расходы
	
	// Прибыль/убыток (тыс. руб.)
	GrossProfit            float64   `json:"gross_profit"`                  // Валовая прибыль
	OperatingProfit        float64   `json:"operating_profit"`              // Прибыль от операционной деятельности
	ProfitBeforeTax        float64   `json:"profit_before_tax"`             // Прибыль до налогообложения
	TaxExpense             float64   `json:"tax_expense"`                   // Налог на прибыль
	NetProfit              float64   `json:"net_profit"`                    // Чистая прибыль
	
	// Дополнительные поля для анализа рисков
	EBITDA                 *float64  `json:"ebitda,omitempty"`              // EBITDA
	OperatingMargin        *float64  `json:"operating_margin,omitempty"`    // Операционная маржа (%)
	NetMargin              *float64  `json:"net_margin,omitempty"`          // Чистая маржа (%)
	
	// Метаданные
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
}

// GetCostStructure возвращает структуру затрат в процентах от себестоимости
func (fr *FinancialResult) GetCostStructure() map[string]float64 {
	if fr.CostTotal == 0 {
		return map[string]float64{}
	}
	
	return map[string]float64{
		"Сырьё и материалы": (fr.CostRawMaterials / fr.CostTotal) * 100,
		"Энергия":           (fr.CostEnergy / fr.CostTotal) * 100,
		"Заработная плата":  (fr.CostLabor / fr.CostTotal) * 100,
		"Амортизация":       (fr.CostDepreciation / fr.CostTotal) * 100,
		"Прочие затраты":    (fr.CostOther / fr.CostTotal) * 100,
	}
}

// GetRevenueStructure возвращает структуру выручки в процентах
func (fr *FinancialResult) GetRevenueStructure() map[string]float64 {
	if fr.RevenueTotal == 0 {
		return map[string]float64{}
	}
	
	return map[string]float64{
		"Экспорт":    (fr.RevenueExport / fr.RevenueTotal) * 100,
		"Внутренний": (fr.RevenueDomestic / fr.RevenueTotal) * 100,
		"Прочие":     (fr.RevenueOther / fr.RevenueTotal) * 100,
	}
}

// GetProfitabilityRatios возвращает коэффициенты рентабельности
func (fr *FinancialResult) GetProfitabilityRatios() map[string]float64 {
	ratios := make(map[string]float64)
	
	if fr.RevenueTotal > 0 {
		ratios["Рентабельность продаж"] = (fr.NetProfit / fr.RevenueTotal) * 100
		ratios["Валовая рентабельность"] = (fr.GrossProfit / fr.RevenueTotal) * 100
		ratios["Операционная рентабельность"] = (fr.OperatingProfit / fr.RevenueTotal) * 100
	}
	
	if fr.OperatingProfit > 0 {
		ratios["Рентабельность операционной деятельности"] = (fr.NetProfit / fr.OperatingProfit) * 100
	}
	
	return ratios
}
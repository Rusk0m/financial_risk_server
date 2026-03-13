package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"financial-risk-server/internal/domain/models"
	"financial-risk-server/internal/repository/interfaces"
)

// FinancialResultRepository реализует интерфейс для работы с отчётами о финансовых результатах
type FinancialResultRepository struct {
	db *sql.DB
}

// NewFinancialResultRepository создаёт новый репозиторий отчётов о ФР
func NewFinancialResultRepository(db *sql.DB) interfaces.FinancialResultRepository {
	return &FinancialResultRepository{db: db}
}

// Create создаёт новый отчёт о ФР
func (r *FinancialResultRepository) Create(ctx context.Context, result *models.FinancialResult) error {
	query := `
		INSERT INTO financial_results (
			enterprise_id, report_id, report_date, period_start, period_end,
			revenue_sales, revenue_export, revenue_domestic, revenue_other, revenue_total,
			cost_of_sales, cost_raw_materials, cost_energy, cost_labor, cost_depreciation, cost_other, cost_total,
			commercial_expenses, administrative_expenses, other_expenses,
			gross_profit, operating_profit, profit_before_tax, tax_expense, net_profit,
			ebitda, operating_margin, net_margin
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28)
		RETURNING id, created_at, updated_at
	`

	var createdAt, updatedAt time.Time
	err := r.db.QueryRowContext(ctx, query,
		result.EnterpriseID,
		result.ReportID,
		result.ReportDate,
		result.PeriodStart,
		result.PeriodEnd,
		result.RevenueSales,
		result.RevenueExport,
		result.RevenueDomestic,
		result.RevenueOther,
		result.RevenueTotal,
		result.CostOfSales,
		result.CostRawMaterials,
		result.CostEnergy,
		result.CostLabor,
		result.CostDepreciation,
		result.CostOther,
		result.CostTotal,
		result.CommercialExpenses,
		result.AdministrativeExpenses,
		result.OtherExpenses,
		result.GrossProfit,
		result.OperatingProfit,
		result.ProfitBeforeTax,
		result.TaxExpense,
		result.NetProfit,
		result.EBITDA,
		result.OperatingMargin,
		result.NetMargin,
	).Scan(&result.ID, &createdAt, &updatedAt)

	if err != nil {
		return fmt.Errorf("failed to create financial result: %w", err)
	}

	result.CreatedAt = createdAt
	result.UpdatedAt = updatedAt

	return nil
}

// GetByID получает отчёт по ID
func (r *FinancialResultRepository) GetByID(ctx context.Context, id int64) (*models.FinancialResult, error) {
	query := `
		SELECT 
			id, enterprise_id, report_id, report_date, period_start, period_end,
			revenue_sales, revenue_export, revenue_domestic, revenue_other, revenue_total,
			cost_of_sales, cost_raw_materials, cost_energy, cost_labor, cost_depreciation, cost_other, cost_total,
			commercial_expenses, administrative_expenses, other_expenses,
			gross_profit, operating_profit, profit_before_tax, tax_expense, net_profit,
			ebitda, operating_margin, net_margin,
			created_at, updated_at
		FROM financial_results
		WHERE id = $1
	`

	result := &models.FinancialResult{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&result.ID,
		&result.EnterpriseID,
		&result.ReportID,
		&result.ReportDate,
		&result.PeriodStart,
		&result.PeriodEnd,
		&result.RevenueSales,
		&result.RevenueExport,
		&result.RevenueDomestic,
		&result.RevenueOther,
		&result.RevenueTotal,
		&result.CostOfSales,
		&result.CostRawMaterials,
		&result.CostEnergy,
		&result.CostLabor,
		&result.CostDepreciation,
		&result.CostOther,
		&result.CostTotal,
		&result.CommercialExpenses,
		&result.AdministrativeExpenses,
		&result.OtherExpenses,
		&result.GrossProfit,
		&result.OperatingProfit,
		&result.ProfitBeforeTax,
		&result.TaxExpense,
		&result.NetProfit,
		&result.EBITDA,
		&result.OperatingMargin,
		&result.NetMargin,
		&result.CreatedAt,
		&result.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("financial result not found with id %d", id)
		}
		return nil, fmt.Errorf("failed to get financial result: %w", err)
	}

	return result, nil
}

// GetByEnterpriseID получает отчёты предприятия
func (r *FinancialResultRepository) GetByEnterpriseID(ctx context.Context, enterpriseID int64) ([]*models.FinancialResult, error) {
	query := `
		SELECT 
			id, enterprise_id, report_id, report_date, period_start, period_end,
			revenue_sales, revenue_export, revenue_domestic, revenue_other, revenue_total,
			cost_of_sales, cost_raw_materials, cost_energy, cost_labor, cost_depreciation, cost_other, cost_total,
			commercial_expenses, administrative_expenses, other_expenses,
			gross_profit, operating_profit, profit_before_tax, tax_expense, net_profit,
			ebitda, operating_margin, net_margin,
			created_at, updated_at
		FROM financial_results
		WHERE enterprise_id = $1
		ORDER BY report_date DESC
	`

	rows, err := r.db.QueryContext(ctx, query, enterpriseID)
	if err != nil {
		return nil, fmt.Errorf("failed to query financial results: %w", err)
	}
	defer rows.Close()

	results := []*models.FinancialResult{}
	for rows.Next() {
		result := &models.FinancialResult{}
		err := rows.Scan(
			&result.ID,
			&result.EnterpriseID,
			&result.ReportID,
			&result.ReportDate,
			&result.PeriodStart,
			&result.PeriodEnd,
			&result.RevenueSales,
			&result.RevenueExport,
			&result.RevenueDomestic,
			&result.RevenueOther,
			&result.RevenueTotal,
			&result.CostOfSales,
			&result.CostRawMaterials,
			&result.CostEnergy,
			&result.CostLabor,
			&result.CostDepreciation,
			&result.CostOther,
			&result.CostTotal,
			&result.CommercialExpenses,
			&result.AdministrativeExpenses,
			&result.OtherExpenses,
			&result.GrossProfit,
			&result.OperatingProfit,
			&result.ProfitBeforeTax,
			&result.TaxExpense,
			&result.NetProfit,
			&result.EBITDA,
			&result.OperatingMargin,
			&result.NetMargin,
			&result.CreatedAt,
			&result.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan financial result: %w", err)
		}
		results = append(results, result)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return results, nil
}

// GetByReportID получает отчёт из отчёта
func (r *FinancialResultRepository) GetByReportID(ctx context.Context, reportID int64) (*models.FinancialResult, error) {
	query := `
		SELECT 
			id, enterprise_id, report_id, report_date, period_start, period_end,
			revenue_sales, revenue_export, revenue_domestic, revenue_other, revenue_total,
			cost_of_sales, cost_raw_materials, cost_energy, cost_labor, cost_depreciation, cost_other, cost_total,
			commercial_expenses, administrative_expenses, other_expenses,
			gross_profit, operating_profit, profit_before_tax, tax_expense, net_profit,
			ebitda, operating_margin, net_margin,
			created_at, updated_at
		FROM financial_results
		WHERE report_id = $1
		ORDER BY report_date DESC
		LIMIT 1
	`

	result := &models.FinancialResult{}
	err := r.db.QueryRowContext(ctx, query, reportID).Scan(
		&result.ID,
		&result.EnterpriseID,
		&result.ReportID,
		&result.ReportDate,
		&result.PeriodStart,
		&result.PeriodEnd,
		&result.RevenueSales,
		&result.RevenueExport,
		&result.RevenueDomestic,
		&result.RevenueOther,
		&result.RevenueTotal,
		&result.CostOfSales,
		&result.CostRawMaterials,
		&result.CostEnergy,
		&result.CostLabor,
		&result.CostDepreciation,
		&result.CostOther,
		&result.CostTotal,
		&result.CommercialExpenses,
		&result.AdministrativeExpenses,
		&result.OtherExpenses,
		&result.GrossProfit,
		&result.OperatingProfit,
		&result.ProfitBeforeTax,
		&result.TaxExpense,
		&result.NetProfit,
		&result.EBITDA,
		&result.OperatingMargin,
		&result.NetMargin,
		&result.CreatedAt,
		&result.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("financial result not found for report id %d", reportID)
		}
		return nil, fmt.Errorf("failed to get financial result: %w", err)
	}

	return result, nil
}

// GetAll получает все отчёты
func (r *FinancialResultRepository) GetAll(ctx context.Context) ([]*models.FinancialResult, error) {
	query := `
		SELECT 
			id, enterprise_id, report_id, report_date, period_start, period_end,
			revenue_sales, revenue_export, revenue_domestic, revenue_other, revenue_total,
			cost_of_sales, cost_raw_materials, cost_energy, cost_labor, cost_depreciation, cost_other, cost_total,
			commercial_expenses, administrative_expenses, other_expenses,
			gross_profit, operating_profit, profit_before_tax, tax_expense, net_profit,
			ebitda, operating_margin, net_margin,
			created_at, updated_at
		FROM financial_results
		ORDER BY report_date DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query financial results: %w", err)
	}
	defer rows.Close()

	results := []*models.FinancialResult{}
	for rows.Next() {
		result := &models.FinancialResult{}
		err := rows.Scan(
			&result.ID,
			&result.EnterpriseID,
			&result.ReportID,
			&result.ReportDate,
			&result.PeriodStart,
			&result.PeriodEnd,
			&result.RevenueSales,
			&result.RevenueExport,
			&result.RevenueDomestic,
			&result.RevenueOther,
			&result.RevenueTotal,
			&result.CostOfSales,
			&result.CostRawMaterials,
			&result.CostEnergy,
			&result.CostLabor,
			&result.CostDepreciation,
			&result.CostOther,
			&result.CostTotal,
			&result.CommercialExpenses,
			&result.AdministrativeExpenses,
			&result.OtherExpenses,
			&result.GrossProfit,
			&result.OperatingProfit,
			&result.ProfitBeforeTax,
			&result.TaxExpense,
			&result.NetProfit,
			&result.EBITDA,
			&result.OperatingMargin,
			&result.NetMargin,
			&result.CreatedAt,
			&result.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan financial result: %w", err)
		}
		results = append(results, result)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return results, nil
}

// Update обновляет отчёт
// Update обновляет отчёт
func (r *FinancialResultRepository) Update(ctx context.Context, result *models.FinancialResult) error {
	query := `
		UPDATE financial_results
		SET report_date = $1, period_start = $2, period_end = $3,
		    revenue_sales = $4, revenue_export = $5, revenue_domestic = $6, revenue_other = $7, revenue_total = $8,
		    cost_of_sales = $9, cost_raw_materials = $10, cost_energy = $11, cost_labor = $12, cost_depreciation = $13, cost_other = $14, cost_total = $15,
		    commercial_expenses = $16, administrative_expenses = $17, other_expenses = $18,
		    gross_profit = $19, operating_profit = $20, profit_before_tax = $21, tax_expense = $22, net_profit = $23,
		    ebitda = $24, operating_margin = $25, net_margin = $26,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $27
	`

	// ✅ Исправлено: используем r.db и корректно обрабатываем результат
	res, err := r.db.ExecContext(ctx, query,
		result.ReportDate,
		result.PeriodStart,
		result.PeriodEnd,
		result.RevenueSales,
		result.RevenueExport,
		result.RevenueDomestic,
		result.RevenueOther,
		result.RevenueTotal,
		result.CostOfSales,
		result.CostRawMaterials,
		result.CostEnergy,
		result.CostLabor,
		result.CostDepreciation,
		result.CostOther,
		result.CostTotal,
		result.CommercialExpenses,
		result.AdministrativeExpenses,
		result.OtherExpenses,
		result.GrossProfit,
		result.OperatingProfit,
		result.ProfitBeforeTax,
		result.TaxExpense,
		result.NetProfit,
		result.EBITDA,
		result.OperatingMargin,
		result.NetMargin,
		result.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update financial result: %w", err)
	}

	rowsAffected, err := res.RowsAffected() // ✅ Исправлено: вызываем у sql.Result
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("financial result not found with id %d", result.ID)
	}

	// ✅ Опционально: обновляем UpdatedAt в объекте модели
	result.UpdatedAt = time.Now()

	return nil
}
// Delete удаляет отчёт по ID
func (r *FinancialResultRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM financial_results WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete financial result: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("financial result not found with id %d", id)
	}

	return nil
}
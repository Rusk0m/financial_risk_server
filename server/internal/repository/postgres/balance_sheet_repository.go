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

// BalanceSheetRepository реализует интерфейс для работы с финансовыми балансами
type BalanceSheetRepository struct {
	db *sql.DB
}

// NewBalanceSheetRepository создаёт новый репозиторий финансовых балансов
func NewBalanceSheetRepository(db *sql.DB) interfaces.BalanceSheetRepository {
	return &BalanceSheetRepository{db: db}
}

// Create создаёт новый баланс
func (r *BalanceSheetRepository) Create(ctx context.Context, balance *models.BalanceSheet) error {
	query := `
		INSERT INTO balance_sheets (
			enterprise_id, report_id, report_date,
			cash_byn, cash_usd, short_term_investments,
			accounts_receivable, vat_receivable,
			inventories, raw_materials, work_in_progress, finished_goods,
			property_plant_equipment, intangible_assets,
			accounts_payable, vat_payable, payroll_payable, short_term_debt,
			long_term_debt,
			authorized_capital, retained_earnings
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)
		RETURNING id, created_at, updated_at
	`

	var createdAt, updatedAt time.Time
	err := r.db.QueryRowContext(ctx, query,
		balance.EnterpriseID,
		balance.ReportID,
		balance.ReportDate,
		balance.CashBYN,
		balance.CashUSD,
		balance.ShortTermInvestments,
		balance.AccountsReceivable,
		balance.VATReceivable,
		balance.Inventories,
		balance.RawMaterials,
		balance.WorkInProgress,
		balance.FinishedGoods,
		balance.PropertyPlantEquipment,
		balance.IntangibleAssets,
		balance.AccountsPayable,
		balance.VATPayable,
		balance.PayrollPayable,
		balance.ShortTermDebt,
		balance.LongTermDebt,
		balance.AuthorizedCapital,
		balance.RetainedEarnings,
	).Scan(&balance.ID, &createdAt, &updatedAt)

	if err != nil {
		return fmt.Errorf("failed to create balance sheet: %w", err)
	}

	balance.CreatedAt = createdAt
	balance.UpdatedAt = updatedAt

	return nil
}

// GetByID получает баланс по ID
func (r *BalanceSheetRepository) GetByID(ctx context.Context, id int64) (*models.BalanceSheet, error) {
	query := `
		SELECT 
			id, enterprise_id, report_id, report_date,
			cash_byn, cash_usd, short_term_investments,
			accounts_receivable, vat_receivable,
			inventories, raw_materials, work_in_progress, finished_goods,
			property_plant_equipment, intangible_assets,
			accounts_payable, vat_payable, payroll_payable, short_term_debt,
			long_term_debt,
			authorized_capital, retained_earnings,
			total_assets, total_liabilities, total_equity,
			created_at, updated_at
		FROM balance_sheets
		WHERE id = $1
	`

	balance := &models.BalanceSheet{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&balance.ID,
		&balance.EnterpriseID,
		&balance.ReportID,
		&balance.ReportDate,
		&balance.CashBYN,
		&balance.CashUSD,
		&balance.ShortTermInvestments,
		&balance.AccountsReceivable,
		&balance.VATReceivable,
		&balance.Inventories,
		&balance.RawMaterials,
		&balance.WorkInProgress,
		&balance.FinishedGoods,
		&balance.PropertyPlantEquipment,
		&balance.IntangibleAssets,
		&balance.AccountsPayable,
		&balance.VATPayable,
		&balance.PayrollPayable,
		&balance.ShortTermDebt,
		&balance.LongTermDebt,
		&balance.AuthorizedCapital,
		&balance.RetainedEarnings,
		&balance.TotalAssets,
		&balance.TotalLiabilities,
		&balance.TotalEquity,
		&balance.CreatedAt,
		&balance.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("balance sheet not found with id %d", id)
		}
		return nil, fmt.Errorf("failed to get balance sheet: %w", err)
	}

	return balance, nil
}

// GetByEnterpriseID получает балансы предприятия
func (r *BalanceSheetRepository) GetByEnterpriseID(ctx context.Context, enterpriseID int64) ([]*models.BalanceSheet, error) {
	query := `
		SELECT 
			id, enterprise_id, report_id, report_date,
			cash_byn, cash_usd, short_term_investments,
			accounts_receivable, vat_receivable,
			inventories, raw_materials, work_in_progress, finished_goods,
			property_plant_equipment, intangible_assets,
			accounts_payable, vat_payable, payroll_payable, short_term_debt,
			long_term_debt,
			authorized_capital, retained_earnings,
			total_assets, total_liabilities, total_equity,
			created_at, updated_at
		FROM balance_sheets
		WHERE enterprise_id = $1
		ORDER BY report_date DESC
	`

	rows, err := r.db.QueryContext(ctx, query, enterpriseID)
	if err != nil {
		return nil, fmt.Errorf("failed to query balance sheets: %w", err)
	}
	defer rows.Close()

	balances := []*models.BalanceSheet{}
	for rows.Next() {
		balance := &models.BalanceSheet{}
		err := rows.Scan(
			&balance.ID,
			&balance.EnterpriseID,
			&balance.ReportID,
			&balance.ReportDate,
			&balance.CashBYN,
			&balance.CashUSD,
			&balance.ShortTermInvestments,
			&balance.AccountsReceivable,
			&balance.VATReceivable,
			&balance.Inventories,
			&balance.RawMaterials,
			&balance.WorkInProgress,
			&balance.FinishedGoods,
			&balance.PropertyPlantEquipment,
			&balance.IntangibleAssets,
			&balance.AccountsPayable,
			&balance.VATPayable,
			&balance.PayrollPayable,
			&balance.ShortTermDebt,
			&balance.LongTermDebt,
			&balance.AuthorizedCapital,
			&balance.RetainedEarnings,
			&balance.TotalAssets,
			&balance.TotalLiabilities,
			&balance.TotalEquity,
			&balance.CreatedAt,
			&balance.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan balance sheet: %w", err)
		}
		balances = append(balances, balance)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return balances, nil
}

// GetByReportID получает баланс из отчёта
func (r *BalanceSheetRepository) GetByReportID(ctx context.Context, reportID int64) (*models.BalanceSheet, error) {
	query := `
		SELECT 
			id, enterprise_id, report_id, report_date,
			cash_byn, cash_usd, short_term_investments,
			accounts_receivable, vat_receivable,
			inventories, raw_materials, work_in_progress, finished_goods,
			property_plant_equipment, intangible_assets,
			accounts_payable, vat_payable, payroll_payable, short_term_debt,
			long_term_debt,
			authorized_capital, retained_earnings,
			total_assets, total_liabilities, total_equity,
			created_at, updated_at
		FROM balance_sheets
		WHERE report_id = $1
		ORDER BY report_date DESC
		LIMIT 1
	`

	balance := &models.BalanceSheet{}
	err := r.db.QueryRowContext(ctx, query, reportID).Scan(
		&balance.ID,
		&balance.EnterpriseID,
		&balance.ReportID,
		&balance.ReportDate,
		&balance.CashBYN,
		&balance.CashUSD,
		&balance.ShortTermInvestments,
		&balance.AccountsReceivable,
		&balance.VATReceivable,
		&balance.Inventories,
		&balance.RawMaterials,
		&balance.WorkInProgress,
		&balance.FinishedGoods,
		&balance.PropertyPlantEquipment,
		&balance.IntangibleAssets,
		&balance.AccountsPayable,
		&balance.VATPayable,
		&balance.PayrollPayable,
		&balance.ShortTermDebt,
		&balance.LongTermDebt,
		&balance.AuthorizedCapital,
		&balance.RetainedEarnings,
		&balance.TotalAssets,
		&balance.TotalLiabilities,
		&balance.TotalEquity,
		&balance.CreatedAt,
		&balance.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("balance sheet not found for report id %d", reportID)
		}
		return nil, fmt.Errorf("failed to get balance sheet: %w", err)
	}

	return balance, nil
}

// GetLatest получает последний баланс предприятия
func (r *BalanceSheetRepository) GetLatest(ctx context.Context, enterpriseID int64) (*models.BalanceSheet, error) {
	query := `
		SELECT 
			id, enterprise_id, report_id, report_date,
			cash_byn, cash_usd, short_term_investments,
			accounts_receivable, vat_receivable,
			inventories, raw_materials, work_in_progress, finished_goods,
			property_plant_equipment, intangible_assets,
			accounts_payable, vat_payable, payroll_payable, short_term_debt,
			long_term_debt,
			authorized_capital, retained_earnings,
			total_assets, total_liabilities, total_equity,
			created_at, updated_at
		FROM balance_sheets
		WHERE enterprise_id = $1
		ORDER BY report_date DESC
		LIMIT 1
	`

	balance := &models.BalanceSheet{}
	err := r.db.QueryRowContext(ctx, query, enterpriseID).Scan(
		&balance.ID,
		&balance.EnterpriseID,
		&balance.ReportID,
		&balance.ReportDate,
		&balance.CashBYN,
		&balance.CashUSD,
		&balance.ShortTermInvestments,
		&balance.AccountsReceivable,
		&balance.VATReceivable,
		&balance.Inventories,
		&balance.RawMaterials,
		&balance.WorkInProgress,
		&balance.FinishedGoods,
		&balance.PropertyPlantEquipment,
		&balance.IntangibleAssets,
		&balance.AccountsPayable,
		&balance.VATPayable,
		&balance.PayrollPayable,
		&balance.ShortTermDebt,
		&balance.LongTermDebt,
		&balance.AuthorizedCapital,
		&balance.RetainedEarnings,
		&balance.TotalAssets,
		&balance.TotalLiabilities,
		&balance.TotalEquity,
		&balance.CreatedAt,
		&balance.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("no balance sheets found for enterprise id %d", enterpriseID)
		}
		return nil, fmt.Errorf("failed to get latest balance sheet: %w", err)
	}

	return balance, nil
}

// GetAll получает все балансы
func (r *BalanceSheetRepository) GetAll(ctx context.Context) ([]*models.BalanceSheet, error) {
	query := `
		SELECT 
			id, enterprise_id, report_id, report_date,
			cash_byn, cash_usd, short_term_investments,
			accounts_receivable, vat_receivable,
			inventories, raw_materials, work_in_progress, finished_goods,
			property_plant_equipment, intangible_assets,
			accounts_payable, vat_payable, payroll_payable, short_term_debt,
			long_term_debt,
			authorized_capital, retained_earnings,
			total_assets, total_liabilities, total_equity,
			created_at, updated_at
		FROM balance_sheets
		ORDER BY report_date DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query balance sheets: %w", err)
	}
	defer rows.Close()

	balances := []*models.BalanceSheet{}
	for rows.Next() {
		balance := &models.BalanceSheet{}
		err := rows.Scan(
			&balance.ID,
			&balance.EnterpriseID,
			&balance.ReportID,
			&balance.ReportDate,
			&balance.CashBYN,
			&balance.CashUSD,
			&balance.ShortTermInvestments,
			&balance.AccountsReceivable,
			&balance.VATReceivable,
			&balance.Inventories,
			&balance.RawMaterials,
			&balance.WorkInProgress,
			&balance.FinishedGoods,
			&balance.PropertyPlantEquipment,
			&balance.IntangibleAssets,
			&balance.AccountsPayable,
			&balance.VATPayable,
			&balance.PayrollPayable,
			&balance.ShortTermDebt,
			&balance.LongTermDebt,
			&balance.AuthorizedCapital,
			&balance.RetainedEarnings,
			&balance.TotalAssets,
			&balance.TotalLiabilities,
			&balance.TotalEquity,
			&balance.CreatedAt,
			&balance.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan balance sheet: %w", err)
		}
		balances = append(balances, balance)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return balances, nil
}

// Update обновляет баланс
func (r *BalanceSheetRepository) Update(ctx context.Context, balance *models.BalanceSheet) error {
	query := `
		UPDATE balance_sheets
		SET report_date = $1,
		    cash_byn = $2, cash_usd = $3, short_term_investments = $4,
		    accounts_receivable = $5, vat_receivable = $6,
		    inventories = $7, raw_materials = $8, work_in_progress = $9, finished_goods = $10,
		    property_plant_equipment = $11, intangible_assets = $12,
		    accounts_payable = $13, vat_payable = $14, payroll_payable = $15, short_term_debt = $16,
		    long_term_debt = $17,
		    authorized_capital = $18, retained_earnings = $19,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $20
	`

	result, err := r.db.ExecContext(ctx, query,
		balance.ReportDate,
		balance.CashBYN,
		balance.CashUSD,
		balance.ShortTermInvestments,
		balance.AccountsReceivable,
		balance.VATReceivable,
		balance.Inventories,
		balance.RawMaterials,
		balance.WorkInProgress,
		balance.FinishedGoods,
		balance.PropertyPlantEquipment,
		balance.IntangibleAssets,
		balance.AccountsPayable,
		balance.VATPayable,
		balance.PayrollPayable,
		balance.ShortTermDebt,
		balance.LongTermDebt,
		balance.AuthorizedCapital,
		balance.RetainedEarnings,
		balance.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update balance sheet: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("balance sheet not found with id %d", balance.ID)
	}

	return nil
}

// Delete удаляет баланс по ID
func (r *BalanceSheetRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM balance_sheets WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete balance sheet: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("balance sheet not found with id %d", id)
	}

	return nil
}
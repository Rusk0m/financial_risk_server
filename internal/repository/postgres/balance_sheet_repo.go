package postgres

import (
	"database/sql"
	"errors"
	"financial-risk-server/internal/domain/models"
	"financial-risk-server/internal/repository/interfaces"
	"time"
)

type balanceSheetRepository struct {
	db *DB
}

// NewBalanceSheetRepository создаёт новый репозиторий финансовых балансов
func NewBalanceSheetRepository(db *DB) interfaces.BalanceSheetRepository {
	return &balanceSheetRepository{db: db}
}

// Create сохраняет финансовый баланс в базу данных
func (r *balanceSheetRepository) Create(balance *models.BalanceSheet) error {
	query := `
		INSERT INTO balance_sheets (
			enterprise_id, report_date, cash_byn, cash_usd,
			accounts_receivable, inventories, accounts_payable,
			short_term_debt, long_term_debt, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id
	`

	now := time.Now()
	err := r.db.QueryRow(query,
		balance.EnterpriseID,
		balance.ReportDate,
		balance.CashBYN,
		balance.CashUSD,
		balance.AccountsReceivable,
		balance.Inventories,
		balance.AccountsPayable,
		balance.ShortTermDebt,
		balance.LongTermDebt,
		now,
		now,
	).Scan(&balance.ID)

	if err != nil {
		return err
	}

	balance.CreatedAt = now
	balance.UpdatedAt = now
	return nil
}

// GetByID получает финансовый баланс по идентификатору
func (r *balanceSheetRepository) GetByID(id int64) (*models.BalanceSheet, error) {
	query := `
		SELECT id, enterprise_id, report_date, cash_byn, cash_usd,
		       accounts_receivable, inventories, accounts_payable,
		       short_term_debt, long_term_debt, created_at, updated_at
		FROM balance_sheets
		WHERE id = $1
	`

	balance := &models.BalanceSheet{}
	err := r.db.QueryRow(query, id).Scan(
		&balance.ID,
		&balance.EnterpriseID,
		&balance.ReportDate,
		&balance.CashBYN,
		&balance.CashUSD,
		&balance.AccountsReceivable,
		&balance.Inventories,
		&balance.AccountsPayable,
		&balance.ShortTermDebt,
		&balance.LongTermDebt,
		&balance.CreatedAt,
		&balance.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("финансовый баланс не найден")
		}
		return nil, err
	}

	return balance, nil
}

// GetByEnterpriseID получает все финансовые балансы для предприятия
func (r *balanceSheetRepository) GetByEnterpriseID(enterpriseID int64) ([]*models.BalanceSheet, error) {
	query := `
		SELECT id, enterprise_id, report_date, cash_byn, cash_usd,
		       accounts_receivable, inventories, accounts_payable,
		       short_term_debt, long_term_debt, created_at, updated_at
		FROM balance_sheets
		WHERE enterprise_id = $1
		ORDER BY report_date DESC
	`

	rows, err := r.db.Query(query, enterpriseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var balances []*models.BalanceSheet
	for rows.Next() {
		balance := &models.BalanceSheet{}
		err := rows.Scan(
			&balance.ID,
			&balance.EnterpriseID,
			&balance.ReportDate,
			&balance.CashBYN,
			&balance.CashUSD,
			&balance.AccountsReceivable,
			&balance.Inventories,
			&balance.AccountsPayable,
			&balance.ShortTermDebt,
			&balance.LongTermDebt,
			&balance.CreatedAt,
			&balance.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		balances = append(balances, balance)
	}

	return balances, rows.Err()
}

// GetLatest получает последний финансовый баланс для предприятия
func (r *balanceSheetRepository) GetLatest(enterpriseID int64) (*models.BalanceSheet, error) {
	query := `
		SELECT id, enterprise_id, report_date, cash_byn, cash_usd,
		       accounts_receivable, inventories, accounts_payable,
		       short_term_debt, long_term_debt, created_at, updated_at
		FROM balance_sheets
		WHERE enterprise_id = $1
		ORDER BY report_date DESC
		LIMIT 1
	`

	balance := &models.BalanceSheet{}
	err := r.db.QueryRow(query, enterpriseID).Scan(
		&balance.ID,
		&balance.EnterpriseID,
		&balance.ReportDate,
		&balance.CashBYN,
		&balance.CashUSD,
		&balance.AccountsReceivable,
		&balance.Inventories,
		&balance.AccountsPayable,
		&balance.ShortTermDebt,
		&balance.LongTermDebt,
		&balance.CreatedAt,
		&balance.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("финансовые балансы не найдены")
		}
		return nil, err
	}

	return balance, nil
}

// Update обновляет финансовый баланс
func (r *balanceSheetRepository) Update(balance *models.BalanceSheet) error {
	query := `
		UPDATE balance_sheets
		SET enterprise_id = $1, report_date = $2, cash_byn = $3,
		    cash_usd = $4, accounts_receivable = $5, inventories = $6,
		    accounts_payable = $7, short_term_debt = $8, long_term_debt = $9,
		    updated_at = $10
		WHERE id = $11
	`

	now := time.Now()
	result, err := r.db.Exec(query,
		balance.EnterpriseID,
		balance.ReportDate,
		balance.CashBYN,
		balance.CashUSD,
		balance.AccountsReceivable,
		balance.Inventories,
		balance.AccountsPayable,
		balance.ShortTermDebt,
		balance.LongTermDebt,
		now,
		balance.ID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("финансовый баланс не найден")
	}

	balance.UpdatedAt = now
	return nil
}

// Delete удаляет финансовый баланс
func (r *balanceSheetRepository) Delete(id int64) error {
	query := `DELETE FROM balance_sheets WHERE id = $1`
	result, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("финансовый баланс не найден")
	}

	return nil
}

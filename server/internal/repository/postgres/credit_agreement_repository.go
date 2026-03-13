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

// CreditAgreementRepository реализует интерфейс для работы с кредитными договорами
type CreditAgreementRepository struct {
	db *sql.DB
}

// NewCreditAgreementRepository создаёт новый репозиторий кредитных договоров
func NewCreditAgreementRepository(db *sql.DB) interfaces.CreditAgreementRepository {
	return &CreditAgreementRepository{db: db}
}

// Create создаёт новый кредитный договор
func (r *CreditAgreementRepository) Create(ctx context.Context, agreement *models.CreditAgreement) error {
	query := `
		INSERT INTO credit_agreements (
			enterprise_id, report_id, agreement_number, agreement_date, creditor_name, creditor_country,
			principal_amount, currency, interest_rate, rate_type, rate_base, rate_spread,
			start_date, maturity_date, term_months,
			collateral_type, collateral_description,
			status, outstanding_balance, next_payment_date, next_payment_amount
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)
		RETURNING id, created_at, updated_at
	`

	var createdAt, updatedAt time.Time
	err := r.db.QueryRowContext(ctx, query,
		agreement.EnterpriseID,
		agreement.ReportID,
		agreement.AgreementNumber,
		agreement.AgreementDate,
		agreement.CreditorName,
		agreement.CreditorCountry,
		agreement.PrincipalAmount,
		agreement.Currency,
		agreement.InterestRate,
		agreement.RateType,
		agreement.RateBase,
		agreement.RateSpread,
		agreement.StartDate,
		agreement.MaturityDate,
		agreement.TermMonths,
		agreement.CollateralType,
		agreement.CollateralDescription,
		agreement.Status,
		agreement.OutstandingBalance,
		agreement.NextPaymentDate,
		agreement.NextPaymentAmount,
	).Scan(&agreement.ID, &createdAt, &updatedAt)

	if err != nil {
		return fmt.Errorf("failed to create credit agreement: %w", err)
	}

	agreement.CreatedAt = createdAt
	agreement.UpdatedAt = updatedAt

	return nil
}

// GetByID получает договор по ID
func (r *CreditAgreementRepository) GetByID(ctx context.Context, id int64) (*models.CreditAgreement, error) {
	query := `
		SELECT 
			id, enterprise_id, report_id, agreement_number, agreement_date, creditor_name, creditor_country,
			principal_amount, currency, interest_rate, rate_type, rate_base, rate_spread,
			start_date, maturity_date, term_months,
			collateral_type, collateral_description,
			status, outstanding_balance, next_payment_date, next_payment_amount,
			created_at, updated_at
		FROM credit_agreements
		WHERE id = $1
	`

	agreement := &models.CreditAgreement{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&agreement.ID,
		&agreement.EnterpriseID,
		&agreement.ReportID,
		&agreement.AgreementNumber,
		&agreement.AgreementDate,
		&agreement.CreditorName,
		&agreement.CreditorCountry,
		&agreement.PrincipalAmount,
		&agreement.Currency,
		&agreement.InterestRate,
		&agreement.RateType,
		&agreement.RateBase,
		&agreement.RateSpread,
		&agreement.StartDate,
		&agreement.MaturityDate,
		&agreement.TermMonths,
		&agreement.CollateralType,
		&agreement.CollateralDescription,
		&agreement.Status,
		&agreement.OutstandingBalance,
		&agreement.NextPaymentDate,
		&agreement.NextPaymentAmount,
		&agreement.CreatedAt,
		&agreement.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("credit agreement not found with id %d", id)
		}
		return nil, fmt.Errorf("failed to get credit agreement: %w", err)
	}

	return agreement, nil
}

// GetByEnterpriseID получает договоры предприятия
func (r *CreditAgreementRepository) GetByEnterpriseID(ctx context.Context, enterpriseID int64) ([]*models.CreditAgreement, error) {
	query := `
		SELECT 
			id, enterprise_id, report_id, agreement_number, agreement_date, creditor_name, creditor_country,
			principal_amount, currency, interest_rate, rate_type, rate_base, rate_spread,
			start_date, maturity_date, term_months,
			collateral_type, collateral_description,
			status, outstanding_balance, next_payment_date, next_payment_amount,
			created_at, updated_at
		FROM credit_agreements
		WHERE enterprise_id = $1
		ORDER BY agreement_date DESC
	`

	rows, err := r.db.QueryContext(ctx, query, enterpriseID)
	if err != nil {
		return nil, fmt.Errorf("failed to query credit agreements: %w", err)
	}
	defer rows.Close()

	agreements := []*models.CreditAgreement{}
	for rows.Next() {
		agreement := &models.CreditAgreement{}
		err := rows.Scan(
			&agreement.ID,
			&agreement.EnterpriseID,
			&agreement.ReportID,
			&agreement.AgreementNumber,
			&agreement.AgreementDate,
			&agreement.CreditorName,
			&agreement.CreditorCountry,
			&agreement.PrincipalAmount,
			&agreement.Currency,
			&agreement.InterestRate,
			&agreement.RateType,
			&agreement.RateBase,
			&agreement.RateSpread,
			&agreement.StartDate,
			&agreement.MaturityDate,
			&agreement.TermMonths,
			&agreement.CollateralType,
			&agreement.CollateralDescription,
			&agreement.Status,
			&agreement.OutstandingBalance,
			&agreement.NextPaymentDate,
			&agreement.NextPaymentAmount,
			&agreement.CreatedAt,
			&agreement.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan credit agreement: %w", err)
		}
		agreements = append(agreements, agreement)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return agreements, nil
}

// GetByReportID получает договоры из отчёта
func (r *CreditAgreementRepository) GetByReportID(ctx context.Context, reportID int64) ([]*models.CreditAgreement, error) {
	query := `
		SELECT 
			id, enterprise_id, report_id, agreement_number, agreement_date, creditor_name, creditor_country,
			principal_amount, currency, interest_rate, rate_type, rate_base, rate_spread,
			start_date, maturity_date, term_months,
			collateral_type, collateral_description,
			status, outstanding_balance, next_payment_date, next_payment_amount,
			created_at, updated_at
		FROM credit_agreements
		WHERE report_id = $1
		ORDER BY agreement_date DESC
	`

	rows, err := r.db.QueryContext(ctx, query, reportID)
	if err != nil {
		return nil, fmt.Errorf("failed to query credit agreements: %w", err)
	}
	defer rows.Close()

	agreements := []*models.CreditAgreement{}
	for rows.Next() {
		agreement := &models.CreditAgreement{}
		err := rows.Scan(
			&agreement.ID,
			&agreement.EnterpriseID,
			&agreement.ReportID,
			&agreement.AgreementNumber,
			&agreement.AgreementDate,
			&agreement.CreditorName,
			&agreement.CreditorCountry,
			&agreement.PrincipalAmount,
			&agreement.Currency,
			&agreement.InterestRate,
			&agreement.RateType,
			&agreement.RateBase,
			&agreement.RateSpread,
			&agreement.StartDate,
			&agreement.MaturityDate,
			&agreement.TermMonths,
			&agreement.CollateralType,
			&agreement.CollateralDescription,
			&agreement.Status,
			&agreement.OutstandingBalance,
			&agreement.NextPaymentDate,
			&agreement.NextPaymentAmount,
			&agreement.CreatedAt,
			&agreement.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan credit agreement: %w", err)
		}
		agreements = append(agreements, agreement)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return agreements, nil
}

// GetAll получает все договоры
func (r *CreditAgreementRepository) GetAll(ctx context.Context) ([]*models.CreditAgreement, error) {
	query := `
		SELECT 
			id, enterprise_id, report_id, agreement_number, agreement_date, creditor_name, creditor_country,
			principal_amount, currency, interest_rate, rate_type, rate_base, rate_spread,
			start_date, maturity_date, term_months,
			collateral_type, collateral_description,
			status, outstanding_balance, next_payment_date, next_payment_amount,
			created_at, updated_at
		FROM credit_agreements
		ORDER BY agreement_date DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query credit agreements: %w", err)
	}
	defer rows.Close()

	agreements := []*models.CreditAgreement{}
	for rows.Next() {
		agreement := &models.CreditAgreement{}
		err := rows.Scan(
			&agreement.ID,
			&agreement.EnterpriseID,
			&agreement.ReportID,
			&agreement.AgreementNumber,
			&agreement.AgreementDate,
			&agreement.CreditorName,
			&agreement.CreditorCountry,
			&agreement.PrincipalAmount,
			&agreement.Currency,
			&agreement.InterestRate,
			&agreement.RateType,
			&agreement.RateBase,
			&agreement.RateSpread,
			&agreement.StartDate,
			&agreement.MaturityDate,
			&agreement.TermMonths,
			&agreement.CollateralType,
			&agreement.CollateralDescription,
			&agreement.Status,
			&agreement.OutstandingBalance,
			&agreement.NextPaymentDate,
			&agreement.NextPaymentAmount,
			&agreement.CreatedAt,
			&agreement.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan credit agreement: %w", err)
		}
		agreements = append(agreements, agreement)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return agreements, nil
}

// Update обновляет договор
func (r *CreditAgreementRepository) Update(ctx context.Context, agreement *models.CreditAgreement) error {
	query := `
		UPDATE credit_agreements
		SET agreement_number = $1, agreement_date = $2, creditor_name = $3, creditor_country = $4,
		    principal_amount = $5, currency = $6, interest_rate = $7, rate_type = $8, rate_base = $9, rate_spread = $10,
		    start_date = $11, maturity_date = $12, term_months = $13,
		    collateral_type = $14, collateral_description = $15,
		    status = $16, outstanding_balance = $17, next_payment_date = $18, next_payment_amount = $19,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $20
	`

	result, err := r.db.ExecContext(ctx, query,
		agreement.AgreementNumber,
		agreement.AgreementDate,
		agreement.CreditorName,
		agreement.CreditorCountry,
		agreement.PrincipalAmount,
		agreement.Currency,
		agreement.InterestRate,
		agreement.RateType,
		agreement.RateBase,
		agreement.RateSpread,
		agreement.StartDate,
		agreement.MaturityDate,
		agreement.TermMonths,
		agreement.CollateralType,
		agreement.CollateralDescription,
		agreement.Status,
		agreement.OutstandingBalance,
		agreement.NextPaymentDate,
		agreement.NextPaymentAmount,
		agreement.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update credit agreement: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("credit agreement not found with id %d", agreement.ID)
	}

	return nil
}

// Delete удаляет договор по ID
func (r *CreditAgreementRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM credit_agreements WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete credit agreement: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("credit agreement not found with id %d", id)
	}

	return nil
}
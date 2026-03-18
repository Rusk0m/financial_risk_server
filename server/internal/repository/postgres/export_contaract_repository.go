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

// ExportContractRepository реализует интерфейс для работы с экспортными контрактами
type ExportContractRepository struct {
	db *sql.DB
}

// NewExportContractRepository создаёт новый репозиторий экспортных контрактов
func NewExportContractRepository(db *sql.DB) interfaces.ExportContractRepository {
	return &ExportContractRepository{db: db}
}

// Create создаёт новый контракт
func (r *ExportContractRepository) Create(ctx context.Context, contract *models.ExportContract) error {
	query := `
		INSERT INTO export_contracts (
			enterprise_id, report_id, contract_number, contract_date, country,
			volume_t, price_contract, currency, payment_term_days, shipment_date,
			payment_status, exchange_rate ) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, created_at, updated_at
	`

	var createdAt, updatedAt interface{}
	err := r.db.QueryRowContext(ctx, query,
		contract.EnterpriseID,
		contract.ReportID,
		contract.ContractNumber,
		contract.ContractDate,
		contract.Country,
		contract.VolumeT,
		contract.PriceContract,
		contract.Currency,
		contract.PaymentTermDays,
		contract.ShipmentDate,
		contract.PaymentStatus,
		contract.ExchangeRate,
	).Scan(&contract.ID, &createdAt, &updatedAt)

	if err != nil {
		return fmt.Errorf("failed to create export contract: %w", err)
	}

	// Преобразуем время
	contract.CreatedAt = createdAt.(time.Time)
	contract.UpdatedAt = updatedAt.(time.Time)

	return nil
}

// GetByID получает контракт по ID
func (r *ExportContractRepository) GetByID(ctx context.Context, id int64) (*models.ExportContract, error) {
	query := `
		SELECT 
			id, enterprise_id, report_id, contract_number, contract_date, country,
			volume_t, price_contract, currency, payment_term_days, shipment_date,
			payment_status, exchange_rate, created_at, updated_at
		FROM export_contracts
		WHERE id = $1
	`

	contract := &models.ExportContract{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&contract.ID,
		&contract.EnterpriseID,
		&contract.ReportID,
		&contract.ContractNumber,
		&contract.ContractDate,
		&contract.Country,
		&contract.VolumeT,
		&contract.PriceContract,
		&contract.Currency,
		&contract.PaymentTermDays,
		&contract.ShipmentDate,
		&contract.PaymentStatus,
		&contract.ExchangeRate,
		&contract.CreatedAt,
		&contract.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("export contract not found with id %d", id)
		}
		return nil, fmt.Errorf("failed to get export contract: %w", err)
	}

	return contract, nil
}

// GetByEnterpriseID получает контракты предприятия
func (r *ExportContractRepository) GetUnpaidContract(ctx context.Context, enterpriseID int64) ([]*models.ExportContract, error) {
	query := `
		SELECT 
			id, enterprise_id, report_id, contract_number, contract_date, country,
			volume_t, price_contract, currency, payment_term_days, shipment_date,
			payment_status, exchange_rate, created_at, updated_at
		FROM export_contracts
		WHERE enterprise_id = $1 
			AND payment_status != 'paid'
		ORDER BY contract_date DESC
	`

	rows, err := r.db.QueryContext(ctx, query, enterpriseID)
	if err != nil {
		return nil, fmt.Errorf("failed to query export contracts: %w", err)
	}
	defer rows.Close()

	contracts := []*models.ExportContract{}
	for rows.Next() {
		contract := &models.ExportContract{}
		err := rows.Scan(
			&contract.ID,
			&contract.EnterpriseID,
			&contract.ReportID,
			&contract.ContractNumber,
			&contract.ContractDate,
			&contract.Country,
			&contract.VolumeT,
			&contract.PriceContract,
			&contract.Currency,
			&contract.PaymentTermDays,
			&contract.ShipmentDate,
			&contract.PaymentStatus,
			&contract.ExchangeRate,
			&contract.CreatedAt,
			&contract.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan export contract: %w", err)
		}
		contracts = append(contracts, contract)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return contracts, nil
}

// GetByReportID получает контракты из отчёта
func (r *ExportContractRepository) GetByReportID(ctx context.Context, reportID int64) ([]*models.ExportContract, error) {
	query := `
		SELECT 
			id, enterprise_id, report_id, contract_number, contract_date, country,
			volume_t, price_contract, currency, payment_term_days, shipment_date,
			payment_status, exchange_rate, created_at, updated_at
		FROM export_contracts
		WHERE report_id = $1
		ORDER BY contract_date DESC
	`

	rows, err := r.db.QueryContext(ctx, query, reportID)
	if err != nil {
		return nil, fmt.Errorf("failed to query export contracts: %w", err)
	}
	defer rows.Close()

	contracts := []*models.ExportContract{}
	for rows.Next() {
		contract := &models.ExportContract{}
		err := rows.Scan(
			&contract.ID,
			&contract.EnterpriseID,
			&contract.ReportID,
			&contract.ContractNumber,
			&contract.ContractDate,
			&contract.Country,
			&contract.VolumeT,
			&contract.PriceContract,
			&contract.Currency,
			&contract.PaymentTermDays,
			&contract.ShipmentDate,
			&contract.PaymentStatus,
			&contract.ExchangeRate,
			&contract.CreatedAt,
			&contract.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan export contract: %w", err)
		}
		contracts = append(contracts, contract)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return contracts, nil
}

// GetAll получает все контракты
func (r *ExportContractRepository) GetAll(ctx context.Context) ([]*models.ExportContract, error) {
	query := `
		SELECT 
			id, enterprise_id, report_id, contract_number, contract_date, country,
			volume_t, price_contract, currency, payment_term_days, shipment_date,
			payment_status, exchange_rate, created_at, updated_at
		FROM export_contracts
		ORDER BY contract_date DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query export contracts: %w", err)
	}
	defer rows.Close()

	contracts := []*models.ExportContract{}
	for rows.Next() {
		contract := &models.ExportContract{}
		err := rows.Scan(
			&contract.ID,
			&contract.EnterpriseID,
			&contract.ReportID,
			&contract.ContractNumber,
			&contract.ContractDate,
			&contract.Country,
			&contract.VolumeT,
			&contract.PriceContract,
			&contract.Currency,
			&contract.PaymentTermDays,
			&contract.ShipmentDate,
			&contract.PaymentStatus,
			&contract.ExchangeRate,
			&contract.CreatedAt,
			&contract.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan export contract: %w", err)
		}
		contracts = append(contracts, contract)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return contracts, nil
}

// Update обновляет контракт
func (r *ExportContractRepository) Update(ctx context.Context, contract *models.ExportContract) error {
	query := `
		UPDATE export_contracts
		SET contract_number = $1, contract_date = $2, country = $3,
		    volume_t = $4, price_contract = $5, currency = $6,
		    payment_term_days = $7, shipment_date = $8, payment_status = $9,
		    exchange_rate = $10, updated_at = CURRENT_TIMESTAMP
		WHERE id = $15
	`

	result, err := r.db.ExecContext(ctx, query,
		contract.ContractNumber,
		contract.ContractDate,
		contract.Country,
		contract.VolumeT,
		contract.PriceContract,
		contract.Currency,
		contract.PaymentTermDays,
		contract.ShipmentDate,
		contract.PaymentStatus,
		contract.ExchangeRate,
		contract.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update export contract: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("export contract not found with id %d", contract.ID)
	}

	return nil
}

// Delete удаляет контракт по ID
func (r *ExportContractRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM export_contracts WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete export contract: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("export contract not found with id %d", id)
	}

	return nil
}
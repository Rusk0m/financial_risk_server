package postgres

import (
	"database/sql"
	"errors"
	"finantial-risk-server/internal/domain/models"
	"finantial-risk-server/internal/repository/interfaces"
	"time"
)

type exportContractRepository struct {
	db *DB
}

// NewExportContractRepository создаёт новый репозиторий экспортных контрактов
func NewExportContractRepository(db *DB) interfaces.ExportContractRepository {
	return &exportContractRepository{db: db}
}

func (r *exportContractRepository) Create(c *models.ExportContract) error {
	query := `
        INSERT INTO export_contracts (
            enterprise_id, contract_date, country, volume_t,
            price_usd_per_t, currency, payment_term_days,
            shipment_date, payment_status, exchange_rate,
            created_at, updated_at
        )
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
        RETURNING id
    `

	now := time.Now()
	err := r.db.QueryRow(query,
		c.EnterpriseID,
		c.ContractDate,
		c.Country,
		c.VolumeT,
		c.PriceUSDPerT,
		c.Currency,
		c.PaymentTermDays,
		c.ShipmentDate,
		c.PaymentStatus,
		c.ExchangeRate,
		now,
		now,
	).Scan(&c.ID)

	return err
}

func (r *exportContractRepository) GetByID(id int64) (*models.ExportContract, error) {
	query := `
        SELECT id, enterprise_id, contract_date, country, volume_t,
               price_usd_per_t, currency, payment_term_days,
               shipment_date, payment_status, exchange_rate,
               created_at, updated_at
        FROM export_contracts
        WHERE id = $1
    `

	c := &models.ExportContract{}
	err := r.db.QueryRow(query, id).Scan(
		&c.ID,
		&c.EnterpriseID,
		&c.ContractDate,
		&c.Country,
		&c.VolumeT,
		&c.PriceUSDPerT,
		&c.Currency,
		&c.PaymentTermDays,
		&c.ShipmentDate,
		&c.PaymentStatus,
		&c.ExchangeRate,
		&c.CreatedAt,
		&c.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("контракт не найден")
		}
		return nil, err
	}

	return c, nil
}

func (r *exportContractRepository) GetByEnterpriseID(enterpriseID int64) ([]*models.ExportContract, error) {
	query := `
        SELECT id, enterprise_id, contract_date, country, volume_t,
               price_usd_per_t, currency, payment_term_days,
               shipment_date, payment_status, exchange_rate,
               created_at, updated_at
        FROM export_contracts
        WHERE enterprise_id = $1
        ORDER BY contract_date DESC
    `

	rows, err := r.db.Query(query, enterpriseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contracts []*models.ExportContract
	for rows.Next() {
		c := &models.ExportContract{}
		err := rows.Scan(
			&c.ID,
			&c.EnterpriseID,
			&c.ContractDate,
			&c.Country,
			&c.VolumeT,
			&c.PriceUSDPerT,
			&c.Currency,
			&c.PaymentTermDays,
			&c.ShipmentDate,
			&c.PaymentStatus,
			&c.ExchangeRate,
			&c.CreatedAt,
			&c.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		contracts = append(contracts, c)
	}

	return contracts, rows.Err()
}

func (r *exportContractRepository) GetPendingContracts(enterpriseID int64) ([]*models.ExportContract, error) {
	query := `
        SELECT id, enterprise_id, contract_date, country, volume_t,
               price_usd_per_t, currency, payment_term_days,
               shipment_date, payment_status, exchange_rate,
               created_at, updated_at
        FROM export_contracts
        WHERE enterprise_id = $1 AND payment_status = 'pending'
        ORDER BY shipment_date ASC
    `

	rows, err := r.db.Query(query, enterpriseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contracts []*models.ExportContract
	for rows.Next() {
		c := &models.ExportContract{}
		err := rows.Scan(
			&c.ID,
			&c.EnterpriseID,
			&c.ContractDate,
			&c.Country,
			&c.VolumeT,
			&c.PriceUSDPerT,
			&c.Currency,
			&c.PaymentTermDays,
			&c.ShipmentDate,
			&c.PaymentStatus,
			&c.ExchangeRate,
			&c.CreatedAt,
			&c.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		contracts = append(contracts, c)
	}

	return contracts, rows.Err()
}

func (r *exportContractRepository) GetByPeriod(enterpriseID int64, start, end time.Time) ([]*models.ExportContract, error) {
	query := `
        SELECT id, enterprise_id, contract_date, country, volume_t,
               price_usd_per_t, currency, payment_term_days,
               shipment_date, payment_status, exchange_rate,
               created_at, updated_at
        FROM export_contracts
        WHERE enterprise_id = $1 
          AND contract_date BETWEEN $2 AND $3
        ORDER BY contract_date DESC
    `

	rows, err := r.db.Query(query, enterpriseID, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contracts []*models.ExportContract
	for rows.Next() {
		c := &models.ExportContract{}
		err := rows.Scan(
			&c.ID,
			&c.EnterpriseID,
			&c.ContractDate,
			&c.Country,
			&c.VolumeT,
			&c.PriceUSDPerT,
			&c.Currency,
			&c.PaymentTermDays,
			&c.ShipmentDate,
			&c.PaymentStatus,
			&c.ExchangeRate,
			&c.CreatedAt,
			&c.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		contracts = append(contracts, c)
	}

	return contracts, rows.Err()
}

func (r *exportContractRepository) Update(c *models.ExportContract) error {
	query := `
        UPDATE export_contracts
        SET enterprise_id = $1, contract_date = $2, country = $3,
            volume_t = $4, price_usd_per_t = $5, currency = $6,
            payment_term_days = $7, shipment_date = $8,
            payment_status = $9, exchange_rate = $10, updated_at = $11
        WHERE id = $12
    `

	c.UpdatedAt = time.Now()
	result, err := r.db.Exec(query,
		c.EnterpriseID,
		c.ContractDate,
		c.Country,
		c.VolumeT,
		c.PriceUSDPerT,
		c.Currency,
		c.PaymentTermDays,
		c.ShipmentDate,
		c.PaymentStatus,
		c.ExchangeRate,
		c.UpdatedAt,
		c.ID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("контракт не найден")
	}

	return nil
}

func (r *exportContractRepository) UpdateStatus(id int64, status string) error {
	query := `
        UPDATE export_contracts
        SET payment_status = $1, updated_at = $2
        WHERE id = $3
    `

	result, err := r.db.Exec(query, status, time.Now(), id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("контракт не найден")
	}

	return nil
}

func (r *exportContractRepository) Delete(id int64) error {
	query := `DELETE FROM export_contracts WHERE id = $1`
	result, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("контракт не найден")
	}

	return nil
}

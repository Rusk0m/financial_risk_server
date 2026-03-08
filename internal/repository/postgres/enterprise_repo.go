package postgres

import (
	"database/sql"
	"errors"
	"financial-risk-server/internal/domain/models"
	"financial-risk-server/internal/repository/interfaces"
	"time"
)

type enterpriseRepository struct {
	db *DB
}

// NewEnterpriseRepository создаёт новый репозиторий предприятий
func NewEnterpriseRepository(db *DB) interfaces.EnterpriseRepository {
	return &enterpriseRepository{db: db}
}

func (r *enterpriseRepository) Create(e *models.Enterprise) error {
	// Валидация данных перед сохранением
	if err := e.IsValid(); err != nil {
		return err
	}
	query := `
        INSERT INTO enterprises (name, industry, annual_production_t, 
                                 export_share_percent, main_currency, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        RETURNING id
    `

	now := time.Now()
	err := r.db.QueryRow(query,
		e.Name,
		e.Industry,
		e.AnnualProductionT,
		e.ExportSharePercent,
		e.MainCurrency,
		now,
		now,
	).Scan(&e.ID)

	return err
}

func (r *enterpriseRepository) GetByID(id int64) (*models.Enterprise, error) {
	query := `
        SELECT id, name, industry, annual_production_t, 
               export_share_percent, main_currency, created_at, updated_at
        FROM enterprises
        WHERE id = $1
    `

	e := &models.Enterprise{}
	err := r.db.QueryRow(query, id).Scan(
		&e.ID,
		&e.Name,
		&e.Industry,
		&e.AnnualProductionT,
		&e.ExportSharePercent,
		&e.MainCurrency,
		&e.CreatedAt,
		&e.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("предприятие не найдено")
		}
		return nil, err
	}

	return e, nil
}

func (r *enterpriseRepository) GetByName(name string) (*models.Enterprise, error) {
	query := `
        SELECT id, name, industry, annual_production_t, 
               export_share_percent, main_currency, created_at, updated_at
        FROM enterprises
        WHERE name = $1
    `

	e := &models.Enterprise{}
	err := r.db.QueryRow(query, name).Scan(
		&e.ID,
		&e.Name,
		&e.Industry,
		&e.AnnualProductionT,
		&e.ExportSharePercent,
		&e.MainCurrency,
		&e.CreatedAt,
		&e.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("предприятие не найдено")
		}
		return nil, err
	}

	return e, nil
}

func (r *enterpriseRepository) GetAll() ([]*models.Enterprise, error) {
	query := `
        SELECT id, name, industry, annual_production_t, 
               export_share_percent, main_currency, created_at, updated_at
        FROM enterprises
        ORDER BY created_at DESC
    `

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var enterprises []*models.Enterprise
	for rows.Next() {
		e := &models.Enterprise{}
		err := rows.Scan(
			&e.ID,
			&e.Name,
			&e.Industry,
			&e.AnnualProductionT,
			&e.ExportSharePercent,
			&e.MainCurrency,
			&e.CreatedAt,
			&e.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		enterprises = append(enterprises, e)
	}

	return enterprises, rows.Err()
}

func (r *enterpriseRepository) Update(e *models.Enterprise) error {
	query := `
        UPDATE enterprises
        SET name = $1, industry = $2, annual_production_t = $3,
            export_share_percent = $4, main_currency = $5, updated_at = $6
        WHERE id = $7
    `

	e.UpdatedAt = time.Now()
	result, err := r.db.Exec(query,
		e.Name,
		e.Industry,
		e.AnnualProductionT,
		e.ExportSharePercent,
		e.MainCurrency,
		e.UpdatedAt,
		e.ID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("предприятие не найдено")
	}

	return nil
}

func (r *enterpriseRepository) Delete(id int64) error {
	query := `DELETE FROM enterprises WHERE id = $1`
	result, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("предприятие не найдено")
	}

	return nil
}

package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"financial-risk-server/internal/domain/models"
	"financial-risk-server/internal/repository/interfaces"
)

// EnterpriseRepository реализует интерфейс для работы с предприятиями
type EnterpriseRepository struct {
	db *sql.DB
}

// NewEnterpriseRepository создаёт новый репозиторий предприятий
func NewEnterpriseRepository(db *sql.DB) interfaces.EnterpriseRepository {
	return &EnterpriseRepository{db: db}
}

// Create создаёт новое предприятие
func (r *EnterpriseRepository) Create(ctx context.Context, enterprise *models.Enterprise) error {
	query := `
		INSERT INTO enterprises (
			name, industry, annual_production_t, export_share_percent, 
			main_currency, is_export_oriented
		) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`

	var createdAt, updatedAt time.Time
	err := r.db.QueryRowContext(ctx, query,
		enterprise.Name,
		enterprise.Industry,
		enterprise.AnnualProductionT,
		enterprise.ExportSharePercent,
		enterprise.MainCurrency,
		enterprise.IsExportOriented,
	).Scan(&enterprise.ID, &createdAt, &updatedAt)

	if err != nil {
		if err.Error() == "pq: duplicate key value violates unique constraint \"enterprises_name_key\"" {
			return fmt.Errorf("enterprise with name '%s' already exists", enterprise.Name)
		}
		return fmt.Errorf("failed to create enterprise: %w", err)
	}

	enterprise.CreatedAt = createdAt
	enterprise.UpdatedAt = updatedAt

	return nil
}

// GetByID получает предприятие по ID
func (r *EnterpriseRepository) GetByID(ctx context.Context, id int64) (*models.Enterprise, error) {
	query := `
		SELECT 
			id, name, industry, annual_production_t, export_share_percent,
			main_currency, is_export_oriented, created_at, updated_at
		FROM enterprises
		WHERE id = $1
	`

	enterprise := &models.Enterprise{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&enterprise.ID,
		&enterprise.Name,
		&enterprise.Industry,
		&enterprise.AnnualProductionT,
		&enterprise.ExportSharePercent,
		&enterprise.MainCurrency,
		&enterprise.IsExportOriented,
		&enterprise.CreatedAt,
		&enterprise.UpdatedAt,
	)

	if err != nil {
		// 🔍 Логирование ошибки с деталями
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("❌ [EnterpriseRepo] Предприятие не найдено: id=%d", id)
			return nil, fmt.Errorf("enterprise not found with id %d", id)
		}
		log.Printf("❌ [EnterpriseRepo] Ошибка БД: %v (id=%d)", err, id)
		return nil, fmt.Errorf("failed to get enterprise: %w", err)
	}
	
	log.Printf("✅ [EnterpriseRepo] Предприятие найдено: id=%d, name=%q", id, enterprise.Name)
	return enterprise, nil
}

// GetAll получает все предприятия
func (r *EnterpriseRepository) GetAll(ctx context.Context) ([]*models.Enterprise, error) {
	query := `
		SELECT 
			id, name, industry, annual_production_t, export_share_percent,
			main_currency, is_export_oriented, created_at, updated_at
		FROM enterprises
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query enterprises: %w", err)
	}
	defer rows.Close()

	enterprises := []*models.Enterprise{}
	for rows.Next() {
		enterprise := &models.Enterprise{}
		err := rows.Scan(
			&enterprise.ID,
			&enterprise.Name,
			&enterprise.Industry,
			&enterprise.AnnualProductionT,
			&enterprise.ExportSharePercent,
			&enterprise.MainCurrency,
			&enterprise.IsExportOriented,
			&enterprise.CreatedAt,
			&enterprise.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan enterprise: %w", err)
		}
		enterprises = append(enterprises, enterprise)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return enterprises, nil
}

// Update обновляет предприятие
func (r *EnterpriseRepository) Update(ctx context.Context, enterprise *models.Enterprise) error {
	query := `
		UPDATE enterprises
		SET name = $1, industry = $2, annual_production_t = $3, 
		    export_share_percent = $4, main_currency = $5, 
		    is_export_oriented = $6, updated_at = CURRENT_TIMESTAMP
		WHERE id = $7
	`

	result, err := r.db.ExecContext(ctx, query,
		enterprise.Name,
		enterprise.Industry,
		enterprise.AnnualProductionT,
		enterprise.ExportSharePercent,
		enterprise.MainCurrency,
		enterprise.IsExportOriented,
		enterprise.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update enterprise: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("enterprise not found with id %d", enterprise.ID)
	}

	return nil
}

// Delete удаляет предприятие по ID
func (r *EnterpriseRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM enterprises WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete enterprise: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("enterprise not found with id %d", id)
	}

	return nil
}
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

// ReportRepository реализует интерфейс для работы с отчётами
type ReportRepository struct {
	db *sql.DB
}

// NewReportRepository создаёт новый репозиторий отчётов
func NewReportRepository(db *sql.DB) *ReportRepository {
	return &ReportRepository{db: db}
}

// Create создаёт новый отчёт
func (r *ReportRepository) Create(ctx context.Context, report *models.Report) error {
	query := `
		INSERT INTO reports (
			enterprise_id, report_type, file_name, original_name, file_path,
			file_size_bytes, upload_date, processing_status, processing_error,
			period_start, period_end, uploaded_by, description
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, created_at, updated_at
	`

	var createdAt, updatedAt time.Time
	err := r.db.QueryRowContext(ctx, query,
		report.EnterpriseID,
		report.ReportType,
		report.FileName,
		report.OriginalName,
		report.FilePath,
		report.FileSizeBytes,
		report.UploadDate,
		report.ProcessingStatus,
		report.ProcessingError,
		report.PeriodStart,
		report.PeriodEnd,
		report.UploadedBy,
		report.Description,
	).Scan(&report.ID, &createdAt, &updatedAt)

	if err != nil {
		return fmt.Errorf("failed to create report: %w", err)
	}

	report.CreatedAt = createdAt
	report.UpdatedAt = updatedAt

	return nil
}

// GetByID получает отчёт по ID
func (r *ReportRepository) GetByID(ctx context.Context, id int64) (*models.Report, error) {
	query := `
		SELECT 
			id, enterprise_id, report_type, file_name, original_name, file_path,
			file_size_bytes, upload_date, processing_status, processing_error,
			period_start, period_end, uploaded_by, description, created_at, updated_at
		FROM reports
		WHERE id = $1
	`

	report := &models.Report{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&report.ID,
		&report.EnterpriseID,
		&report.ReportType,
		&report.FileName,
		&report.OriginalName,
		&report.FilePath,
		&report.FileSizeBytes,
		&report.UploadDate,
		&report.ProcessingStatus,
		&report.ProcessingError,
		&report.PeriodStart,
		&report.PeriodEnd,
		&report.UploadedBy,
		&report.Description,
		&report.CreatedAt,
		&report.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("report not found with id %d", id)
		}
		return nil, fmt.Errorf("failed to get report: %w", err)
	}

	return report, nil
}

// GetAll получает список отчётов с фильтрацией
func (r *ReportRepository) GetAll(ctx context.Context, filter interfaces.ReportFilter) ([]*models.Report, error) {
	query := `
		SELECT 
			id, enterprise_id, report_type, file_name, original_name, file_path,
			file_size_bytes, upload_date, processing_status, processing_error,
			period_start, period_end, uploaded_by, description, created_at, updated_at
		FROM reports
		WHERE enterprise_id = 1
	`

	args := []interface{}{}
	argID := 1

	// Фильтрация по предприятию
	if filter.EnterpriseID != nil {
		query += fmt.Sprintf(" AND enterprise_id = $%d", argID)
		args = append(args, *filter.EnterpriseID)
		argID++
	}

	// Фильтрация по типу отчёта
	if filter.ReportType != nil {
		query += fmt.Sprintf(" AND report_type = $%d", argID)
		args = append(args, *filter.ReportType)
		argID++
	}

	// Фильтрация по статусу
	if filter.Status != nil {
		query += fmt.Sprintf(" AND processing_status = $%d", argID)
		args = append(args, *filter.Status)
		argID++
	}

	// Фильтрация по периоду
	if filter.PeriodStart != nil {
		query += fmt.Sprintf(" AND period_end >= $%d", argID)
		args = append(args, *filter.PeriodStart)
		argID++
	}
	if filter.PeriodEnd != nil {
		query += fmt.Sprintf(" AND period_start <= $%d", argID)
		args = append(args, *filter.PeriodEnd)
		argID++
	}

	// Сортировка и пагинация
	query += " ORDER BY upload_date DESC"
	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argID)
		args = append(args, filter.Limit)
		argID++
	}
	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argID)
		args = append(args, filter.Offset)
	}

	// Выполнение запроса
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query reports: %w", err)
	}
	defer rows.Close()

	reports := []*models.Report{}
	for rows.Next() {
		report := &models.Report{}
		err := rows.Scan(
			&report.ID,
			&report.EnterpriseID,
			&report.ReportType,
			&report.FileName,
			&report.OriginalName,
			&report.FilePath,
			&report.FileSizeBytes,
			&report.UploadDate,
			&report.ProcessingStatus,
			&report.ProcessingError,
			&report.PeriodStart,
			&report.PeriodEnd,
			&report.UploadedBy,
			&report.Description,
			&report.CreatedAt,
			&report.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan report: %w", err)
		}
		reports = append(reports, report)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return reports, nil
}

// UpdateProcessingStatus обновляет статус обработки отчёта
func (r *ReportRepository) UpdateProcessingStatus(ctx context.Context, id int64, status string, errorMessage *string) error {
	query := `
		UPDATE reports
		SET processing_status = $1, processing_error = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $3
	`

	result, err := r.db.ExecContext(ctx, query, status, errorMessage, id)
	if err != nil {
		return fmt.Errorf("failed to update report status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("report not found with id %d", id)
	}

	return nil
}

// Delete удаляет отчёт по ID
func (r *ReportRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM reports WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete report: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("report not found with id %d", id)
	}

	return nil
}

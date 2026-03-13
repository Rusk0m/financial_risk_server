package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"financial-risk-server/internal/domain/models"
	"financial-risk-server/internal/repository/interfaces"

	"github.com/lib/pq"
)

// ComprehensiveRiskReportRepository реализует интерфейс для работы с комплексными отчётами
type ComprehensiveRiskReportRepository struct {
	db *sql.DB
}

// NewComprehensiveRiskReportRepository создаёт новый репозиторий комплексных отчётов
func NewComprehensiveRiskReportRepository(db *sql.DB) interfaces.ComprehensiveRiskReportRepository {
	return &ComprehensiveRiskReportRepository{db: db}
}

// Create создаёт новый комплексный отчёт
func (r *ComprehensiveRiskReportRepository) Create(ctx context.Context, report *models.ComprehensiveRiskReport) error {
	query := `
		INSERT INTO comprehensive_risk_reports (
			enterprise_id, report_date, overall_risk_level, total_risk_value, max_risk_type, max_risk_value,
			currency_risk_id, interest_risk_id, liquidity_risk_id,
			summary, priority_actions, warnings, created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, created_at
	`

	var id int64
	var createdAt time.Time
	err := r.db.QueryRowContext(ctx, query,
		report.EnterpriseID,
		report.ReportDate,
		report.OverallRiskLevel,
		report.TotalRiskValue,
		report.MaxRiskType,
		report.MaxRiskValue,
		report.CurrencyRiskID,
		report.InterestRiskID,
		report.LiquidityRiskID,
		report.Summary,
		pq.Array(report.PriorityActions),
		pq.Array(report.Warnings),
		report.CreatedBy,
	).Scan(&id, &createdAt)

	if err != nil {
		return fmt.Errorf("failed to create comprehensive risk report: %w", err)
	}

	report.ID = id
	report.CreatedAt = createdAt

	return nil
}

// GetByID получает отчёт по ID
func (r *ComprehensiveRiskReportRepository) GetByID(ctx context.Context, id int64) (*models.ComprehensiveRiskReport, error) {
	query := `
		SELECT 
			id, enterprise_id, report_date, overall_risk_level, total_risk_value, max_risk_type, max_risk_value,
			currency_risk_id, interest_risk_id, liquidity_risk_id,
			summary, priority_actions, warnings, created_by, created_at
		FROM comprehensive_risk_reports
		WHERE id = $1
	`

	report := &models.ComprehensiveRiskReport{}
	var priorityActions, warnings []string
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&report.ID,
		&report.EnterpriseID,
		&report.ReportDate,
		&report.OverallRiskLevel,
		&report.TotalRiskValue,
		&report.MaxRiskType,
		&report.MaxRiskValue,
		&report.CurrencyRiskID,
		&report.InterestRiskID,
		&report.LiquidityRiskID,
		&report.Summary,
		pq.Array(&priorityActions),
		pq.Array(&warnings),
		&report.CreatedBy,
		&report.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("comprehensive risk report not found with id %d", id)
		}
		return nil, fmt.Errorf("failed to get comprehensive risk report: %w", err)
	}

	report.PriorityActions = priorityActions
	report.Warnings = warnings
	return report, nil
}

// GetAllByEnterpriseID получает все отчёты для предприятия
func (r *ComprehensiveRiskReportRepository) GetAllByEnterpriseID(ctx context.Context, enterpriseID int64) ([]*models.ComprehensiveRiskReport, error) {
	query := `
		SELECT 
			id, enterprise_id, report_date, overall_risk_level, total_risk_value, max_risk_type, max_risk_value,
			currency_risk_id, interest_risk_id, liquidity_risk_id,
			summary, priority_actions, warnings, created_by, created_at
		FROM comprehensive_risk_reports
		WHERE enterprise_id = $1
		ORDER BY report_date DESC
	`

	rows, err := r.db.QueryContext(ctx, query, enterpriseID)
	if err != nil {
		return nil, fmt.Errorf("failed to query comprehensive risk reports: %w", err)
	}
	defer rows.Close()

	reports := []*models.ComprehensiveRiskReport{}
	for rows.Next() {
		report := &models.ComprehensiveRiskReport{}
		var priorityActions, warnings []string
		err := rows.Scan(
			&report.ID,
			&report.EnterpriseID,
			&report.ReportDate,
			&report.OverallRiskLevel,
			&report.TotalRiskValue,
			&report.MaxRiskType,
			&report.MaxRiskValue,
			&report.CurrencyRiskID,
			&report.InterestRiskID,
			&report.LiquidityRiskID,
			&report.Summary,
			pq.Array(&priorityActions),
			pq.Array(&warnings),
			&report.CreatedBy,
			&report.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan comprehensive risk report: %w", err)
		}
		report.PriorityActions = priorityActions
		report.Warnings = warnings
		reports = append(reports, report)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return reports, nil
}

// GetLatest получает последний отчёт для предприятия
func (r *ComprehensiveRiskReportRepository) GetLatest(ctx context.Context, enterpriseID int64) (*models.ComprehensiveRiskReport, error) {
	query := `
		SELECT 
			id, enterprise_id, report_date, overall_risk_level, total_risk_value, max_risk_type, max_risk_value,
			currency_risk_id, interest_risk_id, liquidity_risk_id,
			summary, priority_actions, warnings, created_by, created_at
		FROM comprehensive_risk_reports
		WHERE enterprise_id = $1
		ORDER BY report_date DESC
		LIMIT 1
	`

	report := &models.ComprehensiveRiskReport{}
	var priorityActions, warnings []string
	err := r.db.QueryRowContext(ctx, query, enterpriseID).Scan(
		&report.ID,
		&report.EnterpriseID,
		&report.ReportDate,
		&report.OverallRiskLevel,
		&report.TotalRiskValue,
		&report.MaxRiskType,
		&report.MaxRiskValue,
		&report.CurrencyRiskID,
		&report.InterestRiskID,
		&report.LiquidityRiskID,
		&report.Summary,
		pq.Array(&priorityActions),
		pq.Array(&warnings),
		&report.CreatedBy,
		&report.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("no comprehensive risk reports found for enterprise id %d", enterpriseID)
		}
		return nil, fmt.Errorf("failed to get latest comprehensive risk report: %w", err)
	}

	report.PriorityActions = priorityActions
	report.Warnings = warnings
	return report, nil
}

// GetAll получает все отчёты
func (r *ComprehensiveRiskReportRepository) GetAll(ctx context.Context) ([]*models.ComprehensiveRiskReport, error) {
	query := `
		SELECT 
			id, enterprise_id, report_date, overall_risk_level, total_risk_value, max_risk_type, max_risk_value,
			currency_risk_id, interest_risk_id, liquidity_risk_id,
			summary, priority_actions, warnings, created_by, created_at
		FROM comprehensive_risk_reports
		ORDER BY report_date DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query comprehensive risk reports: %w", err)
	}
	defer rows.Close()

	reports := []*models.ComprehensiveRiskReport{}
	for rows.Next() {
		report := &models.ComprehensiveRiskReport{}
		var priorityActions, warnings []string
		err := rows.Scan(
			&report.ID,
			&report.EnterpriseID,
			&report.ReportDate,
			&report.OverallRiskLevel,
			&report.TotalRiskValue,
			&report.MaxRiskType,
			&report.MaxRiskValue,
			&report.CurrencyRiskID,
			&report.InterestRiskID,
			&report.LiquidityRiskID,
			&report.Summary,
			pq.Array(&priorityActions),
			pq.Array(&warnings),
			&report.CreatedBy,
			&report.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan comprehensive risk report: %w", err)
		}
		report.PriorityActions = priorityActions
		report.Warnings = warnings
		reports = append(reports, report)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return reports, nil
}

// Update обновляет отчёт
func (r *ComprehensiveRiskReportRepository) Update(ctx context.Context, report *models.ComprehensiveRiskReport) error {
	query := `
		UPDATE comprehensive_risk_reports
		SET report_date = $1, overall_risk_level = $2, total_risk_value = $3, max_risk_type = $4, max_risk_value = $5,
		    currency_risk_id = $6, interest_risk_id = $7, liquidity_risk_id = $8,
		    summary = $9, priority_actions = $10, warnings = $11, created_by = $12
		WHERE id = $13
	`

	result, err := r.db.ExecContext(ctx, query,
		report.ReportDate,
		report.OverallRiskLevel,
		report.TotalRiskValue,
		report.MaxRiskType,
		report.MaxRiskValue,
		report.CurrencyRiskID,
		report.InterestRiskID,
		report.LiquidityRiskID,
		report.Summary,
		pq.Array(report.PriorityActions),
		pq.Array(report.Warnings),
		report.CreatedBy,
		report.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update comprehensive risk report: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("comprehensive risk report not found with id %d", report.ID)
	}

	return nil
}

// Delete удаляет отчёт по ID
func (r *ComprehensiveRiskReportRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM comprehensive_risk_reports WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete comprehensive risk report: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("comprehensive risk report not found with id %d", id)
	}

	return nil
}
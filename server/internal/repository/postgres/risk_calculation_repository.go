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

// RiskCalculationRepository реализует интерфейс для работы с расчётами рисков
type RiskCalculationRepository struct {
	db *sql.DB
}

// NewRiskCalculationRepository создаёт новый репозиторий расчётов рисков
func NewRiskCalculationRepository(db *sql.DB) interfaces.RiskCalculationRepository {
	return &RiskCalculationRepository{db: db}
}

// Create создаёт новый расчёт риска
func (r *RiskCalculationRepository) Create(ctx context.Context, calculation *models.RiskCalculation) error {
	query := `
		INSERT INTO risk_calculations (
			enterprise_id, report_id, risk_type, calculation_date, horizon_days, confidence_level,
			exposure_amount, var_value, stress_test_loss, risk_level,
			calculation_method, scenario_type, assumptions, recommendations
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id, created_at
	`

	var id int64
	var createdAt time.Time
	err := r.db.QueryRowContext(ctx, query,
		calculation.EnterpriseID,
		calculation.ReportID,
		calculation.RiskType,
		calculation.CalculationDate,
		calculation.HorizonDays,
		calculation.ConfidenceLevel,
		calculation.ExposureAmount,
		calculation.VaRValue,
		calculation.StressTestLoss,
		calculation.RiskLevel,
		calculation.CalculationMethod,
		calculation.ScenarioType,
		calculation.Assumptions,
		pq.Array(calculation.Recommendations),
	).Scan(&id, &createdAt)

	if err != nil {
		return fmt.Errorf("failed to create risk calculation: %w", err)
	}

	calculation.ID = id
	calculation.CreatedAt = createdAt

	return nil
}

// GetByID получает расчёт по ID
func (r *RiskCalculationRepository) GetByID(ctx context.Context, id int64) (*models.RiskCalculation, error) {
	query := `
		SELECT 
			id, enterprise_id, report_id, risk_type, calculation_date, horizon_days, confidence_level,
			exposure_amount, var_value, stress_test_loss, risk_level,
			calculation_method, scenario_type, assumptions, recommendations,
			created_at
		FROM risk_calculations
		WHERE id = $1
	`

	calculation := &models.RiskCalculation{}
	var recommendations []string
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&calculation.ID,
		&calculation.EnterpriseID,
		&calculation.ReportID,
		&calculation.RiskType,
		&calculation.CalculationDate,
		&calculation.HorizonDays,
		&calculation.ConfidenceLevel,
		&calculation.ExposureAmount,
		&calculation.VaRValue,
		&calculation.StressTestLoss,
		&calculation.RiskLevel,
		&calculation.CalculationMethod,
		&calculation.ScenarioType,
		&calculation.Assumptions,
		pq.Array(&recommendations),
		&calculation.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("risk calculation not found with id %d", id)
		}
		return nil, fmt.Errorf("failed to get risk calculation: %w", err)
	}

	calculation.Recommendations = recommendations
	return calculation, nil
}

// GetAll получает расчёты с фильтрацией
func (r *RiskCalculationRepository) GetAll(ctx context.Context, filter interfaces.RiskCalculationFilter) ([]*models.RiskCalculation, error) {
	query := `
		SELECT 
			id, enterprise_id, report_id, risk_type, calculation_date, horizon_days, confidence_level,
			exposure_amount, var_value, stress_test_loss, risk_level,
			calculation_method, scenario_type, assumptions, recommendations,
			created_at
		FROM risk_calculations
		WHERE 1=1
	`

	args := []interface{}{}
	argID := 1

	// Фильтрация по предприятию
	if filter.EnterpriseID != nil {
		query += fmt.Sprintf(" AND enterprise_id = $%d", argID)
		args = append(args, *filter.EnterpriseID)
		argID++
	}

	// Фильтрация по типу риска
	if filter.RiskType != nil {
		query += fmt.Sprintf(" AND risk_type = $%d", argID)
		args = append(args, *filter.RiskType)
		argID++
	}

	// Фильтрация по дате
	if filter.StartDate != nil {
		query += fmt.Sprintf(" AND calculation_date >= $%d", argID)
		args = append(args, *filter.StartDate)
		argID++
	}
	if filter.EndDate != nil {
		query += fmt.Sprintf(" AND calculation_date <= $%d", argID)
		args = append(args, *filter.EndDate)
		argID++
	}

	// Сортировка и пагинация
	query += " ORDER BY calculation_date DESC"
	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argID)
		args = append(args, filter.Limit)
		argID++
	}
	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argID)
		args = append(args, filter.Offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query risk calculations: %w", err)
	}
	defer rows.Close()

	calculations := []*models.RiskCalculation{}
	for rows.Next() {
		calculation := &models.RiskCalculation{}
		var recommendations []string
		err := rows.Scan(
			&calculation.ID,
			&calculation.EnterpriseID,
			&calculation.ReportID,
			&calculation.RiskType,
			&calculation.CalculationDate,
			&calculation.HorizonDays,
			&calculation.ConfidenceLevel,
			&calculation.ExposureAmount,
			&calculation.VaRValue,
			&calculation.StressTestLoss,
			&calculation.RiskLevel,
			&calculation.CalculationMethod,
			&calculation.ScenarioType,
			&calculation.Assumptions,
			pq.Array(&recommendations),
			&calculation.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan risk calculation: %w", err)
		}
		calculation.Recommendations = recommendations
		calculations = append(calculations, calculation)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return calculations, nil
}

// GetLatestByType получает последний расчёт по типу риска для предприятия
func (r *RiskCalculationRepository) GetLatestByType(ctx context.Context, enterpriseID int64, riskType string) (*models.RiskCalculation, error) {
	query := `
		SELECT 
			id, enterprise_id, report_id, risk_type, calculation_date, horizon_days, confidence_level,
			exposure_amount, var_value, stress_test_loss, risk_level,
			calculation_method, scenario_type, assumptions, recommendations,
			created_at
		FROM risk_calculations
		WHERE enterprise_id = $1 AND risk_type = $2
		ORDER BY calculation_date DESC
		LIMIT 1
	`

	calculation := &models.RiskCalculation{}
	var recommendations []string
	err := r.db.QueryRowContext(ctx, query, enterpriseID, riskType).Scan(
		&calculation.ID,
		&calculation.EnterpriseID,
		&calculation.ReportID,
		&calculation.RiskType,
		&calculation.CalculationDate,
		&calculation.HorizonDays,
		&calculation.ConfidenceLevel,
		&calculation.ExposureAmount,
		&calculation.VaRValue,
		&calculation.StressTestLoss,
		&calculation.RiskLevel,
		&calculation.CalculationMethod,
		&calculation.ScenarioType,
		&calculation.Assumptions,
		pq.Array(&recommendations),
		&calculation.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("no risk calculation found for enterprise %d and risk type %s", enterpriseID, riskType)
		}
		return nil, fmt.Errorf("failed to get latest risk calculation: %w", err)
	}

	calculation.Recommendations = recommendations
	return calculation, nil
}

// GetAllByEnterpriseID получает все расчёты для предприятия
func (r *RiskCalculationRepository) GetAllByEnterpriseID(ctx context.Context, enterpriseID int64) ([]*models.RiskCalculation, error) {
	query := `
		SELECT 
			id, enterprise_id, report_id, risk_type, calculation_date, horizon_days, confidence_level,
			exposure_amount, var_value, stress_test_loss, risk_level,
			calculation_method, scenario_type, assumptions, recommendations,
			created_at
		FROM risk_calculations
		WHERE enterprise_id = $1
		ORDER BY calculation_date DESC
	`

	rows, err := r.db.QueryContext(ctx, query, enterpriseID)
	if err != nil {
		return nil, fmt.Errorf("failed to query risk calculations: %w", err)
	}
	defer rows.Close()

	calculations := []*models.RiskCalculation{}
	for rows.Next() {
		calculation := &models.RiskCalculation{}
		var recommendations []string
		err := rows.Scan(
			&calculation.ID,
			&calculation.EnterpriseID,
			&calculation.ReportID,
			&calculation.RiskType,
			&calculation.CalculationDate,
			&calculation.HorizonDays,
			&calculation.ConfidenceLevel,
			&calculation.ExposureAmount,
			&calculation.VaRValue,
			&calculation.StressTestLoss,
			&calculation.RiskLevel,
			&calculation.CalculationMethod,
			&calculation.ScenarioType,
			&calculation.Assumptions,
			pq.Array(&recommendations),
			&calculation.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan risk calculation: %w", err)
		}
		calculation.Recommendations = recommendations
		calculations = append(calculations, calculation)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return calculations, nil
}

// Update обновляет расчёт
func (r *RiskCalculationRepository) Update(ctx context.Context, calculation *models.RiskCalculation) error {
	query := `
		UPDATE risk_calculations
		SET risk_type = $1, calculation_date = $2, horizon_days = $3, confidence_level = $4,
		    exposure_amount = $5, var_value = $6, stress_test_loss = $7, risk_level = $8,
		    calculation_method = $9, scenario_type = $10, assumptions = $11, recommendations = $12
		WHERE id = $13
	`

	result, err := r.db.ExecContext(ctx, query,
		calculation.RiskType,
		calculation.CalculationDate,
		calculation.HorizonDays,
		calculation.ConfidenceLevel,
		calculation.ExposureAmount,
		calculation.VaRValue,
		calculation.StressTestLoss,
		calculation.RiskLevel,
		calculation.CalculationMethod,
		calculation.ScenarioType,
		calculation.Assumptions,
		pq.Array(calculation.Recommendations),
		calculation.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update risk calculation: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("risk calculation not found with id %d", calculation.ID)
	}

	return nil
}

// Delete удаляет расчёт по ID
func (r *RiskCalculationRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM risk_calculations WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete risk calculation: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("risk calculation not found with id %d", id)
	}

	return nil
}
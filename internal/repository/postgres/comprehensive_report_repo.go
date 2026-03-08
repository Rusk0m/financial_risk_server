package postgres

import (
	"database/sql"
	"encoding/json"
	"errors"
	"financial-risk-server/internal/domain/models"
	"financial-risk-server/internal/repository/interfaces"
	"time"
)

type comprehensiveReportRepository struct {
	db *DB
}

// NewComprehensiveReportRepository создаёт новый репозиторий комплексных отчётов
func NewComprehensiveReportRepository(db *DB) interfaces.ComprehensiveRiskReportRepository {
	return &comprehensiveReportRepository{db: db}
}

// Create сохраняет комплексный отчёт в базу данных
func (r *comprehensiveReportRepository) Create(report *models.ComprehensiveRiskReport) error {
	// Преобразуем массивы в JSON
	priorityActionsJSON, err := json.Marshal(report.PriorityActions)
	if err != nil {
		return err
	}

	warningsJSON, err := json.Marshal(report.Warnings)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO comprehensive_risk_reports (
			enterprise_id, report_date, period_start, period_end,
			currency_risk_id, credit_risk_id, liquidity_risk_id, 
			market_risk_id, interest_risk_id,
			total_risk_value, max_risk_type, overall_risk_level,
			summary, priority_actions, warnings, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		RETURNING id
	`

	now := time.Now()
	err = r.db.QueryRow(query,
		report.EnterpriseID,
		report.ReportDate,
		report.PeriodStart,
		report.PeriodEnd,
		nullInt64(report.CurrencyRisk.ID),
		nullInt64(report.CreditRisk.ID),
		nullInt64(report.LiquidityRisk.ID),
		nullInt64(report.MarketRisk.ID),
		nullInt64(report.InterestRisk.ID),
		report.TotalRiskValue,
		report.MaxRiskType,
		report.OverallRiskLevel,
		report.Summary,
		priorityActionsJSON,
		warningsJSON,
		now,
	).Scan(&report.ID)

	if err != nil {
		return err
	}

	report.CreatedAt = now
	return nil
}

// GetByID получает комплексный отчёт по идентификатору
func (r *comprehensiveReportRepository) GetByID(id int64) (*models.ComprehensiveRiskReport, error) {
	query := `
		SELECT id, enterprise_id, report_date, period_start, period_end,
		       currency_risk_id, credit_risk_id, liquidity_risk_id,
		       market_risk_id, interest_risk_id,
		       total_risk_value, max_risk_type, overall_risk_level,
		       summary, priority_actions, warnings, created_at
		FROM comprehensive_risk_reports
		WHERE id = $1
	`

	report := &models.ComprehensiveRiskReport{}
	var priorityActionsJSON, warningsJSON []byte
	var currencyRiskID, creditRiskID, liquidityRiskID, marketRiskID, interestRiskID sql.NullInt64

	err := r.db.QueryRow(query, id).Scan(
		&report.ID,
		&report.EnterpriseID,
		&report.ReportDate,
		&report.PeriodStart,
		&report.PeriodEnd,
		&currencyRiskID,
		&creditRiskID,
		&liquidityRiskID,
		&marketRiskID,
		&interestRiskID,
		&report.TotalRiskValue,
		&report.MaxRiskType,
		&report.OverallRiskLevel,
		&report.Summary,
		&priorityActionsJSON,
		&warningsJSON,
		&report.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("комплексный отчёт не найден")
		}
		return nil, err
	}

	// Восстанавливаем внешние ключи
	if currencyRiskID.Valid {
		report.CurrencyRisk = &models.RiskResult{ID: currencyRiskID.Int64}
	}
	if creditRiskID.Valid {
		report.CreditRisk = &models.RiskResult{ID: creditRiskID.Int64}
	}
	if liquidityRiskID.Valid {
		report.LiquidityRisk = &models.RiskResult{ID: liquidityRiskID.Int64}
	}
	if marketRiskID.Valid {
		report.MarketRisk = &models.RiskResult{ID: marketRiskID.Int64}
	}
	if interestRiskID.Valid {
		report.InterestRisk = &models.RiskResult{ID: interestRiskID.Int64}
	}

	// Восстанавливаем массивы из JSON
	if len(priorityActionsJSON) > 0 {
		err = json.Unmarshal(priorityActionsJSON, &report.PriorityActions)
		if err != nil {
			return nil, err
		}
	}
	if len(warningsJSON) > 0 {
		err = json.Unmarshal(warningsJSON, &report.Warnings)
		if err != nil {
			return nil, err
		}
	}

	return report, nil
}

// GetByEnterpriseID получает все комплексные отчёты для предприятия
func (r *comprehensiveReportRepository) GetByEnterpriseID(enterpriseID int64) ([]*models.ComprehensiveRiskReport, error) {
	query := `
		SELECT id, enterprise_id, report_date, period_start, period_end,
		       currency_risk_id, credit_risk_id, liquidity_risk_id,
		       market_risk_id, interest_risk_id,
		       total_risk_value, max_risk_type, overall_risk_level,
		       summary, priority_actions, warnings, created_at
		FROM comprehensive_risk_reports
		WHERE enterprise_id = $1
		ORDER BY report_date DESC
	`

	rows, err := r.db.Query(query, enterpriseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reports []*models.ComprehensiveRiskReport
	for rows.Next() {
		report := &models.ComprehensiveRiskReport{}
		var priorityActionsJSON, warningsJSON []byte
		var currencyRiskID, creditRiskID, liquidityRiskID, marketRiskID, interestRiskID sql.NullInt64

		err := rows.Scan(
			&report.ID,
			&report.EnterpriseID,
			&report.ReportDate,
			&report.PeriodStart,
			&report.PeriodEnd,
			&currencyRiskID,
			&creditRiskID,
			&liquidityRiskID,
			&marketRiskID,
			&interestRiskID,
			&report.TotalRiskValue,
			&report.MaxRiskType,
			&report.OverallRiskLevel,
			&report.Summary,
			&priorityActionsJSON,
			&warningsJSON,
			&report.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Восстанавливаем внешние ключи
		if currencyRiskID.Valid {
			report.CurrencyRisk = &models.RiskResult{ID: currencyRiskID.Int64}
		}
		if creditRiskID.Valid {
			report.CreditRisk = &models.RiskResult{ID: creditRiskID.Int64}
		}
		if liquidityRiskID.Valid {
			report.LiquidityRisk = &models.RiskResult{ID: liquidityRiskID.Int64}
		}
		if marketRiskID.Valid {
			report.MarketRisk = &models.RiskResult{ID: marketRiskID.Int64}
		}
		if interestRiskID.Valid {
			report.InterestRisk = &models.RiskResult{ID: interestRiskID.Int64}
		}

		// Восстанавливаем массивы из JSON
		if len(priorityActionsJSON) > 0 {
			err = json.Unmarshal(priorityActionsJSON, &report.PriorityActions)
			if err != nil {
				return nil, err
			}
		}
		if len(warningsJSON) > 0 {
			err = json.Unmarshal(warningsJSON, &report.Warnings)
			if err != nil {
				return nil, err
			}
		}

		reports = append(reports, report)
	}

	return reports, rows.Err()
}

// GetLatest получает последний комплексный отчёт для предприятия
func (r *comprehensiveReportRepository) GetLatest(enterpriseID int64) (*models.ComprehensiveRiskReport, error) {
	query := `
		SELECT id, enterprise_id, report_date, period_start, period_end,
		       currency_risk_id, credit_risk_id, liquidity_risk_id,
		       market_risk_id, interest_risk_id,
		       total_risk_value, max_risk_type, overall_risk_level,
		       summary, priority_actions, warnings, created_at
		FROM comprehensive_risk_reports
		WHERE enterprise_id = $1
		ORDER BY report_date DESC
		LIMIT 1
	`

	report := &models.ComprehensiveRiskReport{}
	var priorityActionsJSON, warningsJSON []byte
	var currencyRiskID, creditRiskID, liquidityRiskID, marketRiskID, interestRiskID sql.NullInt64

	err := r.db.QueryRow(query, enterpriseID).Scan(
		&report.ID,
		&report.EnterpriseID,
		&report.ReportDate,
		&report.PeriodStart,
		&report.PeriodEnd,
		&currencyRiskID,
		&creditRiskID,
		&liquidityRiskID,
		&marketRiskID,
		&interestRiskID,
		&report.TotalRiskValue,
		&report.MaxRiskType,
		&report.OverallRiskLevel,
		&report.Summary,
		&priorityActionsJSON,
		&warningsJSON,
		&report.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("комплексные отчёты не найдены")
		}
		return nil, err
	}

	// Восстанавливаем внешние ключи и массивы (аналогично GetByID)
	if currencyRiskID.Valid {
		report.CurrencyRisk = &models.RiskResult{ID: currencyRiskID.Int64}
	}
	if creditRiskID.Valid {
		report.CreditRisk = &models.RiskResult{ID: creditRiskID.Int64}
	}
	if liquidityRiskID.Valid {
		report.LiquidityRisk = &models.RiskResult{ID: liquidityRiskID.Int64}
	}
	if marketRiskID.Valid {
		report.MarketRisk = &models.RiskResult{ID: marketRiskID.Int64}
	}
	if interestRiskID.Valid {
		report.InterestRisk = &models.RiskResult{ID: interestRiskID.Int64}
	}

	if len(priorityActionsJSON) > 0 {
		err = json.Unmarshal(priorityActionsJSON, &report.PriorityActions)
		if err != nil {
			return nil, err
		}
	}
	if len(warningsJSON) > 0 {
		err = json.Unmarshal(warningsJSON, &report.Warnings)
		if err != nil {
			return nil, err
		}
	}

	return report, nil
}

// Update обновляет комплексный отчёт
func (r *comprehensiveReportRepository) Update(report *models.ComprehensiveRiskReport) error {
	priorityActionsJSON, err := json.Marshal(report.PriorityActions)
	if err != nil {
		return err
	}

	warningsJSON, err := json.Marshal(report.Warnings)
	if err != nil {
		return err
	}

	query := `
		UPDATE comprehensive_risk_reports
		SET enterprise_id = $1, report_date = $2, period_start = $3,
		    period_end = $4, currency_risk_id = $5, credit_risk_id = $6,
		    liquidity_risk_id = $7, market_risk_id = $8, interest_risk_id = $9,
		    total_risk_value = $10, max_risk_type = $11, overall_risk_level = $12,
		    summary = $13, priority_actions = $14, warnings = $15
		WHERE id = $16
	`

	_, err = r.db.Exec(query,
		report.EnterpriseID,
		report.ReportDate,
		report.PeriodStart,
		report.PeriodEnd,
		nullInt64(report.CurrencyRisk.ID),
		nullInt64(report.CreditRisk.ID),
		nullInt64(report.LiquidityRisk.ID),
		nullInt64(report.MarketRisk.ID),
		nullInt64(report.InterestRisk.ID),
		report.TotalRiskValue,
		report.MaxRiskType,
		report.OverallRiskLevel,
		report.Summary,
		priorityActionsJSON,
		warningsJSON,
		report.ID,
	)

	return err
}

// Delete удаляет комплексный отчёт
func (r *comprehensiveReportRepository) Delete(id int64) error {
	query := `DELETE FROM comprehensive_risk_reports WHERE id = $1`
	result, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("комплексный отчёт не найден")
	}

	return nil
}

// Вспомогательная функция для обработки NULL в внешних ключах
func nullInt64(id int64) interface{} {
	if id == 0 {
		return nil
	}
	return id
}

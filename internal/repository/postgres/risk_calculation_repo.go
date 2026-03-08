package postgres

import (
	"database/sql"
	"encoding/json"
	"errors"
	"financial-risk-server/internal/domain/models"
	"financial-risk-server/internal/repository/interfaces"
	"time"
)

type riskCalculationRepository struct {
	db *DB
}

// NewRiskCalculationRepository создаёт новый репозиторий расчётов рисков
func NewRiskCalculationRepository(db *DB) interfaces.RiskCalculationRepository {
	return &riskCalculationRepository{db: db}
}

// Create сохраняет результат расчёта риска в базу данных
func (r *riskCalculationRepository) Create(result *models.RiskResult) error {
	// Преобразуем рекомендации в JSON
	recommendationsJSON, err := json.Marshal(result.Recommendations)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO risk_calculations (
			enterprise_id, risk_type, calculation_date, horizon_days,
			confidence_level, exposure_amount, var_value, stress_test_loss,
			risk_level, scenario_type, recommendations, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id
	`

	now := time.Now()
	err = r.db.QueryRow(query,
		result.EnterpriseID,
		result.RiskType,
		result.CalculationDate,
		result.HorizonDays,
		result.ConfidenceLevel,
		result.ExposureAmount,
		result.VaRValue,
		result.StressTestLoss,
		result.RiskLevel,
		result.ScenarioType,
		recommendationsJSON,
		now,
	).Scan(&result.ID)

	if err != nil {
		return err
	}

	result.CreatedAt = now
	return nil
}

// GetByID получает расчёт риска по идентификатору
func (r *riskCalculationRepository) GetByID(id int64) (*models.RiskResult, error) {
	query := `
		SELECT id, enterprise_id, risk_type, calculation_date, horizon_days,
		       confidence_level, exposure_amount, var_value, stress_test_loss,
		       risk_level, scenario_type, recommendations, created_at
		FROM risk_calculations
		WHERE id = $1
	`

	result := &models.RiskResult{}
	var recommendationsJSON []byte

	err := r.db.QueryRow(query, id).Scan(
		&result.ID,
		&result.EnterpriseID,
		&result.RiskType,
		&result.CalculationDate,
		&result.HorizonDays,
		&result.ConfidenceLevel,
		&result.ExposureAmount,
		&result.VaRValue,
		&result.StressTestLoss,
		&result.RiskLevel,
		&result.ScenarioType,
		&recommendationsJSON,
		&result.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("расчёт риска не найден")
		}
		return nil, err
	}

	// Преобразуем JSON обратно в массив строк
	if len(recommendationsJSON) > 0 {
		err = json.Unmarshal(recommendationsJSON, &result.Recommendations)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

// GetByEnterpriseID получает все расчёты рисков для предприятия
func (r *riskCalculationRepository) GetByEnterpriseID(enterpriseID int64) ([]*models.RiskResult, error) {
	query := `
		SELECT id, enterprise_id, risk_type, calculation_date, horizon_days,
		       confidence_level, exposure_amount, var_value, stress_test_loss,
		       risk_level, scenario_type, recommendations, created_at
		FROM risk_calculations
		WHERE enterprise_id = $1
		ORDER BY calculation_date DESC
	`

	rows, err := r.db.Query(query, enterpriseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*models.RiskResult
	for rows.Next() {
		result := &models.RiskResult{}
		var recommendationsJSON []byte

		err := rows.Scan(
			&result.ID,
			&result.EnterpriseID,
			&result.RiskType,
			&result.CalculationDate,
			&result.HorizonDays,
			&result.ConfidenceLevel,
			&result.ExposureAmount,
			&result.VaRValue,
			&result.StressTestLoss,
			&result.RiskLevel,
			&result.ScenarioType,
			&recommendationsJSON,
			&result.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if len(recommendationsJSON) > 0 {
			err = json.Unmarshal(recommendationsJSON, &result.Recommendations)
			if err != nil {
				return nil, err
			}
		}

		results = append(results, result)
	}

	return results, rows.Err()
}

// GetByRiskType получает расчёты по типу риска
func (r *riskCalculationRepository) GetByRiskType(enterpriseID int64, riskType models.RiskType) ([]*models.RiskResult, error) {
	query := `
		SELECT id, enterprise_id, risk_type, calculation_date, horizon_days,
		       confidence_level, exposure_amount, var_value, stress_test_loss,
		       risk_level, scenario_type, recommendations, created_at
		FROM risk_calculations
		WHERE enterprise_id = $1 AND risk_type = $2
		ORDER BY calculation_date DESC
	`

	rows, err := r.db.Query(query, enterpriseID, riskType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*models.RiskResult
	for rows.Next() {
		result := &models.RiskResult{}
		var recommendationsJSON []byte

		err := rows.Scan(
			&result.ID,
			&result.EnterpriseID,
			&result.RiskType,
			&result.CalculationDate,
			&result.HorizonDays,
			&result.ConfidenceLevel,
			&result.ExposureAmount,
			&result.VaRValue,
			&result.StressTestLoss,
			&result.RiskLevel,
			&result.ScenarioType,
			&recommendationsJSON,
			&result.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if len(recommendationsJSON) > 0 {
			err = json.Unmarshal(recommendationsJSON, &result.Recommendations)
			if err != nil {
				return nil, err
			}
		}

		results = append(results, result)
	}

	return results, rows.Err()
}

// GetByDateRange получает расчёты за период
func (r *riskCalculationRepository) GetByDateRange(enterpriseID int64, start, end time.Time) ([]*models.RiskResult, error) {
	query := `
		SELECT id, enterprise_id, risk_type, calculation_date, horizon_days,
		       confidence_level, exposure_amount, var_value, stress_test_loss,
		       risk_level, scenario_type, recommendations, created_at
		FROM risk_calculations
		WHERE enterprise_id = $1 
		  AND calculation_date >= $2 
		  AND calculation_date <= $3
		ORDER BY calculation_date DESC
	`

	rows, err := r.db.Query(query, enterpriseID, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*models.RiskResult
	for rows.Next() {
		result := &models.RiskResult{}
		var recommendationsJSON []byte

		err := rows.Scan(
			&result.ID,
			&result.EnterpriseID,
			&result.RiskType,
			&result.CalculationDate,
			&result.HorizonDays,
			&result.ConfidenceLevel,
			&result.ExposureAmount,
			&result.VaRValue,
			&result.StressTestLoss,
			&result.RiskLevel,
			&result.ScenarioType,
			&recommendationsJSON,
			&result.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if len(recommendationsJSON) > 0 {
			err = json.Unmarshal(recommendationsJSON, &result.Recommendations)
			if err != nil {
				return nil, err
			}
		}

		results = append(results, result)
	}

	return results, rows.Err()
}

// GetLatest получает последний расчёт риска по типу
func (r *riskCalculationRepository) GetLatest(enterpriseID int64, riskType models.RiskType) (*models.RiskResult, error) {
	query := `
		SELECT id, enterprise_id, risk_type, calculation_date, horizon_days,
		       confidence_level, exposure_amount, var_value, stress_test_loss,
		       risk_level, scenario_type, recommendations, created_at
		FROM risk_calculations
		WHERE enterprise_id = $1 AND risk_type = $2
		ORDER BY calculation_date DESC
		LIMIT 1
	`

	result := &models.RiskResult{}
	var recommendationsJSON []byte

	err := r.db.QueryRow(query, enterpriseID, riskType).Scan(
		&result.ID,
		&result.EnterpriseID,
		&result.RiskType,
		&result.CalculationDate,
		&result.HorizonDays,
		&result.ConfidenceLevel,
		&result.ExposureAmount,
		&result.VaRValue,
		&result.StressTestLoss,
		&result.RiskLevel,
		&result.ScenarioType,
		&recommendationsJSON,
		&result.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("расчёты риска не найдены")
		}
		return nil, err
	}

	if len(recommendationsJSON) > 0 {
		err = json.Unmarshal(recommendationsJSON, &result.Recommendations)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

// Update обновляет существующий расчёт риска
func (r *riskCalculationRepository) Update(result *models.RiskResult) error {
	recommendationsJSON, err := json.Marshal(result.Recommendations)
	if err != nil {
		return err
	}

	query := `
		UPDATE risk_calculations
		SET enterprise_id = $1, risk_type = $2, calculation_date = $3,
		    horizon_days = $4, confidence_level = $5, exposure_amount = $6,
		    var_value = $7, stress_test_loss = $8, risk_level = $9,
		    scenario_type = $10, recommendations = $11
		WHERE id = $12
	`

	_, err = r.db.Exec(query,
		result.EnterpriseID,
		result.RiskType,
		result.CalculationDate,
		result.HorizonDays,
		result.ConfidenceLevel,
		result.ExposureAmount,
		result.VaRValue,
		result.StressTestLoss,
		result.RiskLevel,
		result.ScenarioType,
		recommendationsJSON,
		result.ID,
	)

	return err
}

// Delete удаляет расчёт риска
func (r *riskCalculationRepository) Delete(id int64) error {
	query := `DELETE FROM risk_calculations WHERE id = $1`
	result, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("расчёт риска не найден")
	}

	return nil
}

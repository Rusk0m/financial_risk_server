package postgres

import (
	"database/sql"
	"errors"
	"financial-risk-server/internal/domain/models"
	"financial-risk-server/internal/repository/interfaces"
	"time"
)

type costStructureRepository struct {
	db *DB
}

// NewCostStructureRepository создает новый репозиторий структуры затрат
func NewCostStructureRepository(db *DB) interfaces.CostStructureRepository {
	return &costStructureRepository{db: db}
}

func (r *costStructureRepository) Create(cost *models.CostStructure) error {
	if err := cost.IsValid(); err != nil {
		return err
	}

	query := `
		INSERT INTO cost_structures (enterprise_id, cost_item, currency, cost_per_t, 
									period_start, period_end, comment, created_at, updated_at)
		VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`
	now := time.Now()
	err := r.db.QueryRow(query,
		cost.EnterpriseID,
		cost.CostItem,
		cost.Currency,
		cost.CostPerT,
		cost.PeriodStart,
		cost.PeriodEnd,
		cost.Comment,
		now,
		now,
	).Scan(&cost.ID)

	return err
}

func (r *costStructureRepository) GetByEnterpriseID(enterpriseID int64) ([]*models.CostStructure, error) {
	query := `
		SELECT id, enterprise_id, cost_item, currency, cost_per_t,
		       period_start, period_end, comment, created_at, updated_at
		FROM cost_structures
		WHERE enterprise_id = $1
		ORDER BY period_start DESC
	`

	rows, err := r.db.Query(query, enterpriseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var costs []*models.CostStructure

	for rows.Next() {
		cost := &models.CostStructure{}
		err := rows.Scan(
			cost.ID,
			&cost.EnterpriseID,
			&cost.CostItem,
			&cost.Currency,
			&cost.CostPerT,
			&cost.PeriodStart,
			&cost.PeriodEnd,
			&cost.Comment,
			&cost.CreatedAt,
			&cost.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		costs = append(costs, cost)
	}

	return costs, rows.Err()

}

func (r *costStructureRepository) GetByID(id int64) (*models.CostStructure, error) {
	qwery := `
		SELECT *
		FROM cost_structures
		WHERE id = $1
	`
	cs := &models.CostStructure{}

	err := r.db.QueryRow(qwery, id).Scan(
		&cs.ID,
		&cs.EnterpriseID,
		&cs.CostItem,
		&cs.Currency,
		&cs.CostPerT,
		&cs.PeriodStart,
		&cs.PeriodEnd,
		&cs.Comment,
		&cs.CreatedAt,
		&cs.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("статьи затрат не найдены")
		}
		return nil, err
	}
	return cs, nil
}
func (r *costStructureRepository) GetByPeriod(enterpriseID int64, start, end time.Time) ([]*models.CostStructure, error) {

	qwery := `
		SELECT *
		FROM cost_structure
		WHERE enterprise_id = $1 
			AND period_start BETWEEN $2 AND $3
		ORDER BY period_start DESC
	`

	rows, err := r.db.Query(qwery, enterpriseID, start, end)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var costs []*models.CostStructure
	for rows.Next() {
		c := &models.CostStructure{}
		err := rows.Scan(
			&c.ID,
			&c.EnterpriseID,
			&c.CostItem,
			&c.Currency,
			&c.CostPerT,
			&c.PeriodStart,
			&c.PeriodEnd,
			&c.Comment,
			&c.CreatedAt,
			&c.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		costs = append(costs, c)
	}
	return costs, nil
}

func (r *costStructureRepository) Update(cost *models.CostStructure) error {
	qwery := `
		UPDATE cost_structure
		SET enterprise_id = $1, cost_item = $2, currency = $3,
			cost_per_t = $4, period_start = $5, period_end = $6,
			comment = $7, updated_at = $8
		WHERE id = $9
	`

	cost.UpdatedAt = time.Now()
	result, err := r.db.Exec(qwery,
		cost.EnterpriseID,
		cost.CostItem,
		cost.Currency,
		cost.CostPerT,
		cost.PeriodStart,
		cost.PeriodEnd,
		cost.Comment,
		cost.UpdatedAt,
		cost.ID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("Статья затрат не найдена")
	}
	return nil
}

func (r *costStructureRepository) Delete(id int64) error {
	query := `DELETE FROM cost_structure WHERE id = $1`
	result, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("Статья затрат не найдена")
	}

	return nil
}

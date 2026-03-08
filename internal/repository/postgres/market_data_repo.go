package postgres

import (
	"database/sql"
	"errors"
	"financial-risk-server/internal/domain/models"
	"financial-risk-server/internal/repository/interfaces"
	"math"
	"time"
)

type marketDataRepository struct {
	db *DB
}

// NewMarketDataRepository создаёт новый репозиторий рыночных данных
func NewMarketDataRepository(db *DB) interfaces.MarketDataRepository {
	return &marketDataRepository{db: db}
}

// Create сохраняет рыночные данные в базу данных
func (r *marketDataRepository) Create(data *models.MarketData) error {
	if err := data.IsValid(); err != nil {
		return err
	}

	query := `
		INSERT INTO market_data (
			data_date, currency_pair, exchange_rate, volatility_30d,
			potassium_price_usd, source, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`

	now := time.Now()
	err := r.db.QueryRow(query,
		data.DataDate,
		data.CurrencyPair,
		data.ExchangeRate,
		data.Volatility30d,
		data.PotassiumPriceUSD,
		data.Source,
		now,
	).Scan(&data.ID)

	if err != nil {
		return err
	}

	data.CreatedAt = now
	return nil
}

// GetByID получает рыночные данные по идентификатору
func (r *marketDataRepository) GetByID(id int64) (*models.MarketData, error) {
	query := `
		SELECT id, data_date, currency_pair, exchange_rate, volatility_30d,
		       potassium_price_usd, source, created_at
		FROM market_data
		WHERE id = $1
	`

	data := &models.MarketData{}
	err := r.db.QueryRow(query, id).Scan(
		&data.ID,
		&data.DataDate,
		&data.CurrencyPair,
		&data.ExchangeRate,
		&data.Volatility30d,
		&data.PotassiumPriceUSD,
		&data.Source,
		&data.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("рыночные данные не найдены")
		}
		return nil, err
	}

	return data, nil
}

// GetLatestByCurrencyPair получает последние рыночные данные для валютной пары
func (r *marketDataRepository) GetLatestByCurrencyPair(pair string) (*models.MarketData, error) {
	query := `
		SELECT id, data_date, currency_pair, exchange_rate, volatility_30d,
		       potassium_price_usd, source, created_at
		FROM market_data
		WHERE currency_pair = $1
		ORDER BY data_date DESC
		LIMIT 1
	`

	data := &models.MarketData{}
	err := r.db.QueryRow(query, pair).Scan(
		&data.ID,
		&data.DataDate,
		&data.CurrencyPair,
		&data.ExchangeRate,
		&data.Volatility30d,
		&data.PotassiumPriceUSD,
		&data.Source,
		&data.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("рыночные данные не найдены")
		}
		return nil, err
	}

	return data, nil
}

// GetByDateRange получает рыночные данные за период
func (r *marketDataRepository) GetByDateRange(pair string, start, end time.Time) ([]*models.MarketData, error) {
	query := `
		SELECT id, data_date, currency_pair, exchange_rate, volatility_30d,
		       potassium_price_usd, source, created_at
		FROM market_data
		WHERE currency_pair = $1
		  AND data_date >= $2
		  AND data_date <= $3
		ORDER BY data_date DESC
	`

	rows, err := r.db.Query(query, pair, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var data []*models.MarketData
	for rows.Next() {
		item := &models.MarketData{}
		err := rows.Scan(
			&item.ID,
			&item.DataDate,
			&item.CurrencyPair,
			&item.ExchangeRate,
			&item.Volatility30d,
			&item.PotassiumPriceUSD,
			&item.Source,
			&item.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		data = append(data, item)
	}

	return data, rows.Err()
}

// GetVolatility рассчитывает волатильность за указанный период
// Волатильность = стандартное отклонение логарифмических доходностей
func (r *marketDataRepository) GetVolatility(pair string, days int) (float64, error) {
	// Получаем исторические данные за период
	query := `
		SELECT exchange_rate
		FROM market_data
		WHERE currency_pair = $1
		  AND data_date >= CURRENT_DATE - $2::INTERVAL
		ORDER BY data_date ASC
	`

	rows, err := r.db.Query(query, pair, days)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	// Собираем курсы для расчёта
	var rates []float64
	for rows.Next() {
		var rate float64
		if err := rows.Scan(&rate); err != nil {
			return 0, err
		}
		rates = append(rates, rate)
	}

	if err := rows.Err(); err != nil {
		return 0, err
	}

	// Нужно минимум 2 точки для расчёта волатильности
	if len(rates) < 2 {
		return 0, errors.New("недостаточно данных для расчёта волатильности")
	}

	// Рассчитываем логарифмические доходности
	var returns []float64
	for i := 1; i < len(rates); i++ {
		if rates[i-1] > 0 {
			returns = append(returns, math.Log(rates[i]/rates[i-1]))
		}
	}

	// Рассчитываем среднее значение доходностей
	var sum float64
	for _, ret := range returns {
		sum += ret
	}
	mean := sum / float64(len(returns))

	// Рассчитываем стандартное отклонение (волатильность)
	var sumSq float64
	for _, ret := range returns {
		diff := ret - mean
		sumSq += diff * diff
	}
	variance := sumSq / float64(len(returns))
	volatility := math.Sqrt(variance) * math.Sqrt(252.0) // Годовая волатильность (252 торговых дня)

	// Ограничиваем разумными значениями (0.01 - 0.5 = 1% - 50%)
	if volatility < 0.01 {
		volatility = 0.01
	}
	if volatility > 0.5 {
		volatility = 0.5
	}

	return volatility, nil
}

// Update обновляет рыночные данные
func (r *marketDataRepository) Update(data *models.MarketData) error {
	if err := data.IsValid(); err != nil {
		return err
	}

	query := `
		UPDATE market_data
		SET data_date = $1, currency_pair = $2, exchange_rate = $3,
		    volatility_30d = $4, potassium_price_usd = $5, source = $6
		WHERE id = $7
	`

	result, err := r.db.Exec(query,
		data.DataDate,
		data.CurrencyPair,
		data.ExchangeRate,
		data.Volatility30d,
		data.PotassiumPriceUSD,
		data.Source,
		data.ID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("рыночные данные не найдены")
	}

	return nil
}

// Delete удаляет рыночные данные
func (r *marketDataRepository) Delete(id int64) error {
	query := `DELETE FROM market_data WHERE id = $1`
	result, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("рыночные данные не найдены")
	}

	return nil
}

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

// MarketDataRepository реализует интерфейс для работы с рыночными данными
type MarketDataRepository struct {
	db *sql.DB
}

// NewMarketDataRepository создаёт новый репозиторий рыночных данных
func NewMarketDataRepository(db *sql.DB) interfaces.MarketDataRepository {
	return &MarketDataRepository{db: db}
}

// Create создаёт новую запись рыночных данных
func (r *MarketDataRepository) Create(ctx context.Context, data *models.MarketData) error {
	query := `
		INSERT INTO market_data (data_date, currency_pair, exchange_rate, volatility_30d, potassium_price_usd, source)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (data_date, currency_pair) DO UPDATE
		SET exchange_rate = EXCLUDED.exchange_rate,
		    volatility_30d = EXCLUDED.volatility_30d,
		    potassium_price_usd = EXCLUDED.potassium_price_usd,
		    source = EXCLUDED.source
		RETURNING id, created_at
	`

	var id int64
	var createdAt time.Time
	err := r.db.QueryRowContext(ctx, query,
		data.DataDate,
		data.CurrencyPair,
		data.ExchangeRate,
		data.Volatility30d,
		data.PotassiumPriceUSD,
		data.Source,
	).Scan(&id, &createdAt)

	if err != nil {
		return fmt.Errorf("failed to create or update market data: %w", err)
	}

	data.ID = id
	data.CreatedAt = createdAt

	return nil
}

// GetByID получает запись по ID
func (r *MarketDataRepository) GetByID(ctx context.Context, id int64) (*models.MarketData, error) {
	query := `
		SELECT id, data_date, currency_pair, exchange_rate, volatility_30d, potassium_price_usd, source, created_at
		FROM market_data
		WHERE id = $1
	`

	data := &models.MarketData{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
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
			return nil, fmt.Errorf("market data not found with id %d", id)
		}
		return nil, fmt.Errorf("failed to get market data: %w", err)
	}

	return data, nil
}

// GetAll получает записи с фильтрацией
func (r *MarketDataRepository) GetAll(ctx context.Context, filter interfaces.MarketDataFilter) ([]*models.MarketData, error) {
	query := `
		SELECT id, data_date, currency_pair, exchange_rate, volatility_30d, potassium_price_usd, source, created_at
		FROM market_data
		WHERE 1=1
	`

	args := []interface{}{}
	argID := 1

	// Фильтрация по дате
	if filter.DataDate != nil {
		query += fmt.Sprintf(" AND data_date = $%d", argID)
		args = append(args, *filter.DataDate)
		argID++
	}

	// Фильтрация по валютной паре
	if filter.CurrencyPair != nil {
		query += fmt.Sprintf(" AND currency_pair = $%d", argID)
		args = append(args, *filter.CurrencyPair)
		argID++
	}

	// Сортировка и пагинация
	query += " ORDER BY data_date DESC"
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
		return nil, fmt.Errorf("failed to query market data: %w", err)
	}
	defer rows.Close()

	dataList := []*models.MarketData{}
	for rows.Next() {
		data := &models.MarketData{}
		err := rows.Scan(
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
			return nil, fmt.Errorf("failed to scan market data: %w", err)
		}
		dataList = append(dataList, data)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return dataList, nil
}

// GetLatestByCurrencyPair получает последнюю запись по валютной паре
func (r *MarketDataRepository) GetLatestByCurrencyPair(ctx context.Context, currencyPair string) (*models.MarketData, error) {
	query := `
		SELECT id, data_date, currency_pair, exchange_rate, volatility_30d, potassium_price_usd, source, created_at
		FROM market_data
		WHERE currency_pair = $1
		ORDER BY data_date DESC
		LIMIT 1
	`

	data := &models.MarketData{}
	err := r.db.QueryRowContext(ctx, query, currencyPair).Scan(
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
			return nil, fmt.Errorf("no market data found for currency pair %s", currencyPair)
		}
		return nil, fmt.Errorf("failed to get latest market data: %w", err)
	}

	return data, nil
}

// Update обновляет запись
func (r *MarketDataRepository) Update(ctx context.Context, data *models.MarketData) error {
	query := `
		UPDATE market_data
		SET exchange_rate = $1, volatility_30d = $2, potassium_price_usd = $3, source = $4
		WHERE id = $5
	`

	result, err := r.db.ExecContext(ctx, query,
		data.ExchangeRate,
		data.Volatility30d,
		data.PotassiumPriceUSD,
		data.Source,
		data.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update market data: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("market data not found with id %d", data.ID)
	}

	return nil
}

// Delete удаляет запись по ID
func (r *MarketDataRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM market_data WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete market data: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("market data not found with id %d", id)
	}

	return nil
}
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

type MarketDataRepository struct {
	db *sql.DB
}

func NewMarketDataRepository(db *sql.DB) interfaces.MarketDataRepository {
	return &MarketDataRepository{db: db}
}

func (r *MarketDataRepository) Create(ctx context.Context, data *models.MarketData) error {
	query := `
		INSERT INTO market_data (
			data_date, currency_pair, exchange_rate, potassium_price_usd,
			source, created_at
		) VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (data_date, currency_pair) DO UPDATE SET
			exchange_rate = EXCLUDED.exchange_rate,
			potassium_price_usd = EXCLUDED.potassium_price_usd,
			source = EXCLUDED.source,
			created_at = EXCLUDED.created_at
		RETURNING id
	`

	var id int64
	err := r.db.QueryRowContext(ctx, query,
		data.DataDate,
		data.CurrencyPair,
		data.ExchangeRate,
		data.PotassiumPriceUSD,
		data.Source,
		data.CreatedAt,
	).Scan(&id)

	if err != nil {
		return fmt.Errorf("failed to create market  %w", err)
	}

	data.ID = id
	return nil
}

func (r *MarketDataRepository) GetByDateAndPair(ctx context.Context, date time.Time, pair string) (*models.MarketData, error) {
	query := `
		SELECT id, data_date, currency_pair, exchange_rate, potassium_price_usd,
		       source, created_at
		FROM market_data
		WHERE data_date = $1 AND currency_pair = $2
	`

	data := &models.MarketData{}
	err := r.db.QueryRowContext(ctx, query, date, pair).Scan(
		&data.ID, &data.DataDate, &data.CurrencyPair,
		&data.ExchangeRate, &data.PotassiumPriceUSD,
		&data.Source, &data.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get market  %w", err)
	}

	return data, nil
}

func (r *MarketDataRepository) GetHistory(ctx context.Context, pair string, days int) ([]*models.MarketData, error) {
	query := `
		SELECT id, data_date, currency_pair, exchange_rate, potassium_price_usd,
		       source, created_at
		FROM market_data
		WHERE currency_pair = $1
		  AND data_date >= $2
		ORDER BY data_date DESC
	`

	fromDate := time.Now().AddDate(0, 0, -days)
	rows, err := r.db.QueryContext(ctx, query, pair, fromDate)
	if err != nil {
		return nil, fmt.Errorf("failed to query market data history: %w", err)
	}
	defer rows.Close()

	data := make([]*models.MarketData, 0)
	for rows.Next() {
		item := &models.MarketData{}
		err := rows.Scan(
			&item.ID, &item.DataDate, &item.CurrencyPair,
			&item.ExchangeRate, &item.PotassiumPriceUSD,
			&item.Source, &item.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan market  %w", err)
		}
		data = append(data, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return data, nil
}

func (r *MarketDataRepository) GetLatest(ctx context.Context, pair string) (*models.MarketData, error) {
	query := `
		SELECT id, data_date, currency_pair, exchange_rate, potassium_price_usd,
		       source, created_at
		FROM market_data
		WHERE currency_pair = $1
		ORDER BY data_date DESC
		LIMIT 1
	`

	data := &models.MarketData{}
	err := r.db.QueryRowContext(ctx, query, pair).Scan(
		&data.ID, &data.DataDate, &data.CurrencyPair,
		&data.ExchangeRate, &data.PotassiumPriceUSD,
		&data.Source, &data.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get latest market data: %w", err)
	}

	return data, nil
}

// GetLatestRates получает последние курсы всех валют
func (r *MarketDataRepository) GetLatestRates(ctx context.Context) (map[string]float64, error) {
	query := `
		SELECT DISTINCT ON (currency_pair)
			currency_pair, exchange_rate
		FROM market_data
		WHERE currency_pair IN ('BYN/USD', 'BYN/EUR', 'BYN/CNY')
		  AND exchange_rate IS NOT NULL
		ORDER BY currency_pair, data_date DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query latest rates: %w", err)
	}
	defer rows.Close()

	rates := make(map[string]float64)
	for rows.Next() {
		var pair string
		var rate sql.NullFloat64
		err := rows.Scan(&pair, &rate)
		if err != nil {
			return nil, fmt.Errorf("failed to scan rate: %w", err)
		}
		if rate.Valid {
			rates[pair] = rate.Float64
		}
	}

	// Fallback: если данных нет, возвращаем стандартные курсы
	if len(rates) == 0 {
		log.Printf("⚠️  [MarketData] No currency rates found, using fallback")
		rates = map[string]float64{
			"BYN/USD": 3.25,
			"BYN/EUR": 3.55,
			"BYN/CNY": 0.45,
		}
	}

	return rates, nil
}

func (r *MarketDataRepository) Update(ctx context.Context, data *models.MarketData) error {
	query := `
		UPDATE market_data
		SET exchange_rate = $1, potassium_price_usd = $2,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $3
	`

	result, err := r.db.ExecContext(ctx, query,
		data.ExchangeRate,
		data.PotassiumPriceUSD,
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
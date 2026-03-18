package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"financial-risk-server/internal/domain/models"
	"financial-risk-server/internal/repository/interfaces"
)

type MarketDataSyncService struct {
	marketRepo   interfaces.MarketDataRepository
	nbrbURLs     map[string]string
	worldBankURL string
}

func NewMarketDataSyncService(marketRepo interfaces.MarketDataRepository) *MarketDataSyncService {
	return &MarketDataSyncService{
		marketRepo: marketRepo,
		nbrbURLs: map[string]string{
			"BYN/USD": "https://api.nbrb.by/ExRates/Rates/Dynamics/431",
			"BYN/EUR": "https://api.nbrb.by/ExRates/Rates/Dynamics/451",
			"BYN/CNY": "https://api.nbrb.by/ExRates/Rates/Dynamics/462",
		},
		worldBankURL: "https://api.worldbank.org/v2/commodities/POTF/price",
	}
}

// SyncAll запускает полную синхронизацию
func (s *MarketDataSyncService) SyncAll(ctx context.Context) error {
	log.Println("🚀 [MarketSync] Starting full market data synchronization")

	var errors []error

	// Синхронизируем все валюты (30 дней)
	for _, pair := range models.SupportedCurrencyPairs() {
		if err := s.SyncCurrencyData(ctx, pair, 270); err != nil {
			log.Printf("❌ [MarketSync] %s sync failed: %v", pair, err)
			errors = append(errors, fmt.Errorf("%s: %w", pair, err))
		}
	}


	if len(errors) > 0 {
		return fmt.Errorf("market data sync completed with %d errors: %v", len(errors), errors)
	}

	log.Println("✅ [MarketSync] Full synchronization completed")
	return nil
}

// SyncCurrencyData синхронизирует курс конкретной валюты
func (s *MarketDataSyncService) SyncCurrencyData(ctx context.Context, pair string, days int) error {
	log.Printf("🔄 [MarketSync] Starting %s sync for %d days", pair, days)

	url, ok := s.nbrbURLs[pair]
	if !ok {
		return fmt.Errorf("unsupported currency pair: %s", pair)
	}

	toDate := time.Now()
	fromDate := toDate.AddDate(0, 0, -days)

	reqURL := fmt.Sprintf("%s?startDate=%s&endDate=%s",
		url,
		fromDate.Format("2006-01-02"),
		toDate.Format("2006-01-02"),
	)

	resp, err := http.Get(reqURL)
	if err != nil {
		return fmt.Errorf("failed to fetch %s from NBRB: %w", pair, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("NBRB API returned status %d for %s: %s", resp.StatusCode, pair, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read NBRB response: %w", err)
	}

	var nbrbData []struct {
		Date            string  `json:"Date"`
		CurOfficialRate float64 `json:"Cur_OfficialRate"`
	}

	if err := json.Unmarshal(body, &nbrbData); err != nil {
		return fmt.Errorf("failed to parse NBRB response for %s: %w", pair, err)
	}

	if len(nbrbData) == 0 {
		return fmt.Errorf("NBRB returned no data for %s", pair)
	}

	log.Printf("📊 [MarketSync] Fetched %d records for %s from NBRB", len(nbrbData), pair)

	savedCount := 0
	for _, record := range nbrbData {
		date, err := time.Parse("2006-01-02T15:04:05", record.Date)
		if err != nil {
			date, _ = time.Parse("2006-01-02", record.Date)
		}

		rate := record.CurOfficialRate
		marketData := &models.MarketData{
			DataDate:     date,
			CurrencyPair: pair,
			ExchangeRate: &rate,
			Source:       "Национальный банк РБ",
			CreatedAt:    time.Now(),
		}

		if err := s.marketRepo.Create(ctx, marketData); err != nil {
			log.Printf("⚠️  [MarketSync] Failed to save %s data for %s: %v", pair, date.Format("2006-01-02"), err)
			continue
		}
		savedCount++
	}

	if savedCount == 0 {
		return fmt.Errorf("failed to save any %s records", pair)
	}

	log.Printf("✅ [MarketSync] Saved %d records for %s", savedCount, pair)
	return nil
}

// SyncPotassiumData синхронизирует цены на калий
func (s *MarketDataSyncService) SyncPotassiumData(ctx context.Context, days int) error {
	log.Printf("🔄 [MarketSync] Starting potassium price sync for %d days", days)

	months := days/30 + 2
	toDate := time.Now()
	fromDate := toDate.AddDate(0, -months, 0)

	url := fmt.Sprintf("%s?format=json&date=%s:%s&freq=M",
		s.worldBankURL,
		fromDate.Format("2006M01"),
		toDate.Format("2006M01"),
	)

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to fetch potassium prices from World Bank: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("World Bank API returned status %d: %s", resp.StatusCode, string(body))
	}

	var rawResponse [][]map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&rawResponse); err != nil {
		return fmt.Errorf("failed to parse World Bank response: %w", err)
	}

	if len(rawResponse) < 2 {
		return fmt.Errorf("unexpected World Bank response format")
	}

	data := rawResponse[1]
	if len(data) == 0 {
		return fmt.Errorf("World Bank returned no potassium price data")
	}

	savedCount := 0
	for _, item := range data {
		dateStr, ok := item["date"].(string)
		if !ok {
			continue
		}

		value, ok := item["value"].(float64)
		if !ok {
			continue
		}

		var year, month int
		fmt.Sscanf(dateStr, "%dM%d", &year, &month)
		date := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)

		marketData := &models.MarketData{
			DataDate:          date,
			CurrencyPair:      "POTASSIUM",
			PotassiumPriceUSD: &value,
			Source:            "World Bank Pink Sheet",
			CreatedAt:         time.Now(),
		}

		if err := s.marketRepo.Create(ctx, marketData); err != nil {
			log.Printf("⚠️  [MarketSync] Failed to save potassium data: %v", err)
			continue
		}
		savedCount++
		log.Printf("✅ [MarketSync] Saved potassium data for %s: $%.2f", date.Format("2006-01-02"), value)
	}

	if savedCount == 0 {
		return fmt.Errorf("failed to save any potassium price records")
	}

	log.Printf("✅ [MarketSync] Saved %d potassium price records", savedCount)
	return nil
}

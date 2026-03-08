package handlers

import (
	"encoding/json"
	"financial-risk-server/internal/domain/models"
	"financial-risk-server/internal/repository/interfaces"
	"financial-risk-server/internal/utils/validator"
	"financial-risk-server/pkg/response"
	"fmt"
	"net/http"
	"time"
)

// MarketDataHandler обрабатывает запросы, связанные с рыночными данными
type MarketDataHandler struct {
	marketRepo interfaces.MarketDataRepository
}

// NewMarketDataHandler создаёт новый обработчик рыночных данных
func NewMarketDataHandler(marketRepo interfaces.MarketDataRepository) *MarketDataHandler {
	return &MarketDataHandler{
		marketRepo: marketRepo,
	}
}

// CreateMarketDataRequest структура запроса на создание рыночных данных
type CreateMarketDataRequest struct {
	DataDate          string  `json:"data_date"`     // YYYY-MM-DD
	CurrencyPair      string  `json:"currency_pair"` // "BYN/USD", "POTASSIUM"
	ExchangeRate      float64 `json:"exchange_rate"`
	Volatility30d     float64 `json:"volatility_30d"`      // опционально
	PotassiumPriceUSD float64 `json:"potassium_price_usd"` // опционально
	Source            string  `json:"source"`
}

// MarketDataResponse структура ответа с данными рынка
type MarketDataResponse struct {
	ID                int64   `json:"id"`
	DataDate          string  `json:"data_date"`
	CurrencyPair      string  `json:"currency_pair"`
	ExchangeRate      float64 `json:"exchange_rate"`
	Volatility30d     float64 `json:"volatility_30d"`
	PotassiumPriceUSD float64 `json:"potassium_price_usd"`
	Source            string  `json:"source"`
}

// toMarketDataResponse преобразует модель в ответ API
func toMarketDataResponse(m *models.MarketData) *MarketDataResponse {
	return &MarketDataResponse{
		ID:                m.ID,
		DataDate:          m.DataDate.Format("2006-01-02"),
		CurrencyPair:      m.CurrencyPair,
		ExchangeRate:      m.ExchangeRate,
		Volatility30d:     m.Volatility30d,
		PotassiumPriceUSD: m.PotassiumPriceUSD,
		Source:            m.Source,
	}
}

// CreateMarketData создаёт новые рыночные данные
func (h *MarketDataHandler) CreateMarketData(w http.ResponseWriter, r *http.Request) {
	var req CreateMarketDataRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.ErrorWithCode(w, http.StatusBadRequest, "INVALID_JSON", "Неверный формат JSON: "+err.Error())
		return
	}

	// Валидация
	validator := validator.NewValidator()
	validationErrors := validator.MergeErrors(
		validator.Required(req.DataDate, "data_date"),
		validator.Required(req.CurrencyPair, "currency_pair"),
		validator.Required(req.Source, "source"),
	)

	if validationErrors != nil {
		response.ValidationError(w, validationErrors.ToMap())
		return
	}

	// Парсим дату
	dataDate, err := time.Parse("2006-01-02", req.DataDate)
	if err != nil {
		validationErrors := map[string]string{"data_date": "Неверный формат даты (ожидается YYYY-MM-DD)"}
		response.ValidationError(w, validationErrors)
		return
	}

	// Создаём модель
	marketData := &models.MarketData{
		DataDate:          dataDate,
		CurrencyPair:      req.CurrencyPair,
		ExchangeRate:      req.ExchangeRate,
		Volatility30d:     req.Volatility30d,
		PotassiumPriceUSD: req.PotassiumPriceUSD,
		Source:            req.Source,
	}

	// Сохраняем в БД
	if err := h.marketRepo.Create(marketData); err != nil {
		response.ErrorWithCode(w, http.StatusInternalServerError, "DB_ERROR", "Ошибка сохранения рыночных данных: "+err.Error())
		return
	}

	response.JSON(w, http.StatusCreated, toMarketDataResponse(marketData))
}

// UpdateMarketData обновляет рыночные данные
func (h *MarketDataHandler) UpdateMarketData(w http.ResponseWriter, r *http.Request) {
	marketDataID := r.PathValue("id")
	if marketDataID == "" {
		marketDataID = r.URL.Query().Get("id")
	}

	if marketDataID == "" {
		response.ErrorWithCode(w, http.StatusBadRequest, "MISSING_ID", "Не указан параметр id")
		return
	}

	var id int64
	_, err := fmt.Sscan(marketDataID, &id)
	if err != nil || id <= 0 {
		response.ErrorWithCode(w, http.StatusBadRequest, "INVALID_ID", "Некорректный идентификатор рыночных данных")
		return
	}

	// Получаем существующие данные
	marketData, err := h.marketRepo.GetByID(id)
	if err != nil {
		response.ErrorWithCode(w, http.StatusNotFound, "NOT_FOUND", "Рыночные данные не найдены: "+err.Error())
		return
	}

	// Парсим тело запроса
	var req CreateMarketDataRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.ErrorWithCode(w, http.StatusBadRequest, "INVALID_JSON", "Неверный формат JSON: "+err.Error())
		return
	}

	// Парсим дату
	dataDate, err := time.Parse("2006-01-02", req.DataDate)
	if err != nil {
		validationErrors := map[string]string{"data_date": "Неверный формат даты"}
		response.ValidationError(w, validationErrors)
		return
	}

	// Обновляем поля
	marketData.DataDate = dataDate
	marketData.CurrencyPair = req.CurrencyPair
	marketData.ExchangeRate = req.ExchangeRate
	marketData.Volatility30d = req.Volatility30d
	marketData.PotassiumPriceUSD = req.PotassiumPriceUSD
	marketData.Source = req.Source

	// Сохраняем в БД
	if err := h.marketRepo.Update(marketData); err != nil {
		response.ErrorWithCode(w, http.StatusInternalServerError, "DB_ERROR", "Ошибка обновления рыночных данных: "+err.Error())
		return
	}

	response.JSON(w, http.StatusOK, toMarketDataResponse(marketData))
}

// DeleteMarketData удаляет рыночные данные
func (h *MarketDataHandler) DeleteMarketData(w http.ResponseWriter, r *http.Request) {
	marketDataID := r.PathValue("id")
	if marketDataID == "" {
		marketDataID = r.URL.Query().Get("id")
	}

	if marketDataID == "" {
		response.ErrorWithCode(w, http.StatusBadRequest, "MISSING_ID", "Не указан параметр id")
		return
	}

	// Парсим ID
	var id int64
	_, err := fmt.Sscan(marketDataID, &id)
	if err != nil || id <= 0 {
		response.ErrorWithCode(w, http.StatusBadRequest, "INVALID_ID", "Некорректный идентификатор marketData")
		return
	}

	// Удаляем контракт из БД
	if err := h.marketRepo.Delete(id); err != nil {
		response.ErrorWithCode(w, http.StatusInternalServerError, "DB_ERROR", "Ошибка удаления контракта: "+err.Error())
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{
		"message": "MarketData успешно удалён",
	})
}

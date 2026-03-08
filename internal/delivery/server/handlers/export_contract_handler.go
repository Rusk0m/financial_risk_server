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

// ExportContractHandler обрабатывает запросы, связанные с экспортными контрактами
type ExportContractHandler struct {
	contractRepo interfaces.ExportContractRepository
}

// NewExportContractHandler создаёт новый обработчик экспортных контрактов
func NewExportContractHandler(contractRepo interfaces.ExportContractRepository) *ExportContractHandler {
	return &ExportContractHandler{
		contractRepo: contractRepo,
	}
}

// CreateContractRequest структура запроса на создание контракта
type CreateContractRequest struct {
	EnterpriseID    int64   `json:"enterprise_id"`
	ContractDate    string  `json:"contract_date"` // YYYY-MM-DD
	Country         string  `json:"country"`
	VolumeT         float64 `json:"volume_t"`
	PriceUSDPerT    float64 `json:"price_usd_per_t"`
	Currency        string  `json:"currency"`
	PaymentTermDays int     `json:"payment_term_days"`
	ShipmentDate    string  `json:"shipment_date"` // YYYY-MM-DD
	PaymentStatus   string  `json:"payment_status"`
	ExchangeRate    float64 `json:"exchange_rate"`
}

// ContractResponse структура ответа с данными контракта
type ContractResponse struct {
	ID              int64   `json:"id"`
	EnterpriseID    int64   `json:"enterprise_id"`
	ContractDate    string  `json:"contract_date"`
	Country         string  `json:"country"`
	VolumeT         float64 `json:"volume_t"`
	PriceUSDPerT    float64 `json:"price_usd_per_t"`
	Currency        string  `json:"currency"`
	PaymentTermDays int     `json:"payment_term_days"`
	ShipmentDate    string  `json:"shipment_date"`
	PaymentStatus   string  `json:"payment_status"`
	ExchangeRate    float64 `json:"exchange_rate"`
	ContractValue   float64 `json:"contract_value"` // Объём × Цена
}

// toContractResponse преобразует модель в ответ API
func toContractResponse(c *models.ExportContract) *ContractResponse {
	return &ContractResponse{
		ID:              c.ID,
		EnterpriseID:    c.EnterpriseID,
		ContractDate:    c.ContractDate.Format("2006-01-02"),
		Country:         c.Country,
		VolumeT:         c.VolumeT,
		PriceUSDPerT:    c.PriceUSDPerT,
		Currency:        c.Currency,
		PaymentTermDays: c.PaymentTermDays,
		ShipmentDate:    c.ShipmentDate.Format("2006-01-02"),
		PaymentStatus:   c.PaymentStatus,
		ExchangeRate:    c.ExchangeRate,
		ContractValue:   c.GetContractValue(),
	}
}

// CreateContract создаёт новый экспортный контракт
func (h *ExportContractHandler) CreateContract(w http.ResponseWriter, r *http.Request) {
	// 1. Парсим тело запроса
	var req CreateContractRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.ErrorWithCode(w, http.StatusBadRequest, "INVALID_JSON", "Неверный формат JSON: "+err.Error())
		return
	}

	// 2. Валидация данных
	validator := validator.NewValidator()
	validationErrors := validator.MergeErrors(
		validator.Required(req.Country, "country"),
		validator.MinValue(req.VolumeT, 1, "volume_t"),
		validator.MaxValue(req.VolumeT, 10000000, "volume_t"),
		validator.MinValue(req.PriceUSDPerT, 100, "price_usd_per_t"),
		validator.MaxValue(req.PriceUSDPerT, 1000, "price_usd_per_t"),
		validator.MinValue(float64(req.PaymentTermDays), 15, "payment_term_days"),
		validator.MaxValue(float64(req.PaymentTermDays), 365, "payment_term_days"),
		validator.CurrencyCode(req.Currency, "currency"),
		validator.MinValue(req.ExchangeRate, 0.1, "exchange_rate"),
		validator.MaxValue(req.ExchangeRate, 10, "exchange_rate"),
	)

	if validationErrors != nil {
		response.ValidationError(w, validationErrors.ToMap())
		return
	}

	// 3. Парсим даты
	contractDate, err := time.Parse("2006-01-02", req.ContractDate)
	if err != nil {
		validationErrors := map[string]string{"contract_date": "Неверный формат даты (ожидается YYYY-MM-DD)"}
		response.ValidationError(w, validationErrors)
		return
	}

	shipmentDate, err := time.Parse("2006-01-02", req.ShipmentDate)
	if err != nil {
		validationErrors := map[string]string{"shipment_date": "Неверный формат даты (ожидается YYYY-MM-DD)"}
		response.ValidationError(w, validationErrors)
		return
	}

	// 4. Создаём модель
	contract := &models.ExportContract{
		EnterpriseID:    req.EnterpriseID,
		ContractDate:    contractDate,
		Country:         req.Country,
		VolumeT:         req.VolumeT,
		PriceUSDPerT:    req.PriceUSDPerT,
		Currency:        req.Currency,
		PaymentTermDays: req.PaymentTermDays,
		ShipmentDate:    shipmentDate,
		PaymentStatus:   req.PaymentStatus,
		ExchangeRate:    req.ExchangeRate,
	}

	// 5. Сохраняем в БД
	if err := h.contractRepo.Create(contract); err != nil {
		response.ErrorWithCode(w, http.StatusInternalServerError, "DB_ERROR", "Ошибка сохранения контракта: "+err.Error())
		return
	}

	// 6. Формируем успешный ответ
	response.JSON(w, http.StatusCreated, toContractResponse(contract))
}

// ListContracts получает список контрактов с фильтрацией
func (h *ExportContractHandler) ListContracts(w http.ResponseWriter, r *http.Request) {
	// Получаем параметры фильтрации
	enterpriseID := r.URL.Query().Get("enterprise_id")
	country := r.URL.Query().Get("country")

	// Получаем контракты из БД
	var contracts []*models.ExportContract
	var err error

	if enterpriseID != "" {
		// Парсим enterprise_id
		var id int64
		_, err := fmt.Sscan(enterpriseID, &id)
		if err != nil || id <= 0 {
			response.ErrorWithCode(w, http.StatusBadRequest, "INVALID_ID", "Некорректный enterprise_id")
			return
		}

		if country != "" {
			// TODO: Добавить метод GetByCountry в репозиторий
			contracts, err = h.contractRepo.GetByEnterpriseID(id)
			// Фильтруем по стране в коде
			filtered := []*models.ExportContract{}
			for _, c := range contracts {
				if c.Country == country {
					filtered = append(filtered, c)
				}
			}
			contracts = filtered
		} else {
			contracts, err = h.contractRepo.GetByEnterpriseID(id)
		}
	} else {
		// Без фильтрации - возвращаем все контракты (не рекомендуется для продакшена)
		// В реальном приложении нужно добавить пагинацию
		contracts = []*models.ExportContract{} // Заглушка
	}

	if err != nil {
		response.ErrorWithCode(w, http.StatusInternalServerError, "DB_ERROR", "Ошибка получения контрактов: "+err.Error())
		return
	}

	// Преобразуем в ответы API
	responses := make([]*ContractResponse, len(contracts))
	for i, c := range contracts {
		responses[i] = toContractResponse(c)
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"contracts": responses,
		"count":     len(responses),
	})
}

// GetContract получает контракт по ID
func (h *ExportContractHandler) GetContract(w http.ResponseWriter, r *http.Request) {
	contractID := r.PathValue("id")
	if contractID == "" {
		contractID = r.URL.Query().Get("id")
	}

	if contractID == "" {
		response.ErrorWithCode(w, http.StatusBadRequest, "MISSING_ID", "Не указан параметр id")
		return
	}

	// Парсим ID
	var id int64
	_, err := fmt.Sscan(contractID, &id)
	if err != nil || id <= 0 {
		response.ErrorWithCode(w, http.StatusBadRequest, "INVALID_ID", "Некорректный идентификатор контракта")
		return
	}

	// Получаем контракт из БД
	contract, err := h.contractRepo.GetByID(id)
	if err != nil {
		response.ErrorWithCode(w, http.StatusNotFound, "NOT_FOUND", "Контракт не найден: "+err.Error())
		return
	}

	response.JSON(w, http.StatusOK, toContractResponse(contract))
}

// UpdateContract обновляет контракт
func (h *ExportContractHandler) UpdateContract(w http.ResponseWriter, r *http.Request) {
	contractID := r.PathValue("id")
	if contractID == "" {
		contractID = r.URL.Query().Get("id")
	}

	if contractID == "" {
		response.ErrorWithCode(w, http.StatusBadRequest, "MISSING_ID", "Не указан параметр id")
		return
	}

	var id int64
	_, err := fmt.Sscan(contractID, &id)
	if err != nil || id <= 0 {
		response.ErrorWithCode(w, http.StatusBadRequest, "INVALID_ID", "Некорректный идентификатор контракта")
		return
	}

	// Получаем существующий контракт
	contract, err := h.contractRepo.GetByID(id)
	if err != nil {
		response.ErrorWithCode(w, http.StatusNotFound, "NOT_FOUND", "Контракт не найден: "+err.Error())
		return
	}

	// Парсим тело запроса
	var req CreateContractRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.ErrorWithCode(w, http.StatusBadRequest, "INVALID_JSON", "Неверный формат JSON: "+err.Error())
		return
	}

	// Валидация (аналогично Create)
	validator := validator.NewValidator()
	validationErrors := validator.MergeErrors(
		validator.Required(req.Country, "country"),
		validator.MinValue(req.VolumeT, 1, "volume_t"),
		validator.MaxValue(req.VolumeT, 10000000, "volume_t"),
		validator.MinValue(req.PriceUSDPerT, 100, "price_usd_per_t"),
		validator.MaxValue(req.PriceUSDPerT, 1000, "price_usd_per_t"),
		validator.CurrencyCode(req.Currency, "currency"),
		validator.MinValue(float64(req.PaymentTermDays), 15, "payment_term_days"),
		validator.MaxValue(float64(req.PaymentTermDays), 365, "payment_term_days"),
		validator.MinValue(req.ExchangeRate, 0.1, "exchange_rate"),
		validator.MaxValue(req.ExchangeRate, 10, "exchange_rate"),
	)

	if validationErrors != nil {
		response.ValidationError(w, validationErrors.ToMap())
		return
	}

	// Парсим даты
	contractDate, err := time.Parse("2006-01-02", req.ContractDate)
	if err != nil {
		validationErrors := map[string]string{"contract_date": "Неверный формат даты (ожидается YYYY-MM-DD)"}
		response.ValidationError(w, validationErrors)
		return
	}

	shipmentDate, err := time.Parse("2006-01-02", req.ShipmentDate)
	if err != nil {
		validationErrors := map[string]string{"shipment_date": "Неверный формат даты (ожидается YYYY-MM-DD)"}
		response.ValidationError(w, validationErrors)
		return
	}

	// Обновляем поля
	contract.EnterpriseID = req.EnterpriseID
	contract.ContractDate = contractDate
	contract.Country = req.Country
	contract.VolumeT = req.VolumeT
	contract.PriceUSDPerT = req.PriceUSDPerT
	contract.Currency = req.Currency
	contract.PaymentTermDays = req.PaymentTermDays
	contract.ShipmentDate = shipmentDate
	contract.PaymentStatus = req.PaymentStatus
	contract.ExchangeRate = req.ExchangeRate

	// Сохраняем в БД
	if err := h.contractRepo.Update(contract); err != nil {
		response.ErrorWithCode(w, http.StatusInternalServerError, "DB_ERROR", "Ошибка обновления контракта: "+err.Error())
		return
	}

	response.JSON(w, http.StatusOK, toContractResponse(contract))
}

// DeleteContract удаляет контракт
func (h *ExportContractHandler) DeleteContract(w http.ResponseWriter, r *http.Request) {
	contractID := r.PathValue("id")
	if contractID == "" {
		contractID = r.URL.Query().Get("id")
	}

	if contractID == "" {
		response.ErrorWithCode(w, http.StatusBadRequest, "MISSING_ID", "Не указан параметр id")
		return
	}

	// Парсим ID
	var id int64
	_, err := fmt.Sscan(contractID, &id)
	if err != nil || id <= 0 {
		response.ErrorWithCode(w, http.StatusBadRequest, "INVALID_ID", "Некорректный идентификатор контракта")
		return
	}

	// Удаляем контракт из БД
	if err := h.contractRepo.Delete(id); err != nil {
		response.ErrorWithCode(w, http.StatusInternalServerError, "DB_ERROR", "Ошибка удаления контракта: "+err.Error())
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{
		"message": "Контракт успешно удалён",
	})
}

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

// CostStructureHandler обрабатывает запросы, связанные со структурой затрат
type CostStructureHandler struct {
	costRepo interfaces.CostStructureRepository
}

// NewCostStructureHandler создаёт новый обработчик структуры затрат
func NewCostStructureHandler(costRepo interfaces.CostStructureRepository) *CostStructureHandler {
	return &CostStructureHandler{
		costRepo: costRepo,
	}
}

// CreateCostRequest структура запроса на создание статьи затрат
type CreateCostRequest struct {
	EnterpriseID int64   `json:"enterprise_id"`
	CostItem     string  `json:"cost_item"`
	Currency     string  `json:"currency"`
	CostPerT     float64 `json:"cost_per_t"`
	PeriodStart  string  `json:"period_start"` // YYYY-MM-DD
	PeriodEnd    string  `json:"period_end"`   // YYYY-MM-DD
	Comment      string  `json:"comment"`
}

// CostResponse структура ответа с данными статьи затрат
type CostResponse struct {
	ID           int64   `json:"id"`
	EnterpriseID int64   `json:"enterprise_id"`
	CostItem     string  `json:"cost_item"`
	Currency     string  `json:"currency"`
	CostPerT     float64 `json:"cost_per_t"`
	PeriodStart  string  `json:"period_start"`
	PeriodEnd    string  `json:"period_end"`
	Comment      string  `json:"comment"`
}

// toCostResponse преобразует модель в ответ API
func toCostResponse(c *models.CostStructure) *CostResponse {
	return &CostResponse{
		ID:           c.ID,
		EnterpriseID: c.EnterpriseID,
		CostItem:     c.CostItem,
		Currency:     c.Currency,
		CostPerT:     c.CostPerT,
		PeriodStart:  c.PeriodStart.Format("2006-01-02"),
		PeriodEnd:    c.PeriodEnd.Format("2006-01-02"),
		Comment:      c.Comment,
	}
}

// CreateCost создаёт новую статью затрат
func (h *CostStructureHandler) CreateCost(w http.ResponseWriter, r *http.Request) {
	// 1. Парсим тело запроса
	var req CreateCostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.ErrorWithCode(w, http.StatusBadRequest, "INVALID_JSON", "Неверный формат JSON: "+err.Error())
		return
	}

	// 2. Валидация данных
	validator := validator.NewValidator()
	validationErrors := validator.MergeErrors(
		validator.Required(req.CostItem, "cost_item"),
		validator.Required(req.Currency, "currency"),
		validator.MinValue(req.CostPerT, 0, "cost_per_t"),
		validator.CurrencyCode(req.Currency, "currency"),
	)

	if validationErrors != nil {
		response.ValidationError(w, validationErrors.ToMap())
		return
	}

	// 3. Парсим даты периода
	periodStart, err := time.Parse("2006-01-02", req.PeriodStart)
	if err != nil {
		validationErrors := map[string]string{"period_start": "Неверный формат даты (ожидается YYYY-MM-DD)"}
		response.ValidationError(w, validationErrors)
		return
	}

	periodEnd, err := time.Parse("2006-01-02", req.PeriodEnd)
	if err != nil {
		validationErrors := map[string]string{"period_end": "Неверный формат даты (ожидается YYYY-MM-DD)"}
		response.ValidationError(w, validationErrors)
		return
	}

	// 4. Создаём модель
	cost := &models.CostStructure{
		EnterpriseID: req.EnterpriseID,
		CostItem:     req.CostItem,
		Currency:     req.Currency,
		CostPerT:     req.CostPerT,
		PeriodStart:  periodStart,
		PeriodEnd:    periodEnd,
		Comment:      req.Comment,
	}

	// 5. Сохраняем в БД
	if err := h.costRepo.Create(cost); err != nil {
		response.ErrorWithCode(w, http.StatusInternalServerError, "DB_ERROR", "Ошибка сохранения статьи затрат: "+err.Error())
		return
	}

	// 6. Формируем успешный ответ
	response.JSON(w, http.StatusCreated, toCostResponse(cost))
}

// ListCosts получает список статей затрат
func (h *CostStructureHandler) ListCosts(w http.ResponseWriter, r *http.Request) {
	enterpriseID := r.URL.Query().Get("enterprise_id")

	if enterpriseID == "" {
		response.ErrorWithCode(w, http.StatusBadRequest, "MISSING_PARAM", "Необходим параметр enterprise_id")
		return
	}

	// Парсим enterprise_id
	var id int64
	_, err := fmt.Sscan(enterpriseID, &id)
	if err != nil || id <= 0 {
		response.ErrorWithCode(w, http.StatusBadRequest, "INVALID_ID", "Некорректный enterprise_id")
		return
	}

	// Получаем статьи затрат из БД
	costs, err := h.costRepo.GetByEnterpriseID(id)
	if err != nil {
		response.ErrorWithCode(w, http.StatusInternalServerError, "DB_ERROR", "Ошибка получения статей затрат: "+err.Error())
		return
	}

	// Преобразуем в ответы API
	responses := make([]*CostResponse, len(costs))
	for i, c := range costs {
		responses[i] = toCostResponse(c)
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"costs": responses,
		"count": len(responses),
	})
}

// GetCost получает статью затрат по ID
func (h *CostStructureHandler) GetCost(w http.ResponseWriter, r *http.Request) {
	costID := r.PathValue("id")
	if costID == "" {
		costID = r.URL.Query().Get("id")
	}

	if costID == "" {
		response.ErrorWithCode(w, http.StatusBadRequest, "MISSING_ID", "Не указан параметр id")
		return
	}

	// Парсим ID
	var id int64
	_, err := fmt.Sscan(costID, &id)
	if err != nil || id <= 0 {
		response.ErrorWithCode(w, http.StatusBadRequest, "INVALID_ID", "Некорректный идентификатор статьи затрат")
		return
	}

	// Получаем контракт из БД
	cost, err := h.costRepo.GetByID(id)
	if err != nil {
		response.ErrorWithCode(w, http.StatusNotFound, "NOT_FOUND", "Статья затрат не найдена: "+err.Error())
		return
	}

	response.JSON(w, http.StatusOK, toCostResponse(cost))
}

// UpdateCost обновляет статью затрат
func (h *CostStructureHandler) UpdateCost(w http.ResponseWriter, r *http.Request) {
	costID := r.PathValue("id")
	if costID == "" {
		costID = r.URL.Query().Get("id")
	}

	if costID == "" {
		response.ErrorWithCode(w, http.StatusBadRequest, "MISSING_ID", "Не указан параметр id")
		return
	}

	// Парсим ID
	var id int64
	_, err := fmt.Sscan(costID, &id)
	if err != nil || id <= 0 {
		response.ErrorWithCode(w, http.StatusBadRequest, "INVALID_ID", "Некорректный идентификатор статьи затрат")
		return
	}

	// Получаем существующую статью затрат
	cost, err := h.costRepo.GetByID(id)
	if err != nil {
		response.ErrorWithCode(w, http.StatusNotFound, "NOT_FOUND", "Статья затрат не найдена: "+err.Error())
		return
	}

	// Парсим тело запроса
	var req CreateCostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.ErrorWithCode(w, http.StatusBadRequest, "INVALID_JSON", "Неверный формат JSON: "+err.Error())
		return
	}

	// Валидация (аналогично CreateCost)
	validator := validator.NewValidator()
	validationErrors := validator.MergeErrors(
		validator.Required(req.CostItem, "cost_item"),
		validator.Required(req.Currency, "currency"),
		validator.MinValue(req.CostPerT, 0, "cost_per_t"),
		validator.CurrencyCode(req.Currency, "currency"),
	)

	if validationErrors != nil {
		response.ValidationError(w, validationErrors.ToMap())
		return
	}

	// Парсим даты
	periodStart, err := time.Parse("2006-01-02", req.PeriodStart)
	if err != nil {
		validationErrors := map[string]string{"period_start": "Неверный формат даты"}
		response.ValidationError(w, validationErrors)
		return
	}

	periodEnd, err := time.Parse("2006-01-02", req.PeriodEnd)
	if err != nil {
		validationErrors := map[string]string{"period_end": "Неверный формат даты"}
		response.ValidationError(w, validationErrors)
		return
	}

	// Обновляем поля
	cost.EnterpriseID = req.EnterpriseID
	cost.CostItem = req.CostItem
	cost.Currency = req.Currency
	cost.CostPerT = req.CostPerT
	cost.PeriodStart = periodStart
	cost.PeriodEnd = periodEnd
	cost.Comment = req.Comment

	// Сохраняем в БД
	if err := h.costRepo.Update(cost); err != nil {
		response.ErrorWithCode(w, http.StatusInternalServerError, "DB_ERROR", "Ошибка обновления статьи затрат: "+err.Error())
		return
	}

	response.JSON(w, http.StatusOK, toCostResponse(cost))
}

// DeleteCost удаляет статью затрат
func (h *CostStructureHandler) DeleteCost(w http.ResponseWriter, r *http.Request) {
	costID := r.PathValue("id")
	if costID == "" {
		costID = r.URL.Query().Get("id")
	}

	if costID == "" {
		response.ErrorWithCode(w, http.StatusBadRequest, "MISSING_ID", "Не указан параметр id")
		return
	}

	// Парсим ID
	var id int64
	_, err := fmt.Sscan(costID, &id)
	if err != nil || id <= 0 {
		response.ErrorWithCode(w, http.StatusBadRequest, "INVALID_ID", "Некорректный идентификатор статьи затрат")
		return
	}

	// Удаляем статью затрат из БД
	if err := h.costRepo.Delete(id); err != nil {
		response.ErrorWithCode(w, http.StatusInternalServerError, "DB_ERROR", "Ошибка удаления статьи затрат: "+err.Error())
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{
		"message": "Статья затрат успешно удалена",
	})
}

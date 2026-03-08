package handlers

import (
	"encoding/json"
	"financial-risk-server/internal/domain/models"
	"financial-risk-server/internal/repository/interfaces"
	"financial-risk-server/internal/utils/validator"
	"financial-risk-server/pkg/response"
	"fmt"
	"net/http"
)

// EnterpriseHandler обрабатывает запросы, связанные с предприятиями
type EnterpriseHandler struct {
	enterpriseRepo interfaces.EnterpriseRepository
}

// NewEnterpriseHandler создаёт новый обработчик предприятий
func NewEnterpriseHandler(enterpriseRepo interfaces.EnterpriseRepository) *EnterpriseHandler {
	return &EnterpriseHandler{
		enterpriseRepo: enterpriseRepo,
	}
}

// CreateEnterpriseRequest структура запроса на создание предприятия
type CreateEnterpriseRequest struct {
	Name               string  `json:"name"`
	Industry           string  `json:"industry"`
	AnnualProductionT  float64 `json:"annual_production_t"`
	ExportSharePercent float64 `json:"export_share_percent"`
	MainCurrency       string  `json:"main_currency"`
}

// EnterpriseResponse структура ответа с данными предприятия
type EnterpriseResponse struct {
	ID                 int64   `json:"id"`
	Name               string  `json:"name"`
	Industry           string  `json:"industry"`
	AnnualProductionT  float64 `json:"annual_production_t"`
	ExportSharePercent float64 `json:"export_share_percent"`
	MainCurrency       string  `json:"main_currency"`
	IsExportOriented   bool    `json:"is_export_oriented"`
}

// toResponse преобразует доменную модель в ответ API
func toResponse(e *models.Enterprise) *EnterpriseResponse {
	return &EnterpriseResponse{
		ID:                 e.ID,
		Name:               e.Name,
		Industry:           e.Industry,
		AnnualProductionT:  e.AnnualProductionT,
		ExportSharePercent: e.ExportSharePercent,
		MainCurrency:       e.MainCurrency,
		IsExportOriented:   e.IsExportOriented(),
	}
}

// CreateEnterprise создаёт новое предприятие
func (h *EnterpriseHandler) CreateEnterprise(w http.ResponseWriter, r *http.Request) {
	// 1. Парсим тело запроса
	var req CreateEnterpriseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.ErrorWithCode(w, http.StatusBadRequest, "INVALID_JSON", "Неверный формат JSON: "+err.Error())
		return
	}

	// 2. Валидация данных
	validator := validator.NewValidator()
	validationErrors := validator.MergeErrors(
		validator.Required(req.Name, "name"),
		validator.MinLength(req.Name, 2, "name"),
		validator.MaxLength(req.Name, 255, "name"),
		validator.MinValue(req.AnnualProductionT, 1, "annual_production_t"),
		validator.MaxValue(req.AnnualProductionT, 100000000, "annual_production_t"),
		validator.MinValue(req.ExportSharePercent, 0, "export_share_percent"),
		validator.MaxValue(req.ExportSharePercent, 100, "export_share_percent"),
		validator.CurrencyCode(req.MainCurrency, "main_currency"),
	)

	if validationErrors != nil {
		response.ValidationError(w, validationErrors.ToMap())
		return
	}

	// 3. Создаём доменную модель
	enterprise := &models.Enterprise{
		Name:               req.Name,
		Industry:           req.Industry,
		AnnualProductionT:  req.AnnualProductionT,
		ExportSharePercent: req.ExportSharePercent,
		MainCurrency:       req.MainCurrency,

		// 5. Формируем успешный ответ
	}

	// 4. Сохраняем в БД
	if err := h.enterpriseRepo.Create(enterprise); err != nil {
		// Обработка ошибки дублирования (если имя уже существует)
		if err.Error() == "предприятие не найдено" {
			response.ErrorWithCode(w, http.StatusConflict, "DUPLICATE_ENTERPRISE", "Предприятие с таким именем уже существует")
			return
		}
		response.ErrorWithCode(w, http.StatusInternalServerError, "DB_ERROR", "Ошибка сохранения предприятия: "+err.Error())
		return
	}

	// 5. Формируем успешный ответ
	response.JSON(w, http.StatusCreated, toResponse(enterprise))
}

// GetEnterprise получает предприятие по ID
func (h *EnterpriseHandler) GetEnterprise(w http.ResponseWriter, r *http.Request) {
	// Получаем ID из переменных маршрута (будет реализовано в маршрутизаторе)
	// Для упрощения сейчас используем параметр из URL
	// В реальном приложении используйте маршрутизатор с поддержкой переменных (chi, gorilla/mux)

	// Пример: /api/v1/enterprises/1
	// В этом упрощённом примере мы будем получать ID из строки запроса
	enterpriseID := r.PathValue("id")
	if enterpriseID == "" {
		// Попробуем получить из query параметра
		enterpriseID = r.URL.Query().Get("id")
	}

	if enterpriseID == "" {
		response.ErrorWithCode(w, http.StatusBadRequest, "MISSING_ID", "Не указан параметр id")
		return
	}

	// Парсим ID
	var id int64
	_, err := fmt.Sscan(enterpriseID, &id)
	if err != nil || id <= 0 {
		response.ErrorWithCode(w, http.StatusBadRequest, "INVALID_ID", "Некорректный идентификатор предприятия")
		return
	}

	// Получаем предприятие из БД
	enterprise, err := h.enterpriseRepo.GetByID(id)
	if err != nil {
		response.ErrorWithCode(w, http.StatusNotFound, "NOT_FOUND", "Предприятие не найдено: "+err.Error())
		return
	}

	response.JSON(w, http.StatusOK, toResponse(enterprise))
}

// ListEnterprises получает список всех предприятий
func (h *EnterpriseHandler) ListEnterprises(w http.ResponseWriter, r *http.Request) {
	enterprises, err := h.enterpriseRepo.GetAll()
	if err != nil {
		response.ErrorWithCode(w, http.StatusInternalServerError, "DB_ERROR", "Ошибка получения списка предприятий: "+err.Error())
		return
	}

	// Преобразуем список в ответы API
	responses := make([]*EnterpriseResponse, len(enterprises))
	for i, e := range enterprises {
		responses[i] = toResponse(e)
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"enterprises": responses,
		"count":       len(responses),
	})
}

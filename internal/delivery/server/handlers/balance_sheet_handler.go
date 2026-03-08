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

// BalanceSheetHandler обрабатывает запросы, связанные с финансовыми балансами
type BalanceSheetHandler struct {
	balanceRepo interfaces.BalanceSheetRepository
}

// NewBalanceSheetHandler создаёт новый обработчик финансовых балансов
func NewBalanceSheetHandler(balanceRepo interfaces.BalanceSheetRepository) *BalanceSheetHandler {
	return &BalanceSheetHandler{
		balanceRepo: balanceRepo,
	}
}

// CreateBalanceRequest структура запроса на создание баланса
type CreateBalanceRequest struct {
	EnterpriseID       int64   `json:"enterprise_id"`
	ReportDate         string  `json:"report_date"` // YYYY-MM-DD
	CashBYN            float64 `json:"cash_byn"`
	CashUSD            float64 `json:"cash_usd"`
	AccountsReceivable float64 `json:"accounts_receivable"`
	Inventories        float64 `json:"inventories"`
	AccountsPayable    float64 `json:"accounts_payable"`
	ShortTermDebt      float64 `json:"short_term_debt"`
	LongTermDebt       float64 `json:"long_term_debt"`
}

// BalanceResponse структура ответа с данными баланса
type BalanceResponse struct {
	ID                 int64   `json:"id"`
	EnterpriseID       int64   `json:"enterprise_id"`
	ReportDate         string  `json:"report_date"`
	CashBYN            float64 `json:"cash_byn"`
	CashUSD            float64 `json:"cash_usd"`
	AccountsReceivable float64 `json:"accounts_receivable"`
	Inventories        float64 `json:"inventories"`
	AccountsPayable    float64 `json:"accounts_payable"`
	ShortTermDebt      float64 `json:"short_term_debt"`
	LongTermDebt       float64 `json:"long_term_debt"`
	CurrentRatio       float64 `json:"current_ratio"` // Текущая ликвидность
	QuickRatio         float64 `json:"quick_ratio"`   // Быстрая ликвидность
}

// toBalanceResponse преобразует модель в ответ API
func toBalanceResponse(b *models.BalanceSheet) *BalanceResponse {
	return &BalanceResponse{
		ID:                 b.ID,
		EnterpriseID:       b.EnterpriseID,
		ReportDate:         b.ReportDate.Format("2006-01-02"),
		CashBYN:            b.CashBYN,
		CashUSD:            b.CashUSD,
		AccountsReceivable: b.AccountsReceivable,
		Inventories:        b.Inventories,
		AccountsPayable:    b.AccountsPayable,
		ShortTermDebt:      b.ShortTermDebt,
		LongTermDebt:       b.LongTermDebt,
		CurrentRatio:       b.GetCurrentRatio(),
		QuickRatio:         b.GetQuickRatio(),
	}
}

// CreateBalance создаёт новый финансовый баланс
func (h *BalanceSheetHandler) CreateBalance(w http.ResponseWriter, r *http.Request) {
	// 1. Парсим тело запроса
	var req CreateBalanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.ErrorWithCode(w, http.StatusBadRequest, "INVALID_JSON", "Неверный формат JSON: "+err.Error())
		return
	}

	// 2. Валидация данных
	validator := validator.NewValidator()
	validationErrors := validator.MergeErrors(
		validator.Required(req.ReportDate, "report_date"),
		validator.MinValue(req.CashBYN, 0, "cash_byn"),
		validator.MinValue(req.CashUSD, 0, "cash_usd"),
		validator.MinValue(req.AccountsReceivable, 0, "accounts_receivable"),
		validator.MinValue(req.Inventories, 0, "inventories"),
		validator.MinValue(req.AccountsPayable, 0, "accounts_payable"),
		validator.MinValue(req.ShortTermDebt, 0, "short_term_debt"),
		validator.MinValue(req.LongTermDebt, 0, "long_term_debt"),
	)

	if validationErrors != nil {
		response.ValidationError(w, validationErrors.ToMap())
		return
	}

	// 3. Парсим дату отчёта
	reportDate, err := time.Parse("2006-01-02", req.ReportDate)
	if err != nil {
		validationErrors := map[string]string{"report_date": "Неверный формат даты (ожидается YYYY-MM-DD)"}
		response.ValidationError(w, validationErrors)
		return
	}

	// 4. Создаём модель
	balance := &models.BalanceSheet{
		EnterpriseID:       req.EnterpriseID,
		ReportDate:         reportDate,
		CashBYN:            req.CashBYN,
		CashUSD:            req.CashUSD,
		AccountsReceivable: req.AccountsReceivable,
		Inventories:        req.Inventories,
		AccountsPayable:    req.AccountsPayable,
		ShortTermDebt:      req.ShortTermDebt,
		LongTermDebt:       req.LongTermDebt,
	}

	// 5. Сохраняем в БД
	if err := h.balanceRepo.Create(balance); err != nil {
		response.ErrorWithCode(w, http.StatusInternalServerError, "DB_ERROR", "Ошибка сохранения баланса: "+err.Error())
		return
	}

	// 6. Формируем успешный ответ
	response.JSON(w, http.StatusCreated, toBalanceResponse(balance))
}

// ListBalances получает список балансов
func (h *BalanceSheetHandler) ListBalances(w http.ResponseWriter, r *http.Request) {
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

	// Получаем балансы из БД
	balances, err := h.balanceRepo.GetByEnterpriseID(id)
	if err != nil {
		response.ErrorWithCode(w, http.StatusInternalServerError, "DB_ERROR", "Ошибка получения балансов: "+err.Error())
		return
	}

	// Преобразуем в ответы API
	responses := make([]*BalanceResponse, len(balances))
	for i, b := range balances {
		responses[i] = toBalanceResponse(b)
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"balances": responses,
		"count":    len(responses),
	})
}

// GetBalance получает баланс по ID
func (h *BalanceSheetHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	balanceID := r.PathValue("id")
	if balanceID == "" {
		balanceID = r.URL.Query().Get("id")
	}

	if balanceID == "" {
		response.ErrorWithCode(w, http.StatusBadRequest, "MISSING_ID", "Не указан параметр id")
		return
	}

	// Парсим ID
	var id int64
	_, err := fmt.Sscan(balanceID, &id)
	if err != nil || id <= 0 {
		response.ErrorWithCode(w, http.StatusBadRequest, "INVALID_ID", "Некорректный идентификатор баланса")
		return
	}

	// Получаем контракт из БД
	balance, err := h.balanceRepo.GetByID(id)
	if err != nil {
		response.ErrorWithCode(w, http.StatusNotFound, "NOT_FOUND", "Статья затрат не найдена: "+err.Error())
		return
	}

	response.JSON(w, http.StatusOK, toBalanceResponse(balance))
}

// GetLatestBalance получает последний баланс предприятия
func (h *BalanceSheetHandler) GetLatestBalance(w http.ResponseWriter, r *http.Request) {
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

	// Получаем последний баланс из БД
	balance, err := h.balanceRepo.GetLatest(id)
	if err != nil {
		response.ErrorWithCode(w, http.StatusNotFound, "NOT_FOUND", "Баланс не найден: "+err.Error())
		return
	}

	response.JSON(w, http.StatusOK, toBalanceResponse(balance))
}

// UpdateBalance обновляет баланс
func (h *BalanceSheetHandler) UpdateBalance(w http.ResponseWriter, r *http.Request) {
	balanceID := r.PathValue("id")
	if balanceID == "" {
		balanceID = r.URL.Query().Get("id")
	}

	if balanceID == "" {
		response.ErrorWithCode(w, http.StatusBadRequest, "MISSING_ID", "Не указан параметр id")
		return
	}

	var id int64
	_, err := fmt.Sscan(balanceID, &id)
	if err != nil || id <= 0 {
		response.ErrorWithCode(w, http.StatusBadRequest, "INVALID_ID", "Некорректный идентификатор баланса")
		return
	}

	// Получаем существующий баланс
	balance, err := h.balanceRepo.GetByID(id)
	if err != nil {
		response.ErrorWithCode(w, http.StatusNotFound, "NOT_FOUND", "Баланс не найден: "+err.Error())
		return
	}

	// Парсим тело запроса
	var req CreateBalanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.ErrorWithCode(w, http.StatusBadRequest, "INVALID_JSON", "Неверный формат JSON: "+err.Error())
		return
	}

	// Валидация (аналогично Create)
	validator := validator.NewValidator()
	validationErrors := validator.MergeErrors(
		validator.Required(req.ReportDate, "report_date"),
		validator.MinValue(req.CashBYN, 0, "cash_byn"),
		validator.MinValue(req.CashUSD, 0, "cash_usd"),
		validator.MinValue(req.AccountsReceivable, 0, "accounts_receivable"),
		validator.MinValue(req.Inventories, 0, "inventories"),
		validator.MinValue(req.AccountsPayable, 0, "accounts_payable"),
		validator.MinValue(req.ShortTermDebt, 0, "short_term_debt"),
		validator.MinValue(req.LongTermDebt, 0, "long_term_debt"),
	)

	if validationErrors != nil {
		response.ValidationError(w, validationErrors.ToMap())
		return
	}

	// Парсим дату отчёта
	reportDate, err := time.Parse("2006-01-02", req.ReportDate)
	if err != nil {
		validationErrors := map[string]string{"report_date": "Неверный формат даты (ожидается YYYY-MM-DD)"}
		response.ValidationError(w, validationErrors)
		return
	}

	// Обновляем поля
	balance.EnterpriseID = req.EnterpriseID
	balance.ReportDate = reportDate
	balance.CashBYN = req.CashBYN
	balance.CashUSD = req.CashUSD
	balance.AccountsReceivable = req.AccountsReceivable
	balance.Inventories = req.Inventories
	balance.AccountsPayable = req.AccountsPayable
	balance.ShortTermDebt = req.ShortTermDebt
	balance.LongTermDebt = req.LongTermDebt

	// Сохраняем в БД
	if err := h.balanceRepo.Update(balance); err != nil {
		response.ErrorWithCode(w, http.StatusInternalServerError, "DB_ERROR", "Ошибка обновления баланса: "+err.Error())
		return
	}

	response.JSON(w, http.StatusOK, toBalanceResponse(balance))
}

// DeleteBalance удаляет баланс
func (h *BalanceSheetHandler) DeleteBalance(w http.ResponseWriter, r *http.Request) {
	balanceID := r.PathValue("id")
	if balanceID == "" {
		balanceID = r.URL.Query().Get("id")
	}

	if balanceID == "" {
		response.ErrorWithCode(w, http.StatusBadRequest, "MISSING_ID", "Не указан параметр id")
		return
	}

	// Парсим ID
	var id int64
	_, err := fmt.Sscan(balanceID, &id)
	if err != nil || id <= 0 {
		response.ErrorWithCode(w, http.StatusBadRequest, "INVALID_ID", "Некорректный идентификатор баланса")
		return
	}

	// Удаляем контракт из БД
	if err := h.balanceRepo.Delete(id); err != nil {
		response.ErrorWithCode(w, http.StatusInternalServerError, "DB_ERROR", "Ошибка удаления контракта: "+err.Error())
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{
		"message": "Контракт успешно удалён",
	})
}

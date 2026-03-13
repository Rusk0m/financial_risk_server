package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"financial-risk-server/internal/domain/models"
	"financial-risk-server/internal/repository/interfaces"
	"financial-risk-server/pkg/response"
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

	// 2. Валидация данных (простая)
	if req.ReportDate == "" {
		response.ValidationError(w, "report_date", "Дата отчёта обязательна")
		return
	}
	if req.CashBYN < 0 || req.CashUSD < 0 || req.AccountsReceivable < 0 ||
		req.Inventories < 0 || req.AccountsPayable < 0 ||
		req.ShortTermDebt < 0 || req.LongTermDebt < 0 {
		response.ErrorWithCode(w, http.StatusBadRequest, "INVALID_VALUE", "Все финансовые показатели должны быть неотрицательными")
		return
	}

	// 3. Парсим дату отчёта
	reportDate, err := time.Parse("2006-01-02", req.ReportDate)
	if err != nil {
		response.ValidationError(w, "report_date", "Неверный формат даты (ожидается YYYY-MM-DD)")
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
	if err := h.balanceRepo.Create(r.Context(), balance); err != nil {
		response.ErrorWithCode(w, http.StatusInternalServerError, "DB_ERROR", "Ошибка сохранения баланса: "+err.Error())
		return
	}

	// 6. Формируем успешный ответ
	response.JSON(w, http.StatusCreated, toBalanceResponse(balance))
}

// ListBalances получает список балансов
func (h *BalanceSheetHandler) ListBalances(w http.ResponseWriter, r *http.Request) {
	enterpriseIDStr := r.URL.Query().Get("enterprise_id")
	if enterpriseIDStr == "" {
		response.ErrorWithCode(w, http.StatusBadRequest, "MISSING_PARAM", "Необходим параметр enterprise_id")
		return
	}

	// Парсим enterprise_id
	enterpriseID, err := strconv.ParseInt(enterpriseIDStr, 10, 64)
	if err != nil || enterpriseID <= 0 {
		response.ErrorWithCode(w, http.StatusBadRequest, "INVALID_ID", "Некорректный enterprise_id")
		return
	}

	// Получаем балансы из БД
	balances, err := h.balanceRepo.GetByEnterpriseID(r.Context(), enterpriseID)
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
	balanceIDStr := r.PathValue("id")
	if balanceIDStr == "" {
		balanceIDStr = r.URL.Query().Get("id")
	}

	if balanceIDStr == "" {
		response.ErrorWithCode(w, http.StatusBadRequest, "MISSING_ID", "Не указан параметр id")
		return
	}

	// Парсим ID
	balanceID, err := strconv.ParseInt(balanceIDStr, 10, 64)
	if err != nil || balanceID <= 0 {
		response.ErrorWithCode(w, http.StatusBadRequest, "INVALID_ID", "Некорректный идентификатор баланса")
		return
	}

	// Получаем баланс из БД
	balance, err := h.balanceRepo.GetByID(r.Context(), balanceID)
	if err != nil {
		response.ErrorWithCode(w, http.StatusNotFound, "NOT_FOUND", "Баланс не найден: "+err.Error())
		return
	}

	response.JSON(w, http.StatusOK, toBalanceResponse(balance))
}

// GetLatestBalance получает последний баланс предприятия
func (h *BalanceSheetHandler) GetLatestBalance(w http.ResponseWriter, r *http.Request) {
	enterpriseIDStr := r.URL.Query().Get("enterprise_id")
	if enterpriseIDStr == "" {
		response.ErrorWithCode(w, http.StatusBadRequest, "MISSING_PARAM", "Необходим параметр enterprise_id")
		return
	}

	// Парсим enterprise_id
	enterpriseID, err := strconv.ParseInt(enterpriseIDStr, 10, 64)
	if err != nil || enterpriseID <= 0 {
		response.ErrorWithCode(w, http.StatusBadRequest, "INVALID_ID", "Некорректный enterprise_id")
		return
	}

	// Получаем последний баланс из БД
	balance, err := h.balanceRepo.GetLatest(r.Context(), enterpriseID)
	if err != nil {
		response.ErrorWithCode(w, http.StatusNotFound, "NOT_FOUND", "Баланс не найден: "+err.Error())
		return
	}

	response.JSON(w, http.StatusOK, toBalanceResponse(balance))
}

// UpdateBalance обновляет баланс
func (h *BalanceSheetHandler) UpdateBalance(w http.ResponseWriter, r *http.Request) {
	balanceIDStr := r.PathValue("id")
	if balanceIDStr == "" {
		balanceIDStr = r.URL.Query().Get("id")
	}

	if balanceIDStr == "" {
		response.ErrorWithCode(w, http.StatusBadRequest, "MISSING_ID", "Не указан параметр id")
		return
	}

	// Парсим ID
	balanceID, err := strconv.ParseInt(balanceIDStr, 10, 64)
	if err != nil || balanceID <= 0 {
		response.ErrorWithCode(w, http.StatusBadRequest, "INVALID_ID", "Некорректный идентификатор баланса")
		return
	}

	// Получаем существующий баланс
	balance, err := h.balanceRepo.GetByID(r.Context(), balanceID)
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
	if req.ReportDate == "" {
		response.ValidationError(w, "report_date", "Дата отчёта обязательна")
		return
	}
	if req.CashBYN < 0 || req.CashUSD < 0 || req.AccountsReceivable < 0 ||
		req.Inventories < 0 || req.AccountsPayable < 0 ||
		req.ShortTermDebt < 0 || req.LongTermDebt < 0 {
		response.ErrorWithCode(w, http.StatusBadRequest, "INVALID_VALUE", "Все финансовые показатели должны быть неотрицательными")
		return
	}

	// Парсим дату отчёта
	reportDate, err := time.Parse("2006-01-02", req.ReportDate)
	if err != nil {
		response.ValidationError(w, "report_date", "Неверный формат даты (ожидается YYYY-MM-DD)")
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
	if err := h.balanceRepo.Update(r.Context(), balance); err != nil {
		response.ErrorWithCode(w, http.StatusInternalServerError, "DB_ERROR", "Ошибка обновления баланса: "+err.Error())
		return
	}

	response.JSON(w, http.StatusOK, toBalanceResponse(balance))
}

// DeleteBalance удаляет баланс
func (h *BalanceSheetHandler) DeleteBalance(w http.ResponseWriter, r *http.Request) {
	balanceIDStr := r.PathValue("id")
	if balanceIDStr == "" {
		balanceIDStr = r.URL.Query().Get("id")
	}

	if balanceIDStr == "" {
		response.ErrorWithCode(w, http.StatusBadRequest, "MISSING_ID", "Не указан параметр id")
		return
	}

	// Парсим ID
	balanceID, err := strconv.ParseInt(balanceIDStr, 10, 64)
	if err != nil || balanceID <= 0 {
		response.ErrorWithCode(w, http.StatusBadRequest, "INVALID_ID", "Некорректный идентификатор баланса")
		return
	}

	// Удаляем баланс из БД
	if err := h.balanceRepo.Delete(r.Context(), balanceID); err != nil {
		response.ErrorWithCode(w, http.StatusInternalServerError, "DB_ERROR", "Ошибка удаления баланса: "+err.Error())
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{
		"message": "Баланс успешно удалён",
	})
}
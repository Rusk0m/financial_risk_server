package handlers

import (
	"financial-risk-server/internal/domain/models"
	"financial-risk-server/internal/repository/interfaces"
	"financial-risk-server/internal/service"
	"financial-risk-server/pkg/response"
	"net/http"
	"strconv"
	"time"
)

// RiskHandler обрабатывает запросы, связанные с расчётами рисков
type RiskHandler struct {
	riskService *service.RiskCalculationService
}

// NewRiskHandler создаёт новый хэндлер расчётов рисков
func NewRiskHandler(riskService *service.RiskCalculationService) *RiskHandler {
	return &RiskHandler{
		riskService: riskService,
	}
}

// GetRiskCalculations возвращает список расчётов рисков с фильтрацией
func (h *RiskHandler) GetRiskCalculations(w http.ResponseWriter, r *http.Request) {
	filter := interfaces.RiskCalculationFilter{
		Limit:  50,
		Offset: 0,
	}

	// Фильтрация по предприятию
	if enterpriseIDStr := r.URL.Query().Get("enterprise_id"); enterpriseIDStr != "" {
		id, err := strconv.ParseInt(enterpriseIDStr, 10, 64)
		if err == nil && id > 0 {
			filter.EnterpriseID = &id
		}
	}

	// Фильтрация по типу риска
	if riskType := r.URL.Query().Get("risk_type"); riskType != "" {
		filter.RiskType = &riskType
	}

	// Фильтрация по дате
	if startDate := r.URL.Query().Get("start_date"); startDate != "" {
		// Валидация формата даты
		_, err := time.Parse("2006-01-02", startDate)
		if err == nil {
			filter.StartDate = &startDate
		}
	}
	if endDate := r.URL.Query().Get("end_date"); endDate != "" {
		_, err := time.Parse("2006-01-02", endDate)
		if err == nil {
			filter.EndDate = &endDate
		}
	}

	// Пагинация
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err == nil && limit > 0 && limit <= 100 {
			filter.Limit = limit
		}
	}
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err == nil && offset >= 0 {
			filter.Offset = offset
		}
	}

	// Получаем расчёты из сервиса
	calculations, err := h.riskService.GetAllRiskCalculations(r.Context(), filter)
	if err != nil {
		response.InternalServerError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"calculations": calculations,
			"count":        len(calculations),
			"total":        len(calculations), // Для упрощения (в реальном приложении нужно отдельное поле для общего количества)
			"limit":        filter.Limit,
			"offset":       filter.Offset,
		},
	})
}

// GetRiskCalculationByID возвращает расчёт риска по ID
func (h *RiskHandler) GetRiskCalculationByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.ValidationError(w, "id", "Invalid risk calculation ID")
		return
	}

	// Получаем расчёт из сервиса
	calculation, err := h.riskService.GetRiskCalculationByID(r.Context(), id)
	if err != nil {
		response.NotFound(w)
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    calculation,
	})
}

// CalculateRisks рассчитывает риски для заданного типа
func (h *RiskHandler) CalculateRisks(w http.ResponseWriter, r *http.Request) {
	// Получаем параметры из query string
	enterpriseIDStr := r.URL.Query().Get("enterprise_id")
	if enterpriseIDStr == "" {
		enterpriseIDStr = "1" // По умолчанию для одного предприятия
	}
	
	enterpriseID, err := strconv.ParseInt(enterpriseIDStr, 10, 64)
	if err != nil || enterpriseID <= 0 {
		response.ErrorWithCode(w, http.StatusBadRequest, "INVALID_ENTERPRISE_ID", "Valid enterprise_id is required")
		return
	}

	riskType := r.URL.Query().Get("risk_type")
	if riskType == "" {
		response.ErrorWithCode(w, http.StatusBadRequest, "RISK_TYPE_REQUIRED", "risk_type parameter is required")
		return
	}

	// Валидация типа риска
	validTypes := map[string]bool{
		"currency":   true,
		"interest":   true,
		"liquidity":  true,
	}
	if !validTypes[riskType] {
		response.ErrorWithCode(w, http.StatusBadRequest, "INVALID_RISK_TYPE", 
			"Invalid risk type. Must be one of: currency, interest, liquidity")
		return
	}

	horizonDaysStr := r.URL.Query().Get("horizon_days")
	horizonDays := 30 // По умолчанию 30 дней
	if horizonDaysStr != "" {
		h, err := strconv.Atoi(horizonDaysStr)
		if err == nil && h > 0 && h <= 365 {
			horizonDays = h
		}
	}

	confidenceLevelStr := r.URL.Query().Get("confidence_level")
	confidenceLevel := 0.95 // По умолчанию 95%
	if confidenceLevelStr != "" {
		c, err := strconv.ParseFloat(confidenceLevelStr, 64)
		if err == nil && c >= 0.90 && c <= 0.99 {
			confidenceLevel = c
		}
	}

	// Вызываем соответствующий метод расчёта
	var calculation *models.RiskCalculation
	switch riskType {
	case "currency":
		calculation, err = h.riskService.CalculateCurrencyRisk(r.Context(), enterpriseID, horizonDays, confidenceLevel)
	case "interest":
		calculation, err = h.riskService.CalculateInterestRisk(r.Context(), enterpriseID, horizonDays, confidenceLevel)
	case "liquidity":
		calculation, err = h.riskService.CalculateLiquidityRisk(r.Context(), enterpriseID, horizonDays, confidenceLevel)
	}

	if err != nil {
		response.InternalServerError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    calculation,
		"message": "Risk calculated successfully",
	})
}

// CalculateAllRisks рассчитывает все три типа рисков
func (h *RiskHandler) CalculateAllRisks(w http.ResponseWriter, r *http.Request) {
	// Получаем параметры из query string
	enterpriseIDStr := r.URL.Query().Get("enterprise_id")
	if enterpriseIDStr == "" {
		enterpriseIDStr = "1" // По умолчанию для одного предприятия
	}
	
	enterpriseID, err := strconv.ParseInt(enterpriseIDStr, 10, 64)
	if err != nil || enterpriseID <= 0 {
		response.ErrorWithCode(w, http.StatusBadRequest, "INVALID_ENTERPRISE_ID", "Valid enterprise_id is required")
		return
	}

	horizonDaysStr := r.URL.Query().Get("horizon_days")
	horizonDays := 30 // По умолчанию 30 дней
	if horizonDaysStr != "" {
		h, err := strconv.Atoi(horizonDaysStr)
		if err == nil && h > 0 && h <= 365 {
			horizonDays = h
		}
	}

	confidenceLevelStr := r.URL.Query().Get("confidence_level")
	confidenceLevel := 0.95 // По умолчанию 95%
	if confidenceLevelStr != "" {
		c, err := strconv.ParseFloat(confidenceLevelStr, 64)
		if err == nil && c >= 0.90 && c <= 0.99 {
			confidenceLevel = c
		}
	}

	// Вызываем сервис для расчёта всех рисков
	results, err := h.riskService.CalculateAllRisks(r.Context(), enterpriseID, horizonDays, confidenceLevel)
	if err != nil {
		response.InternalServerError(w, err)
		return
	}

	// Формируем ответ с агрегированными данными
	summary := h.buildRiskSummary(results)

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"currency_risk":    results["currency"],
			"interest_risk":    results["interest"],
			"liquidity_risk":   results["liquidity"],
			"summary":          summary,
			"calculation_date": time.Now().Format(time.RFC3339),
		},
		"message": "All risks calculated successfully",
	})
}

// buildRiskSummary формирует сводку по всем рискам
func (h *RiskHandler) buildRiskSummary(results map[string]*models.RiskCalculation) map[string]interface{} {
	totalRisk := results["currency"].VaRValue + results["interest"].VaRValue + results["liquidity"].VaRValue
	
	// Определяем максимальный риск
	maxRisk := results["currency"]
	maxRiskType := "currency"
	
	if results["interest"].VaRValue > maxRisk.VaRValue {
		maxRisk = results["interest"]
		maxRiskType = "interest"
	}
	if results["liquidity"].VaRValue > maxRisk.VaRValue {
		maxRisk = results["liquidity"]
		maxRiskType = "liquidity"
	}
	
	// Определяем общий уровень риска
	overallRiskLevel := "low"
	if maxRisk.RiskLevel == "high" {
		overallRiskLevel = "high"
	} else if maxRisk.RiskLevel == "medium" {
		overallRiskLevel = "medium"
	}
	
	return map[string]interface{}{
		"total_risk_value": totalRisk,
		"max_risk_type":    maxRiskType,
		"max_risk_value":   maxRisk.VaRValue,
		"overall_risk_level": overallRiskLevel,
		"currency_risk_percentage": (results["currency"].VaRValue / totalRisk) * 100,
		"interest_risk_percentage": (results["interest"].VaRValue / totalRisk) * 100,
		"liquidity_risk_percentage": (results["liquidity"].VaRValue / totalRisk) * 100,
	}
}
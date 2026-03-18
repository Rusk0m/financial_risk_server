package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"financial-risk-server/internal/domain/models"
	"financial-risk-server/internal/repository/interfaces"
	"financial-risk-server/internal/service"
)

// RiskHandler обрабатывает HTTP запросы для расчёта рисков
type RiskHandler struct {
	riskService *service.RiskCalculationService
}

// NewRiskHandler создаёт новый handler
func NewRiskHandler(riskService *service.RiskCalculationService) *RiskHandler {
	return &RiskHandler{
		riskService: riskService,
	}
}

// CalculateAllRisksRequest запрос на расчёт всех рисков
type CalculateAllRisksRequest struct {
	EnterpriseID    int64   `json:"enterprise_id"`
	HorizonDays     int     `json:"horizon_days"`
	ConfidenceLevel float64 `json:"confidence_level"`
}

// RiskCalculationDTO Data Transfer Object для RiskCalculation
type RiskCalculationDTO struct {
	ID                int64      `json:"id"`
	EnterpriseID      int64      `json:"enterprise_id"`
	RiskType          string     `json:"risk_type"`
	CalculationDate   string     `json:"calculation_date"`
	HorizonDays       int        `json:"horizon_days"`
	ConfidenceLevel   float64    `json:"confidence_level"`
	ExposureAmount    float64    `json:"exposure_amount"`
	VaRValue          float64    `json:"var_value"`
	StressTestLoss    float64    `json:"stress_test_loss"`
	RiskLevel         string     `json:"risk_level"`
	RiskPercentage    float64    `json:"risk_percentage"`
	StressPercentage  float64    `json:"stress_percentage"`
	CalculationMethod *string    `json:"calculation_method,omitempty"`
	Recommendations   []string   `json:"recommendations"`
}

// ToDTO конвертирует RiskCalculation в DTO
func ToDTO(risk *models.RiskCalculation) *RiskCalculationDTO {
	return &RiskCalculationDTO{
		ID:                risk.ID,
		EnterpriseID:      risk.EnterpriseID,
		RiskType:          risk.RiskType,
		CalculationDate:   risk.CalculationDate.Format(time.RFC3339),
		HorizonDays:       risk.HorizonDays,
		ConfidenceLevel:   risk.ConfidenceLevel,
		ExposureAmount:    risk.ExposureAmount,
		VaRValue:          risk.VaRValue,
		StressTestLoss:    risk.StressTestLoss,
		RiskLevel:         risk.RiskLevel,
		RiskPercentage:    risk.GetRiskPercentage(),
		StressPercentage:  risk.GetStressTestPercentage(),
		CalculationMethod: risk.CalculationMethod,
		Recommendations:   risk.Recommendations,
	}
}

// Рассчитывает все три типа рисков для предприятия
func (h *RiskHandler) CalculateRisks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req CalculateAllRisksRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	// Валидация
	if req.EnterpriseID <= 0 {
		writeJSONError(w, http.StatusBadRequest, "enterprise_id must be positive")
		return
	}
	if req.HorizonDays <= 0 || req.HorizonDays > 365 {
		writeJSONError(w, http.StatusBadRequest, "horizon_days must be between 1 and 365")
		return
	}
	if req.ConfidenceLevel < 0.90 || req.ConfidenceLevel > 0.99 {
		writeJSONError(w, http.StatusBadRequest, "confidence_level must be between 0.90 and 0.99")
		return
	}

	// Расчёт рисков
	results, err := h.riskService.CalculateAllRisks(
		ctx,
		req.EnterpriseID,
		req.HorizonDays,
		req.ConfidenceLevel,
	)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to calculate risks: "+err.Error())
		return
	}

	// Конвертация в DTO
	data := make(map[string]*RiskCalculationDTO)
	for riskType, risk := range results {
		data[riskType] = ToDTO(risk)
	}

	response := map[string]interface{}{
		"success": true,
		"data":    data,
	}

	writeJSON(w, http.StatusOK, response)
}

// CalculateAllRisks GET /api/v1/risks/all
// Альтернативный endpoint для расчёта (через query params)
func (h *RiskHandler) CalculateAllRisks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	query := r.URL.Query()
	
	// EnterpriseID (обязательный)
	enterpriseIDStr := query.Get("enterprise_id")
	if enterpriseIDStr == "" {
		writeJSONError(w, http.StatusBadRequest, "enterprise_id is required")
		return
	}
	enterpriseID, err := strconv.ParseInt(enterpriseIDStr, 10, 64)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid enterprise_id")
		return
	}

	// HorizonDays (по умолчанию 10)
	horizonDays := 10
	if hd := query.Get("horizon_days"); hd != "" {
		horizonDays, _ = strconv.Atoi(hd)
	}

	// ConfidenceLevel (по умолчанию 0.95)
	confidenceLevel := 0.95
	if cl := query.Get("confidence_level"); cl != "" {
		confidenceLevel, _ = strconv.ParseFloat(cl, 64)
	}

	// Валидация
	if horizonDays <= 0 || horizonDays > 365 {
		writeJSONError(w, http.StatusBadRequest, "horizon_days must be between 1 and 365")
		return
	}
	if confidenceLevel < 0.90 || confidenceLevel > 0.99 {
		writeJSONError(w, http.StatusBadRequest, "confidence_level must be between 0.90 and 0.99")
		return
	}

	// Расчёт рисков
	results, err := h.riskService.CalculateAllRisks(
		ctx,
		enterpriseID,
		horizonDays,
		confidenceLevel,
	)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to calculate risks: "+err.Error())
		return
	}

	// Конвертация в DTO
	data := make(map[string]*RiskCalculationDTO)
	for riskType, risk := range results {
		data[riskType] = ToDTO(risk)
	}

	response := map[string]interface{}{
		"success": true,
		"data":    data,
	}

	writeJSON(w, http.StatusOK, response)
}

// GetRiskCalculations GET /api/v1/risks
// Получает список расчётов с фильтрацией
func (h *RiskHandler) GetRiskCalculations(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	query := r.URL.Query()
	
	var filter interfaces.RiskCalculationFilter

	if enterpriseID := query.Get("enterprise_id"); enterpriseID != "" {
		id, err := strconv.ParseInt(enterpriseID, 10, 64)
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "Invalid enterprise_id")
			return
		}
		filter.EnterpriseID = &id
	}

	if riskType := query.Get("risk_type"); riskType != "" {
		filter.RiskType = &riskType
	}

	if limit := query.Get("limit"); limit != "" {
		l, err := strconv.Atoi(limit)
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "Invalid limit")
			return
		}
		filter.Limit = l
	} else {
		filter.Limit = 100
	}

	if offset := query.Get("offset"); offset != "" {
		o, err := strconv.Atoi(offset)
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "Invalid offset")
			return
		}
		filter.Offset = o
	}

	results, err := h.riskService.GetAllRiskCalculations(ctx, filter)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to get risk calculations: "+err.Error())
		return
	}

	data := make([]*RiskCalculationDTO, 0, len(results))
	for _, risk := range results {
		data = append(data, ToDTO(risk))
	}

	response := map[string]interface{}{
		"success": true,
		"data":    data,
		"total":   len(data),
	}

	writeJSON(w, http.StatusOK, response)
}

// GetRiskCalculationByID GET /api/v1/risks/{id}
// Получает расчёт по ID
func (h *RiskHandler) GetRiskCalculationByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Извлекаем ID из пути (Go 1.22+ синтаксис)
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid risk ID")
		return
	}

	result, err := h.riskService.GetRiskCalculationByID(ctx, id)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "Risk calculation not found")
		return
	}

	response := map[string]interface{}{
		"success": true,
		"data":    ToDTO(result),
	}

	writeJSON(w, http.StatusOK, response)
}

// Вспомогательные функции

// writeJSON записывает JSON ответ
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// writeJSONError записывает JSON ответ об ошибке
func writeJSONError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]interface{}{
		"success": false,
		"error":   message,
		"code":    status,
	})
}
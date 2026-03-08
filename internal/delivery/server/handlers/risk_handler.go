package handlers

import (
	"encoding/json"
	"financial-risk-server/internal/domain/models"
	"financial-risk-server/internal/service"
	"financial-risk-server/pkg/response"
	"net/http"
	"strconv"
)

// RiskHandler обрабатывает запросы, связанные с расчётом рисков
type RiskHandler struct {
	riskService service.RiskCalculationService
}

// NewRiskHandler создаёт новый обработчик расчёта рисков
func NewRiskHandler(riskService service.RiskCalculationService) *RiskHandler {
	return &RiskHandler{
		riskService: riskService,
	}
}

// CalculateRiskRequest структура запроса на расчёт риска
type CalculateRiskRequest struct {
	EnterpriseID    int64   `json:"enterprise_id"`
	RiskType        string  `json:"risk_type"`        // currency, credit, liquidity, market, interest
	HorizonDays     int     `json:"horizon_days"`     // для VaR
	ConfidenceLevel float64 `json:"confidence_level"` // 0.95 или 0.99
	PriceChangePct  float64 `json:"price_change_pct"` // для фондового риска
	RateChangePct   float64 `json:"rate_change_pct"`  // для процентного риска
}

// RiskResultResponse структура ответа с результатом расчёта риска
type RiskResultResponse struct {
	ID              int64    `json:"id"`
	RiskType        string   `json:"risk_type"`
	CalculationDate string   `json:"calculation_date"`
	HorizonDays     int      `json:"horizon_days"`
	ConfidenceLevel float64  `json:"confidence_level"`
	ExposureAmount  float64  `json:"exposure_amount"`
	VaRValue        float64  `json:"var_value"`
	StressTestLoss  float64  `json:"stress_test_loss"`
	RiskLevel       string   `json:"risk_level"`
	Recommendations []string `json:"recommendations"`
	ScenarioType    string   `json:"scenario_type"`
}

// toRiskResponse преобразует доменную модель в ответ API
func toRiskResponse(r *models.RiskResult) *RiskResultResponse {
	if r == nil {
		return nil
	}

	return &RiskResultResponse{
		ID:              r.ID,
		RiskType:        string(r.RiskType),
		CalculationDate: r.CalculationDate.Format("2006-01-02 15:04:05"),
		HorizonDays:     r.HorizonDays,
		ConfidenceLevel: r.ConfidenceLevel,
		ExposureAmount:  r.ExposureAmount,
		VaRValue:        r.VaRValue,
		StressTestLoss:  r.StressTestLoss,
		RiskLevel:       string(r.RiskLevel),
		Recommendations: r.Recommendations,
		ScenarioType:    r.ScenarioType,
	}
}

// CalculateRisk рассчитывает указанный тип риска
func (h *RiskHandler) CalculateRisk(w http.ResponseWriter, r *http.Request) {
	// 1. Парсим тело запроса
	var req CalculateRiskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.ErrorWithCode(w, http.StatusBadRequest, "INVALID_JSON", "Неверный формат JSON: "+err.Error())
		return
	}

	// 2. Валидация базовых параметров
	validationErrors := make(map[string]string)

	if req.EnterpriseID <= 0 {
		validationErrors["enterprise_id"] = "ID предприятия должен быть больше 0"
	}

	// 3. Вызываем соответствующий метод сервиса в зависимости от типа риска
	var result *models.RiskResult
	var err error

	switch req.RiskType {
	case "currency":
		if req.HorizonDays == 0 {
			req.HorizonDays = 30
		}
		if req.ConfidenceLevel == 0 {
			req.ConfidenceLevel = 0.95
		}
		result, err = h.riskService.CalculateCurrencyRisk(
			req.EnterpriseID,
			req.HorizonDays,
			req.ConfidenceLevel,
		)

	case "credit":
		result, err = h.riskService.CalculateCreditRisk(req.EnterpriseID)

	case "liquidity":
		result, err = h.riskService.CalculateLiquidityRisk(req.EnterpriseID)

	case "market":
		if req.PriceChangePct == 0 {
			req.PriceChangePct = -15.0 // пессимистичный сценарий
		}
		result, err = h.riskService.CalculateMarketRisk(
			req.EnterpriseID,
			req.PriceChangePct,
		)

	case "interest":
		if req.RateChangePct == 0 {
			req.RateChangePct = 1.0 // +1% к ставке
		}
		result, err = h.riskService.CalculateInterestRisk(
			req.EnterpriseID,
			req.RateChangePct,
		)

	default:
		validationErrors["risk_type"] = "неизвестный тип риска (допустимые значения: currency, credit, liquidity, market, interest)"
	}

	// Если были ошибки валидации, возвращаем их
	if len(validationErrors) > 0 {
		response.ValidationError(w, validationErrors)
		return
	}

	// Если была ошибка при расчёте
	if err != nil {
		response.ErrorWithCode(w, http.StatusInternalServerError, "RISK_CALCULATION_ERROR", "Ошибка расчёта риска: "+err.Error())
		return
	}

	// 4. Формируем успешный ответ
	response.JSON(w, http.StatusOK, toRiskResponse(result))
}

// CalculateAllRisks рассчитывает все риски для предприятия
func (h *RiskHandler) CalculateAllRisks(w http.ResponseWriter, r *http.Request) {
	// Получаем enterprise_id из query параметра
	enterpriseID, err := strconv.ParseInt(r.URL.Query().Get("enterprise_id"), 10, 64)
	if err != nil || enterpriseID <= 0 {
		response.ValidationError(w, map[string]string{
			"enterprise_id": "некорректный или отсутствующий параметр enterprise_id",
		})
		return
	}

	// Получаем дополнительные параметры из query
	horizonDays := 30
	if hd := r.URL.Query().Get("horizon_days"); hd != "" {
		if h, err := strconv.Atoi(hd); err == nil && h > 0 {
			horizonDays = h
		}
	}

	confidenceLevel := 0.95
	if cl := r.URL.Query().Get("confidence_level"); cl != "" {
		if c, err := strconv.ParseFloat(cl, 64); err == nil && c > 0 && c <= 1 {
			confidenceLevel = c
		}
	}

	priceChangePct := -15.0
	if pc := r.URL.Query().Get("price_change_pct"); pc != "" {
		if p, err := strconv.ParseFloat(pc, 64); err == nil {
			priceChangePct = p
		}
	}

	rateChangePct := 1.0
	if rc := r.URL.Query().Get("rate_change_pct"); rc != "" {
		if r, err := strconv.ParseFloat(rc, 64); err == nil {
			rateChangePct = r
		}
	}

	// Вызываем агрегирующий метод
	report, err := h.riskService.CalculateAllRisks(
		enterpriseID,
		horizonDays,
		confidenceLevel,
		priceChangePct,
		rateChangePct,
	)
	if err != nil {
		response.ErrorWithCode(w, http.StatusInternalServerError, "RISK_CALCULATION_ERROR", "Ошибка расчёта рисков: "+err.Error())
		return
	}

	// Формируем ответ
	// Создаём упрощённую структуру для ответа
	type RiskSummary struct {
		RiskType       string  `json:"risk_type"`
		RiskLevel      string  `json:"risk_level"`
		VaRValue       float64 `json:"var_value"`
		StressTestLoss float64 `json:"stress_test_loss"`
	}

	responseData := map[string]interface{}{
		"enterprise_id":      report.EnterpriseID,
		"report_date":        report.ReportDate.Format("2006-01-02 15:04:05"),
		"overall_risk_level": report.OverallRiskLevel,
		"total_risk_value":   report.TotalRiskValue,
		"max_risk_type":      report.MaxRiskType,
		"summary":            report.Summary,
		"priority_actions":   report.PriorityActions,
		"warnings":           report.Warnings,
	}

	// Добавляем детали по каждому типу риска
	if report.CurrencyRisk != nil {
		responseData["currency_risk"] = RiskSummary{
			RiskType:       string(report.CurrencyRisk.RiskType),
			RiskLevel:      string(report.CurrencyRisk.RiskLevel),
			VaRValue:       report.CurrencyRisk.VaRValue,
			StressTestLoss: report.CurrencyRisk.StressTestLoss,
		}
	}
	if report.CreditRisk != nil {
		responseData["credit_risk"] = RiskSummary{
			RiskType:       string(report.CreditRisk.RiskType),
			RiskLevel:      string(report.CreditRisk.RiskLevel),
			VaRValue:       report.CreditRisk.VaRValue,
			StressTestLoss: report.CreditRisk.StressTestLoss,
		}
	}
	if report.LiquidityRisk != nil {
		responseData["liquidity_risk"] = RiskSummary{
			RiskType:       string(report.LiquidityRisk.RiskType),
			RiskLevel:      string(report.LiquidityRisk.RiskLevel),
			VaRValue:       report.LiquidityRisk.VaRValue,
			StressTestLoss: report.LiquidityRisk.StressTestLoss,
		}
	}
	if report.MarketRisk != nil {
		responseData["market_risk"] = RiskSummary{
			RiskType:       string(report.MarketRisk.RiskType),
			RiskLevel:      string(report.MarketRisk.RiskLevel),
			VaRValue:       report.MarketRisk.VaRValue,
			StressTestLoss: report.MarketRisk.StressTestLoss,
		}
	}
	if report.InterestRisk != nil {
		responseData["interest_risk"] = RiskSummary{
			RiskType:       string(report.InterestRisk.RiskType),
			RiskLevel:      string(report.InterestRisk.RiskLevel),
			VaRValue:       report.InterestRisk.VaRValue,
			StressTestLoss: report.InterestRisk.StressTestLoss,
		}
	}

	response.JSON(w, http.StatusOK, responseData)
}

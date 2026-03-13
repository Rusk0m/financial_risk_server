package handlers

import (
	"financial-risk-server/internal/domain/models"
	"financial-risk-server/internal/service"
	"financial-risk-server/pkg/response"
	"net/http"
	"strconv"
)

// ComprehensiveReportHandler обрабатывает запросы, связанные с комплексными отчётами
type ComprehensiveReportHandler struct {
	comprehensiveReportService *service.ComprehensiveReportService
}

// NewComprehensiveReportHandler создаёт новый хэндлер комплексных отчётов
func NewComprehensiveReportHandler(comprehensiveReportService *service.ComprehensiveReportService) *ComprehensiveReportHandler {
	return &ComprehensiveReportHandler{
		comprehensiveReportService: comprehensiveReportService,
	}
}

// GetReports возвращает список комплексных отчётов с фильтрацией
func (h *ComprehensiveReportHandler) GetReports(w http.ResponseWriter, r *http.Request) {
	filterEnterpriseIDStr := r.URL.Query().Get("enterprise_id")
	var filterEnterpriseID *int64
	
	if filterEnterpriseIDStr != "" {
		id, err := strconv.ParseInt(filterEnterpriseIDStr, 10, 64)
		if err == nil && id > 0 {
			filterEnterpriseID = &id
		}
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 50
	if limitStr != "" {
		l, err := strconv.Atoi(limitStr)
		if err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	offsetStr := r.URL.Query().Get("offset")
	offset := 0
	if offsetStr != "" {
		o, err := strconv.Atoi(offsetStr)
		if err == nil && o >= 0 {
			offset = o
		}
	}

	// Получаем отчёты из сервиса
	var reports []*models.ComprehensiveRiskReport
	var err error
	
	if filterEnterpriseID != nil {
		reports, err = h.comprehensiveReportService.GetAllReportsByEnterpriseID(r.Context(), *filterEnterpriseID)
	} else {
		reports, err = h.comprehensiveReportService.GetAllReports(r.Context())
	}
	
	if err != nil {
		response.InternalServerError(w, err)
		return
	}

	// Применяем пагинацию на уровне приложения
	start := offset
	end := offset + limit
	if start > len(reports) {
		start = len(reports)
	}
	if end > len(reports) {
		end = len(reports)
	}
	
	paginatedReports := reports[start:end]

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"reports": paginatedReports,
			"count":   len(paginatedReports),
			"total":   len(reports),
			"limit":   limit,
			"offset":  offset,
		},
	})
}

// GetReportByID возвращает комплексный отчёт по ID
func (h *ComprehensiveReportHandler) GetReportByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.ValidationError(w, "id", "Invalid report ID")
		return
	}

	// Получаем отчёт из сервиса
	report, err := h.comprehensiveReportService.GetReportByID(r.Context(), id)
	if err != nil {
		response.NotFound(w)
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    report,
	})
}

// GenerateReport генерирует новый комплексный отчёт
func (h *ComprehensiveReportHandler) GenerateReport(w http.ResponseWriter, r *http.Request) {
	enterpriseIDStr := r.URL.Query().Get("enterprise_id")
	if enterpriseIDStr == "" {
		enterpriseIDStr = "1" // По умолчанию для одного предприятия
	}
	
	enterpriseID, err := strconv.ParseInt(enterpriseIDStr, 10, 64)
	if err != nil || enterpriseID <= 0 {
		response.ErrorWithCode(w, http.StatusBadRequest, "INVALID_ENTERPRISE_ID", "Valid enterprise_id is required")
		return
	}

	// Вызываем сервис для генерации отчёта
	report, err := h.comprehensiveReportService.GenerateReport(r.Context(), enterpriseID)
	if err != nil {
		response.InternalServerError(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, map[string]interface{}{
		"success": true,
		"data":    report,
		"message": "Comprehensive risk report generated successfully",
	})
}
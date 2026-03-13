package handlers

import (
	"financial-risk-server/internal/domain/models"
	"financial-risk-server/internal/repository/interfaces"
	"financial-risk-server/internal/service"
	"financial-risk-server/pkg/response"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
)

// ReportHandler обрабатывает запросы, связанные с отчётами
type ReportHandler struct {
	reportService *service.ReportService
}

// NewReportHandler создаёт новый хэндлер отчётов
func NewReportHandler(reportService *service.ReportService) *ReportHandler {
	return &ReportHandler{
		reportService: reportService,
	}
}

// GetReports возвращает список отчётов с фильтрацией
func (h *ReportHandler) GetReports(w http.ResponseWriter, r *http.Request) {
	filter := interfaces.ReportFilter{
		Limit:  50,
		Offset: 0,
	}

	if enterpriseIDStr := r.URL.Query().Get("enterprise_id"); enterpriseIDStr != "" {
		id, err := strconv.ParseInt(enterpriseIDStr, 10, 64)
		if err == nil && id > 0 {
			filter.EnterpriseID = &id
		}
	}

	if reportType := r.URL.Query().Get("report_type"); reportType != "" {
		filter.ReportType = &reportType
	}

	if status := r.URL.Query().Get("status"); status != "" {
		filter.Status = &status
	}

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

	reports, err := h.reportService.GetAllReports(r.Context(), filter)
	if err != nil {
		response.InternalServerError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"reports": reports,
		"count":   len(reports),
		"total":   len(reports), // Для упрощения
		"limit":   filter.Limit,
		"offset":  filter.Offset,
	})
}

// GetReport возвращает отчёт по ID
func (h *ReportHandler) GetReport(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.ValidationError(w, "id", "Invalid report ID")
		return
	}

	report, err := h.reportService.GetReport(r.Context(), id)
	if err != nil {
		response.NotFound(w)
		return
	}

	// Формируем полный URL для скачивания
	baseURL := fmt.Sprintf("%s://%s", r.URL.Scheme, r.Host)
	if r.TLS != nil {
		baseURL = "https://" + r.Host
	} else {
		baseURL = "http://" + r.Host
	}
	report.DownloadURL = fmt.Sprintf("%s/api/v1/reports/%d/download", baseURL, report.ID)

	response.JSON(w, http.StatusOK, report)
}

// UploadReport загружает новый отчёт
func (h *ReportHandler) UploadReport(w http.ResponseWriter, r *http.Request) {
	// Проверяем тип контента
	if r.Header.Get("Content-Type") == "" {
		response.ErrorWithCode(w, http.StatusBadRequest, "INVALID_CONTENT_TYPE", "Content-Type header is required")
		return
	}

	// Парсим multipart/form-data (максимум 32 МБ в памяти)
	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		response.ErrorWithCode(w, http.StatusBadRequest, "INVALID_FORM", "Failed to parse form: "+err.Error())
		return
	}

	// Получаем файл
	file, handler, err := r.FormFile("file")
	if err != nil {
		response.ErrorWithCode(w, http.StatusBadRequest, "FILE_REQUIRED", "File field 'file' is required")
		return
	}
	defer file.Close()

	// Валидация размера файла
	if handler.Size == 0 {
		response.ErrorWithCode(w, http.StatusBadRequest, "EMPTY_FILE", "File cannot be empty")
		return
	}
	if handler.Size > 10*1024*1024 { // 10 МБ
		response.ErrorWithCode(w, http.StatusBadRequest, "FILE_TOO_LARGE", "File size exceeds 10 MB limit")
		return
	}

	// Получаем метаданные из формы
	enterpriseIDStr := r.FormValue("enterprise_id")
	enterpriseID, err := strconv.ParseInt(enterpriseIDStr, 10, 64)
	if err != nil || enterpriseID <= 0 {
		response.ErrorWithCode(w, http.StatusBadRequest, "INVALID_ENTERPRISE_ID", "Valid enterprise_id is required")
		return
	}

	reportType := r.FormValue("report_type")
	if reportType == "" {
		response.ErrorWithCode(w, http.StatusBadRequest, "REPORT_TYPE_REQUIRED", "report_type is required")
		return
	}

	// Создаём запрос на загрузку
	req := &models.ReportUploadRequest{
		EnterpriseID: enterpriseID,
		ReportType:   reportType,
		PeriodStart:  r.FormValue("period_start"),
		PeriodEnd:    r.FormValue("period_end"),
		Description:  r.FormValue("description"),
		UploadedBy:   r.FormValue("uploaded_by"),
	}

	// Вызываем сервис для загрузки (передаём оригинальный файл)
	createdReport, err := h.reportService.UploadReport(r.Context(), handler, req)
	if err != nil {
		response.ErrorWithCode(w, http.StatusInternalServerError, "UPLOAD_FAILED", "Failed to upload report: "+err.Error())
		return
	}

	response.JSON(w, http.StatusCreated, createdReport)
}
// DeleteReport удаляет отчёт по ID
func (h *ReportHandler) DeleteReport(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.ValidationError(w, "id", "Invalid report ID")
		return
	}

	err = h.reportService.DeleteReport(r.Context(), id)
	if err != nil {
		response.InternalServerError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{
		"message": "Report deleted successfully",
	})
}

// DownloadReport скачивает файл отчёта
func (h *ReportHandler) DownloadReport(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.ValidationError(w, "id", "Invalid report ID")
		return
	}

	// Получаем метаданные отчёта
	report, err := h.reportService.GetReport(r.Context(), id)
	if err != nil {
		response.NotFound(w)
		return
	}

	// Проверяем существование файла
	if report.FilePath == "" {
		response.ErrorWithCode(w, http.StatusNotFound, "FILE_NOT_FOUND", "Report file path is not set")
		return
	}

	// Формируем полный путь к файлу
	filePath := report.FilePath
	if !filepath.IsAbs(filePath) {
		filePath = filepath.Join("./uploads", filePath)
	}

	// Проверяем существование файла
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		response.ErrorWithCode(w, http.StatusNotFound, "FILE_NOT_FOUND", "Report file does not exist on server")
		return
	}

	// Открываем файл для чтения
	file, err := os.Open(filePath)
	if err != nil {
		response.InternalServerError(w, fmt.Errorf("failed to open report file: %w", err))
		return
	}
	defer file.Close()

	// Определяем MIME-тип по расширению
	ext := filepath.Ext(report.OriginalName)
	mimeType := "application/octet-stream"
	switch ext {
	case ".xlsx":
		mimeType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case ".xls":
		mimeType = "application/vnd.ms-excel"
	case ".csv":
		mimeType = "text/csv"
	}

	// Устанавливаем заголовки для скачивания
	// Кодируем имя файла для поддержки кириллицы (RFC 5987)
	encodedFileName := url.PathEscape(report.OriginalName)

	w.Header().Set("Content-Type", mimeType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename*=UTF-8''%s", encodedFileName))
	w.Header().Set("Content-Length", strconv.FormatInt(report.FileSizeBytes, 10))
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	// Копируем файл в ответ
	_, err = io.Copy(w, file)
	if err != nil {
		// Ошибка при отправке файла - логируем, но не прерываем
	}
}
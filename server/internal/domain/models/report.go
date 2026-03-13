package models

import (
	"time"
)

// Report представляет метаданные загруженного отчёта
type Report struct {
	ID              int64     `json:"id"`
	EnterpriseID    int64     `json:"enterprise_id"`
	ReportType      string    `json:"report_type"`       // export_contracts, balance_sheet, financial_results, credit_agreements
	FileName        string    `json:"file_name"`         // Уникальное имя файла на сервере
	OriginalName    string    `json:"original_name"`     // Оригинальное имя файла
	FilePath        string    `json:"file_path"`         // Путь к файлу относительно корня сервера
	FileSizeBytes   int64     `json:"file_size_bytes"`
	UploadDate      time.Time `json:"upload_date"`
	ProcessingStatus string   `json:"processing_status"` // pending, processed, error, archived
	ProcessingError string    `json:"processing_error,omitempty"`
	PeriodStart     *time.Time `json:"period_start,omitempty"`
	PeriodEnd       *time.Time `json:"period_end,omitempty"`
	UploadedBy      string    `json:"uploaded_by,omitempty"`
	Description     string    `json:"description,omitempty"`
	DownloadURL     string    `json:"download_url"`      // URL для скачивания файла
		// Метаданные
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ReportUploadRequest представляет запрос на загрузку отчёта
type ReportUploadRequest struct {
	EnterpriseID int64  `form:"enterprise_id" binding:"required"`
	ReportType   string `form:"report_type" binding:"required,oneof=export_contracts balance_sheet financial_results credit_agreements"`
	PeriodStart  string `form:"period_start"` // YYYY-MM-DD
	PeriodEnd    string `form:"period_end"`   // YYYY-MM-DD
	Description  string `form:"description"`
	UploadedBy   string `form:"uploaded_by"`
}

// ReportFilter представляет параметры фильтрации отчётов
type ReportFilter struct {
	EnterpriseID *int64
	ReportType   *string
	Status       *string
	PeriodStart  *time.Time
	PeriodEnd    *time.Time
	Limit        int
	Offset       int
}
// GetDownloadURL формирует URL для скачивания отчёта
func (r *Report) GetDownloadURL(baseURL string) string {
	return baseURL + "/api/v1/reports/" + string(r.ID) + "/download"
}

// IsProcessed проверяет, обработан ли отчёт
func (r *Report) IsProcessed() bool {
	return r.ProcessingStatus == "processed"
}

// HasError проверяет, есть ли ошибка при обработке
func (r *Report) HasError() bool {
	return r.ProcessingStatus == "error"
}
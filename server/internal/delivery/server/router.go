package server

import (
	"financial-risk-server/internal/delivery/server/handlers"
	"financial-risk-server/internal/service"
	"net/http"
)

// SetupRouter настраивает маршруты приложения
func SetupRouter(
	enterpriseService *service.EnterpriseService,
	reportService *service.ReportService,
	riskService *service.RiskCalculationService,
	comprehensiveReportService *service.ComprehensiveReportService,
) *http.ServeMux {
	mux := http.NewServeMux()

	// Создаём хэндлеры
	enterpriseHandler := handlers.NewEnterpriseHandler(enterpriseService)
	reportHandler := handlers.NewReportHandler(reportService)
	riskHandler := handlers.NewRiskHandler(riskService)
	comprehensiveReportHandler := handlers.NewComprehensiveReportHandler(comprehensiveReportService)
	healthHandler := handlers.NewHealthHandler()

	// Health check
	mux.HandleFunc("GET /health", healthHandler.Check)

	// API версии 1
	// Предприятия
	mux.HandleFunc("GET /api/v1/enterprises/{id}", enterpriseHandler.GetEnterpriseByID)

	// Отчёты
	mux.HandleFunc("GET /api/v1/reports", reportHandler.GetReports)
	mux.HandleFunc("GET /api/v1/reports/{id}", reportHandler.GetReport)
	mux.HandleFunc("POST /api/v1/reports/upload", reportHandler.UploadReport)
	mux.HandleFunc("DELETE /api/v1/reports/{id}", reportHandler.DeleteReport)
	mux.HandleFunc("GET /api/v1/reports/{id}/download", reportHandler.DownloadReport)

	// Расчёты рисков
	mux.HandleFunc("GET /api/v1/risks", riskHandler.GetRiskCalculations)
	mux.HandleFunc("GET /api/v1/risks/{id}", riskHandler.GetRiskCalculationByID)
	mux.HandleFunc("POST /api/v1/risks/calculate", riskHandler.CalculateRisks)
	mux.HandleFunc("GET /api/v1/risks/all", riskHandler.CalculateAllRisks)

	// Комплексные отчёты
	mux.HandleFunc("GET /api/v1/comprehensive-reports", comprehensiveReportHandler.GetReports)
	mux.HandleFunc("GET /api/v1/comprehensive-reports/{id}", comprehensiveReportHandler.GetReportByID)
	mux.HandleFunc("POST /api/v1/comprehensive-reports/generate", comprehensiveReportHandler.GenerateReport)

	// Обработчик 404
	mux.HandleFunc("GET /{path...}", func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})

	return mux
}
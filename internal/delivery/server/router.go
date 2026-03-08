package server

import (
	"financial-risk-server/internal/delivery/server/handlers"
	"financial-risk-server/internal/delivery/server/middleware"
	"net/http"
)

// SetupRouter настраивает все маршруты приложения
func SetupRouter(
	// Обработчики предприятий
	enterpriseHandler *handlers.EnterpriseHandler,

	// Обработчики данных
	contractHandler *handlers.ExportContractHandler,
	costHandler *handlers.CostStructureHandler,
	balanceHandler *handlers.BalanceSheetHandler,
	marketHandler *handlers.MarketDataHandler,

	// Обработчики рисков
	riskHandler *handlers.RiskHandler,

	// Health check
	healthHandler *handlers.HealthHandler,
) http.Handler {
	mux := http.NewServeMux()

	// ========== Health Check ==========
	mux.HandleFunc("GET /health", healthHandler.Check)

	// ========== Предприятия ==========
	mux.HandleFunc("POST /api/v1/enterprises", enterpriseHandler.CreateEnterprise)
	mux.HandleFunc("GET /api/v1/enterprises", enterpriseHandler.ListEnterprises)
	mux.HandleFunc("GET /api/v1/enterprises/", enterpriseHandler.GetEnterprise)

	// ========== Экспортные контракты ==========
	mux.HandleFunc("POST /api/v1/export-contracts", contractHandler.CreateContract)
	mux.HandleFunc("GET /api/v1/export-contracts", contractHandler.ListContracts)
	mux.HandleFunc("GET /api/v1/export-contracts/", contractHandler.GetContract)
	mux.HandleFunc("PUT /api/v1/export-contracts/", contractHandler.UpdateContract)
	mux.HandleFunc("DELETE /api/v1/export-contracts/", contractHandler.DeleteContract)

	// ========== Структура затрат ==========
	mux.HandleFunc("POST /api/v1/cost-structures", costHandler.CreateCost)
	mux.HandleFunc("GET /api/v1/cost-structures", costHandler.ListCosts)
	mux.HandleFunc("GET /api/v1/cost-structures/", costHandler.GetCost)
	mux.HandleFunc("PUT /api/v1/cost-structures/", costHandler.UpdateCost)
	mux.HandleFunc("DELETE /api/v1/cost-structures/", costHandler.DeleteCost)

	// ========== Рыночные данные ==========
	mux.HandleFunc("POST /api/v1/market-data", marketHandler.CreateMarketData)
	mux.HandleFunc("PUT /api/v1/cost-market-data/", marketHandler.UpdateMarketData)
	mux.HandleFunc("DELETE /api/v1/cost-structures/", marketHandler.DeleteMarketData)

	// ========== Финансовые балансы ==========
	mux.HandleFunc("POST /api/v1/balance-sheets", balanceHandler.CreateBalance)
	mux.HandleFunc("GET /api/v1/balance-sheets", balanceHandler.ListBalances)
	mux.HandleFunc("GET /api/v1/balance-sheets/", balanceHandler.GetBalance)
	mux.HandleFunc("GET /api/v1/balance-sheets/latest", balanceHandler.GetLatestBalance)
	mux.HandleFunc("PUT /api/v1/balance-sheets/", balanceHandler.UpdateBalance)
	mux.HandleFunc("DELETE /api/v1/balance-sheets/", balanceHandler.DeleteBalance)

	// ========== Расчёт рисков ==========
	mux.HandleFunc("POST /api/v1/risks/calculate", riskHandler.CalculateRisk)
	mux.HandleFunc("GET /api/v1/risks/all", riskHandler.CalculateAllRisks)

	// Применяем мидлвары
	handler := middleware.RecoveryMiddleware(
		middleware.LoggerMiddleware(
			middleware.CORSMiddleware(mux),
		),
	)

	return handler
}

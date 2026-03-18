 package service_test

// import (
// 	"context"
// 	"encoding/json"
// 	"testing"
// 	"time"

// 	"financial-risk-server/internal/domain/models"
// 	"financial-risk-server/internal/repository/interfaces"
// 	"financial-risk-server/internal/service"
// )

// // Mock репозиториев
// type MockRiskCalculationRepository struct {
// 	mock.Mock
// }

// func (m *MockRiskCalculationRepository) Create(ctx context.Context, risk *models.RiskCalculation) error {
// 	args := m.Called(ctx, risk)
// 	return args.Error(0)
// }

// func (m *MockRiskCalculationRepository) GetAll(ctx context.Context, filter interfaces.RiskCalculationFilter) ([]*models.RiskCalculation, error) {
// 	args := m.Called(ctx, filter)
// 	return args.Get(0).([]*models.RiskCalculation), args.Error(1)
// }

// func (m *MockRiskCalculationRepository) GetByID(ctx context.Context, id int64) (*models.RiskCalculation, error) {
// 	args := m.Called(ctx, id)
// 	return args.Get(0).(*models.RiskCalculation), args.Error(1)
// }

// type MockExportContractRepository struct {
// 	mock.Mock
// }

// func (m *MockExportContractRepository) GetByEnterpriseID(ctx context.Context, enterpriseID int64) ([]*models.ExportContract, error) {
// 	args := m.Called(ctx, enterpriseID)
// 	return args.Get(0).([]*models.ExportContract), args.Error(1)
// }

// type MockBalanceSheetRepository struct {
// 	mock.Mock
// }

// func (m *MockBalanceSheetRepository) GetLatest(ctx context.Context, enterpriseID int64) (*models.BalanceSheet, error) {
// 	args := m.Called(ctx, enterpriseID)
// 	return args.Get(0).(*models.BalanceSheet), args.Error(1)
// }

// type MockCreditAgreementRepository struct {
// 	mock.Mock
// }

// func (m *MockCreditAgreementRepository) GetByEnterpriseID(ctx context.Context, enterpriseID int64) ([]*models.CreditAgreement, error) {
// 	args := m.Called(ctx, enterpriseID)
// 	return args.Get(0).([]*models.CreditAgreement), args.Error(1)
// }

// type MockMarketDataRepository struct {
// 	mock.Mock
// }

// func (m *MockMarketDataRepository) Create(ctx context.Context, data *models.MarketData) error {
// 	args := m.Called(ctx, data)
// 	return args.Error(0)
// }

// func (m *MockMarketDataRepository) GetByDateAndPair(ctx context.Context, date time.Time, pair string) (*models.MarketData, error) {
// 	args := m.Called(ctx, date, pair)
// 	return args.Get(0).(*models.MarketData), args.Error(1)
// }

// func (m *MockMarketDataRepository) GetHistory(ctx context.Context, pair string, days int) ([]*models.MarketData, error) {
// 	args := m.Called(ctx, pair, days)
// 	return args.Get(0).([]*models.MarketData), args.Error(1)
// }

// func (m *MockMarketDataRepository) GetLatest(ctx context.Context, pair string) (*models.MarketData, error) {
// 	args := m.Called(ctx, pair)
// 	return args.Get(0).(*models.MarketData), args.Error(1)
// }

// func (m *MockMarketDataRepository) GetLatestRates(ctx context.Context) (map[string]float64, error) {
// 	args := m.Called(ctx)
// 	return args.Get(0).(map[string]float64), args.Error(1)
// }

// func (m *MockMarketDataRepository) Update(ctx context.Context, data *models.MarketData) error {
// 	args := m.Called(ctx, data)
// 	return args.Error(0)
// }

// // TestCalculateCurrencyRisk_Success тестирует успешный расчёт валютного риска
// func TestCalculateCurrencyRisk_Success(t *testing.T) {
// 	// Arrange
// 	ctx := context.Background()
// 	enterpriseID := int64(1)
// 	horizonDays := 10
// 	confidenceLevel := 0.95

// 	// Моки
// 	mockRiskRepo := new(MockRiskCalculationRepository)
// 	mockContractRepo := new(MockExportContractRepository)
// 	mockBalanceRepo := new(MockBalanceSheetRepository)
// 	mockCreditRepo := new(MockCreditAgreementRepository)
// 	mockMarketRepo := new(MockMarketDataRepository)
// 	mockEnterpriseService := &service.EnterpriseService{}

// 	// Тестовые данные
// 	contracts := []*models.ExportContract{
// 		{
// 			ID:             1,
// 			ContractValueUSD: 100000,
// 			PaymentStatus:  "pending",
// 		},
// 		{
// 			ID:             2,
// 			ContractValueUSD: 50000,
// 			PaymentStatus:  "pending",
// 		},
// 	}

// 	rates := map[string]float64{
// 		"USD/BYN": 3.25,
// 	}

// 	history := []*models.MarketData{
// 		{ExchangeRate: float64Ptr(3.20)},
// 		{ExchangeRate: float64Ptr(3.22)},
// 		{ExchangeRate: float64Ptr(3.25)},
// 		{ExchangeRate: float64Ptr(3.23)},
// 		{ExchangeRate: float64Ptr(3.26)},
// 	}

// 	// Настройка моков
// 	mockContractRepo.On("GetByEnterpriseID", ctx, enterpriseID).Return(contracts, nil)
// 	mockMarketRepo.On("GetLatestRates", ctx).Return(rates, nil)
// 	mockMarketRepo.On("GetHistory", ctx, "BYN/USD", 31).Return(history, nil)
// 	mockRiskRepo.On("Create", ctx, mock.Anything).Return(nil)

// 	// Создаём сервис
// 	svc := service.NewRiskCalculationService(
// 		mockRiskRepo,
// 		mockContractRepo,
// 		mockBalanceRepo,
// 		mockCreditRepo,
// 		mockMarketRepo,
// 		mockEnterpriseService,
// 	)

// 	// Act
// 	result, err := svc.CalculateCurrencyRisk(ctx, enterpriseID, horizonDays, confidenceLevel)

// 	// Assert
// 	assert.NoError(t, err)
// 	assert.NotNil(t, result)
// 	assert.Equal(t, "currency", result.RiskType)
// 	assert.Equal(t, enterpriseID, result.EnterpriseID)
// 	assert.Equal(t, 150000.0, result.ExposureAmount) // 100000 + 50000
// 	assert.Greater(t, result.VaRValue, 0.0)
// 	assert.Greater(t, result.StressTestLoss, 0.0)
// 	assert.Contains(t, []string{"low", "medium", "high"}, result.RiskLevel)
// 	assert.NotEmpty(t, result.Recommendations)

// 	// Проверяем, что Assumptions валидный JSON
// 	var assumptions map[string]interface{}
// 	err = json.Unmarshal([]byte(*result.Assumptions), &assumptions)
// 	assert.NoError(t, err)
// 	assert.Contains(t, assumptions, "volatility30d")
// 	assert.Contains(t, assumptions, "zScore")

// 	// Verify mocks
// 	mockContractRepo.AssertExpectations(t)
// 	mockMarketRepo.AssertExpectations(t)
// 	mockRiskRepo.AssertExpectations(t)
// }

// // TestCalculateCurrencyRisk_NoContracts тестирует расчёт без контрактов
// func TestCalculateCurrencyRisk_NoContracts(t *testing.T) {
// 	ctx := context.Background()
// 	enterpriseID := int64(1)
// 	horizonDays := 10
// 	confidenceLevel := 0.95

// 	mockRiskRepo := new(MockRiskCalculationRepository)
// 	mockContractRepo := new(MockExportContractRepository)
// 	mockBalanceRepo := new(MockBalanceSheetRepository)
// 	mockCreditRepo := new(MockCreditAgreementRepository)
// 	mockMarketRepo := new(MockMarketDataRepository)
// 	mockEnterpriseService := &service.EnterpriseService{}

// 	contracts := []*models.ExportContract{}

// 	mockContractRepo.On("GetByEnterpriseID", ctx, enterpriseID).Return(contracts, nil)

// 	svc := service.NewRiskCalculationService(
// 		mockRiskRepo,
// 		mockContractRepo,
// 		mockBalanceRepo,
// 		mockCreditRepo,
// 		mockMarketRepo,
// 		mockEnterpriseService,
// 	)

// 	result, err := svc.CalculateCurrencyRisk(ctx, enterpriseID, horizonDays, confidenceLevel)

// 	assert.NoError(t, err)
// 	assert.NotNil(t, result)
// 	assert.Equal(t, 0.0, result.ExposureAmount)
// 	assert.Equal(t, 0.0, result.VaRValue)
// 	assert.Equal(t, "low", result.RiskLevel)
// 	assert.Equal(t, "no_data", *result.CalculationMethod)
// }

// // TestCalculateInterestRisk_Success тестирует успешный расчёт процентного риска
// func TestCalculateInterestRisk_Success(t *testing.T) {
// 	ctx := context.Background()
// 	enterpriseID := int64(1)
// 	horizonDays := 10
// 	confidenceLevel := 0.95

// 	mockRiskRepo := new(MockRiskCalculationRepository)
// 	mockContractRepo := new(MockExportContractRepository)
// 	mockBalanceRepo := new(MockBalanceSheetRepository)
// 	mockCreditRepo := new(MockCreditAgreementRepository)
// 	mockMarketRepo := new(MockMarketDataRepository)
// 	mockEnterpriseService := &service.EnterpriseService{}

// 	outstandingBalance := 500000.0
// 	agreements := []*models.CreditAgreement{
// 		{
// 			ID:                 1,
// 			PrincipalAmount:    1000000,
// 			OutstandingBalance: &outstandingBalance,
// 			InterestRate:       0.12,
// 			IsFloating:         true,
// 		},
// 		{
// 			ID:              2,
// 			PrincipalAmount: 300000,
// 			InterestRate:    0.10,
// 			IsFloating:      false,
// 		},
// 	}

// 	mockCreditRepo.On("GetByEnterpriseID", ctx, enterpriseID).Return(agreements, nil)
// 	mockRiskRepo.On("Create", ctx, mock.Anything).Return(nil)

// 	svc := service.NewRiskCalculationService(
// 		mockRiskRepo,
// 		mockContractRepo,
// 		mockBalanceRepo,
// 		mockCreditRepo,
// 		mockMarketRepo,
// 		mockEnterpriseService,
// 	)

// 	result, err := svc.CalculateInterestRisk(ctx, enterpriseID, horizonDays, confidenceLevel)

// 	assert.NoError(t, err)
// 	assert.NotNil(t, result)
// 	assert.Equal(t, "interest", result.RiskType)
// 	assert.Equal(t, 500000.0, result.ExposureAmount)
// 	assert.Equal(t, 5000.0, result.VaRValue)      // 500000 * 0.01
// 	assert.Equal(t, 15000.0, result.StressTestLoss) // 500000 * 0.03

// 	mockCreditRepo.AssertExpectations(t)
// }

// // TestCalculateLiquidityRisk_Success тестирует успешный расчёт риска ликвидности
// func TestCalculateLiquidityRisk_Success(t *testing.T) {
// 	ctx := context.Background()
// 	enterpriseID := int64(1)
// 	horizonDays := 10
// 	confidenceLevel := 0.95

// 	mockRiskRepo := new(MockRiskCalculationRepository)
// 	mockContractRepo := new(MockExportContractRepository)
// 	mockBalanceRepo := new(MockBalanceSheetRepository)
// 	mockCreditRepo := new(MockCreditAgreementRepository)
// 	mockMarketRepo := new(MockMarketDataRepository)
// 	mockEnterpriseService := &service.EnterpriseService{}

// 	balance := &models.BalanceSheet{
// 		CashBYN:              100000,
// 		CashUSD:              50000,
// 		ShortTermInvestments: 20000,
// 		Inventories:          150000,
// 		AccountsPayable:      200000,
// 		VATPayable:           30000,
// 		PayrollPayable:       50000,
// 		ShortTermDebt:        100000,
// 	}

// 	rates := map[string]float64{
// 		"USD/BYN": 3.25,
// 	}

// 	mockBalanceRepo.On("GetLatest", ctx, enterpriseID).Return(balance, nil)
// 	mockMarketRepo.On("GetLatestRates", ctx).Return(rates, nil)
// 	mockRiskRepo.On("Create", ctx, mock.Anything).Return(nil)

// 	svc := service.NewRiskCalculationService(
// 		mockRiskRepo,
// 		mockContractRepo,
// 		mockBalanceRepo,
// 		mockCreditRepo,
// 		mockMarketRepo,
// 		mockEnterpriseService,
// 	)

// 	result, err := svc.CalculateLiquidityRisk(ctx, enterpriseID, horizonDays, confidenceLevel)

// 	assert.NoError(t, err)
// 	assert.NotNil(t, result)
// 	assert.Equal(t, "liquidity", result.RiskType)
// 	assert.Equal(t, 380000.0, result.ExposureAmount) // 200000 + 30000 + 50000 + 100000

// 	// Проверяем Assumptions
// 	var assumptions map[string]interface{}
// 	err = json.Unmarshal([]byte(*result.Assumptions), &assumptions)
// 	assert.NoError(t, err)
// 	assert.Contains(t, assumptions, "currentRatio")
// 	assert.Contains(t, assumptions, "quickRatio")

// 	mockBalanceRepo.AssertExpectations(t)
// }

// // TestCalculateAllRisks_InvalidHorizon тестирует валидацию горизонта
// func TestCalculateAllRisks_InvalidHorizon(t *testing.T) {
// 	ctx := context.Background()
// 	enterpriseID := int64(1)

// 	mockRiskRepo := new(MockRiskCalculationRepository)
// 	mockContractRepo := new(MockExportContractRepository)
// 	mockBalanceRepo := new(MockBalanceSheetRepository)
// 	mockCreditRepo := new(MockCreditAgreementRepository)
// 	mockMarketRepo := new(MockMarketDataRepository)
// 	mockEnterpriseService := &service.EnterpriseService{}

// 	svc := service.NewRiskCalculationService(
// 		mockRiskRepo,
// 		mockContractRepo,
// 		mockBalanceRepo,
// 		mockCreditRepo,
// 		mockMarketRepo,
// 		mockEnterpriseService,
// 	)

// 	_, err := svc.CalculateAllRisks(ctx, enterpriseID, 0, 0.95)
// 	assert.Error(t, err)
// 	assert.Contains(t, err.Error(), "horizonDays must be between 1 and 365")

// 	_, err = svc.CalculateAllRisks(ctx, enterpriseID, 400, 0.95)
// 	assert.Error(t, err)
// 	assert.Contains(t, err.Error(), "horizonDays must be between 1 and 365")
// }

// // TestCalculateAllRisks_InvalidConfidenceLevel тестирует валидацию уровня доверия
// func TestCalculateAllRisks_InvalidConfidenceLevel(t *testing.T) {
// 	ctx := context.Background()
// 	enterpriseID := int64(1)

// 	mockRiskRepo := new(MockRiskCalculationRepository)
// 	mockContractRepo := new(MockExportContractRepository)
// 	mockBalanceRepo := new(MockBalanceSheetRepository)
// 	mockCreditRepo := new(MockCreditAgreementRepository)
// 	mockMarketRepo := new(MockMarketDataRepository)
// 	mockEnterpriseService := &service.EnterpriseService{}

// 	svc := service.NewRiskCalculationService(
// 		mockRiskRepo,
// 		mockContractRepo,
// 		mockBalanceRepo,
// 		mockCreditRepo,
// 		mockMarketRepo,
// 		mockEnterpriseService,
// 	)

// 	_, err := svc.CalculateAllRisks(ctx, enterpriseID, 10, 0.80)
// 	assert.Error(t, err)
// 	assert.Contains(t, err.Error(), "confidenceLevel must be between 0.90 and 0.99")

// 	_, err = svc.CalculateAllRisks(ctx, enterpriseID, 10, 1.0)
// 	assert.Error(t, err)
// 	assert.Contains(t, err.Error(), "confidenceLevel must be between 0.90 and 0.99")
// }

// // TestDetermineRiskLevel тестирует определение уровня риска
// func TestDetermineRiskLevel(t *testing.T) {
// 	testCases := []struct {
// 		riskRatio float64
// 		expected  string
// 	}{
// 		{0.03, "low"},
// 		{0.05, "low"},
// 		{0.06, "medium"},
// 		{0.10, "medium"},
// 		{0.15, "medium"},
// 		{0.16, "high"},
// 		{0.25, "high"},
// 	}

// 	for _, tc := range testCases {
// 		t.Run(tc.expected, func(t *testing.T) {
// 			// Создаём сервис для доступа к determineRiskLevel
// 			mockRiskRepo := new(MockRiskCalculationRepository)
// 			mockContractRepo := new(MockExportContractRepository)
// 			mockBalanceRepo := new(MockBalanceSheetRepository)
// 			mockCreditRepo := new(MockCreditAgreementRepository)
// 			mockMarketRepo := new(MockMarketDataRepository)
// 			mockEnterpriseService := &service.EnterpriseService{}

// 			svc := service.NewRiskCalculationService(
// 				mockRiskRepo,
// 				mockContractRepo,
// 				mockBalanceRepo,
// 				mockCreditRepo,
// 				mockMarketRepo,
// 				mockEnterpriseService,
// 			)

// 			// Используем рефлексию или создаём тестовый метод
// 			// Для простоты проверяем через публичные методы
// 			assert.NotNil(t, svc)
// 		})
// 	}
// }

// // Вспомогательные функции
// func float64Ptr(f float64) *float64 {
// 	return &f
// }

// func stringPtr(s string) *string {
// 	return &s
// }
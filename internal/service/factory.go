package service

import "finantial-risk-server/internal/repository/interfaces"

// ServiceFactory фабрика для создания сервисов
type ServiceFactory struct {
	config *RiskCalculationServiceConfig
}

// NewServiceFactory создаёт новую фабрику сервисов
func NewServiceFactory(config *RiskCalculationServiceConfig) *ServiceFactory {
	if config == nil {
		// Используем значения по умолчанию
		config = &RiskCalculationServiceConfig{
			DefaultHorizonDays:     30,
			DefaultConfidenceLevel: 0.95,
			DefaultPriceChangePct:  -15.0,
			DefaultRateChangePct:   1.0,
		}
	}

	return &ServiceFactory{
		config: config,
	}
}

// CreateRiskCalculationService создаёт полный сервис расчёта рисков
func (f *ServiceFactory) CreateRiskCalculationService(
	exportRepo interfaces.ExportContractRepository,
	marketRepo interfaces.MarketDataRepository,
	balanceRepo interfaces.BalanceSheetRepository,
	costRepo interfaces.CostStructureRepository,
	enterpriseRepo interfaces.EnterpriseRepository,
	riskCalcRepo interfaces.RiskCalculationRepository,
) RiskCalculationService {

	// Создаём отдельные сервисы для каждого типа риска
	currencyRiskService := NewCurrencyRiskService(exportRepo, marketRepo, f.config)
	creditRiskService := NewCreditRiskService(exportRepo, f.config)
	liquidityRiskService := NewLiquidityRiskService(balanceRepo, costRepo, f.config)
	marketRiskService := NewMarketRiskService(enterpriseRepo, marketRepo, exportRepo, f.config)
	interestRiskService := NewInterestRiskService(balanceRepo, f.config)

	// Создаём агрегирующий сервис
	aggregator := NewRiskAggregatorService(
		currencyRiskService,
		creditRiskService,
		liquidityRiskService,
		marketRiskService,
		interestRiskService,
		riskCalcRepo,
		f.config,
	)

	return aggregator
}

// CreateCurrencyRiskService создаёт только сервис валютного риска
func (f *ServiceFactory) CreateCurrencyRiskService(
	exportRepo interfaces.ExportContractRepository,
	marketRepo interfaces.MarketDataRepository,
) *currencyRiskService {
	return NewCurrencyRiskService(exportRepo, marketRepo, f.config)
}

// CreateCreditRiskService создаёт только сервис кредитного риска
func (f *ServiceFactory) CreateCreditRiskService(
	exportRepo interfaces.ExportContractRepository,
) *creditRiskService {
	return NewCreditRiskService(exportRepo, f.config)
}

// CreateLiquidityRiskService создаёт только сервис риска ликвидности
func (f *ServiceFactory) CreateLiquidityRiskService(
	balanceRepo interfaces.BalanceSheetRepository,
	costRepo interfaces.CostStructureRepository,
) *liquidityRiskService {
	return NewLiquidityRiskService(balanceRepo, costRepo, f.config)
}

// CreateMarketRiskService создаёт только сервис фондового риска
func (f *ServiceFactory) CreateMarketRiskService(
	enterpriseRepo interfaces.EnterpriseRepository,
	marketRepo interfaces.MarketDataRepository,
	exportRepo interfaces.ExportContractRepository,
) *marketRiskService {
	return NewMarketRiskService(enterpriseRepo, marketRepo, exportRepo, f.config)
}

// CreateInterestRiskService создаёт только сервис процентного риска
func (f *ServiceFactory) CreateInterestRiskService(
	balanceRepo interfaces.BalanceSheetRepository,
) *interestRiskService {
	return NewInterestRiskService(balanceRepo, f.config)
}

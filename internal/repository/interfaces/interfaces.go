package interfaces

import (
	"finantial-risk-server/internal/domain/models"
	"time"
)

// EnterpriseRepository определяет методы для работы с предприятиями
type EnterpriseRepository interface {
	Create(enterprise *models.Enterprise) error
	GetByID(id int64) (*models.Enterprise, error)
	GetByName(name string) (*models.Enterprise, error)
	GetAll() ([]*models.Enterprise, error)
	Update(enterprise *models.Enterprise) error
	Delete(id int64) error
}

// ExportContractRepository определяет методы для работы с экспортными контрактами
type ExportContractRepository interface {
	Create(contract *models.ExportContract) error
	GetByID(id int64) (*models.ExportContract, error)
	GetByEnterpriseID(enterpriseID int64) ([]*models.ExportContract, error)
	GetPendingContracts(enterpriseID int64) ([]*models.ExportContract, error)
	GetByPeriod(enterpriseID int64, start, end time.Time) ([]*models.ExportContract, error)
	Update(contract *models.ExportContract) error
	UpdateStatus(id int64, status string) error
	Delete(id int64) error
}

// CostStructureRepository определяет методы для работы со структурой затрат
type CostStructureRepository interface {
	Create(cost *models.CostStructure) error
	GetByID(id int64) (*models.CostStructure, error)
	GetByEnterpriseID(enterpriseID int64) ([]*models.CostStructure, error)
	GetByPeriod(enterpriseID int64, start, end time.Time) ([]*models.CostStructure, error)
	Update(cost *models.CostStructure) error
	Delete(id int64) error
}

// BalanceSheetRepository определяет методы для работы с балансами
type BalanceSheetRepository interface {
	Create(balance *models.BalanceSheet) error
	GetByID(id int64) (*models.BalanceSheet, error)
	GetByEnterpriseID(enterpriseID int64) ([]*models.BalanceSheet, error)
	GetLatest(enterpriseID int64) (*models.BalanceSheet, error)
	Update(balance *models.BalanceSheet) error
	Delete(id int64) error
}

// MarketDataRepository определяет методы для работы с рыночными данными
type MarketDataRepository interface {
	Create(data *models.MarketData) error
	GetByID(id int64) (*models.MarketData, error)
	GetLatestByCurrencyPair(pair string) (*models.MarketData, error)
	GetByDateRange(pair string, start, end time.Time) ([]*models.MarketData, error)
	GetVolatility(pair string, days int) (float64, error)
	Update(data *models.MarketData) error
	Delete(id int64) error
}

// RiskCalculationRepository определяет методы для работы с результатами расчётов
type RiskCalculationRepository interface {
	Create(result *models.RiskResult) error
	GetByID(id int64) (*models.RiskResult, error)
	GetByEnterpriseID(enterpriseID int64) ([]*models.RiskResult, error)
	GetByRiskType(enterpriseID int64, riskType models.RiskType) ([]*models.RiskResult, error)
	GetByDateRange(enterpriseID int64, start, end time.Time) ([]*models.RiskResult, error)
	GetLatest(enterpriseID int64, riskType models.RiskType) (*models.RiskResult, error)
	Update(result *models.RiskResult) error
	Delete(id int64) error
}

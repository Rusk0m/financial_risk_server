package service

import (
	"context"
	"fmt"
	"math"
	"time"

	"financial-risk-server/internal/domain/models"
	"financial-risk-server/internal/repository/interfaces"
)

// RiskCalculationService предоставляет бизнес-логику для расчёта финансовых рисков
type RiskCalculationService struct {
	riskRepo          interfaces.RiskCalculationRepository
	contractRepo      interfaces.ExportContractRepository
	balanceRepo       interfaces.BalanceSheetRepository
	creditRepo        interfaces.CreditAgreementRepository
	marketDataRepo    interfaces.MarketDataRepository
	enterpriseService *EnterpriseService
}

// NewRiskCalculationService создаёт новый сервис расчёта рисков
func NewRiskCalculationService(
	riskRepo interfaces.RiskCalculationRepository,
	contractRepo interfaces.ExportContractRepository,
	balanceRepo interfaces.BalanceSheetRepository,
	creditRepo interfaces.CreditAgreementRepository,
	marketDataRepo interfaces.MarketDataRepository,
	enterpriseService *EnterpriseService,
) *RiskCalculationService {
	return &RiskCalculationService{
		riskRepo:          riskRepo,
		contractRepo:      contractRepo,
		balanceRepo:       balanceRepo,
		creditRepo:        creditRepo,
		marketDataRepo:    marketDataRepo,
		enterpriseService: enterpriseService,
	}
}

// CalculateAllRisks рассчитывает все три типа рисков для предприятия
func (s *RiskCalculationService) CalculateAllRisks(
	ctx context.Context,
	enterpriseID int64,
	horizonDays int,
	confidenceLevel float64,
) (map[string]*models.RiskCalculation, error) {
	results := make(map[string]*models.RiskCalculation)

	// Рассчитываем валютный риск
	currencyRisk, err := s.CalculateCurrencyRisk(ctx, enterpriseID, horizonDays, confidenceLevel)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate currency risk: %w", err)
	}
	results["currency"] = currencyRisk

	// Рассчитываем процентный риск
	interestRisk, err := s.CalculateInterestRisk(ctx, enterpriseID, horizonDays, confidenceLevel)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate interest risk: %w", err)
	}
	results["interest"] = interestRisk

	// Рассчитываем риск ликвидности
	liquidityRisk, err := s.CalculateLiquidityRisk(ctx, enterpriseID, horizonDays, confidenceLevel)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate liquidity risk: %w", err)
	}
	results["liquidity"] = liquidityRisk

	return results, nil
}

// CalculateCurrencyRisk рассчитывает валютный риск
func (s *RiskCalculationService) CalculateCurrencyRisk(
	ctx context.Context,
	enterpriseID int64,
	horizonDays int,
	confidenceLevel float64,
) (*models.RiskCalculation, error) {
	// 1. Получаем все незакрытые экспортные контракты
	contracts, err := s.contractRepo.GetByEnterpriseID(ctx, enterpriseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get export contracts: %w", err)
	}

	// 2. Рассчитываем экспозицию (сумма неполученной выручки в USD)
	exposureAmount := 0.0
	for _, contract := range contracts {
		if contract.PaymentStatus != "paid" {
			exposureAmount += contract.GetContractValueUSD()
		}
	}

	// 3. Получаем текущую волатильность курса BYN/USD
	marketData, err := s.marketDataRepo.GetLatestByCurrencyPair(ctx, "BYN/USD")
	if err != nil {
		// Используем волатильность по умолчанию 9%
		marketData = &models.MarketData{Volatility30d: float64Ptr(0.09)}
	}

	volatility := 0.09
	if marketData.Volatility30d != nil {
		volatility = *marketData.Volatility30d
	}

	// 4. Рассчитываем VaR (Value at Risk) по методу дисперсии-ковариации
	// VaR = Экспозиция × Волатильность × Z-оценка × √(Горизонт/252)
	// Z-оценка для 95% = 1.65, для 99% = 2.33
	zScore := getZScore(confidenceLevel)
	
	varValue := exposureAmount * volatility * zScore * math.Sqrt(float64(horizonDays)/252.0)

	// 5. Рассчитываем стресс-тест (сценарий ослабления курса на 25%)
	stressTestLoss := exposureAmount * 0.25

	// 6. Определяем уровень риска
	riskLevel := s.determineRiskLevel(varValue / exposureAmount)

	// 7. Формируем рекомендации
	recommendations := []string{
		"Хеджировать 30-40% валютной экспозиции через форвардные контракты",
		"Сократить срок оплаты экспортных контрактов до 45-60 дней",
		"Диверсифицировать валюту расчётов (включить до 20% в CNY, EUR)",
	}

	// 8. Создаём расчёт риска
	risk := &models.RiskCalculation{
		EnterpriseID:    enterpriseID,
		RiskType:        "currency",
		CalculationDate: time.Now(),
		HorizonDays:     horizonDays,
		ConfidenceLevel: confidenceLevel,
		ExposureAmount:  exposureAmount,
		VaRValue:        varValue,
		StressTestLoss:  stressTestLoss,
		RiskLevel:       riskLevel,
		CalculationMethod: stringPtr("variance_covariance"),
		ScenarioType:    stringPtr("base"),
		Recommendations: recommendations,
	}

	// 9. Сохраняем в БД
	if err := s.riskRepo.Create(ctx, risk); err != nil {
		return nil, fmt.Errorf("failed to save currency risk calculation: %w", err)
	}

	return risk, nil
}

// CalculateInterestRisk рассчитывает процентный риск
func (s *RiskCalculationService) CalculateInterestRisk(
	ctx context.Context,
	enterpriseID int64,
	horizonDays int,
	confidenceLevel float64,
) (*models.RiskCalculation, error) {
	// 1. Получаем все кредитные договоры предприятия
	agreements, err := s.creditRepo.GetByEnterpriseID(ctx, enterpriseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get credit agreements: %w", err)
	}

	// 2. Рассчитываем экспозицию (сумма кредитов с плавающей ставкой)
	exposureAmount := 0.0
	floatingRateCount := 0
	for _, agreement := range agreements {
		if agreement.IsFloatingRate() {
			amount := agreement.PrincipalAmount
			if agreement.OutstandingBalance != nil {
				amount = *agreement.OutstandingBalance
			}
			exposureAmount += amount
			floatingRateCount++
		}
	}

	// 3. Рассчитываем дополнительные процентные расходы при повышении ставки на 1%
	// Доп. расходы = Экспозиция × 0.01 (1%)
	varValue := exposureAmount * 0.01

	// 4. Стресс-тест (повышение ставки на 3%)
	stressTestLoss := exposureAmount * 0.03

	// 5. Определяем уровень риска
	riskLevel := s.determineRiskLevel(varValue / exposureAmount)

	// 6. Формируем рекомендации
	recommendations := []string{
		"Поддерживать текущую структуру капитала с фиксированной ставкой",
		"Мониторинг процентных ставок 1 раз в месяц",
		"Рассмотреть рефинансирование части долга при снижении ставок",
	}

	// 7. Создаём расчёт риска
	risk := &models.RiskCalculation{
		EnterpriseID:    enterpriseID,
		RiskType:        "interest",
		CalculationDate: time.Now(),
		HorizonDays:     horizonDays,
		ConfidenceLevel: confidenceLevel,
		ExposureAmount:  exposureAmount,
		VaRValue:        varValue,
		StressTestLoss:  stressTestLoss,
		RiskLevel:       riskLevel,
		CalculationMethod: stringPtr("scenario_analysis"),
		ScenarioType:    stringPtr("base"),
		Recommendations: recommendations,
	}

	// 8. Сохраняем в БД
	if err := s.riskRepo.Create(ctx, risk); err != nil {
		return nil, fmt.Errorf("failed to save interest risk calculation: %w", err)
	}

	return risk, nil
}

// CalculateLiquidityRisk рассчитывает риск ликвидности
func (s *RiskCalculationService) CalculateLiquidityRisk(
	ctx context.Context,
	enterpriseID int64,
	horizonDays int,
	confidenceLevel float64,
) (*models.RiskCalculation, error) {
	// 1. Получаем последний финансовый баланс
	balance, err := s.balanceRepo.GetLatest(ctx, enterpriseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest balance sheet: %w", err)
	}

	// 2. Рассчитываем ликвидные активы
	// Ликвидные активы = ДС в BYN + ДС в USD (в пересчёте по курсу) + КФВ
	liquidAssets := balance.CashBYN + balance.CashUSD*3.25 + balance.ShortTermInvestments

	// 3. Рассчитываем краткосрочные обязательства
	shortTermLiabilities := balance.AccountsPayable + balance.VATPayable + 
		balance.PayrollPayable + balance.ShortTermDebt

	// 4. Экспозиция = Краткосрочные обязательства
	exposureAmount := shortTermLiabilities

	// 5. Рассчитываем VaR на основе коэффициента текущей ликвидности
	// Текущая ликвидность = Ликвидные активы / Краткосрочные обязательства
	currentRatio := liquidAssets / shortTermLiabilities

	varValue := 0.0
	if currentRatio < 1.0 {
		// Критическая ситуация: нехватка ликвидности
		varValue = shortTermLiabilities * 0.30
	} else if currentRatio < 1.5 {
		// Средний риск
		varValue = shortTermLiabilities * 0.15
	} else {
		// Низкий риск
		varValue = shortTermLiabilities * 0.05
	}

	// 6. Стресс-тест (снижение ликвидных активов на 40%)
	stressTestLoss := shortTermLiabilities * 0.40

	// 7. Определяем уровень риска
	riskLevel := s.determineRiskLevel(varValue / exposureAmount)

	// 8. Формируем рекомендации
	recommendations := []string{
		"Усилить контроль за дебиторской задолженностью (еженедельный анализ)",
		"Переговорить с ключевыми поставщиками о более выгодных условиях оплаты",
		"Создать резервный фонд на 2-3 месяца операционных расходов",
	}

	// 9. Создаём расчёт риска
	risk := &models.RiskCalculation{
		EnterpriseID:    enterpriseID,
		RiskType:        "liquidity",
		CalculationDate: time.Now(),
		HorizonDays:     horizonDays,
		ConfidenceLevel: confidenceLevel,
		ExposureAmount:  exposureAmount,
		VaRValue:        varValue,
		StressTestLoss:  stressTestLoss,
		RiskLevel:       riskLevel,
		CalculationMethod: stringPtr("ratio_analysis"),
		ScenarioType:    stringPtr("base"),
		Recommendations: recommendations,
	}

	// 10. Сохраняем в БД
	if err := s.riskRepo.Create(ctx, risk); err != nil {
		return nil, fmt.Errorf("failed to save liquidity risk calculation: %w", err)
	}

	return risk, nil
}

// determineRiskLevel определяет уровень риска на основе отношения VaR к экспозиции
func (s *RiskCalculationService) determineRiskLevel(riskRatio float64) string {
	if riskRatio > 0.15 {
		return "high"
	}
	if riskRatio > 0.05 {
		return "medium"
	}
	return "low"
}

// getZScore возвращает Z-оценку для заданного уровня доверия
func getZScore(confidenceLevel float64) float64 {
	switch {
	case confidenceLevel >= 0.99:
		return 2.33 // 99% уровень доверия
	case confidenceLevel >= 0.95:
		return 1.65 // 95% уровень доверия
	default:
		return 1.28 // 90% уровень доверия
	}
}

// stringPtr вспомогательная функция для создания указателя на строку
func stringPtr(s string) *string {
	return &s
}

// GetAllRiskCalculations получает список расчётов с фильтрацией
func (s *RiskCalculationService) GetAllRiskCalculations(
	ctx context.Context,
	filter interfaces.RiskCalculationFilter,
) ([]*models.RiskCalculation, error) {
	return s.riskRepo.GetAll(ctx, filter)
}

// GetRiskCalculationByID получает расчёт по ID
func (s *RiskCalculationService) GetRiskCalculationByID(
	ctx context.Context,
	id int64,
) (*models.RiskCalculation, error) {
	return s.riskRepo.GetByID(ctx, id)
}


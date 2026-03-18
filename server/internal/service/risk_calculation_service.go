package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	"financial-risk-server/internal/domain/models"
	"financial-risk-server/internal/repository/interfaces"
)

// RiskCalculationService предоставляет бизнес-логику для расчёта финансовых рисков
type RiskCalculationService struct {
	riskRepo          interfaces.RiskCalculationRepository
	contractRepo      interfaces.ExportContractRepository
	balanceRepo       interfaces.BalanceSheetRepository
	finResultRepo	  interfaces.FinancialResultRepository
	creditRepo        interfaces.CreditAgreementRepository
	marketDataRepo    interfaces.MarketDataRepository
	volatilityCalc    *VolatilityCalculator
	enterpriseService *EnterpriseService
}

func NewRiskCalculationService(
	riskRepo interfaces.RiskCalculationRepository,
	contractRepo interfaces.ExportContractRepository,
	balanceRepo interfaces.BalanceSheetRepository,
	finResultsRepo interfaces.FinancialResultRepository,
	creditRepo interfaces.CreditAgreementRepository,
	marketDataRepo interfaces.MarketDataRepository,
	enterpriseService *EnterpriseService,
) *RiskCalculationService {
	return &RiskCalculationService{
		riskRepo:          riskRepo,
		contractRepo:      contractRepo,
		balanceRepo:       balanceRepo,
		finResultRepo: 	   finResultsRepo,
		creditRepo:        creditRepo,
		marketDataRepo:    marketDataRepo,
		volatilityCalc:    NewVolatilityCalculator(marketDataRepo),
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

	// Валидация входных параметров
	if horizonDays <= 0 || horizonDays > 365 {
		return nil, fmt.Errorf("horizonDays must be between 1 and 365, got: %d", horizonDays)
	}
	if confidenceLevel < 0.90 || confidenceLevel > 0.99 {
		return nil, fmt.Errorf("confidenceLevel must be between 0.90 and 0.99, got: %.2f", confidenceLevel)
	}

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

// service/risk_calculation.go

// CalculateCurrencyRisk рассчитывает валютный риск
func (s *RiskCalculationService) CalculateCurrencyRisk(
	ctx context.Context,
	enterpriseID int64,
	horizonDays int,
	confidenceLevel float64,
) (*models.RiskCalculation, error) {
	// 1. Получаем все незакрытые экспортные контракты
	contracts, err := s.contractRepo.GetUnpaidContract(ctx, enterpriseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get export contracts: %w", err)
	}

	// 🔍 Логирование для отладки
	log.Printf("🔍 [RiskCalc] Enterprise ID: %d", enterpriseID)
	log.Printf("🔍 [RiskCalc] Total unpaid contracts: %d", len(contracts))

	// 2. Рассчитываем экспозицию по каждой валюте
	exposureByCurrency := make(map[string]float64)
	totalExposureUSD := 0.0
	unpaidContracts := 0

	for _, contract := range contracts {
		if contract.PaymentStatus != "paid" {
			valueUSD := contract.GetContractValueUSD()
			currency := strings.ToUpper(contract.Currency)
			if currency == "" {
				currency = "USD"
			}

			exposureByCurrency[currency] += valueUSD
			totalExposureUSD += valueUSD
			unpaidContracts++

			log.Printf("🔍 [RiskCalc] Contract %s: Currency=%s, Value=%.2f USD",
				contract.ContractNumber, currency, valueUSD)
		}
	}

	log.Printf("🔍 [RiskCalc] Total exposure: %.2f USD", totalExposureUSD)
	log.Printf("🔍 [RiskCalc] Exposure by currency: %v", exposureByCurrency)

	// Если нет контрактов - риск нулевой
	if totalExposureUSD == 0 {
		return s.createEmptyRiskCalculation(enterpriseID, "currency", horizonDays, confidenceLevel), nil
	}

	// 3. Рассчитываем волатильность для каждой валютной пары
	volatilityByCurrency := make(map[string]float64)
	
	// 🔴 Обработка разных валютных пар
	currencyPairs := map[string]string{
		"USD": "BYN/USD",
		"EUR": "BYN/EUR",
		"CNY": "BYN/CNY",
	}

	for currency, pair := range currencyPairs {
		if exposureByCurrency[currency] > 0 {
			dailyVol, err := s.volatilityCalc.CalculateVolatility(ctx, pair, 30)
			if err != nil {
				dailyVol = 0.09 / math.Sqrt(252) // Fallback: 9% годовых → дневная
				log.Printf("⚠️  [RiskCalc] Using fallback volatility for %s: %.6f", currency, dailyVol)
			}
			volatilityByCurrency[currency] = dailyVol
			log.Printf("🔍 [RiskCalc] Volatility for %s (%s): %.6f (%.2f%% annual)",
				currency, pair, dailyVol, dailyVol*math.Sqrt(252)*100)
		}
	}

	// 4. Получаем текущие обменные курсы
	// rates, err := s.marketDataRepo.GetLatestRates(ctx)
	// if err != nil {
	// 	log.Printf("⚠️  [RiskCalc] Failed to get exchange rates: %v", err)
	// }

	// 5. Рассчитываем VaR для каждой валюты и суммируем
	totalVaR := 0.0
	zScore := getZScore(confidenceLevel)

	for currency, exposure := range exposureByCurrency {
		volatility := volatilityByCurrency[currency]
		if volatility == 0 {
			volatility = 0.09 / math.Sqrt(252) // Fallback
		}

		// VaR для этой валюты
		currencyVaR := exposure * volatility * zScore * math.Sqrt(float64(horizonDays))
		totalVaR += currencyVaR

		log.Printf("🔍 [RiskCalc] VaR for %s: Exposure=%.2f, Vol=%.6f, VaR=%.2f",
			currency, exposure, volatility, currencyVaR)
	}

	// 6. Стресс-тест (сценарий ослабления курса на 25%)
	stressTestLoss := totalExposureUSD * 0.25

	// 7. Определяем уровень риска
	riskRatio := totalVaR / totalExposureUSD
	riskLevel := s.determineRiskLevel(riskRatio)

	log.Printf("🔍 [RiskCalc] Risk Ratio: %.4f (%.2f%%), Level: %s", riskRatio, riskRatio*100, riskLevel)

	// 🔴 Проверка на аномалии
	if totalExposureUSD > 10_000_000_000 { // > 10 млрд USD
		log.Printf("⚠️  [RiskCalc] WARNING: Exposure seems too high! Check data.")
	}

	// 8. Формируем рекомендации
	recommendations := s.generateCurrencyRiskRecommendations(riskLevel, unpaidContracts)

	// 9. Создаём метаданные расчёта
	assumptions, _ := json.Marshal(map[string]interface{}{
		"volatilityByCurrency": volatilityByCurrency,
		"exposureByCurrency":   exposureByCurrency,
		"totalExposureUSD":     totalExposureUSD,
		"zScore":               zScore,
		"contractsCount":       unpaidContracts,
		"horizonDays":          horizonDays,
	})

	// 10. Создаём расчёт риска
	risk := &models.RiskCalculation{
		EnterpriseID:      enterpriseID,
		RiskType:          "currency",
		CalculationDate:   time.Now(),
		HorizonDays:       horizonDays,
		ConfidenceLevel:   confidenceLevel,
		ExposureAmount:    totalExposureUSD,
		VaRValue:          totalVaR,
		StressTestLoss:    stressTestLoss,
		RiskLevel:         riskLevel,
		CalculationMethod: stringPtr("variance_covariance"),
		ScenarioType:      stringPtr("base"),
		Assumptions:       stringPtr(string(assumptions)),
		Recommendations:   recommendations,
	}

	// 11. Сохраняем в БД
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

	// 2. Рассчитываем экспозицию (кредиты с плавающей ставкой)
	exposureAmount := 0.0
	floatingRateAmount := 0.0
	fixedRateAmount := 0.0
	floatingRateCount := 0
	weightedAvgMaturity := 0.0

	for _, agreement := range agreements {
		amount := agreement.PrincipalAmount
		if agreement.OutstandingBalance != nil {
			amount = *agreement.OutstandingBalance
		}

		if agreement.IsFloatingRate() {
			floatingRateAmount += amount
			floatingRateCount++
			// Учитываем срок до погашения для взвешенной волатильности
			weightedAvgMaturity += float64(agreement.TermMonths) * amount
		} else {
			fixedRateAmount += amount
		}
	}

	exposureAmount = floatingRateAmount

	// Если нет кредитов с плавающей ставкой - риск минимальный
	if exposureAmount == 0 {
		return s.createEmptyRiskCalculation(enterpriseID, "interest", horizonDays, confidenceLevel), nil
	}

	// Средневзвешенный срок
	if floatingRateAmount > 0 {
		weightedAvgMaturity = weightedAvgMaturity / floatingRateAmount
	}

	// 3. Получаем волатильность процентной ставки (историческая волатильность ключевой ставки)
	// Для Беларуси используем волатильность ставки рефинансирования НБ РБ
	rateVolatility, err := s.volatilityCalc.CalculateVolatility(ctx, "KEY_RATE", 90)
	if err != nil {
		// Fallback: 20% годовая волатильность ставки (типично для развивающихся рынков)
		rateVolatility = 0.20 / math.Sqrt(252)
		log.Printf("⚠️  [InterestRisk] Using fallback volatility: %.6f", rateVolatility)
	}

	// 4. Рассчитываем VaR с учётом confidence и horizon
	// Формула: VaR = Exposure × Volatility × Z-Score × √(Horizon/252) × Duration
	// Duration ≈ Term/12 для кредитов
	zScore := getZScore(confidenceLevel)
	duration := weightedAvgMaturity / 12.0 // Конвертируем месяцы в годы
	if duration <= 0 {
		duration = 1.0
	}

	// 🔴 Теперь VaR зависит от confidenceLevel и horizonDays!
	varValue := exposureAmount * rateVolatility * zScore * math.Sqrt(float64(horizonDays)/252.0) * duration

	// 5. Стресс-тест (повышение ставки на 3% на год)
	stressTestLoss := exposureAmount * 0.03 * duration

	// 6. Определяем уровень риска
	riskRatio := varValue / exposureAmount
	riskLevel := s.determineRiskLevel(riskRatio)

	// 7. Формируем рекомендации
	currentKeyRate := 0.10 // 10% по умолчанию
	recommendations := s.generateInterestRiskRecommendations(riskLevel, floatingRateCount, currentKeyRate)

	// 8. Создаём метаданные расчёта
	assumptions, _ := json.Marshal(map[string]interface{}{
		"floatingRateAmount":  floatingRateAmount,
		"fixedRateAmount":     fixedRateAmount,
		"currentKeyRate":      currentKeyRate,
		"agreementsCount":     floatingRateCount,
		"rateVolatility":      rateVolatility,
		"rateVolatilityAnn":   rateVolatility * math.Sqrt(252),
		"zScore":              zScore,
		"duration":            duration,
		"horizonDays":         horizonDays,
		"confidenceLevel":     confidenceLevel,
		"rateShock":           0.01,
		"stressTestShock":     0.03,
	})

	// 9. Создаём расчёт риска
	risk := &models.RiskCalculation{
		EnterpriseID:      enterpriseID,
		RiskType:          "interest",
		CalculationDate:   time.Now(),
		HorizonDays:       horizonDays,
		ConfidenceLevel:   confidenceLevel,
		ExposureAmount:    exposureAmount,
		VaRValue:          varValue,
		StressTestLoss:    stressTestLoss,
		RiskLevel:         riskLevel,
		CalculationMethod: stringPtr("duration_analysis"),
		ScenarioType:      stringPtr("base"),
		Assumptions:       stringPtr(string(assumptions)),
		Recommendations:   recommendations,
	}

	// 10. Сохраняем в БД
	if err := s.riskRepo.Create(ctx, risk); err != nil {
		return nil, fmt.Errorf("failed to save interest risk calculation: %w", err)
	}

	log.Printf("🔍 [InterestRisk] Exposure=%.2f, VaR=%.2f (%.2f%%), Duration=%.2f years",
		exposureAmount, varValue, riskRatio*100, duration)

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

	// 2. 🔴 Получаем последние финансовые результаты
	finResults, err := s.finResultRepo.GetLatest(ctx, enterpriseID)
	if err != nil {
		log.Printf("⚠️  [LiquidityRisk] Failed to get financial results: %v", err)
		// 🔴 НЕ продолжаем с nil, а устанавливаем в nil явно
		finResults = nil
	}

	// 3. Получаем текущий обменный курс
	rates, err := s.marketDataRepo.GetLatestRates(ctx)
	exchangeRate := 3.25
	if err == nil && rates != nil {
		if rate, ok := rates["USD/BYN"]; ok {
			exchangeRate = rate
		} else if rate, ok := rates["BYN/USD"]; ok {
			exchangeRate = 1.0 / rate
		}
	}

	// 4. 🔴 Рассчитываем ликвидные активы
	liquidAssets := balance.CashBYN + balance.CashUSD*exchangeRate + balance.ShortTermInvestments

	// 5. 🔴 Рассчитываем быстро реализуемые активы
	quickAssets := liquidAssets + balance.AccountsReceivable*0.8

	// 6. 🔴 Краткосрочные обязательства
	shortTermLiabilities := balance.AccountsPayable + balance.VATPayable +
		balance.PayrollPayable + balance.ShortTermDebt

	// Защита от деления на ноль
	if shortTermLiabilities == 0 {
		return s.createEmptyRiskCalculation(enterpriseID, "liquidity", horizonDays, confidenceLevel), nil
	}

	// 7. 🔴 Liquidity Gap
	liquidityGap := liquidAssets - shortTermLiabilities
	liquidityGapRatio := liquidityGap / shortTermLiabilities

	// 8. 🔴 Рассчитываем месячные операционные расходы
	monthlyOperatingExpenses := s.calculateMonthlyOperatingExpenses(finResults)
	log.Printf("🔍 [LiquidityRisk] Monthly Operating Expenses: %.2f BYN", monthlyOperatingExpenses)

	// 9. 🔴 Months of Coverage
	monthsOfCoverage := 0.0
	if monthlyOperatingExpenses > 0 {
		monthsOfCoverage = liquidAssets / monthlyOperatingExpenses
	}
	log.Printf("🔍 [LiquidityRisk] Months of Coverage: %.2f", monthsOfCoverage)

	// 10. 🔴 Волатильность денежных потоков
	cashFlowVolatility := s.calculateCashFlowVolatility(ctx, enterpriseID)
	if cashFlowVolatility <= 0 {
		cashFlowVolatility = 0.15 / math.Sqrt(12)
	}
	log.Printf("🔍 [LiquidityRisk] Cash Flow Volatility: %.6f (monthly)", cashFlowVolatility)

	// 11. 🔴 CFaR
	zScore := getZScore(confidenceLevel)
	horizonMonths := float64(horizonDays) / 30.0

	varValue := liquidAssets * cashFlowVolatility * zScore * math.Sqrt(horizonMonths)

	// 🔴 Корректировка на Liquidity Gap
	if liquidityGap < 0 {
		gapAdjustment := 1.0 + (math.Abs(liquidityGap) / liquidAssets) * 0.5
		varValue = varValue * gapAdjustment
		log.Printf("🔍 [LiquidityRisk] Gap adjustment: %.2f (deficit=%.2f)", gapAdjustment, math.Abs(liquidityGap))
	}

	// 12. 🔴 Exposure = ЛИКВИДНЫЕ АКТИВЫ
	exposureAmount := liquidAssets

	// 13. Стресс-тест
	stressLiquidAssets := liquidAssets * 0.70
	stressExpenses := monthlyOperatingExpenses * 1.20
	stressMonthsOfCoverage := stressLiquidAssets / stressExpenses
	stressTestLoss := liquidAssets - stressLiquidAssets
	if stressMonthsOfCoverage < 1.0 {
		stressTestLoss += (stressExpenses - stressLiquidAssets/stressMonthsOfCoverage) * 0.5
	}

	// 14. 🔴 Уровень риска
	riskLevel := s.determineLiquidityRiskLevel(monthsOfCoverage, liquidityGapRatio)

	// 15. Risk Ratio
	riskRatio := 0.0
	if exposureAmount > 0 {
		riskRatio = varValue / exposureAmount
	}
	log.Printf("🔍 [LiquidityRisk] VaR=%.2f (%.2f%% of liquid assets)", varValue, riskRatio*100)

	// 16. Рекомендации
	quickRatio := quickAssets / shortTermLiabilities
	currentRatio := liquidAssets / shortTermLiabilities
	recommendations := s.generateLiquidityRiskRecommendations(riskLevel, monthsOfCoverage, liquidityGapRatio, currentRatio)

	// 17. Метаданные
	assumptions, _ := json.Marshal(map[string]interface{}{
		"liquidAssets":             liquidAssets,
		"shortTermLiabilities":     shortTermLiabilities,
		"liquidityGap":             liquidityGap,
		"liquidityGapRatio":        liquidityGapRatio,
		"monthsOfCoverage":         monthsOfCoverage,
		"monthlyOperatingExpenses": monthlyOperatingExpenses,
		"cashFlowVolatility":       cashFlowVolatility,
		"cashFlowVolatilityAnn":    cashFlowVolatility * math.Sqrt(12),
		"zScore":                   zScore,
		"horizonMonths":            horizonMonths,
		"horizonDays":              horizonDays,
		"confidenceLevel":          confidenceLevel,
		"exchangeRate":             exchangeRate,
		"quickRatio":               quickRatio,
		"currentRatio":             currentRatio,
		"riskRatio":                riskRatio,
		"stressMonthsOfCoverage":   stressMonthsOfCoverage,
	})

	// 18. Создаём расчёт риска
	risk := &models.RiskCalculation{
		EnterpriseID:      enterpriseID,
		RiskType:          "liquidity",
		CalculationDate:   time.Now(),
		HorizonDays:       horizonDays,
		ConfidenceLevel:   confidenceLevel,
		ExposureAmount:    exposureAmount,
		VaRValue:          varValue,
		StressTestLoss:    stressTestLoss,
		RiskLevel:         riskLevel,
		CalculationMethod: stringPtr("cash_flow_at_risk"),
		ScenarioType:      stringPtr("base"),
		Assumptions:       stringPtr(string(assumptions)),
		Recommendations:   recommendations,
	}

	// 19. Сохраняем в БД
	if err := s.riskRepo.Create(ctx, risk); err != nil {
		return nil, fmt.Errorf("failed to save liquidity risk calculation: %w", err)
	}

	return risk, nil
}
// calculateMonthlyOperatingExpenses рассчитывает месячные операционные расходы
func (s *RiskCalculationService) calculateMonthlyOperatingExpenses(finResults *models.FinancialResult) float64 {
	// 🔴 ПРОВЕРКА НА NIL
	if finResults == nil {
		log.Printf("⚠️  [LiquidityRisk] No financial results, using fallback: 500M BYN/month")
		return 500_000_000 // Fallback: 500M BYN
	}

	// Операционные расходы = Себестоимость + Коммерческие + Административные
	quarterlyOperatingExpenses := finResults.CostTotal +
		finResults.CommercialExpenses +
		finResults.AdministrativeExpenses

	monthlyExpenses := quarterlyOperatingExpenses / 3.0

	log.Printf("🔍 [LiquidityRisk] Quarterly expenses: %.2f, Monthly: %.2f",
		quarterlyOperatingExpenses, monthlyExpenses)

	return monthlyExpenses
}
// calculateCashFlowVolatility рассчитывает волатильность денежных потоков
func (s *RiskCalculationService) calculateCashFlowVolatility(ctx context.Context, enterpriseID int64) float64 {
	// Получаем историю финансовых результатов
	results, err := s.finResultRepo.GetByEnterpriseID(ctx, enterpriseID)
	if err != nil || len(results) < 4 {
		log.Printf("⚠️  [LiquidityRisk] Not enough history for volatility calculation")
		return 0
	}

	// 🔴 Проверка на nil в массиве
	cashFlows := make([]float64, 0, len(results))
	for _, result := range results {
		if result == nil {
			continue
		}
		ocf := result.NetProfit + result.CostDepreciation
		cashFlows = append(cashFlows, ocf)
	}

	if len(cashFlows) < 2 {
		return 0
	}

	mean := 0.0
	for _, cf := range cashFlows {
		mean += cf
	}
	mean /= float64(len(cashFlows))

	if mean == 0 {
		return 0
	}

	variance := 0.0
	for _, cf := range cashFlows {
		variance += (cf - mean) * (cf - mean)
	}
	variance /= float64(len(cashFlows) - 1)

	stdDev := math.Sqrt(variance)
	coefficientOfVariation := stdDev / math.Abs(mean)
	monthlyVolatility := coefficientOfVariation / math.Sqrt(3)

	log.Printf("🔍 [LiquidityRisk] Cash Flow Volatility: %.2f%% (CV=%.4f)", monthlyVolatility*100, coefficientOfVariation)

	return monthlyVolatility
}
// determineLiquidityRiskLevel определяет уровень риска ликвидности
func (s *RiskCalculationService) determineLiquidityRiskLevel(monthsOfCoverage float64, gapRatio float64) string {
	// 🔴 Основной критерий: Months of Coverage
	if monthsOfCoverage < 1.5 {
		return "high" // Критический риск (< 1.5 месяцев)
	}
	if monthsOfCoverage < 3.0 {
		return "medium" // Средний риск (1.5-3 месяца)
	}
	return "low" // Низкий риск (> 3 месяца)
}

// generateLiquidityRiskRecommendations генерирует рекомендации
func (s *RiskCalculationService) generateLiquidityRiskRecommendations(
	riskLevel string,
	monthsOfCoverage float64,
	gapRatio float64,
	currentRatio float64,
) []string {
	baseRecommendations := []string{
		"Мониторить денежный поток еженедельно",
		"Поддерживать кредитную линию на 2-3 месяца операционных расходов",
		"Оптимизировать условия оплаты с поставщиками и клиентами",
	}

	switch riskLevel {
	case "high":
		return append([]string{
			fmt.Sprintf("⚠️ КРИТИЧЕСКИ: Запас ликвидности всего %.1f месяцев (норма 3-6)", monthsOfCoverage),
			fmt.Sprintf("⚠️ Коэффициент текущей ликвидности %.2f (норма ≥ 1.5)", currentRatio),
			"СРОЧНО: Привлечь краткосрочное финансирование",
			"Ускорить инкассацию дебиторской задолженности",
			"Отложить капитальные расходы и дивиденды",
			"Рассмотреть продажу неосновных активов",
		}, baseRecommendations...)
	case "medium":
		return append([]string{
			fmt.Sprintf("⚠️ Запас ликвидности %.1f месяцев (рекомендуется 3-6)", monthsOfCoverage),
			fmt.Sprintf("⚠️ Коэффициент текущей ликвидности %.2f (норма ≥ 1.5)", currentRatio),
			"Увеличить кредитную линию на 20-30%",
			"Сократить срок инкассации на 5-10 дней",
		}, baseRecommendations...)
	default:
		return append([]string{
			fmt.Sprintf("✅ Запас ликвидности %.1f месяцев - достаточный уровень", monthsOfCoverage),
			fmt.Sprintf("✅ Коэффициент текущей ликвидности %.2f", currentRatio),
			"Поддерживать текущий уровень ликвидности",
			"Рассмотреть размещение избыточных средств в краткосрочные инструменты",
		}, baseRecommendations...)
	}
}

// generateCurrencyRiskRecommendations генерирует рекомендации для валютного риска
func (s *RiskCalculationService) generateCurrencyRiskRecommendations(riskLevel string, contractsCount int) []string {
	baseRecommendations := []string{
		"Хеджировать 30-40% валютной экспозиции через форвардные контракты",
		"Сократить срок оплаты экспортных контрактов до 45-60 дней",
		"Диверсифицировать валюту расчётов (включить до 20% в CNY, EUR)",
	}

	switch riskLevel {
	case "high":
		return append([]string{
			"СРОЧНО: Увеличить долю хеджирования до 60-70%",
			"Рассмотреть опционные контракты для защиты от неблагоприятных движений курса",
			fmt.Sprintf("Приоритетная работа с %d незакрытыми контрактами", contractsCount),
		}, baseRecommendations...)
	case "medium":
		return baseRecommendations
	default:
		return []string{
			"Продолжать мониторинг валютных позиций",
			"Поддерживать текущую стратегию хеджирования",
		}
	}
}

// generateInterestRiskRecommendations генерирует рекомендации для процентного риска
func (s *RiskCalculationService) generateInterestRiskRecommendations(riskLevel string, floatingCount int, keyRate float64) []string {
	baseRecommendations := []string{
		"Поддерживать текущую структуру капитала с фиксированной ставкой",
		"Мониторинг процентных ставок 1 раз в месяц",
		"Рассмотреть рефинансирование части долга при снижении ставок",
	}

	switch riskLevel {
	case "high":
		return append([]string{
			"СРОЧНО: Конвертировать плавающие ставки в фиксированные через свопы",
			"Рассмотреть досрочное погашение наиболее дорогих кредитов",
			fmt.Sprintf("Текущая ключевая ставка: %.2f%%", keyRate*100),
		}, baseRecommendations...)
	case "medium":
		return baseRecommendations
	default:
		return []string{
			"Продолжать мониторинг ключевой ставки НБ РБ",
			"Поддерживать текущую структуру долга",
		}
	}
}

// createEmptyRiskCalculation создаёт пустой расчёт риска (когда нет данных для расчёта)
func (s *RiskCalculationService) createEmptyRiskCalculation(
	enterpriseID int64,
	riskType string,
	horizonDays int,
	confidenceLevel float64,
) *models.RiskCalculation {
	assumptions, _ := json.Marshal(map[string]interface{}{
		"reason":  "no_data",
		"message": "Недостаточно данных для расчёта риска",
	})

	return &models.RiskCalculation{
		EnterpriseID:      enterpriseID,
		RiskType:          riskType,
		CalculationDate:   time.Now(),
		HorizonDays:       horizonDays,
		ConfidenceLevel:   confidenceLevel,
		ExposureAmount:    0,
		VaRValue:          0,
		StressTestLoss:    0,
		RiskLevel:         "low",
		CalculationMethod: stringPtr("no_data"),
		ScenarioType:      stringPtr("base"),
		Assumptions:       stringPtr(string(assumptions)),
		Recommendations:   []string{"Недостаточно данных для расчёта риска"},
	}
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
		return 2.33
	case confidenceLevel >= 0.95:
		return 1.65
	default:
		return 1.28
	}
}

// Вспомогательные функции
func stringPtr(s string) *string {
	return &s
}

func float64Ptr(f float64) *float64 {
	return &f
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

// GetRiskCalculationsByEnterprise получает все расчёты для предприятия
func (s *RiskCalculationService) GetRiskCalculationsByEnterprise(
	ctx context.Context,
	enterpriseID int64,
	limit int,
	offset int,
) ([]*models.RiskCalculation, error) {
	filter := interfaces.RiskCalculationFilter{
		EnterpriseID: &enterpriseID,
		Limit:        limit,
		Offset:       offset,
	}
	return s.riskRepo.GetAll(ctx, filter)
}

// GetLatestRiskCalculations получает последние расчёты для каждого типа риска
func (s *RiskCalculationService) GetLatestRiskCalculations(
	ctx context.Context,
	enterpriseID int64,
) (map[string]*models.RiskCalculation, error) {
	results := make(map[string]*models.RiskCalculation)

	riskTypes := []string{"currency", "interest", "liquidity"}

	for _, riskType := range riskTypes {
		filter := interfaces.RiskCalculationFilter{
			EnterpriseID: &enterpriseID,
			RiskType:     &riskType,
			Limit:        1,
			Offset:       0,
		}

		calculations, err := s.riskRepo.GetAll(ctx, filter)
		if err != nil {
			continue
		}

		if len(calculations) > 0 {
			results[riskType] = calculations[0]
		}
	}

	return results, nil
}
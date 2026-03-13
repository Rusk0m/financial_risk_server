package service

import (
	"context"
	"fmt"
	"time"

	"financial-risk-server/internal/domain/models"
	"financial-risk-server/internal/repository/interfaces"
)

// ComprehensiveReportService предоставляет бизнес-логику для генерации комплексных отчётов
type ComprehensiveReportService struct {
	reportRepo        interfaces.ComprehensiveRiskReportRepository
	riskRepo          interfaces.RiskCalculationRepository
	enterpriseService *EnterpriseService
}

// NewComprehensiveReportService создаёт новый сервис комплексных отчётов
func NewComprehensiveReportService(
	reportRepo interfaces.ComprehensiveRiskReportRepository,
	riskRepo interfaces.RiskCalculationRepository,
	enterpriseService *EnterpriseService,
) *ComprehensiveReportService {
	return &ComprehensiveReportService{
		reportRepo:        reportRepo,
		riskRepo:          riskRepo,
		enterpriseService: enterpriseService,
	}
}

// GenerateReport генерирует комплексный отчёт по рискам
func (s *ComprehensiveReportService) GenerateReport(
	ctx context.Context,
	enterpriseID int64,
) (*models.ComprehensiveRiskReport, error) {
	// 1. Получаем последние расчёты по каждому типу риска
	currencyRisk, err := s.riskRepo.GetLatestByType(ctx, enterpriseID, "currency")
	if err != nil {
		return nil, fmt.Errorf("failed to get latest currency risk: %w", err)
	}

	interestRisk, err := s.riskRepo.GetLatestByType(ctx, enterpriseID, "interest")
	if err != nil {
		return nil, fmt.Errorf("failed to get latest interest risk: %w", err)
	}

	liquidityRisk, err := s.riskRepo.GetLatestByType(ctx, enterpriseID, "liquidity")
	if err != nil {
		return nil, fmt.Errorf("failed to get latest liquidity risk: %w", err)
	}

	// 2. Рассчитываем суммарный риск
	totalRiskValue := currencyRisk.VaRValue + interestRisk.VaRValue + liquidityRisk.VaRValue

	// 3. Определяем максимальный риск
	maxRiskType := "currency"
	maxRiskValue := currencyRisk.VaRValue

	if interestRisk.VaRValue > maxRiskValue {
		maxRiskType = "interest"
		maxRiskValue = interestRisk.VaRValue
	}
	if liquidityRisk.VaRValue > maxRiskValue {
		maxRiskType = "liquidity"
		maxRiskValue = liquidityRisk.VaRValue
	}

	// 4. Определяем общий уровень риска
	overallRiskLevel := s.determineOverallRiskLevel(
		currencyRisk.RiskLevel,
		interestRisk.RiskLevel,
		liquidityRisk.RiskLevel,
	)

	// 5. Формируем краткое резюме
	summary := s.generateSummary(
		overallRiskLevel,
		totalRiskValue,
		maxRiskType,
		maxRiskValue,
	)

	// 6. Формируем приоритетные действия
	priorityActions := s.generatePriorityActions(
		currencyRisk,
		interestRisk,
		liquidityRisk,
	)

	// 7. Формируем предупреждения
	warnings := s.generateWarnings(
		currencyRisk,
		interestRisk,
		liquidityRisk,
	)

	// 8. Создаём комплексный отчёт
	report := &models.ComprehensiveRiskReport{
		EnterpriseID:       enterpriseID,
		ReportDate:         time.Now(),
		OverallRiskLevel:   overallRiskLevel,
		TotalRiskValue:     totalRiskValue,
		MaxRiskType:        maxRiskType,
		MaxRiskValue:       maxRiskValue,
		CurrencyRiskID:     &currencyRisk.ID,
		InterestRiskID:     &interestRisk.ID,
		LiquidityRiskID:    &liquidityRisk.ID,
		Summary:            summary,
		PriorityActions:    priorityActions,
		Warnings:           warnings,
		CreatedBy:          stringPtr("System"),
	}

	// 9. Сохраняем в БД
	if err := s.reportRepo.Create(ctx, report); err != nil {
		return nil, fmt.Errorf("failed to save comprehensive risk report: %w", err)
	}

	return report, nil
}

// GetAllReports получает все комплексные отчёты
func (s *ComprehensiveReportService) GetAllReports(ctx context.Context) ([]*models.ComprehensiveRiskReport, error) {
	return s.reportRepo.GetAll(ctx)
}

// GetAllReportsByEnterpriseID получает все отчёты для предприятия
func (s *ComprehensiveReportService) GetAllReportsByEnterpriseID(ctx context.Context, enterpriseID int64) ([]*models.ComprehensiveRiskReport, error) {
	return s.reportRepo.GetAllByEnterpriseID(ctx, enterpriseID)
}

// GetReportByID получает отчёт по ID
func (s *ComprehensiveReportService) GetReportByID(ctx context.Context, id int64) (*models.ComprehensiveRiskReport, error) {
	return s.reportRepo.GetByID(ctx, id)
}

// determineOverallRiskLevel определяет общий уровень риска на основе трёх типов
func (s *ComprehensiveReportService) determineOverallRiskLevel(
	currencyLevel, interestLevel, liquidityLevel string,
) string {
	// Если хотя бы один риск высокий - общий уровень высокий
	if currencyLevel == "high" || interestLevel == "high" || liquidityLevel == "high" {
		return "high"
	}
	
	// Если хотя бы два риска средние - общий уровень средний
	mediumCount := 0
	if currencyLevel == "medium" {
		mediumCount++
	}
	if interestLevel == "medium" {
		mediumCount++
	}
	if liquidityLevel == "medium" {
		mediumCount++
	}
	
	if mediumCount >= 2 {
		return "medium"
	}
	
	// Иначе низкий
	return "low"
}

// generateSummary формирует краткое резюме отчёта
func (s *ComprehensiveReportService) generateSummary(
	overallRiskLevel string,
	totalRiskValue float64,
	maxRiskType string,
	maxRiskValue float64,
) string {
	var levelDesc string
	switch overallRiskLevel {
	case "high":
		levelDesc = "критический"
	case "medium":
		levelDesc = "средний"
	case "low":
		levelDesc = "низкий"
	default:
		levelDesc = "неопределённый"
	}

	var riskTypeDesc string
	switch maxRiskType {
	case "currency":
		riskTypeDesc = "валютного"
	case "interest":
		riskTypeDesc = "процентного"
	case "liquidity":
		riskTypeDesc = "ликвидности"
	}

	return fmt.Sprintf(
		"Общий уровень риска: %s. Суммарный риск: $%.2f млн. Максимальный риск: %s ($%.2f млн).",
		levelDesc,
		totalRiskValue/1000000,
		riskTypeDesc,
		maxRiskValue/1000000,
	)
}

// generatePriorityActions формирует список приоритетных действий
func (s *ComprehensiveReportService) generatePriorityActions(
	currencyRisk, interestRisk, liquidityRisk *models.RiskCalculation,
) []string {
	actions := []string{}

	// Добавляем действия для высоких рисков
	if currencyRisk.IsHighRisk() {
		actions = append(actions, "Срочно хеджировать валютную экспозицию")
	}
	if interestRisk.IsHighRisk() {
		actions = append(actions, "Пересмотреть структуру долгового портфеля")
	}
	if liquidityRisk.IsHighRisk() {
		actions = append(actions, "Усилить управление ликвидностью")
	}

	// Если нет высоких рисков, добавляем общие рекомендации
	if len(actions) == 0 {
		actions = append(actions, "Продолжить мониторинг ключевых рисков")
		actions = append(actions, "Оптимизировать структуру затрат")
		actions = append(actions, "Диверсифицировать экспортные рынки")
	}

	// Ограничиваем до 5 действий
	if len(actions) > 5 {
		actions = actions[:5]
	}

	return actions
}

// generateWarnings формирует список предупреждений
func (s *ComprehensiveReportService) generateWarnings(
	currencyRisk, interestRisk, liquidityRisk *models.RiskCalculation,
) []string {
	warnings := []string{}

	if currencyRisk.IsHighRisk() {
		warnings = append(warnings, "Валютный риск превышает допустимый уровень")
	}
	if interestRisk.IsHighRisk() {
		warnings = append(warnings, "Процентный риск требует немедленных действий")
	}
	if liquidityRisk.IsHighRisk() {
		warnings = append(warnings, "Критический уровень риска ликвидности")
	}

	// Добавляем предупреждения для средних рисков
	if !currencyRisk.IsHighRisk() && currencyRisk.IsMediumRisk() {
		warnings = append(warnings, "Валютный риск находится на среднем уровне")
	}
	if !interestRisk.IsHighRisk() && interestRisk.IsMediumRisk() {
		warnings = append(warnings, "Процентный риск требует внимания")
	}
	if !liquidityRisk.IsHighRisk() && liquidityRisk.IsMediumRisk() {
		warnings = append(warnings, "Риск ликвидности на контроле")
	}

	// Ограничиваем до 5 предупреждений
	if len(warnings) > 5 {
		warnings = warnings[:5]
	}

	return warnings
}

// stringPtr вспомогательная функция
// func stringPtr(s string) *string {
// 	return &s
// }
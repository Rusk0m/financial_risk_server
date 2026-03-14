import 'package:client_flutter/core/constants/app_colors.dart';
import 'package:client_flutter/presentation/widgets/animated_card.dart';
import 'package:client_flutter/presentation/widgets/risk_card.dart';
import 'package:client_flutter/presentation/widgets/risk_heatmap.dart';
import 'package:client_flutter/presentation/widgets/risk_pie_chart.dart';
import 'package:client_flutter/presentation/widgets/risk_trend_chart.dart';
import 'package:flutter/material.dart';

class RiskResultsWidget extends StatelessWidget {
  final Map<String, dynamic> results;

  const RiskResultsWidget({
    super.key,
    required this.results,
  });

  @override
  Widget build(BuildContext context) {
    // Извлекаем данные о рисках ТОЛЬКО для ключей, которые точно содержат данные риска
    final risks = [
      if (results['currency_risk'] is Map) results['currency_risk'] as Map<String, dynamic>,
      if (results['credit_risk'] is Map) results['credit_risk'] as Map<String, dynamic>,
      if (results['liquidity_risk'] is Map) results['liquidity_risk'] as Map<String, dynamic>,
      if (results['market_risk'] is Map) results['market_risk'] as Map<String, dynamic>,
      if (results['interest_risk'] is Map) results['interest_risk'] as Map<String, dynamic>,
    ];

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        // Заголовок
        Row(
          children: [
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
              decoration: BoxDecoration(
                color: AppColors.primaryLight,
                borderRadius: BorderRadius.circular(20),
              ),
              child: const Text(
                'Результаты расчёта',
                style: TextStyle(
                  fontSize: 14,
                  fontWeight: FontWeight.w600,
                  color: AppColors.primaryDark,
                ),
              ),
            ),
          ],
        ),
        const SizedBox(height: 16),
        
        // Сводка
        _buildSummaryCard(results, risks),
        const SizedBox(height: 24),
        
        // Круговая диаграмма
        RiskPieChart(risks: risks),
        const SizedBox(height: 24),
        
        // Тепловая карта
        RiskHeatmapChart(risks: risks),
        const SizedBox(height: 24),
        
        // Динамика рисков
        const RiskTrendChart(historicalData: []),
        const SizedBox(height: 24),
        
        // Детальные карточки рисков с анимацией
        ...risks.asMap().entries.map((entry) => 
          AnimatedCard(
            delay: Duration(milliseconds: entry.key * 150),
            child: RiskCard(
              title: _getRiskName(_toString(entry.value['risk_type'])),
              riskLevel: _getRiskLevel(_toString(entry.value['risk_level'])),
              exposure: _toDouble(entry.value['exposure_amount']),
              varValue: _toDouble(entry.value['var_value']),
              stressTest: _toDouble(entry.value['stress_test_loss']),
              recommendations: [
                'Рекомендация 1 для ${_getRiskName(_toString(entry.value['risk_type']))}',
                'Рекомендация 2 для ${_getRiskName(_toString(entry.value['risk_type']))}',
                'Рекомендация 3 для ${_getRiskName(_toString(entry.value['risk_type']))}',
              ],
              icon: _getRiskIcon(_toString(entry.value['risk_type'])),
            ),
          ),
        ),
      ],
    );
  }

  Widget _buildSummaryCard(Map<String, dynamic> results, List<Map<String, dynamic>> risks) {
    // Рассчитываем суммарный риск с безопасным преобразованием типов
    final totalRisk = risks.fold<double>(0, (sum, risk) => 
      sum + _toDouble(risk['var_value'])
    );
    
    // Находим максимальный риск
    MapEntry<String, double>? maxRiskEntry;
    if (risks.isNotEmpty) {
      maxRiskEntry = risks
          .map((risk) => MapEntry(
                _toString(risk['risk_type']),
                _toDouble(risk['var_value']),
              ))
          .reduce((a, b) => a.value > b.value ? a : b);
    }
    
    // Проверяем наличие высоких рисков
    final hasHighRisk = risks.any((risk) => _toString(risk['risk_level']) == 'high');
    
    return Card(
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
      color: hasHighRisk 
          ? Colors.red.withOpacity(0.05) 
          : Colors.blue.withOpacity(0.05),
      child: Padding(
        padding: const EdgeInsets.all(20.0),
        child: Column(
          children: [
            Row(
              children: [
                Container(
                  padding: const EdgeInsets.all(12),
                  decoration: BoxDecoration(
                    color: hasHighRisk ? AppColors.riskHigh : AppColors.primary,
                    shape: BoxShape.circle,
                  ),
                  child: Icon(
                    hasHighRisk ? Icons.warning : Icons.analytics,
                    color: Colors.white,
                    size: 28,
                  ),
                ),
                const SizedBox(width: 16),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        hasHighRisk ? 'КРИТИЧЕСКИЙ УРОВЕНЬ РИСКА' : 'СРЕДНИЙ УРОВЕНЬ РИСКА',
                        style: TextStyle(
                          fontSize: 18,
                          fontWeight: FontWeight.bold,
                          color: hasHighRisk ? AppColors.riskHigh : AppColors.primary,
                        ),
                      ),
                      const SizedBox(height: 4),
                      Text(
                        hasHighRisk
                            ? 'Требуются немедленные действия для снижения рисков'
                            : 'Рекомендуется регулярный мониторинг и оптимизация',
                        style: const TextStyle(
                          fontSize: 14,
                          color: AppColors.textSecondary,
                        ),
                      ),
                    ],
                  ),
                ),
              ],
            ),
            const SizedBox(height: 20),
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceAround,
              children: [
                _buildSummaryItem(
                  'Суммарный риск',
                  '\$${(totalRisk / 1000000).toStringAsFixed(1)} млн',
                  Icons.trending_up,
                  AppColors.primary,
                ),
                _buildSummaryItem(
                  'Критический риск',
                  maxRiskEntry != null ? _getRiskName(maxRiskEntry.key) : '—',
                  Icons.warning,
                  AppColors.riskHigh,
                ),
                _buildSummaryItem(
                  'Рекомендаций',
                  '15',
                  Icons.lightbulb,
                  Colors.amber,
                ),
              ],
            ),
            const SizedBox(height: 20),
            SizedBox(
              width: double.infinity,
              child: OutlinedButton.icon(
                onPressed: () {
                  // Экспорт отчёта
                },
                icon: const Icon(Icons.picture_as_pdf, color: AppColors.primary),
                label: const Text(
                  'ЭКСПОРТИРОВАТЬ ОТЧЁТ В PDF',
                  style: TextStyle(
                    fontSize: 16,
                    fontWeight: FontWeight.w700,
                    color: AppColors.primary,
                  ),
                ),
                style: OutlinedButton.styleFrom(
                  padding: const EdgeInsets.symmetric(vertical: 14),
                  side: const BorderSide(color: AppColors.primary, width: 2),
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(12),
                  ),
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildSummaryItem(String label, String value, IconData icon, Color color) {
    return Column(
      children: [
        Icon(icon, color: color, size: 24),
        const SizedBox(height: 8),
        Text(
          label,
          style: const TextStyle(
            fontSize: 12,
            color: AppColors.textSecondary,
          ),
          textAlign: TextAlign.center,
        ),
        const SizedBox(height: 4),
        Text(
          value,
          style: const TextStyle(
            fontSize: 18,
            fontWeight: FontWeight.bold,
            color: AppColors.textPrimary,
          ),
          textAlign: TextAlign.center,
        ),
      ],
    );
  }

  // Вспомогательные функции для безопасного преобразования типов
  double _toDouble(dynamic value) {
    if (value == null) return 0.0;
    if (value is num) return value.toDouble();
    if (value is String) return double.tryParse(value) ?? 0.0;
    return 0.0;
  }

  String _toString(dynamic value) {
    if (value == null) return '';
    if (value is String) return value;
    return value.toString();
  }

  String _getRiskName(String riskType) {
    switch (riskType) {
      case 'currency': return 'Валютный';
      case 'credit': return 'Кредитный';
      case 'liquidity': return 'Ликвидность';
      case 'market': return 'Фондовый';
      case 'interest': return 'Процентный';
      default: return 'Риск';
    }
  }

  RiskLevel _getRiskLevel(String level) {
    switch (level) {
      case 'low': return RiskLevel.low;
      case 'medium': return RiskLevel.medium;
      case 'high': return RiskLevel.high;
      default: return RiskLevel.medium;
    }
  }

  String _getRiskIcon(String riskType) {
    switch (riskType) {
      case 'currency': return '💱';
      case 'credit': return '💳';
      case 'liquidity': return '💧';
      case 'market': return '📉';
      case 'interest': return '📈';
      default: return '⚠️';
    }
  }
}
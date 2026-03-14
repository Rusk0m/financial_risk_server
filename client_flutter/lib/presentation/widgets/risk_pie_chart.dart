import 'package:client_flutter/core/constants/app_colors.dart';
import 'package:flutter/material.dart';
import 'package:fl_chart/fl_chart.dart';

class RiskPieChart extends StatelessWidget {
  final List<Map<String, dynamic>> risks;

  const RiskPieChart({
    super.key,
    required this.risks,
  });

  @override
  Widget build(BuildContext context) {
final totalRisk = risks.fold<double>(0, (sum, risk) => 
      sum + _toDouble(risk['var_value'])
    );     
    return Card(
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
      child: Padding(
        padding: const EdgeInsets.all(16.0),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                const Icon(Icons.pie_chart, color: AppColors.primary, size: 20),
                const SizedBox(width: 8),
                const Text(
                  'Структура финансовых рисков',
                  style: TextStyle(
                    fontSize: 18,
                    fontWeight: FontWeight.bold,
                    color: AppColors.textPrimary,
                  ),
                ),
              ],
            ),
            const SizedBox(height: 24),
            Row(
              children: [
                // Круговая диаграмма
                Expanded(
                  child: AspectRatio(
                    aspectRatio: 1,
                    child: PieChart(
                      PieChartData(
                        sections: _generateSections(totalRisk),
                        sectionsSpace: 2,
                        centerSpaceRadius: 60,
                        //sectionsRadius: 80,
                        pieTouchData: PieTouchData(
                          touchCallback: (FlTouchEvent event, pieTouchResponse) {
                            // Обработка касаний
                          },
                        ),
                      ),
                    ),
                  ),
                ),
                const SizedBox(width: 24),
                // Легенда
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: risks.asMap().entries.map((entry) {
                      final risk = entry.value;
                      final color = _getRiskColor(entry.key);
                      
                      return Padding(
                        padding: const EdgeInsets.only(bottom: 16),
                        child: Row(
                          children: [
                            Container(
                              width: 16,
                              height: 16,
                              decoration: BoxDecoration(
                                color: color,
                                shape: BoxShape.circle,
                                border: Border.all(color: Colors.white, width: 2),
                              ),
                            ),
                            const SizedBox(width: 12),
                            Expanded(
                              child: Column(
                                crossAxisAlignment: CrossAxisAlignment.start,
                                children: [
                                  Text(
                                    _getRiskName(risk['risk_type'] as String),
                                    style: const TextStyle(
                                      fontSize: 14,
                                      fontWeight: FontWeight.w600,
                                      color: AppColors.textPrimary,
                                    ),
                                  ),
                                  const SizedBox(height: 2),
                                  Text(
                                    '\$${(_toDouble(risk['var_value'] ) / 1000000).toStringAsFixed(1)} млн (${(_toDouble(risk['var_value'] ) / totalRisk * 100).toStringAsFixed(0)}%)',
                                    style: const TextStyle(
                                      fontSize: 12,
                                      color: AppColors.textSecondary,
                                    ),
                                  ),
                                ],
                              ),
                            ),
                          ],
                        ),
                      );
                    }).toList(),
                  ),
                ),
              ],
            ),
            const SizedBox(height: 12),
             Text(
              'Суммарный риск: \$${(totalRisk / 1000000).toStringAsFixed(1)} млн',
              style: TextStyle(
                fontSize: 16,
                fontWeight: FontWeight.bold,
                color: AppColors.textPrimary,
              ),
              textAlign: TextAlign.center,
            ),
          ],
        ),
      ),
    );
  }

  List<PieChartSectionData> _generateSections(double totalRisk) {
    return risks.asMap().entries.map((entry) {
      final risk = entry.value;
      final index = entry.key;
      final value = _toDouble(['var_value'] );
      final percentage = (value / totalRisk * 100).roundToDouble();
      final color = _getRiskColor(index);
      
      return PieChartSectionData(
        value: value,
        title: '${percentage.toStringAsFixed(0)}%',
        color: color,
        radius: 70,
        titleStyle: const TextStyle(
          fontSize: 14,
          fontWeight: FontWeight.bold,
          color: Colors.white,
        ),
        badgeWidget: index == 0
            ? Container(
                width: 20,
                height: 20,
                decoration: BoxDecoration(
                  color: Colors.white,
                  shape: BoxShape.circle,
                  border: Border.all(color: color, width: 2),
                ),
                child: Center(
                  child: Text(
                    _getRiskIcon(risk['risk_type'] as String),
                    style: const TextStyle(fontSize: 10),
                  ),
                ),
              )
            : null,
        badgePositionPercentageOffset: .98,
      );
    }).toList();
  }

  Color _getRiskColor(int index) {
    // Цвета для каждого типа риска
    final colors = [
      AppColors.riskHigh,    // Фондовый
      AppColors.riskMedium,  // Валютный
      AppColors.riskMedium,  // Кредитный
      AppColors.riskMedium,  // Ликвидность
      AppColors.riskLow,     // Процентный
    ];
    return colors[index % colors.length];
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
double _toDouble(dynamic value) {
    if (value == null) return 0.0;
    if (value is num) return value.toDouble();
    if (value is String) return double.tryParse(value) ?? 0.0;
    return 0.0;
  }
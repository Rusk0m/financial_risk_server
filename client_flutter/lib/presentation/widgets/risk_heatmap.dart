import 'package:client_flutter/core/constants/app_colors.dart';
import 'package:flutter/material.dart';

class RiskHeatmapChart extends StatelessWidget {
  final List<Map<String, dynamic>> risks;

  const RiskHeatmapChart({
    super.key,
    required this.risks,
  });

  @override
  Widget build(BuildContext context) {
    return Card(
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
      child: Padding(
        padding: const EdgeInsets.all(16.0),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                const Icon(Icons.thermostat, color: AppColors.primary, size: 20),
                const SizedBox(width: 8),
                const Text(
                  'Тепловая карта рисков',
                  style: TextStyle(
                    fontSize: 18,
                    fontWeight: FontWeight.bold,
                    color: AppColors.textPrimary,
                  ),
                ),
              ],
            ),
            const SizedBox(height: 16),
            SizedBox(
              height: 300,
              child: Stack(
                children: [
                  // Фоновая сетка
                  Container(
                    decoration: BoxDecoration(
                      gradient: LinearGradient(
                        begin: Alignment.bottomLeft,
                        end: Alignment.topRight,
                        colors: [
                          AppColors.riskLow.withOpacity(0.1),
                          AppColors.riskMedium.withOpacity(0.1),
                          AppColors.riskHigh.withOpacity(0.15),
                        ],
                      ),
                      borderRadius: BorderRadius.circular(12),
                    ),
                  ),
                  
                  // Матрица рисков
                  Padding(
                    padding: const EdgeInsets.all(8.0),
                    child: Column(
                      mainAxisAlignment: MainAxisAlignment.spaceEvenly,
                      children: risks.asMap().entries.map((entry) {
                        final index = entry.key;
                        final risk = entry.value;
                        final riskLevel = _getRiskLevel(_toDouble(risk['var_value']));
                        final color = _getRiskColor(riskLevel);
                        
                        return Expanded(
                          child: Row(
                            mainAxisAlignment: MainAxisAlignment.spaceEvenly,
                            children: [
                              _buildRiskBubble(
                                context,
                                risk['risk_type'] as String,
                                _toDouble(['var_value'] ),
                                _toDouble(risk['exposure_amount'] ),
                                color,
                                index,
                              ),
                            ],
                          ),
                        );
                      }).toList(),
                    ),
                  ),
                  
                  // Легенда
                  Positioned(
                    bottom: 16,
                    left: 16,
                    right: 16,
                    child: Row(
                      mainAxisAlignment: MainAxisAlignment.spaceEvenly,
                      children: [
                        _buildLegendItem(AppColors.riskLow, 'Низкий риск'),
                        _buildLegendItem(AppColors.riskMedium, 'Средний риск'),
                        _buildLegendItem(AppColors.riskHigh, 'Высокий риск'),
                      ],
                    ),
                  ),
                ],
              ),
            ),
            const SizedBox(height: 12),
            const Text(
              'Матрица вероятность × воздействие: чем темнее цвет, тем выше риск',
              style: TextStyle(
                fontSize: 12,
                color: AppColors.textSecondary,
                fontStyle: FontStyle.italic,
              ),
              textAlign: TextAlign.center,
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildRiskBubble(
    BuildContext context,
    String riskType,
    double varValue,
    double exposure,
    Color color,
    int index,
  ) {
    final riskPercentage = (varValue / exposure * 100).clamp(5.0, 30.0);
    final size = 40.0 + (riskPercentage * 2);
    
    return GestureDetector(
      onTap: () {
        // Показать детали риска
        showDialog(
          context: context,
          builder: (context) => _buildRiskDetailDialog(context,riskType, varValue, exposure, color),
        );
      },
      child: Container(
        width: size,
        height: size,
        decoration: BoxDecoration(
          color: color.withOpacity(0.7),
          shape: BoxShape.circle,
          border: Border.all(color: color, width: 2),
          boxShadow: [
            BoxShadow(
              color: color.withOpacity(0.3),
              blurRadius: 10,
              spreadRadius: 2,
            ),
          ],
        ),
        child: Center(
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Text(
                _getRiskIcon(riskType),
                style: const TextStyle(fontSize: 16, fontWeight: FontWeight.bold),
              ),
              const SizedBox(height: 2),
              Text(
                '${riskPercentage.toStringAsFixed(1)}%',
                style: TextStyle(
                  fontSize: 10,
                  fontWeight: FontWeight.bold,
                  color: color.computeLuminance() > 0.5 ? AppColors.textPrimary : Colors.white,
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildLegendItem(Color color, String label) {
    return Row(
      children: [
        Container(
          width: 12,
          height: 12,
          decoration: BoxDecoration(
            color: color,
            shape: BoxShape.circle,
          ),
        ),
        const SizedBox(width: 4),
        Text(
          label,
          style: const TextStyle(
            fontSize: 11,
            color: AppColors.textSecondary,
          ),
        ),
      ],
    );
  }

  Widget _buildRiskDetailDialog(BuildContext context,String riskType, double varValue, double exposure, Color color) {
    return AlertDialog(
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
      title: Row(
        children: [
          Container(
            width: 32,
            height: 32,
            decoration: BoxDecoration(
              color: color.withOpacity(0.2),
              shape: BoxShape.circle,
            ),
            child: Center(
              child: Text(
                _getRiskIcon(riskType),
                style: TextStyle(fontSize: 16, color: color),
              ),
            ),
          ),
          const SizedBox(width: 12),
          Text(
            _getRiskName(riskType),
            style: const TextStyle(fontWeight: FontWeight.bold),
          ),
        ],
      ),
      content: Column(
        mainAxisSize: MainAxisSize.min,
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          _buildDialogRow('Экспозиция', '\$${(exposure / 1000000).toStringAsFixed(2)} млн'),
          _buildDialogRow('VaR (95%)', '\$${(varValue / 1000000).toStringAsFixed(2)} млн'),
          _buildDialogRow('Уровень риска', _getRiskLevelText(varValue / exposure)),
          const SizedBox(height: 16),
          const Text(
            'Рекомендации:',
            style: TextStyle(fontWeight: FontWeight.bold),
          ),
          const SizedBox(height: 8),
          ..._getRecommendations(riskType).map((rec) => Padding(
            padding: const EdgeInsets.only(bottom: 4),
            child: Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const Text('• ', style: TextStyle(color: AppColors.primary)),
                Expanded(child: Text(rec, style: const TextStyle(fontSize: 13))),
              ],
            ),
          )),
        ],
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.pop(context),
          child: const Text('Закрыть'),
        ),
      ],
    );
  }

  Widget _buildDialogRow(String label, String value) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 8),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Text(label, style: const TextStyle(color: AppColors.textSecondary)),
          Text(value, style: const TextStyle(fontWeight: FontWeight.bold)),
        ],
      ),
    );
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

  String _getRiskName(String riskType) {
    switch (riskType) {
      case 'currency': return 'Валютный риск';
      case 'credit': return 'Кредитный риск';
      case 'liquidity': return 'Риск ликвидности';
      case 'market': return 'Фондовый риск';
      case 'interest': return 'Процентный риск';
      default: return 'Риск';
    }
  }

  String _getRiskLevelText(double ratio) {
    if (ratio < 0.05) return 'Низкий';
    if (ratio < 0.15) return 'Средний';
    return 'Высокий';
  }

  List<String> _getRecommendations(String riskType) {
    switch (riskType) {
      case 'currency':
        return [
          'Хеджировать 30-40% валютной экспозиции',
          'Сократить срок оплаты до 45-60 дней',
          'Диверсифицировать валюту расчётов',
        ];
      case 'market':
        return [
          'Оптимизировать себестоимость на 10-15%',
          'Диверсифицировать продуктовую линейку',
          'Усилить мониторинг мировых цен',
        ];
      default:
        return [
          'Регулярный мониторинг показателей',
          'Оптимизация структуры затрат',
          'Разработка плана действий',
        ];
    }
  }

  int _getRiskLevel(double varValue) {
    if (varValue < 10000000) return 0; // low
    if (varValue < 100000000) return 1; // medium
    return 2; // high
  }

  Color _getRiskColor(int level) {
    switch (level) {
      case 0: return AppColors.riskLow;
      case 1: return AppColors.riskMedium;
      case 2: return AppColors.riskHigh;
      default: return AppColors.textSecondary;
    }
  }
}
double _toDouble(dynamic value) {
    if (value == null) return 0.0;
    if (value is num) return value.toDouble();
    if (value is String) return double.tryParse(value) ?? 0.0;
    return 0.0;
  }
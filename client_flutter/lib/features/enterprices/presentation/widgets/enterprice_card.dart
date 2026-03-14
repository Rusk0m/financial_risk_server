import 'package:flutter/material.dart';
import 'package:client_flutter/core/constants/app_colors.dart';
import 'package:client_flutter/features/enterprices/domain/entities/enterprise.dart';

class EnterpriseCard extends StatelessWidget {
  final Enterprise enterprise;
  final bool isSelected;
  final VoidCallback onTap;

  const EnterpriseCard({
    super.key,
    required this.enterprise,
    required this.isSelected,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    // Безопасное получение данных с значениями по умолчанию
    final name = enterprise.name.isNotEmpty ? enterprise.name : 'Без названия';
    final industry = enterprise.industry.isNotEmpty ? enterprise.industry : 'Не указана';
    final production = enterprise.annualProductionT > 0 
        ? '${(enterprise.annualProductionT / 1000000).toStringAsFixed(1)} млн т' 
        : '0 т';
    final exportPercent = enterprise.exportSharePercent >= 0 && enterprise.exportSharePercent <= 100
        ? '${enterprise.exportSharePercent.toStringAsFixed(0)}%'
        : '0%';
    final currency = enterprise.mainCurrency.isNotEmpty ? enterprise.mainCurrency : 'USD';
    
    // Определение цвета на основе доли экспорта (без использования свойства exportRiskLevel)
    final Color riskColor;
    final String riskLabel;
    
    if (enterprise.exportSharePercent > 90) {
      riskColor = AppColors.riskHigh;
      riskLabel = 'Очень высокий';
    } else if (enterprise.exportSharePercent > 75) {
      riskColor = Colors.red;
      riskLabel = 'Высокий';
    } else if (enterprise.exportSharePercent > 50) {
      riskColor = AppColors.riskMedium;
      riskLabel = 'Средний';
    } else {
      riskColor = AppColors.riskLow;
      riskLabel = 'Низкий';
    }
    
    // Определение иконки отрасли
    final icon = _getIndustryIcon(industry);

    return Card(
      // ФИКСИРОВАННАЯ ВЫСОТА ДЛЯ СТАБИЛЬНОСТИ
      clipBehavior: Clip.antiAlias,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(16),
      ),
      elevation: isSelected ? 4 : 2,
      margin: EdgeInsets.zero,
      child: SizedBox(
        height: 190, // Строго фиксированная высота
        child: InkWell(
          onTap: onTap,
          child: Container(
            decoration: BoxDecoration(
              gradient: isSelected
                  ? LinearGradient(
                      colors: [AppColors.primaryLight.withOpacity(0.3), Colors.white],
                      begin: Alignment.topLeft,
                      end: Alignment.bottomRight,
                    )
                  : null,
              border: isSelected
                  ? Border.all(color: AppColors.primary, width: 2)
                  : null,
              borderRadius: BorderRadius.circular(16),
            ),
            child: Padding(
              padding: const EdgeInsets.all(16.0),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  // Верхняя часть: иконка + название
                  Row(
                    children: [
                      Container(
                        width: 44,
                        height: 44,
                        decoration: BoxDecoration(
                          color: riskColor.withOpacity(0.1),
                          shape: BoxShape.circle,
                        ),
                        child: Center(
                          child: Text(
                            icon,
                            style: TextStyle(
                              fontSize: 20,
                              color: riskColor,
                            ),
                          ),
                        ),
                      ),
                      const SizedBox(width: 14),
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              name,
                              style: const TextStyle(
                                fontSize: 17,
                                fontWeight: FontWeight.bold,
                                color: AppColors.textPrimary,
                              ),
                              maxLines: 1,
                              overflow: TextOverflow.ellipsis,
                            ),
                            const SizedBox(height: 3),
                            Text(
                              industry,
                              style: const TextStyle(
                                fontSize: 13,
                                color: AppColors.textSecondary,
                              ),
                              maxLines: 1,
                              overflow: TextOverflow.ellipsis,
                            ),
                          ],
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 14),
                  
                  // Статистика
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      _buildStatItem('Объём', production),
                      _buildStatItem('Экспорт', exportPercent),
                      _buildStatItem('Валюта', currency),
                    ],
                  ),
                  const SizedBox(height: 12),
                  
                  // Индикатор риска
                  Container(
                    padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
                    decoration: BoxDecoration(
                      color: riskColor.withOpacity(0.08),
                      borderRadius: BorderRadius.circular(12),
                      border: Border.all(
                        color: riskColor.withOpacity(0.3),
                        width: 1,
                      ),
                    ),
                    child: Row(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        Container(
                          width: 8,
                          height: 8,
                          decoration: BoxDecoration(
                            color: riskColor,
                            shape: BoxShape.circle,
                          ),
                        ),
                        const SizedBox(width: 8),
                        Text(
                          '$riskLabel риск',
                          style: TextStyle(
                            fontSize: 13,
                            fontWeight: FontWeight.w600,
                            color: riskColor,
                          ),
                        ),
                      ],
                    ),
                  ),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildStatItem(String label, String value) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          label,
          style: const TextStyle(
            fontSize: 11,
            color: AppColors.textSecondary,
            fontWeight: FontWeight.w500,
          ),
        ),
        const SizedBox(height: 4),
        Text(
          value,
          style: const TextStyle(
            fontSize: 14,
            fontWeight: FontWeight.bold,
            color: AppColors.textPrimary,
          ),
          maxLines: 1,
        ),
      ],
    );
  }

  String _getIndustryIcon(String industry) {
    final lower = industry.toLowerCase();
    
    if (lower.contains('калий') || lower.contains('удобрени')) {
      return '🧂';
    }
    if (lower.contains('азот')) {
      return '🧪';
    }
    if (lower.contains('машиностроение') || lower.contains('самосвал') || lower.contains('техника')) {
      return '🚛';
    }
    if (lower.contains('автомобиль') || lower.contains('авто')) {
      return '🚗';
    }
    if (lower.contains('химическ')) {
      return '⚗️';
    }
    return '🏭';
  }
}   
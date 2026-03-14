import 'package:client_flutter/core/constants/app_colors.dart';
import 'package:flutter/material.dart';

enum RiskLevel { low, medium, high }

class RiskCard extends StatelessWidget {
  final String title;
  final RiskLevel riskLevel;
  final num exposure;
  final double varValue;
  final double stressTest;
  final List<String> recommendations;
  final String icon;

  const RiskCard({
    super.key,
    required this.title,
    required this.riskLevel,
    required this.exposure,
    required this.varValue,
    required this.stressTest,
    required this.recommendations,
    required this.icon,
  });

  Color get _riskColor {
    switch (riskLevel) {
      case RiskLevel.low: return AppColors.riskLow;
      case RiskLevel.medium: return AppColors.riskMedium;
      case RiskLevel.high: return AppColors.riskHigh;
    }
  }

  String get _riskText {
    switch (riskLevel) {
      case RiskLevel.low: return 'НИЗКИЙ РИСК';
      case RiskLevel.medium: return 'СРЕДНИЙ РИСК';
      case RiskLevel.high: return 'ВЫСОКИЙ РИСК';
    }
  }

  @override
  Widget build(BuildContext context) {
    return Card(
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
      child: Container(
        decoration: BoxDecoration(
          gradient: LinearGradient(
            begin: Alignment.topLeft,
            end: Alignment.bottomRight,
            colors: [
              Colors.white,
              _riskColor.withOpacity(0.05),
            ],
          ),
          borderRadius: BorderRadius.circular(16),
          border: Border.all(color: _riskColor.withOpacity(0.2), width: 1),
        ),
        child: Padding(
          padding: const EdgeInsets.all(16.0),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              // Заголовок
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Row(
                    children: [
                      Container(
                        width: 40,
                        height: 40,
                        decoration: BoxDecoration(
                          color: _riskColor.withOpacity(0.1),
                          shape: BoxShape.circle,
                        ),
                        child: Center(
                          child: Text(
                            icon,
                            style: TextStyle(
                              fontSize: 20,
                              color: _riskColor,
                              fontWeight: FontWeight.bold,
                            ),
                          ),
                        ),
                      ),
                      const SizedBox(width: 12),
                      Text(
                        title,
                        style: const TextStyle(
                          fontSize: 18,
                          fontWeight: FontWeight.bold,
                          color: AppColors.textPrimary,
                        ),
                      ),
                    ],
                  ),
                  Container(
                    padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
                    decoration: BoxDecoration(
                      color: _riskColor.withOpacity(0.1),
                      borderRadius: BorderRadius.circular(20),
                    ),
                    child: Text(
                      _riskText,
                      style: TextStyle(
                        fontSize: 12,
                        fontWeight: FontWeight.w600,
                        color: _riskColor,
                      ),
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 16),

              // Статистика
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceAround,
                children: [
                  _buildStatItem('Экспозиция', '\$${(exposure / 1000000).toStringAsFixed(1)}M'),
                  const VerticalDivider(width: 1, color: AppColors.divider),
                  _buildStatItem('VaR (95%)', '\$${(varValue / 1000000).toStringAsFixed(1)}M'),
                  const VerticalDivider(width: 1, color: AppColors.divider),
                  _buildStatItem('Стресс', '\$${(stressTest / 1000000).toStringAsFixed(1)}M', isStress: true),
                ],
              ),
              const SizedBox(height: 16),

              // Рекомендации
              if (recommendations.isNotEmpty) ...[
                const Divider(color: AppColors.divider),
                const SizedBox(height: 12),
                Row(
                  children: [
                    Icon(Icons.lightbulb_outlined, size: 16, color: AppColors.textSecondary),
                    const SizedBox(width: 8),
                    const Text(
                      'Рекомендации',
                      style: TextStyle(
                        fontSize: 14,
                        fontWeight: FontWeight.w600,
                        color: AppColors.textPrimary,
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 8),
                ...recommendations.take(2).map((rec) => _buildRecommendation(rec)),
                if (recommendations.length > 2)
                  Padding(
                    padding: const EdgeInsets.only(top: 8),
                    child: Text(
                      '+ ещё ${recommendations.length - 2} рекомендаций',
                      style: const TextStyle(
                        fontSize: 12,
                        color: AppColors.primary,
                      ),
                    ),
                  ),
              ],
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildStatItem(String label, String value, {bool isStress = false}) {
    return Column(
      children: [
        Text(
          label,
          style: const TextStyle(
            fontSize: 12,
            color: AppColors.textSecondary,
          ),
        ),
        const SizedBox(height: 4),
        Text(
          value,
          style: TextStyle(
            fontSize: 16,
            fontWeight: FontWeight.bold,
            color: isStress ? AppColors.riskHigh : AppColors.textPrimary,
          ),
        ),
      ],
    );
  }

  Widget _buildRecommendation(String text) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 6),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Text(
            '•',
            style: TextStyle(
              fontSize: 18,
              color: AppColors.primary,
              height: 0.8,
            ),
          ),
          const SizedBox(width: 8),
          Expanded(
            child: Text(
              text,
              style: const TextStyle(
                fontSize: 13,
                color: AppColors.textPrimary,
              ),
              maxLines: 2,
              overflow: TextOverflow.ellipsis,
            ),
          ),
        ],
      ),
    );
  }
}
import 'dart:math';

import 'package:client_flutter/core/constants/app_colors.dart';
import 'package:flutter/material.dart';
import 'package:fl_chart/fl_chart.dart';

class RiskTrendChart extends StatelessWidget {
  final List<Map<String, dynamic>> historicalData;

  const RiskTrendChart({
    super.key,
    required this.historicalData,
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
                const Icon(Icons.trending_up, color: AppColors.primary, size: 20),
                const SizedBox(width: 8),
                const Text(
                  'Динамика рисков (последние 30 дней)',
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
              height: 350,
              child: LineChart(
                mainData(),
                duration: const Duration(milliseconds: 250),
                //swapAnimationDuration: const Duration(milliseconds: 250),
              ),
            ),
            const SizedBox(height: 12),
            Wrap(
              spacing: 16,
              runSpacing: 8,
              children: [
                _buildIndicator(AppColors.riskHigh, 'Фондовый риск'),
                _buildIndicator(AppColors.riskMedium, 'Валютный риск'),
                _buildIndicator(AppColors.riskLow, 'Процентный риск'),
              ],
            ),
          ],
        ),
      ),
    );
  }

  LineChartData mainData() {
    return LineChartData(
      gridData: FlGridData(
        show: true,
        drawVerticalLine: true,
        horizontalInterval: 1,
        verticalInterval: 50,
        getDrawingHorizontalLine: (value) {
          return FlLine(
            color: AppColors.divider,
            strokeWidth: 1,
          );
        },
        getDrawingVerticalLine: (value) {
          return FlLine(
            color: AppColors.divider,
            strokeWidth: 1,
          );
        },
      ),
      titlesData: FlTitlesData(
        show: true,
        rightTitles: const AxisTitles(
          sideTitles: SideTitles(showTitles: false),
        ),
        topTitles: const AxisTitles(
          sideTitles: SideTitles(showTitles: false),
        ),
        bottomTitles: AxisTitles(
          sideTitles: SideTitles(
            showTitles: true,
            reservedSize: 30,
            interval: 5,
            getTitlesWidget: (value, meta) {
              return Text(
                '${value.toInt()}',
                style: const TextStyle(
                  color: AppColors.textSecondary,
                  fontSize: 11,
                ),
              );
            },
          ),
        ),
        leftTitles: AxisTitles(
          sideTitles: SideTitles(
            showTitles: true,
            interval: 1,
            reservedSize: 40,
            getTitlesWidget: (value, meta) {
              return Text(
                '\$${value.toInt()}M',
                style: const TextStyle(
                  color: AppColors.textSecondary,
                  fontSize: 11,
                ),
              );
            },
          ),
        ),
      ),
      borderData: FlBorderData(
        show: true,
        border: Border.all(color: AppColors.border, width: 1),
      ),
      minX: 0,
      maxX: 30,
      minY: 0,
      maxY: 400,
      lineBarsData: [
        // Фондовый риск (красный)
        LineChartBarData(
          spots: _generateSpots('market'),
          isCurved: true,
          gradient: LinearGradient(
            colors: [AppColors.riskHigh.withOpacity(0.2), AppColors.riskHigh],
          ),
          barWidth: 3,
          isStrokeCapRound: true,
          dotData: FlDotData(
            show: true,
            getDotPainter: (spot, percent, barData, index) => FlDotCirclePainter(
              radius: 5,
              color: AppColors.riskHigh,
              strokeWidth: 2,
              strokeColor: Colors.white,
            ),
          ),
          belowBarData: BarAreaData(
            show: true,
            gradient: LinearGradient(
              colors: [
                AppColors.riskHigh.withOpacity(0.1),
                AppColors.riskHigh.withOpacity(0.02),
              ],
              begin: Alignment.topCenter,
              end: Alignment.bottomCenter,
            ),
          ),
        ),
        // Валютный риск (жёлтый)
        LineChartBarData(
          spots: _generateSpots('currency'),
          isCurved: true,
          gradient: LinearGradient(
            colors: [AppColors.riskMedium.withOpacity(0.2), AppColors.riskMedium],
          ),
          barWidth: 3,
          isStrokeCapRound: true,
          dotData: FlDotData(
            show: true,
            getDotPainter: (spot, percent, barData, index) => FlDotCirclePainter(
              radius: 4,
              color: AppColors.riskMedium,
              strokeWidth: 2,
              strokeColor: Colors.white,
            ),
          ),
          belowBarData: BarAreaData(
            show: true,
            gradient: LinearGradient(
              colors: [
                AppColors.riskMedium.withOpacity(0.1),
                AppColors.riskMedium.withOpacity(0.02),
              ],
              begin: Alignment.topCenter,
              end: Alignment.bottomCenter,
            ),
          ),
        ),
        // Процентный риск (зелёный)
        LineChartBarData(
          spots: _generateSpots('interest'),
          isCurved: true,
          gradient: LinearGradient(
            colors: [AppColors.riskLow.withOpacity(0.2), AppColors.riskLow],
          ),
          barWidth: 3,
          isStrokeCapRound: true,
          dotData: FlDotData(
            show: true,
            getDotPainter: (spot, percent, barData, index) => FlDotCirclePainter(
              radius: 4,
              color: AppColors.riskLow,
              strokeWidth: 2,
              strokeColor: Colors.white,
            ),
          ),
          belowBarData: BarAreaData(
            show: true,
            gradient: LinearGradient(
              colors: [
                AppColors.riskLow.withOpacity(0.1),
                AppColors.riskLow.withOpacity(0.02),
              ],
              begin: Alignment.topCenter,
              end: Alignment.bottomCenter,
            ),
          ),
        ),
      ],
    );
  }

  List<FlSpot> _generateSpots(String riskType) {
    // Генерируем случайные данные для демонстрации
    // В реальном приложении данные будут приходить с сервера
    final random = Random();
    final baseValue = riskType == 'market' ? 300 : riskType == 'currency' ? 100 : 20;
    
    return List.generate(31, (index) {
      final variation = (random.nextDouble() - 0.5) * 50;
      final value = (baseValue + variation).clamp(10.0, 380.0);
      return FlSpot(index.toDouble(), value);
    });
  }

  Widget _buildIndicator(Color color, String text) {
    return Row(
      children: [
        Container(
          width: 12,
          height: 12,
          decoration: BoxDecoration(
            color: color,
            shape: BoxShape.circle,
            border: Border.all(color: Colors.white, width: 2),
          ),
        ),
        const SizedBox(width: 6),
        Text(
          text,
          style: const TextStyle(
            fontSize: 13,
            color: AppColors.textPrimary,
          ),
        ),
      ],
    );
  }
}
import 'package:client_flutter/core/constants/app_colors.dart';
import 'package:flutter/material.dart';

class RiskParametersWidget extends StatefulWidget {
  const RiskParametersWidget({super.key});

  @override
  State<RiskParametersWidget> createState() => _RiskParametersWidgetState();
}

class _RiskParametersWidgetState extends State<RiskParametersWidget> {
  int _horizonDays = 30;
  double _confidenceLevel = 0.95;
  double _priceChangePct = -15.0;
  double _rateChangePct = 1.0;

  Color _getHorizonColor(int days) {
    if (days <= 15) return AppColors.riskLow;
    if (days <= 45) return AppColors.riskMedium;
    return AppColors.riskHigh;
  }

  Color _getConfidenceColor(double level) {
    if (level >= 0.99) return AppColors.riskLow;
    if (level >= 0.95) return AppColors.riskMedium;
    return AppColors.riskHigh;
  }

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
                const Icon(Icons.settings, color: AppColors.primary, size: 20),
                const SizedBox(width: 8),
                const Text(
                  'Параметры расчёта',
                  style: TextStyle(
                    fontSize: 18,
                    fontWeight: FontWeight.bold,
                    color: AppColors.textPrimary,
                  ),
                ),
              ],
            ),
            const SizedBox(height: 16),
            
            // Горизонт расчёта
            _buildParameterSlider(
              label: 'Горизонт расчёта',
              value: _horizonDays.toDouble(),
              min: 1,
              max: 365,
              divisions: 364,
              unit: 'дней',
              color: _getHorizonColor(_horizonDays),
              onChanged: (value) {
                setState(() => _horizonDays = value.toInt());
              },
            ),
            const SizedBox(height: 24),
            
            // Уровень доверия
            _buildParameterSlider(
              label: 'Уровень доверия',
              value: _confidenceLevel * 100,
              min: 90,
              max: 99,
              divisions: 9,
              unit: '%',
              color: _getConfidenceColor(_confidenceLevel),
              onChanged: (value) {
                setState(() => _confidenceLevel = value / 100);
              },
            ),
            const SizedBox(height: 24),
            
            // Изменение цены калия
            _buildParameterSlider(
              label: 'Изменение цены калия',
              value: _priceChangePct,
              min: -30,
              max: 30,
              divisions: 60,
              unit: '%',
              color: Colors.red,
              onChanged: (value) {
                setState(() => _priceChangePct = value);
              },
            ),
            const SizedBox(height: 24),
            
            // Изменение ставки
            _buildParameterSlider(
              label: 'Изменение ставки',
              value: _rateChangePct,
              min: -3,
              max: 5,
              divisions: 16,
              unit: '%',
              color: Colors.amber,
              onChanged: (value) {
                setState(() => _rateChangePct = value);
              },
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildParameterSlider({
    required String label,
    required double value,
    required double min,
    required double max,
    required int divisions,
    required String unit,
    required Color color,
    required ValueChanged<double> onChanged,
  }) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            Text(
              label,
              style: const TextStyle(
                fontSize: 14,
                fontWeight: FontWeight.w600,
                color: AppColors.textPrimary,
              ),
            ),
            Text(
              '${value.toStringAsFixed(value.truncateToDouble() == value ? 0 : 1)}$unit',
              style: TextStyle(
                fontSize: 16,
                fontWeight: FontWeight.bold,
                color: color,
              ),
            ),
          ],
        ),
        const SizedBox(height: 12),
        SliderTheme(
          data: SliderThemeData(
            activeTrackColor: color,
            inactiveTrackColor: AppColors.border,
            thumbColor: color,
            overlayColor: color.withOpacity(0.2),
            thumbShape: const RoundSliderThumbShape(enabledThumbRadius: 10),
            overlayShape: const RoundSliderOverlayShape(overlayRadius: 25),
          ),
          child: Slider(
            value: value,
            min: min,
            max: max,
            divisions: divisions,
            onChanged: onChanged,
          ),
        ),
      ],
    );
  }
}
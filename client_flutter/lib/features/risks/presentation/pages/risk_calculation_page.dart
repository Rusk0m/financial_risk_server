import 'package:client_flutter/core/constants/app_colors.dart';
import 'package:client_flutter/di/service_locator.dart';
import 'package:client_flutter/features/risks/presentation/bloc/risk_calculation_bloc.dart';
import 'package:client_flutter/features/risks/presentation/widgets/risk_parameters_widget.dart';
import 'package:client_flutter/presentation/widgets/app_bar.dart';
import 'package:client_flutter/features/risks/presentation/widgets/risk_results_widget.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

class RiskCalculationPage extends StatelessWidget {
  const RiskCalculationPage({super.key});

  @override
  Widget build(BuildContext context) {
    return BlocProvider(
      create: (context) => sl<RiskCalculationBloc>(),
      child: Scaffold(
        appBar: const CustomAppBar(
          title: '🧮 Расчёт финансовых рисков',
        ),
        body: const RiskCalculationContent(),
      ),
    );
  }
}

class RiskCalculationContent extends StatelessWidget {
  const RiskCalculationContent({super.key});

  @override
  Widget build(BuildContext context) {
    return SingleChildScrollView(
      child: Padding(
        padding: const EdgeInsets.all(16.0),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Приветствие
            _buildHeader(),
            const SizedBox(height: 24),
            
            // Параметры расчёта
            const RiskParametersWidget(),
            const SizedBox(height: 24),
            
            // Кнопка расчёта
            _buildCalculateButton(context),
            const SizedBox(height: 24),
            
            // Результаты
            BlocBuilder<RiskCalculationBloc, RiskCalculationState>(
              builder: (context, state) {
                return switch (state.status) {
                  RiskCalculationStatus.initial => const SizedBox.shrink(),
                  RiskCalculationStatus.loading => _buildLoading(),
                  RiskCalculationStatus.loaded => RiskResultsWidget(results: state.results!),
                  RiskCalculationStatus.error => _buildError(context,state.error!),
                };
              },
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildHeader() {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Container(
          padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
          decoration: BoxDecoration(
            color: AppColors.primaryLight,
            borderRadius: BorderRadius.circular(20),
          ),
          child: const Text(
            'Анализ рисков ОАО «Беларуськалий»',
            style: TextStyle(
              fontSize: 14,
              fontWeight: FontWeight.w600,
              color: AppColors.primaryDark,
            ),
          ),
        ),
        const SizedBox(height: 16),
        const Text(
          'Расчёт финансовых рисков',
          style: TextStyle(
            fontSize: 28,
            fontWeight: FontWeight.bold,
            color: AppColors.textPrimary,
          ),
        ),
        const SizedBox(height: 8),
        const Text(
          'Произведите комплексный анализ всех типов финансовых рисков для предприятия ОАО «Беларуськалий»',
          style: TextStyle(
            fontSize: 16,
            color: AppColors.textSecondary,
          ),
        ),
      ],
    );
  }

  Widget _buildCalculateButton(BuildContext context) {
    return BlocBuilder<RiskCalculationBloc, RiskCalculationState>(
      builder: (context, state) {
        return SizedBox(
          width: double.infinity,
          child: ElevatedButton(
            onPressed: state.status == RiskCalculationStatus.loading
                ? null
                : () => context.read<RiskCalculationBloc>().add(
                      CalculateRiskRequested(
                        enterpriseId: 1,
                        horizonDays: 30,
                        confidenceLevel: 0.95,
                        priceChangePct: -15.0,
                        rateChangePct: 1.0,
                      ),
                    ),
            style: ElevatedButton.styleFrom(
              padding: const EdgeInsets.symmetric(vertical: 16),
              backgroundColor: AppColors.primary,
              foregroundColor: Colors.white,
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(16),
              ),
              elevation: 3,
            ),
            child: state.status == RiskCalculationStatus.loading
                ? const SizedBox(
                    width: 24,
                    height: 24,
                    child: CircularProgressIndicator(
                      strokeWidth: 2,
                      valueColor: AlwaysStoppedAnimation<Color>(Colors.white),
                    ),
                  )
                : const Text(
                    'РАССЧИТАТЬ ВСЕ РИСКИ',
                    style: TextStyle(
                      fontSize: 18,
                      fontWeight: FontWeight.w700,
                    ),
                  ),
          ),
        );
      },
    );
  }

  Widget _buildLoading() {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          const SizedBox(height: 40),
          Container(
            width: 80,
            height: 80,
            decoration: BoxDecoration(
              color: AppColors.primaryLight,
              borderRadius: BorderRadius.circular(20),
            ),
            child: const Center(
              child: SizedBox(
                width: 40,
                height: 40,
                child: CircularProgressIndicator(
                  strokeWidth: 3,
                  valueColor: AlwaysStoppedAnimation<Color>(AppColors.primary),
                ),
              ),
            ),
          ),
          const SizedBox(height: 24),
          const Text(
            'Расчёт рисков...',
            style: TextStyle(
              fontSize: 20,
              fontWeight: FontWeight.bold,
              color: AppColors.textPrimary,
            ),
          ),
          const SizedBox(height: 8),
          const Text(
            'Анализируем данные предприятия и рыночные условия',
            style: TextStyle(
              fontSize: 16,
              color: AppColors.textSecondary,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildError(BuildContext context, String error) {
    return Card(
      color: Colors.red.withOpacity(0.05),
      child: Padding(
        padding: const EdgeInsets.all(16.0),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Icon(Icons.error_outline, color: Colors.red, size: 24),
                const SizedBox(width: 12),
                const Text(
                  'Ошибка расчёта',
                  style: TextStyle(
                    fontSize: 18,
                    fontWeight: FontWeight.bold,
                    color: Colors.red,
                  ),
                ),
              ],
            ),
            const SizedBox(height: 12),
            Text(
              error,
              style: const TextStyle(
                fontSize: 14,
                color: Colors.red,
              ),
            ),
            const SizedBox(height: 16),
            Center(
              child: ElevatedButton(
                onPressed: () {
                  context.read<RiskCalculationBloc>().add(
                     CalculateRiskRequested(
                      enterpriseId: 1,
                      horizonDays: 30,
                      confidenceLevel: 0.95,
                      priceChangePct: -15.0,
                      rateChangePct: 1.0,
                    ),
                  );
                },
                child: const Text('Повторить расчёт'),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
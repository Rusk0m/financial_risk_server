import 'package:flutter/material.dart';
import 'package:client_flutter/core/constants/app_colors.dart';
import 'package:client_flutter/presentation/widgets/app_bar.dart';

class EnterpriseDashboard extends StatelessWidget {
  const EnterpriseDashboard({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: const CustomAppBar(title: 'ОАО «Беларуськалий»'),
      body: SingleChildScrollView(
        child: Padding(
          padding: const EdgeInsets.all(16.0),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              // Шапка предприятия
              _buildEnterpriseHeader(),
              const SizedBox(height: 24),

              // Аналитика рисков
              _buildRiskAnalysisSection(),
              const SizedBox(height: 24),

              // Статистика отчётов
              _buildReportsSection(),
              const SizedBox(height: 24),

              // Последние расчёты
              _buildRecentCalculations(),
              const SizedBox(height: 24),

              // Быстрые действия
              _buildQuickActions(context),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildEnterpriseHeader() {
    return Container(
      decoration: BoxDecoration(
        gradient: const LinearGradient(
          colors: [Color(0xFF0284C7), Color(0xFF0369A1)],
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
        ),
        borderRadius: BorderRadius.circular(20),
      ),
      child: Padding(
        padding: const EdgeInsets.all(24.0),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Логотип и название
            Row(
              children: [
                Container(
                  width: 64,
                  height: 64,
                  decoration: BoxDecoration(
                    color: Colors.white.withOpacity(0.2),
                    shape: BoxShape.circle,
                  ),
                  child: const Center(
                    child: Text('🇧🇾', style: TextStyle(fontSize: 28)),
                  ),
                ),
                const SizedBox(width: 20),
                Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    const Text(
                      'ОАО «Беларуськалий»',
                      style: TextStyle(
                        fontSize: 32,
                        fontWeight: FontWeight.bold,
                        color: Colors.white,
                      ),
                    ),
                    const SizedBox(height: 4),
                    const Text(
                      'Крупнейший в мире производитель калийных удобрений',
                      style: TextStyle(fontSize: 16, color: Colors.white70),
                    ),
                  ],
                ),
              ],
            ),
            const SizedBox(height: 24),

            // Ключевые показатели
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceAround,
              children: [
                _buildKeyMetric(
                  'Годовой объём',
                  '8.5 млн т',
                  'Добыча калийных руд',
                  Icons.factory,
                ),
                _buildKeyMetric(
                  'Экспорт',
                  '95%',
                  'Доля выручки в иностранной валюте',
                  Icons.ios_share,
                ),
                _buildKeyMetric(
                  'Рыночная доля',
                  '18%',
                  'На мировом рынке калия',
                  Icons.pie_chart,
                ),
              ],
            ),
            const SizedBox(height: 24),

            // Статус системы
            Container(
              padding: const EdgeInsets.all(16),
              decoration: BoxDecoration(
                color: Colors.white.withOpacity(0.15),
                borderRadius: BorderRadius.circular(16),
              ),
              child: Row(
                children: [
                  Container(
                    width: 12,
                    height: 12,
                    decoration: BoxDecoration(
                      color: Colors.green,
                      shape: BoxShape.circle,
                    ),
                  ),
                  const SizedBox(width: 12),
                  const Text(
                    'Система анализа рисков активна • Последнее обновление данных: 04.03.2026 14:30',
                    style: TextStyle(
                      fontSize: 14,
                      color: Colors.white,
                      height: 1.4,
                    ),
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildKeyMetric(
    String label,
    String value,
    String description,
    IconData icon,
  ) {
    return Column(
      children: [
        Container(
          width: 48,
          height: 48,
          decoration: BoxDecoration(
            color: Colors.white.withOpacity(0.2),
            shape: BoxShape.circle,
          ),
          child: Center(child: Icon(icon, size: 24, color: Colors.white)),
        ),
        const SizedBox(height: 12),
        Text(
          value,
          style: const TextStyle(
            fontSize: 24,
            fontWeight: FontWeight.bold,
            color: Colors.white,
          ),
        ),
        const SizedBox(height: 4),
        Text(
          label,
          style: const TextStyle(fontSize: 14, color: Colors.white70),
        ),
        const SizedBox(height: 8),
        Container(width: 80, height: 2, color: Colors.white.withOpacity(0.3)),
        const SizedBox(height: 8),
        Text(
          description,
          style: const TextStyle(
            fontSize: 12,
            color: Colors.white54,
            height: 1.3,
          ),
          textAlign: TextAlign.center,
        ),
      ],
    );
  }

  Widget _buildRiskAnalysisSection() {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        _buildSectionHeader(
          '📈 Анализ финансовых рисков',
          'Текущий уровень рисков на основе последних отчётов',
        ),
        const SizedBox(height: 16),
        Row(
          children: [
            Expanded(
              child: _buildRiskCard(
                'Валютный',
                'СРЕДНИЙ',
                '\$12.87 млн',
                '💱',
                AppColors.riskMedium,
                'Экспорт в USD при затратах в BYN',
              ),
            ),
            const SizedBox(width: 12),
            Expanded(
              child: _buildRiskCard(
                'Процентный',
                'НИЗКИЙ',
                '\$5.00 млн',
                '📈',
                AppColors.riskLow,
                'Долг с фиксированной ставкой',
              ),
            ),
            const SizedBox(width: 12),
            Expanded(
              child: _buildRiskCard(
                'Ликвидности',
                'СРЕДНИЙ',
                '\$57.00 млн',
                '💧',
                AppColors.riskMedium,
                'Краткосрочные обязательства',
              ),
            ),
          ],
        ),
        const SizedBox(height: 16),
        _buildRiskSummary(),
      ],
    );
  }

  Widget _buildRiskCard(
    String title,
    String level,
    String value,
    String icon,
    Color color,
    String description,
  ) {
    return Card(
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
      child: Container(
        decoration: BoxDecoration(
          gradient: LinearGradient(
            colors: [color.withOpacity(0.05), Colors.white],
            begin: Alignment.topLeft,
            end: Alignment.bottomRight,
          ),
          borderRadius: BorderRadius.circular(16),
          border: Border.all(color: color.withOpacity(0.3), width: 1),
        ),
        child: Padding(
          padding: const EdgeInsets.all(16.0),
          child: Column(
            children: [
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Text(icon, style: const TextStyle(fontSize: 32)),
                  Container(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 10,
                      vertical: 4,
                    ),
                    decoration: BoxDecoration(
                      color: color.withOpacity(0.15),
                      borderRadius: BorderRadius.circular(20),
                    ),
                    child: Text(
                      level,
                      style: TextStyle(
                        fontSize: 12,
                        fontWeight: FontWeight.w700,
                        color: color,
                      ),
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 12),
              Text(
                title,
                style: const TextStyle(
                  fontSize: 18,
                  fontWeight: FontWeight.bold,
                  color: AppColors.textPrimary,
                ),
                textAlign: TextAlign.center,
              ),
              const SizedBox(height: 8),
              Text(
                value,
                style: TextStyle(
                  fontSize: 24,
                  fontWeight: FontWeight.bold,
                  color: color,
                ),
                textAlign: TextAlign.center,
              ),
              const SizedBox(height: 12),
              Container(
                height: 4,
                decoration: BoxDecoration(
                  color: color.withOpacity(0.2),
                  borderRadius: BorderRadius.circular(2),
                ),
                child: FractionallySizedBox(
                  widthFactor: _getRiskWidth(level),
                  child: Container(
                    decoration: BoxDecoration(
                      color: color,
                      borderRadius: BorderRadius.circular(2),
                    ),
                  ),
                ),
              ),
              const SizedBox(height: 8),
              Text(
                description,
                style: const TextStyle(
                  fontSize: 12,
                  color: AppColors.textSecondary,
                  height: 1.4,
                ),
                textAlign: TextAlign.center,
              ),
            ],
          ),
        ),
      ),
    );
  }

  double _getRiskWidth(String level) {
    switch (level) {
      case 'НИЗКИЙ':
        return 0.3;
      case 'СРЕДНИЙ':
        return 0.6;
      case 'ВЫСОКИЙ':
        return 0.9;
      default:
        return 0.5;
    }
  }

  Widget _buildRiskSummary() {
    return Card(
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
      child: Padding(
        padding: const EdgeInsets.all(20.0),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Container(
                  width: 8,
                  height: 32,
                  decoration: BoxDecoration(
                    color: AppColors.primary,
                    borderRadius: BorderRadius.circular(4),
                  ),
                ),
                const SizedBox(width: 12),
                const Text(
                  'Рекомендации по управлению рисками',
                  style: TextStyle(
                    fontSize: 18,
                    fontWeight: FontWeight.bold,
                    color: AppColors.textPrimary,
                  ),
                ),
              ],
            ),
            const SizedBox(height: 16),
            _buildRecommendation(
              '1. Валютный риск',
              'Хеджировать 30-40% валютной экспозиции через форвардные контракты с банками РБ. Сократить срок оплаты экспортных контрактов до 45-60 дней.',
              Icons.swap_horiz,
              AppColors.riskMedium,
            ),
            _buildRecommendation(
              '2. Процентный риск',
              'Поддерживать текущую структуру капитала. Мониторинг процентных ставок 1 раз в месяц. Рассмотреть рефинансирование части долга при снижении ставок.',
              Icons.trending_up,
              AppColors.riskLow,
            ),
            _buildRecommendation(
              '3. Риск ликвидности',
              'Усилить контроль за дебиторской задолженностью (еженедельный анализ). Переговорить с ключевыми поставщиками о более выгодных условиях оплаты.',
              Icons.water_drop,
              AppColors.riskMedium,
            ),
            const SizedBox(height: 20),
            Center(
              child: ElevatedButton.icon(
                onPressed: () {
                  // Переход к странице расчёта рисков
                },
                icon: const Icon(Icons.calculate),
                label: const Text(
                  'РАССЧИТАТЬ РИСКИ',
                  style: TextStyle(fontWeight: FontWeight.bold),
                ),
                style: ElevatedButton.styleFrom(
                  padding: const EdgeInsets.symmetric(
                    horizontal: 24,
                    vertical: 14,
                  ),
                  backgroundColor: AppColors.primary,
                  foregroundColor: Colors.white,
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(14),
                  ),
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildRecommendation(
    String title,
    String description,
    IconData icon,
    Color color,
  ) {
    return Container(
      margin: const EdgeInsets.only(bottom: 16),
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        border: Border(left: BorderSide(width: 4, color: color)),
        color: color.withOpacity(0.03),
        borderRadius: BorderRadius.circular(12),
      ),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Container(
            width: 36,
            height: 36,
            decoration: BoxDecoration(
              color: color.withOpacity(0.1),
              shape: BoxShape.circle,
            ),
            child: Center(child: Icon(icon, size: 20, color: color)),
          ),
          const SizedBox(width: 14),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  title,
                  style: TextStyle(
                    fontSize: 15,
                    fontWeight: FontWeight.w700,
                    color: color,
                  ),
                ),
                const SizedBox(height: 6),
                Text(
                  description,
                  style: const TextStyle(
                    fontSize: 14,
                    color: AppColors.textPrimary,
                    height: 1.5,
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildReportsSection() {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        _buildSectionHeader(
          '📊 Управление отчётами',
          'Загруженные бухгалтерией документы для расчёта рисков',
        ),
        const SizedBox(height: 16),
        Row(
          children: [
            Expanded(
              child: _buildReportCard(
                'Экспортные контракты',
                '127',
                '02.03.2026',
                Icons.description,
                Colors.blue,
              ),
            ),
            const SizedBox(width: 12),
            Expanded(
              child: _buildReportCard(
                'Финансовые балансы',
                '4',
                '01.01.2026',
                Icons.account_balance,
                Colors.green,
              ),
            ),
          ],
        ),
        const SizedBox(height: 12),
        Row(
          children: [
            Expanded(
              child: _buildReportCard(
                'Отчёты о ФР',
                '12',
                '01.02.2026',
                Icons.show_chart,
                Colors.purple,
              ),
            ),
            const SizedBox(width: 12),
            Expanded(
              child: _buildReportCard(
                'Кредитные договоры',
                '8',
                '15.12.2025',
                Icons.credit_card,
                Colors.red,
              ),
            ),
          ],
        ),
        const SizedBox(height: 20),
        Center(
          child: Wrap(
            spacing: 12,
            runSpacing: 12,
            children: [
              _buildUploadButton(
                'Загрузить контракты',
                Icons.file_upload,
                Colors.blue,
              ),
              _buildUploadButton(
                'Загрузить баланс',
                Icons.file_upload,
                Colors.green,
              ),
              _buildUploadButton(
                'Загрузить отчёт ФР',
                Icons.file_upload,
                Colors.purple,
              ),
              _buildUploadButton(
                'Загрузить кредиты',
                Icons.file_upload,
                Colors.red,
              ),
            ],
          ),
        ),
      ],
    );
  }

  Widget _buildReportCard(
    String title,
    String count,
    String date,
    IconData icon,
    Color color,
  ) {
    return Card(
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
      child: Padding(
        padding: const EdgeInsets.all(16.0),
        child: Column(
          children: [
            Container(
              width: 48,
              height: 48,
              decoration: BoxDecoration(
                color: color.withOpacity(0.1),
                shape: BoxShape.circle,
              ),
              child: Center(child: Icon(icon, size: 24, color: color)),
            ),
            const SizedBox(height: 12),
            Text(
              title,
              style: const TextStyle(
                fontSize: 15,
                fontWeight: FontWeight.w600,
                color: AppColors.textPrimary,
              ),
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: 8),
            Text(
              count,
              style: TextStyle(
                fontSize: 28,
                fontWeight: FontWeight.bold,
                color: color,
              ),
            ),
            const SizedBox(height: 4),
            Text(
              'последний: $date',
              style: const TextStyle(
                fontSize: 12,
                color: AppColors.textSecondary,
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildUploadButton(String label, IconData icon, Color color) {
    return ElevatedButton.icon(
      onPressed: () {
        // Открыть диалог загрузки файла
      },
      icon: Icon(icon, size: 18),
      label: Text(
        label,
        style: const TextStyle(fontSize: 13, fontWeight: FontWeight.w600),
      ),
      style: ElevatedButton.styleFrom(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
        backgroundColor: color.withOpacity(0.1),
        foregroundColor: color,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
        elevation: 0,
      ),
    );
  }

  Widget _buildRecentCalculations() {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        _buildSectionHeader(
          '📅 История расчётов',
          'Последние 5 расчётов финансовых рисков',
        ),
        const SizedBox(height: 16),
        _buildCalculationItem(
          '04.03.2026 14:30',
          'Валютный риск',
          'СРЕДНИЙ',
          '\$12.87 млн',
          AppColors.riskMedium,
        ),
        _buildCalculationItem(
          '04.03.2026 14:30',
          'Процентный риск',
          'НИЗКИЙ',
          '\$5.00 млн',
          AppColors.riskLow,
        ),
        _buildCalculationItem(
          '04.03.2026 14:30',
          'Риск ликвидности',
          'СРЕДНИЙ',
          '\$57.00 млн',
          AppColors.riskMedium,
        ),
        _buildCalculationItem(
          '01.03.2026 09:15',
          'Валютный риск',
          'СРЕДНИЙ',
          '\$13.21 млн',
          AppColors.riskMedium,
        ),
        _buildCalculationItem(
          '01.03.2026 09:15',
          'Процентный риск',
          'НИЗКИЙ',
          '\$4.95 млн',
          AppColors.riskLow,
        ),
        const SizedBox(height: 16),
        Center(
          child: TextButton(
            onPressed: () {
              // Переход к странице истории расчётов
            },
            child: const Text(
              'Показать всю историю расчётов →',
              style: TextStyle(
                color: AppColors.primary,
                fontWeight: FontWeight.w600,
              ),
            ),
          ),
        ),
      ],
    );
  }

  Widget _buildCalculationItem(
    String dateTime,
    String riskType,
    String level,
    String value,
    Color color,
  ) {
    return Card(
      margin: const EdgeInsets.only(bottom: 8),
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
      child: ListTile(
        leading: Container(
          width: 40,
          height: 40,
          decoration: BoxDecoration(
            color: color.withOpacity(0.1),
            shape: BoxShape.circle,
          ),
          child: Center(
            child: Text(
              _getRiskIcon(riskType),
              style: TextStyle(fontSize: 18, color: color),
            ),
          ),
        ),
        title: Text(
          '$riskType • $level',
          style: const TextStyle(fontWeight: FontWeight.w600),
        ),
        subtitle: Text(
          'Рассчитано: $dateTime',
          style: const TextStyle(fontSize: 13, color: AppColors.textSecondary),
        ),
        trailing: Text(
          value,
          style: TextStyle(
            fontSize: 16,
            fontWeight: FontWeight.bold,
            color: color,
          ),
        ),
      ),
    );
  }

  String _getRiskIcon(String riskType) {
    switch (riskType) {
      case 'Валютный риск':
        return '💱';
      case 'Процентный риск':
        return '📈';
      case 'Риск ликвидности':
        return '💧';
      default:
        return '⚠️';
    }
  }

  Widget _buildQuickActions(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        _buildSectionHeader(
          '🚀 Быстрые действия',
          'Оперативное управление рисками предприятия',
        ),
        const SizedBox(height: 16),
        SizedBox(
          height: 120,
          child: ListView(
            scrollDirection: Axis.horizontal,
            children: [
              _buildActionCard(
                context,
                'Рассчитать риски',
                'Полный анализ 3 ключевых рисков',
                Icons.calculate,
                AppColors.primary,
                () => Navigator.pushNamed(context, '/risks/calculate'),
              ),
              _buildActionCard(
                context,
                'Загрузить отчёты',
                'Импорт данных из бухгалтерии',
                Icons.upload_file,
                Colors.green,
                () => Navigator.pushNamed(context, '/reports'),
              ),
              _buildActionCard(
                context,
                'История расчётов',
                'Все проведённые анализы',
                Icons.history,
                Colors.purple,
                () {},
              ),
              _buildActionCard(
                context,
                'Экспорт в PDF',
                'Сформировать отчёт для руководства',
                Icons.picture_as_pdf,
                Colors.red,
                () {},
              ),
            ],
          ),
        ),
      ],
    );
  }

  Widget _buildActionCard(
    BuildContext context,
    String title,
    String description,
    IconData icon,
    Color color,
    VoidCallback onPressed,
  ) {
    return GestureDetector(
      onTap: onPressed,
      child: IntrinsicHeight(
        child: Container(
          width: 220,
          margin: const EdgeInsets.only(right: 16),
          decoration: BoxDecoration(
            gradient: LinearGradient(
              colors: [color.withOpacity(0.1), Colors.white],
              begin: Alignment.topLeft,
              end: Alignment.bottomRight,
            ),
            borderRadius: BorderRadius.circular(16),
            border: Border.all(color: color.withOpacity(0.3), width: 1),
          ),
          child: Padding(
            padding: const EdgeInsets.all(12.0),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                Center(
                  child: Container(
                    width: 36,
                    height: 36,
                    decoration: BoxDecoration(
                      color: color.withOpacity(0.15),
                      shape: BoxShape.circle,
                    ),
                    child: Center(child: Icon(icon, size: 20, color: color)),
                  ),
                ),
                const SizedBox(height: 10),
                Text(
                  title,
                  style: TextStyle(
                    fontSize: 16,
                    fontWeight: FontWeight.bold,
                    color: color,
                  ),
                ),
                const SizedBox(height: 4),
                Flexible(
                  fit: FlexFit.loose,
                  child: Text(
                    description,
                    style: const TextStyle(
                      fontSize: 13,
                      color: AppColors.textSecondary,
                      height: 1.4,
                    ),
                    overflow: TextOverflow.ellipsis,
                    maxLines: 3,
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildSectionHeader(String title, String subtitle) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          title,
          style: const TextStyle(
            fontSize: 24,
            fontWeight: FontWeight.bold,
            color: AppColors.textPrimary,
          ),
        ),
        const SizedBox(height: 6),
        Text(
          subtitle,
          style: const TextStyle(fontSize: 15, color: AppColors.textSecondary),
        ),
      ],
    );
  }
}

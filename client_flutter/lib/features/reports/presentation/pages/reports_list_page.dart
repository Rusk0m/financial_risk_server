import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:client_flutter/features/reports/presentation/bloc/report_bloc.dart';
import 'package:client_flutter/features/reports/domain/entities/report.dart';
import 'package:client_flutter/features/reports/presentation/widgets/report_card.dart';

class ReportsListPage extends StatefulWidget {
  const ReportsListPage({super.key});

  @override
  State<ReportsListPage> createState() => _ReportsListPageState();
}

class _ReportsListPageState extends State<ReportsListPage> {
  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (mounted) context.read<ReportBloc>().add(LoadReportsEvent());
    });
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('📊 Отчёты'),
        actions: [
          IconButton(
            icon: const Icon(Icons.refresh),
            onPressed: () => context.read<ReportBloc>().add(LoadReportsEvent()),
          ),
        ],
      ),
      body: BlocBuilder<ReportBloc, ReportState>(
        builder: (context, state) {
          if (state.status == ReportStatus.loading && state.reports.isEmpty) {
            return const Center(child: CircularProgressIndicator());
          }

          if (state.status == ReportStatus.error) {
            return Center(
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Icon(Icons.error_outline, size: 48, color: Colors.red[300]),
                  const SizedBox(height: 16),
                  Text(state.error ?? 'Произошла ошибка', textAlign: TextAlign.center),
                  const SizedBox(height: 16),
                  ElevatedButton.icon(
                    onPressed: () => context.read<ReportBloc>().add(LoadReportsEvent()),
                    icon: const Icon(Icons.refresh),
                    label: const Text('Повторить'),
                  ),
                ],
              ),
            );
          }

          if (state.reports.isEmpty) {
            return _buildEmptyState(context);
          }

          final reports = _filterReports(state.reports, state.filterType, state.filterStatus);

          return Column(
            children: [
              _buildFilterBar(context, state),
              Expanded(
                child: ListView.builder(
                  padding: const EdgeInsets.all(16),
                  itemCount: reports.length,
                  itemBuilder: (context, index) {
                    final report = reports[index];
                    return ReportCard(
                      report: report,
                      onTap: () => _showDetails(context, report),
                      onDownload: () => context.read<ReportBloc>().add(DownloadReportEvent(report.id, report.fileName)),
                      onDelete: () => _confirmDelete(context, report.id),
                    );
                  },
                ),
              ),
            ],
          );
        },
      ),
      floatingActionButton: FloatingActionButton.extended(
        onPressed: () => Navigator.pushNamed(context, '/reports/upload')
            .then((updated) {
          if (updated == true && mounted) {
            context.read<ReportBloc>().add(LoadReportsEvent());
          }
        }),
        icon: const Icon(Icons.cloud_upload),
        label: const Text('Загрузить'),
      ),
    );
  }

  Widget _buildFilterBar(BuildContext context, ReportState state) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      color: Colors.grey[50],
      child: Row(
        children: [
          Expanded(
            child: DropdownButtonHideUnderline(
              child: DropdownButton<String>(
                value: state.filterType ?? 'all',
                isExpanded: true,
                items: const [
                  DropdownMenuItem(value: 'all', child: Text('Все типы')),
                  DropdownMenuItem(value: 'export_contracts', child: Text('📄 Экспорт')),
                  DropdownMenuItem(value: 'balance_sheet', child: Text('📊 Баланс')),
                  DropdownMenuItem(value: 'financial_results', child: Text('📈 ФР')),
                  DropdownMenuItem(value: 'credit_agreements', child: Text('💳 Кредиты')),
                ],
                onChanged: (value) => context.read<ReportBloc>().add(
                  FilterReportsEvent(type: value == 'all' ? null : value, status: state.filterStatus),
                ),
              ),
            ),
          ),
          const SizedBox(width: 8),
          Expanded(
            child: DropdownButtonHideUnderline(
              child: DropdownButton<String>(
                value: state.filterStatus ?? 'all',
                isExpanded: true,
                items: const [
                  DropdownMenuItem(value: 'all', child: Text('Все статусы')),
                  DropdownMenuItem(value: 'processed', child: Text('✅ Обработан')),
                  DropdownMenuItem(value: 'pending', child: Text('⏳ В обработке')),
                  DropdownMenuItem(value: 'failed', child: Text('❌ Ошибка')),
                ],
                onChanged: (value) => context.read<ReportBloc>().add(
                  FilterReportsEvent(type: state.filterType, status: value == 'all' ? null : value),
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildEmptyState(BuildContext context) {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(Icons.folder_open, size: 64, color: Theme.of(context).primaryColor.withOpacity(0.5)),
          const SizedBox(height: 16),
          const Text('Нет отчётов', style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold)),
          const SizedBox(height: 8),
          Text('Загрузите первый отчёт, чтобы начать работу', style: TextStyle(color: Colors.grey[600])),
          const SizedBox(height: 24),
          ElevatedButton.icon(
            onPressed: () => Navigator.pushNamed(context, '/reports/upload'),
            icon: const Icon(Icons.cloud_upload),
            label: const Text('Загрузить отчёт'),
          ),
        ],
      ),
    );
  }

  List<Report> _filterReports(List<Report> reports, String? type, String? status) {
    return reports.where((r) {
      if (type != null && r.type != type) return false;
      if (status != null && r.status != status) return false;
      return true;
    }).toList();
  }

  void _showDetails(BuildContext context, Report report) {
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      builder: (context) => DraggableScrollableSheet(
        initialChildSize: 0.7,
        minChildSize: 0.4,
        maxChildSize: 0.9,
        builder: (_, controller) => _buildDetailsSheet(report),
      ),
    );
  }

  Widget _buildDetailsSheet(Report report) {
    return Container(
      decoration: const BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.vertical(top: Radius.circular(20)),
      ),
      child: Column(
        children: [
          Container(
            padding: const EdgeInsets.all(20),
            decoration: BoxDecoration(
              color: _getColor(report.type).withOpacity(0.1),
              borderRadius: const BorderRadius.vertical(top: Radius.circular(20)),
            ),
            child: Row(
              children: [
                Container(
                  width: 56,
                  height: 56,
                  decoration: BoxDecoration(
                    color: _getColor(report.type).withOpacity(0.2),
                    borderRadius: BorderRadius.circular(14),
                  ),
                  child: Center(child: Icon(report.typeIcon, size: 28, color: _getColor(report.type))),
                ),
                const SizedBox(width: 16),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(report.originalName, style: const TextStyle(fontSize: 18, fontWeight: FontWeight.bold), maxLines: 2, overflow: TextOverflow.ellipsis),
                      const SizedBox(height: 4),
                      Text(report.typeDisplayName, style: TextStyle(fontSize: 14, color: Colors.grey[600])),
                    ],
                  ),
                ),
                Container(
                  padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 8),
                  decoration: BoxDecoration(
                    color: report.statusColor.withOpacity(0.15),
                    borderRadius: BorderRadius.circular(20),
                  ),
                  child: Text(report.statusBadge, style: TextStyle(fontSize: 13, fontWeight: FontWeight.w600, color: report.statusColor)),
                ),
              ],
            ),
          ),
          Expanded(
            child: ListView(
              padding: const EdgeInsets.all(20),
              children: [
                _buildDetailRow('ID', '#${report.id}'),
                _buildDetailRow('Файл', report.fileName),
                _buildDetailRow('Размер', report.formattedSize),
                _buildDetailRow('Загружен', report.formattedDate),
                if (report.uploadedBy != null) _buildDetailRow('Загрузил', report.uploadedBy!),
                if (report.period != null) _buildDetailRow('Период', report.period!),
                if (report.description != null) ...[
                  const SizedBox(height: 16),
                  const Text('Описание', style: TextStyle(fontSize: 16, fontWeight: FontWeight.w600)),
                  const SizedBox(height: 8),
                  Text(report.description!, style: TextStyle(color: Colors.grey[700], height: 1.5)),
                ],
              ],
            ),
          ),
          Container(
            padding: const EdgeInsets.all(20),
            decoration: BoxDecoration(color: Colors.grey[50], borderRadius: const BorderRadius.vertical(bottom: Radius.circular(20))),
            child: Row(
              children: [
                Expanded(
                  child: OutlinedButton.icon(
                    onPressed: () => context.read<ReportBloc>().add(DownloadReportEvent(report.id, report.fileName)),
                    icon: const Icon(Icons.download),
                    label: const Text('Скачать'),
                  ),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: ElevatedButton.icon(
                    onPressed: () {
                      // TODO: Переход к расчёту рисков
                    },
                    icon: const Icon(Icons.analytics),
                    label: const Text('Анализ'),
                    style: ElevatedButton.styleFrom(backgroundColor: Theme.of(context).primaryColor, foregroundColor: Colors.white),
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildDetailRow(String label, String value) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 12),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          SizedBox(width: 100, child: Text(label, style: TextStyle(fontSize: 14, color: Colors.grey[600], fontWeight: FontWeight.w500))),
          Expanded(child: Text(value, style: const TextStyle(fontSize: 14, fontWeight: FontWeight.w600))),
        ],
      ),
    );
  }

  void _confirmDelete(BuildContext context, int id) {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Удалить отчёт?'),
        content: const Text('Это действие нельзя отменить.'),
        actions: [
          TextButton(onPressed: () => Navigator.pop(context), child: const Text('Отмена')),
          TextButton(
            onPressed: () {
              Navigator.pop(context);
              context.read<ReportBloc>().add(DeleteReportEvent(id));
            },
            style: TextButton.styleFrom(foregroundColor: Colors.red),
            child: const Text('Удалить'),
          ),
        ],
      ),
    );
  }

  Color _getColor(String type) {
    switch (type) {
      case 'export_contracts': return Colors.blue;
      case 'balance_sheet': return Colors.green;
      case 'financial_results': return Colors.purple;
      case 'credit_agreements': return Colors.red;
      default: return Colors.grey;
    }
  }
}
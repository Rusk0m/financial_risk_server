import 'dart:io';
import 'package:client_flutter/features/reports/domain/entities/report.dart';
import 'package:client_flutter/features/reports/domain/repositories/report_repository.dart';

class UploadReport {
  final ReportRepository repository;
  UploadReport(this.repository);

  Future<Report> call({
    required File file,
    required String type,
    String? description,
    DateTime? periodStart,
    DateTime? periodEnd,
  }) async {
    // Валидация типа
    const validTypes = ['export_contracts', 'balance_sheet', 'financial_results', 'credit_agreements'];
    if (!validTypes.contains(type)) {
      throw ArgumentError('Неверный тип отчёта: $type');
    }

    // Валидация файла
    if (!file.existsSync()) throw FileSystemException('Файл не найден', file.path);
    
    final ext = file.path.split('.').last.toLowerCase();
    if (!['xlsx', 'xls'].contains(ext)) {
      throw ArgumentError('Поддерживаются только файлы Excel (.xlsx, .xls)');
    }
    
    if (file.lengthSync() > 10 * 1024 * 1024) {
      throw ArgumentError('Максимальный размер файла: 10 МБ');
    }

    return await repository.upload(
      file: file,
      type: type,
      description: description,
      periodStart: periodStart,
      periodEnd: periodEnd,
    );
  }
}
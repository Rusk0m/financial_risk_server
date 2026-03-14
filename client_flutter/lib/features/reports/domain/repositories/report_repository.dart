import 'dart:io';
import 'package:client_flutter/features/reports/domain/entities/report.dart';

abstract class ReportRepository {
  /// Получить список отчётов с фильтрацией
  Future<List<Report>> getAll({String? type, String? status});
  
  /// Загрузить новый отчёт
  Future<Report> upload({
    required File file,
    required String type,
    String? description,
    DateTime? periodStart,
    DateTime? periodEnd,
  });
  
  /// Скачать отчёт (возвращает ссылку или открывает в браузере)
  Future<String> download(int id);
  
  /// Удалить отчёт
  Future<void> delete(int id);
}
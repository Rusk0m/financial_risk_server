import 'package:equatable/equatable.dart';
import 'package:flutter/material.dart';

class Report extends Equatable {
  final int id;
  final String type;              // export_contracts | balance_sheet | financial_results | credit_agreements
  final String fileName;          // Уникальное имя на сервере
  final String originalName;      // Оригинальное имя файла
  final DateTime uploadDate;
  final String status;            // pending | processed | failed
  final int sizeBytes;
  final String? description;
  final String? period;           // "01.02.2026 — 28.02.2026"
  final String? uploadedBy;
  final String downloadUrl;

  const Report({
    required this.id,
    required this.type,
    required this.fileName,
    required this.originalName,
    required this.uploadDate,
    required this.status,
    required this.sizeBytes,
    this.description,
    this.period,
    this.uploadedBy,
    required this.downloadUrl,
  });

  // === UI Helpers ===
  String get typeDisplayName {
    switch (type) {
      case 'export_contracts': return 'Экспортные контракты';
      case 'balance_sheet': return 'Финансовый баланс';
      case 'financial_results': return 'Отчёт о финансовых результатах';
      case 'credit_agreements': return 'Кредитные договоры';
      default: return 'Неизвестный тип';
    }
  }

  String get statusBadge {
    switch (status) {
      case 'processed': return 'Обработан';
      case 'pending': return 'В обработке';
      case 'failed': return 'Ошибка';
      default: return status;
    }
  }

  Color get statusColor {
    switch (status) {
      case 'processed': return Colors.green;
      case 'pending': return Colors.orange;
      case 'failed': return Colors.red;
      default: return Colors.grey;
    }
  }

  String get formattedSize {
    if (sizeBytes < 1024) return '$sizeBytes Б';
    if (sizeBytes < 1024 * 1024) return '${(sizeBytes / 1024).toStringAsFixed(1)} КБ';
    return '${(sizeBytes / (1024 * 1024)).toStringAsFixed(1)} МБ';
  }

  String get formattedDate {
    return '${uploadDate.day.toString().padLeft(2, '0')}.${uploadDate.month.toString().padLeft(2, '0')}.${uploadDate.year}';
  }

  IconData get typeIcon {
    switch (type) {
      case 'export_contracts': return Icons.description;
      case 'balance_sheet': return Icons.account_balance;
      case 'financial_results': return Icons.show_chart;
      case 'credit_agreements': return Icons.credit_card;
      default: return Icons.folder;
    }
  }

  @override
  List<Object?> get props => [id, type, fileName, status];
}
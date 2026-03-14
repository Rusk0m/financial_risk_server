import 'package:flutter/material.dart';
import 'package:client_flutter/features/reports/domain/entities/report.dart';

class ReportCard extends StatelessWidget {
  final Report report;
  final VoidCallback? onDownload;
  final VoidCallback? onDelete;
  final VoidCallback? onTap;

  const ReportCard({
    Key? key,
    required this.report,
    this.onDownload,
    this.onDelete,
    this.onTap,
  }) : super(key: key);

  @override
  Widget build(BuildContext context) {
    return Card(
      margin: const EdgeInsets.only(bottom: 12),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(12),
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                children: [
                  Container(
                    width: 48,
                    height: 48,
                    decoration: BoxDecoration(
                      color: _getColor(report.type).withOpacity(0.15),
                      borderRadius: BorderRadius.circular(12),
                    ),
                    child: Icon(report.typeIcon, color: _getColor(report.type)),
                  ),
                  const SizedBox(width: 16),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          report.originalName,
                          style: const TextStyle(fontWeight: FontWeight.w600, fontSize: 16),
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                        ),
                        const SizedBox(height: 4),
                        Text(report.typeDisplayName, style: TextStyle(color: Colors.grey[600], fontSize: 13)),
                        if (report.period != null) ...[
                          const SizedBox(height: 2),
                          Text('📅 ${report.period}', style: TextStyle(fontSize: 12, color: Colors.grey[500])),
                        ],
                      ],
                    ),
                  ),
                  Chip(
                    label: Text(report.statusBadge, style: TextStyle(fontSize: 12, color: report.statusColor)),
                    backgroundColor: report.statusColor.withOpacity(0.15),
                    padding: EdgeInsets.zero,
                    materialTapTargetSize: MaterialTapTargetSize.shrinkWrap,
                  ),
                ],
              ),
              const Divider(height: 24),
              Row(
                children: [
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text('📤 ${report.formattedDate}', style: TextStyle(fontSize: 12, color: Colors.grey[600])),
                        const SizedBox(height: 2),
                        Text('💾 ${report.formattedSize}', style: TextStyle(fontSize: 12, color: Colors.grey[600])),
                        if (report.uploadedBy != null)
                          Text('👤 ${report.uploadedBy}', style: TextStyle(fontSize: 12, color: Colors.grey[600])),
                      ],
                    ),
                  ),
                  Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      IconButton(
                        icon: const Icon(Icons.download, size: 20),
                        onPressed: onDownload,
                        tooltip: 'Скачать',
                        constraints: const BoxConstraints(),
                        padding: EdgeInsets.zero,
                      ),
                      IconButton(
                        icon: const Icon(Icons.delete_outline, size: 20),
                        onPressed: onDelete,
                        tooltip: 'Удалить',
                        constraints: const BoxConstraints(),
                        padding: EdgeInsets.zero,
                        color: Colors.red[400],
                      ),
                    ],
                  ),
                ],
              ),
            ],
          ),
        ),
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
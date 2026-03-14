import 'dart:io';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:file_picker/file_picker.dart';
import 'package:client_flutter/features/reports/presentation/bloc/report_bloc.dart';

class ReportUploadPage extends StatefulWidget {
  const ReportUploadPage({super.key});

  @override
  State<ReportUploadPage> createState() => _ReportUploadPageState();
}

class _ReportUploadPageState extends State<ReportUploadPage> {
  String _selectedType = 'export_contracts';
  File? _selectedFile;
  String? _description;

  final _reportTypes = const {
    'export_contracts': _ReportTypeInfo('Экспортные контракты', 'Данные по экспортным контрактам', Icons.description, Colors.blue),
    'balance_sheet': _ReportTypeInfo('Финансовый баланс', 'Форма №1 по ГОСТ 10520-2013', Icons.account_balance, Colors.green),
    'financial_results': _ReportTypeInfo('Отчёт о финансовых результатах', 'Форма №2 по ГОСТ 10520-2013', Icons.show_chart, Colors.purple),
    'credit_agreements': _ReportTypeInfo('Кредитные договоры', 'Данные по кредитам и займам', Icons.credit_card, Colors.red),
  };

  @override
  Widget build(BuildContext context) {
    final info = _reportTypes[_selectedType]!;

    return Scaffold(
      appBar: AppBar(title: const Text('📤 Загрузка отчёта')),
      body: BlocListener<ReportBloc, ReportState>(
        listenWhen: (p, c) => p.uploadStatus != c.uploadStatus,
        listener: (context, state) {
          if (state.uploadStatus == UploadStatus.success) {
            ScaffoldMessenger.of(context).showSnackBar(
              const SnackBar(content: Text('✅ Отчёт успешно загружен'), backgroundColor: Colors.green),
            );
            Future.delayed(const Duration(seconds: 1), () {
              if (mounted) Navigator.pop(context, true);
            });
          } else if (state.uploadStatus == UploadStatus.error) {
            ScaffoldMessenger.of(context).showSnackBar(
              SnackBar(content: Text('❌ ${state.uploadError ?? 'Ошибка загрузки'}'), backgroundColor: Colors.red),
            );
          }
        },
        child: SingleChildScrollView(
          padding: const EdgeInsets.all(16),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              _buildTypeSelector(),
              const SizedBox(height: 20),
              _buildInfoCard(info),
              const SizedBox(height: 20),
              _buildFilePicker(),
              if (_selectedFile != null) ...[
                const SizedBox(height: 16),
                _buildDescriptionField(),
              ],
              if (state.uploadStatus == UploadStatus.uploading) ...[
                const SizedBox(height: 20),
                LinearProgressIndicator(value: state.uploadProgress),
                const SizedBox(height: 8),
                Center(child: Text('Загрузка: ${(state.uploadProgress * 100).toStringAsFixed(0)}%')),
              ],
            ],
          ),
        ),
      ),
    );
  }

  ReportState get state => context.read<ReportBloc>().state;

  Widget _buildTypeSelector() {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        const Text('Тип отчёта', style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold)),
        const SizedBox(height: 12),
        Wrap(
          spacing: 8,
          runSpacing: 8,
          children: _reportTypes.entries.map((e) {
            final isSelected = e.key == _selectedType;
            return FilterChip(
              label: Text(e.value.name),
              selected: isSelected,
              onSelected: (selected) {
                if (selected) setState(() => _selectedType = e.key);
              },
              avatar: Icon(e.value.icon, size: 20, color: isSelected ? e.value.color : null),
              selectedColor: e.value.color.withOpacity(0.2),
              checkmarkColor: e.value.color,
            );
          }).toList(),
        ),
      ],
    );
  }

  Widget _buildInfoCard(_ReportTypeInfo info) {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Row(
          children: [
            Container(
              width: 48, height: 48,
              decoration: BoxDecoration(color: info.color.withOpacity(0.15), borderRadius: BorderRadius.circular(12)),
              child: Center(child: Icon(info.icon, color: info.color)),
            ),
            const SizedBox(width: 16),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(info.name, style: const TextStyle(fontSize: 16, fontWeight: FontWeight.bold)),
                  const SizedBox(height: 4),
                  Text(info.description, style: TextStyle(color: Colors.grey[600], fontSize: 14)),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildFilePicker() {
    return GestureDetector(
      onTap: _pickFile,
      child: Container(
        width: double.infinity,
        padding: const EdgeInsets.all(32),
        decoration: BoxDecoration(
          color: Colors.grey[50],
          borderRadius: BorderRadius.circular(16),
          border: Border.all(color: Colors.grey[300]!),
        ),
        child: Column(
          children: [
            Icon(Icons.upload_file, size: 48, color: Theme.of(context).primaryColor),
            const SizedBox(height: 16),
            Text(
              _selectedFile?.path.split('/').last ?? 'Нажмите для выбора файла',
              style: const TextStyle(fontSize: 16, fontWeight: FontWeight.w600),
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: 8),
            Text('Формат: XLSX, макс. 10 МБ', style: TextStyle(color: Colors.grey[600])),
          ],
        ),
      ),
    );
  }

  Widget _buildDescriptionField() {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        const Text('Описание (необязательно)', style: TextStyle(fontWeight: FontWeight.w600)),
        const SizedBox(height: 8),
        TextFormField(
          initialValue: _description,
          decoration: const InputDecoration(
            hintText: 'Например: Отчёт за февраль 2026',
            border: OutlineInputBorder(),
            contentPadding: EdgeInsets.symmetric(horizontal: 12, vertical: 10),
          ),
          maxLines: 2,
          onChanged: (v) => _description = v,
        ),
        const SizedBox(height: 16),
        SizedBox(
          width: double.infinity,
          child: ElevatedButton.icon(
            onPressed: _selectedFile == null || state.uploadStatus == UploadStatus.uploading ? null : _upload,
            icon: state.uploadStatus == UploadStatus.uploading
                ? const SizedBox(width: 20, height: 20, child: CircularProgressIndicator(strokeWidth: 2))
                : const Icon(Icons.cloud_upload),
            label: Text(state.uploadStatus == UploadStatus.uploading ? 'Загрузка...' : 'ЗАГРУЗИТЬ'),
          ),
        ),
      ],
    );
  }

  Future<void> _pickFile() async {
    final result = await FilePicker.platform.pickFiles(
      type: FileType.custom,
      allowedExtensions: ['xlsx', 'xls'],
    );
    if (result != null && result.files.single.path != null) {
      final file = File(result.files.single.path!);
      if (file.lengthSync() > 10 * 1024 * 1024) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Файл слишком большой (макс. 10 МБ)'), backgroundColor: Colors.red),
        );
        return;
      }
      setState(() {
        _selectedFile = file;
      });
    }
  }

  void _upload() {
    if (_selectedFile == null) return;
    context.read<ReportBloc>().add(
      UploadReportEvent(
        file: _selectedFile!,
        type: _selectedType,
        description: _description?.isNotEmpty == true ? _description : null,
      ),
    );
  }
}

class _ReportTypeInfo {
  final String name;
  final String description;
  final IconData icon;
  final Color color;
  const _ReportTypeInfo(this.name, this.description, this.icon, this.color);
}
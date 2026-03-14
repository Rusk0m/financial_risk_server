import 'package:flutter/material.dart';
import 'package:client_flutter/core/constants/app_colors.dart';
import 'package:client_flutter/features/enterprices/domain/entities/enterprise.dart';

class EnterpriseForm extends StatefulWidget {
  final Enterprise? enterprise;
  final void Function(Enterprise) onSubmit;

  const EnterpriseForm({
    super.key,
    this.enterprise,
    required this.onSubmit,
  });

  @override
  State<EnterpriseForm> createState() => _EnterpriseFormState();
}

class _EnterpriseFormState extends State<EnterpriseForm> {
  final _formKey = GlobalKey<FormState>();
  late String _name;
  late String _industry;
  late String _annualProduction;
  late String _exportShare;
  late String _currency;

  @override
  void initState() {
    super.initState();
    _name = widget.enterprise?.name ?? '';
    _industry = widget.enterprise?.industry ?? 'Горнодобывающая (калийные удобрения)';
    _annualProduction = widget.enterprise?.annualProductionT.toString() ?? '8500000';
    _exportShare = widget.enterprise?.exportSharePercent.toString() ?? '95';
    _currency = widget.enterprise?.mainCurrency ?? 'USD';
  }

  @override
  Widget build(BuildContext context) {
    return Form(
      key: _formKey,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          _buildTextField(
            label: 'Название *',
            controller: TextEditingController(text: _name),
            onChanged: (v) => _name = v,
            validator: (v) => v == null || v.trim().isEmpty ? 'Введите название' : null,
            hint: 'ОАО «Беларуськалий»',
          ),
          const SizedBox(height: 16),
          
          _buildTextField(
            label: 'Отрасль *',
            controller: TextEditingController(text: _industry),
            onChanged: (v) => _industry = v,
            validator: (v) => v == null || v.trim().isEmpty ? 'Введите отрасль' : null,
            hint: 'Горнодобывающая',
          ),
          const SizedBox(height: 16),
          
          _buildRowFields(
            left: _buildTextField(
              label: 'Объём (т) *',
              controller: TextEditingController(text: _annualProduction),
              onChanged: (v) => _annualProduction = v,
              validator: (v) {
                if (v == null || v.trim().isEmpty) return 'Введите объём';
                final num = double.tryParse(v.trim());
                return (num != null && num > 0) ? null : 'Введите корректный объём';
              },
              hint: '8500000',
              keyboardType: TextInputType.number,
            ),
            right: _buildTextField(
              label: 'Экспорт (%) *',
              controller: TextEditingController(text: _exportShare),
              onChanged: (v) => _exportShare = v,
              validator: (v) {
                if (v == null || v.trim().isEmpty) return 'Введите долю экспорта';
                final num = double.tryParse(v.trim());
                return (num != null && num >= 0 && num <= 100) 
                    ? null 
                    : 'Доля должна быть от 0 до 100%';
              },
              hint: '95',
              keyboardType: TextInputType.number,
            ),
          ),
          const SizedBox(height: 16),
          
          _buildDropdown(
            label: 'Валюта *',
            items: const ['USD', 'EUR', 'BYN', 'CNY'],
            value: _currency,
            onChanged: (v) => setState(() => _currency = v!),
          ),
          const SizedBox(height: 8),
        ],
      ),
    );
  }

  Widget _buildTextField({
    required String label,
    required TextEditingController controller,
    required ValueChanged<String> onChanged,
    required String? Function(String?)? validator,
    String? hint,
    TextInputType keyboardType = TextInputType.text,
  }) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          label,
          style: const TextStyle(
            fontSize: 14,
            fontWeight: FontWeight.w600,
            color: AppColors.textPrimary,
          ),
        ),
        const SizedBox(height: 8),
        TextFormField(
          controller: controller,
          onChanged: onChanged,
          validator: validator,
          keyboardType: keyboardType,
          decoration: InputDecoration(
            hintText: hint,
            hintStyle: const TextStyle(color: AppColors.textSecondary),
            filled: true,
            fillColor: Colors.grey[50],
            border: OutlineInputBorder(
              borderRadius: BorderRadius.circular(12),
              borderSide: const BorderSide(color: AppColors.border),
            ),
            enabledBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(12),
              borderSide: const BorderSide(color: AppColors.border),
            ),
            focusedBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(12),
              borderSide: const BorderSide(color: AppColors.primary, width: 2),
            ),
          ),
        ),
      ],
    );
  }

  Widget _buildRowFields({required Widget left, required Widget right}) {
    return Row(
      children: [
        Expanded(child: left),
        const SizedBox(width: 12),
        Expanded(child: right),
      ],
    );
  }

  Widget _buildDropdown({
    required String label,
    required List<String> items,
    required String value,
    required ValueChanged<String?> onChanged,
  }) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          label,
          style: const TextStyle(
            fontSize: 14,
            fontWeight: FontWeight.w600,
            color: AppColors.textPrimary,
          ),
        ),
        const SizedBox(height: 8),
        Container(
          decoration: BoxDecoration(
            color: Colors.grey[50],
            borderRadius: BorderRadius.circular(12),
            border: Border.all(color: AppColors.border),
          ),
          child: DropdownButtonFormField<String>(
            value: value,
            items: items.map((item) => DropdownMenuItem(value: item, child: Text(item))).toList(),
            onChanged: onChanged,
            decoration: const InputDecoration(
              border: InputBorder.none,
              contentPadding: EdgeInsets.symmetric(horizontal: 16, vertical: 14),
            ),
            dropdownColor: Colors.white,
            borderRadius: BorderRadius.circular(12),
          ),
        ),
      ],
    );
  }

  void _submitForm() {
    if (_formKey.currentState!.validate()) {
      final enterprise = Enterprise(
        id: widget.enterprise?.id ?? 0,
        name: _name.trim(),
        industry: _industry.trim(),
        annualProductionT: double.tryParse(_annualProduction.trim()) ?? 8500000,
        exportSharePercent: double.tryParse(_exportShare.trim()) ?? 95,
        mainCurrency: _currency,
      );
      
      widget.onSubmit(enterprise);
    }
  }
}
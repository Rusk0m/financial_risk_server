import 'package:client_flutter/features/enterprices/presentation/widgets/enterprice_card.dart';
import 'package:client_flutter/features/enterprices/presentation/widgets/enterprice_form_dialog.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:client_flutter/core/constants/app_colors.dart';
import 'package:client_flutter/features/enterprices/domain/entities/enterprise.dart';
import 'package:client_flutter/features/enterprices/domain/usecases/create_enterprise.dart';
import 'package:client_flutter/features/enterprices/domain/usecases/delete_enterprise.dart';
import 'package:client_flutter/features/enterprices/domain/usecases/get_all_enterprises.dart';
import 'package:client_flutter/features/enterprices/domain/usecases/get_by_id_enterprise.dart';
import 'package:client_flutter/features/enterprices/domain/usecases/update_enterprise.dart';
import 'package:client_flutter/features/enterprices/presentation/bloc/enterprise_bloc.dart';
import 'package:client_flutter/presentation/widgets/app_bar.dart';
import 'package:client_flutter/di/service_locator.dart';

class EnterprisesPage extends StatelessWidget {
  const EnterprisesPage({super.key});

  @override
  Widget build(BuildContext context) {
    return BlocProvider(
      create: (context) => EnterpriseBloc(
        getAllEnterprises: sl<GetAllEnterprises>(),
        getByIdEnterprise: sl<GetByIdEnterprise>(),
        createEnterprise: sl<CreateEnterprise>(),
        updateEnterprise: sl<UpdateEnterprise>(),
        deleteEnterprise: sl<DeleteEnterprise>(),
      )..add(LoadEnterprises()),
      child: const _EnterprisesView(),
    );
  }
}

class _EnterprisesView extends StatefulWidget {
  const _EnterprisesView();

  @override
  State<_EnterprisesView> createState() => _EnterprisesViewState();
}

class _EnterprisesViewState extends State<_EnterprisesView> {
  bool _showError = false;
  String _errorMessage = '';

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: const CustomAppBar(
        title: '🏢 Предприятия',
      ),
      body: BlocConsumer<EnterpriseBloc, EnterpriseState>(
        listener: (context, state) {
          // Обработка ошибок
          if (state.error != null && !_showError) {
            setState(() {
              _showError = true;
              _errorMessage = state.error!;
            });
            
            // Автоматическое скрытие через 5 секунд
            Future.delayed(const Duration(seconds: 5), () {
              if (mounted) {
                setState(() => _showError = false);
              }
            });
          }
        },
        builder: (context, state) {
          // Отображение ошибки поверх контента
          if (_showError) {
            WidgetsBinding.instance.addPostFrameCallback((_) {
              _showErrorSnackBar(_errorMessage);
            });
            setState(() => _showError = false);
          }

          return _buildContent(state);
        },
      ),
      floatingActionButton: FloatingActionButton(
        onPressed: _showCreateDialog,
        backgroundColor: AppColors.primary,
        child: const Icon(Icons.add, color: Colors.white),
      ),
    );
  }

  Widget _buildContent(EnterpriseState state) {
    switch (state.status) {
      case EnterpriseStatus.initial:
      case EnterpriseStatus.loading:
        return _buildLoading();

      case EnterpriseStatus.loaded:
      case EnterpriseStatus.success:
        return _buildEnterpriseList(state.enterprises ?? [],state);

      case EnterpriseStatus.error:
        return _buildError(state.error ?? 'Неизвестная ошибка');

      case EnterpriseStatus.creating:
        return _buildLoading(message: 'Создание предприятия...');

      case EnterpriseStatus.updating:
        return _buildLoading(message: 'Обновление предприятия...');

      case EnterpriseStatus.deleting:
        return _buildLoading(message: 'Удаление предприятия...');
    }
  }

  Widget _buildLoading({String message = 'Загрузка данных...'}) {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          const SizedBox(
            width: 50,
            height: 50,
            child: CircularProgressIndicator(
              strokeWidth: 3,
              valueColor: AlwaysStoppedAnimation<Color>(AppColors.primary),
            ),
          ),
          const SizedBox(height: 20),
          Text(
            message,
            style: const TextStyle(
              fontSize: 17,
              fontWeight: FontWeight.w500,
              color: AppColors.textSecondary,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildError(String message) {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(
            Icons.error_outline,
            size: 72,
            color: Colors.red.withOpacity(0.6),
          ),
          const SizedBox(height: 20),
          Text(
            'Ошибка загрузки',
            style: const TextStyle(
              fontSize: 22,
              fontWeight: FontWeight.bold,
              color: Colors.red,
            ),
          ),
          const SizedBox(height: 10),
          SizedBox(
            width: 300,
            child: Text(
              message,
              style: const TextStyle(
                fontSize: 15,
                color: AppColors.textSecondary,
                height: 1.4,
              ),
              textAlign: TextAlign.center,
            ),
          ),
          const SizedBox(height: 30),
          ElevatedButton.icon(
            onPressed: () => context.read<EnterpriseBloc>().add(LoadEnterprises()),
            icon: const Icon(Icons.refresh),
            label: const Text('Повторить попытку'),
            style: ElevatedButton.styleFrom(
              padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 14),
              backgroundColor: AppColors.primary,
              foregroundColor: Colors.white,
              shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(14)),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildEnterpriseList(List<Enterprise> enterprises, EnterpriseState state) {
    if (enterprises.isEmpty) {
      return _buildEmptyState();
    }

    return RefreshIndicator(
      onRefresh: () async {
        context.read<EnterpriseBloc>().add(LoadEnterprises());
      },
      child: ListView.separated(
        padding: const EdgeInsets.all(16.0),
        itemCount: enterprises.length,
        separatorBuilder: (context, index) => const SizedBox(height: 12),
        itemBuilder: (context, index) {
          final enterprise = enterprises[index];
          final isSelected = state.selectedEnterprise?.id == enterprise.id;
          
          return EnterpriseCard(
            enterprise: enterprise,
            isSelected: isSelected,
            onTap: () => _showDetails(context, enterprise),
          );
        },
      ),
    );
  }

  Widget _buildEmptyState() {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Container(
            width: 90,
            height: 90,
            decoration: BoxDecoration(
              gradient: const LinearGradient(
                colors: [Color(0xFF3B82F6), Color(0xFF1D4ED8)],
              ),
              shape: BoxShape.circle,
            ),
            child: const Center(
              child: Icon(Icons.business, size: 44, color: Colors.white),
            ),
          ),
          const SizedBox(height: 28),
          const Text(
            'Нет предприятий',
            style: TextStyle(
              fontSize: 24,
              fontWeight: FontWeight.bold,
              color: AppColors.textPrimary,
            ),
          ),
          const SizedBox(height: 12),
          const Text(
            'Добавьте предприятие для анализа финансовых рисков',
            style: TextStyle(
              fontSize: 16,
              color: AppColors.textSecondary,
              height: 1.5,
              //textAlign: TextAlign.center,
            ),
            textAlign: TextAlign.center,
          ),
          const SizedBox(height: 36),
          SizedBox(
            width: 280,
            child: ElevatedButton.icon(
              onPressed: _showCreateDialog,
              icon: const Icon(Icons.add, size: 20),
              label: const Text(
                'Добавить предприятие',
                style: TextStyle(fontSize: 16, fontWeight: FontWeight.w600),
              ),
              style: ElevatedButton.styleFrom(
                padding: const EdgeInsets.symmetric(vertical: 16),
                backgroundColor: AppColors.primary,
                foregroundColor: Colors.white,
                shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
                elevation: 3,
              ),
            ),
          ),
        ],
      ),
    );
  }

  void _showCreateDialog() {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text(
          'Добавить предприятие',
          style: TextStyle(fontWeight: FontWeight.bold, fontSize: 20),
        ),
        content: SizedBox(
          width: double.maxFinite,
          child: SingleChildScrollView(
            child: ConstrainedBox(
              constraints: const BoxConstraints(minHeight: 400),
              child: EnterpriseForm(
                onSubmit: (enterprise) {
                  Navigator.pop(context);
                  context.read<EnterpriseBloc>().add(CreateEnterpriseEvent(enterprise));
                },
              ),
            ),
          ),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('Отмена'),
          ),
        ],
      ),
    );
  }

  void _showDetails(BuildContext context, Enterprise enterprise) {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: Text(
          enterprise.name,
          style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 20),
        ),
        content: SizedBox(
          width: double.maxFinite,
          child: SingleChildScrollView(
            child: ConstrainedBox(
              constraints: const BoxConstraints(minHeight: 300),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  _buildDetailRow('Отрасль', enterprise.industry),
                  _buildDetailRow('Годовой объём', '${(enterprise.annualProductionT / 1000000).toStringAsFixed(1)} млн т'),
                  _buildDetailRow('Доля экспорта', '${enterprise.exportSharePercent.toStringAsFixed(0)}%'),
                  _buildDetailRow('Основная валюта', enterprise.mainCurrency),
                  const SizedBox(height: 20),
                  const Text(
                    'Экспортный риск',
                    style: TextStyle(fontWeight: FontWeight.bold, fontSize: 16),
                  ),
                  const SizedBox(height: 10),
                  _buildRiskBadge(enterprise.exportSharePercent),
                ],
              ),
            ),
          ),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('Закрыть'),
          ),
        ],
      ),
    );
  }

  Widget _buildDetailRow(String label, String value) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 8),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          SizedBox(
            width: 130,
            child: Text(
              '$label:',
              style: const TextStyle(fontWeight: FontWeight.w500),
            ),
          ),
          Expanded(
            child: Text(
              value,
              style: const TextStyle(fontWeight: FontWeight.normal),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildRiskBadge(double exportShare) {
    Color color;
    String label;
    
    if (exportShare > 90) {
      color = AppColors.riskHigh;
      label = 'Очень высокий';
    } else if (exportShare > 75) {
      color = Colors.red;
      label = 'Высокий';
    } else if (exportShare > 50) {
      color = AppColors.riskMedium;
      label = 'Средний';
    } else {
      color = AppColors.riskLow;
      label = 'Низкий';
    }
    
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 8),
      decoration: BoxDecoration(
        color: color.withOpacity(0.1),
        borderRadius: BorderRadius.circular(14),
        border: Border.all(color: color.withOpacity(0.3), width: 1.5),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Container(
            width: 10,
            height: 10,
            decoration: BoxDecoration(
              color: color,
              shape: BoxShape.circle,
            ),
          ),
          const SizedBox(width: 10),
          Text(
            '$label экспортный риск',
            style: TextStyle(
              fontSize: 15,
              fontWeight: FontWeight.w600,
              color: color,
            ),
          ),
        ],
      ),
    );
  }

  void _showErrorSnackBar(String message) {
    if (!mounted) return;
    
    ScaffoldMessenger.of(context).hideCurrentSnackBar();
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Row(
          children: [
            const Icon(Icons.error, size: 20, color: Colors.white),
            const SizedBox(width: 12),
            Expanded(
              child: Text(
                message,
                style: const TextStyle(color: Colors.white, fontSize: 14),
              ),
            ),
          ],
        ),
        backgroundColor: Colors.red,
        duration: const Duration(seconds: 4),
        behavior: SnackBarBehavior.floating,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
        margin: const EdgeInsets.all(16),
      ),
    );
  }
}
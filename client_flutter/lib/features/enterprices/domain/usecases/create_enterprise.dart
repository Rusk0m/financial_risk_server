
import 'package:client_flutter/features/enterprices/domain/entities/enterprise.dart';
import 'package:client_flutter/features/enterprices/domain/repositories/enterprise_repository.dart';

class CreateEnterprise {
  final EnterpriseRepository repository;

  CreateEnterprise(this.repository);

  Future<Enterprise> call(Enterprise enterprise) async {
    // Валидация данных перед отправкой
    if (enterprise.name.isEmpty) {
      throw Exception('Название предприятия не может быть пустым');
    }
    if (enterprise.annualProductionT <= 0) {
      throw Exception('Годовой объём производства должен быть положительным');
    }
    if (enterprise.exportSharePercent < 0 || enterprise.exportSharePercent > 100) {
      throw Exception('Доля экспорта должна быть в диапазоне 0-100%');
    }
    
    return await repository.create(enterprise);
  }
}
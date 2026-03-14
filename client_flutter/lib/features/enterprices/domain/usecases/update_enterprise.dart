
import 'package:client_flutter/features/enterprices/domain/entities/enterprise.dart';
import 'package:client_flutter/features/enterprices/domain/repositories/enterprise_repository.dart';

class UpdateEnterprise {
  final EnterpriseRepository repository;

  UpdateEnterprise(this.repository);

  Future<Enterprise> call(int id, Enterprise enterprise) async {
    if (id <= 0) {
      throw Exception('Некорректный идентификатор предприятия');
    }
    
    // Валидация данных
    if (enterprise.name.isEmpty) {
      throw Exception('Название предприятия не может быть пустым');
    }
    
    return await repository.update(id, enterprise);
  }
}
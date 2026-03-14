import 'package:client_flutter/features/enterprices/domain/repositories/enterprise_repository.dart';

class DeleteEnterprise {
  final EnterpriseRepository repository;

  DeleteEnterprise(this.repository);

  Future<void> call(int id) async {
    if (id <= 0) {
      throw Exception('Некорректный идентификатор предприятия');
    }
    
    return await repository.delete(id);
  }
}
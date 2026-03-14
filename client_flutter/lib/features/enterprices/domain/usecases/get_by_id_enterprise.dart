
import 'package:client_flutter/features/enterprices/domain/entities/enterprise.dart';
import 'package:client_flutter/features/enterprices/domain/repositories/enterprise_repository.dart';

class GetByIdEnterprise {
  final EnterpriseRepository repository;

  GetByIdEnterprise(this.repository);

  Future<Enterprise> call(int id) async {
    if (id <= 0) {
      throw Exception('Некорректный идентификатор предприятия');
    }
    return await repository.getById(id);
  }
}
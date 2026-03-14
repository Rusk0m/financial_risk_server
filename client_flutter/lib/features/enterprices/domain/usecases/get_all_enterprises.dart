import 'package:client_flutter/features/enterprices/domain/entities/enterprise.dart';
import 'package:client_flutter/features/enterprices/domain/repositories/enterprise_repository.dart';

class GetAllEnterprises {
  final EnterpriseRepository repository;

  GetAllEnterprises(this.repository);

  Future<List<Enterprise>> call() async {
    return await repository.getAll();
  }
}
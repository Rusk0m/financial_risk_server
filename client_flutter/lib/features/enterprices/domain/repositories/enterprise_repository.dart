import 'package:client_flutter/features/enterprices/domain/entities/enterprise.dart';

abstract class EnterpriseRepository {
  Future<List<Enterprise>> getAll();
  Future<Enterprise> getById(int id);
  Future<Enterprise> create(Enterprise enterprise);
  Future<Enterprise> update(int id, Enterprise enterprise);
  Future<void> delete(int id);
}

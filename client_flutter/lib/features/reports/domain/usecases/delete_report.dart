import 'package:client_flutter/features/reports/domain/repositories/report_repository.dart';

class DeleteReport {
  final ReportRepository repository;
  DeleteReport(this.repository);

  Future<void> call(int id) async => await repository.delete(id);
}

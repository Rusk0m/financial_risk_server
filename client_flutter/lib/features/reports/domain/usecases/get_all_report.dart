import 'package:client_flutter/features/reports/domain/entities/report.dart';
import 'package:client_flutter/features/reports/domain/repositories/report_repository.dart';

class GetAllReports {
  final ReportRepository repository;
  GetAllReports(this.repository);

  Future<List<Report>> call({String? type, String? status}) async {
    return await repository.getAll(type: type, status: status);
  }
}
import 'package:client_flutter/features/reports/domain/repositories/report_repository.dart';

class DownloadReport {
  final ReportRepository repository;
  DownloadReport(this.repository);

  Future<String> call(int id) async => await repository.download(id);
}

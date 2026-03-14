part of 'report_bloc.dart';

@immutable
sealed class ReportEvent {}

final class LoadReportsEvent extends ReportEvent {
  final String? type;
  final String? status;
  LoadReportsEvent({this.type, this.status});
}

final class UploadReportEvent extends ReportEvent {
  final File file;
  final String type;
  final String? description;
  final DateTime? periodStart;
  final DateTime? periodEnd;
  
  UploadReportEvent({
    required this.file,
    required this.type,
    this.description,
    this.periodStart,
    this.periodEnd,
  });
}

final class DeleteReportEvent extends ReportEvent {
  final int id;
  DeleteReportEvent(this.id);
}

final class DownloadReportEvent extends ReportEvent {
  final int id;
  final String fileName;
  DownloadReportEvent(this.id, this.fileName);
}

final class FilterReportsEvent extends ReportEvent {
  final String? type;
  final String? status;
  FilterReportsEvent({this.type, this.status});
}
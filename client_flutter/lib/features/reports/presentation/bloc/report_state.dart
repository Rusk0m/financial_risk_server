part of 'report_bloc.dart';

enum ReportStatus { initial, loading, loaded, error }
enum UploadStatus { initial, uploading, success, error }

@immutable
class ReportState {
  final ReportStatus status;
  final List<Report> reports;
  final String? error;
  final String? filterType;
  final String? filterStatus;
  
  // Upload state
  final UploadStatus uploadStatus;
  final double uploadProgress;
  final String? uploadError;

  const ReportState({
    required this.status,
    required this.reports,
    this.error,
    this.filterType,
    this.filterStatus,
    this.uploadStatus = UploadStatus.initial,
    this.uploadProgress = 0.0,
    this.uploadError,
  });

  ReportState.initial() : this(status: ReportStatus.initial, reports: []);
  ReportState.loading() : this(status: ReportStatus.loading, reports: []);
  
  ReportState.loaded(List<Report> reports, {String? type, String? status}) 
    : this(status: ReportStatus.loaded, reports: reports, filterType: type, filterStatus: status);
  
  ReportState.error(String message) 
    : this(status: ReportStatus.error, reports: [], error: message);

  ReportState copyWith({
    ReportStatus? status,
    List<Report>? reports,
    String? error,
    String? filterType,
    String? filterStatus,
    UploadStatus? uploadStatus,
    double? uploadProgress,
    String? uploadError,
  }) {
    return ReportState(
      status: status ?? this.status,
      reports: reports ?? this.reports,
      error: error ?? this.error,
      filterType: filterType ?? this.filterType,
      filterStatus: filterStatus ?? this.filterStatus,
      uploadStatus: uploadStatus ?? this.uploadStatus,
      uploadProgress: uploadProgress ?? this.uploadProgress,
      uploadError: uploadError ?? this.uploadError,
    );
  }

  @override
  List<Object?> get props => [status, reports, error, uploadStatus];
}
import 'package:client_flutter/features/reports/domain/entities/report.dart';
import 'package:json_annotation/json_annotation.dart';

part 'report_model.g.dart';

@JsonSerializable()
class ReportModel {
  final int id;
  
  @JsonKey(name: "report_type")
  final String type;
  
  @JsonKey(name: "file_name")
  final String fileName;
  
  @JsonKey(name: "original_name")
  final String originalName;
  
  @JsonKey(name: "upload_date")
  final DateTime uploadDate;
  
  @JsonKey(name: "processing_status")
  final String status;
  
  @JsonKey(name: "file_size_bytes")
  final int sizeBytes;
  
  final String? description;
  
  @JsonKey(name: "period_start")
  final DateTime? periodStart;
  
  @JsonKey(name: "period_end")
  final DateTime? periodEnd;
  
  @JsonKey(name: "uploaded_by")
  final String? uploadedBy;
  
  @JsonKey(name: "download_url")
  final String downloadUrl;

  ReportModel({
    required this.id,
    required this.type,
    required this.fileName,
    required this.originalName,
    required this.uploadDate,
    required this.status,
    required this.sizeBytes,
    this.description,
    this.periodStart,
    this.periodEnd,
    this.uploadedBy,
    required this.downloadUrl,
  });

  factory ReportModel.fromJson(Map<String, dynamic> json) => _$ReportModelFromJson(json);
  Map<String, dynamic> toJson() => _$ReportModelToJson(this);

  Report toEntity() => Report(
    id: id,
    type: type,
    fileName: fileName,
    originalName: originalName,
    uploadDate: uploadDate,
    status: status,
    sizeBytes: sizeBytes,
    description: description,
    period: periodStart != null && periodEnd != null
        ? '${periodStart!.day.toString().padLeft(2, '0')}.${periodStart!.month.toString().padLeft(2, '0')}.${periodStart!.year} — ${periodEnd!.day.toString().padLeft(2, '0')}.${periodEnd!.month.toString().padLeft(2, '0')}.${periodEnd!.year}'
        : null,
    uploadedBy: uploadedBy,
    downloadUrl: downloadUrl,
  );

  factory ReportModel.fromEntity(Report entity) => ReportModel(
    id: entity.id,
    type: entity.type,
    fileName: entity.fileName,
    originalName: entity.originalName,
    uploadDate: entity.uploadDate,
    status: entity.status,
    sizeBytes: entity.sizeBytes,
    description: entity.description,
    uploadedBy: entity.uploadedBy,
    downloadUrl: entity.downloadUrl,
  );
}
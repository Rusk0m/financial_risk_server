// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'report_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

ReportModel _$ReportModelFromJson(Map<String, dynamic> json) => ReportModel(
  id: (json['id'] as num).toInt(),
  type: json['report_type'] as String,
  fileName: json['file_name'] as String,
  originalName: json['original_name'] as String,
  uploadDate: DateTime.parse(json['upload_date'] as String),
  status: json['processing_status'] as String,
  sizeBytes: (json['file_size_bytes'] as num).toInt(),
  description: json['description'] as String?,
  periodStart: json['period_start'] == null
      ? null
      : DateTime.parse(json['period_start'] as String),
  periodEnd: json['period_end'] == null
      ? null
      : DateTime.parse(json['period_end'] as String),
  uploadedBy: json['uploaded_by'] as String?,
  downloadUrl: json['download_url'] as String,
);

Map<String, dynamic> _$ReportModelToJson(ReportModel instance) =>
    <String, dynamic>{
      'id': instance.id,
      'report_type': instance.type,
      'file_name': instance.fileName,
      'original_name': instance.originalName,
      'upload_date': instance.uploadDate.toIso8601String(),
      'processing_status': instance.status,
      'file_size_bytes': instance.sizeBytes,
      'description': instance.description,
      'period_start': instance.periodStart?.toIso8601String(),
      'period_end': instance.periodEnd?.toIso8601String(),
      'uploaded_by': instance.uploadedBy,
      'download_url': instance.downloadUrl,
    };

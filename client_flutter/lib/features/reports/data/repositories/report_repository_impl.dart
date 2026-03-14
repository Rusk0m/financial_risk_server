import 'dart:io';
import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import 'package:path/path.dart' as path;
import 'package:universal_html/html.dart' as html;

import 'package:client_flutter/core/network/api_client.dart';
import 'package:client_flutter/features/reports/data/models/report_model.dart';
import 'package:client_flutter/features/reports/domain/entities/report.dart';
import 'package:client_flutter/features/reports/domain/repositories/report_repository.dart';

class ReportRepositoryImpl implements ReportRepository {
  final ApiClient apiClient;

  ReportRepositoryImpl(this.apiClient);

  @override
  Future<List<Report>> getAll({String? type, String? status}) async {
    try {
      final params = <String, dynamic>{};
      if (type != null && type != 'all') params['type'] = type;
      if (status != null && status != 'all') params['status'] = status;

      final response = await apiClient.dio.get(
        '/reports',
        queryParameters: params.isEmpty ? null : params,
      );

      if (response.data?['success'] == true) {
        final data = response.data['data'] as Map<String, dynamic>;
        if (data['reports'] is List) {
          return (data['reports'] as List)
              .map((json) => ReportModel.fromJson(json as Map<String, dynamic>).toEntity())
              .toList();
        }
      }
      return [];
    } on DioException catch (e) {
      _logError('getAll', e);
      rethrow;
    }
  }

@override
Future<Report> upload({
  required File file,
  required String type,
  String? description,
  DateTime? periodStart,
  DateTime? periodEnd,
}) async {
  try {
    final formData = FormData.fromMap({
      'file': await MultipartFile.fromFile(
        file.path,
        filename: path.basename(file.path),
        contentType: DioMediaType.parse('application/vnd.openxmlformats-officedocument.spreadsheetml.sheet'),
      ),
      'report_type': type,
      // ✅ ДОБАВЛЕНО: enterprise_id (для вашего приложения всегда = 1)
      'enterprise_id': 1,
      // ✅ Опциональные поля
      if (description != null) 'description': description,
      if (periodStart != null) 'period_start': periodStart.toIso8601String(),
      if (periodEnd != null) 'period_end': periodEnd.toIso8601String(),
    });

    final response = await apiClient.dio.post(
      '/reports/upload',
      data:  formData,
      options: Options(
        contentType: 'multipart/form-data',
        receiveTimeout: const Duration(minutes: 5),
        sendTimeout: const Duration(minutes: 5),
      ),
    );

    if (response.data?['success'] == true) {
      return ReportModel.fromJson(response.data['data'] as Map<String, dynamic>).toEntity();
    }

    final error = response.data?['error'];
    throw Exception(error?['message'] ?? 'Ошибка загрузки');
    
  } on DioException catch (e) {
    _logError('upload', e);
    throw Exception(_formatDioError(e));
  }
}

  @override
  Future<String> download(int id) async {
    final url = '${apiClient.dio.options.baseUrl}/reports/$id/download';
    
    // Для Web: открываем в новой вкладке
    if (kIsWeb) { // force web behavior for testing
      html.window.open(url, '_blank');
      return url;
    }
    
    // Для Mobile: можно скачать через Dio и сохранить
    // Реализуется при необходимости
    return url;
  }

  @override
  Future<void> delete(int id) async {
    try {
      await apiClient.dio.delete('/reports/$id');
    } on DioException catch (e) {
      _logError('delete', e);
      rethrow;
    }
  }

  void _logError(String method, DioException e) {
    print('❌ [ReportRepo.$method] ${_formatDioError(e)}');
    print('   URL: ${e.requestOptions.uri}');
    if (e.response != null) {
      print('   Status: ${e.response?.statusCode}');
      print('   Data: ${e.response?.data}');
    }
  }

  String _formatDioError(DioException e) {
    switch (e.type) {
      case DioExceptionType.connectionTimeout:
      case DioExceptionType.sendTimeout:
      case DioExceptionType.receiveTimeout:
        return 'Превышено время ожидания сервера';
      case DioExceptionType.badResponse:
        final data = e.response?.data;
        if (data is Map && data['error'] != null) {
          return data['error']['message'] ?? 'Ошибка сервера';
        }
        return 'Ошибка сервера: ${e.response?.statusCode}';
      case DioExceptionType.connectionError:
        return 'Нет соединения с сервером';
      case DioExceptionType.cancel:
        return 'Запрос отменён';
      default:
        return e.message ?? 'Неизвестная ошибка';
    }
  }
}
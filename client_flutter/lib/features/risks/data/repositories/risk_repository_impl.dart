import 'package:client_flutter/core/network/api_client.dart';
import 'package:client_flutter/features/risks/data/models/risk_result_model.dart';
import 'package:client_flutter/features/risks/domain/repositories/risk_repository.dart';
import 'package:dio/dio.dart';

class RiskRepositoryImpl implements RiskRepository {
  final ApiClient _apiClient;

  RiskRepositoryImpl(this._apiClient);

Map<String, dynamic> _validateRiskResponse(Map<String, dynamic> rawResponse) {
  final validated = <String, dynamic>{};
  
  // Валидируем только ключи рисков
  final riskKeys = ['currency_risk', 'credit_risk', 'liquidity_risk', 'market_risk', 'interest_risk'];
  
  for (final key in riskKeys) {
    final value = rawResponse[key];
    if (value is Map && 
        value.containsKey('risk_type') && 
        value.containsKey('var_value') &&
        value.containsKey('risk_level')) {
      validated[key] = value;
    }
  }
  
  // Добавляем служебные поля
  if (rawResponse.containsKey('overall_risk_level')) {
    validated['overall_risk_level'] = rawResponse['overall_risk_level'];
  }
  if (rawResponse.containsKey('total_risk_value')) {
    validated['total_risk_value'] = rawResponse['total_risk_value'];
  }
  if (rawResponse.containsKey('max_risk_type')) {
    validated['max_risk_type'] = rawResponse['max_risk_type'];
  }
  
  return validated;
}
  @override
  Future<RiskResultModel> calculateRisk({
    required int enterpriseId,
    required String riskType,
    int horizonDays = 30,
    double confidenceLevel = 0.95,
    double? priceChangePct,
    double? rateChangePct,
  }) async {
    try {
      final response = await _apiClient.dio.post(
        '/risks/calculate',
        data: {
          'enterprise_id': enterpriseId,
          'risk_type': riskType,
          'horizon_days': horizonDays,
          'confidence_level': confidenceLevel,
          if (priceChangePct != null) 'price_change_pct': priceChangePct,
          if (rateChangePct != null) 'rate_change_pct': rateChangePct,
        },
      );

      return RiskResultModel.fromJson(response.data['data']);
    } on DioException catch (e) {
      throw Exception('Ошибка расчёта риска: ${e.message}');
    }
  }

  @override
  Future<Map<String, dynamic>> calculateAllRisks({
    required int enterpriseId,
    int horizonDays = 30,
    double confidenceLevel = 0.95,
    double priceChangePct = -15.0,
    double rateChangePct = 1.0,
  }) async {
    try {
      final response = await _apiClient.dio.get(
        '/risks/all',
        queryParameters: {
          'enterprise_id': enterpriseId,
          'horizon_days': horizonDays,
          'confidence_level': confidenceLevel,
          'price_change_pct': priceChangePct,
          'rate_change_pct': rateChangePct,
        },
      );
      // Валидируем ответ перед возвратом
      return _validateRiskResponse(response.data['data'] as Map<String, dynamic>);
    } on DioException catch (e) {
      // Для демо-режима возвращаем mock-данные при ошибке подключения
      if (e.type == DioExceptionType.connectionError) {
        return _getMockRiskData();
      }
      throw Exception('Ошибка расчёта всех рисков: ${e.message}');
    }
  }

  Map<String, dynamic> _getMockRiskData() {
    return {
      'currency_risk': {
        'risk_type': 'currency',
        'risk_level': 'medium',
        'exposure_amount': 143000000,
        'var_value': 12870000,
        'stress_test_loss': 35750000,
      },
      'credit_risk': {
        'risk_type': 'credit',
        'risk_level': 'medium',
        'exposure_amount': 210000000,
        'var_value': 8700000,
        'stress_test_loss': 45000000,
      },
      'liquidity_risk': {
        'risk_type': 'liquidity',
        'risk_level': 'medium',
        'exposure_amount': 2750000000,
        'var_value': 57000000,
        'stress_test_loss': 190000000,
      },
      'market_risk': {
        'risk_type': 'market',
        'risk_level': 'high',
        'exposure_amount': 2422500000,
        'var_value': 363375000,
        'stress_test_loss': 726750000,
      },
      'interest_risk': {
        'risk_type': 'interest',
        'risk_level': 'low',
        'exposure_amount': 5050000000,
        'var_value': 5000000,
        'stress_test_loss': 15000000,
      },
      'overall_risk_level': 'high',
      'total_risk_value': 4469475000,
      'max_risk_type': 'market',
    };
  }
}
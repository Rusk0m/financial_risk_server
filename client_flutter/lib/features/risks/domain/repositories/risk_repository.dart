import 'package:client_flutter/features/risks/data/models/risk_result_model.dart';

abstract class RiskRepository {
  Future<RiskResultModel> calculateRisk({
    required int enterpriseId,
    required String riskType,
    int horizonDays = 30,
    double confidenceLevel = 0.95,
    double? priceChangePct,
    double? rateChangePct,
  });

  Future<Map<String, dynamic>> calculateAllRisks({
    required int enterpriseId,
    int horizonDays = 30,
    double confidenceLevel = 0.95,
    double priceChangePct = -15.0,
    double rateChangePct = 1.0,
  });
}
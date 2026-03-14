import 'package:client_flutter/features/risks/domain/repositories/risk_repository.dart';

class CalculateAllRisks {
  final RiskRepository repository;

  CalculateAllRisks(this.repository);

  Future<Map<String, dynamic>> call(CalculateAllRisksParams params) async {
    return await repository.calculateAllRisks(
      enterpriseId: params.enterpriseId,
      horizonDays: params.horizonDays,
      confidenceLevel: params.confidenceLevel,
      priceChangePct: params.priceChangePct,
      rateChangePct: params.rateChangePct,
    );
  }
}

class CalculateAllRisksParams {
  final int enterpriseId;
  final int horizonDays;
  final double confidenceLevel;
  final double priceChangePct;
  final double rateChangePct;

  CalculateAllRisksParams({
    required this.enterpriseId,
    this.horizonDays = 30,
    this.confidenceLevel = 0.95,
    this.priceChangePct = -15.0,
    this.rateChangePct = 1.0,
  });
}
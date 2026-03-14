part of 'risk_calculation_bloc.dart';

@immutable
class RiskCalculationEvent {}

final class CalculateRiskRequested extends RiskCalculationEvent {
  final int enterpriseId;
  final int horizonDays;
  final double confidenceLevel;
  final double priceChangePct;
  final double rateChangePct;

  CalculateRiskRequested({
    required this.enterpriseId,
    this.horizonDays = 30,
    this.confidenceLevel = 0.95,
    this.priceChangePct = -15.0,
    this.rateChangePct = 1.0,
  });
}
part of 'risk_calculation_bloc.dart';

class RiskCalculationState {
  final RiskCalculationStatus status;
  final Map<String, dynamic>? results;
  final String? error;

  const RiskCalculationState({
    required this.status,
    this.results,
    this.error,
  });

  const RiskCalculationState.initial()
      : this(status: RiskCalculationStatus.initial);

  const RiskCalculationState.loading()
      : this(status: RiskCalculationStatus.loading);

  const RiskCalculationState.loaded(Map<String, dynamic> results)
      : this(status: RiskCalculationStatus.loaded, results: results);

  const RiskCalculationState.error(String error)
      : this(status: RiskCalculationStatus.error, error: error);
}

enum RiskCalculationStatus {
  initial,
  loading,
  loaded,
  error,
}